package notifier

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Notifier struct {
	LogPath string
	Wall    bool
}

func (n *Notifier) Log(event, user, msg string) {
	line := fmt.Sprintf("[%s] event=%s user=%s %s\n",
		time.Now().Format(time.RFC3339), event, user, msg)
	logFile := filepath.Join(n.LogPath, "audit.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "vigilon: write log: %v\n", err)
		return
	}
	defer f.Close()
	f.WriteString(line)
}

func (n *Notifier) Warn(user, msg string) {
	n.Log("warning", user, msg)
	if n.Wall {
		wallMsg := fmt.Sprintf("[Vigilon] @%s: %s", user, msg)
		exec.Command("wall", wallMsg).Run()
	}
}

func (n *Notifier) Kill(user, msg string) {
	n.Log("kill", user, msg)
	if n.Wall {
		wallMsg := fmt.Sprintf("[Vigilon] @%s: %s", user, msg)
		exec.Command("wall", wallMsg).Run()
	}
}
