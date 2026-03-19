package daemon

import (
	"fmt"
	"log"
	"time"

	"github.com/Pthahnix/Vigilon/internal/config"
	"github.com/Pthahnix/Vigilon/internal/enforcer"
	"github.com/Pthahnix/Vigilon/internal/monitor"
	"github.com/Pthahnix/Vigilon/internal/notifier"
	"github.com/Pthahnix/Vigilon/internal/state"
)

type Daemon struct {
	Config   *config.Config
	Enforcer *enforcer.Enforcer
}

func New(cfg *config.Config) (*Daemon, error) {
	if err := monitor.Init(); err != nil {
		return nil, fmt.Errorf("nvml init: %w", err)
	}

	st, err := state.Load(cfg.State.Path)
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}

	n := &notifier.Notifier{
		LogPath: cfg.Notify.LogPath,
		Wall:    cfg.Notify.Wall,
	}

	grace, _ := time.ParseDuration(cfg.Daemon.GracePeriod)
	if grace == 0 {
		grace = 30 * time.Minute
	}

	e := enforcer.New(cfg, st, n, grace)

	return &Daemon{Config: cfg, Enforcer: e}, nil
}

func (d *Daemon) Run() error {
	interval, _ := time.ParseDuration(d.Config.Daemon.CheckInterval)
	if interval == 0 {
		interval = 10 * time.Minute
	}

	log.Printf("Vigilon daemon started (interval=%s, grace=%s)",
		d.Config.Daemon.CheckInterval, d.Config.Daemon.GracePeriod)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	d.cycle()
	for range ticker.C {
		d.cycle()
	}
	return nil
}

func (d *Daemon) cycle() {
	st, err := state.Load(d.Config.State.Path)
	if err != nil {
		log.Printf("reload state: %v", err)
		return
	}
	d.Enforcer.State = st

	violations, err := d.Enforcer.Check()
	if err != nil {
		log.Printf("check: %v", err)
		return
	}

	for _, v := range violations {
		d.Enforcer.Enforce(v)
	}
}
