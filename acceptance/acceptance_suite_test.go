package acceptance_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"
	"strconv"
)

var volmanPath string
var volmanServerPort int
var debugServerAddress string
var volmanRunner *ginkgomon.Runner

var driverPath string
var driverServerPort int
var debugServerAddress2 string
var driverRunner *ginkgomon.Runner
var tmpDriversPath string

var keyringFileContents string
var clusterIp string

func TestVolman(t *testing.T) {
	// these integration tests can take a bit, especially under load;
	// 1 second is too harsh
	SetDefaultEventuallyTimeout(10 * time.Second)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Cephdriver Cmd Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	volmanPath, err = gexec.Build("github.com/cloudfoundry-incubator/volman/cmd/volman", "-race")
	Expect(err).NotTo(HaveOccurred())

	driverPath, err = gexec.Build("github.com/cloudfoundry-incubator/cephdriver/cmd/cephdriver", "-race")
	Expect(err).NotTo(HaveOccurred())
	return []byte(strings.Join([]string{volmanPath, driverPath}, ","))
}, func(pathsByte []byte) {
	path := string(pathsByte)
	volmanPath = strings.Split(path, ",")[0]
	driverPath = strings.Split(path, ",")[1]

	// read config files
	keyringFileContents = os.Getenv("CEPH_KEYRING")
	clusterIp = os.Getenv("CEPH_CLUSTER_IP")
})

var _ = BeforeEach(func() {
	var err error

	tmpDriversPath = os.Getenv("VOLMAN_DRIVERSPATH")
	if len(tmpDriversPath) == 0 {
		tmpDriversPath, err = ioutil.TempDir(os.TempDir(), "ceph-driver-test")
		Expect(err).NotTo(HaveOccurred())
	}

	driverServerPort = 9750 + GinkgoParallelNode()
	debugServerAddress2 = fmt.Sprintf("127.0.0.1:%d", 9850+GinkgoParallelNode())

	isPipeline := false
	isPipeline, _ = strconv.ParseBool(os.Getenv("PERSI_PIPELINE"))

	if isPipeline {
		driverRunner = ginkgomon.New(ginkgomon.Config{
			Name: "noop",
		})
	} else {
		driverRunner = ginkgomon.New(ginkgomon.Config{
			Name: "cephdriverServer",
			Command: exec.Command(
				driverPath,
				"-listenAddr", fmt.Sprintf("127.0.0.1:%d", driverServerPort),
				"-debugAddr", debugServerAddress2,
				"-driversPath", tmpDriversPath,
			),
			StartCheck: "cephdriverServer.started",
		})
	}

	volmanServerPort = 8750 + GinkgoParallelNode()
	debugServerAddress = fmt.Sprintf("127.0.0.1:%d", 8850+GinkgoParallelNode())
	volmanRunner = ginkgomon.New(ginkgomon.Config{
		Name: "volman",
		Command: exec.Command(
			volmanPath,
			"-listenAddr", fmt.Sprintf("127.0.0.1:%d", volmanServerPort),
			"-debugAddr", debugServerAddress,
			"-driversPath", tmpDriversPath,
		),
		StartCheck: "volman.started",
	})
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})
