package main_test

import (
	"time"
	"testing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"strings"
)

var cephdriverPath string

func TestCephdriver(t *testing.T) {
	// these integration tests can take a bit, especially under load;
	// 1 second is too harsh
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
	cephdriverPath = strings.TrimSuffix(cephdriverPath,"/cephdriver")
})



var _ = SynchronizedAfterSuite(func() {

}, func() {
	gexec.CleanupBuildArtifacts()
})
