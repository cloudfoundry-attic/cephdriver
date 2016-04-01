package acceptance_test

import (
	"github.com/cloudfoundry-incubator/volman/certification"
	"github.com/nu7hatch/gouuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Ceph Driver Certification", func() {
	certification.CertifiyWith("Cephdriver", func() (*ginkgomon.Runner, *ginkgomon.Runner, int, string, string, int, string, func() (string, map[string]interface{})) {

		uuid, err := uuid.NewV4()
		Expect(err).NotTo(HaveOccurred())
		volumeName := "ceph-volume-name_" + uuid.String()
		volumeId := "ceph-volume-id_" + uuid.String()

		volumeInfo := func() (string, map[string]interface{}) {
			localMountPoint := tmpDriversPath + "/_cephdriver-" + volumeId
			opts := map[string]interface{}{"keyring": keyringFileContents, "ip": clusterIp, "localMountPoint": localMountPoint, "remoteMountPoint": "unused"}
			return volumeName, opts
		}

		return driverRunner, volmanRunner, volmanServerPort, debugServerAddress, tmpDriversPath, driverServerPort, "cephdriver", volumeInfo
	})
})
