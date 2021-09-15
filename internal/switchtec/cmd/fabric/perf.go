package cmd

import (
	"fmt"
	"time"

	"stash.us.cray.com/rabsw/nnf-ec/internal/switchtec/pkg/switchtec"
)

type BandwidthCmd struct {
	Device  string `kong:"arg,required,type='existingFile',env='SWITCHTEC_DEV',help='The switchtec device.'"`
	Time    int    `kong:"optional,short='t',default='5',help='measurement time in seconds'"`
	Details bool   `kong:"optional,short='d',default='false',help='print posted, non-posted and completion results'"`
	Type    string `kong:"optional,short='t',default='raw',enum='raw,payload',help='bandwidth type ${enum}'"`
}

var (
	typeMap = map[string]switchtec.BandwidthType{
		"raw":     switchtec.Raw_BandwidthType,
		"payload": switchtec.Payload_BandwidthType,
	}
)

func (cmd *BandwidthCmd) Run() error {
	dev, err := switchtec.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	if err := dev.BandwidthCounterSetAll(typeMap[cmd.Type]); err != nil {
		return err
	}

	// Resetting the bandwidht counter takes about 1s
	time.Sleep(time.Second)

	// Record the bandwidth counters
	ports, start, err := dev.BandwidthCounterAll(false)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * time.Duration(cmd.Time))

	_, end, err := dev.BandwidthCounterAll(false)
	if err != nil {
		return err
	}

	siSuffixGet := func(val float64) (float64, string) {
		type siSuffix struct {
			magnitude float64
			suffix    string
		}

		siSuffixes := []siSuffix{
			{1e15, "P"},
			{1e12, "T"},
			{1e9, "G"},
			{1e6, "M"},
			{1e3, "k"},
			{1e0, ""},
			{1e-3, "m"},
			{1e-6, "u"},
			{1e-9, "n"},
			{1e-12, "p"},
			{1e-15, "f"},
		}

		for _, s := range siSuffixes {
			if val >= s.magnitude {
				val /= s.magnitude
				return val, s.suffix
			}
		}

		return val, ""
	}

	printBandwidth := func(msg string, timeUs uint64, total uint64) {
		rate := float64(total) / (float64(timeUs) * 1e-6)
		rate, suffix := siSuffixGet(rate)

		fmt.Printf("\t%-8s\t%5.3g %sB/s\n", msg, rate, suffix)
	}

	printPortTitleFunc := func() func(switchtec.PortId) {
		lastPartition := uint8(0xFF)

		return func(port switchtec.PortId) {
			if port.Partition != lastPartition {
				fmt.Printf("Partition %d:\n", port.Partition)
			}
			lastPartition = port.Partition

			dir := "DSP"
			if port.Upstream != 0 {
				dir = "USP"
			}

			fmt.Printf("\tLogical Port ID %d (%s):\n", port.LogPortId, dir)
		}
	}

	printPortTitle := printPortTitleFunc()
	for i, port := range ports {

		printPortTitle(port)

		end[i].Subtract(&start[i])

		elapsedMicroseconds := end[i].TimeInMicroseconds
		egressTotal := end[i].Egress.Total()
		ingressTotal := end[i].Ingress.Total()

		if !cmd.Details {
			printBandwidth("Out:", elapsedMicroseconds, egressTotal)
			printBandwidth("In:", elapsedMicroseconds, ingressTotal)
		} else {
			fmt.Printf("\tOut:\n")
			printBandwidth("  Posted:", elapsedMicroseconds, end[i].Egress.Posted)
			printBandwidth("  Non-Posted:", elapsedMicroseconds, end[i].Egress.NonPosted)
			printBandwidth("  Completion:", elapsedMicroseconds, end[i].Egress.Completion)
			fmt.Printf("\tIn:\n")
			printBandwidth("  Posted:", elapsedMicroseconds, end[i].Ingress.Posted)
			printBandwidth("  Non-Posted:", elapsedMicroseconds, end[i].Ingress.NonPosted)
			printBandwidth("  Completion:", elapsedMicroseconds, end[i].Ingress.Completion)
		}
	}

	return nil
}
