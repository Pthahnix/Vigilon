package main

import (
	"log"

	"github.com/Pthahnix/Vigilon/internal/config"
	"github.com/Pthahnix/Vigilon/internal/daemon"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the GPU watchdog daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			log.Fatalf("load config: %v", err)
		}
		d, err := daemon.New(cfg)
		if err != nil {
			log.Fatalf("init daemon: %v", err)
		}
		return d.Run()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
