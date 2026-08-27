package goxf

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/flag"
)

func parseFlags(appID, configFile string) error {
	// 注册短参数与全量参数标志 (-h, -c, -w, -v, -e)
	flag.Register(
		&flag.BoolFlag{
			Name:  "help,h",
			Usage: "show help information",
			Action: func(name string, fs *flag.FlagSet) {
				fs.PrintDefaults()
				os.Exit(0)
			},
		},
		&flag.StringFlag{
			Name:    "config,c",
			Usage:   "path to application configuration file",
			EnvVar:  "GOXF_CONFIG",
			Default: configFile,
		},
		&flag.BoolFlag{
			Name:    "watch,w",
			Usage:   "watch configuration file changes dynamically",
			Default: false,
			EnvVar:  "GOXF_CONFIG_WATCH",
		},
		&flag.BoolFlag{
			Name:    "version,v",
			Usage:   "show application version information",
			Default: false,
			Action: func(name string, fs *flag.FlagSet) {
				fmt.Printf("appId: %s\n", appID)
				application.PrintVersion()
				os.Exit(0)
			},
		},
		&flag.BoolFlag{
			Name:    "crypt-conf,e",
			Usage:   "input password to generate encrypted config file (system.enc)",
			Default: false,
		},
	)

	// 注册短子命令体系 (Subcommands: version/v, help/h)
	flag.AddCommand(&cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "show application version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("appId: %s\n", appID)
			application.PrintVersion()
			os.Exit(0)
		},
	})

	// 快速命中短子命令 (如: app v, app version, app help)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "v":
			fmt.Printf("appId: %s\n", appID)
			application.PrintVersion()
			os.Exit(0)
		case "help":
			flag.PrintDefaults()
			os.Exit(0)
		}
	}

	return flag.Parse()
}
