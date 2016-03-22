package main_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/volman/voldriver"
	"github.com/cloudfoundry-incubator/volman/voldriver/driverhttp"

	"github.com/cloudfoundry-incubator/cf_http/unix_transport"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Cephdriver Cmd test ", func() {
	Context("given a started cephdriver Unix server", func() {
		BeforeEach(func() {
			cephdriverUnixServerProcess = ginkgomon.Invoke(unixRunner)
			time.Sleep(time.Millisecond * 1000)
		})

		It("should get a 404 for root to prove it started", func() {
			unixSocketUrl := "unix://" + socketPath
			_, status, err := getUnix(unixSocketUrl, socketPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(status).Should(ContainSubstring("404"))
		})
	})

	Context("given a mounted volume", func() {
		var volumeName string
		var mountResponse voldriver.MountResponse
		var client voldriver.Driver
		var testLogger *lagertest.TestLogger
		JustBeforeEach(func() {
			cephdriverUnixServerProcess = ginkgomon.Invoke(unixRunner)
			time.Sleep(time.Millisecond * 1000)

			client = driverhttp.NewRemoteUnixClient(socketPath)
			testLogger = lagertest.NewTestLogger("cephDriver Server Test")
			node := GinkgoParallelNode()
			volumeName = "fake-volume-name_" + strconv.Itoa(node)
			testLogger.Info("creating-volume", lager.Data{"name": volumeName})

			opts := map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "localMountPoint": "some-localmountpoint", "remoteMountPoint": "some-remote-mountpoint"}
			createRequest := voldriver.CreateRequest{Name: volumeName, Opts: opts}
			createResponse := client.Create(testLogger, createRequest)
			Expect(createResponse.Err).To(Equal(""))

			mountRequest := voldriver.MountRequest{Name: volumeName}
			mountResponse = client.Mount(testLogger, mountRequest)
			Expect(mountResponse.Err).To(Equal(""))
		})

		It("should return proper mountpoint", func() {
			Expect(mountResponse.Mountpoint).NotTo(Equal(""))
		})

		It("should unmount a volume given same volume name", func() {
			unmountRequest := voldriver.UnmountRequest{Name: volumeName}
			unmountErr := client.Unmount(testLogger, unmountRequest)
			Expect(unmountErr.Err).To(Equal(""))
		})

	})
	Context("given a started cephdriver server", func() {
		BeforeEach(func() {
			cephdriverServerProcess = ginkgomon.Invoke(runner)
			time.Sleep(time.Millisecond * 1000)
		})

		It("should not exit", func() {
			Consistently(runner).ShouldNot(Exit())
		})

		It("should get a 404 for root", func() {
			_, status, err := get("/")
			Expect(err).NotTo(HaveOccurred())
			Expect(status).Should(ContainSubstring("404"))
		})

		It("should have a debug server endpoint", func() {
			_, err := http.Get(fmt.Sprintf("http://%s/debug/pprof/goroutine", debugServerAddress))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("given a mounted volume", func() {
			var volumeId string
			var volumeName string
			var mountResponse voldriver.MountResponse

			JustBeforeEach(func() {
				client := driverhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", cephdriverServerPort))
				testLogger := lagertest.NewTestLogger("cephDriver Server Test")
				node := GinkgoParallelNode()
				volumeId = "fake-volume-id_" + strconv.Itoa(node)
				volumeName = "fake-volume-name_" + strconv.Itoa(node)
				testLogger.Info("creating-volume", lager.Data{"name": volumeName})
				opts := map[string]interface{}{"keyring": "some-keyring", "ip": "some-ip", "localMountPoint": "some-localmountpoint", "remoteMountPoint": "some-remote-mountpoint"}
				createRequest := voldriver.CreateRequest{Name: volumeName, Opts: opts}
				createResponse := client.Create(testLogger, createRequest)
				Expect(createResponse.Err).To(Equal(""))

				mountRequest := voldriver.MountRequest{Name: volumeName}
				mountResponse = client.Mount(testLogger, mountRequest)
				Expect(mountResponse.Err).To(Equal(""))
			})

			It("should return a propoer mountpoint", func() {
				Expect(mountResponse.Mountpoint).NotTo(Equal(""))
			})

			It("should unmount a volume given same volume name", func() {
				client := driverhttp.NewRemoteClient(fmt.Sprintf("http://0.0.0.0:%d", cephdriverServerPort))
				testLogger := lagertest.NewTestLogger("cephDriver Server Test")
				unmountRequest := voldriver.UnmountRequest{Name: volumeName}
				unmountErr := client.Unmount(testLogger, unmountRequest)
				Expect(unmountErr.Err).To(Equal(""))

			})

		})
	})
})

func get(path string) (body string, status string, err error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://0.0.0.0:%d%s", cephdriverServerPort, path), nil)
	response, _ := (&http.Client{}).Do(req)
	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	return string(bodyBytes[:]), response.Status, err
}

func getUnix(path string, socketPath string) (body string, status string, err error) {
	req, _ := http.NewRequest("GET", path, nil)
	client := http.Client{
		Transport: unix_transport.New(socketPath),
	}
	response, _ := client.Do(req)
	defer response.Body.Close()
	bodyBytes, err := ioutil.ReadAll(response.Body)
	return string(bodyBytes[:]), response.Status, err
}
