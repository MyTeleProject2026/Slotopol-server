package cmd

import (
	"fmt"
	"os"

	"github.com/MyTeleProject2026/Slotopol-server/config"

	"github.com/spf13/cobra"
)

const rootShort = "Slots games backend"
const rootLong = `This application implements web server and reels scanner for slots games.`

var (
	rootCmd = &cobra.Command{
		Use:     config.AppName,
		Version: config.BuildVers,
		Short:   rootShort,
		Long:    rootLong,
	}
)

func init() {
	cobra.OnInitialize(config.InitConfig)

	var pf = rootCmd.PersistentFlags()
	pf.StringVarP(&config.CfgFile, "config", "c", "", "config file (default is config/slot-app.yaml at executable location)")
	pf.StringVarP(&config.SqlPath, "sqlite", "q", "", "sqlite databases path (default same as config file path)")
	pf.StringArrayVarP(&config.ObjPath, "fpath", "f", nil, "additional paths to yaml files or folders with game specific data (can be repeated)")
	pf.BoolVarP(&config.Verbose, "verbose", "v", false, "print more verbose information to log")
	rootCmd.SetVersionTemplate(fmt.Sprintf("version: %s, builton: %s", config.BuildVers, config.BuildTime))
}

// Execute executes the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
