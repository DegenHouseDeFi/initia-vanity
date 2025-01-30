package main

import (
	"fmt"
	"os"
	"time"

	"github.com/degenhousedefi/initia-vanity/internal/config"
	"github.com/degenhousedefi/initia-vanity/internal/output"
	"github.com/degenhousedefi/initia-vanity/pkg/vanity"
	"github.com/spf13/cobra"
)

var cfg *config.Config

func main() {
	rootCmd := &cobra.Command{
		Use:   "initia-vanity [pattern]",
		Short: "Generate vanity addresses for Initia",
		Long: `Initia Vanity Address Generator

A tool to generate custom Initia's cosmos based public key that match specific patterns.
The generator supports searching for patterns at the start, end, or anywhere in the address.
All generated addresses will start with 'init1'.`,
		Args: cobra.MaximumNArgs(1),
		RunE: run,
		Example: `  # Generate an address ending with "alice"
  initia-vanity -p end alice

  # Find 3 addresses containing "bob" with statistics
  initia-vanity -p any -c 3 --stats bob

  # Generate a case-sensitive address starting with "Charlie"
  initia-vanity -p start --case-sensitive Charlie

  # Generate address using mnemonic
  initia-vanity -p end --use-mnemonic alice

  # Generate address using specific mnemonic
  initia-vanity -p end --use-mnemonic --mnemonic "your twelve words here" alice

  # Save results to a JSON file
  initia-vanity -p any --format json -o addresses.json alice

  # Use custom number of threads
  initia-vanity -p end -t 8 bob`,
	}

	cfg = config.DefaultConfig()

	// Pattern Matching Options
	rootCmd.Flags().StringVarP(&cfg.Position, "position", "p", cfg.Position,
		`Match position in address (one of: start, end, any)
- start: Match after init1 prefix
- end:   Match at the end
- any:   Match anywhere in address`)
	rootCmd.Flags().BoolVar(&cfg.CaseSensitive, "case-sensitive", cfg.CaseSensitive,
		"Enable case-sensitive pattern matching")
	rootCmd.Flags().IntVarP(&cfg.Count, "count", "c", cfg.Count,
		"Number of matching addresses to generate")

	// Key Generation Options
	rootCmd.Flags().BoolVar(&cfg.UseMnemonic, "use-mnemonic", cfg.UseMnemonic,
		"Use mnemonic-based key generation instead of random")
	rootCmd.Flags().StringVar(&cfg.Mnemonic, "mnemonic", cfg.Mnemonic,
		"Specify mnemonic phrase (optional, will be generated if not provided)")
	rootCmd.Flags().Uint32Var(&cfg.AccountNumber, "account", cfg.AccountNumber,
		"Account number for HD derivation path (default: 0)")
	rootCmd.Flags().Uint32Var(&cfg.AddressIndex, "address-index", cfg.AddressIndex,
		"Address index for HD derivation path (default: 0)")

	// Performance Options
	rootCmd.Flags().IntVarP(&cfg.Threads, "threads", "t", cfg.Threads,
		"Number of threads for parallel processing")
	rootCmd.Flags().BoolVar(&cfg.Stats, "stats", cfg.Stats,
		"Show performance statistics")

	// Output Options
	rootCmd.Flags().StringVarP(&cfg.OutputFile, "output", "o", cfg.OutputFile,
		"Output file path (if not specified, prints to stdout)")
	rootCmd.Flags().StringVar(&cfg.Format, "format", cfg.Format,
		"Output format (one of: text, json)")
	rootCmd.Flags().BoolVar(&cfg.Quiet, "quiet", cfg.Quiet,
		"Suppress progress output")

	rootCmd.Version = "v1.0.0"

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// If no pattern is provided, display help
	if len(args) == 0 {
		return cmd.Help()
	}

	// Set pattern from args
	cfg.Pattern = args[0]

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}

	// Create formatter
	formatter := output.NewFormatter(cfg.Format, cfg.Quiet)

	if !cfg.Quiet {
		fmt.Printf("Searching for pattern: %s\n", cfg.Pattern)
		fmt.Printf("Position: %s\n", cfg.Position)
		fmt.Printf("Using %d threads\n", cfg.Threads)
		if cfg.UseMnemonic {
			fmt.Println("Using mnemonic-based generation")
			if cfg.Mnemonic != "" {
				fmt.Println("Using provided mnemonic")
			}
		}
	}

	// Create and start generator
	startTime := time.Now()
	generator := vanity.NewGenerator(cfg.Pattern, cfg.Position, cfg.CaseSensitive, cfg.Count, cfg.UseMnemonic, cfg.Mnemonic)

	// Start generation
	if err := generator.Generate(cfg.Threads); err != nil {
		return fmt.Errorf("generation failed: %v", err)
	}

	// Get results
	results := generator.GetResults()
	output, err := formatter.FormatResults(results)
	if err != nil {
		return fmt.Errorf("error formatting results: %v", err)
	}

	// Write output
	if cfg.OutputFile != "" {
		if err := os.WriteFile(cfg.OutputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
		if !cfg.Quiet {
			fmt.Printf("Results written to %s\n", cfg.OutputFile)
		}
	} else {
		fmt.Println(output)
	}

	// Print statistics if requested
	if cfg.Stats && !cfg.Quiet {
		stats := generator.GetStats()
		fmt.Print(formatter.FormatStats(stats, time.Since(startTime)))
	}

	return nil
}
