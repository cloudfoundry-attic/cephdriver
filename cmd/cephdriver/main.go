package main

import (
	"encoding/json"
	"fmt"
	_"io/ioutil"
	"os"
	"github.com/cloudfoundry-incubator/cephdriver"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman"
	flags "github.com/jessevdk/go-flags"
	_"os/exec"
)

type InfoCommand struct {
	Info func() `short:"i" long:"info" description:"Print program information"`
}

func (x *InfoCommand) Execute(args []string) error {
	InfoResponse := voldriver.InfoResponse{
		Name: "cephdriver",
		Path: "/fake/path",
	}

	jsonBlob, err := json.Marshal(InfoResponse)
	if err != nil {
		panic("Error Marshaling the driver")
	}
	fmt.Println(string(jsonBlob))

	return nil
}



type MountCommand struct {
	Mount func() `short:"m" long:"mount" description:"Mount a volume Id to a path"`
}

func (x *MountCommand) Execute(args []string) error {

	//eg: ceph-fuse -k ceph.client.admin.keyring -m $cephfs_ip:6789 ~/mycephfs
	//cmd := "ceph-fuse"
	// need to deserialize config object to get a)keyring file name, b)ceph ip, c)mount point
	//volumeId := args[0]
	var config cephdriver.MountConfig
	err := json.Unmarshal([]byte(args[1]),&config)
	if err == nil{
		panic("json parsing error: config cannot be parsed")
	}
	cmdArgs := []string{"-k",config.Keyring,"-m" ,fmt.Sprintf("%s:6789",config.IP), config.MountPoint}
	// if err := exec.Command(cmd, cmdArgs...).Run(); err != nil {
	// 	fmt.Fprintln(os.Stderr, err)
	// 	panic("Error mounting")
	// }
	fmt.Println(cmdArgs)

	mountPoint := volman.MountResponse{config.MountPoint}

	jsonBlob, err := json.Marshal(mountPoint)
	if err != nil {
		panic("Error Marshaling the mount point")
	}
	fmt.Println(string(jsonBlob))

	return nil
}

type Options struct{}

func main() {
	var infoCmd InfoCommand
	var mountCmd MountCommand
	var options Options
	var parser = flags.NewParser(&options, flags.Default)

	parser.AddCommand("info",
		"Print Info",
		"The info command print the driver name and version.",
		&infoCmd)
	parser.AddCommand("mount",
		"Mount Volume",
		"Mount a volume Id to a path - returning the path.",
		&mountCmd)
	_, err := parser.Parse()

	if err != nil {
		panic(err)
		os.Exit(1)
	}
}
