package vanity

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator("test", "end", false, 1)
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if g.pattern != "test" {
		t.Errorf("expected pattern 'test', got '%s'", g.pattern)
	}
}

func TestGenerateAddress(t *testing.T) {
	g := NewGenerator("test", "end", false, 1)

	// Generate a sample public key for testing
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		t.Fatalf("failed to generate private key: %v", err)
	}
	pubKey := privateKey.PubKey().SerializeCompressed()

	address := g.generateAddress(pubKey)
	if address == "" {
		t.Error("generateAddress returned empty string")
	}
	if !strings.HasPrefix(address, "init1") {
		t.Errorf("address does not start with init1: %s", address)
	}
}

func TestIsMatch(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		position      string
		caseSensitive bool
		address       string
		want          bool
	}{
		{
			name:     "end match",
			pattern:  "test",
			position: "end",
			address:  "init1abctest",
			want:     true,
		},
		{
			name:     "start match",
			pattern:  "test",
			position: "start",
			address:  "init1testabc",
			want:     true,
		},
		{
			name:     "any match",
			pattern:  "test",
			position: "any",
			address:  "init1abctestxyz",
			want:     true,
		},
		{
			name:     "no match",
			pattern:  "test",
			position: "end",
			address:  "init1abcxyz",
			want:     false,
		},
		{
			name:          "case sensitive match",
			pattern:       "Test",
			position:      "end",
			caseSensitive: true,
			address:       "init1abcTest",
			want:          true,
		},
		{
			name:          "case sensitive no match",
			pattern:       "Test",
			position:      "end",
			caseSensitive: true,
			address:       "init1abctest",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.pattern, tt.position, tt.caseSensitive, 1)
			got := g.isMatch(tt.address)
			if got != tt.want {
				t.Errorf("isMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		position string
		count    int
		threads  int
		timeout  time.Duration
	}{
		{
			name:     "single result",
			pattern:  "a",
			position: "end",
			count:    1,
			threads:  2,
			timeout:  10 * time.Second,
		},
		{
			name:     "multiple results",
			pattern:  "a",
			position: "any",
			count:    3,
			threads:  4,
			timeout:  20 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.pattern, tt.position, false, tt.count)

			err := g.Generate(tt.threads)
			if err != nil {
				t.Errorf("Generate() error = %v", err)
			}

			results := g.GetResults()
			if len(results) != tt.count {
				t.Errorf("expected %d results, got %d", tt.count, len(results))
			}

			for _, result := range results {
				// Verify each result
				if !g.isMatch(result.Address) {
					t.Errorf("generated address does not match pattern: %s", result.Address)
				}
				if !strings.HasPrefix(result.Address, "init1") {
					t.Errorf("address does not start with init1: %s", result.Address)
				}
			}

			stats := g.GetStats()
			if stats.Found != uint64(tt.count) {
				t.Errorf("stats.Found = %d, want %d", stats.Found, tt.count)
			}
		})
	}
}

func TestGetStats(t *testing.T) {
	g := NewGenerator("test", "end", false, 1)
	atomic.StoreUint64(&g.stats.Attempts, 100)
	atomic.StoreUint64(&g.stats.Found, 2)

	stats := g.GetStats()
	if stats.Attempts != 100 {
		t.Errorf("expected 100 attempts, got %d", stats.Attempts)
	}
	if stats.Found != 2 {
		t.Errorf("expected 2 found, got %d", stats.Found)
	}
}

func TestStop(t *testing.T) {
	g := NewGenerator("a", "end", false, 1000) // Large count to ensure it doesn't finish naturally
	done := make(chan bool)

	go func() {
		g.Generate(2)
		done <- true
	}()

	// Let it run for a bit
	time.Sleep(100 * time.Millisecond)

	// Stop the generator
	g.Stop()

	select {
	case <-done:
		// Success - generator stopped
	case <-time.After(2 * time.Second):
		t.Fatal("generator did not stop after Stop() called")
	}
}
