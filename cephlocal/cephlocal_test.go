package cephlocal_test

import (
	"bytes"
	"fmt"

	"github.com/cloudfoundry-incubator/cephdriver/cephlocal"
	"github.com/cloudfoundry-incubator/cephdriver/cephlocal/cephfakes"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("cephlocal", func() {

	var (
		driver         voldriver.Driver
		fakeInvoker    *cephfakes.FakeInvoker
		fakeSystemUtil *cephfakes.FakeSystemUtil
		testLogger     lager.Logger
	)

	BeforeEach(func() {
		fakeInvoker = new(cephfakes.FakeInvoker)
		fakeSystemUtil = new(cephfakes.FakeSystemUtil)
		driver = cephlocal.NewLocalDriverWithInvokerAndSystemUtil(fakeInvoker, fakeSystemUtil)
		testLogger = lagertest.NewTestLogger("CephdriverTest")
	})

	Describe("Create and Get", func() {

		var (
			createResponse voldriver.ErrorResponse
			opts           map[string]interface{}
		)
		Context("when creating a volume", func() {
			Context("when successful", func() {
				BeforeEach(func() {
					opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "localMountPoint": "some-localmountpoint", "remoteMountPoint": "some-remote-mountpoint"}
					createSuccessful(testLogger, driver, "some-volume-name", opts)
				})
				It("should be able to retrieve volume", func() {
					getResponse := getSuccessful(testLogger, driver, "some-volume-name")
					Expect(getResponse.Volume.Mountpoint).To(Equal(""))
				})

			})
			Context("when unsuccessful", func() {
				Context("when missing opts params", func() {
					BeforeEach(func() {
						opts = map[string]interface{}{}
					})
					It("should error with missing remoteMountPoint", func() {
						opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "localMountPoint": "some-localmountpoint"}
						createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
						createResponse = driver.Create(testLogger, createRequest)
						Expect(createResponse.Err).To(Equal("Missing mandatory 'remoteMountPoint' field in 'Opts'"))
					})
					It("should error with missing localMountPoint", func() {
						opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "remoteMountPoint": "some-remotemountpoint"}
						createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
						createResponse = driver.Create(testLogger, createRequest)
						Expect(createResponse.Err).To(Equal("Missing mandatory 'localMountPoint' field in 'Opts'"))
					})
					It("should error with missing keyring", func() {
						opts = map[string]interface{}{"ip": "some-ip", "remoteMountPoint": "some-remotemountpoint", "localMountPoint": "some-localmoutnpoint"}
						createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
						createResponse = driver.Create(testLogger, createRequest)
						Expect(createResponse.Err).To(Equal("Missing mandatory 'keyring' field in 'Opts'"))
					})
					It("should error with missing ip", func() {
						opts = map[string]interface{}{"keyring": "some-keyring", "remoteMountPoint": "some-remotemountpoint", "localMountPoint": "some-localmoutnpoint"}
						createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
						createResponse = driver.Create(testLogger, createRequest)
						Expect(createResponse.Err).To(Equal("Missing mandatory 'ip' field in 'Opts'"))
					})
					It("should not be able to retrieve volume", func() {
						getUnsuccessful(testLogger, driver, "some-volume-name")
					})
				})
			})
			Context("when volume already exists", func() {
				BeforeEach(func() {
					opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "localMountPoint": "some-localmountpoint", "remoteMountPoint": "some-remote-mountpoint"}
					createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
					createResponse = driver.Create(testLogger, createRequest)
					Expect(createResponse.Err).To(Equal(""))
				})
				It("fails when given different metadata.", func() {
					opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "localMountPoint": "some-localmountpoint", "remoteMountPoint": "someother-remote-mountpoint"}
					createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
					createResponse = driver.Create(testLogger, createRequest)
					Expect(createResponse.Err).To(Equal("Volume 'some-volume-name' already exists with different Opts"))
				})
				It("succeeds when given same metadata", func() {
					opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "localMountPoint": "some-localmountpoint", "remoteMountPoint": "some-remote-mountpoint"}
					createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
					createResponse = driver.Create(testLogger, createRequest)
					Expect(createResponse.Err).To(Equal(""))
				})
			})
		})

	})

	Describe(".Mount", func() {
		var (
			mountResponse  voldriver.MountResponse
			volumeName     string
			opts           map[string]interface{}
		)
		Context("when there is a created/attached volume", func() {
			BeforeEach(func() {
				volumeName = "volume-name"
				opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "localMountPoint": "some-localmountpoint", "remoteMountPoint": "some-remote-mountpoint"}
				createSuccessful(testLogger, driver, volumeName, opts)
			})

			It("should report an error for volume name mismatch", func() {
				mountRequest := voldriver.MountRequest{Name: "garbage"}
				mountResponse = driver.Mount(testLogger, mountRequest)
				Expect(mountResponse.Err).To(Equal("Volume 'garbage' not found"))
				Expect(mountResponse.Mountpoint).To(Equal(""))
			})

			It("should report an error if keyfile creation errors", func() {
				fakeSystemUtil.WriteFileReturns(fmt.Errorf("error writing file"))
				mountRequest := voldriver.MountRequest{Name: volumeName}
				mountResponse = driver.Mount(testLogger, mountRequest)
				Expect(mountResponse.Err).To(Equal(fmt.Sprintf("Error mounting '%s' (error writing file)", volumeName)))
			})

			It("should report an error if CLI invocation fails", func() {
				fakeInvoker.InvokeReturns(fmt.Errorf("invocation fails"))
				mountRequest := voldriver.MountRequest{Name: volumeName}
				mountResponse = driver.Mount(testLogger, mountRequest)
				Expect(mountResponse.Err).To(Equal(fmt.Sprintf("Error mounting '%s' (invocation fails)", volumeName)))
			})

			Context("when the mount completes successfully", func() {
				BeforeEach(func() {
					fakeInvoker.InvokeReturns(nil)
					mountSuccessful(testLogger, driver, volumeName)
				})
				It("creates a keyfile", func() {
					Expect(fakeSystemUtil.WriteFileCallCount()).To(Equal(1))
				})
				It("can get the volume and it is mounted path", func() {
					getResponse := getSuccessful(testLogger, driver, volumeName)
					Expect(getResponse.Volume.Mountpoint).To(Equal("some-localmountpoint"))
				})
				It("should return mountpoint", func() {

					mountResponse = driver.Mount(testLogger, voldriver.MountRequest{
						Name: volumeName,
					})

					Expect(mountResponse.Mountpoint).To(Equal("some-localmountpoint"))
					By("not calling ceph executable again.")
					Expect(fakeInvoker.InvokeCallCount()).To(Equal(1))
				})
			})
		})
	})

	Describe(".Unmount", func() {
		var (
			unmountResponse voldriver.ErrorResponse
			volumeName      string
			opts            map[string]interface{}
		)
		Context("when there is a created/attached volume", func() {
			BeforeEach(func() {
				volumeName = "volume-name"
				opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "localMountPoint": "some-localmountpoint", "remoteMountPoint": "some-remote-mountpoint"}
				createSuccessful(testLogger, driver, volumeName, opts)
			})

			It("should report an error for volume name mismatch", func() {
				unmountRequest := voldriver.UnmountRequest{Name: "garbage"}
				unmountResponse = driver.Unmount(testLogger, unmountRequest)
				Expect(unmountResponse.Err).To(Equal("Volume 'garbage' is unknown"))
			})

			It("should error when volume is not mounted", func() {
				unmountRequest := voldriver.UnmountRequest{Name: volumeName}
				unmountResponse = driver.Unmount(testLogger, unmountRequest)

				Expect(unmountResponse.Err).To(Equal(fmt.Sprintf("Volume '%s' not mounted", volumeName)))
			})
			It("should error when volume is not created", func() {
				unmountRequest := voldriver.UnmountRequest{Name: "non-existent-volume"}
				unmountResponse = driver.Unmount(testLogger, unmountRequest)
				Expect(unmountResponse.Err).To(Equal(fmt.Sprintf("Volume '%s' is unknown", "non-existent-volume")))
			})

			Context("when volume mounted", func() {
				BeforeEach(func() {
					mountSuccessful(testLogger, driver, volumeName)
				})
				It("should report an error if remove config file fails", func() {
					fakeSystemUtil.RemoveReturns(fmt.Errorf("file deletion failed"))
					unmountRequest := voldriver.UnmountRequest{Name: volumeName}
					unmountResponse = driver.Unmount(testLogger, unmountRequest)
					Expect(unmountResponse.Err).To(Equal(fmt.Sprintf("Error unmounting '%s' (file deletion failed)", volumeName)))
				})
				It("should report an error if CLI invocation fails", func() {
					fakeInvoker.InvokeReturns(fmt.Errorf("invocation fails"))
					unmountRequest := voldriver.UnmountRequest{Name: volumeName}
					unmountResponse = driver.Unmount(testLogger, unmountRequest)
					Expect(unmountResponse.Err).To(Equal(fmt.Sprintf("Error unmounting '%s' (invocation fails)", volumeName)))
				})

				Context("when umount successful", func() {
					BeforeEach(func() {
						fakeInvoker.InvokeReturns(nil)
						unmountSuccessful(testLogger, driver, volumeName)
					})
					It("only gets volume name, without Mountpoint", func() {
						getResponse := getSuccessful(testLogger, driver, volumeName)
						Expect(getResponse.Volume.Mountpoint).To(Equal(""))
					})
					It("removes keyfile", func() {
						Expect(fakeSystemUtil.RemoveCallCount()).To(Equal(1))
					})
				})
			})
		})
	})

	Describe(".Remove", func() {
		const volumeName = "volume-name"
		var  opts            map[string]interface{}

		It("should fail if no volume name provided", func() {
			removeResponse := driver.Remove(testLogger, voldriver.RemoveRequest{
				Name: "",
			})
			Expect(removeResponse.Err).To(Equal("Missing mandatory 'volume_name'"))
		})

		It("should fail if no volume was created", func() {
			removeResponse := driver.Remove(testLogger, voldriver.RemoveRequest{
				Name: volumeName,
			})
			Expect(removeResponse.Err).To(Equal("Volume 'volume-name' not found"))
		})

		Context("when there is a created/attached volume", func() {
			BeforeEach(func() {
				opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "localMountPoint": "some-localmountpoint", "remoteMountPoint": "some-remote-mountpoint"}
				createSuccessful(testLogger, driver, volumeName, opts)
			})
			It("destroys volume", func() {
				removeResponse := driver.Remove(testLogger, voldriver.RemoveRequest{
					Name: volumeName,
				})
				Expect(removeResponse.Err).To(Equal(""))
				getUnsuccessful(testLogger, driver, volumeName)
			})

			Context("when volume mounted", func() {
				BeforeEach(func() {
					mountSuccessful(testLogger, driver, volumeName)
				})
				It("unmounts and destroys volume", func() {
					removeResponse := driver.Remove(testLogger, voldriver.RemoveRequest{
						Name: volumeName,
					})
					Expect(removeResponse.Err).To(Equal(""))
					getUnsuccessful(testLogger, driver, volumeName)
					Expect(fakeSystemUtil.RemoveCallCount()).To(Equal(1))
				})
				Context("when unmount fails",func(){
					BeforeEach(func(){
						fakeInvoker.InvokeReturns(fmt.Errorf("invocation fails"))
					})
					It("returns error",func(){
						removeResponse := driver.Remove(testLogger, voldriver.RemoveRequest{
							Name: volumeName,
						})
						Expect(removeResponse.Err).To(Equal("Error unmounting '"+ volumeName+ "' (invocation fails)"))
					})

				})
			})
		})
	})
	
})

var _ = Describe("RealInvoker", func() {
	var (
		subject    cephlocal.Invoker
		fakeCmd    *volmanfakes.FakeCmd
		fakeExec   *volmanfakes.FakeExec
		testLogger = lagertest.NewTestLogger("InvokerTest")
		cmd        = "some-fake-command"
		args       = []string{"fake-args-1"}
	)
	Context("when invoking an executable", func() {
		BeforeEach(func() {
			fakeExec = new(volmanfakes.FakeExec)
			fakeCmd = new(volmanfakes.FakeCmd)
			fakeExec.CommandReturns(fakeCmd)
			subject = cephlocal.NewRealInvokerWithExec(fakeExec)
		})

		It("should report an error when unable to attach to stdout", func() {
			fakeCmd.StdoutPipeReturns(errCloser{bytes.NewBufferString("")}, fmt.Errorf("unable to attach to stdout"))
			err := subject.Invoke(testLogger, cmd, args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unable to attach to stdout"))
		})

		It("should report an error when unable to start binary", func() {
			fakeCmd.StdoutPipeReturns(errCloser{bytes.NewBufferString("cmdfails")}, nil)
			fakeCmd.StartReturns(fmt.Errorf("unable to start binary"))
			err := subject.Invoke(testLogger, cmd, args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unable to start binary"))
		})
		It("should report an error when executing the driver binary fails", func() {
			fakeCmd.WaitReturns(fmt.Errorf("executing driver binary fails"))

			err := subject.Invoke(testLogger, cmd, args)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("executing driver binary fails"))
		})
		It("should successfully invoke cli", func() {
			err := subject.Invoke(testLogger, cmd, args)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

func createSuccessful(logger lager.Logger, driver voldriver.Driver, volumeName string, opts map[string]interface{}){
	createRequest := voldriver.CreateRequest{Name: volumeName, Opts: opts}
	createResponse := driver.Create(logger, createRequest)
	Expect(createResponse.Err).To(Equal(""))
}

func getUnsuccessful(logger lager.Logger, localDriver voldriver.Driver, volumeName string){
	getResponse := localDriver.Get(logger, voldriver.GetRequest{
		Name: volumeName,
	})

	Expect(getResponse.Err).To(Equal("Volume '" + volumeName + "' not found"))
	Expect(getResponse.Volume.Name).To(Equal(""))
}

func getSuccessful(logger lager.Logger, localDriver voldriver.Driver, volumeName string) voldriver.GetResponse {
	getResponse := localDriver.Get(logger, voldriver.GetRequest{
		Name: volumeName,
	})

	Expect(getResponse.Err).To(Equal(""))
	Expect(getResponse.Volume.Name).To(Equal(volumeName))
	return getResponse
}

func mountSuccessful(logger lager.Logger, localDriver voldriver.Driver, volumeName string){
	mountResponse := localDriver.Mount(logger, voldriver.MountRequest{
		Name: volumeName,
	})
	Expect(mountResponse.Err).To(Equal(""))
	Expect(mountResponse.Mountpoint).To(Equal("some-localmountpoint"))
}

func unmountSuccessful(logger lager.Logger, localDriver voldriver.Driver, volumeName string) {
	unmountResponse := localDriver.Unmount(logger, voldriver.UnmountRequest{
		Name: volumeName,
	})
	Expect(unmountResponse.Err).To(Equal(""))
}