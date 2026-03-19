package enforcer

import (
	"fmt"
	"os"
	"sort"
	"syscall"
	"time"

	"github.com/Pthahnix/Vigilon/internal/config"
	"github.com/Pthahnix/Vigilon/internal/monitor"
	"github.com/Pthahnix/Vigilon/internal/notifier"
	"github.com/Pthahnix/Vigilon/internal/state"
)

type Enforcer struct {
	Config        *config.Config
	State         *state.State
	Notifier      *notifier.Notifier
	Grace         map[string]time.Time
	GraceDuration time.Duration
}

func New(cfg *config.Config, st *state.State, n *notifier.Notifier, grace time.Duration) *Enforcer {
	return &Enforcer{
		Config:        cfg,
		State:         st,
		Notifier:      n,
		Grace:         make(map[string]time.Time),
		GraceDuration: grace,
	}
}

type Violation struct {
	User       string
	Priority   string
	Allowed    int
	Actual     int
	ExcessPIDs []uint32
}

func (e *Enforcer) Check() ([]Violation, error) {
	userGPUs, err := monitor.UserGPUMap()
	if err != nil {
		return nil, err
	}
	procs, err := monitor.ListProcesses()
	if err != nil {
		return nil, err
	}

	var violations []Violation
	for user, gpus := range userGPUs {
		priority := e.State.GetPriority(user)
		tier, ok := e.Config.Priority[priority]
		if !ok {
			tier = config.PriorityTier{MaxGPUs: 1}
		}
		if len(gpus) <= tier.MaxGPUs {
			delete(e.Grace, user)
			continue
		}
		allowed := make(map[int]bool)
		for i := 0; i < tier.MaxGPUs && i < len(gpus); i++ {
			allowed[gpus[i]] = true
		}
		var excess []monitor.GPUProcess
		for _, p := range procs {
			if p.User == user && !allowed[p.GPU] {
				excess = append(excess, p)
			}
		}
		sort.Slice(excess, func(i, j int) bool {
			return excess[i].PID > excess[j].PID
		})
		pids := make([]uint32, len(excess))
		for i, p := range excess {
			pids[i] = p.PID
		}
		violations = append(violations, Violation{
			User:       user,
			Priority:   priority,
			Allowed:    tier.MaxGPUs,
			Actual:     len(gpus),
			ExcessPIDs: pids,
		})
	}
	return violations, nil
}

func (e *Enforcer) Enforce(v Violation) {
	first, inGrace := e.Grace[v.User]
	if !inGrace {
		e.Grace[v.User] = time.Now()
		msg := fmt.Sprintf("Using %d GPUs but %s allows %d. You have %s to comply.",
			v.Actual, v.Priority, v.Allowed, e.GraceDuration)
		e.Notifier.Warn(v.User, msg)
		return
	}
	if time.Since(first) < e.GraceDuration {
		return
	}
	for _, pid := range v.ExcessPIDs {
		proc, err := os.FindProcess(int(pid))
		if err != nil {
			continue
		}
		proc.Signal(syscall.SIGTERM)
		msg := fmt.Sprintf("Killed PID %d (exceeded %s limit: %d/%d GPUs)",
			pid, v.Priority, v.Actual, v.Allowed)
		e.Notifier.Kill(v.User, msg)
	}
	delete(e.Grace, v.User)
}

// CheckIdle detects users with elevated priority but no GPU activity.
// - 3 consecutive idle cycles → auto-downgrade to P0
// - Expired but still actively using GPU → tolerate, don't kill
func (e *Enforcer) CheckIdle(statePath string) {
	userGPUs, err := monitor.UserGPUMap()
	if err != nil {
		return
	}

	for user, u := range e.State.Users {
		if u.Priority == "P0" || u.Priority == "" {
			continue
		}

		_, hasGPU := userGPUs[user]
		expired := u.Expires != nil && time.Now().After(*u.Expires)

		if !hasGPU {
			// User has elevated priority but no GPU processes
			u.IdleCount++
			if u.IdleCount >= 3 {
				// 3 consecutive idle checks → auto-reclaim
				e.Notifier.Log("auto-reclaim", user,
					fmt.Sprintf("idle for %d cycles, downgraded from %s to P0", u.IdleCount, u.Priority))
				u.Priority = "P0"
				u.GPUs = nil
				u.Expires = nil
				u.IdleCount = 0
			}
		} else {
			// User is actively using GPU
			u.IdleCount = 0
			if expired {
				// Expired but still running → tolerate, just log
				e.Notifier.Log("overtime-tolerate", user,
					fmt.Sprintf("expired but GPU still active, tolerating %s", u.Priority))
			}
		}
	}

	// Save updated idle counts
	state.Save(statePath, e.State)
}
