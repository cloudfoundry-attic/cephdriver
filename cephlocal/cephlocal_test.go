package cephlocal_test

import (
	"github.com/cloudfoundry-incubator/cephdriver/cephlocal"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("cephlocal", func() {

	var (
		driver     *cephlocal.LocalDriver
		fakeExec   *volmanfakes.FakeExec
		fakeCmd    *volmanfakes.FakeCmd
		testLogger lager.Logger

		volumeName      string
		localMountPoint string
	)

	BeforeEach(func() {
		fakeExec = new(volmanfakes.FakeExec)
		fakeCmd = new(volmanfakes.FakeCmd)

		driver = cephlocal.NewLocalDriverWithExec(fakeExec)
		testLogger = lagertest.NewTestLogger("CephdriverTest")
	})

	Describe(".Mount", func() {
		var (
			mountResponse voldriver.MountResponse
		)

		BeforeEach(func() {
			volumeName = "volume-name"
			localMountPoint = "/local/mount/point"

			createRequest := voldriver.CreateRequest{
				Name: volumeName,
				Opts: map[string]interface{}{
					"ip":                "127.0.0.1",
					"port":              "6789",
					"key-ring":          "a-key-ring-string",
					"remote-mountpoint": "/remote/server/file/to/be/mounted",
					"local-mountpoint":  localMountPoint,
				},
			}
			createResponse := driver.Create(testLogger, createRequest)
			Expect(createResponse.Err).To(Equal(""))
		})

		JustBeforeEach(func() {
			mountResponse = driver.Mount(testLogger, voldriver.MountRequest{})
		})

		Context("when there are no errors", func() {
			JustBeforeEach(func() {
				Expect(mountResponse.Err).To(Equal(""))
			})

			It("properly invokes the ceph-fuse executable", func() {
				Expect(fakeExec.CommandCallCount()).To(Equal(1))
				command, _ := fakeExec.CommandArgsForCall(0)
				Expect(command).To(Equal("ceph-fuse"))
			})

			It("returns a proper mountResponse", func() {
				Expect(mountResponse.Mountpoint).To(Equal(localMountPoint))
			})
		})

		Context("errors", func() {
			Context("from the ceph fuse executable", func() {
				BeforeEach(func() {

				})

				It("sets the error on the mountResponse", func() {})
			})
		})
	})
})

//var cephdriverPath string
// ip         string
// keyring    string
// mountpoint string

// fakeExec.CommandReturns(fakeCmd)

// keyring = "/tmp/test-keyring"
// ip = "192.168.100.50"
// mountpoint = "/mnt/test-mount"

// config := &cephdriver.MountConfig{Keyring: keyring, MountPoint: mountpoint, IP: ip}

// configJsonBlob, err := json.Marshal(config)
// Expect(err).NotTo(HaveOccurred())

// mountRequest := voldriver.MountRequest{Name: "test"}
// mountResponse := driver.Mount(testLogger, mountRequest)

// Expect(err).NotTo(HaveOccurred())
// Expect(fakeExec.CommandCallCount()).To(Equal(1))
// executable, args := fakeExec.CommandArgsForCall(0)
// Expect(executable).To(Equal("ceph-fuse"))
// Expect(len(args)).To(Equal(5))
// Expect(args[0]).To(Equal("-k"))
// // unable to test arg[1] as it is a keyring file generated from the supplied content
// Expect(args[2]).To(Equal("-m"))
// Expect(args[3]).To(Equal(fmt.Sprintf("%s:6789", ip)))
// Expect(args[4]).To(Equal(mountpoint))

// Expect(fakeCmd.StartCallCount()).To(Equal(1))
