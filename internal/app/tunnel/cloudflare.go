package tunnel

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/talkincode/toughradius/v9/config"
	"go.uber.org/zap"
)

type CloudflareTunnel struct {
	cfg       config.TunnelConfig
	cmd       *exec.Cmd
	status    *TunnelStatus
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

func NewCloudflareTunnel(cfg config.TunnelConfig) *CloudflareTunnel {
	return &CloudflareTunnel{
		cfg: cfg,
		status: &TunnelStatus{
			Type:   "cloudflare",
			Status: "stopped",
		},
	}
}

func (t *CloudflareTunnel) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cfg.Token == "" {
		return fmt.Errorf("cloudflare tunnel token is required")
	}

	ctx, cancel := context.WithCancel(context.Background())
	t.cancel = cancel

	// Start cloudflared via command line
	// Assuming cloudflared is installed and in PATH
	// Or we could embed it, but that's complex. Shell execution is simpler for MVP.
	t.cmd = exec.CommandContext(ctx, "cloudflared", "tunnel", "run", "--token", t.cfg.Token)
	
	if err := t.cmd.Start(); err != nil {
		t.status.Status = "error"
		t.status.Error = err.Error()
		return fmt.Errorf("failed to start cloudflared: %v", err)
	}

	t.status.Status = "running"
	t.status.StartedAt = time.Now()
	t.status.ID = "cloudflare-" + time.Now().Format("20060102150405")

	go t.monitor()

	return nil
}

func (t *CloudflareTunnel) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancel != nil {
		t.cancel()
	}

	if t.cmd != nil {
		_ = t.cmd.Wait()
	}

	t.status.Status = "stopped"
	return nil
}

func (t *CloudflareTunnel) Status() (*TunnelStatus, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.status, nil
}

func (t *CloudflareTunnel) Type() string {
	return "cloudflare"
}

func (t *CloudflareTunnel) monitor() {
	if t.cmd == nil {
		return
	}
	
	err := t.cmd.Wait()
	
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if err != nil {
		zap.S().Errorf("cloudflared exited with error: %v", err)
		t.status.Status = "error"
		t.status.Error = err.Error()
	} else {
		zap.S().Info("cloudflared exited normally")
		t.status.Status = "stopped"
	}
}
