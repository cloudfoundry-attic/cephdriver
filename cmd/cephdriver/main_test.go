package main_test

import (
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry-incubator/cephdriver"
	"github.com/cloudfoundry-incubator/volman/system"
	"github.com/cloudfoundry-incubator/volman/voldriver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-golang/lager/lagertest"
)

var _ = Describe("cephdriver", func() {
	var driver *voldriver.DriverClientCli

	BeforeEach(func() {
		driver = voldriver.NewDriverClientCli(cephdriverPath, &system.SystemExec{}, "cephdriver")
	})

	Context("after valid driver install ", func() {
		It("should respond with valid info", func() {
			testLogger := lagertest.NewTestLogger("CephdriverTest")
			infoResponse, err := driver.Info(testLogger)
			Expect(err).NotTo(HaveOccurred())
			Expect(infoResponse).NotTo(Equal(nil))
			Expect(infoResponse.Name).To(Equal("cephdriver"))
		})
		It("should be able to mount", func() {
			testLogger := lagertest.NewTestLogger("CephdriverTest")
			config := &cephdriver.MountConfig{Keyring: keyring, MountPoint: mountpoint, IP: ip}
			configJsonBlob, err := json.Marshal(config)
			Expect(err).NotTo(HaveOccurred())

			mountRequest := voldriver.MountRequest{VolumeId: "test", Config: string(configJsonBlob)}
			mountResponse, err := driver.Mount(testLogger, mountRequest)

			Expect(err).NotTo(HaveOccurred())
			Expect(mountResponse.Path).NotTo(Equal(nil))
			fmt.Println(mountResponse.Path)

		})

	})
})
