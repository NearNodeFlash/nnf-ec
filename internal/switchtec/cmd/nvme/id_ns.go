package nvme

import (
	"fmt"

	"stash.us.cray.com/rabsw/nnf-ec/internal/switchtec/pkg/nvme"
)

// IdNsCmd Send an Identify Namespace command to the given device
type IdNsCmd struct {
	Device      string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	NamespaceId int32  `kong:"optional,short='n',default='-1',help='Namespace to identify.'"`
	Present     bool   `kong:"optional,short='p',default='false',help='Return the namespace only if present on the device.'"`
}

// Run will run the Identify Namespace Command.
func (cmd *IdNsCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	ns, err := dev.IdentifyNamespace(uint32(cmd.NamespaceId), cmd.Present)
	if err != nil {
		return err
	}

	fmt.Printf("NVME Identify Namespace %d\n", cmd.NamespaceId)
	fmt.Printf("  %-8s: %-32s : %#016x\n", "NSZE", "Namespace Size", ns.Size)
	fmt.Printf("  %-8s: %-32s : %#016x\n", "NCAP", "Namespace Capacity", ns.Capacity)
	fmt.Printf("  %-8s: %-32s : %#016x\n", "NUSE", "Namespace Utilization", ns.Utilization)
	fmt.Printf("  %-8s: %-32s :\n", "NUSE", "Features")

	// TODO: More details

	fmt.Printf("  %-8s: %-32s : 0x%s", "NGUID", "Namespace GUID\n", ns.GloballyUniqueIdentifier.String())
	fmt.Printf("  %-8s: %-32s : %-2d\n", "NLBAS", "Number LBA Formats", ns.NumberOfLBAFormats)
	for i := 0; i < int(ns.NumberOfLBAFormats); i++ {
		f := &ns.LBAFormats[i]
		rp := f.RelativePerformance
		rpStr := map[uint8]string{3: "Degraded", 2: "Good", 1: "Better", 0: "Best"}
		inUse := ""
		if i == int(ns.FormattedLBASize.Format) {
			inUse = "(in use)"
		}

		fmt.Printf("  %-8s  %-2d: Data Size: %4dB - Metadata Size: %dB - Relative Performance: %#x %s %s\n",
			"", i, 1<<f.LBADataSize, f.MetadataSize, rp, rpStr[rp], inUse)
	}

	return nil
}