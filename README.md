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

```bash
go install github.com/degenhousedefi/initia-vanity/cmd/initia-vanity@latest
```

Or build from source:

```bash
git clone https://github.com/degenhousedefi/initia-vanity.git
cd initia-vanity
make build
```

## Usage

```bash
# Generate an address ending with "alice"
initia-vanity -p end alice

# Generate 5 addresses containing "bob" anywhere
initia-vanity -p any -c 5 bob

# Generate a case-sensitive address starting with "Charlie"
initia-vanity -p start --case-sensitive Charlie

# Save results to JSON file
initia-vanity -p any --format json -o addresses.json alice

# Show generation statistics
initia-vanity --stats rohan
```

### Options

- `-p, --position`: Match position (start|end|any)
- `-t, --threads`: Number of threads (default: CPU cores)
- `--case-sensitive`: Enable case-sensitive matching
- `-o, --output`: Output file path
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