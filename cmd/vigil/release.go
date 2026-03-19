package main

import (
	"fmt"
	"os/user"

	"github.com/Pthahnix/Vigilon/internal/config"
	"github.com/Pthahnix/Vigilon/internal/notifier"
	"github.com/Pthahnix/Vigilon/internal/state"
	"github.com/spf13/cobra"
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release your GPU priority back to P0",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}

		u, _ := user.Current()
		username := u.Username

		st, _ := state.Load(cfg.State.Path)
		prev, ok := st.Users[username]
		if !ok || prev.Priority == "P0" || prev.Priority == "" {
			fmt.Println("You are already at P0. Nothing to release.")
			return nil
		}

		oldPriority := prev.Priority
		delete(st.Users, username)
		state.Save(cfg.State.Path, st)

		n := &notifier.Notifier{LogPath: cfg.Notify.LogPath, Wall: false}
		n.Log("release", username, fmt.Sprintf("manually released %s → P0", oldPriority))

		fmt.Printf("Released %s → P0. Thank you for freeing up resources.\n", oldPriority)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}
