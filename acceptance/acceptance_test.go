package acceptance_test

import (
	"github.com/cloudfoundry-incubator/volman/certification"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tedsuo/ifrit/ginkgomon"
	"github.com/nu7hatch/gouuid"
)

var _ = Describe("Ceph Driver Certification", func() {
	certification.CertifiyWith("Cephdriver", func()(*ginkgomon.Runner, *ginkgomon.Runner, int, string, string, int, string, func() (string, map[string]interface{})) {

		volumeInfo := func()(string, map[string]interface{}){
			uuid, err := uuid.NewV4()
			Expect(err).NotTo(HaveOccurred())
			volumeName := "ceph-volume-name_" + uuid.String()
			volumeId := "ceph-volume-id_" + uuid.String()
			opts := map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "localMountPoint": "some-local-mountpoint-"+volumeId, "remoteMountPoint": "some-remote-mountpoint-"+volumeId}
			return volumeName, opts
		}

		return driverRunner, volmanRunner, volmanServerPort, debugServerAddress, tmpDriversPath, driverServerPort, "cephdriver", volumeInfo
	})
})
