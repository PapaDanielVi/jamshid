# Contributing to Jamshid

Thank you for considering contributing to Jamshid!

## How to Contribute

### Reporting Bugs

- Use the GitHub issue tracker
- Include steps to reproduce
- Mention your OS and Go version

### Suggesting Enhancements

- Open an issue with the "enhancement" label
- Describe the feature and why it would be useful
- Discuss implementation approach if possible

### Pull Requests

1. Fork the repo and create a branch from `main`
2. Make your changes
3. Run `make all` to ensure fmt, vet, lint, and tests pass
4. Commit with a clear message
5. Push and create a PR

## Development Setup

```bash
git clone https://github.com/PapaDanielVi/jamshid.git
cd jamshid
go mod download
make build
```

## Code Style

- Follow standard Go conventions
- Run `gofmt` before committing
- Add tests for new functionality
- Keep it simple - no over-engineering

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
