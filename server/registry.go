package server

import (
	"fmt"
	"sync"

	"github.com/tianhongw/grp/conf"
	"github.com/tianhongw/grp/pkg/log"
)

type TunnelRegistry struct {
	mu      sync.Mutex
	tunnels map[string]*Tunnel

	lg log.Logger
}

func newTunnelRegistry(cfg *conf.Config) *TunnelRegistry {
	tr := &TunnelRegistry{
		tunnels: map[string]*Tunnel{},
	}

	lg, err := log.NewLogger(cfg.Log.Type,
		log.WithLevel(cfg.Log.Level),
		log.WithPrefix("tunnel-registry"))
	if err != nil {
		panic("new logger for tunnel registry failed")
	}

	tr.lg = lg

	return tr
}

func (tr *TunnelRegistry) Get(url string) *Tunnel {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	return tr.tunnels[url]
}

func (tr *TunnelRegistry) Register(t *Tunnel, url string) error {
	tr.mu.Lock()
	defer tr.mu.Unlock()

	if tr.tunnels[url] != nil {
		return fmt.Errorf("tunnel: %s is already registered", url)
	}

	tr.tunnels[url] = t

	return nil
}

type ControlRegistry struct {
	mu       sync.Mutex
	controls map[string]*Control

	lg log.Logger
}

func newControlRegistry(cfg *conf.Config) *ControlRegistry {
	cr := &ControlRegistry{
		controls: make(map[string]*Control),
	}

	lg, err := log.NewLogger(cfg.Log.Type,
		log.WithLevel(cfg.Log.Level),
		log.WithPrefix("control-registry"))
	if err != nil {
		panic("new logger for control registry failed")
	}

	cr.lg = lg

	return cr
}

func (cr *ControlRegistry) Get(clientId string) *Control {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	return cr.controls[clientId]
}

func (cr *ControlRegistry) Add(clientId string, ctl *Control) (oldCtl *Control) {
	cr.mu.Lock()
	oldCtl = cr.controls[clientId]
	cr.mu.Unlock()

	if oldCtl != nil {
		oldCtl.Replace(ctl)
	}

	cr.mu.Lock()
	cr.controls[clientId] = ctl
	cr.mu.Unlock()
	cr.lg.Infof("add control with client id: %s", clientId)
	return
}

func (cr *ControlRegistry) Remove(clientId string) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.controls[clientId] == nil {
		cr.lg.Errorf("remove control failed, no control find for client: %s", clientId)
		return fmt.Errorf("no control find for client: %s", clientId)
	} else {
		cr.lg.Infof("remove control for client: %s success", clientId)
		delete(cr.controls, clientId)
	}

	return nil
}

func (cr *ControlRegistry) exit() {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	for _, ctl := range cr.controls {
		ctl.exit()
	}

	return
}
