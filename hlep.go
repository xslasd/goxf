package goxf

import (
	"fmt"
	"os"

	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/flag"
)

func parseFlags(appID, configFile string) error {
	flag.Register(&flag.BoolFlag{
		Name:  "help",
		Usage: "--help, show help information",
		Action: func(name string, fs *flag.FlagSet) {
			fs.PrintDefaults()
			os.Exit(0)
		},
	})
	flag.Register(&flag.StringFlag{
		Name:    "config",
		Usage:   "--config",
		EnvVar:  "GOXF_CONFIG",
		Default: configFile,
	})

	flag.Register(&flag.BoolFlag{
		Name:    "watch",
		Usage:   "--watch, watch config change event",
		Default: false,
		EnvVar:  "GOXF_CONFIG_WATCH",
	})
	flag.Register(&flag.BoolFlag{
		Name:    "version",
		Usage:   "--version,show app info",
		Default: false,
		Action: func(name string, fs *flag.FlagSet) {
			fmt.Printf("appId: %s", appID)
			application.PrintVersion()
			os.Exit(0)
		},
	})
	flag.Register(&flag.BoolFlag{
		Name:    "crypt-conf",
		Usage:   "--crypt-conf, input password to generate encrypt/decrypt config file",
		Default: false,
	})
	return flag.Parse()
}
