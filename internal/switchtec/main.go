/*
 * Copyright 2020, 2021, 2022 Hewlett Packard Enterprise Development LP
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

package main

import (
	"github.com/alecthomas/kong"

	ctx "github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/cmd"
	cfg "github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/cmd/config"
	cmd "github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/cmd/fabric"
	nvme "github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/cmd/nvme"
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
	GetSmartLog    nvme.GetSmartLogCmd    `kong:"cmd,help='Retrieve SMART log for the given device.'"`
	BuildMiFeature nvme.BuildMiFeatureCmd `kong:"cmd,help='Build MI Metadata feature file with interactive terminal'"`
	Discover       nvme.DiscoverCmd       `kong:"cmd,help='Discover NVMe devices by Serial Number, Model Number, or NQN.'"`

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
