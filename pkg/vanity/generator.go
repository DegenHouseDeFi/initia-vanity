package vanity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil/bech32"
	"golang.org/x/crypto/ripemd160"
)

const (
	HRP = "init" // Fixed HRP for Initia addresses
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

// generateAddress creates a bech32 address from a public key
func (g *Generator) generateAddress(pubKey []byte) string {
	sha256Hash := sha256.Sum256(pubKey)
	ripemdHasher := ripemd160.New()
	ripemdHasher.Write(sha256Hash[:])
	ripemdHash := ripemdHasher.Sum(nil)

	converted, err := bech32.ConvertBits(ripemdHash, 8, 5, true)
	if err != nil {
		return ""
	}

	address, err := bech32.EncodeFromBase256(HRP, converted)
	if err != nil {
		return ""
	}

	if !g.caseSensitive {
		address = strings.ToLower(address)
	}
	return address
}

// isMatch checks if an address matches the pattern
func (g *Generator) isMatch(address string) bool {
	pattern := g.pattern
	if !g.caseSensitive {
		pattern = strings.ToLower(pattern)
	}

	switch g.position {
	case "start":
		return strings.HasPrefix(address, HRP+"1"+pattern)
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

			privateKey, err := btcec.NewPrivateKey()
			if err != nil {
				continue
			}

			pubKey := privateKey.PubKey().SerializeCompressed()
			address := g.generateAddress(pubKey)
			if address == "" {
				continue
			}

			if g.isMatch(address) {
				result := Result{
					Address:    address,
					PrivateKey: hex.EncodeToString(privateKey.Serialize()),
					PublicKey:  hex.EncodeToString(pubKey),
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
