package main

import (
	"fmt"
	"os/exec"

	"github.com/Pthahnix/Vigilon/internal/config"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Vigilon daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		if err := verifyPassphrase(cfg); err != nil {
			return err
		}

		out, err := exec.Command("sudo", "systemctl", "start", "vigilon").CombinedOutput()
		if err != nil {
			return fmt.Errorf("start failed: %s", out)
		}
		fmt.Println("Vigilon daemon started.")
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Vigilon daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		if err := verifyPassphrase(cfg); err != nil {
			return err
		}

		out, err := exec.Command("sudo", "systemctl", "stop", "vigilon").CombinedOutput()
		if err != nil {
			return fmt.Errorf("stop failed: %s", out)
		}
		fmt.Println("Vigilon daemon stopped.")
		return nil
	},
}

func init() {
	adminCmd.AddCommand(startCmd)
	adminCmd.AddCommand(stopCmd)
}
