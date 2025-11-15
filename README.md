# create-go-api

A CLI tool to scaffold production-ready Go API services with database support (DynamoDB, Postgres), framework support (Chi, ConnectRPC), and one-click deployment.

## Features

- 🗄️ **Database Support**: PostgreSQL and DynamoDB
- 🚀 **Framework Support**: Chi and ConnectRPC
- 🎨 **Interactive TUI**: Beautiful terminal UI for project creation
- 📦 **One-click Deployment**: Optional Fly.io deployment setup
- 🧪 **Testing**: Built-in testcontainers for database testing
- 📊 **Observability**: Prometheus and Grafana integration

## Installation

### Using Go Install (Recommended)

```bash
go install github.com/anmho/create-go-api@latest
```

### Building from Source

```bash
git clone https://github.com/anmho/create-go-api.git
cd create-go-api
make cli-build
```

## Usage

### Interactive Mode (Recommended)

Simply run the command without any flags:

```bash
create-go-api create
```

Or explicitly use the interactive flag:

```bash
create-go-api create --interactive
```

### Non-Interactive Mode

Provide all required flags:

```bash
create-go-api create \
  --name my-api \
  --module-path github.com/username/my-api \
  --driver postgres \
  --framework chi \
  --output ./my-api
```

### Options

- `--name, -n`: Project name (required)
- `--module-path, -m`: Go module path (required)
- `--driver, -d`: Database driver (`postgres` or `dynamodb`)
- `--framework, -f`: API framework (`chi` or `connectrpc`)
- `--output, -o`: Output directory (defaults to project name)
- `--deploy`: Enable deployment setup (Fly.io)
- `--interactive, -i`: Use interactive TUI mode

### Check Version

```bash
create-go-api version
```

## Generated Project Structure

The tool generates a complete Go API project with:

- Production-ready project structure
- Database integration (PostgreSQL or DynamoDB)
- API handlers (Chi or ConnectRPC)
- Configuration management
- Docker Compose setup
- Database migrations (PostgreSQL)
- Testing setup with testcontainers
- Prometheus metrics
- Grafana dashboards
- Deployment scripts (optional)

## Development

### Prerequisites

- Go 1.25.4 or later
- Make

### Building

```bash
make cli-build
```

### Running Tests

```bash
make test
```

### Code Quality

```bash
make fmt    # Format code
make vet    # Vet code
make lint   # Lint code
```

## Releasing

This project uses [goreleaser](https://goreleaser.com/) for releases.

### Creating a Release

1. Create and push a git tag:
   ```bash
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```

2. Run the release command:
   ```bash
   make release
   ```

This will:
- Build binaries for multiple platforms
- Create a GitHub release
- Update the Homebrew formula in the `anmho/homebrew-taps` repository

### Testing Releases

To test the release process without creating an actual release:

```bash
make release-snapshot
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT

