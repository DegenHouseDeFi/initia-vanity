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

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/go-bip39"
)

// Result represents a generated vanity address and its keys
type Result struct {
	Address        string `json:"address"`
	PrivateKey     string `json:"private_key"`
	PublicKey      string `json:"public_key"`
	Mnemonic       string `json:"mnemonic,omitempty"`
	DerivationPath string `json:"derivation_path,omitempty"`
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
	useMnemonic   bool
	mnemonic      string
	stats         *Stats
	results       []Result
	stopCh        chan struct{}
	progressCh    chan struct{}
	stopped       atomic.Bool
	mu            sync.Mutex
}

// NewGenerator creates a new vanity address generator
func NewGenerator(pattern, position string, caseSensitive bool, count int, useMnemonic bool, mnemonic string) *Generator {
	return &Generator{
		pattern:       pattern,
		position:      position,
		caseSensitive: caseSensitive,
		count:         count,
		useMnemonic:   useMnemonic,
		mnemonic:      mnemonic,
		stats:         &Stats{},
		stopCh:        make(chan struct{}),
	}
}

// generateMnemonic generates a new random mnemonic
func (g *Generator) generateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %v", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %v", err)
	}

	return mnemonic, nil
}

// generateAddress creates a new random Cosmos SDK compatible address
func (g *Generator) generateAddress() (string, string, string, error) {
	// Generate private key using Cosmos SDK's secp256k1
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()

	// Get address from public key
	addr := sdk.AccAddress(pubKey.Address())

	// Convert to bech32 with "init" prefix
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

// generateAddressFromMnemonic generates an address using HD wallet derivation
func (g *Generator) generateAddressFromMnemonic() (string, string, string, string, string, error) {
	var mnemonic string
	if g.mnemonic != "" {
		// Validate provided mnemonic
		if !bip39.IsMnemonicValid(g.mnemonic) {
			return "", "", "", "", "", fmt.Errorf("invalid mnemonic provided")
		}
		mnemonic = g.mnemonic
	} else {
		var err error
		mnemonic, err = g.generateMnemonic()
		if err != nil {
			return "", "", "", "", "", err
		}
	}

	// Derive seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// Create master key and derive path
	master, ch := hd.ComputeMastersFromSeed(seed)

	// Use BIP44 path: m/44'/118'/0'/0/index
	// 44' : BIP 44 purpose
	// 118': Cosmos coin type
	// 0'  : Account number
	// 0   : External branch
	// index: Address index
	path := "m/44'/118'/0'/0/0"

	derivedPrivKey, err := hd.DerivePrivateKeyForPath(master, ch, path)
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to derive private key: %v", err)
	}

	// Create private key from derived bytes
	privKey := &secp256k1.PrivKey{Key: derivedPrivKey}
	pubKey := privKey.PubKey()

	// Get address from public key
	addr := sdk.AccAddress(pubKey.Address())

	// Convert to bech32 with "init" prefix
	address, err := bech32.ConvertAndEncode("init", addr)
	if err != nil {
		return "", "", "", "", "", err
	}

	// Format keys
	privKeyHex := hex.EncodeToString(privKey.Bytes())
	pubKeyJSON := map[string]interface{}{
		"@type": "/cosmos.crypto.secp256k1.PubKey",
		"key":   base64.StdEncoding.EncodeToString(pubKey.Bytes()),
	}
	pubKeyBytes, err := json.Marshal(pubKeyJSON)
	if err != nil {
		return "", "", "", "", "", err
	}

	return address, privKeyHex, string(pubKeyBytes), mnemonic, path, nil
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

			var address, privKey, pubKey, mnemonic, derivationPath string
			var err error

			if g.useMnemonic {
				address, privKey, pubKey, mnemonic, derivationPath, err = g.generateAddressFromMnemonic()
			} else {
				address, privKey, pubKey, err = g.generateAddress()
			}

			if err != nil {
				continue
			}

			if g.isMatch(address) {
				result := Result{
					Address:    address,
					PrivateKey: privKey,
					PublicKey:  pubKey,
				}

				if g.useMnemonic {
					result.Mnemonic = mnemonic
					result.DerivationPath = derivationPath
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
