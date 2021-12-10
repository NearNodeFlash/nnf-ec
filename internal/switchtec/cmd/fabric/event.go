package cmd

import (
	"fmt"
	"sort"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/switchtec"
)

type EventListCmd struct {
	Device string `kong:"arg,required,type='existingFile',env='SWITCHTEC_DEV',help='The switchtec device.'"`
	All    bool   `kong:"optional,short='a',help='Show events in all partitions'"`
	Reset  bool   `kong:"optional,short='r',help='Clear all events'"`
	Event  int    `kong:"optional,short='e',help='Clear all events of the specified type'"`
}

func (cmd *EventListCmd) Run() error {
	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	fmt.Println("Retrieving Event Summary...")
	summary, err := dev.EventSummary()
	if err != nil {
		return err
	}

	fmt.Println("Retrieving Events...")
	events, err := dev.GetEvents(summary, switchtec.EventId(cmd.Event), cmd.All, cmd.Reset, -1)
	if err != nil {
		return err
	}

	fmt.Println("Sort Events...")
	sort.Sort(switchtec.Events(events))

	fmt.Println("Print Events...")
	printEvents(events)

	return nil
}

type EventWaitCmd struct {
	Device    string `kong:"arg,required,type='existingFile',env='SWITCHTEC_DEV',help='The switchtec device.'"`
	Event     uint   `kong:"arg,required,short='e',help='Event to wait on'"`
	Partition int32  `kong:"optional,short='p',default=-1,help='Partition ID for the event'"`
	Port      int32  `kong:"optional,short='p',default=-1,help='Logical port ID for the event'"`
	Timeout   int64  `kong:"optional,short='t',default=-1,help='Timeout in milliseconds (-1 = forever)'"`
}

func (cmd *EventWaitCmd) Run() error {

	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	event := switchtec.EventId(cmd.Event)
	index := int32(0)

	switch event.Type() {
	case switchtec.Invalid_EventType:
		return fmt.Errorf("Event %d not regonized as a valid event", cmd.Event)
	case switchtec.Global_EventType:
		// NO-OP
		break
	case switchtec.Partition_EventType:
		if cmd.Port >= 0 {
			return fmt.Errorf("Port cannot be specified for Partition Event Type")
		}

		if cmd.Partition < 0 {
			index = switchtec.IndexAll
		} else {
			index = cmd.Partition
		}
	case switchtec.Port_EventType:
		if cmd.Partition < 0 && cmd.Port < 0 {
			index = switchtec.IndexAll
		} else if cmd.Partition < 0 || cmd.Port < 0 {
			return fmt.Errorf("Port and Partition are required for Port event type")
		} else {
			pff, err := dev.PortToPff(int32(cmd.Partition), int32(cmd.Port))
			if err != nil {
				return err
			}

			index = int32(pff)
		}
	}

	fmt.Println("Wait For Event...")
	summary, err := dev.EventWaitFor(event, index, cmd.Timeout)
	if err != nil {
		return err
	}

	fmt.Println("Retrieving Events...")
	events, err := dev.GetEvents(summary, event, false, false, index)

	fmt.Println("Sort Events...")
	sort.Sort(switchtec.Events(events))

	fmt.Println("Print Events...")
	printEvents(events)

	return nil
}

func printEvents(events []switchtec.Event) {

	lastPartition, lastPort := -1, -1

	for _, event := range events {

		if event.Partition != lastPartition {
			if event.Partition == -1 {
				fmt.Printf("Global Events:\n")
			} else {
				fmt.Printf("Partition %d Events:\n", event.Partition)
			}
		}

		if event.Port != lastPort && event.Port != -1 {
			if event.Port == switchtec.PffVep {
				fmt.Printf("Port VEP:\n")
			} else {
				fmt.Printf("Port %d:\n", event.Port)
			}
		}

		lastPartition = event.Partition
		lastPort = event.Port

		fmt.Printf("\t%s\n", event.String())
	}
}

type EventGfmsCmd struct {
	Device string `kong:"arg,required,type='existingFile',env='SWITCHTEC_DEV',help='The switchtec device.'"`
	Clear  bool   `kong:"optional,short='c',default=false,help='Clear all GFMS events'"`
}

func (cmd *EventGfmsCmd) Run() error {
	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	if cmd.Clear {
		return dev.ClearGfmsEvents()
	}

	events, err := dev.GetGfmsEvents()
	if err != nil {
		return err
	}

	type eventDetails struct {
		msg     string
		newFunc func(switchtec.GfmsEvent) switchtec.GfmsEventInterface
	}

	eventMap := map[int]eventDetails{
		switchtec.HostLinkUp_GfmsEvent:      {msg: "HOST_LINK_UP        ", newFunc: switchtec.NewHostGfmsEvent},
		switchtec.HostLinkDown_GfmsEvent:    {msg: "HOST_LINK_DOWN      ", newFunc: switchtec.NewHostGfmsEvent},
		switchtec.DeviceAdd_GfmsEvent:       {msg: "DEVICE_ADD          ", newFunc: switchtec.NewDeviceGfmsEvent},
		switchtec.DeviceDelete_GfmsEvent:    {msg: "DEVICE_DELETE       ", newFunc: switchtec.NewDeviceGfmsEvent},
		switchtec.FabricLinkUp_GfmsEvent:    {msg: "FABRIC_LINK_UP      ", newFunc: nil},
		switchtec.FabricLinkDown_GfmsEvent:  {msg: "FABRIC_LINK_DOWN    ", newFunc: nil},
		switchtec.Bind_GfmsEvent:            {msg: "BIND                ", newFunc: switchtec.NewBindGfmsEvent},
		switchtec.Unbind_GfmsEvent:          {msg: "UNBIND              ", newFunc: switchtec.NewBindGfmsEvent},
		switchtec.DatabaseChanged_GfmsEvent: {msg: "DATABASE_CHANGED    ", newFunc: nil},
		switchtec.HvdInstEnable_GfmsEvent:   {msg: "HVD_INSTANCE_ENABLE ", newFunc: switchtec.NewHvdGfmsEvent},
		switchtec.HvdInstDisable_GfmsEvent:  {msg: "HVD_INSTANCE_DISABLE", newFunc: switchtec.NewHvdGfmsEvent},
		switchtec.EpPortAdd_GfmsEvent:       {msg: "EP_PORT_ADD         ", newFunc: switchtec.NewPortGfmsEvent},
		switchtec.EpPortRemove_GfmsEvent:    {msg: "EP_PORT_REMOVE      ", newFunc: switchtec.NewPortGfmsEvent},
		switchtec.Aer_GfmsEvent:             {msg: "AER                 ", newFunc: switchtec.NewAerGfmsEvent},
	}

	for _, event := range events {
		if details, ok := eventMap[event.Code]; ok {
			fmt.Printf("%s (PAX ID %d):\n", details.msg, event.Id)
			if details.newFunc != nil {
				details.newFunc(event).Print()
			}
		} else {
			fmt.Printf("WARNING: Unknown Event Code %d\n", event.Code)
		}
	}

	return nil
}

type EventCmd struct {
	List EventListCmd `kong:"cmd,help='List events (Non-PAX)'"`
	Wait EventWaitCmd `kong:"cmd,help='Wait on events (Non-PAX)'"`
	Gfms EventGfmsCmd `kong:"cmd,help='Display and control GFMS event information (PAX Only)'"`
}
