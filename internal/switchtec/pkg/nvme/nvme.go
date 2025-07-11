/*
 * Copyright 2020-2025 Hewlett Packard Enterprise Development LP
 * Other additional copyright holders may be indicated within.
 *
 * The entirety of this work is licensed under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 *
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nvme

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/HewlettPackard/structex"
	"github.com/google/uuid"

	"github.com/NearNodeFlash/nnf-ec/internal/switchtec/pkg/switchtec"
	"github.com/NearNodeFlash/nnf-ec/pkg/ec"
	sf "github.com/NearNodeFlash/nnf-ec/pkg/rfsf/pkg/models"
)

// Device describes the NVMe Device and its attributes
type Device struct {
	Path string
	ops  ops
}

func (dev *Device) String() string { return dev.Path }

// UserIoCmd represents an NVMe User I/O Request
// This is a copy of the C definition in '/include/linux/nvme_ioctl.h' struct nvme_user_io{}
type UserIoCmd struct {
	Opcode   uint8
	Flags    uint8
	Control  uint16
	NBlocks  uint16
	Reserved uint16
	Metadata uint64
	Addr     uint64
	StartLBA uint64
	DsMgmt   uint32
	RefTag   uint32
	AppTag   uint16
	AppMask  uint16
}

// PassthruCmd is used for NVMe Passthrough from user space to the NVMe Device Driver
// This is a copy of the C definition in 'include/linux/nvme_ioctl.h' struct nvme_passthru_cmd{}
type PassthruCmd struct {
	Opcode      uint8  // Byte 0
	Flags       uint8  // Byte 1
	Reserved1   uint16 // Bytes 2-3
	NSID        uint32 // Bytes 4-7
	Cdw2        uint32 // Bytes 8-11
	Cdw3        uint32 // Bytes 12-15
	Metadata    uint64 // Bytes 16-23
	Addr        uint64 // Bytes 24-31
	MetadataLen uint32 // Bytes 32-35
	DataLen     uint32 // Bytes 36-39
	Cdw10       uint32 // Bytes 40-43
	Cdw11       uint32 // Bytes 44-47
	Cdw12       uint32 // Bytes 48-51
	Cdw13       uint32 // Bytes 52-55
	Cdw14       uint32 // Bytes 56-60
	Cdw15       uint32 // Bytes 60-63
	TimeoutMs   uint32 // Bytes 64-67
	Result      uint32 // Bytes 68-71
}

// AdminCmd aliases the PassthruCmd
type AdminCmd = PassthruCmd

func (cmd *AdminCmd) String() string {
	return fmt.Sprintf("OpCode: %s (%#02x)", AdminCommandOpCode(cmd.Opcode), cmd.Opcode)
}

// AdminCommandOpCode sizes the opcodes listed below
type AdminCommandOpCode uint8

// Admin Command Op Codes. These are from the NVMe Specification
const (
	GetLogPage                AdminCommandOpCode = 0x02
	IdentifyOpCode            AdminCommandOpCode = 0x06
	SetFeatures               AdminCommandOpCode = 0x09
	GetFeatures               AdminCommandOpCode = 0x0A
	NamespaceManagementOpCode AdminCommandOpCode = 0x0D
	NamespaceAttachOpCode     AdminCommandOpCode = 0x15
	VirtualMgmtOpCode         AdminCommandOpCode = 0x1c
	NvmeMiSend                AdminCommandOpCode = 0x1D
	NvmeMiRecv                AdminCommandOpCode = 0x1E
	FormatNvmOpCode           AdminCommandOpCode = 0x80
)

func (code AdminCommandOpCode) String() string {
	switch code {
	case GetLogPage:
		return "Get Log Page"
	case IdentifyOpCode:
		return "Identify"
	case SetFeatures:
		return "Set Features"
	case GetFeatures:
		return "Get Features"
	case NamespaceManagementOpCode:
		return "Namespace Management"
	case NamespaceAttachOpCode:
		return "Namespace Attach/Detach"
	case VirtualMgmtOpCode:
		return "Virtualization Management"
	case NvmeMiSend:
		return "Management Interface Send"
	case NvmeMiRecv:
		return "Management Interface Recv"
	case FormatNvmOpCode:
		return "Format"
	}

	return "UNKNOWN"
}

// StatusCode sizes the status codes listed below
type StatusCode uint32

// Constants applying to the StatusCode
const (
	StatusCodeMask         = 0x7FF
	CommandRetryDelayMask  = 0x1800
	CommandRetryDelayShift = 11
	MoreMask               = 0x2000
	DoNotRetryMask         = 0x4000
)

// Status codes
const (
	NamespaceAlreadyAttached StatusCode = 0x118
	NamespaceNotAttached     StatusCode = 0x11a
)

func (sc StatusCode) String() string {
	switch sc {
	case NamespaceAlreadyAttached:
		return "Namespace Already Attached"
	case NamespaceNotAttached:
		return "Namespace Not Attached"
	}

	return "UNKNOWN"
}

// CommandError captures an error and details about handling
type CommandError struct {
	StatusCode        StatusCode
	CommandRetryDelay uint8
	More              bool
	DoNotRetry        bool
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("NVMe Status: %s (%#03x) CRD: %d More: %t DNR: %t", e.StatusCode, uint32(e.StatusCode), e.CommandRetryDelay, e.More, e.DoNotRetry)
}

// NewCommandError generates the default CommandError for use
func newCommandError(status uint32) error {
	return &CommandError{
		StatusCode:        StatusCode(status & StatusCodeMask),
		CommandRetryDelay: uint8((status & CommandRetryDelayMask) >> CommandRetryDelayShift),
		More:              status&MoreMask != 0,
		DoNotRetry:        status&DoNotRetryMask != 0,
	}
}

// Open establishes a connection with the switchtec switch to which subsequent commands are issued.
func Open(devPath string) (*Device, error) {

	if strings.HasPrefix(path.Base(devPath), "nvme") {
		return nil, fmt.Errorf("direct nvme device not suppported yet")
	}

	ops := switchOps{}

	deviceExp := regexp.MustCompile("(?P<pdfid>[^@]*)?@(?P<device>.*)")
	if m := deviceExp.FindStringSubmatch(devPath); m != nil {
		pdfidStr := m[1]
		deviceStr := m[2]

		pdfid, err := strconv.ParseUint(pdfidStr, 0, 16)
		if err != nil {
			return nil, err
		}

		ops.pdfid = uint16(pdfid)
		ops.devPath = deviceStr

		dev, err := switchtec.Open(ops.devPath)
		if err != nil {
			return nil, err
		}

		if err := connectDevice(dev, &ops); err != nil {
			dev.Close()
			return nil, err
		}

		ops.dev = dev

	} else {
		return nil, fmt.Errorf("Unknown device path")
	}

	return &Device{
		Path: ops.devPath,
		ops:  &ops}, nil
}

// Connect connects a PDFID to an existing switchtec device
func Connect(dev *switchtec.Device, pdfid uint16) (*Device, error) {
	ops := switchOps{
		dev:   dev,
		pdfid: pdfid,
	}

	if err := connectDevice(dev, &ops); err != nil {
		return nil, err
	}

	device := &Device{
		Path: fmt.Sprintf("%#x@%s", pdfid, dev.Path),
		ops:  &ops,
	}

	return device, nil
}

func connectDevice(dev *switchtec.Device, ops *switchOps) (err error) {
	if ops.tunnelStatus, err = dev.EndPointTunnelStatus(ops.pdfid); err != nil {
		return err
	}

	if ops.tunnelStatus == switchtec.DisabledEpTunnelStatus {
		if err = dev.EndPointTunnelEnable(ops.pdfid); err != nil {
			return err
		}

		for true {
			status, err := dev.EndPointTunnelStatus(ops.pdfid)
			if err != nil {
				return err
			}

			if status == switchtec.EnabledEpTunnelStatus {
				break
			}

			time.Sleep(100 * time.Millisecond)
		}

		ops.tunnelStatus = switchtec.EnabledEpTunnelStatus
	}

	return nil
}

// Close closes connection with the switchtec switch
func (dev *Device) Close() {
	if dev.ops != nil {
		dev.ops.close()
	}

	dev.ops = nil
}

const (
	IdentifyDataSize uint32 = 4096
)

// IdentifyControllerOrNamespaceType constrains the type for the constants below.
type IdentifyControllerOrNamespaceType int // CNS

// CNS constants
const (
	Namespace_CNS                     IdentifyControllerOrNamespaceType = 0x00
	Controller_CNS                    IdentifyControllerOrNamespaceType = 0x01
	NamespaceList_CNS                 IdentifyControllerOrNamespaceType = 0x02
	NamespaceDescriptorList_CNS       IdentifyControllerOrNamespaceType = 0x03
	NVMSetList_CNS                    IdentifyControllerOrNamespaceType = 0x04
	NamespacePresentList_CNS          IdentifyControllerOrNamespaceType = 0x10
	NamespacePresent_CNS              IdentifyControllerOrNamespaceType = 0x11
	ControllerNamespaceList_CNS       IdentifyControllerOrNamespaceType = 0x12
	ControllerList_CNS                IdentifyControllerOrNamespaceType = 0x13
	PrimaryControllerCapabilities_CNS IdentifyControllerOrNamespaceType = 0x14
	SecondaryControllerList_CNS       IdentifyControllerOrNamespaceType = 0x15
	NamespaceGranularityList_CNS      IdentifyControllerOrNamespaceType = 0x16
	UUIDList_CNS                      IdentifyControllerOrNamespaceType = 0x17
)

type NamespaceIdentifier uint32

const (
	// COMMON_NAMESPACE_IDENTIFIER identifies the capabilities and settings that are common to all namespaces
	// in the Identify Namespace data structure
	COMMON_NAMESPACE_IDENTIFIER NamespaceIdentifier = 0xFFFFFFFF
)

// NamespaceGloballyUniqueIdentifier uniquely identifies the namespace
type NamespaceGloballyUniqueIdentifier [16]byte

func (a NamespaceGloballyUniqueIdentifier) String() string {
	return fmt.Sprintf("%x", string(a[:]))
}

// Parse - ensure the uuid is parseable
func (a *NamespaceGloballyUniqueIdentifier) Parse(s string) {
	*a = NamespaceGloballyUniqueIdentifier(uuid.MustParse(s))
}

// IdentifyNamespace -
func (dev *Device) IdentifyNamespace(namespaceId uint32, present bool) (*IdNs, error) {

	ns := new(IdNs)

	cns := Namespace_CNS
	if present {
		cns = NamespacePresent_CNS
	}

	buf := structex.NewBuffer(ns)

	if err := dev.Identify(namespaceId, cns, buf.Bytes()); err != nil {
		return nil, err
	}

	if err := structex.Decode(buf, ns); err != nil {
		return nil, err
	}

	return ns, nil
}

// IdentifyNamespaceList -
func (dev *Device) IdentifyNamespaceList(namespaceId uint32, all bool) ([1024]NamespaceIdentifier, error) {

	var data struct {
		Identifiers [1024]NamespaceIdentifier
	}

	cns := NamespaceList_CNS // Display only controller active NS'
	if all {
		cns = NamespacePresentList_CNS // Display all namespaces present on the device
	}

	buf := structex.NewBuffer(data)

	if err := dev.Identify(namespaceId, cns, buf.Bytes()); err != nil {
		return data.Identifiers, err
	}

	if err := structex.Decode(buf, &data); err != nil {
		return data.Identifiers, err
	}

	return data.Identifiers, nil
}

// IdentifyController -
func (dev *Device) IdentifyController() (*IdCtrl, error) {
	id := new(IdCtrl)

	buf := structex.NewBuffer(id)

	if err := dev.Identify(0, Controller_CNS, buf.Bytes()); err != nil {
		return nil, err
	}

	if err := structex.Decode(buf, id); err != nil {
		return nil, err
	}

	return id, nil
}

// IdentifyNamespaceControllerList - Identifies the list of controllers attached to the provided namespace
func (dev *Device) IdentifyNamespaceControllerList(namespaceId uint32) (*CtrlList, error) {

	ctrlList := new(CtrlList)
	buf := structex.NewBuffer(ctrlList)

	if err := dev.Identify(namespaceId, ControllerNamespaceList_CNS, buf.Bytes()); err != nil {
		return nil, err
	}

	if err := structex.Decode(buf, ctrlList); err != nil {
		return nil, err
	}

	return ctrlList, nil
}

func (dev *Device) IdentifyPrimaryControllerCapabilities(controllerId uint16) (*CtrlCaps, error) {

	caps := new(CtrlCaps)
	buf := structex.NewBuffer(caps)

	var cdw10 uint32
	cdw10 = uint32(controllerId) << 16
	cdw10 |= uint32(PrimaryControllerCapabilities_CNS)

	if err := dev.IdentifyRaw(0, cdw10, 0, buf.Bytes()); err != nil {
		return nil, err
	}

	if err := structex.Decode(buf, caps); err != nil {
		return nil, err
	}

	return caps, nil
}

// Identify -
func (dev *Device) Identify(namespaceId uint32, cns IdentifyControllerOrNamespaceType, data []byte) error {
	return dev.IdentifyRaw(namespaceId, uint32(cns), 0, data)
}

// IdentifyRaw -
func (dev *Device) IdentifyRaw(namespaceId uint32, cdw10 uint32, cdw11 uint32, data []byte) error {
	var addr uint64 = 0
	if data != nil {
		addr = uint64(uintptr(unsafe.Pointer(&data[0])))
	}

	cmd := AdminCmd{
		Opcode:  uint8(IdentifyOpCode),
		NSID:    namespaceId,
		Addr:    addr,
		DataLen: IdentifyDataSize,
		Cdw10:   cdw10,
		Cdw11:   cdw11,
	}

	return dev.ops.submitAdminPassthru(dev, &cmd, data)
}

type OptionalControllerCapabilities uint16

const (
	SecuritySendReceiveCapabilities OptionalControllerCapabilities = (1 << 0)
	FormatNVMCommandSupport                                        = (1 << 1)
	FirmwareCommitImageDownload                                    = (1 << 2)
	NamespaceManagementCapability                                  = (1 << 3)
	DeviceSelfTestCommand                                          = (1 << 4)
	DirectivesSupport                                              = (1 << 5)
	NVMeMISendReceiveSupport                                       = (1 << 6)
	VirtualizationManagementSupport                                = (1 << 7)
	DoorbellBufferConfigCommand                                    = (1 << 8)
	GetLBAStatusCapability                                         = (1 << 9)
)

// Identify Controller Data Structure
// Figure 249 from NVM-Express 1_4a-2020.03.09-Ratified specification
type id_ctrl struct {
	PCIVendorId                 uint16   // VID
	PCISubsystemVendorId        uint16   // SSVID
	SerialNumber                [20]byte // SN
	ModelNumber                 [40]byte // MN
	FirmwareRevision            [8]byte  // FR
	RecommendedArbitrationBurst uint8    // RAB
	IEEOUIIdentifier            [3]byte  // IEEE
	ControllerCapabilities      struct {
		MultiPort                 uint8 `bitfield:"1"` // Bit 0: NVM subsystem may contain more than one NVM subsystem port
		MultiController           uint8 `bitfield:"1"` // Bit 1: NVM subsystem may contain two or more controllers
		SRIOVVirtualFunction      uint8 `bitfield:"1"` // Bit 2: The controller is associated with an SR-IOV Virtual Function.
		AsymmetricNamespaceAccess uint8 `bitfield:"1"` // Bit 3: NVM subsystem supports Asymmetric Namespace Access Reporting.
		Reserved                  uint8 `bitfield:"4"`
	} // CMIC
	MaximumDataTransferSize                   uint8  // MDTS
	ControllerId                              uint16 // CNTLID
	Version                                   uint32 // VER
	RTD3ResumeLatency                         uint32 // RTD3R
	RTD3EntryLatency                          uint32 // RTD3E
	AsynchronousEventsSupport                 uint32 // OAES
	ControllerAttributes                      uint32 // CTRATT
	ReadRecoveryLevelsSupported               uint16 // RRLS
	Reserved102                               [9]byte
	ControllerType                            uint8    // CNTRLTYPE
	FRUGloballyUniqueIdentifier               [16]byte // FGUID
	CommandRetryDelayTime1                    uint16   // CRDT1
	CommandRetryDelayTime2                    uint16   // CRDT2
	CommandRetryDelayTime3                    uint16   // CRDT3
	Reserved134                               [122]byte
	OptionalAdminCommandSupport               uint16               // OACS
	AbortCommandLimit                         uint8                // ACL
	AsynchronousEventRequestLimit             uint8                // AERL
	FirmwareUpdates                           uint8                // FRMW
	LogPageAttributes                         uint8                // LPA
	ErrorLogPageEntries                       uint8                // ELPE
	NumberPowerStatesSupport                  uint8                // NPSS
	AdminVendorSpecificCommandConfiguration   uint8                // AVSCC
	AutonomousPowerStateTranstionAttributes   uint8                // APSTA
	WarningCompositeTemperatureThreshold      uint16               // WCTEMP
	CriticalCompositeTemperatureThreashold    uint16               // CCTEMP
	MaximumTimeForFirmwareActivation          uint16               // MTFA 100ms unites
	HostMemoryBufferPreferredSize             uint32               // HMPRE 4 KiB Units
	HostMemoryBufferMinimumSize               uint32               // HMMIN 4KiB Units
	TotalNVMCapacity                          [16]byte             // TNVMCAP
	UnallocatedNVMCapacity                    [16]byte             // UNVMCAP
	ReplayProtectedMemoryBlockSupport         uint32               // RPMBS
	ExtendedDriveSelfTestTime                 uint16               // EDSTT
	DeviceSelfTestOptions                     uint8                // DSTO
	FirmwareUpdateGranularity                 uint8                // FWUG
	KeepAliveSupport                          uint16               // KAS
	HostControllerThermalManagementAttributes uint16               // HCTMA
	MinimumThermalManagementTemperature       uint16               // MNTMT
	MaximumThermalManagementTemperature       uint16               // MXTMT
	Sanitize                                  SanitizeCapabilities //SANICAP
	HostMemoryBufferMinimDescriptorEntrySize  uint32               // HMMINDS
	HostMemoryMaximumDescriptorEntries        uint16               // HMMAXD
	NVMSetIdentifierMaximum                   uint16               // NSETIDMAX
	EnduraceGroupIdentifierMaximum            uint16               // ENDGIDMAX
	ANATranstitionTime                        uint8                // ANATT
	AsymmetricNamespaceAccessCapabilities     uint8                // ANACAP
	ANAGroupIdenfierMaximum                   uint32               // ANAGRPMAX
	NumberOfANAGroupIdentifiers               uint32               // NANAGRPID
	PersistentEventLogSize                    uint32               // PELS
	Reserved356                               [156]byte
	CommandSetAttributes                      struct {
		SubmissionQueueEntrySize              uint8  // SQES
		CompletionQueueEntrySize              uint8  // CQES
		MaximumOutstandingCommands            uint16 // MAXCMD
		NumberOfNamespaces                    uint32 // NN
		OptionalNVMCommandSupport             uint16 // ONCS
		FuesedOperationSupport                uint16 // FUSES
		FormatNVMAttributes                   uint8  // FNA
		VolatileWriteCache                    uint8  // VWC
		AtomicWriteUnitNormal                 uint16 // AWUN
		AtomicWriteUnitPowerFail              uint16 // AWUPF
		NVMVendorSpecificCommandConfiguration uint8  // NVSCC
		NamespaceWriteProtectionCapabilities  uint8  // NWPC
		AtomicCompareAndWriteUnit             uint16 // ACWU
	}
	Reserved534                      [2]byte
	SGLSupport                       uint32 // SGLS
	MaximumNumberOfAllowedNamespaces uint32 // MNAN
	Reserved544                      [224]byte
	NVMSubsystemNVMeQualifiedName    [256]byte // SUBNQN
	Reserved1024                     [768]byte
	IOCCSZ                           uint32
	IORCSZ                           uint32
	ICDOFF                           uint16
	CTRATTR                          uint8
	MSBDB                            uint8
	Reserved1804                     [244]byte
	PowerStateDescriptors            [32]id_power_state
	VendorSpecific                   [1024]byte
}

// SanitizeCapabilities data structure
// structex annotations where applicable.
type SanitizeCapabilities struct {
	CryptoErase               uint8  `bitfield:"1"` // Bits 0     controller supports Crypto Erase sanitize operation (1), 0 - no support
	BlockErase                uint8  `bitfield:"1"` // Bits 1     controller supports Block Erase sanitize operation (1), 0 - no support
	Overwrite                 uint8  `bitfield:"1"` // Bits 2     controller supports Overwrite sanitize operation (1), 0 - no support
	Unused0                   uint8  `bitfield:"5"` // Bits 7:3   unused
	Unused1                   uint16 // Bits 23:8  unused
	Unused2                   uint8  `bitfield:"5"` // Bits 28:24 unused
	NoDeallocateAfterSanitize uint8  `bitfield:"1"` // Bits 29    No-deallocate After Sanitize bit in Sanitize inhibitied command support (1), 0 - no support
	NoDeallocateMask          uint8  `bitfield:"2"` // Bits 31:30 No-deallocate After Sanitize mask to extract value.
}

type id_power_state struct {
	MaxPowerInCentiwatts    uint16
	Reserved2               uint8
	Flags                   uint8
	EntryLatency            uint32 // EXLAT microseconds
	ExitLatency             uint32 // EXLAT microseconds
	RelativeReadThroughpout uint8  // RRT
	RelativeReadLatency     uint8  // RRL
	RelativeWriteThroughput uint8  // RWT
	RelativeWriteLatency    uint8  // RWL
	IdlePower               uint16 // IDLP
	IdlePowerScale          uint8  // IPS
	Reserved19              uint8
	ActivePower             uint16 // ACTP
	ActivePowerScale        uint8  // APS
	Reserved23              [9]byte
}

type IdCtrl id_ctrl

// GetCapability -
func (id *IdCtrl) GetCapability(cap OptionalControllerCapabilities) bool {
	return id.OptionalAdminCommandSupport&uint16(cap) != 0
}

// Identify Namespace Data Structure
// Figure 247 from NVM-Express 1_4a-2020.03.09-Ratified specification
//
// All the parameters listed are the long-form name with the 'Namespace'
// prefix dropped, if it exists. structex annotations where applicable.
type IdNs struct {
	Size        uint64 // NSZE
	Capacity    uint64 // NCAP
	Utilization uint64 // NUSE
	Features    struct {
		Thinp    uint8 `bitfield:"1"` // Bit 0: Thin provisioning supported for this namespace
		Nsabp    uint8 `bitfield:"1"` // Bit 1: NAWUN, NAWUPF, NACWU are defined for this namespace (TODO: Wtf is this?)
		Dae      uint8 `bitfield:"1"` // Bit 2: Deallocated or Unwritten Logical Block error support for this namespace
		Uidreuse uint8 `bitfield:"1"` // Bit 3: NGUID field for this namespace, if non-zero, is never reused by the controller.
		OptPerf  uint8 `bitfield:"1"` // Bit 4: If 1, indicates NPWG, NPWA, NPDG, NPDA, NOWS are defined for this namespace and should be used for host I/O optimization.
		Reserved uint8 `bitfield:"3"` // Bits 7:5 are reserved
	} // NSFEAT
	NumberOfLBAFormats   uint8            // NLBAF
	FormattedLBASize     FormattedLBASize // FLABS
	MetadataCapabilities struct {
		ExtendedSupport uint8 `bitfield:"1"` // Bit 0: If 1, indicates support metadata transfer as part of extended data LBA
		PointerSupport  uint8 `bitfield:"1"` // Bit 1: If 1, indicates support for metadata transfer in a separate buffer specified by Metadata Pointer in SQE
		Reserved        uint8 `bitfield:"6"` // Bits 7:2 are reserved
	} // MC
	EndToEndDataProtectionCaps         uint8                 // DPC
	EndToEndDataProtectionTypeSettings uint8                 // DPS
	MultiPathIOSharingCapabilities     NamespaceCapabilities // NMIC
	ReservationCapabilities            struct {
		PersistThroughPowerLoss        uint8 `bitfield:"1"` // Bit 0: If 1 indicates namespace supports the Persist Through Power Loss capability
		WriteExclusiveReservation      uint8 `bitfield:"1"` // Bit 1: If 1 indicates namespace supports the Write Exclusive reservation type
		ExclusiveAccessReservation     uint8 `bitfield:"1"` // Bit 2: If 1 indicates namespace supports Exclusive Access reservation type
		WriteExclusiveRegistrantsOnly  uint8 `bitfield:"1"` // Bit 3:
		ExclusiveAccessRegistrantsOnly uint8 `bitfield:"1"` // Bit 4:
		WriteExclusiveAllRegistrants   uint8 `bitfield:"1"` // Bit 5:
		ExclusiveAllRegistrants        uint8 `bitfield:"1"` // Bit 6:
		IgnoreExistingKey              uint8 `bitfield:"1"` // Bit 7: If 1 indicates that the Ingore Existing Key is used as defined in 1.3 or later specification. If 0, 1.2.1 or earlier definition
	} // RESCAP
	FormatProgressIndicator struct {
		PercentageRemaining   uint8 `bitfield:"7"` // Bits 6:0 indicates percentage of the Format NVM command that remains completed. If bit 7 is set to '1' then a value of 0 indicates the namespace is formatted with FLBAs and DPS fields and no format is in progress
		FormatProgressSupport uint8 `bitfield:"1"` // Bit 7 if 1 indicates the namespace supports the Format Progress Indicator field
	} // FPI
	DeallocateLogiclBlockFeatures  uint8     // DLFEAT
	AtomicWriteUnitNormal          uint16    // NAWUN
	AtomicWriteUnitPowerFail       uint16    // NAWUPF
	AtomicCompareAndWriteUnit      uint16    // NACWU
	AtomicBoundarySizeNormal       uint16    // NABSN
	AtomicBoundaryOffset           uint16    // NABO
	AtomicBoundarySizePowerFail    uint16    // NABSPF
	OptimalIOBoundary              uint16    // NOIOB
	NVMCapacity                    [16]uint8 // NVMCAP Total size of the NVM allocated to this namespace, in bytes.
	PreferredWriteGranularity      uint16    // NPWG
	PreferredWriteAlignment        uint16    // NPWA
	PreferredDeallocateGranularity uint16    // NPDG
	PreferredDeallocateAlignment   uint16    // NPDA
	OptimalWriteSize               uint16    // NOWS The size in logical block for optimal write performance for this namespace. Should be a multiple of NPWG. Refer to 8.25 for details
	Reserved74                     [18]uint8
	AnaGroupIdentifier             uint32 // ANAGRPID indicates the ANA Group Identifier for thsi ANA group (refer to section 8.20.2) of which the namespace is a member.
	Reserved96                     [3]uint8
	Attributes                     struct {
		WriteProtected uint8 `bitfield:"1"` // Bit 0: indicates namespace is currently write protected due to any error condition
		Reserved       uint8 `bitfield:"7"` // Bits 7:1 are reserved
	} // NSATTR
	NvmSetIdentifier             uint16                            // NVMSETID
	EnduranceGroupIdentifier     uint16                            // ENDGID
	GloballyUniqueIdentifier     NamespaceGloballyUniqueIdentifier // NGUID
	IeeeExtendedUniqueIdentifier [8]uint8                          // EUI64
	LBAFormats                   [16]struct {
		MetadataSize        uint16 // MS
		LBADataSize         uint8  // LBADS Indicates the LBA data size supported in terms of a power-of-two value. If the value is 0, the the LBA format is not supported
		RelativePerformance uint8  // RP indicates the relative performance of this LBA format relative to other LBA formats
	}
	Reserved192    [192]uint8
	VendorSpecific [3712]uint8
}

// Format Options Data Structure
// All the parameters listed are the long-form name
// structex annotations where applicable.
type FormatNs struct {
	Format                 uint8    `bitfield:"4"` // Bits 3:0  Indicate one of the 16 supported LBA Formats
	MetadataSetting        uint8    `bitfield:"1"` // Bits 4    Indicates metadata transfer setting. If the Metadata size is 0, the bit is N/A
	ProtectionInfo         uint8    `bitfield:"3"` // Bits 7:5  Indicates end-to-end protection information type
	ProtectionInfoLocation uint8    `bitfield:"1"` // Bits 8    Indicates protection information location in the metadata
	SecureEraseSetting     uint8    `bitfield:"3"` // Bits 11-9 Indicates secure erase setting, 0 - none, 1 - user data erase, 2 - crypto-erase
	Fill                   uint8    `bitfield:"4"` // Filler to pad to byte
	Unused                 [2]uint8 // Pad to uint32
}

type FormattedLBASize struct {
	Format   uint8 `bitfield:"4"` // Bits 3:0: Indicates one of the 16 supported LBA Formats
	Metadata uint8 `bitfield:"1"` // Bit 4: If 1, indicates metadata is transfer at the end of the data LBA
	Reserved uint8 `bitfield:"3"` // Bits 7:5 are reserved
} // FLBAS

type NamespaceCapabilities struct {
	Sharing  uint8 `bitfield:"1"` // Bit 0: If 1, namespace may be attached to two ro more controllers in the NVM subsystem concurrently
	Reserved uint8 `bitfield:"7"` // Bits 7:1 are reserved
} // NMIC

type CtrlList struct {
	Num         uint16 `countOf:"Identifiers"`
	Identifiers [2047]uint16
}

type CtrlCaps struct {
	ControllerId                           uint16 // CNTLID
	PortId                                 uint16 // PORTID
	ControllerResourceType                 uint8  // CRT
	Reserved5                              [27]uint8
	VQResourcesFlexibleTotal               uint32 // VQFRT
	VQResourcesFlexibleAssigned            uint32 // VQRFA
	VQResourcesFlexibleAllocatedToPrimary  uint16 // VQRFAP
	VQResourcesPrivateTotal                uint16 // VQPRT
	VQResourcesFlexibleSecondaryMaximum    uint16 // VQFRSM
	VQFlexibleResourcePreferredGranularity uint16 // VQGRAN
	Reserved48                             [16]uint8
	VIResourcesFlexibleTotal               uint32 // VIFRT
	VIResourcesFlexibleAssigned            uint32 // VIRFA
	VIResourcesFlexibleAllocatedToPrimary  uint16 // VIRFAP
	VIResourcesPrivateTotal                uint16 // VIPRT
	VIResourcesFlexibleSecondaryMaximum    uint16 // VIFRSM
	VIFlexibleResourcePreferredGranularity uint16 // VIGRAN
	Reserved90                             [4016]uint8
}

// CreateNamespace creates a new namespace with the specified parameters
func (dev *Device) CreateNamespace(size uint64, capacity uint64, format uint8, dps uint8, sharing uint8, anagrpid uint32, nvmsetid uint16, timeout uint32) (uint32, error) {

	ns := IdNs{
		Size:                               size,
		Capacity:                           capacity,
		FormattedLBASize:                   FormattedLBASize{Format: format},
		EndToEndDataProtectionTypeSettings: dps,
		MultiPathIOSharingCapabilities:     NamespaceCapabilities{Sharing: sharing},
		AnaGroupIdentifier:                 anagrpid,
		NvmSetIdentifier:                   nvmsetid,
	}

	buf, err := structex.EncodeByteBuffer(ns)
	if err != nil {
		return 0, err
	}

	cmd := AdminCmd{
		Opcode:    uint8(NamespaceManagementOpCode),
		Addr:      uint64(uintptr(unsafe.Pointer(&ns))),
		Cdw10:     0,
		DataLen:   uint32(len(buf)),
		TimeoutMs: timeout,
	}

	if err := dev.ops.submitAdminPassthru(dev, &cmd, buf); err != nil {
		return 0, err
	}

	return cmd.Result, nil
}

// DeleteNamespace deletes the specified namespace
func (dev *Device) DeleteNamespace(namespaceID uint32) error {
	cmd := AdminCmd{
		Opcode: uint8(NamespaceManagementOpCode),
		NSID:   namespaceID,
		Cdw10:  1,
	}

	return dev.ops.submitAdminPassthru(dev, &cmd, nil)
}

// FormatNamespace issues a suitable format command to the namespace.
// The existing format is queried and reused, and if crypto erase
// is supported we chose that.
func (dev *Device) FormatNamespace(namespaceID uint32) error {

	idctrl, err := dev.IdentifyController()
	if err != nil {
		return err
	}

	idns, err := dev.IdentifyNamespace(namespaceID, true /* namespace present */)
	if err != nil {
		return err
	}

	var secureEraseSetting uint8
	if idctrl.Sanitize.CryptoErase == 1 {
		secureEraseSetting = 2
	}

	// If the drive supports crypto erase, use it.
	formatOptions := FormatNs{
		Format:             idns.FormattedLBASize.Format,
		SecureEraseSetting: secureEraseSetting,
	}

	buf, err := structex.EncodeByteBuffer(formatOptions)
	if err != nil {
		return err
	}

	cmd := AdminCmd{
		Opcode: uint8(FormatNvmOpCode),
		NSID:   namespaceID,
		Cdw10:  binary.LittleEndian.Uint32(buf),
	}

	return dev.ops.submitAdminPassthru(dev, &cmd, nil)
}

func (dev *Device) manageNamespace(namespaceID uint32, controllers []uint16, attach bool) error {

	list := CtrlList{
		Num: uint16(len(controllers)),
	}

	copy(list.Identifiers[:], controllers[:])

	buf := structex.NewBuffer(list)
	if buf == nil {
		return fmt.Errorf("Buffer allocation failed")
	}

	if err := structex.Encode(buf, list); err != nil {
		return fmt.Errorf("Encoding error")
	}

	cmd := AdminCmd{
		Opcode:  uint8(NamespaceAttachOpCode),
		NSID:    namespaceID,
		Addr:    uint64(uintptr(unsafe.Pointer(&list))),
		Cdw10:   map[bool]uint32{true: 0, false: 1}[attach],
		DataLen: 0x1000,
	}

	return dev.ops.submitAdminPassthru(dev, &cmd, buf.Bytes())
}

func (dev *Device) AttachNamespace(namespaceID uint32, controllers []uint16) error {
	return dev.manageNamespace(namespaceID, controllers, true)
}

func (dev *Device) DetachNamespace(namespaceID uint32, controllers []uint16) error {
	return dev.manageNamespace(namespaceID, controllers, false)
}

type VirtualManagementResourceType uint32

const (
	VQResourceType VirtualManagementResourceType = 0
	VIResourceType VirtualManagementResourceType = 1
)

var (
	VirtualManagementResourceTypeName = map[VirtualManagementResourceType]string{
		VQResourceType: "VQ",
		VIResourceType: "VI",
	}
)

type VirtualManagementAction uint32

const (
	PrimaryFlexibleAction  VirtualManagementAction = 0x1
	SecondaryOfflineAction VirtualManagementAction = 0x7
	SecondaryAssignAction  VirtualManagementAction = 0x8
	SecondaryOnlineAction  VirtualManagementAction = 0x9
)

func (dev *Device) VirtualMgmt(ctrlId uint16, action VirtualManagementAction, resourceType VirtualManagementResourceType, numResources uint32) error {

	var cdw10 uint32
	cdw10 = uint32(ctrlId) << 16
	cdw10 = cdw10 | uint32(resourceType<<8)
	cdw10 = cdw10 | uint32(action<<0)

	cmd := AdminCmd{
		Opcode: uint8(VirtualMgmtOpCode),
		Cdw10:  cdw10,
		Cdw11:  numResources,
	}

	return dev.ops.submitAdminPassthru(dev, &cmd, nil)

}

type SecondaryControllerEntry struct { // TODO: Add structex annotations
	SecondaryControllerID       uint16
	PrimaryControllerID         uint16
	SecondaryControllerState    uint8
	Reserved5                   [3]uint8
	VirtualFunctionNumber       uint16
	VQFlexibleResourcesAssigned uint16
	VIFlexibleResourcesAssigned uint16
	Reserved14                  [18]uint8
}

type SecondaryControllerList struct {
	Count    uint8 `countOf:"Entries"`
	Reserved [31]uint8
	Entries  [127]SecondaryControllerEntry
}

// ListSecondary retries the secondary controller list associated with the
// primary controller of the given device. This differs from the C implementation
// in that num-entries is not an option. The maximum number of entries is
// always returned. It is up to the caller to trim down the response.
func (dev *Device) ListSecondary(startingCtrlId uint32, namespaceId uint32) (*SecondaryControllerList, error) {

	list := new(SecondaryControllerList)

	buf := structex.NewBuffer(list)
	if buf == nil {
		return nil, fmt.Errorf("Cannot allocate bufffer")
	}

	if err := dev.IdentifyRaw(namespaceId, (startingCtrlId<<16)|uint32(SecondaryControllerList_CNS), 0, buf.Bytes()); err != nil {
		return nil, err
	}

	if err := structex.Decode(buf, &list); err != nil {
		return nil, err
	}

	return list, nil
}

type Feature uint8

const (
	NoneFeature                  Feature = 0x00
	ArbitrationFeature                   = 0x01
	PowerManagmentFeature                = 0x02
	LBARangeFeature                      = 0x03
	TemperatureThresholdFeature          = 0x04
	ErrorRecoveryFeature                 = 0x05
	VolatileWriteCacheFeature            = 0x06
	NumQueuesFeature                     = 0x07
	IRQCoalesceFeature                   = 0x08
	IRQConfigFeature                     = 0x09
	WriteAtomicFeature                   = 0x0a
	AsyncEventFeature                    = 0x0b
	AutoPSTFeature                       = 0x0c
	HostMemoryBufferFeature              = 0x0d
	TimestampFeature                     = 0x0e
	KATOFeature                          = 0x0f
	HCTMFeature                          = 0x10
	NoPSCFeature                         = 0x11
	RRLFeature                           = 0x12
	PLMConfigFeature                     = 0x13
	PLMWindowFeature                     = 0x14
	LBAStatusInfoFeature                 = 0x15
	HostBehaviorFeature                  = 0x16
	SanitizeFeature                      = 0x17
	EnduranceFeature                     = 0x18
	IOCSProfileFeature                   = 0x19
	SWProgressFeature                    = 0x80
	HostIDFeature                        = 0x81
	ReservationMaskFeature               = 0x82
	ReservationPersistentFeature         = 0x83
	WriteProtectFeature                  = 0x84
	MiControllerMetadata                 = 0x7E
	MiNamespaceMetadata                  = 0x7F
)

var FeatureBufferLength = [256]uint32{
	LBARangeFeature:         4096,
	AutoPSTFeature:          256,
	HostMemoryBufferFeature: 256,
	TimestampFeature:        8,
	PLMConfigFeature:        512,
	HostBehaviorFeature:     512,
	HostIDFeature:           8,
	MiControllerMetadata:    4096,
	MiNamespaceMetadata:     4096,
}

func (dev *Device) GetFeature(nsid uint32, fid Feature, sel int, cdw11 uint32, len uint32, buf []byte) error {
	cdw10 := uint32(fid) | uint32(sel)<<8

	return dev.feature(GetFeatures, nsid, cdw10, cdw11, 0, len, buf)
}

func (dev *Device) SetFeature(nsid uint32, fid Feature, cdw12 uint32, save bool, len uint32, buf []byte) error {
	cdw10 := uint32(fid)
	if save {
		cdw10 = cdw10 | (1 << 31)
	}

	return dev.feature(SetFeatures, nsid, cdw10, 0, cdw12, len, buf)
}

func (dev *Device) feature(opCode AdminCommandOpCode, nsid uint32, cdw10, cdw11, cdw12 uint32, len uint32, buf []byte) error {

	cmd := AdminCmd{
		Opcode:  uint8(opCode),
		NSID:    nsid,
		Cdw10:   cdw10,
		Cdw11:   cdw11,
		Cdw12:   cdw12,
		Addr:    uint64(uintptr(unsafe.Pointer(&buf[0]))),
		DataLen: len,
	}

	return dev.ops.submitAdminPassthru(dev, &cmd, buf)
}

type MIHostMetadata struct {
	NumDescriptors uint8
	Reserved1      uint8
	DescriptorData []byte
}

type MIHostMetadataElementDescriptor struct {
	Type uint8
	Rev  uint8
	Len  uint16
	Val  []byte
}

// HostMetadataElementType constains the constants below
type HostMetadataElementType uint8

// Metadata constants
const (
	OsCtrlNameElementType           HostMetadataElementType = 0x01
	OsDriverNameElementType                                 = 0x02
	OsDriverVersionElementType                              = 0x03
	PreBootCtrlNameElementType                              = 0x04
	PreBootDriverNameElementType                            = 0x05
	PreBootDriverVersionElementType                         = 0x06

	OsNamespaceNameElementType      HostMetadataElementType = 0x01
	PreBootNamespaceNameElementType                         = 0x02
)

type MiMeatadataFeatureBuilder struct {
	data [4096]byte

	offset int
}

func NewMiFeatureBuilder() *MiMeatadataFeatureBuilder {
	return &MiMeatadataFeatureBuilder{offset: 2}
}

func (builder *MiMeatadataFeatureBuilder) AddElement(typ uint8, rev uint8, data []byte) *MiMeatadataFeatureBuilder {

	// Increment number of elements
	builder.data[0] = builder.data[0] + 1

	offset := builder.offset
	builder.data[offset] = typ
	offset++
	builder.data[offset] = rev
	offset++

	binary.LittleEndian.PutUint16(builder.data[offset:], uint16(len(data)))
	offset += 2

	copy(builder.data[offset:], data)
	offset += len(data)

	builder.offset = offset

	return builder
}

func (builder *MiMeatadataFeatureBuilder) Bytes() []byte {
	return builder.data[:builder.offset]
}

// SmartLog health information
// Get Log Page - SMART / Health Information Log
// Figure 196 from NVM-Express 1_4a-2020.03.09-Ratified specification
type SmartLog struct {
	CriticalWarning struct {
		SpareCapacity                  uint8 `bitfield:"1"` // Bit 0: If set to ‘1’, then the available spare capacity has fallen below the threshold.
		Temperature                    uint8 `bitfield:"1"` // Bit 1: If set to ‘1’, then a temperature is: a) greater than or equal to an over temerature threshold; or b) less than or equal to an under temperature threshold
		Degraded                       uint8 `bitfield:"1"` // Bit 2: If set to ‘1’, then the NVM subsystem reliability has been degraded due to significant media related errors or any internal error that degrades NVM subsystem reliability.
		ReadOnly                       uint8 `bitfield:"1"` // Bit 3: If set to ‘1’, then the media has been placed in read only mode.
		BackupFailed                   uint8 `bitfield:"1"` // Bit 4: If set to ‘1’, then the volatile memory backup device has failed.
		PersistentMemoryRegionReadOnly uint8 `bitfield:"1"` // If set to ‘1’, then the Persistent Memory Region has become read-only or unreliable
		Reserved                       uint8 `bitfield:"2"`
	}
	CompositeTemperature                         uint16 // Contains a value corresponding to a temperature in degrees Kelvin that represents the current composite temperature of the controller and namespace(s) associated with that controller.
	AvailableSpare                               uint8  // Contains a normalized percentage (0% to 100%) of the remaining spare capacity available.
	AvailableSpareThreshold                      uint8  // When the Available Spare falls below the threshold indicated in this field, an asynchronous event completion may occur. The value is indicated as a normalized percentage (0% to 100%). The values 101 to 255 are reserved.
	PercentageUsed                               uint8  // Contains a vendor specific estimate of the percentage of NVM subsystem life used based on the actual usage and the manufacturer’s prediction of NVM life.
	EnduranceGroupCriticalWarningSummary         uint8  // This field indicates critical warnings for the state of Endurance Groups.
	Reserved7                                    [25]uint8
	DataUnitsReadLo                              uint64
	DataUnitsReadHi                              uint64
	DataUnitsWrittenLo                           uint64
	DataUnitsWrittenHi                           uint64
	HostReadsLo                                  uint64
	HostReadsHi                                  uint64
	HostWritesLo                                 uint64
	HostWritesHi                                 uint64
	ControllerBusyTimeLo                         uint64
	ControllerBusyTimeHi                         uint64
	PowerCyclesLo                                uint64
	PowerCyclesHi                                uint64
	PowerOnHoursLo                               uint64
	PowerOnHoursHi                               uint64
	UnsafeShutdownsLo                            uint64
	UnsafeShutdownsHi                            uint64
	MediaErrorsLo                                uint64
	MediaErrorsHi                                uint64
	NumberErrorLogEntriesLo                      uint64
	NumberErrorLogEntriesHi                      uint64
	WarningCompositeTemperatureTime              uint32
	CriticalCompositeTemperatureTime             uint32
	TemperatureSensor                            [8]uint16
	ThermalManagementTemperature1TransitionCount uint32
	ThermalManagementTemperature2TransitionCount uint32
	TotalTimeForThermalManagementTemperature1    uint32
	TotalTimeForThermalManagementTemperature2    uint32
	Reserved232                                  [280]uint8
}

// GetSmartLog - retrieve the smartlog information
func (dev *Device) GetSmartLog() (*SmartLog, error) {

	log := new(SmartLog)

	buf := structex.NewBuffer(log)
	if buf == nil {
		return nil, fmt.Errorf("Cannot allocate buffer")
	}

	if err := dev.getNsidLog(2, 0, 0xFFFFFFFF, buf.Bytes()); err != nil {
		return nil, err
	}

	if err := structex.Decode(buf, log); err != nil {
		return nil, err
	}

	return log, nil
}

// InterpretSmartLog - calculate the drive's swordfish status
func InterpretSmartLog(log *SmartLog) sf.ResourceState {
	// Check critical warnings first - check individual bitfield members
	if log.CriticalWarning.SpareCapacity != 0 { // Spare capacity
		return sf.DISABLED_RST
	}
	if log.CriticalWarning.ReadOnly != 0 { // Read-only mode
		return sf.DISABLED_RST
	}

	// Check other critical warnings for degraded state
	if log.CriticalWarning.Temperature != 0 ||
		log.CriticalWarning.Degraded != 0 ||
		log.CriticalWarning.BackupFailed != 0 ||
		log.CriticalWarning.PersistentMemoryRegionReadOnly != 0 {
		return sf.STANDBY_OFFLINE_RST
	}

	return sf.ENABLED_RST // Healthy
}

func MangleSmartLog(log *SmartLog) {

	// Occaionally managle the smart log data
	if rand.Intn(100) < 10 { // 10% chance of simulating critical condition
		warningType := rand.Intn(6)
		switch warningType {
		case 0:
			log.CriticalWarning.SpareCapacity = 1
			log.AvailableSpare = 5 // Below threshold
		case 1:
			log.CriticalWarning.Temperature = 1
			log.CompositeTemperature = 358 // ~85°C
		case 2:
			log.CriticalWarning.Degraded = 1
		case 3:
			log.CriticalWarning.ReadOnly = 1
		case 4:
			log.CriticalWarning.BackupFailed = 1
		case 5:
			log.CriticalWarning.PersistentMemoryRegionReadOnly = 1
		}
	}

}

// Log contstants
const (
	LogCdw10LogPageIdentiferMask         = 0xFF
	LogCdw10LogPageIdentiferShift        = 0
	LogCdw10LogSpecificFieldMask         = 0x0f
	LogCdw10LogSpecificFieldShift        = 8
	LogCdw10RetainAsynchronousEventMask  = 0x1
	LogCdw10RetainAsynchronousEventShift = 15
	LogCdw10NumberOfDwordsLowerMask      = 0xFFFF
	LogCdw10NumberOfDwordsLowerShift     = 16

	LogCdw11NumberOfDwordsUpperMask    = 0xFFFF
	LogCdw11NumberOfDwordsUpperShift   = 0
	LogCdw11LogSpecificIdentifierMask  = 0xFF
	LogCdw11LogSpecificIdentifierShift = 16

	LogCdw14UuidMask  = 0x7F
	LogCdw14UuidShift = 0
)

func (dev *Device) getNsidLog(logPageIdentifier uint8, retainAsynchronousEvent uint8, nsid uint32, buf []byte) error {
	return dev.getLog(logPageIdentifier, 0, retainAsynchronousEvent, nsid, 0, buf)
}

func (dev *Device) getLog(logPageIdentifier uint8, logSpecificField uint8, retainAsynchronousEvent uint8, nsid uint32, logPageOffset uint64, buf []byte) error {

	var numberDwords uint32 = uint32(len(buf)>>2) - 1

	// Get Log Page - Command Dword 10
	var cdw10 uint32 = (((uint32(logPageIdentifier) & LogCdw10LogPageIdentiferMask) << LogCdw10LogPageIdentiferShift) |
		((uint32(logSpecificField) & LogCdw10LogSpecificFieldMask) << LogCdw10LogSpecificFieldShift) |
		((uint32(retainAsynchronousEvent) & LogCdw10RetainAsynchronousEventMask) << LogCdw10RetainAsynchronousEventShift) |
		((uint32(numberDwords) & LogCdw10NumberOfDwordsLowerMask) << LogCdw10NumberOfDwordsLowerShift))

	var cdw11 uint32 = ((((numberDwords >> 16) & LogCdw11NumberOfDwordsUpperMask) << LogCdw11NumberOfDwordsUpperShift) |
		((0 /* Not Supported */ & LogCdw11LogSpecificIdentifierMask) << LogCdw11LogSpecificIdentifierShift))

	var cdw12 uint32 = uint32(logPageOffset & 0xFFFFFFFF)
	var cdw13 uint32 = uint32(logPageOffset >> 32)

	cmd := AdminCmd{
		Opcode:  uint8(GetLogPage),
		NSID:    nsid,
		Cdw10:   cdw10,
		Cdw11:   cdw11,
		Cdw12:   cdw12,
		Cdw13:   cdw13,
		Cdw14:   0, /* Not Supported */
		Addr:    uint64(uintptr(unsafe.Pointer(&buf[0]))),
		DataLen: uint32(len(buf)),
	}

	return dev.ops.submitAdminPassthru(dev, &cmd, buf)

}

// LogSmartLog outputs the SMART log data in a structured way for logging.
func LogSmartLog(log ec.Logger, smartLog *SmartLog, context ...interface{}) {
	if smartLog == nil {
		log.Error(nil, "SMART log is nil", context...)
		return
	}
	log.Info("SMART log data",
		append(context,
			"criticalWarning.spareCapacity", smartLog.CriticalWarning.SpareCapacity,
			"criticalWarning.temperature", smartLog.CriticalWarning.Temperature,
			"criticalWarning.degraded", smartLog.CriticalWarning.Degraded,
			"criticalWarning.readOnly", smartLog.CriticalWarning.ReadOnly,
			"criticalWarning.backupFailed", smartLog.CriticalWarning.BackupFailed,
			"criticalWarning.persistentMemoryRegionReadOnly", smartLog.CriticalWarning.PersistentMemoryRegionReadOnly,
			"availableSpare", smartLog.AvailableSpare,
			"availableSpareThreshold", smartLog.AvailableSpareThreshold,
			"percentageUsed", smartLog.PercentageUsed,
			"compositeTemperature", smartLog.CompositeTemperature,
			"mediaErrorsLo", smartLog.MediaErrorsLo,
			"numberErrorLogEntriesLo", smartLog.NumberErrorLogEntriesLo,
		)...)
}
