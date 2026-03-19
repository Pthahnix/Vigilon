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

		u, err := user.Current()
		if err != nil {
			return fmt.Errorf("get current user: %v", err)
		}
		username := u.Username

		var oldPriority string
		err = state.LoadAndModify(cfg.State.Path, func(st *state.State) error {
			prev, ok := st.Users[username]
			if !ok || prev.Priority == "P0" || prev.Priority == "" {
				return nil
			}
			oldPriority = prev.Priority
			delete(st.Users, username)
			return nil
		})
		if err != nil {
			return fmt.Errorf("update state: %v", err)
		}
		if oldPriority == "" {
			fmt.Println("You are already at P0. Nothing to release.")
			return nil
		}

		n := &notifier.Notifier{LogPath: cfg.Notify.LogPath, Wall: false}
		n.Log("release", username, fmt.Sprintf("manually released %s → P0", oldPriority))

		fmt.Printf("Released %s → P0. Thank you for freeing up resources.\n", oldPriority)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
}
