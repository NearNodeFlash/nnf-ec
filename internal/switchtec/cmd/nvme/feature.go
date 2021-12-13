package nvme

import (
	"bufio"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.hpe.com/hpe/hpc-rabsw-nnf-ec/internal/switchtec/pkg/nvme"
)

type GetFeatureCmd struct {
	Device      string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	FeatureId   uint32 `kong:"arg,short='f',required,help='Feature identifier'"`
	NamespaceId uint32 `kong:"arg,short='n',help='Identifier of desired namespace'"`
	Select      int    `kong:"arg,short='s',help='[0-3]: current/default/saved/supported'"`
	Output      string `kong:"arg,short='o',help='Output file to write'"`
}

func (cmd *GetFeatureCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	flen := nvme.FeatureBufferLength[cmd.FeatureId]

	if cmd.Select == 3 {
		flen = 0
	}

	buf := make([]byte, flen)
	if err := dev.GetFeature(cmd.NamespaceId, nvme.Feature(cmd.FeatureId), cmd.Select, 0, flen, buf); err != nil {
		return err
	}

	if err := ioutil.WriteFile(cmd.Output, buf, fs.ModePerm); err != nil {
		return err
	}

	fmt.Println("Success")

	return nil
}

type SetFeatureCmd struct {
	Device      string `kong:"arg,required,type='existingFile',help='The nvme device or device over switchtec tunnel'"`
	FeatureId   string `kong:"arg,short='f',required,help='Feature identifier'"`
	NamespaceId uint32 `kong:"arg,short='n',help='Identifier of desired namespace'"`
	Data        string `kong:"arg,type='existingFile',help='file for feature data'"`
	Save        bool   `kong:"optional,short='s',help='Save data to persistent storage'"`
}

func (cmd *SetFeatureCmd) Run() error {
	dev, err := nvme.Open(cmd.Device)
	if err != nil {
		return err
	}
	defer dev.Close()

	data, err := ioutil.ReadFile(cmd.Data)

	fid, err := strconv.ParseUint(cmd.FeatureId, 0, 32)
	if err != nil {
		return fmt.Errorf("Failed to parse feature id %s", cmd.FeatureId)
	}

	flen := nvme.FeatureBufferLength[fid]
	if len(data) > int(flen) {
		return fmt.Errorf("Data is larger than feature file")
	}

	buf := make([]byte, flen)
	copy(buf[:], data)

	fmt.Printf("Performing Set-Feature %#02x NSID %d of %d bytes\n", fid, cmd.NamespaceId, flen)
	if err := dev.SetFeature(cmd.NamespaceId, nvme.Feature(fid), 0, cmd.Save, flen, buf); err != nil {
		return err
	}

	fmt.Println("Success")

	return nil
}

type BuildMiFeatureCmd struct {
	File string `kong:"arg,required,help='Output file'"`
}

func (cmd *BuildMiFeatureCmd) Run() error {
	reader := bufio.NewReader(os.Stdin)
	builder := nvme.NewMiFeatureBuilder()

	for {
		fmt.Print("Enter Element Type: ")
		ets, _ := reader.ReadString('\n')
		if ets == "\n" {
			break
		}

		fmt.Print("Enter Element Rev: ")
		ers, _ := reader.ReadString('\n')

		fmt.Print("Enter Element Data: ")
		eds, _ := reader.ReadString('\n')

		et, _ := strconv.ParseUint(strings.TrimRight(ets, "\n"), 0, 8)
		er, _ := strconv.ParseUint(strings.TrimRight(ers, "\n"), 0, 8)
		ed := strings.TrimRight(eds, "\n")

		builder.AddElement(uint8(et), uint8(er), []byte(ed))

	}

	fmt.Printf("Writing %d Bytes to %s...\n", len(builder.Bytes()), cmd.File)
	return ioutil.WriteFile(cmd.File, builder.Bytes(), fs.ModePerm)
}
