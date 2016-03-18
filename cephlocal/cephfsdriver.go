package cephlocal

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/pivotal-golang/lager"
)

type LocalDriver struct { // see voldriver.resources.go
	rootDir       string
	logFile       string
	volumes       map[string]*volumeMetadata
	useInvoker    Invoker
	useSystemUtil SystemUtil
}
type volumeMetadata struct {
	Keyring          string
	IP               string
	RemoteMountPoint string
	LocalMountPoint  string
	Mounted          bool
	KeyPath          string
}

func (v *volumeMetadata) equals(volume *volumeMetadata) bool {
	return volume.LocalMountPoint == v.LocalMountPoint && volume.RemoteMountPoint == v.RemoteMountPoint && volume.Keyring == v.Keyring && volume.IP == v.IP
}

func NewLocalDriver() *LocalDriver {
	return NewLocalDriverWithInvokerAndSystemUtil(NewRealInvoker(), NewRealSystemUtil())
}

func NewLocalDriverWithInvokerAndSystemUtil(invoker Invoker, systemUtil SystemUtil) *LocalDriver {
	return &LocalDriver{"_cephdriver/", "/tmp/cephdriver.log", map[string]*volumeMetadata{}, invoker, systemUtil}
}

func (d *LocalDriver) Info(logger lager.Logger) (voldriver.InfoResponse, error) {
	return voldriver.InfoResponse{
		Name: "cephdriver",
		Path: "/fake/path",
	}, nil
}

func (d *LocalDriver) Create(logger lager.Logger, createRequest voldriver.CreateRequest) voldriver.ErrorResponse {
	logger = logger.Session("create", lager.Data{"request": createRequest})
	logger.Info("start")
	defer logger.Info("end")

	var (
		localmountpoint  string
		remotemountpoint string
		ip               string
		keyring          string
		err              *voldriver.ErrorResponse
	)

	ip, err = extractValue(logger, "ip", createRequest.Opts)
	if err != nil {
		return *err
	}

	keyring, err = extractValue(logger, "keyring", createRequest.Opts)
	if err != nil {
		return *err
	}

	remotemountpoint, err = extractValue(logger, "remoteMountPoint", createRequest.Opts)
	if err != nil {
		return *err
	}

	localmountpoint, err = extractValue(logger, "localMountPoint", createRequest.Opts)
	if err != nil {
		return *err
	}

	return d.create(logger, createRequest.Name, ip, keyring, remotemountpoint, localmountpoint)
}

func successfullResponse() voldriver.ErrorResponse {
	return voldriver.ErrorResponse{}
}

func (d *LocalDriver) create(logger lager.Logger, name, ip, keyring, remotemountpoint, localmountpoint string) voldriver.ErrorResponse {
	var volume *volumeMetadata
	var ok bool

	newVolume := &volumeMetadata{LocalMountPoint: localmountpoint, RemoteMountPoint: remotemountpoint, Keyring: keyring, IP: ip}

	if volume, ok = d.volumes[name]; !ok {
		logger.Info("create-volume", lager.Data{"volume_name": name})
		d.volumes[name] = newVolume
		return successfullResponse()
	}

	if volume.equals(newVolume) {
		logger.Info("duplicate-volume", lager.Data{"volume_name": name})
		return successfullResponse()
	}

	logger.Info("duplicate-volume-with-different-opts", lager.Data{"volume_name": name, "existing-volume": volume})
	return voldriver.ErrorResponse{Err: fmt.Sprintf("Volume '%s' already exists with different Opts", name)}

}

func extractValue(logger lager.Logger, value string, opts map[string]interface{}) (string, *voldriver.ErrorResponse) {
	var aString interface{}
	var str string
	var ok bool
	if aString, ok = opts[value]; !ok {
		logger.Info("missing-" + strings.ToLower(value))
		return "", &voldriver.ErrorResponse{Err: "Missing mandatory '" + value + "' field in 'Opts'"}
	}
	if str, ok = aString.(string); !ok {
		logger.Info("missing-" + strings.ToLower(value))
		return "", &voldriver.ErrorResponse{Err: "Unable to string convert '" + value + "' field in 'Opts'"}
	}
	logger.Info("creating-volume", lager.Data{"value - " + value: str})
	return str, nil
}

func (d *LocalDriver) Get(logger lager.Logger, getRequest voldriver.GetRequest) voldriver.GetResponse {
	logger = logger.Session("Get")
	logger.Info("start")
	defer logger.Info("end")
	if volume, ok := d.volumes[getRequest.Name]; ok {
		logger.Info("get-volume", lager.Data{"volume_name": getRequest.Name})
		if volume.Mounted == true {
			return voldriver.GetResponse{Volume: voldriver.VolumeInfo{Name: getRequest.Name, Mountpoint: volume.LocalMountPoint}}
		}
		return voldriver.GetResponse{Volume: voldriver.VolumeInfo{Name: getRequest.Name}}
	}
	logger.Info("get-volume-not-found", lager.Data{"volume_name": getRequest.Name})
	return voldriver.GetResponse{Err: fmt.Sprintf("Volume '%s' not found", getRequest.Name)}
}

func (d *LocalDriver) Mount(logger lager.Logger, mountRequest voldriver.MountRequest) voldriver.MountResponse {
	logger = logger.Session("Mount")
	logger.Info("start")
	defer logger.Info("end")
	var volume *volumeMetadata
	var ok bool
	if volume, ok = d.volumes[mountRequest.Name]; !ok {

		logger.Info("mount-volume-not-found", lager.Data{"volume_name": mountRequest.Name})
		return voldriver.MountResponse{Err: fmt.Sprintf("Volume '%s' not found", mountRequest.Name)}
	}
	if volume.Mounted == true {
		logger.Info("mount-volume-already-mounted", lager.Data{"volume": volume})
		return voldriver.MountResponse{Mountpoint: volume.LocalMountPoint}
	}
	logger.Info("mounting-volume-"+mountRequest.Name, lager.Data{"volume": volume})
	content := []byte(volume.Keyring)

	volume.KeyPath = "/tmp/keypath_" + string(time.Now().UnixNano())

	err := d.useSystemUtil.WriteFile(volume.KeyPath, content, 0777)
	if err != nil {
		logger.Error("Error mounting volume", err)
		return voldriver.MountResponse{Err: fmt.Sprintf("Error mounting '%s' (%s)", mountRequest.Name, err.Error())}
	}
	cmdArgs := []string{"-k", volume.KeyPath, "-m", fmt.Sprintf("%s:6789", volume.IP), volume.LocalMountPoint}

	if err := d.callCeph(logger, cmdArgs); err != nil {
		logger.Error("Error mounting volume", err)
		return voldriver.MountResponse{Err: fmt.Sprintf("Error mounting '%s' (%s)", mountRequest.Name, err.Error())}
	}

	volume.Mounted = true

	return voldriver.MountResponse{Mountpoint: volume.LocalMountPoint}

}

func (d *LocalDriver) Unmount(logger lager.Logger, unmountRequest voldriver.UnmountRequest) voldriver.ErrorResponse {
	logger = logger.Session("Unmount")
	logger.Info("start")
	defer logger.Info("end")

	var volume *volumeMetadata
	var ok bool
	if volume, ok = d.volumes[unmountRequest.Name]; !ok {
		logger.Info("unmount-volume-not-found", lager.Data{"volume_name": unmountRequest.Name})
		return voldriver.ErrorResponse{Err: fmt.Sprintf("Volume '%s' is unknown", unmountRequest.Name)}
	}
	if volume.Mounted == false {
		logger.Info("unmount-volume-not-mounted", lager.Data{"volume_name": unmountRequest.Name})
		return voldriver.ErrorResponse{Err: fmt.Sprintf("Volume '%s' not mounted", unmountRequest.Name)}
	}
	logger.Info("umount-found-volume", lager.Data{"metadata": volume})
	cmdArgs := []string{volume.LocalMountPoint}
	if err := d.useInvoker.Invoke(logger, "unmount", cmdArgs); err != nil {
		logger.Error("Error invoking CLI", err)
		return voldriver.ErrorResponse{Err: fmt.Sprintf("Error unmounting '%s' (%s)", unmountRequest.Name, err.Error())}
	}
	volume.Mounted = false
	err := d.useSystemUtil.Remove(volume.KeyPath)
	if err != nil {
		logger.Error("Error deleting file", err)
		return voldriver.ErrorResponse{Err: fmt.Sprintf("Error unmounting '%s' (%s)", unmountRequest.Name, err.Error())}
	}
	return voldriver.ErrorResponse{}
}

func (d *LocalDriver) callCeph(logger lager.Logger, args []string) error {
	cmd := "ceph-fuse"
	return d.useInvoker.Invoke(logger, cmd, args)
}

//go:generate counterfeiter -o ./cephfakes/fake_system_util.go . SystemUtil

type SystemUtil interface {
	WriteFile(filename string, data []byte, perm os.FileMode) error
	Remove(string) error
}
type realSystemUtil struct{}

func NewRealSystemUtil() SystemUtil {
	return &realSystemUtil{}
}

func (f *realSystemUtil) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func (f *realSystemUtil) Remove(path string) error {
	return os.Remove(path)
}

//go:generate counterfeiter -o ./cephfakes/fake_invoker.go . Invoker

type Invoker interface {
	Invoke(logger lager.Logger, executable string, args []string) error
}

type realInvoker struct {
	useExec system.Exec
}

func NewRealInvoker() Invoker {
	return NewRealInvokerWithExec(&system.SystemExec{})
}

func NewRealInvokerWithExec(useExec system.Exec) Invoker {
	return &realInvoker{useExec}
}

func (r *realInvoker) Invoke(logger lager.Logger, executable string, cmdArgs []string) error {
	cmdHandle := r.useExec.Command(executable, cmdArgs...)

	_, err := cmdHandle.StdoutPipe()
	if err != nil {
		logger.Error("unable to get stdout", err)
		return err
	}

	if err = cmdHandle.Start(); err != nil {
		logger.Error("starting command", err)
		return err
	}

	if err = cmdHandle.Wait(); err != nil {
		logger.Error("waiting for command", err)
		return err
	}

	// could validate stdout, but defer until actually need it
	return nil
}
