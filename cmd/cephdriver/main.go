package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry-incubator/cephdriver"
	//cf_lager "github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry-incubator/volman"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	flags "github.com/jessevdk/go-flags"
	"io/ioutil"
	"os"
	"os/exec"
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
	//logger, _ := cf_lager.New("cephdriver")

	//eg: ceph-fuse -k ceph.client.admin.keyring -m $cephfs_ip:6789 ~/mycephfs
	cmd := "ceph-fuse"
	var config cephdriver.MountConfig
	//logger.Info("before unmarshall ")
	err := json.Unmarshal([]byte(args[1]), &config)
	//logger.Info("after unmarshall")
	if err != nil {
		panic("json parsing error: config cannot be parsed")
	}
	content := []byte(config.Keyring)

	ioutil.WriteFile("/tmp/keyring", content, 0777)
	//logger.Info("after writing keyring file")
	cmdArgs := []string{"-k", "/tmp/keyring", "-m", fmt.Sprintf("%s:6789", config.IP), "/tmp/test"}
	cmdHandle := exec.Command(cmd, cmdArgs...)
	var out bytes.Buffer
	cmdHandle.Stdout = &out
	err = cmdHandle.Run()
	if err != nil {
		fmt.Println(os.Stderr)
		fmt.Println(err)
		fmt.Println(out)
	}

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
