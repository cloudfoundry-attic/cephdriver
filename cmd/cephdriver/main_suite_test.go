package main_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

var cephdriverPath string
var ip string
var keyring string
var mountpoint string

func TestCephdriver(t *testing.T) {

	SetDefaultEventuallyTimeout(10 * time.Second)

	RegisterFailHandler(Fail)
	RunSpecs(t, "Cephdriver Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error

	cephdriverPath, err = gexec.Build("github.com/cloudfoundry-incubator/cephdriver/cmd/cephdriver", "-race")
	Expect(err).NotTo(HaveOccurred())
	return []byte(cephdriverPath)
}, func(pathsByte []byte) {
	cephdriverPath = string(pathsByte)
	cephdriverPath = strings.TrimSuffix(cephdriverPath, "/cephdriver")
	bytesconfig, err := ioutil.ReadFile("/tmp/cephfs-files/ip.conf")
	if err != nil {
		fmt.Printf("%#v", err)
	}
	ip = string(bytesconfig)

	bytesconfig, err = ioutil.ReadFile("/tmp/cephfs-files/ceph.client.admin.keyring")
	if err != nil {
		fmt.Printf("%#v", err)
	}
	keyring = string(bytesconfig)

	bytesconfig, err = ioutil.ReadFile("/tmp/cephfs-files/mountpoint.conf")
	if err != nil {
		fmt.Printf("%#v", err)
	}
	mountpoint = string(bytesconfig)
})

var _ = SynchronizedAfterSuite(func() {

}, func() {
	//gexec.CleanupBuildArtifacts()
})
