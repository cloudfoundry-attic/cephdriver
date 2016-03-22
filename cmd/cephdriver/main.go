package main

import (
	"flag"
	"os"

	"github.com/cloudfoundry-incubator/cephdriver/cephlocal"
	"github.com/cloudfoundry-incubator/cephdriver/cephlocal/cephfakes"
	cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	cf_lager "github.com/cloudfoundry-incubator/cf-lager"

	"github.com/cloudfoundry-incubator/volman/voldriver/driverhttp"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
)

var atAddress = flag.String(
	"listenAddr",
	"0.0.0.0:9750",
	"host:port to serve volume management functions",
)

var driversPath = flag.String(
	"driversPath",
	"",
	"Path to directory where drivers are installed",
)

var transport = flag.String(
	"transport",
	"tcp",
	"Transport protocol to transmit HTTP over",
)
var isUnit = flag.Bool(
	"unit",
	false,
	"Set to true for unit testing with fakes",
)

func init() {
	// no command line parsing can happen here in go 1.6git soll
}

func main() {
	parseCommandLine()

	var withLogger lager.Logger
	var logTap *lager.ReconfigurableSink

	var cephDriverServer ifrit.Runner

	withLogger, logTap = logger(*transport)
	cephDriverServer = createCephDriverServer(withLogger, *atAddress, *driversPath, *isUnit, *transport)

	servers := grouper.Members{
		{"cephdriver-server", cephDriverServer},
	}
	if dbgAddr := cf_debug_server.DebugAddress(flag.CommandLine); dbgAddr != "" {
		servers = append(grouper.Members{
			{"debug-server", cf_debug_server.Runner(dbgAddr, logTap)},
		}, servers...)
	}
	process := ifrit.Invoke(processRunnerFor(servers))
	untilTerminated(withLogger, process)
}

func exitOnFailure(logger lager.Logger, err error) {
	if err != nil {
		logger.Error("fatal-err..aborting", err)
		panic(err.Error())
	}
}

func untilTerminated(logger lager.Logger, process ifrit.Process) {
	err := <-process.Wait()
	exitOnFailure(logger, err)
}

func processRunnerFor(servers grouper.Members) ifrit.Runner {
	return sigmon.New(grouper.NewOrdered(os.Interrupt, servers))
}

func createCephDriverServer(logger lager.Logger, atAddress string, driversPath string, isUnit bool, transport string) ifrit.Runner {
	logger.Info("started")
	defer logger.Info("ends")
	var client *cephlocal.LocalDriver
	if isUnit == true {
		fakeInvoker := new(cephfakes.FakeInvoker)
		fakeSystemUtil := new(cephfakes.FakeSystemUtil)
		client = cephlocal.NewLocalDriverWithInvokerAndSystemUtil(fakeInvoker, fakeSystemUtil)
	} else {
		client = cephlocal.NewLocalDriver()
	}
	handler, err := driverhttp.NewHandler(logger, client)
	exitOnFailure(logger, err)
	if transport == "tcp" {
		return http_server.New(atAddress, handler)
	}
	return http_server.NewUnixServer(atAddress, handler)
}

func logger(transport string) (lager.Logger, *lager.ReconfigurableSink) {
	if transport == "tcp" {
		return cf_lager.New("cephdriverServer")
	}
	return cf_lager.New("cephdriverUnixServer")
}
func parseCommandLine() {
	cf_lager.AddFlags(flag.CommandLine)
	cf_debug_server.AddFlags(flag.CommandLine)
	flag.Parse()
}
