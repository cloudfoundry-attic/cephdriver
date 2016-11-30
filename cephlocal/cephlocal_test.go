package cephlocal_test

import (
	"fmt"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"

	"code.cloudfoundry.org/cephdriver/cephlocal"
	"code.cloudfoundry.org/voldriver"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"context"

	"code.cloudfoundry.org/goshims/ioutilshim/ioutil_fake"
	"code.cloudfoundry.org/goshims/osshim/os_fake"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"code.cloudfoundry.org/voldriver/voldriverfakes"
)

var _ = Describe("cephlocal", func() {

	var (
		driver      voldriver.Driver
		fakeInvoker *voldriverfakes.FakeInvoker
		fakeOs      *os_fake.FakeOs
		fakeIoutil  *ioutil_fake.FakeIoutil
		testLogger  lager.Logger
		testCtx     context.Context
		testEnv     voldriver.Env
	)

	BeforeEach(func() {
		fakeInvoker = new(voldriverfakes.FakeInvoker)
		fakeOs = new(os_fake.FakeOs)
		fakeIoutil = new(ioutil_fake.FakeIoutil)
		driver = cephlocal.NewLocalDriverWithInvokerAndSystemUtil(fakeInvoker, fakeOs, fakeIoutil)
		testLogger = lagertest.NewTestLogger("CephdriverTest")
		testCtx = context.TODO()
		testEnv = driverhttp.NewHttpDriverEnv(testLogger, testCtx)
	})

	Describe(".Activate", func() {

		It("should return VolumeDrivers json", func() {
			response := driver.Activate(testEnv)
			Expect(len(response.Implements)).To(BeNumerically(">", 0))
			Expect(response.Implements[0]).To(Equal("VolumeDriver"))
		})

	})

	Describe("Create and Get", func() {

		var (
			createResponse voldriver.ErrorResponse
			opts           map[string]interface{}
		)
		Context("when creating a volume", func() {
			Context("when successful", func() {
				BeforeEach(func() {
					opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "local_mount_point": "some-localmountpoint", "remote_mount_point": "some-remote-mountpoint"}
					createSuccessful(testEnv, driver, "some-volume-name", opts)
				})
				It("should be able to retrieve volume", func() {
					getResponse := getSuccessful(testEnv, driver, "some-volume-name")
					Expect(getResponse.Volume.Mountpoint).To(Equal(""))
				})

			})
			Context("when unsuccessful", func() {
				Context("when missing opts params", func() {
					BeforeEach(func() {
						opts = map[string]interface{}{}
					})
					It("should error with missing remote_mount_point", func() {
						opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "local_mount_point": "some-localmountpoint"}
						createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
						createResponse = driver.Create(testEnv, createRequest)
						Expect(createResponse.Err).To(Equal("Missing mandatory 'remote_mount_point' field in 'Opts'"))
					})
					It("should error with missing local_mount_point", func() {
						opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "remote_mount_point": "some-remotemountpoint"}
						createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
						createResponse = driver.Create(testEnv, createRequest)
						Expect(createResponse.Err).To(Equal("Missing mandatory 'local_mount_point' field in 'Opts'"))
					})
					It("should error with missing keyring", func() {
						opts = map[string]interface{}{"ip": "some-ip", "remote_mount_point": "some-remotemountpoint", "local_mount_point": "some-localmoutnpoint"}
						createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
						createResponse = driver.Create(testEnv, createRequest)
						Expect(createResponse.Err).To(Equal("Missing mandatory 'keyring' field in 'Opts'"))
					})
					It("should error with missing ip", func() {
						opts = map[string]interface{}{"keyring": "some-keyring", "remote_mount_point": "some-remotemountpoint", "local_mount_point": "some-localmoutnpoint"}
						createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
						createResponse = driver.Create(testEnv, createRequest)
						Expect(createResponse.Err).To(Equal("Missing mandatory 'ip' field in 'Opts'"))
					})
					It("should not be able to retrieve volume", func() {
						getUnsuccessful(testEnv, driver, "some-volume-name")
					})
				})
			})
			Context("when volume already exists", func() {
				BeforeEach(func() {
					opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "local_mount_point": "some-localmountpoint", "remote_mount_point": "some-remote-mountpoint"}
					createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
					createResponse = driver.Create(testEnv, createRequest)
					Expect(createResponse.Err).To(Equal(""))
				})
				It("fails when given different metadata.", func() {
					opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "local_mount_point": "some-localmountpoint", "remote_mount_point": "someother-remote-mountpoint"}
					createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
					createResponse = driver.Create(testEnv, createRequest)
					Expect(createResponse.Err).To(Equal("Volume 'some-volume-name' already exists with different Opts"))
				})
				It("succeeds when given same metadata", func() {
					opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "local_mount_point": "some-localmountpoint", "remote_mount_point": "some-remote-mountpoint"}
					createRequest := voldriver.CreateRequest{Name: "some-volume-name", Opts: opts}
					createResponse = driver.Create(testEnv, createRequest)
					Expect(createResponse.Err).To(Equal(""))
				})
			})
		})

	})

	Describe(".List", func() {
		var (
			volumeName string
			opts       map[string]interface{}
		)

		Context("when there is a created/attached volume", func() {
			BeforeEach(func() {
				volumeName = "volume-name"
				opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "local_mount_point": "some-localmountpoint", "remote_mount_point": "some-remote-mountpoint"}
				createSuccessful(testEnv, driver, volumeName, opts)
			})

			It("should list the volume with an empty mountpoint for unmounted volumes", func() {
				listResponse := driver.List(testEnv)
				Expect(listResponse.Err).To(Equal(""))
				Expect(listResponse.Volumes[0].Name).To(Equal("volume-name"))
				Expect(listResponse.Volumes[0].Mountpoint).To(Equal(""))
			})

			Context("when the mount completes successfully", func() {
				BeforeEach(func() {
					fakeInvoker.InvokeReturns(nil, nil)
					mountSuccessful(testEnv, driver, volumeName)

					Expect(fakeOs.MkdirAllCallCount()).To(Equal(1))
					path, _ := fakeOs.MkdirAllArgsForCall(0)
					Expect(path).To(Equal("some-localmountpoint"))
				})

				It("should list the volume with an empty mountpoint for unmounted volumes", func() {
					listResponse := driver.List(testEnv)
					Expect(listResponse.Err).To(Equal(""))
					Expect(listResponse.Volumes[0].Name).To(Equal("volume-name"))
					Expect(listResponse.Volumes[0].Mountpoint).To(Equal("some-localmountpoint"))
				})
			})
		})
	})

	Describe(".Path", func() {
		var (
			volumeName string
			opts       map[string]interface{}
		)

		It("should report an error for non-existent volume", func() {
			pathResponse := driver.Path(testEnv, voldriver.PathRequest{
				Name: "unknown",
			})

			Expect(pathResponse.Err).NotTo(Equal(""))
		})

		Context("when there is a created/attached volume", func() {
			BeforeEach(func() {
				volumeName = "volume-name"
				opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "local_mount_point": "some-localmountpoint", "remote_mount_point": "some-remote-mountpoint"}
				createSuccessful(testEnv, driver, volumeName, opts)
			})

			It("should report an error for Path on unmounted volume", func() {
				pathResponse := driver.Path(testEnv, voldriver.PathRequest{
					Name: volumeName,
				})

				Expect(pathResponse.Err).NotTo(Equal(""))
			})

			Context("when the mount completes successfully", func() {
				BeforeEach(func() {
					fakeInvoker.InvokeReturns(nil, nil)
					mountSuccessful(testEnv, driver, volumeName)

					Expect(fakeOs.MkdirAllCallCount()).To(Equal(1))
					path, _ := fakeOs.MkdirAllArgsForCall(0)
					Expect(path).To(Equal("some-localmountpoint"))
				})

				It("should return Path correctly", func() {
					pathResponse := driver.Path(testEnv, voldriver.PathRequest{
						Name: volumeName,
					})

					Expect(pathResponse.Err).To(Equal(""))
					Expect(pathResponse.Mountpoint).To(Equal("some-localmountpoint"))
				})
			})
		})
	})

	Describe(".Mount", func() {
		var (
			mountResponse voldriver.MountResponse
			volumeName    string
			opts          map[string]interface{}
		)
		Context("when there is a created/attached volume", func() {
			BeforeEach(func() {
				volumeName = "volume-name"
				opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "local_mount_point": "some-localmountpoint", "remote_mount_point": "some-remote-mountpoint"}
				createSuccessful(testEnv, driver, volumeName, opts)
			})

			It("should report an error for volume name mismatch", func() {
				mountRequest := voldriver.MountRequest{Name: "garbage"}
				mountResponse = driver.Mount(testEnv, mountRequest)
				Expect(mountResponse.Err).To(Equal("Volume 'garbage' not found"))
				Expect(mountResponse.Mountpoint).To(Equal(""))
			})

			It("should report an error if keyfile creation errors", func() {
				fakeIoutil.WriteFileReturns(fmt.Errorf("error writing file"))
				mountRequest := voldriver.MountRequest{Name: volumeName}
				mountResponse = driver.Mount(testEnv, mountRequest)
				Expect(mountResponse.Err).To(Equal(fmt.Sprintf("Error mounting '%s' (error writing file)", volumeName)))
			})

			It("should report an error if CLI invocation fails", func() {
				fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("invocation fails"))
				mountRequest := voldriver.MountRequest{Name: volumeName}
				mountResponse = driver.Mount(testEnv, mountRequest)
				Expect(mountResponse.Err).To(Equal(fmt.Sprintf("Error mounting '%s' (invocation fails)", volumeName)))
			})

			Context("when the mount completes successfully", func() {
				BeforeEach(func() {
					fakeInvoker.InvokeReturns(nil, nil)
					mountSuccessful(testEnv, driver, volumeName)

					Expect(fakeOs.MkdirAllCallCount()).To(Equal(1))
					path, _ := fakeOs.MkdirAllArgsForCall(0)
					Expect(path).To(Equal("some-localmountpoint"))
				})

				It("invokes Ceph with a remote mountpoint", func() {
					_, _, args := fakeInvoker.InvokeArgsForCall(0)
					Expect(args).To(ContainElement("-r"))
					Expect(args).To(ContainElement("some-remote-mountpoint"))
				})

				It("creates a keyfile", func() {
					Expect(fakeIoutil.WriteFileCallCount()).To(Equal(1))
				})

				It("can get the volume and it is mounted path", func() {
					getResponse := getSuccessful(testEnv, driver, volumeName)
					Expect(getResponse.Volume.Mountpoint).To(Equal("some-localmountpoint"))
				})

				It("should return mountpoint", func() {
					mountResponse = driver.Mount(testEnv, voldriver.MountRequest{
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
				opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "local_mount_point": "some-localmountpoint", "remote_mount_point": "some-remote-mountpoint"}
				createSuccessful(testEnv, driver, volumeName, opts)
			})

			It("should report an error for volume name mismatch", func() {
				unmountRequest := voldriver.UnmountRequest{Name: "garbage"}
				unmountResponse = driver.Unmount(testEnv, unmountRequest)
				Expect(unmountResponse.Err).To(Equal("Volume 'garbage' is unknown"))
			})

			It("should error when volume is not mounted", func() {
				unmountRequest := voldriver.UnmountRequest{Name: volumeName}
				unmountResponse = driver.Unmount(testEnv, unmountRequest)

				Expect(unmountResponse.Err).To(Equal(fmt.Sprintf("Volume '%s' not mounted", volumeName)))
			})
			It("should error when volume is not created", func() {
				unmountRequest := voldriver.UnmountRequest{Name: "non-existent-volume"}
				unmountResponse = driver.Unmount(testEnv, unmountRequest)
				Expect(unmountResponse.Err).To(Equal(fmt.Sprintf("Volume '%s' is unknown", "non-existent-volume")))
			})

			Context("when volume mounted", func() {
				BeforeEach(func() {
					mountSuccessful(testEnv, driver, volumeName)

					_, cmd, _ := fakeInvoker.InvokeArgsForCall(0)
					Expect(cmd).To(Equal("ceph-fuse"))
				})
				It("should report an error if remove config file fails", func() {
					fakeOs.RemoveReturns(fmt.Errorf("file deletion failed"))
					unmountRequest := voldriver.UnmountRequest{Name: volumeName}
					unmountResponse = driver.Unmount(testEnv, unmountRequest)
					Expect(unmountResponse.Err).To(Equal(fmt.Sprintf("Error unmounting '%s' (file deletion failed)", volumeName)))
				})
				It("should report an error if CLI invocation fails", func() {
					fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("invocation fails"))
					unmountRequest := voldriver.UnmountRequest{Name: volumeName}
					unmountResponse = driver.Unmount(testEnv, unmountRequest)
					Expect(unmountResponse.Err).To(Equal(fmt.Sprintf("Error unmounting '%s' (invocation fails)", volumeName)))
				})

				Context("when fusermount -u successful", func() {
					BeforeEach(func() {
						fakeInvoker.InvokeReturns(nil, nil)

						unmountSuccessful(testEnv, driver, volumeName)

						Expect(fakeInvoker.InvokeCallCount()).To(Equal(2)) // mount and umount commands
						_, cmd, _ := fakeInvoker.InvokeArgsForCall(1)
						Expect(cmd).To(Equal("fusermount"))
					})
					It("only gets volume name, without Mountpoint", func() {
						getResponse := getSuccessful(testEnv, driver, volumeName)
						Expect(getResponse.Volume.Mountpoint).To(Equal(""))
					})
					It("removes keyfile and local mountpoint directory", func() {
						Expect(fakeOs.RemoveCallCount()).To(Equal(2))
						mountPointPath := fakeOs.RemoveArgsForCall(1)
						Expect(mountPointPath).To(Equal("some-localmountpoint"))
					})
				})

				Context("when the volume is mounted for a second time and then unmounted", func() {
					BeforeEach(func() {
						mountSuccessful(testEnv, driver, volumeName)
						unmountSuccessful(testEnv, driver, volumeName)
					})
					It("can still get the volume and it is mounted path", func() {
						getResponse := getSuccessful(testEnv, driver, volumeName)
						Expect(getResponse.Volume.Mountpoint).To(Equal("some-localmountpoint"))
					})
				})
			})
		})
	})

	Describe(".Remove", func() {
		const volumeName = "volume-name"
		var opts map[string]interface{}

		It("should fail if no volume name provided", func() {
			removeResponse := driver.Remove(testEnv, voldriver.RemoveRequest{
				Name: "",
			})
			Expect(removeResponse.Err).To(Equal("Missing mandatory 'volume_name'"))
		})

		It("should fail if no volume was created", func() {
			removeResponse := driver.Remove(testEnv, voldriver.RemoveRequest{
				Name: volumeName,
			})
			Expect(removeResponse.Err).To(Equal("Volume 'volume-name' not found"))
		})

		Context("when there is a created/attached volume", func() {
			BeforeEach(func() {
				opts = map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "local_mount_point": "some-localmountpoint", "remote_mount_point": "some-remote-mountpoint"}
				createSuccessful(testEnv, driver, volumeName, opts)
			})
			It("destroys volume", func() {
				removeResponse := driver.Remove(testEnv, voldriver.RemoveRequest{
					Name: volumeName,
				})
				Expect(removeResponse.Err).To(Equal(""))
				getUnsuccessful(testEnv, driver, volumeName)
			})

			Context("when volume mounted", func() {
				BeforeEach(func() {
					mountSuccessful(testEnv, driver, volumeName)
				})
				It("unmounts and destroys volume", func() {
					removeResponse := driver.Remove(testEnv, voldriver.RemoveRequest{
						Name: volumeName,
					})
					Expect(removeResponse.Err).To(Equal(""))
					getUnsuccessful(testEnv, driver, volumeName)
					Expect(fakeOs.RemoveCallCount()).To(Equal(2))
					mountPointPath := fakeOs.RemoveArgsForCall(1)
					Expect(mountPointPath).To(Equal("some-localmountpoint"))
				})
				Context("when unmount fails", func() {
					BeforeEach(func() {
						fakeInvoker.InvokeReturns([]byte("error"), fmt.Errorf("invocation fails"))
					})
					It("returns error", func() {
						removeResponse := driver.Remove(testEnv, voldriver.RemoveRequest{
							Name: volumeName,
						})
						Expect(removeResponse.Err).To(Equal("Error unmounting '" + volumeName + "' (invocation fails)"))
					})

				})
			})
		})
	})

})

func createSuccessful(env voldriver.Env, driver voldriver.Driver, volumeName string, opts map[string]interface{}) {
	createRequest := voldriver.CreateRequest{Name: volumeName, Opts: opts}
	createResponse := driver.Create(env, createRequest)
	Expect(createResponse.Err).To(Equal(""))
}

func getUnsuccessful(env voldriver.Env, localDriver voldriver.Driver, volumeName string) {
	getResponse := localDriver.Get(env, voldriver.GetRequest{
		Name: volumeName,
	})

	Expect(getResponse.Err).To(Equal("Volume '" + volumeName + "' not found"))
	Expect(getResponse.Volume.Name).To(Equal(""))
}

func getSuccessful(env voldriver.Env, localDriver voldriver.Driver, volumeName string) voldriver.GetResponse {
	getResponse := localDriver.Get(env, voldriver.GetRequest{
		Name: volumeName,
	})

	Expect(getResponse.Err).To(Equal(""))
	Expect(getResponse.Volume.Name).To(Equal(volumeName))
	return getResponse
}

func mountSuccessful(env voldriver.Env, localDriver voldriver.Driver, volumeName string) {
	mountResponse := localDriver.Mount(env, voldriver.MountRequest{
		Name: volumeName,
	})
	Expect(mountResponse.Err).To(Equal(""))
	Expect(mountResponse.Mountpoint).To(Equal("some-localmountpoint"))
}

func unmountSuccessful(env voldriver.Env, localDriver voldriver.Driver, volumeName string) {
	unmountResponse := localDriver.Unmount(env, voldriver.UnmountRequest{
		Name: volumeName,
	})
	Expect(unmountResponse.Err).To(Equal(""))
}
