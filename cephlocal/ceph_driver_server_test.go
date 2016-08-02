package cephlocal_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"

	"code.cloudfoundry.org/cephdriver/cephlocal"
)

var _ = Describe("Ceph Driver Server", func() {

	var (
		err    error
		logger lager.Logger

		tmpDir                 string
		atAddress, driversPath string

		cephDriverServer *cephlocal.CephDriverServerStruct
		cephDriverConfig cephlocal.CephServerConfig
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("ceph-driver-server")

		tmpDir, err = ioutil.TempDir("", "ceph-driver-server-unit-test")
		Expect(err).NotTo(HaveOccurred())

		cephDriverConfig = cephlocal.CephServerConfig{
			AtAddress:   "fake-address",
			DriversPath: "fake-drivers-path",
		}
	})

	AfterEach(func() {
		err = os.RemoveAll(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("#Runner", func() {
		Context("when address and transport are IP and tcp", func() {
			BeforeEach(func() {
				cephDriverConfig = cephlocal.CephServerConfig{
					AtAddress:   "0.0.0.0:9750",
					DriversPath: tmpDir,
				}
				cephDriverServer = cephlocal.NewCephDriverServer(cephDriverConfig).(*cephlocal.CephDriverServerStruct)
			})

			It("creates a ifrit.Runner", func() {
				runner, err := cephDriverServer.CreateTcpServer(logger, cephDriverConfig.AtAddress, cephDriverConfig.DriversPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(runner).NotTo(BeNil())
			})
		})

		Context("when address and transport are sock file and unix", func() {
			BeforeEach(func() {
				cephDriverConfig = cephlocal.CephServerConfig{
					AtAddress:   "fake.sock",
					DriversPath: tmpDir,
				}
				cephDriverServer = cephlocal.NewCephDriverServer(cephDriverConfig).(*cephlocal.CephDriverServerStruct)
			})

			It("creates a ifrit.Runner", func() {
				runner, err := cephDriverServer.CreateUnixServer(logger, cephDriverConfig.AtAddress, cephDriverConfig.DriversPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(runner).NotTo(BeNil())
			})
		})
	})

	Describe("#DetermineTransport", func() {
		BeforeEach(func() {
			cephDriverServer = cephlocal.NewCephDriverServer(cephDriverConfig).(*cephlocal.CephDriverServerStruct)
		})

		Context("when address is an IP", func() {
			It("returns transport as tcp", func() {
				transport := cephDriverServer.DetermineTransport("0.0.0.0:9750")
				Expect(transport).To(Equal("tcp"))
			})
		})

		Context("when address is a sock file", func() {
			It("returns transport as unix", func() {
				transport := cephDriverServer.DetermineTransport("fake.sock")
				Expect(transport).To(Equal("unix"))
			})
		})
	})

	Describe("#CreateTcpServer", func() {
		BeforeEach(func() {
			atAddress = "0.0.0.0:9750"
			driversPath = tmpDir
		})

		Context("when we have valid atAddress and driversPath", func() {
			It("returns a ifrit.Runner", func() {
				runner, err := cephDriverServer.CreateTcpServer(logger, atAddress, driversPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(runner).NotTo(BeNil())
			})
		})

		Context("when we have invalid arguments", func() {
			Context("when atAddress is invalid", func() {
				It("fails creating Runner", func() {
					runner, err := cephDriverServer.CreateTcpServer(logger, "...", driversPath)
					Expect(err).To(HaveOccurred())
					Expect(runner).To(BeNil())
				})
			})

			Context("when atAddress and driversPath are invalid", func() {
				It("fails creating Runner", func() {
					runner, err := cephDriverServer.CreateTcpServer(logger, "...", "/root/../..")
					Expect(err).To(HaveOccurred())
					Expect(runner).To(BeNil())
				})
			})
		})
	})

	Describe("#CreateUnixServer", func() {
		BeforeEach(func() {
			atAddress = "something.sock"
			driversPath = tmpDir
		})

		Context("when we have valid atAddress and driversPath", func() {
			It("returns a ifrit.Runner", func() {
				runner, err := cephDriverServer.CreateUnixServer(logger, atAddress, driversPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(runner).NotTo(BeNil())
			})
		})

		Context("when we have invalid arguments", func() {
			Context("when atAddress is invalid", func() {
				It("fails creating Runner", func() {
					runner, err := cephDriverServer.CreateUnixServer(logger, "~/fake-address", driversPath)
					Expect(err).To(HaveOccurred())
					Expect(runner).To(BeNil())
				})
			})

			Context("when atAddress and driversPath are invalid", func() {
				It("fails creating Runner", func() {
					runner, err := cephDriverServer.CreateUnixServer(logger, "~/fake-address", "/root/../..")
					Expect(err).To(HaveOccurred())
					Expect(runner).To(BeNil())
				})
			})
		})
	})
})
