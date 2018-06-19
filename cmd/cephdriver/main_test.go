package main_test

import (
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"io/ioutil"
	"net"
	"os/exec"

	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	var (
		session *gexec.Session
		command *exec.Cmd
		err     error
		logger  lager.Logger
	)

	BeforeEach(func() {
		command = exec.Command(driverPath)
		logger = lagertest.NewTestLogger("test-ceph-driver")
	})

	JustBeforeEach(func() {
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		session.Kill().Wait()
	})

	Context("with a driver path", func() {
		var (
			listenAddr string
			debugAddr  string
		)

		BeforeEach(func() {
			dir, err := ioutil.TempDir("", "driversPath")
			Expect(err).ToNot(HaveOccurred())
			port := 9750 + GinkgoParallelNode()
			listenAddr = fmt.Sprintf("127.0.0.1:%d", port)
			debugPort := 9850 + GinkgoParallelNode()
			debugAddr = fmt.Sprintf("127.0.0.1:%d", debugPort)
			command.Args = append(command.Args, "-driversPath="+dir)
			command.Args = append(command.Args, "-listenAddr="+listenAddr)
			command.Args = append(command.Args, "-debugAddr="+debugAddr)
		})

		It("listens on the specified port", func() {
			EventuallyWithOffset(1, func() error {
				_, err := net.Dial("tcp", listenAddr)
				return err
			}, 5).ShouldNot(HaveOccurred())
		})

		It("listens on the debug address for admin reqs by default", func() {
			EventuallyWithOffset(1, func() error {
				_, err := net.Dial("tcp", debugAddr)
				return err
			}, 5).ShouldNot(HaveOccurred())
		})
	})
})
