package vanity

import (
	"encoding/json"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cosmos/go-bip39"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator("test", "end", false, 1, false, "")
	if g == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if g.pattern != "test" {
		t.Errorf("expected pattern 'test', got '%s'", g.pattern)
	}
}

func TestGenerateAddress(t *testing.T) {
	g := NewGenerator("test", "end", false, 1, false, "")

	addr, privKey, pubKey, err := g.generateAddress()
	if err != nil {
		t.Fatalf("generateAddress error: %v", err)
	}

	// Check address format
	if addr == "" {
		t.Error("address is empty")
	}
	if !strings.HasPrefix(addr, "init1") {
		t.Errorf("address does not start with init1: %s", addr)
	}

	// Check private key format (should be hex-encoded 32 bytes)
	if len(privKey) != 64 {
		t.Errorf("private key length should be 64 chars, got %d", len(privKey))
	}

	// Check public key format (should be JSON with @type and key fields)
	var pubKeyJSON map[string]interface{}
	if err := json.Unmarshal([]byte(pubKey), &pubKeyJSON); err != nil {
		t.Errorf("invalid public key JSON: %v", err)
	}
	if pubKeyJSON["@type"] != "/cosmos.crypto.secp256k1.PubKey" {
		t.Errorf("unexpected public key type: %v", pubKeyJSON["@type"])
	}
	if _, ok := pubKeyJSON["key"].(string); !ok {
		t.Error("public key missing 'key' field")
	}
}

func TestGenerateAddressFromMnemonic(t *testing.T) {
	tests := []struct {
		name         string
		mnemonic     string
		expectError  bool
		checkAddress bool
	}{
		{
			name:         "generate with new mnemonic",
			mnemonic:     "",
			expectError:  false,
			checkAddress: true,
		},
		{
			name:         "use provided valid mnemonic",
			mnemonic:     "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			expectError:  false,
			checkAddress: true,
		},
		{
			name:         "invalid mnemonic",
			mnemonic:     "invalid mnemonic phrase",
			expectError:  true,
			checkAddress: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator("test", "end", false, 1, true, tt.mnemonic)
			addr, privKey, pubKey, mnemonic, path, err := g.generateAddressFromMnemonic()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkAddress {
				// Check address format
				if !strings.HasPrefix(addr, "init1") {
					t.Errorf("address does not start with init1: %s", addr)
				}

				// Verify mnemonic is valid
				if !bip39.IsMnemonicValid(mnemonic) {
					t.Error("generated mnemonic is invalid")
				}

				// Check derivation path is always index 0
				expectedPath := "m/44'/118'/0'/0/0"
				if path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, path)
				}

				// Check private key format
				if len(privKey) != 64 {
					t.Errorf("private key length should be 64 chars, got %d", len(privKey))
				}

				// Check public key format
				var pubKeyJSON map[string]interface{}
				if err := json.Unmarshal([]byte(pubKey), &pubKeyJSON); err != nil {
					t.Errorf("invalid public key JSON: %v", err)
				}
			}
		})
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
			g := NewGenerator(tt.pattern, tt.position, tt.caseSensitive, 1, false, "")
			got := g.isMatch(tt.address)
			if got != tt.want {
				t.Errorf("isMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		position    string
		count       int
		threads     int
		timeout     time.Duration
		useMnemonic bool
		mnemonic    string
	}{
		{
			name:        "single result without mnemonic",
			pattern:     "a",
			position:    "end",
			count:       1,
			threads:     2,
			timeout:     10 * time.Second,
			useMnemonic: false,
		},
		{
			name:        "multiple results without mnemonic",
			pattern:     "a",
			position:    "any",
			count:       3,
			threads:     4,
			timeout:     20 * time.Second,
			useMnemonic: false,
		},
		{
			name:        "single result with mnemonic",
			pattern:     "a",
			position:    "end",
			count:       1,
			threads:     2,
			timeout:     10 * time.Second,
			useMnemonic: true,
		},
		{
			name:        "multiple results with mnemonic",
			pattern:     "a",
			position:    "any",
			count:       3,
			threads:     4,
			timeout:     20 * time.Second,
			useMnemonic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.pattern, tt.position, false, tt.count, tt.useMnemonic, tt.mnemonic)

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
				// Verify public key format
				var pubKeyJSON map[string]interface{}
				if err := json.Unmarshal([]byte(result.PublicKey), &pubKeyJSON); err != nil {
					t.Errorf("invalid public key JSON: %v", err)
				}
				// Verify private key length (should be 64 chars hex)
				if len(result.PrivateKey) != 64 {
					t.Errorf("private key length should be 64 chars, got %d", len(result.PrivateKey))
				}

				// Verify mnemonic-specific fields when using mnemonic
				if tt.useMnemonic {
					if result.Mnemonic == "" {
						t.Error("mnemonic is empty when using mnemonic generation")
					}
					if !bip39.IsMnemonicValid(result.Mnemonic) {
						t.Error("invalid mnemonic generated")
					}
					if result.DerivationPath != "m/44'/118'/0'/0/0" {
						t.Errorf("incorrect derivation path, expected m/44'/118'/0'/0/0, got %s", result.DerivationPath)
					}
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
	g := NewGenerator("test", "end", false, 1, false, "")
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
	g := NewGenerator("a", "end", false, 1000, false, "") // Large count to ensure it doesn't finish naturally
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
