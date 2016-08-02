package main

import (
	"flag"
	"os"

	"code.cloudfoundry.org/lager"

	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/sigmon"

	"code.cloudfoundry.org/cephdriver/cephlocal"

	"syscall"

	cf_lager "code.cloudfoundry.org/cflager"
	cf_debug_server "code.cloudfoundry.org/debugserver"
)

func init() {
	// no command line parsing can happen here in go 1.6git soll
}

func main() {
	cephServerConfig := cephlocal.CephServerConfig{}
	parseCommandLine(&cephServerConfig)

	withLogger, _ := cf_lager.New("ceph-driver-server")

	syscall.Umask(000)

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
