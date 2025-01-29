package output

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/degenhousedefi/initia-vanity/pkg/vanity"
)

func TestFormatResults(t *testing.T) {
	sampleResults := []vanity.Result{
		{
			Address:    "init1test123",
			PrivateKey: "privatekey1",
			PublicKey:  "publickey1",
		},
		{
			Address:    "init1test456",
			PrivateKey: "privatekey2",
			PublicKey:  "publickey2",
		},
	}

	tests := []struct {
		name        string
		format      string
		results     []vanity.Result
		wantEmpty   bool
		wantErr     bool
		checkFormat func(string) error
	}{
		{
			name:      "empty results",
			format:    "text",
			results:   nil,
			wantEmpty: true,
		},
		{
			name:    "text format",
			format:  "text",
			results: sampleResults,
			checkFormat: func(output string) error {
				expected := []string{
					"Address: init1test123",
					"Private key: privatekey1",
					"Public key: publickey1",
					"Address: init1test456",
				}
				for _, exp := range expected {
					if !strings.Contains(output, exp) {
						t.Errorf("expected output to contain '%s'", exp)
					}
				}
				return nil
			},
		},
		{
			name:    "json format",
			format:  "json",
			results: sampleResults,
			checkFormat: func(output string) error {
				var results []vanity.Result
				if err := json.Unmarshal([]byte(output), &results); err != nil {
					return err
				}
				if len(results) != 2 {
					t.Errorf("expected 2 results, got %d", len(results))
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFormatter(tt.format, false)
			output, err := f.FormatResults(tt.results)

			if (err != nil) != tt.wantErr {
				t.Errorf("FormatResults() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantEmpty && output != "" {
				t.Error("FormatResults() expected empty output")
				return
			}

			if tt.checkFormat != nil {
				if err := tt.checkFormat(output); err != nil {
					t.Errorf("FormatResults() output format check failed: %v", err)
				}
			}
		})
	}
}

func TestFormatStats(t *testing.T) {
	stats := vanity.Stats{
		Attempts: 1000,
		Found:    5,
	}
	duration := 2 * time.Second

	f := NewFormatter("text", false)
	output := f.FormatStats(stats, duration)

	expectedStrings := []string{
		"Duration:",
		"Total attempts: 1000",
		"Addresses found: 5",
		"Average speed:",
		"Attempts per match:",
	}

	for _, exp := range expectedStrings {
		if !strings.Contains(output, exp) {
			t.Errorf("FormatStats() output missing '%s'", exp)
		}
	}
}
