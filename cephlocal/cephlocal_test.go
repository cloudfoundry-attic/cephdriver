package cephlocal_test

import (
	"encoding/json"
	"fmt"

	"github.com/cloudfoundry-incubator/cephdriver"
	"github.com/cloudfoundry-incubator/cephdriver/cephlocal"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/volmanfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("cephlocal", func() {
	var driver *cephlocal.LocalDriver
	var fakeExec *volmanfakes.FakeExec
	var fakeCmd *volmanfakes.FakeCmd

	//var cephdriverPath string
	var ip string
	var keyring string
	var mountpoint string

	BeforeEach(func() {
		fakeExec = new(volmanfakes.FakeExec)
		fakeCmd = new(volmanfakes.FakeCmd)

		driver = cephlocal.NewLocalDriverWithExec(fakeExec)
	})

	It("should be able to mount", func() {
		testLogger := lagertest.NewTestLogger("CephdriverTest")

		fakeExec.CommandReturns(fakeCmd)

		keyring = "/tmp/test-keyring"
		ip = "192.168.100.50"
		mountpoint = "/mnt/test-mount"

		config := &cephdriver.MountConfig{Keyring: keyring, MountPoint: mountpoint, IP: ip}
		//fakeExec.Command("ceph-fuse", "-k", "/tmp/keyring", "-m", fmt.Sprintf("%s:6789", config.IP), "/tmp/test")

		configJsonBlob, err := json.Marshal(config)
		Expect(err).NotTo(HaveOccurred())

		mountRequest := voldriver.MountRequest{VolumeId: "test", Config: string(configJsonBlob)}
		_, err = driver.Mount(testLogger, mountRequest)

		Expect(err).NotTo(HaveOccurred())
		Expect(fakeExec.CommandCallCount()).To(Equal(1))
		executable, args := fakeExec.CommandArgsForCall(0)
		Expect(executable).To(Equal("ceph-fuse"))
		Expect(len(args)).To(Equal(5))
		Expect(args[0]).To(Equal("-k"))
		// unable to test arg[1] as it is a keyring file generated from the supplied content
		Expect(args[2]).To(Equal("-m"))
		Expect(args[3]).To(Equal(fmt.Sprintf("%s:6789", ip)))
		Expect(args[4]).To(Equal(mountpoint))

		Expect(fakeCmd.StartCallCount()).To(Equal(1))

	})

})
