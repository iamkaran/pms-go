package broker

import mqtt "github.com/mochi-mqtt/server/v2"

type ReadyHook struct {
	mqtt.HookBase
	ready chan struct{}
}

func (h *ReadyHook) ID() string           { return "ready-hook" }
func (h *ReadyHook) Provides(b byte) bool { return b == mqtt.OnStarted }
func (h *ReadyHook) OnStarted()           { close(h.ready) }
