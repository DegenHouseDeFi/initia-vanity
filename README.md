# Initia Vanity Address Generator

A high-performance vanity address generator for Initia. Generate custom addresses that match specific patterns.

## Features

- Generate addresses with custom patterns
- Multiple pattern matching modes (start, end, any)
- Multi-threaded for high performance
- Case-sensitive/insensitive matching
- JSON/Text output formats
- Progress reporting and statistics
- File output support

## Installation

### Using Go
```bash
go install github.com/degenhousedefi/initia-vanity/cmd/initia-vanity@latest
```

Or build from source:
```bash
git clone https://github.com/degenhousedefi/initia-vanity.git
cd initia-vanity
make build
```

### Using Homebrew
```bash
# Use the full repository URL in the tap command
brew tap degenhousedefi/initia-vanity https://github.com/degenhousedefi/initia-vanity
brew install initia-vanity
```

## Usage

Basic usage examples:
```bash
# Generate an address ending with "alice"
initia-vanity -p end alice

# Generate 5 addresses containing "bob" anywhere
initia-vanity -p any -c 5 bob

# Generate a case-sensitive address starting with "Charlie"
initia-vanity -p start --case-sensitive Charlie
```

Advanced usage:
```bash
# Save results to JSON file and show statistics
initia-vanity -p any --format json --stats -o addresses.json alice

# Use 8 threads and suppress progress output
initia-vanity -p end -t 8 --quiet rohan

# Generate 10 addresses with pattern matching anywhere
initia-vanity -p any -c 10 test

# Multiple options combined
initia-vanity -p start --case-sensitive -c 3 --stats --format json Bob
```

### Options

- `-p, --position`: Match position (start|end|any)
  - `start`: Match after init1 prefix
  - `end`: Match at the end
  - `any`: Match anywhere in address
- `-t, --threads`: Number of threads (default: CPU cores)
- `--case-sensitive`: Enable case-sensitive matching
- `-o, --output`: Output file path (if not specified, prints to stdout)
- `--format`: Output format (text|json)
- `--quiet`: Suppress progress output
- `-c, --count`: Number of addresses to generate
- `--stats`: Show performance statistics

## Development

```bash
# Run tests
make test

# Run all checks
make check

# Build for all platforms
make build-all
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.