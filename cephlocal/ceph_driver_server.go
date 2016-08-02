package cephlocal

import (
	"errors"
	"io/ioutil"
	"os"

	"strings"

	"fmt"

	"encoding/json"

	"regexp"

	"code.cloudfoundry.org/lager"

	"code.cloudfoundry.org/voldriver"
	"code.cloudfoundry.org/voldriver/driverhttp"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/http_server"
)

const IPV4_REGEX string = `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\:([0-9]{1,4}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`

type CephServerConfig struct {
	AtAddress   string
	DriversPath string
	Transport   string
}

type CephDriverServer interface {
	Runner(logger lager.Logger) (ifrit.Runner, error)
}

type CephDriverServerStruct struct {
	config CephServerConfig
}

func NewCephDriverServer(config CephServerConfig) CephDriverServer {
	return &CephDriverServerStruct{
		config: config,
	}
}

func (server *CephDriverServerStruct) Runner(logger lager.Logger) (ifrit.Runner, error) {
	var err error
	var cephDriverServer ifrit.Runner

	server.config.Transport = server.DetermineTransport(server.config.AtAddress)

	if server.config.Transport == "tcp" {
		cephDriverServer, err = server.CreateTcpServer(logger, server.config.AtAddress, server.config.DriversPath)
	} else {
		cephDriverServer, err = server.CreateUnixServer(logger, server.config.AtAddress, server.config.DriversPath)

	}
	if err != nil {
		return nil, err
	}

	return cephDriverServer, nil
}

func (server *CephDriverServerStruct) CreateTcpServer(logger lager.Logger, atAddress string, driversPath string) (ifrit.Runner, error) {
	logger = logger.Session("create-tcp-server")
	logger.Info("start")
	defer logger.Info("ends")

	valid := server.isValidTcpAddress(atAddress)
	if !valid {
		return nil, errors.New(fmt.Sprintf("invalid-address %s", atAddress))
	}

	spec := voldriver.DriverSpec{
		Name:    "cephdriver",
		Address: server.protocolify(atAddress, "http"),
	}
	specJson, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	err = voldriver.WriteDriverSpec(logger, driversPath, "cephdriver", "json", specJson)
	if err != nil {
		return nil, err
	}

	handler, err := driverhttp.NewHandler(logger, NewLocalDriver())
	if err != nil {
		return nil, err
	}

	return http_server.New(atAddress, handler), nil
}

func (server *CephDriverServerStruct) CreateUnixServer(logger lager.Logger, atAddress string, driversPath string) (ifrit.Runner, error) {
	logger = logger.Session("create-unix-server")
	logger.Info("start")
	defer logger.Info("ends")

	err := server.isValidUnixSocketPath(atAddress)
	if err != nil {
		return nil, err
	}

	url := server.protocolify(atAddress, "unix")
	err = voldriver.WriteDriverSpec(logger, driversPath, "cephdriver", "spec", []byte(url))
	if err != nil {
		return nil, err
	}

	handler, err := driverhttp.NewHandler(logger, NewLocalDriver())
	if err != nil {
		return nil, err
	}

	return http_server.NewUnixServer(atAddress, handler), nil
}

func (server *CephDriverServerStruct) DetermineTransport(address string) string {
	if strings.HasSuffix(address, ".sock") {
		return "unix"
	}

	return "tcp"
}

// Private

func (server *CephDriverServerStruct) protocolify(address string, protocol string) string {
	if !strings.HasPrefix(address, protocol+"://") {
		return fmt.Sprintf("%s://%s", protocol, address)
	}
	return address
}

func (server *CephDriverServerStruct) isValidUnixSocketPath(socketPath string) error {
	_, err := os.Stat(socketPath)
	if err == nil {
		return nil
	}

	err = ioutil.WriteFile(socketPath, []byte{}, 0644)
	if err != nil {
		return err
	} else {
		os.Remove(socketPath)
	}

	return nil
}

func (server *CephDriverServerStruct) isValidTcpAddress(address string) bool {
	re := regexp.MustCompile(IPV4_REGEX)
	matches := re.FindStringSubmatch(address)
	return len(matches) > 0
}
