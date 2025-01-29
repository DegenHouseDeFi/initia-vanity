package config

import (
	"runtime"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Position != "end" {
		t.Errorf("expected default position 'end', got '%s'", cfg.Position)
	}
	if cfg.Threads != runtime.NumCPU() {
		t.Errorf("expected default threads %d, got %d", runtime.NumCPU(), cfg.Threads)
	}
	if cfg.CaseSensitive {
		t.Error("expected case-sensitive to be false by default")
	}
	if cfg.Format != "text" {
		t.Errorf("expected default format 'text', got '%s'", cfg.Format)
	}
	if cfg.Count != 1 {
		t.Errorf("expected default count 1, got %d", cfg.Count)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Pattern:  "test",
				Position: "end",
				Threads:  1,
				Format:   "text",
				Count:    1,
			},
			wantErr: false,
		},
		{
			name: "invalid position",
			config: &Config{
				Pattern:  "test",
				Position: "invalid",
				Threads:  1,
				Format:   "text",
				Count:    1,
			},
			wantErr: true,
		},
		{
			name: "empty pattern",
			config: &Config{
				Pattern:  "",
				Position: "end",
				Threads:  1,
				Format:   "text",
				Count:    1,
			},
			wantErr: true,
		},
		{
			name: "invalid threads",
			config: &Config{
				Pattern:  "test",
				Position: "end",
				Threads:  0,
				Format:   "text",
				Count:    1,
			},
			wantErr: true,
		},
		{
			name: "invalid count",
			config: &Config{
				Pattern:  "test",
				Position: "end",
				Threads:  1,
				Format:   "text",
				Count:    0,
			},
			wantErr: true,
		},
		{
			name: "invalid format",
			config: &Config{
				Pattern:  "test",
				Position: "end",
				Threads:  1,
				Format:   "invalid",
				Count:    1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
