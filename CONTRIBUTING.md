# Contributing to Recommendation Engine

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on what is best for the community
- Show empathy towards other community members

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in Issues
2. If not, create a new issue with:
   - Clear title and description
   - Steps to reproduce
   - Expected vs actual behavior
   - System information (OS, Go version, etc.)
   - Relevant logs or screenshots

### Suggesting Features

1. Check if the feature has already been suggested
2. Create a new issue with:
   - Clear description of the feature
   - Use cases and benefits
   - Potential implementation approach (optional)

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Format code (`go fmt ./...`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to your fork (`git push origin feature/amazing-feature`)
9. Open a Pull Request

### Development Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/reco-engine.git
cd reco-engine

# Install dependencies
go mod download

# Start infrastructure
docker-compose up -d postgres redis kafka

# Run tests
go test ./...

# Run a service locally
go run ./cmd/ingest
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Use `golangci-lint` for linting
- Write meaningful commit messages
- Add comments for complex logic
- Keep functions small and focused

### Testing

- Write unit tests for new functionality
- Maintain or improve code coverage
- Test edge cases and error conditions
- Use table-driven tests where appropriate

Example:

```go
func TestValidateEvent(t *testing.T) {
    tests := []struct {
        name    string
        event   *models.Event
        wantErr bool
    }{
        {
            name: "valid event",
            event: &models.Event{
                UserID: 1,
                ItemID: 2,
                EventType: "VIEW",
            },
            wantErr: false,
        },
        // Add more test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateEvent(tt.event)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateEvent() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Documentation

- Update README.md for user-facing changes
- Update API.md for API changes
- Add/update code comments
- Update CHANGELOG.md

### Commit Messages

Follow the conventional commits format:

```
type(scope): subject

body

footer
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting)
- `refactor`: Code refactoring
- `test`: Adding/updating tests
- `chore`: Maintenance tasks

Example:

```
feat(api): add category filter to popular endpoint

Added optional category query parameter to /popular endpoint
to allow filtering popular items by category.

Closes #123
```

### Review Process

1. Maintainers will review your PR
2. Address any feedback or requested changes
3. Once approved, a maintainer will merge your PR

## Project Structure

```
reco-engine/
├── cmd/              # Service entry points
├── internal/         # Internal packages
│   ├── api/         # API service
│   ├── ingest/      # Ingest service
│   ├── processor/   # Stream processor
│   ├── store/       # Database/Redis clients
│   ├── models/      # Data models
│   └── util/        # Utilities
├── infra/           # Infrastructure configs
├── scripts/         # Utility scripts
└── docs/            # Documentation
```

## Questions?

Feel free to:
- Open an issue for questions
- Join discussions in GitHub Discussions
- Contact maintainers

Thank you for contributing! 🎉
