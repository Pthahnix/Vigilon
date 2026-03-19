package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "vigil",
	Short: "Vigilon - AI-powered GPU watchdog",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/etc/vigilon/config.yaml", "config file path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
