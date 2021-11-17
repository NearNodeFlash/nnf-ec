package main

import (
	"github.com/alecthomas/kong"

	ctx "stash.us.cray.com/rabsw/nnf-ec/internal/switchtec/cmd"
	cfg "stash.us.cray.com/rabsw/nnf-ec/internal/switchtec/cmd/config"
	cmd "stash.us.cray.com/rabsw/nnf-ec/internal/switchtec/cmd/fabric"
	nvme "stash.us.cray.com/rabsw/nnf-ec/internal/switchtec/cmd/nvme"
)

var cli struct {
	Debug   bool `kong:"optional,help='Enable debug'"`
	Verbose int  `kong:"optional,hidden,type='counter',help='Debug verbosity level.'"`

	Echo     cmd.EchoCmd      `kong:"cmd,help='Send an echo command to the device.'"`
	Identify cmd.IdentifyCmd  `kong:"cmd,help='Identify the PAX device.'"`
	LinkStat cmd.LinkStatCmd  `kong:"cmd,help='Report link status for a device.'"`
	Bind     cmd.BindCmd      `kong:"cmd,help='Bind an endpoint to a port within a host virtualization domain.'"`
	Unbind   cmd.UnbindCmd    `kong:"cmd,help='Unbind an endpoint on a port within a host virtualization domain.'"`
	Dump     cmd.DumpCmd      `kong:"cmd,help='Dump various types of GFMS information.'"`
	Gas      cmd.GasCmd       `kong:"cmd,help='Global Address Space access commands.'"`
	Mfg      cmd.MfgCmd       `kong:"cmd,help='Manufacturing commands.'"`
	Event    cmd.EventCmd     `kong:"cmd,help='Event commands.'"`
	Bw       cmd.BandwidthCmd `kong:"cmd,help='Measure the traffic bandwidth through each port.'"`

	IdCtrl         nvme.IdCtrlCmd         `kong:"cmd,help='Identify Controller.'"`
	IdNs           nvme.IdNsCmd           `kong:"cmd,help='IdentifyNamespace.'"`
	ListNs         nvme.ListNsCmd         `kong:"cmd,help='List Namespace.'"`
	IdNsCtrls      nvme.IdNamespaceCtrls  `kong:"cmd,help='Identify Controller attached to Namespace.'"`
	CreateNs       nvme.CreateNsCmd       `kong:"cmd,help='Create namespace.'"`
	DeleteNs       nvme.DeleteNsCmd       `kong:"cmd,help='Delete namespace.'"`
	AttachNs       nvme.AttachNsCmd       `kong:"cmd,help='attaches a namespace to the given controller or comma-sep list of controllers.'"`
	DetachNs       nvme.DetachNsCmd       `kong:"cmd,help='detaches a namespace from the given controller or comma-sep list of controllers.'"`
	ListSecondary  nvme.ListSecondaryCmd  `kong:"cmd,help='Show secondary controller list associated with the primary controller of the given device.'"`
	VirtMgmt       nvme.VirtualMgmtCmd    `kong:"cmd,help='Virtualization command supported by primary NVMe controlers.'"`
	GetFeature     nvme.GetFeatureCmd     `kong:"cmd,help='Get feature.'"`
	SetFeature     nvme.SetFeatureCmd     `kong:"cmd,help='Set feature.'"`
	BuildMiFeature nvme.BuildMiFeatureCmd `kong:"cmd,help='Build MI Metadata feature file with interactive terminal'"`

	Config cfg.ConfigCmd `kong:"cmd,help='Configure device.'"`
}

func main() {
	c := kong.Parse(&cli)

	debugLevel := cli.Verbose
	if debugLevel == ctx.Disabled && cli.Debug {
		debugLevel = ctx.Debug
	}

	ctx.ApplyContext(ctx.Context{
		DebugLevel: ctx.DebugLevel(debugLevel),
		LogLevel:   ctx.LogLevel(ctx.Disabled),
	})

	err := c.Run()
	c.FatalIfErrorf(err)
}