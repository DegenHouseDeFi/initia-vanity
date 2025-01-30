package output

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/degenhousedefi/initia-vanity/pkg/vanity"
)

// Formatter handles the output formatting
type Formatter struct {
	format string
	quiet  bool
}

// NewFormatter creates a new output formatter
func NewFormatter(format string, quiet bool) *Formatter {
	return &Formatter{
		format: format,
		quiet:  quiet,
	}
}

// FormatResults formats the generation results
func (f *Formatter) FormatResults(results []vanity.Result) (string, error) {
	if len(results) == 0 {
		return "", nil
	}

	if f.format == "json" {
		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return "", fmt.Errorf("error encoding JSON: %v", err)
		}
		return string(jsonData), nil
	}

	var builder strings.Builder
	for _, result := range results {
		// Always present fields
		builder.WriteString(fmt.Sprintf("Address: %s\n", result.Address))
		builder.WriteString(fmt.Sprintf("Private key: %s\n", result.PrivateKey))
		builder.WriteString(fmt.Sprintf("Public key: %s\n", result.PublicKey))

		// Mnemonic-specific fields
		if result.Mnemonic != "" {
			builder.WriteString(fmt.Sprintf("Mnemonic: %s\n", result.Mnemonic))
			builder.WriteString(fmt.Sprintf("Derivation path: %s\n", result.DerivationPath))
			builder.WriteString(fmt.Sprintf("Note: Import this mnemonic in your wallet to access this address\n"))
		}

		// Separator between results
		builder.WriteString("---\n")
	}
	return builder.String(), nil
}

// FormatStats formats the generation statistics
func (f *Formatter) FormatStats(stats vanity.Stats, duration time.Duration) string {
	speed := float64(stats.Attempts) / duration.Seconds()
	var builder strings.Builder

	builder.WriteString("\nStatistics:\n")
	builder.WriteString(fmt.Sprintf("Duration: %v\n", duration.Round(time.Second)))
	builder.WriteString(fmt.Sprintf("Total attempts: %d\n", stats.Attempts))
	builder.WriteString(fmt.Sprintf("Addresses found: %d\n", stats.Found))
	builder.WriteString(fmt.Sprintf("Average speed: %.2f addresses/second\n", speed))

	if stats.Found > 0 {
		attemptsPerMatch := float64(stats.Attempts) / float64(stats.Found)
		builder.WriteString(fmt.Sprintf("Attempts per match: %.2f\n", attemptsPerMatch))
	}

	return builder.String()
}

// PrintProgress prints the current progress
func (f *Formatter) PrintProgress(current, total int) {
	if !f.quiet {
		fmt.Printf("\rFound %d/%d addresses", current, total)
	}
}

// PrintSpeed prints the current speed
func (f *Formatter) PrintSpeed(attempts uint64, duration time.Duration) {
	if !f.quiet {
		speed := float64(attempts) / duration.Seconds()
		fmt.Printf("\rTried %d addresses (%.2f/s)", attempts, speed)
	}
}

// PrintMnemonicInfo prints information about mnemonic generation
func (f *Formatter) PrintMnemonicInfo(useMnemonic bool, customMnemonic bool) {
	if !f.quiet && useMnemonic {
		if customMnemonic {
			fmt.Println("Using provided mnemonic for address generation")
		} else {
			fmt.Println("Using new random mnemonic for address generation")
		}
	}
}
