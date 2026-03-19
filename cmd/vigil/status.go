package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/Pthahnix/Vigilon/internal/config"
	"github.com/Pthahnix/Vigilon/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show all users' GPU priority and usage",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		st, err := state.Load(cfg.State.Path)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "USER\tPRIORITY\tGPUs\tEXPIRES")
		fmt.Fprintln(w, "----\t--------\t----\t-------")
		for user, u := range st.Users {
			exp := "-"
			if u.Expires != nil {
				exp = u.Expires.Format("2006-01-02 15:04")
			}
			fmt.Fprintf(w, "%s\t%s\t%v\t%s\n", user, u.Priority, u.GPUs, exp)
		}
		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
