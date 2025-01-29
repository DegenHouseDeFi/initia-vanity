package vanity

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// Result represents a generated vanity address and its keys
type Result struct {
	Address    string `json:"address"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
}

// Stats holds generation statistics
type Stats struct {
	Attempts  uint64
	Found     uint64
	StartTime int64
	EndTime   int64
}

// Generator handles the vanity address generation
type Generator struct {
	pattern       string
	position      string
	caseSensitive bool
	count         int
	stats         *Stats
	results       []Result
	stopCh        chan struct{}
	progressCh    chan struct{}
	stopped       atomic.Bool
	mu            sync.Mutex
}

// NewGenerator creates a new vanity address generator
func NewGenerator(pattern, position string, caseSensitive bool, count int) *Generator {
	return &Generator{
		pattern:       pattern,
		position:      position,
		caseSensitive: caseSensitive,
		count:         count,
		stats:         &Stats{},
		stopCh:        make(chan struct{}),
	}
}

// generateAddress creates a new Cosmos SDK compatible address
func (g *Generator) generateAddress() (string, string, string, error) {
	// Generate private key using Cosmos SDK's secp256k1
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	// Get address from public key
	addr := sdk.AccAddress(pubKey.Address())

	// Encode address to bech32 with "init" prefix
	address, err := bech32.ConvertAndEncode("init", addr)
	if err != nil {
		return "", "", "", err
	}

	// Get private key hex
	privKeyHex := hex.EncodeToString(privKey.Bytes())

	// Format public key as JSON
	pubKeyJSON := map[string]interface{}{
		"@type": "/cosmos.crypto.secp256k1.PubKey",
		"key":   base64.StdEncoding.EncodeToString(pubKey.Bytes()),
	}
	pubKeyBytes, err := json.Marshal(pubKeyJSON)
	if err != nil {
		return "", "", "", err
	}

	return address, privKeyHex, string(pubKeyBytes), nil
}

// isMatch checks if an address matches the pattern
func (g *Generator) isMatch(address string) bool {
	pattern := g.pattern
	if !g.caseSensitive {
		pattern = strings.ToLower(pattern)
		address = strings.ToLower(address)
	}

	switch g.position {
	case "start":
		return strings.HasPrefix(address, "init1"+pattern)
	case "end":
		return strings.HasSuffix(address, pattern)
	case "any":
		return strings.Contains(address, pattern)
	default:
		return false
	}
}

// GetResults returns the generated results
func (g *Generator) GetResults() []Result {
	g.mu.Lock()
	defer g.mu.Unlock()
	return append([]Result{}, g.results...)
}

// GetStats returns the current statistics
func (g *Generator) GetStats() Stats {
	return Stats{
		Attempts: atomic.LoadUint64(&g.stats.Attempts),
		Found:    atomic.LoadUint64(&g.stats.Found),
	}
}

// Stop stops the generation process
func (g *Generator) Stop() {
	if !g.stopped.Swap(true) {
		close(g.stopCh)
	}
}

func (g *Generator) worker(wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-g.stopCh:
			return
		case <-ticker.C:
			select {
			case g.progressCh <- struct{}{}:
			default:
			}
		default:
			g.mu.Lock()
			if len(g.results) >= g.count {
				g.mu.Unlock()
				g.Stop()
				return
			}
			g.mu.Unlock()

			if g.stopped.Load() {
				return
			}

			address, privKey, pubKey, err := g.generateAddress()
			if err != nil {
				continue
			}

			if g.isMatch(address) {
				result := Result{
					Address:    address,
					PrivateKey: privKey,
					PublicKey:  pubKey,
				}

				g.mu.Lock()
				if len(g.results) < g.count {
					g.results = append(g.results, result)
					atomic.AddUint64(&g.stats.Found, 1)
				}
				g.mu.Unlock()
			}

			atomic.AddUint64(&g.stats.Attempts, 1)
		}
	}
}

// Generate starts the address generation process
func (g *Generator) Generate(threads int) error {
	g.progressCh = make(chan struct{}, 1)
	defer close(g.progressCh)

	var wg sync.WaitGroup
	wg.Add(threads)

	startTime := time.Now()

	// Start progress reporter
	go func() {
		for range g.progressCh {
			if g.stopped.Load() {
				return
			}
			attempts := atomic.LoadUint64(&g.stats.Attempts)
			found := atomic.LoadUint64(&g.stats.Found)
			speed := float64(attempts) / time.Since(startTime).Seconds()

			fmt.Printf("\rProgress: %d/%d found | Attempts: %d | Speed: %.2f/s",
				found, g.count, attempts, speed)
		}
	}()

	// Start workers
	for i := 0; i < threads; i++ {
		go g.worker(&wg)
	}

	wg.Wait()
	fmt.Println() // New line after progress
	return nil
}
