# Scrum Master

A beautiful, modular Go application that uses AI to analyze project descriptions and automatically create JIRA epics and stories with proper linking.

## 🚀 Features

- **AI-Powered Analysis**: Analyzes project descriptions to create epics and stories
- **JIRA Integration**: Automatically creates JIRA tickets with proper epic linking
- **Modular Architecture**: Clean separation of concerns with services, repositories, and helpers
- **Beautiful CLI**: Colorful, informative terminal output with progress indicators
- **Configuration Management**: YAML-based configuration with validation
- **Error Handling**: Comprehensive error handling with retry logic
- **Dry Run Mode**: Preview changes before creating actual JIRA tickets

## 📦 Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd scrum-master
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o bin/scrum-master cmd/scrum-master/main.go
```

## ⚙️ Configuration

Create a `config.yaml` file with your settings:

```yaml
anthropic:
  api_key: your-anthropic-api-key
  model: claude-sonnet-4-20250514
  timeout_seconds: 120
  max_tokens: 4000
  chunk_size_chars: 15000
  retry_count: 3
  retry_delay_seconds: 5

jira:
  base_url: https://your-domain.atlassian.net
  username: your-email@example.com
  api_token: your-jira-api-token
  project_key: YOUR_PROJECT_KEY
  timeout_seconds: 30

processing:
  mode: full
  output_dir: ./output
  save_intermediate: true
```

## 🎯 Usage

### Process a Project Description

Analyze a project description file and create a breakdown:

```bash
./bin/scrum-master process project-desc.md
```

Options:
- `--mode, -m`: Processing mode (`analyze-only`, `full`)
- `--config, -c`: Configuration file path (default: `config.yaml`)

### Create JIRA Tickets from Analysis

Load an analysis file and create JIRA tickets:

```bash
./bin/scrum-master create-from-analysis ./output/project-desc-analysis-20250101-120000.json
```

Options:
- `--dry-run, -d`: Show what would be created without actually creating tickets
- `--config, -c`: Configuration file path (default: `config.yaml`)

## 🏛️ Architecture Details

### Services Layer (`internal/services/`)

- **AnalysisService**: Handles project analysis and breakdown display
- **JiraService**: Manages JIRA ticket creation with business logic

### Repository Layer (`internal/repositories/`)

- **JiraRepository**: Handles all JIRA API interactions with proper error handling

### Models (`internal/models/`)

- **Project Models**: Epic, Story, and ProjectBreakdown structures
- **JIRA Models**: JiraIssue, JiraFields, and API response structures

### Helpers (`internal/helpers/`)

- **Colors**: Beautiful terminal output with emojis and colors
- **Files**: File operations, JSON handling, and path utilities

### Configuration (`internal/config/`)

- **Config**: Centralized configuration management with validation
- **LoadConfig**: YAML parsing with comprehensive error handling

## 🔧 Development

### Adding New Features

1. **Models**: Add new data structures in `internal/models/`
2. **Repositories**: Add data access logic in `internal/repositories/`
3. **Services**: Add business logic in `internal/services/`
4. **Helpers**: Add utilities in `internal/helpers/`
5. **CLI**: Add commands in `cmd/scrum-master/main.go`

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/services
```

### Building

```bash
# Build for current platform
go build -o bin/scrum-master cmd/scrum-master/main.go

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o bin/scrum-master-linux cmd/scrum-master/main.go
GOOS=darwin GOARCH=amd64 go build -o bin/scrum-master-mac cmd/scrum-master/main.go
GOOS=windows GOARCH=amd64 go build -o bin/scrum-master-windows.exe cmd/scrum-master/main.go
```

## 🎨 Code Style

This project follows Go best practices:

- **Package Naming**: Use descriptive, lowercase package names
- **Error Handling**: Use `fmt.Errorf` with `%w` for error wrapping
- **Documentation**: All exported functions have godoc comments
- **Testing**: Comprehensive unit tests for all packages
- **Dependencies**: Minimal external dependencies, prefer standard library

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI
- Beautiful colors with [Fatih Color](https://github.com/fatih/color)
- Configuration with [YAML](https://gopkg.in/yaml.v2)