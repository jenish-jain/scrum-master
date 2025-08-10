# Scrum Master

A powerful CLI tool that uses AI to automatically break down project descriptions into actionable JIRA epics and user stories. Perfect for project managers and development teams who want to streamline their project planning process.

## Features

🤖 **AI-Powered Analysis**: Uses Claude AI to intelligently break down complex project descriptions
📝 **Markdown Input**: Accepts project descriptions in markdown format
🎫 **JIRA Integration**: Automatically creates epics and stories in your JIRA board
🖥️ **Beautiful CLI**: Rich terminal interface with colors and progress indicators
🔍 **Dry Run Mode**: Preview what will be created before committing to JIRA
⚙️ **Configurable**: Easy YAML configuration for API keys and settings
📊 **Multiple Processing Modes**: Choose between full processing, analysis-only, or JIRA creation-only
🔄 **Batch Processing**: Automatically splits large files into chunks to prevent timeouts
💾 **Intermediate Results**: Save analysis results for later JIRA creation
🔁 **Retry Logic**: Automatic retry with exponential backoff for failed API requests

## Installation

### Prerequisites

- Go 1.21 or higher
- JIRA account with API access
- Anthropic API key

### Build from Source

```bash
git clone <repository-url>
cd scrum-master
go mod tidy
go build -o scrum-master main.go
```

### Using Go Install

```bash
go install github.com/yourusername/scrum-master@latest
```

## Quick Start

### 1. Initialize Configuration

```bash
./scrum-master init
```

This creates a `config.yaml` file with placeholder values.

### 2. Configure API Keys

Edit the generated `config.yaml` file:

```yaml
anthropic:
  api_key: "your-anthropic-api-key-here"
  model: "claude-sonnet-4-20250514"

jira:
  base_url: "https://your-domain.atlassian.net"
  username: "your-email@example.com"
  api_token: "your-jira-api-token"
  project_key: "PROJ"
```

### 3. Create Your Project Description

Create a markdown file (e.g., `project.md`) with your project description:

```markdown
# My Awesome Project

## Overview
This project aims to...

## Requirements
- Feature 1
- Feature 2
- Feature 3

## Goals
- Improve user experience
- Increase performance
- Add new capabilities
```

### 4. Process the Project

```bash
# Full processing (analyze + create JIRA tickets)
./scrum-master process -i project.md

# Analyze only (save results without creating JIRA tickets)
./scrum-master process -i project.md --mode analyze-only

# Dry run (preview only)
./scrum-master process -i project.md --dry-run

# Custom output directory
./scrum-master process -i project.md --mode analyze-only --output ./my-analysis

# Create JIRA tickets from previously saved analysis
./scrum-master create-from-analysis --analysis ./output/project-analysis-20240110-143022.json
```

## Processing Modes

### Full Mode (Default)
- Analyzes the project with AI
- Saves intermediate results (if enabled)
- Creates JIRA tickets
- Best for: Complete workflow in one command

### Analyze-Only Mode
- Only processes the project with AI
- Saves detailed analysis to JSON and Markdown files
- No JIRA tickets created
- Best for: Large projects, reviewing analysis before creating tickets, or when you want to save on API costs

### Create-Only Mode
- Creates JIRA tickets from previously saved analysis
- Use the `create-from-analysis` command
- Best for: When you've already analyzed and want to create tickets later

## Handling Large Files

The tool automatically handles large project descriptions by:

1. **Chunking**: Files larger than the configured chunk size are split into overlapping chunks
2. **Batch Processing**: Each chunk is processed separately with rate limiting
3. **Merging**: Results are intelligently merged and deduplicated
4. **Intermediate Saving**: Each chunk's results are saved for debugging
5. **Retry Logic**: Failed requests are automatically retried with exponential backoff

## Configuration

### Getting API Keys

#### Anthropic API Key
1. Visit [Anthropic Console](https://console.anthropic.com/)
2. Create an account or sign in
3. Navigate to API Keys section
4. Generate a new API key

#### JIRA API Token
1. Visit [Atlassian API Tokens](https://id.atlassian.com/manage-profile/security/api-tokens)
2. Create an API token
3. Use your Atlassian email as username
4. Use the generated token as api_token

### Configuration File

The configuration file supports the following options:

```yaml
anthropic:
  api_key: string              # Your Anthropic API key
  model: string                # AI model to use (default: claude-sonnet-4-20250514)
  timeout_seconds: 120         # API request timeout
  max_tokens: 4000             # Maximum tokens per request
  chunk_size_chars: 15000      # Size for splitting large files
  retry_count: 3               # Number of retries for failed requests
  retry_delay_seconds: 5       # Delay between retries

jira:
  base_url: string             # Your JIRA instance URL
  username: string             # Your JIRA username (email)
  api_token: string            # Your JIRA API token
  project_key: string          # JIRA project key where tickets will be created
  timeout_seconds: 30          # JIRA API request timeout

processing:
  mode: "full"                 # Processing mode: "full", "analyze-only", "create-only"
  output_dir: "./output"       # Directory for saving analysis files
  save_intermediate: true      # Save intermediate chunk results
```

## Usage

### Commands

#### `init`
Initialize a new configuration file:
```bash
scrum-master init [--config config.yaml]
```

#### `process`
Process a project description and create JIRA tickets:
```bash
scrum-master process -i <input-file> [options]
```

Options:
- `-i, --input`: Input markdown file (required)
- `-d, --dry-run`: Preview mode - don't create actual JIRA tickets
- `-m, --mode`: Processing mode ("full", "analyze-only", "create-only")
- `-o, --output`: Output directory for analysis files
- `-c, --config`: Configuration file path (default: config.yaml)

#### `create-from-analysis`
Create JIRA tickets from previously saved analysis:
```bash
scrum-master create-from-analysis --analysis <analysis-file> [options]
```

Options:
- `-a, --analysis`: Analysis JSON file (required)
- `-d, --dry-run`: Preview mode - don't create actual JIRA tickets
- `-c, --config`: Configuration file path (default: config.yaml)

### Input Format

Your markdown file should include:

- **Project title** (H1 heading)
- **Overview/Description** section
- **Requirements** or **Features** section
- **Goals** or **Objectives** section
- Any additional context that helps with breakdown

The AI will analyze the entire document and create structured epics and stories based on the content.

### Output

The tool creates:

1. **Epics**: Major functional areas or phases
2. **Stories**: Specific user stories with:
   - Proper user story format ("As a ... I want ... so that ...")
   - Story points (Fibonacci sequence: 1, 2, 3, 5, 8)
   - Priority levels (High, Medium, Low)
   - Acceptance criteria
   - Dependencies (where applicable)

## Example Output

```
🤖 Starting Project Breakdown Process
📝 Loading configuration...
✅ Configuration loaded - Mode: analyze-only
📖 Reading input file: large-project.md
✅ Read 55132 bytes from input file
🧠 Processing with AI (4 chunks)...
  Processing chunk 1/4...
  Processing chunk 2/4...
  Processing chunk 3/4...
  Processing chunk 4/4...
✅ AI processing complete - 4 chunks processed, 6 epics found
✅ Analysis saved:
  📄 JSON: ./output/large-project-analysis-20240110-143022.json  
  📝 Markdown: ./output/large-project-summary-20240110-143022.md

📋 Project Breakdown: E-Commerce Platform Modernization
Overview: Modernize existing e-commerce platform for better performance and UX
Processed in 4 chunks

Epic 1: Frontend Modernization | Chunk: 1
Priority: High
Description: Migrate from jQuery to React with responsive design

  Story 1.1: Set up React development environment
    Points: 3 | Priority: High
    Description: As a developer, I want a modern React development environment so that I can build efficient components
    Acceptance Criteria:
      • React 18+ with TypeScript configured
      • ESLint and Prettier set up
      • Hot reloading working

📊 Summary: 6 epics, 34 stories, 127 story points total
🎉 Analysis completed and saved!
```

### Output Files

The tool generates several output files:

1. **Main Analysis JSON**: Complete breakdown with all epics and stories
2. **Summary Markdown**: Human-readable summary with formatting
3. **Chunk Analysis Files**: Individual chunk results (if processing large files)

Example file structure:
```
output/
├── large-project-analysis-20240110-143022.json
├── large-project-summary-20240110-143022.md
├── large-project-chunk-1-20240110-143015.json
├── large-project-chunk-2-20240110-143018.json
├── large-project-chunk-3-20240110-143020.json
└── large-project-chunk-4-20240110-143022.json
```

## Best Practices

### Writing Good Project Descriptions

1. **Be Specific**: Include concrete requirements and features
2. **Add Context**: Explain the "why" behind features
3. **Include Constraints**: Mention technical limitations, budget, timeline
4. **Define Success**: Clear metrics and goals
5. **Structure Well**: Use headings and bullet points for clarity

### Managing Large Projects

1. **Use Analyze-Only Mode First**: For large projects (>50KB), analyze first to review before creating tickets
2. **Adjust Chunk Size**: Modify `chunk_size_chars` in config for optimal processing
3. **Review Merged Results**: Always review the final analysis as chunk merging may need manual adjustment
4. **Save Intermediate Results**: Enable `save_intermediate` to debug chunk processing

### Cost Optimization

1. **Use Analyze-Only Mode**: Avoid repeated AI processing by saving analysis results
2. **Optimize Chunk Size**: Larger chunks = fewer API calls but higher timeout risk
3. **Review Before Creating**: Use dry-run mode to validate before creating JIRA tickets

### Project Organization

1. **Use Consistent Naming**: Follow your team's naming conventions
2. **Regular Updates**: Re-process when requirements change significantly
3. **Version Your Analysis**: Keep timestamped analysis files for tracking changes