package ws

import "testing"

func TestConfigureUpgraderAppliesConnectionDefaults(t *testing.T) {
	previous := currentConnectionConfig()
	t.Cleanup(func() { ConfigureUpgrader(previous) })

	ConfigureUpgrader(Config{AllowedOrigins: []string{"https://app.example"}})
	cfg := currentConnectionConfig()
	if cfg.MaxMessageSize <= 0 || cfg.PingInterval <= 0 || cfg.PongTimeout <= cfg.PingInterval {
		t.Fatalf("invalid connection defaults: %+v", cfg)
	}
	if Upgrader.ReadBufferSize <= 0 || Upgrader.WriteBufferSize <= 0 {
		t.Fatal("upgrader buffers were not initialized")
	}
}
