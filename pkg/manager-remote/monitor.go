package remote

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"os/signal"
	"unsafe"

	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	DEVPATH = "/sys/devices"

	UEventMagicNumber = 0xFEEDCAFE
)

type RemoteServer struct {

	// Local Server Controller for managing devices which are on this server
	//ctrl LocalServerController
}

func (s *RemoteServer) Run() error {

	m := UDevMonitor{}

	if err := m.Open(); err != nil {
		fmt.Printf("Failed to open UDev Monitor: %s", err)
		return err
	}

	events := m.Run()

	// Start the UDev Monitor for receiving UDev Events
	for {
		select {
		case event := <-events:
			fmt.Printf("Event: %s\n", event)

			switch event.Action {
			case KObject_Add, KObject_Remove:

				//if err := s.ctrl.refresh(); err != nil {
				//	return err
				//}
			}
		}
	}
}

type UDevMonitorNetlinkGroup uint32

const (
	UDevMonitor_None UDevMonitorNetlinkGroup = iota
	UDevMonitor_Kernel
	UDevMonitor_UDev
)

// KObject Action defines an Actions for a kernel object
// https://elixir.bootlin.com/linux/v5.12-rc8/source/include/linux/kobject.h
type KObjectAction string

const (
	KObject_Add     KObjectAction = "add"
	KObject_Remove                = "remove"
	KObject_Change                = "change"
	KObject_Move                  = "move"
	KObject_Online                = "online"
	KObject_Offline               = "offline"
	KObject_Bind                  = "bind"
	KObject_Unbind                = "unbind"
)

func ParseKObjectAction(action string) KObjectAction {
	switch KObjectAction(action) {
	case KObject_Add, KObject_Remove, KObject_Change, KObject_Move, KObject_Online, KObject_Offline, KObject_Bind, KObject_Unbind:
		return KObjectAction(action)
	}

	panic(fmt.Sprintf("Unsupported Action %s", action))
}

type UEvent struct {
	Action KObjectAction
	Object string
	Env    map[string]string
}

func (uevent *UEvent) String() string {
	return fmt.Sprintf("%s@%s", uevent.Action, uevent.Object)
}

func (uevent *UEvent) IsNvmeEvent() bool {
	if dev, ok := uevent.Env["DEVNAME"]; ok {
		return strings.HasPrefix(dev, "/dev/nvme")
	}

	return false
}

func NewUEventFromPath(path string) (*UEvent, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	event := UEvent{
		Action: KObject_Add,
		Object: filepath.Dir(path),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.SplitN(scanner.Text(), "=", 2)
		if len(fields) != 2 {
			continue
		}
		event.Env[fields[0]] = fields[1]
	}

	return &event, nil
}

func NewUEventFromBytes(buf []byte) (*UEvent, uint32, error) {

	// [ 0 -  7] "libudev" prefix distinguishes libudev events from kernel messages
	// [ 8 - 11] Magic number to protect against message mismatch; stored in network order
	// [12 - 15] Total length of header structure known to the sender
	// [16 - 19] Properties String Buffer Offset
	// [20 - 23] Properties String Buffer Length
	// [  ...  ] some other fields
	// [40 - ..] usually the data

	if len(buf) < 8 {
		return nil, 0, nil
	}

	if bytes.Compare(buf[:8], []byte("libudev\x00")) == 0 {
		if len(buf) < 40 {
			return nil, 0, nil
		}

		magic := binary.BigEndian.Uint32(buf[8:])
		if magic != UEventMagicNumber {
			return nil, 0, fmt.Errorf("Unsupported UDev Event %#8x", magic)
		}

		offset := *(*uint32)(unsafe.Pointer(&buf[16]))
		length := *(*uint32)(unsafe.Pointer(&buf[20]))

		if len(buf) < int(offset+length) {
			return nil, 0, nil
		}

		fields := bytes.Split(buf[offset:offset+length], []byte{0x00})
		if len(fields) == 0 {
			return nil, 0, fmt.Errorf("Unsupported field count in event: data missing")
		}

		env := make(map[string]string, 0)
		for _, envs := range fields[:len(fields)-1] {
			e := bytes.Split(envs, []byte("="))
			if len(e) != 2 {
				return nil, 0, fmt.Errorf("Unsupported field in event: invalid entry %s", string(envs))
			}

			env[string(e[0])] = string(e[1])
		}

		event := UEvent{
			Action: ParseKObjectAction(env["ACTION"]),
			Object: env["DEVPATH"],
			Env:    env,
		}

		return &event, offset + length, nil
	}

	panic(fmt.Sprintf("Unsupported UDev Event +%v", buf))
}

type UDevMonitor struct {
	fd int

	// Channel for Exit signals
	exit chan struct{}
}

func (m *UDevMonitor) Open() error {
	fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_KOBJECT_UEVENT)
	if err != nil {
		return err
	}

	netlink := syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Groups: uint32(UDevMonitor_UDev),
	}

	if err := syscall.Bind(fd, &netlink); err != nil {
		syscall.Close(fd)
		return err
	}

	m.fd = fd
	m.exit = make(chan struct{}, 1)

	return nil
}

func (m *UDevMonitor) Close() error {
	fmt.Println("UDev Monitor Closing...")
	m.exit <- struct{}{}
	return syscall.Close(m.fd)
}

func (m *UDevMonitor) Run() chan UEvent {

	m.setupSignalHandler()

	events := make(chan UEvent, 1)

	// Process loop - this starts a go routine that continuously
	// reads from the UDev NetLink and processes and received
	// events.
	go func() {
		err := func() error {
			buf := make([]byte, os.Getpagesize())
			offset := uint32(0)

			for {
				// This will read into buffer beginning at offset bytes
				//                     |- offset
				// [---- Previous Data | ------ New Data -------]
				n, _, err := syscall.Recvfrom(m.fd, buf[offset:], 0)
				if err != nil {
					return err
				}

				offset += uint32(n)

				// Start processing the events. There may be zero or more events in
				// the buffer, the contents of which are valid for bytes [0, offset).
				for {
					// Try and parse a UEvent from the buffer bytes, this will return
					// the UEvent (if any) and the number of consumed bytes from the buffer
					// in processing the event, or an error.
					event, consumed, err := NewUEventFromBytes(buf[:offset])
					if err != nil {
						return err
					}

					if event != nil {
						// Notify event occurred
						events <- *event

						// Copy the remaining buffer contents down to the lower buffer
						// bytes and subtract the bytes consumed by the event.
						copy(buf, buf[consumed:offset])
						offset -= consumed
					} else {
						break
					}
				}

				// We've read an entire buffer of data, but that was insufficient for
				// processing the event. Need to grow the buffer by a page size to
				// accomadate for all the event data.
				if int(offset) == len(buf) {
					buf = append(buf, make([]byte, os.Getpagesize())...)
					offset = 0
				}
			}
		}()

		if err != nil {
			fmt.Printf("UDev Monitor Error %s", err)
		}
	}()

	return events
}

// Setup a Signal Handler to quit properly on varios signals. This runs
// a signal handler in the background; when any of the SIG* occur, closes
// the UDevMonitor
func (m *UDevMonitor) setupSignalHandler() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		sig := <-signals
		fmt.Printf("UDev Monitor Signal %s\n", sig)
		m.Close()
	}()
}
