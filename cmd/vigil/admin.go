package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Pthahnix/Vigilon/internal/config"
	"github.com/Pthahnix/Vigilon/internal/daemon"
	"github.com/Pthahnix/Vigilon/internal/notifier"
	"github.com/Pthahnix/Vigilon/internal/reviewer"
	"github.com/Pthahnix/Vigilon/internal/state"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Admin commands (requires passphrase)",
}

func loadPassphrase(cfg *config.Config) string {
	if cfg.LLM.EnvFile != "" {
		reviewer.LoadEnvFile(cfg.LLM.EnvFile)
	}
	return os.Getenv("ADMIN_PASSPHRASE")
}

func verifyPassphrase(cfg *config.Config) error {
	expected := loadPassphrase(cfg)
	if expected == "" {
		return fmt.Errorf("ADMIN_PASSPHRASE not set in .env")
	}
	fmt.Print("Enter admin passphrase: ")
	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		// fallback for non-terminal (e.g. pipe)
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			raw = []byte(scanner.Text())
		}
	}
	if strings.TrimSpace(string(raw)) != expected {
		return fmt.Errorf("invalid passphrase")
	}
	return nil
}

var grantCmd = &cobra.Command{
	Use:   "grant <user> <P0|P1|P2> [duration]",
	Short: "Manually set a user's priority",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		if err := verifyPassphrase(cfg); err != nil {
			return err
		}

		user := args[0]
		priority := strings.ToUpper(args[1])
		if _, ok := cfg.Priority[priority]; !ok {
			return fmt.Errorf("invalid priority: %s (use P0, P1, P2)", priority)
		}

		dur := 24 * time.Hour
		if len(args) == 3 {
			d, err := time.ParseDuration(args[2])
			if err != nil {
				return fmt.Errorf("invalid duration: %v", err)
			}
			dur = d
		}
		expires := time.Now().Add(dur)

		if err := state.LoadAndModify(cfg.State.Path, func(st *state.State) error {
			st.Users[user] = &state.UserState{
				Priority: priority,
				GPUs:     nil,
				Expires:  &expires,
			}
			return nil
		}); err != nil {
			return fmt.Errorf("save state: %v", err)
		}

		n := &notifier.Notifier{LogPath: cfg.Notify.LogPath, Wall: false}
		n.Log("admin-grant", user, fmt.Sprintf("priority=%s expires=%s", priority, expires.Format("2006-01-02 15:04")))

		fmt.Printf("Granted %s → %s (expires %s)\n", user, priority, expires.Format("2006-01-02 15:04"))
		return nil
	},
}

var resetCmd = &cobra.Command{
	Use:   "reset <user>",
	Short: "Reset a user back to P0",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		if err := verifyPassphrase(cfg); err != nil {
			return err
		}

		user := args[0]
		if err := state.LoadAndModify(cfg.State.Path, func(st *state.State) error {
			delete(st.Users, user)
			return nil
		}); err != nil {
			return fmt.Errorf("save state: %v", err)
		}

		n := &notifier.Notifier{LogPath: cfg.Notify.LogPath, Wall: false}
		n.Log("admin-reset", user, "reset to P0")

		fmt.Printf("Reset %s → P0\n", user)
		return nil
	},
}

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Reset ALL users to P0",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		if err := verifyPassphrase(cfg); err != nil {
			return err
		}

		st := &state.State{Users: make(map[string]*state.UserState)}
		if err := state.Save(cfg.State.Path, st); err != nil {
			return fmt.Errorf("save state: %w", err)
		}

		n := &notifier.Notifier{LogPath: cfg.Notify.LogPath, Wall: false}
		n.Log("admin-purge", "admin", "all users reset to P0")

		fmt.Println("All users reset to P0")
		return nil
	},
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run one detection cycle immediately",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return err
		}
		if err := verifyPassphrase(cfg); err != nil {
			return err
		}

		d, err := daemon.New(cfg)
		if err != nil {
			return fmt.Errorf("init: %v", err)
		}
		fmt.Println("Running detection cycle...")
		d.Cycle()
		fmt.Println("Done.")
		return nil
	},
}

func init() {
	adminCmd.AddCommand(grantCmd)
	adminCmd.AddCommand(resetCmd)
	adminCmd.AddCommand(purgeCmd)
	adminCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(adminCmd)
}
