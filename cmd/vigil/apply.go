package main

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"time"

	"github.com/Pthahnix/Vigilon/internal/config"
	"github.com/Pthahnix/Vigilon/internal/monitor"
	"github.com/Pthahnix/Vigilon/internal/notifier"
	"github.com/Pthahnix/Vigilon/internal/reviewer"
	"github.com/Pthahnix/Vigilon/internal/state"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply <file.md>",
	Short: "Submit a GPU priority upgrade request",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("load config: %v", err)
		}

		content, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("read application: %v", err)
		}

		r := reviewer.New(&cfg.LLM)
		result, err := r.Review(string(content))
		if err != nil {
			return fmt.Errorf("AI review failed: %v", err)
		}

		u, err := user.Current()
		if err != nil {
			return fmt.Errorf("get current user: %v", err)
		}
		username := u.Username

		dur, err := time.ParseDuration(result.Duration)
		if err != nil {
			dur = 24 * time.Hour
		}
		// Apply duration buffer to estimated duration
		dur = time.Duration(float64(dur) * cfg.Daemon.DurationBuffer)
		expires := time.Now().Add(dur)

		if err := monitor.Init(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: GPU monitor init failed: %v (using fallback assignment)\n", err)
		} else {
			defer monitor.Shutdown()
		}
		tier := cfg.Priority[result.Priority]
		gpus := assignGPUs(tier.MaxGPUs)

		if err := state.LoadAndModify(cfg.State.Path, func(st *state.State) error {
			st.Users[username] = &state.UserState{
				Priority: result.Priority,
				GPUs:     gpus,
				Expires:  &expires,
			}
			return nil
		}); err != nil {
			return fmt.Errorf("save state: %v", err)
		}

		n := &notifier.Notifier{LogPath: cfg.Notify.LogPath, Wall: false}
		n.Log("grant", username, fmt.Sprintf("priority=%s gpus=%v expires=%s reason=%s",
			result.Priority, gpus, expires.Format("2006-01-02 15:04"), result.Reason))

		gpuStr := formatGPUs(gpus)
		fmt.Printf("\n✅ Priority upgraded: P0 → %s\n", result.Priority)
		fmt.Printf("   Assigned GPUs: %s\n", gpuStr)
		fmt.Printf("   Expires: %s\n", expires.Format("2006-01-02 15:04"))
		fmt.Printf("   Run: export CUDA_VISIBLE_DEVICES=%s\n", gpuNums(gpus))
		fmt.Printf("   Reason: %s\n\n", result.Reason)

		grantPath := filepath.Join(u.HomeDir, "vigilon_grant.md")
		grantContent := fmt.Sprintf(`# Vigilon GPU Grant
- Priority: %s
- GPUs: %s
- Granted: %s
- Expires: %s
- Reason: %s
`, result.Priority, gpuStr,
			time.Now().Format("2006-01-02 15:04"),
			expires.Format("2006-01-02 15:04"),
			result.Reason)
		if err := os.WriteFile(grantPath, []byte(grantContent), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not write grant file: %v\n", err)
		}

		return nil
	},
}

func assignGPUs(count int) []int {
	gpus, err := monitor.ListGPUs()
	if err != nil || len(gpus) == 0 {
		fallback := make([]int, count)
		for i := range fallback {
			fallback[i] = i
		}
		return fallback
	}
	sort.Slice(gpus, func(i, j int) bool {
		return gpus[i].MemUsed < gpus[j].MemUsed
	})
	result := make([]int, 0, count)
	for i := 0; i < count && i < len(gpus); i++ {
		result = append(result, gpus[i].Index)
	}
	sort.Ints(result)
	return result
}

func formatGPUs(gpus []int) string {
	s := ""
	for i, g := range gpus {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("CUDA:%d", g)
	}
	return s
}

func gpuNums(gpus []int) string {
	s := ""
	for i, g := range gpus {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf("%d", g)
	}
	return s
}

func init() {
	rootCmd.AddCommand(applyCmd)
}
