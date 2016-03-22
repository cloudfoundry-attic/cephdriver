package main_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	"testing"
)

var cephdriverServerPath string

var runner *ginkgomon.Runner
var unixRunner *ginkgomon.Runner
var cephdriverServerPort int
var cephdriverServerProcess ifrit.Process
var cephdriverUnixServerProcess ifrit.Process
var debugServerAddress string
var socketPath string

func TestCephdriver(t *testing.T) {

	SetDefaultEventuallyTimeout(10 * time.Second)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Cephdriver Cmd Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	cephdriverServerPath, err = gexec.Build("github.com/cloudfoundry-incubator/cephdriver/cmd/cephdriver", "-race")
	Expect(err).NotTo(HaveOccurred())

	return []byte(cephdriverServerPath)
}, func(pathsByte []byte) {
	cephdriverServerPath = string(pathsByte)
})

var _ = BeforeEach(func() {

	cephdriverServerPort = 9750 + GinkgoParallelNode()
	debugServerAddress = fmt.Sprintf("0.0.0.0:%d", 9850+GinkgoParallelNode())
	runner = ginkgomon.New(ginkgomon.Config{
		Name: "cephdriverServer",
		Command: exec.Command(
			cephdriverServerPath,
			"-listenAddr", fmt.Sprintf("0.0.0.0:%d", cephdriverServerPort),
			"-debugAddr", debugServerAddress,
			"-unit", "true",
		),
		StartCheck: "cephdriverServer.started",
	})

	tmpdir, err := ioutil.TempDir(os.TempDir(), "ceph-driver-test")
	Ω(err).ShouldNot(HaveOccurred())

	socketPath = path.Join(tmpdir, "cephdriver.sock")

	unixRunner = ginkgomon.New(ginkgomon.Config{
		Name: "cephdriverUnixServer",
		Command: exec.Command(
			cephdriverServerPath,
			"-listenAddr", socketPath,
			"-transport", "unix",
			"-unit", "true",
		),
		StartCheck: "cephdriverUnixServer.started",
	})
})

var _ = AfterEach(func() {
	ginkgomon.Kill(cephdriverServerProcess)
	ginkgomon.Kill(cephdriverUnixServerProcess)
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})
