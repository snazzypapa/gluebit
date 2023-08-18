package main

import (
	"fmt"

	"github.com/alexflint/go-arg"
)

const Version = "0.1.0"
const Description = "set listening port in qbittorrent based on gluetun connection"

// Config represents the configuration options for the gluebit program.
// It is used by "github.com/alexflint/go-arg" to parse command-line arguments
// and environment variables.
type Config struct {
	QbitUsername    string `arg:"--qbituser,env:QBITUSER" default:"" help:"qbittorrent username"`
	QbitPassword    string `arg:"--qbitpass,env:QBITPASS" default:"" help:"qbittorrent password"`
	QbitHost        string `arg:"--qbithost,env:QBITHOST" default:"localhost" help:"host to reach qbittorrent on. If this is run on the same docker network as gluetun, this can be set to the container name"`
	QbitPort        int    `arg:"--qbitport,env:QBITPORT" default:"8080" help:"port to reach qbittorrent on"`
	GlueTunHost     string `arg:"--gluetunhost,env:GLUETUNHOST" default:"localhost" help:"host to reach gluetun on. If this is run on the same docker network as gluetun, this can be set to the container name"`
	GlueTunPort     int    `arg:"--gluetunport,env:GLUETUNPORT" default:"8000" help:"port to reach gluetun on"`
	GlueTunPortFile string `arg:"--gluetunportfile,env:GLUETUNPORTFILE" default:"" help:"path to gluetun port file"`
	UpdateInterval  int    `arg:"--interval,env:GLUEBIT_INTERVAL" default:"" help:"Update interval in seconds"`
}

// Description returns a string describing the purpose of the program.
// It is used by "github.com/alexflint/go-arg" to display help text.
func (Config) Description() string {
	return fmt.Sprintf("\n%s\n", Description)
}

// Version returns the version of the program.
func (Config) Version() string {
	return fmt.Sprintf("gluebit %s\n", Version)
}

// qbitUrl returns the url to reach qbittorrent.
func (c Config) qbitUrl() string {
	return fmt.Sprintf("http://%s:%d", c.QbitHost, c.QbitPort)
}

// gluetunUrl returns the url to reach gluetun.
func (c Config) gluetunUrl() string {
	return fmt.Sprintf("http://%s:%d", c.GlueTunHost, c.GlueTunPort)
}

// loadConfig returns a Config struct.
// It loads the configuration from command-line arguments and environment variables.
func loadConfig() Config {
	var cli Config
	p := arg.MustParse(&cli)
	if (cli.GlueTunHost == "" || cli.GlueTunPort == 0) && cli.GlueTunPortFile == "" {
		p.Fail("Invalid config: must specify either --gluetunhost and --gluetunport or --gluetunportfile")
	}
	if cli.QbitHost == "" || cli.QbitPort == 0 {
		p.Fail("Invalid config: need --qbithost and --qbitport")
	}
	return cli
}
