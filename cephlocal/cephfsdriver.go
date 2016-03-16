package cephlocal

import (
	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/lager"
)

type LocalDriver struct { // see voldriver.resources.go
	rootDir string
	logFile string
	useExec system.Exec
}

func NewLocalDriver() *LocalDriver {
	return &LocalDriver{"_cephdriver/", "/tmp/cephdriver.log", &system.SystemExec{}}
}

func NewLocalDriverWithExec(useExec system.Exec) *LocalDriver {
	return &LocalDriver{"_cephdriver/", "/tmp/cephdriver.log", useExec}
}

func (d *LocalDriver) Info(logger lager.Logger) (voldriver.InfoResponse, error) {
	return voldriver.InfoResponse{
		Name: "cephdriver",
		Path: "/fake/path",
	}, nil
}

func (d *LocalDriver) Mount(logger lager.Logger, mountRequest voldriver.MountRequest) voldriver.MountResponse {
	return voldriver.MountResponse{}
}

func (d *LocalDriver) Unmount(logger lager.Logger, unmountRequest voldriver.UnmountRequest) voldriver.ErrorResponse {
	return voldriver.ErrorResponse{}
}

func (d *LocalDriver) Create(logger lager.Logger, createRequest voldriver.CreateRequest) voldriver.ErrorResponse {
	return voldriver.ErrorResponse{}
}

func (d *LocalDriver) Get(logger lager.Logger, getRequest voldriver.GetRequest) voldriver.GetResponse {
	return voldriver.GetResponse{}
}

// func (d *LocalDriver) Mount(logger lager.Logger, mountRequest voldriver.MountRequest) (voldriver.MountResponse, error) {

// 	f, _ := d.openLog()
// 	defer f.Close()

// 	mountPath := os.TempDir() + d.rootDir + mountRequest.VolumeId
// 	d.writeLog(f, "Mounting volume %s", mountRequest.VolumeId)
// 	d.writeLog(f, "Creating volume path %s", mountPath)
// 	cmd := "ceph-fuse"
// 	var config cephdriver.MountConfig
// 	//logger.Info("before unmarshall ")
// 	err := json.Unmarshal([]byte(mountRequest.Config), &config)
// 	//logger.Info("after unmarshall")
// 	if err != nil {
// 		panic("json parsing error: config cannot be parsed")
// 	}
// 	content := []byte(config.Keyring)

// 	ioutil.WriteFile("/tmp/keyring", content, 0777)
// 	//logger.Info("after writing keyring file")
// 	cmdArgs := []string{"-k", "/tmp/keyring", "-m", fmt.Sprintf("%s:6789", config.IP), config.MountPoint}
// 	cmdHandle := d.useExec.Command(cmd, cmdArgs...)

// 	_, err = cmdHandle.StdoutPipe()
// 	if err != nil {
// 		d.writeLog(f, "unable to get stdout")
// 		return voldriver.MountResponse{}, fmt.Errorf(fmt.Sprintf("unable to get stdoutv"), err)
// 	}

// 	if err = cmdHandle.Start(); err != nil {
// 		d.writeLog(f, "starting command")
// 		return voldriver.MountResponse{}, fmt.Errorf(fmt.Sprintf("starting command"), err)
// 	}

// 	if err = cmdHandle.Wait(); err != nil {
// 		d.writeLog(f, "waiting for command")
// 		return voldriver.MountResponse{}, fmt.Errorf(fmt.Sprintf("waiting for command"), err)
// 	}

// 	mountPoint := voldriver.MountResponse{config.MountPoint}

// 	return mountPoint, nil
// }

// func (d *localDriver) Unmount(logger lager.Logger, unmountRequest voldriver.UnmountRequest) error {

// 	f, _ := d.openLog()
// 	defer f.Close()

// 	mountPath := os.TempDir() + d.rootDir + unmountRequest.VolumeId
// 	exists, err := exists(mountPath)
// 	if err != nil {
// 		d.writeLog(f, "Error establishing if volume exists")
// 		return fmt.Errorf("Error establishing if volume exists")
// 	}
// 	if !exists {
// 		d.writeLog(f, "Volume %s does not exist, nothing to do!", unmountRequest.VolumeId)
// 		return fmt.Errorf("Volume %s does not exist, nothing to do!", unmountRequest.VolumeId)
// 	} else {
// 		d.writeLog(f, "Removing volume path %s", mountPath)
// 		err := os.RemoveAll(mountPath)
// 		if err != nil {
// 			d.writeLog(f, "Unexpected error removing mount path %s", unmountRequest.VolumeId)
// 			return fmt.Errorf("Unexpected error removing mount path %s", unmountRequest.VolumeId)
// 		}
// 		d.writeLog(f, "Unmounted volume %s", unmountRequest.VolumeId)
// 	}
// 	return nil
// }

// func (d *LocalDriver) openLog() (*os.File, error) {
// 	f, err := os.OpenFile(d.logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
// 	if err != nil {
// 		panic(fmt.Sprintf("Can't create cephdriver log file %s", d.logFile))
// 	}
// 	return f, nil
// }

// func (d *LocalDriver) writeLog(f *os.File, msg string, args ...string) error {
// 	t := time.Now()
// 	_, err := f.WriteString(fmt.Sprintf("[%s] "+msg+"\n", t.Format(time.RFC3339), args))
// 	return err
// }
