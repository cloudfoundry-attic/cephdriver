package main

import (
	"flag"
	"os"

	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"

	"github.com/cloudfoundry-incubator/cephdriver/cephlocal"

	cf_debug_server "github.com/cloudfoundry-incubator/cf-debug-server"
	cf_lager "github.com/cloudfoundry-incubator/cf-lager"
)

func init() {
	// no command line parsing can happen here in go 1.6git soll
}

func main() {
	cephServerConfig := cephlocal.CephServerConfig{}
	parseCommandLine(&cephServerConfig)

	withLogger := lager.NewLogger("ceph-driver-server")

	cephServer := cephlocal.NewCephDriverServer(cephServerConfig)
	cephDriverServer, err := cephServer.Runner(withLogger)
	exitOnFailure(withLogger, err)

	servers := grouper.Members{
		{"cephdriver-server", cephDriverServer},
	}

	var logTap *lager.ReconfigurableSink

	if dbgAddr := cf_debug_server.DebugAddress(flag.CommandLine); dbgAddr != "" {
		servers = append(grouper.Members{
			{"debug-server", cf_debug_server.Runner(dbgAddr, logTap)},
		}, servers...)
	}

	runner := sigmon.New(grouper.NewOrdered(os.Interrupt, servers))
	process := ifrit.Invoke(runner)
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

func parseCommandLine(config *cephlocal.CephServerConfig) {
	flag.StringVar(&config.AtAddress, "listenAddr", "0.0.0.0:9750", "host:port to serve volume management functions")
	flag.StringVar(&config.DriversPath, "driversPath", "", "Path to directory where drivers are installed")
	flag.StringVar(&config.Transport, "transport", "tcp", "Transport protocol to transmit HTTP over")

	cf_lager.AddFlags(flag.CommandLine)
	cf_debug_server.AddFlags(flag.CommandLine)

	flag.Parse()
}
