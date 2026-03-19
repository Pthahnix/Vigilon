package main

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Pthahnix/Vigilon/internal/config"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show your Vigilon history",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		u, _ := user.Current()
		logFile := filepath.Join(cfg.Notify.LogPath, "audit.log")
		f, err := os.Open(logFile)
		if err != nil {
			fmt.Println("No log history found.")
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "user="+u.Username) {
				fmt.Println(line)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logCmd)
}
