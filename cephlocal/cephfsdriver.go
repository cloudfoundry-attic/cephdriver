package cephlocal

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/voldriver"

	"github.com/cloudfoundry/gunk/os_wrap/exec_wrap"
	"code.cloudfoundry.org/goshims/ioutil"
	osshim "code.cloudfoundry.org/goshims/os"
	"os"
)

type LocalDriver struct { // see voldriver.resources.go
	rootDir       string
	logFile       string
	volumes       map[string]*volumeMetadata
	useInvoker    Invoker
	os 						osshim.Os
	ioutil 				ioutilshim.Ioutil
}
type volumeMetadata struct {
	Keyring          string
	IP               string
	RemoteMountPoint string
	LocalMountPoint  string
	MountCount       int
	KeyPath          string
}

func (v *volumeMetadata) equals(volume *volumeMetadata) bool {
	return volume.LocalMountPoint == v.LocalMountPoint && volume.RemoteMountPoint == v.RemoteMountPoint && volume.Keyring == v.Keyring && volume.IP == v.IP
}

func NewLocalDriver() *LocalDriver {
	return NewLocalDriverWithInvokerAndSystemUtil(NewRealInvoker(), &osshim.OsShim{}, &ioutilshim.IoutilShim{})
}

func NewLocalDriverWithInvokerAndSystemUtil(invoker Invoker, os osshim.Os, ioutil ioutilshim.Ioutil) *LocalDriver {
	return &LocalDriver{"_cephdriver/", "/tmp/cephdriver.log", map[string]*volumeMetadata{}, invoker, os, ioutil}
}

func (d *LocalDriver) Create(logger lager.Logger, createRequest voldriver.CreateRequest) voldriver.ErrorResponse {
	logger = logger.Session("create", lager.Data{"request": createRequest})
	logger.Info("start")
	defer logger.Info("end")

	var (
		localMountPoint  string
		remoteMountPoint string
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

	remoteMountPoint, err = extractValue(logger, "remote_mount_point", createRequest.Opts)
	if err != nil {
		return *err
	}

	localMountPoint, err = extractValue(logger, "local_mount_point", createRequest.Opts)
	if err != nil {
		return *err
	}

	return d.create(logger, createRequest.Name, ip, keyring, remoteMountPoint, localMountPoint)
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
		logger.Info("missing-config-value", lager.Data{"key": value})
		return "", &voldriver.ErrorResponse{Err: "Missing mandatory '" + value + "' field in 'Opts'"}
	}
	if str, ok = aString.(string); !ok {
		logger.Info("missing-" + strings.ToLower(value))
		return "", &voldriver.ErrorResponse{Err: "Unable to string convert '" + value + "' field in 'Opts'"}
	}
	return str, nil
}

func (d *LocalDriver) Get(logger lager.Logger, getRequest voldriver.GetRequest) voldriver.GetResponse {
	logger = logger.Session("Get")
	logger.Info("start")
	defer logger.Info("end")
	if volume, ok := d.volumes[getRequest.Name]; ok {
		logger.Info("get-volume", lager.Data{"volume_name": getRequest.Name})
		if volume.MountCount > 0 {
			return voldriver.GetResponse{Volume: voldriver.VolumeInfo{Name: getRequest.Name, Mountpoint: volume.LocalMountPoint}}
		}
		return voldriver.GetResponse{Volume: voldriver.VolumeInfo{Name: getRequest.Name}}
	}
	logger.Info("get-volume-not-found", lager.Data{"volume_name": getRequest.Name})
	return voldriver.GetResponse{Err: fmt.Sprintf("Volume '%s' not found", getRequest.Name)}
}

func (d *LocalDriver) Path(logger lager.Logger, getRequest voldriver.PathRequest) voldriver.PathResponse {
	logger = logger.Session("Path")
	logger.Info("start")
	defer logger.Info("end")

	if volume, ok := d.volumes[getRequest.Name]; ok {
		if volume.MountCount > 0 {
			logger.Info("volume-path", lager.Data{"volume_name": getRequest.Name, "volume_path": volume.LocalMountPoint})
			return voldriver.PathResponse{Mountpoint: volume.LocalMountPoint}
		}
		logger.Info("volume-path-not-mounted", lager.Data{"volume_name": getRequest.Name})
		return voldriver.PathResponse{Err: fmt.Sprintf("Volume %s not mounted", getRequest.Name)}
	}
	logger.Info("volume-path-not-found", lager.Data{"volume_name": getRequest.Name})
	return voldriver.PathResponse{Err: fmt.Sprintf("Volume '%s' not found", getRequest.Name)}
}

func (d *LocalDriver) Activate(logger lager.Logger) voldriver.ActivateResponse {

	return voldriver.ActivateResponse{
		Implements: []string{"VolumeDriver"},
	}
}

func (d *LocalDriver) Capabilities(logger lager.Logger) voldriver.CapabilitiesResponse {
	return voldriver.CapabilitiesResponse{
		Capabilities: voldriver.CapabilityInfo{Scope: "global"},
	}
}

func (d *LocalDriver) List(logger lager.Logger) voldriver.ListResponse {
	listResponse := voldriver.ListResponse{}
	volInfo := voldriver.VolumeInfo{}
	for volumeName, volume := range d.volumes {
		volInfo.Name = volumeName
		if volume.MountCount > 0 {
			volInfo.Mountpoint = volume.LocalMountPoint
		} else {
			volInfo.Mountpoint = ""
		}
		listResponse.Volumes = append(listResponse.Volumes, volInfo)
	}
	listResponse.Err = ""
	return listResponse
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
	if volume.MountCount > 0 {
		volume.MountCount++
		logger.Info("mount-volume-already-mounted", lager.Data{"volume": volume})
		return voldriver.MountResponse{Mountpoint: volume.LocalMountPoint}
	}
	logger.Info("mounting-volume-"+mountRequest.Name, lager.Data{"volume": volume})
	content := []byte(volume.Keyring)

	volume.KeyPath = fmt.Sprintf("/tmp/keypath_%#v", time.Now().UnixNano())

	err := d.ioutil.WriteFile(volume.KeyPath, content, 0777)
	if err != nil {
		logger.Error("Error mounting volume", err)
		return voldriver.MountResponse{Err: fmt.Sprintf("Error mounting '%s' (%s)", mountRequest.Name, err.Error())}
	}

	err = d.os.MkdirAll(volume.LocalMountPoint, os.ModePerm)
	if err != nil {
		logger.Error("failed-creating-localmountpoint", err)
		return voldriver.MountResponse{Err: fmt.Sprintf("Unable to create local mount point for volume '%s'", mountRequest.Name)}
	}

	cmdArgs := []string{"-k", volume.KeyPath, "-m", fmt.Sprintf("%s:6789", volume.IP), "-r", volume.RemoteMountPoint, volume.LocalMountPoint}
	if err := d.callCeph(logger, cmdArgs); err != nil {
		logger.Error("Error mounting volume", err)
		return voldriver.MountResponse{Err: fmt.Sprintf("Error mounting '%s' (%s)", mountRequest.Name, err.Error())}
	}

	volume.MountCount = 1

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
	if volume.MountCount == 0 {
		logger.Info("unmount-volume-not-mounted", lager.Data{"volume_name": unmountRequest.Name})
		return voldriver.ErrorResponse{Err: fmt.Sprintf("Volume '%s' not mounted", unmountRequest.Name)}
	}
	return d.unmount(logger, volume, unmountRequest.Name)
}

func (d *LocalDriver) Remove(logger lager.Logger, removeRequest voldriver.RemoveRequest) voldriver.ErrorResponse {
	logger = logger.Session("remove", lager.Data{"volume": removeRequest})
	logger.Info("start")
	defer logger.Info("end")

	if removeRequest.Name == "" {
		return voldriver.ErrorResponse{Err: "Missing mandatory 'volume_name'"}
	}

	var response voldriver.ErrorResponse
	var vol *volumeMetadata
	var exists bool
	if vol, exists = d.volumes[removeRequest.Name]; !exists {
		logger.Error("failed-volume-removal", fmt.Errorf(fmt.Sprintf("Volume %s not found", removeRequest.Name)))
		return voldriver.ErrorResponse{fmt.Sprintf("Volume '%s' not found", removeRequest.Name)}
	}

	for vol.MountCount > 0 {
		response = d.unmount(logger, vol, removeRequest.Name)
		if response.Err != "" {
			return response
		}
	}

	logger.Info("removing-volume", lager.Data{"name": removeRequest.Name})
	delete(d.volumes, removeRequest.Name)
	return voldriver.ErrorResponse{}
}

func (d *LocalDriver) unmount(logger lager.Logger, volume *volumeMetadata, volumeName string) voldriver.ErrorResponse {
	logger.Info("umount-found-volume", lager.Data{"metadata": volume})

	if volume.MountCount > 1 {
		volume.MountCount--
		logger.Info("unmount-volume-in-use", lager.Data{"metadata": volume})
		return voldriver.ErrorResponse{}
	}

	cmdArgs := []string{"-u", volume.LocalMountPoint}
	if err := d.useInvoker.Invoke(logger, "fusermount", cmdArgs); err != nil {
		logger.Error("Error invoking CLI", err)
		return voldriver.ErrorResponse{Err: fmt.Sprintf("Error unmounting '%s' (%s)", volumeName, err.Error())}
	}
	volume.MountCount = 0
	err := d.os.Remove(volume.KeyPath)
	if err != nil {
		logger.Error("Error deleting file", err)
		return voldriver.ErrorResponse{Err: fmt.Sprintf("Error unmounting '%s' (%s)", volumeName, err.Error())}
	}
	err = d.os.Remove(volume.LocalMountPoint)
	if err != nil {
		logger.Error("Error deleting local mountpoint", err)
		return voldriver.ErrorResponse{Err: fmt.Sprintf("Error unmounting '%s' (%s)", volumeName, err.Error())}
	}
	return voldriver.ErrorResponse{}
}

func (d *LocalDriver) callCeph(logger lager.Logger, args []string) error {
	cmd := "ceph-fuse"
	return d.useInvoker.Invoke(logger, cmd, args)
}

//go:generate counterfeiter -o ./cephfakes/fake_invoker.go . Invoker

type Invoker interface {
	Invoke(logger lager.Logger, executable string, args []string) error
}

type realInvoker struct {
	useExec exec_wrap.Exec
}

func NewRealInvoker() Invoker {
	return NewRealInvokerWithExec(exec_wrap.NewExec())
}

func NewRealInvokerWithExec(useExec exec_wrap.Exec) Invoker {
	return &realInvoker{useExec}
}

func (r *realInvoker) Invoke(logger lager.Logger, executable string, cmdArgs []string) error {
	logger = logger.Session("invoking-command", lager.Data{"executable": executable, "args": cmdArgs})
	logger.Info("start")
	defer logger.Info("end")

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
		logger.Error("command-exited", err)
		return err
	}

	// could validate stdout, but defer until actually need it
	return nil
}
