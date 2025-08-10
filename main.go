package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// Configuration structure
type Config struct {
	Anthropic struct {
		APIKey     string `yaml:"api_key"`
		Model      string `yaml:"model"`
		Timeout    int    `yaml:"timeout_seconds"`
		MaxTokens  int    `yaml:"max_tokens"`
		ChunkSize  int    `yaml:"chunk_size_chars"`
		RetryCount int    `yaml:"retry_count"`
		RetryDelay int    `yaml:"retry_delay_seconds"`
	} `yaml:"anthropic"`
	Jira struct {
		BaseURL    string `yaml:"base_url"`
		Username   string `yaml:"username"`
		APIToken   string `yaml:"api_token"`
		ProjectKey string `yaml:"project_key"`
		Timeout    int    `yaml:"timeout_seconds"`
	} `yaml:"jira"`
	Processing struct {
		Mode             string `yaml:"mode"` // "full", "analyze-only", "create-only"
		OutputDir        string `yaml:"output_dir"`
		SaveIntermediate bool   `yaml:"save_intermediate"`
	} `yaml:"processing"`
}

// AI Response structures
type Epic struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	Stories     []Story `json:"stories"`
	Chunk       int     `json:"chunk,omitempty"`
}

type Story struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Priority           string   `json:"priority"`
	StoryPoints        int      `json:"story_points"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Dependencies       []string `json:"dependencies,omitempty"`
	EpicTitle          string   `json:"epic_title,omitempty"`
}

type ProjectBreakdown struct {
	ProjectName    string    `json:"project_name"`
	Overview       string    `json:"overview"`
	Epics          []Epic    `json:"epics"`
	ProcessedAt    time.Time `json:"processed_at"`
	TotalChunks    int       `json:"total_chunks,omitempty"`
	InputFileSize  int64     `json:"input_file_size"`
	ProcessingMode string    `json:"processing_mode"`
}

type ChunkAnalysis struct {
	ChunkIndex int    `json:"chunk_index"`
	Content    string `json:"content"`
	Epics      []Epic `json:"epics"`
}

// JIRA structures
type JiraIssue struct {
	Fields JiraFields `json:"fields"`
}

type JiraFields struct {
	Project     JiraProject   `json:"project"`
	Summary     string        `json:"summary"`
	Description string        `json:"description"`
	IssueType   JiraIssueType `json:"issuetype"`
	Parent      *JiraParent   `json:"parent,omitempty"` // Parent field for epic linking (pointer to allow nil)
	// Priority field removed as it's not available in this JIRA project
}

type JiraProject struct {
	Key string `json:"key"`
}

type JiraIssueType struct {
	Name string `json:"name"`
}

type JiraParent struct {
	Key string `json:"key"`
}

type JiraPriority struct {
	Name string `json:"name"`
}

type JiraResponse struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

// CLI App structure
type App struct {
	config Config
}

// Color definitions for terminal output
var (
	successColor = color.New(color.FgGreen, color.Bold)
	errorColor   = color.New(color.FgRed, color.Bold)
	infoColor    = color.New(color.FgCyan)
	warningColor = color.New(color.FgYellow)
	headerColor  = color.New(color.FgMagenta, color.Bold)
)

func main() {
	app := &App{}

	var rootCmd = &cobra.Command{
		Use:   "project-breakdown",
		Short: "Break down project descriptions into JIRA epics and stories using AI",
		Long: `A CLI tool that reads markdown project descriptions and uses AI to break them down
into manageable epics and stories, then creates them as tickets in JIRA.

Supports multiple processing modes:
- full: Analyze and create JIRA tickets (default)
- analyze-only: Only analyze and save results to files
- create-only: Create JIRA tickets from previously saved analysis`,
	}

	var configFile string
	var inputFile string
	var dryRun bool
	var mode string
	var outputDir string

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "Configuration file path")

	var processCmd = &cobra.Command{
		Use:   "process",
		Short: "Process a markdown file and create JIRA tickets",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.processProject(configFile, inputFile, dryRun, mode, outputDir)
		},
	}

	processCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input markdown file (required)")
	processCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show what would be created without actually creating JIRA tickets")
	processCmd.Flags().StringVarP(&mode, "mode", "m", "", "Processing mode: full, analyze-only, create-only")
	processCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for analysis files")
	processCmd.MarkFlagRequired("input")

	var createCmd = &cobra.Command{
		Use:   "create-from-analysis",
		Short: "Create JIRA tickets from previously saved analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.createFromAnalysis(configFile, inputFile, dryRun)
		},
	}

	createCmd.Flags().StringVarP(&inputFile, "analysis", "a", "", "Analysis JSON file (required)")
	createCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show what would be created without actually creating JIRA tickets")
	createCmd.MarkFlagRequired("analysis")

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return app.initConfig(configFile)
		},
	}

	rootCmd.AddCommand(processCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		errorColor.Println("Error:", err)
		os.Exit(1)
	}
}

func (app *App) initConfig(configPath string) error {
	headerColor.Println("🚀 Initializing Project Breakdown Bot Configuration")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Configuration file already exists at %s\n", configPath)
		fmt.Print("Do you want to overwrite it? (y/N): ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		response := strings.ToLower(strings.TrimSpace(scanner.Text()))

		if response != "y" && response != "yes" {
			infoColor.Println("Configuration initialization cancelled.")
			return nil
		}
	}

	// Create sample config with enhanced settings
	config := Config{}
	config.Anthropic.APIKey = "your-anthropic-api-key-here"
	config.Anthropic.Model = "claude-sonnet-4-20250514"
	config.Anthropic.Timeout = 120
	config.Anthropic.MaxTokens = 4000
	config.Anthropic.ChunkSize = 15000
	config.Anthropic.RetryCount = 3
	config.Anthropic.RetryDelay = 5
	config.Jira.BaseURL = "https://your-domain.atlassian.net"
	config.Jira.Username = "your-email@example.com"
	config.Jira.APIToken = "your-jira-api-token"
	config.Jira.ProjectKey = "PROJ"
	config.Jira.Timeout = 30
	config.Processing.Mode = "full"
	config.Processing.OutputDir = "./output"
	config.Processing.SaveIntermediate = true

	// Write config file
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	err = os.WriteFile(configPath, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	successColor.Printf("✅ Configuration file created at %s\n", configPath)
	warningColor.Println("⚠️  Please edit the configuration file and add your API keys before running the process command.")

	return nil
}

func (app *App) loadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(data, &app.config)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// Set defaults for new fields
	if app.config.Anthropic.Timeout == 0 {
		app.config.Anthropic.Timeout = 120
	}
	if app.config.Anthropic.MaxTokens == 0 {
		app.config.Anthropic.MaxTokens = 4000
	}
	if app.config.Anthropic.ChunkSize == 0 {
		app.config.Anthropic.ChunkSize = 15000
	}
	if app.config.Anthropic.RetryCount == 0 {
		app.config.Anthropic.RetryCount = 3
	}
	if app.config.Anthropic.RetryDelay == 0 {
		app.config.Anthropic.RetryDelay = 5
	}
	if app.config.Jira.Timeout == 0 {
		app.config.Jira.Timeout = 30
	}
	if app.config.Processing.Mode == "" {
		app.config.Processing.Mode = "full"
	}
	if app.config.Processing.OutputDir == "" {
		app.config.Processing.OutputDir = "./output"
	}

	return nil
}

func (app *App) processProject(configPath, inputFile string, dryRun bool, mode, outputDir string) error {
	headerColor.Println("🤖 Starting Project Breakdown Process")

	// Load configuration
	infoColor.Println("📝 Loading configuration...")
	if err := app.loadConfig(configPath); err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Override config with CLI flags
	if mode != "" {
		app.config.Processing.Mode = mode
	}
	if outputDir != "" {
		app.config.Processing.OutputDir = outputDir
	}

	successColor.Printf("✅ Configuration loaded - Mode: %s\n", app.config.Processing.Mode)

	// Validate mode-specific requirements
	if app.config.Processing.Mode != "create-only" {
		if app.config.Anthropic.APIKey == "" || app.config.Anthropic.APIKey == "your-anthropic-api-key-here" {
			return fmt.Errorf("Anthropic API key not configured (required for analysis modes)")
		}
	}

	if app.config.Processing.Mode != "analyze-only" && !dryRun {
		if app.config.Jira.APIToken == "" || app.config.Jira.APIToken == "your-jira-api-token" {
			return fmt.Errorf("JIRA API token not configured (required for ticket creation)")
		}
	}

	// Create output directory
	if err := os.MkdirAll(app.config.Processing.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Read input file
	infoColor.Printf("📖 Reading input file: %s\n", inputFile)
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}
	successColor.Printf("✅ Read %d bytes from input file\n", len(content))

	var breakdown *ProjectBreakdown

	// Process based on mode
	switch app.config.Processing.Mode {
	case "analyze-only":
		breakdown, err = app.analyzeProject(string(content), inputFile)
		if err != nil {
			return fmt.Errorf("analysis failed: %v", err)
		}

		// Save analysis
		if err := app.saveAnalysis(breakdown, inputFile); err != nil {
			return fmt.Errorf("failed to save analysis: %v", err)
		}

		successColor.Println("🎉 Analysis completed and saved!")
		return nil

	case "create-only":
		return fmt.Errorf("create-only mode requires using 'create-from-analysis' command")

	case "full":
		breakdown, err = app.analyzeProject(string(content), inputFile)
		if err != nil {
			return fmt.Errorf("analysis failed: %v", err)
		}

		// Save intermediate results
		if app.config.Processing.SaveIntermediate {
			if err := app.saveAnalysis(breakdown, inputFile); err != nil {
				warningColor.Printf("⚠️  Failed to save intermediate analysis: %v\n", err)
			}
		}

		// Continue to JIRA creation
		break

	default:
		return fmt.Errorf("unknown processing mode: %s", app.config.Processing.Mode)
	}

	// Display breakdown
	app.displayBreakdown(breakdown)

	if dryRun {
		warningColor.Println("🔍 Dry run mode - no JIRA tickets will be created")
		return nil
	}

	// Confirm creation
	fmt.Print("\nDo you want to create these tickets in JIRA? (y/N): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if response != "y" && response != "yes" {
		infoColor.Println("Ticket creation cancelled.")
		return nil
	}

	// Create JIRA tickets
	infoColor.Println("🎫 Creating JIRA tickets...")
	if err := app.createJiraTickets(breakdown); err != nil {
		return fmt.Errorf("failed to create JIRA tickets: %v", err)
	}

	successColor.Println("🎉 Project breakdown completed successfully!")
	return nil
}

func (app *App) createFromAnalysis(configPath, analysisFile string, dryRun bool) error {
	headerColor.Println("🎫 Creating JIRA Tickets from Analysis")

	// Load configuration
	if err := app.loadConfig(configPath); err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if !dryRun {
		if app.config.Jira.APIToken == "" || app.config.Jira.APIToken == "your-jira-api-token" {
			return fmt.Errorf("JIRA API token not configured")
		}
	}

	// Read analysis file
	infoColor.Printf("📖 Reading analysis file: %s\n", analysisFile)
	content, err := os.ReadFile(analysisFile)
	if err != nil {
		return fmt.Errorf("failed to read analysis file: %v", err)
	}

	var breakdown ProjectBreakdown
	if err := json.Unmarshal(content, &breakdown); err != nil {
		return fmt.Errorf("failed to parse analysis file: %v", err)
	}

	successColor.Printf("✅ Loaded analysis for project: %s\n", breakdown.ProjectName)

	// Display breakdown
	app.displayBreakdown(&breakdown)

	if dryRun {
		warningColor.Println("🔍 Dry run mode - no JIRA tickets will be created")
		return nil
	}

	// Confirm creation
	fmt.Print("\nDo you want to create these tickets in JIRA? (y/N): ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(scanner.Text()))

	if response != "y" && response != "yes" {
		infoColor.Println("Ticket creation cancelled.")
		return nil
	}

	// Create JIRA tickets
	if err := app.createJiraTickets(&breakdown); err != nil {
		return fmt.Errorf("failed to create JIRA tickets: %v", err)
	}

	successColor.Println("🎉 JIRA tickets created successfully!")
	return nil
}

func (app *App) analyzeProject(content, inputFile string) (*ProjectBreakdown, error) {
	contentSize := len(content)
	chunkSize := app.config.Anthropic.ChunkSize

	if contentSize <= chunkSize {
		// Process as single chunk
		infoColor.Println("🧠 Processing with AI (single chunk)...")
		return app.processWithAI(content, 1, 1)
	}

	// Process in chunks
	chunks := app.splitIntoChunks(content, chunkSize)
	infoColor.Printf("🧠 Processing with AI (%d chunks)...\n", len(chunks))

	var allEpics []Epic
	var projectName, overview string

	for i, chunk := range chunks {
		infoColor.Printf("  Processing chunk %d/%d...\n", i+1, len(chunks))

		breakdown, err := app.processChunkWithRetry(chunk, i+1, len(chunks))
		if err != nil {
			return nil, fmt.Errorf("failed to process chunk %d: %v", i+1, err)
		}

		if i == 0 {
			projectName = breakdown.ProjectName
			overview = breakdown.Overview
		}

		// Add chunk information to epics
		for j := range breakdown.Epics {
			breakdown.Epics[j].Chunk = i + 1
		}

		allEpics = append(allEpics, breakdown.Epics...)

		// Save intermediate chunk results
		if app.config.Processing.SaveIntermediate {
			chunkAnalysis := ChunkAnalysis{
				ChunkIndex: i + 1,
				Content:    chunk,
				Epics:      breakdown.Epics,
			}
			app.saveChunkAnalysis(chunkAnalysis, inputFile, i+1)
		}

		// Rate limiting between chunks
		if i < len(chunks)-1 {
			time.Sleep(time.Duration(app.config.Anthropic.RetryDelay) * time.Second)
		}
	}

	// Merge and deduplicate epics
	mergedEpics := app.mergeEpics(allEpics)

	result := &ProjectBreakdown{
		ProjectName:    projectName,
		Overview:       overview,
		Epics:          mergedEpics,
		ProcessedAt:    time.Now(),
		TotalChunks:    len(chunks),
		InputFileSize:  int64(contentSize),
		ProcessingMode: app.config.Processing.Mode,
	}

	successColor.Printf("✅ AI processing complete - %d chunks processed, %d epics found\n",
		len(chunks), len(mergedEpics))

	return result, nil
}

func (app *App) splitIntoChunks(content string, chunkSize int) []string {
	if len(content) <= chunkSize {
		return []string{content}
	}

	var chunks []string
	lines := strings.Split(content, "\n")

	var currentChunk strings.Builder
	var currentSize int

	for _, line := range lines {
		lineSize := len(line) + 1 // +1 for newline

		// If adding this line would exceed chunk size, save current chunk
		if currentSize+lineSize > chunkSize && currentSize > 0 {
			chunks = append(chunks, currentChunk.String())
			currentChunk.Reset()
			currentSize = 0
		}

		currentChunk.WriteString(line + "\n")
		currentSize += lineSize
	}

	// Add the last chunk if it has content
	if currentSize > 0 {
		chunks = append(chunks, currentChunk.String())
	}

	return chunks
}

func (app *App) processChunkWithRetry(content string, chunkIndex, totalChunks int) (*ProjectBreakdown, error) {
	var lastErr error

	for attempt := 1; attempt <= app.config.Anthropic.RetryCount; attempt++ {
		breakdown, err := app.processWithAI(content, chunkIndex, totalChunks)
		if err == nil {
			return breakdown, nil
		}

		lastErr = err

		if attempt < app.config.Anthropic.RetryCount {
			warningColor.Printf("    Attempt %d failed, retrying in %d seconds...\n",
				attempt, app.config.Anthropic.RetryDelay)
			time.Sleep(time.Duration(app.config.Anthropic.RetryDelay) * time.Second)
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %v", app.config.Anthropic.RetryCount, lastErr)
}

func (app *App) processWithAI(content string, chunkIndex, totalChunks int) (*ProjectBreakdown, error) {
	var prompt string

	if totalChunks == 1 {
		prompt = fmt.Sprintf(`You are a senior project manager and technical lead. Analyze the following project description and break it down into actionable epics and user stories for a development team.

Project Description:
%s

Please respond with a JSON object that follows this exact structure:
{
  "project_name": "string",
  "overview": "brief project overview",
  "epics": [
    {
      "title": "Epic title",
      "description": "Detailed epic description",
      "priority": "High|Medium|Low",
      "stories": [
        {
          "title": "User story title",
          "description": "As a [user type], I want [goal] so that [benefit]",
          "priority": "High|Medium|Low",
          "story_points": 1-8,
          "acceptance_criteria": ["criteria1", "criteria2"],
          "dependencies": ["optional dependency references"]
        }
      ]
    }
  ]
}

Guidelines:
- Create 3-7 epics that represent major functional areas
- Each epic should have 3-8 user stories
- Story points should follow Fibonacci sequence (1,2,3,5,8)
- Write clear acceptance criteria for each story
- Identify dependencies between stories where relevant
- Prioritize based on business value and technical dependencies
- Use proper user story format: "As a [persona], I want [goal] so that [benefit]"

Respond ONLY with valid JSON. Do not include any markdown formatting or explanations.`, content)
	} else {
		prompt = fmt.Sprintf(`You are analyzing chunk %d of %d from a larger project description. Focus on the content in this chunk while being aware it's part of a larger project.

Content to analyze:
%s

Please respond with a JSON object focusing on epics and stories that can be derived from THIS SPECIFIC CONTENT:
{
  "project_name": "string (extract from content or use generic name)",
  "overview": "brief overview based on this chunk",
  "epics": [
    {
      "title": "Epic title (specific to this chunk's content)",
      "description": "Detailed epic description",
      "priority": "High|Medium|Low",
      "stories": [
        {
          "title": "User story title",
          "description": "As a [user type], I want [goal] so that [benefit]",
          "priority": "High|Medium|Low", 
          "story_points": 1-8,
          "acceptance_criteria": ["criteria1", "criteria2"],
          "dependencies": ["optional dependency references"]
        }
      ]
    }
  ]
}

Guidelines for chunk processing:
- Focus only on what's clearly described in this chunk
- Create 1-4 epics based on the chunk content
- Each epic should have 2-6 user stories
- Use story points (1,2,3,5,8) appropriate for individual stories
- Be specific about acceptance criteria based on chunk content
- If the chunk seems incomplete, create stories for what IS described

Respond ONLY with valid JSON. Do not include any markdown formatting or explanations.`, chunkIndex, totalChunks, content)
	}

	// Call Anthropic API
	reqBody := map[string]interface{}{
		"model":      app.config.Anthropic.Model,
		"max_tokens": app.config.Anthropic.MaxTokens,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", app.config.Anthropic.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: time.Duration(app.config.Anthropic.Timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode API response: %v", err)
	}

	if len(apiResponse.Content) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	// Parse the AI response
	var breakdown ProjectBreakdown
	responseText := strings.TrimSpace(apiResponse.Content[0].Text)

	// Remove any potential markdown formatting
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	if err := json.Unmarshal([]byte(responseText), &breakdown); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %v\nResponse: %s", err, responseText)
	}

	return &breakdown, nil
}

func (app *App) mergeEpics(epics []Epic) []Epic {
	epicMap := make(map[string]*Epic)

	for _, epic := range epics {
		key := strings.ToLower(strings.TrimSpace(epic.Title))

		if existing, exists := epicMap[key]; exists {
			// Merge stories, avoiding duplicates
			existing.Stories = append(existing.Stories, epic.Stories...)

			// Update description if the new one is more detailed
			if len(epic.Description) > len(existing.Description) {
				existing.Description = epic.Description
			}

			// Keep highest priority
			if epic.Priority == "High" || (epic.Priority == "Medium" && existing.Priority == "Low") {
				existing.Priority = epic.Priority
			}
		} else {
			// Create a copy to avoid modifying the original
			newEpic := epic
			epicMap[key] = &newEpic
		}
	}

	// Convert map back to slice
	var result []Epic
	for _, epic := range epicMap {
		// Deduplicate stories within each epic
		epic.Stories = app.deduplicateStories(epic.Stories)
		result = append(result, *epic)
	}

	return result
}

func (app *App) deduplicateStories(stories []Story) []Story {
	storyMap := make(map[string]Story)

	for _, story := range stories {
		key := strings.ToLower(strings.TrimSpace(story.Title))

		if existing, exists := storyMap[key]; exists {
			// Keep the story with more detailed information
			if len(story.Description) > len(existing.Description) ||
				len(story.AcceptanceCriteria) > len(existing.AcceptanceCriteria) {
				storyMap[key] = story
			}
		} else {
			storyMap[key] = story
		}
	}

	// Convert map back to slice
	var result []Story
	for _, story := range storyMap {
		result = append(result, story)
	}

	return result
}

func (app *App) saveAnalysis(breakdown *ProjectBreakdown, inputFile string) error {
	baseName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
	timestamp := breakdown.ProcessedAt.Format("20060102-150405")

	// Save JSON analysis
	jsonFile := filepath.Join(app.config.Processing.OutputDir, fmt.Sprintf("%s-analysis-%s.json", baseName, timestamp))
	jsonData, err := json.MarshalIndent(breakdown, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %v", err)
	}

	// Save markdown summary
	mdFile := filepath.Join(app.config.Processing.OutputDir, fmt.Sprintf("%s-summary-%s.md", baseName, timestamp))
	mdContent := app.generateMarkdownSummary(breakdown)

	if err := os.WriteFile(mdFile, []byte(mdContent), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %v", err)
	}

	successColor.Printf("✅ Analysis saved:\n")
	infoColor.Printf("  📄 JSON: %s\n", jsonFile)
	infoColor.Printf("  📝 Markdown: %s\n", mdFile)

	return nil
}

func (app *App) saveChunkAnalysis(chunk ChunkAnalysis, inputFile string, chunkIndex int) error {
	baseName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))
	timestamp := time.Now().Format("20060102-150405")

	chunkFile := filepath.Join(app.config.Processing.OutputDir,
		fmt.Sprintf("%s-chunk-%d-%s.json", baseName, chunkIndex, timestamp))

	chunkData, err := json.MarshalIndent(chunk, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal chunk JSON: %v", err)
	}

	return os.WriteFile(chunkFile, chunkData, 0644)
}

func (app *App) generateMarkdownSummary(breakdown *ProjectBreakdown) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("# %s - Project Breakdown\n\n", breakdown.ProjectName))
	md.WriteString(fmt.Sprintf("**Generated:** %s\n", breakdown.ProcessedAt.Format("2006-01-02 15:04:05")))
	md.WriteString(fmt.Sprintf("**Processing Mode:** %s\n", breakdown.ProcessingMode))
	if breakdown.TotalChunks > 0 {
		md.WriteString(fmt.Sprintf("**Total Chunks:** %d\n", breakdown.TotalChunks))
	}
	md.WriteString(fmt.Sprintf("**Input File Size:** %d bytes\n\n", breakdown.InputFileSize))

	md.WriteString("## Project Overview\n\n")
	md.WriteString(fmt.Sprintf("%s\n\n", breakdown.Overview))

	totalStories := 0
	totalPoints := 0

	for i, epic := range breakdown.Epics {
		md.WriteString(fmt.Sprintf("## Epic %d: %s\n\n", i+1, epic.Title))
		md.WriteString(fmt.Sprintf("**Priority:** %s\n", epic.Priority))
		if epic.Chunk > 0 {
			md.WriteString(fmt.Sprintf("**Source Chunk:** %d\n", epic.Chunk))
		}
		md.WriteString(fmt.Sprintf("**Description:** %s\n\n", epic.Description))

		md.WriteString("### Stories\n\n")

		for j, story := range epic.Stories {
			md.WriteString(fmt.Sprintf("#### %d.%d %s\n\n", i+1, j+1, story.Title))
			md.WriteString(fmt.Sprintf("- **Points:** %d\n", story.StoryPoints))
			md.WriteString(fmt.Sprintf("- **Priority:** %s\n", story.Priority))
			md.WriteString(fmt.Sprintf("- **Description:** %s\n\n", story.Description))

			if len(story.AcceptanceCriteria) > 0 {
				md.WriteString("**Acceptance Criteria:**\n")
				for _, criteria := range story.AcceptanceCriteria {
					md.WriteString(fmt.Sprintf("- %s\n", criteria))
				}
				md.WriteString("\n")
			}

			if len(story.Dependencies) > 0 {
				md.WriteString(fmt.Sprintf("**Dependencies:** %s\n\n", strings.Join(story.Dependencies, ", ")))
			}

			totalStories++
			totalPoints += story.StoryPoints
		}

		md.WriteString("---\n\n")
	}

	md.WriteString("## Summary\n\n")
	md.WriteString(fmt.Sprintf("- **Total Epics:** %d\n", len(breakdown.Epics)))
	md.WriteString(fmt.Sprintf("- **Total Stories:** %d\n", totalStories))
	md.WriteString(fmt.Sprintf("- **Total Story Points:** %d\n", totalPoints))

	return md.String()
}

func (app *App) displayBreakdown(breakdown *ProjectBreakdown) {
	headerColor.Printf("\n📋 Project Breakdown: %s\n", breakdown.ProjectName)
	fmt.Printf("Overview: %s\n", breakdown.Overview)

	if breakdown.TotalChunks > 0 {
		infoColor.Printf("Processed in %d chunks\n", breakdown.TotalChunks)
	}

	fmt.Println()

	totalStories := 0
	totalPoints := 0

	for i, epic := range breakdown.Epics {
		headerColor.Printf("Epic %d: %s\n", i+1, epic.Title)
		infoColor.Printf("Priority: %s", epic.Priority)
		if epic.Chunk > 0 {
			infoColor.Printf(" | Chunk: %d", epic.Chunk)
		}
		fmt.Printf("\nDescription: %s\n\n", epic.Description)

		for j, story := range epic.Stories {
			fmt.Printf("  Story %d.%d: %s\n", i+1, j+1, story.Title)
			fmt.Printf("    Points: %d | Priority: %s\n", story.StoryPoints, story.Priority)
			fmt.Printf("    Description: %s\n", story.Description)

			if len(story.AcceptanceCriteria) > 0 {
				fmt.Printf("    Acceptance Criteria:\n")
				for _, criteria := range story.AcceptanceCriteria {
					fmt.Printf("      • %s\n", criteria)
				}
			}

			if len(story.Dependencies) > 0 {
				fmt.Printf("    Dependencies: %s\n", strings.Join(story.Dependencies, ", "))
			}
			fmt.Println()

			totalStories++
			totalPoints += story.StoryPoints
		}
	}

	successColor.Printf("📊 Summary: %d epics, %d stories, %d story points total\n",
		len(breakdown.Epics), totalStories, totalPoints)
}

func (app *App) createJiraTickets(breakdown *ProjectBreakdown) error {
	// Test JIRA connection first
	infoColor.Println("🔍 Testing JIRA connection...")
	if err := app.testJiraConnection(); err != nil {
		return fmt.Errorf("JIRA connection test failed: %v", err)
	}
	successColor.Println("✅ JIRA connection successful")

	createdEpics := make(map[string]string) // epic title -> JIRA key

	// Create epics first
	for i, epic := range breakdown.Epics {
		infoColor.Printf("Creating epic %d/%d: %s\n", i+1, len(breakdown.Epics), epic.Title)

		epicKey, err := app.createJiraIssueWithRetry(epic.Title, epic.Description, "Epic", epic.Priority, "")
		if err != nil {
			return fmt.Errorf("failed to create epic '%s': %v", epic.Title, err)
		}

		createdEpics[epic.Title] = epicKey
		successColor.Printf("✅ Created epic: %s\n", epicKey)

		// Create stories for this epic
		for j, story := range epic.Stories {
			infoColor.Printf("  Creating story %d/%d: %s\n", j+1, len(epic.Stories), story.Title)

			// Format story description with acceptance criteria
			fullDescription := story.Description + "\n\n*Acceptance Criteria:*\n"
			for _, criteria := range story.AcceptanceCriteria {
				fullDescription += "• " + criteria + "\n"
			}

			if len(story.Dependencies) > 0 {
				fullDescription += "\n*Dependencies:* " + strings.Join(story.Dependencies, ", ")
			}

			storyKey, err := app.createJiraIssueWithRetry(story.Title, fullDescription, "Task", story.Priority, epicKey)
			if err != nil {
				warningColor.Printf("⚠️  Failed to create story '%s': %v\n", story.Title, err)
				continue
			}

			successColor.Printf("  ✅ Created story: %s\n", storyKey)

			// Small delay between story creation to avoid rate limits
			time.Sleep(100 * time.Millisecond)
		}

		// Delay between epics
		if i < len(breakdown.Epics)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return nil
}

func (app *App) createJiraIssueWithRetry(title, description, issueType, priority, epicLink string) (string, error) {
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		key, err := app.createJiraIssue(title, description, issueType, priority, epicLink)
		if err == nil {
			return key, nil
		}

		lastErr = err

		if attempt < 3 {
			time.Sleep(2 * time.Second)
		}
	}

	return "", fmt.Errorf("failed after 3 attempts: %v", lastErr)
}

func (app *App) createJiraIssue(title, description, issueType, priority, epicLink string) (string, error) {
	issue := JiraIssue{
		Fields: JiraFields{
			Project: JiraProject{
				Key: app.config.Jira.ProjectKey,
			},
			Summary:     title,
			Description: description,
			IssueType: JiraIssueType{
				Name: issueType,
			},
		},
	}

	// Set parent (epic) if provided and issue type is not Epic
	if epicLink != "" && issueType != "Epic" {
		issue.Fields.Parent = &JiraParent{Key: epicLink}
	}

	// Note: Priority field is commented out as it's not available in this JIRA project
	// if priority != "" {
	// 	issue.Fields.Priority = JiraPriority{Name: priority}
	// }

	jsonData, err := json.Marshal(issue)
	if err != nil {
		return "", fmt.Errorf("failed to marshal issue: %v", err)
	}

	url := fmt.Sprintf("%s/rest/api/2/issue", app.config.Jira.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(app.config.Jira.Username, app.config.Jira.APIToken)

	// Debug: Print request details (without sensitive data)
	infoColor.Printf("Making JIRA API request to: %s\n", url)
	infoColor.Printf("Project Key: %s, Issue Type: %s\n", app.config.Jira.ProjectKey, issueType)

	client := &http.Client{Timeout: time.Duration(app.config.Jira.Timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		errorColor.Printf("JIRA API Error - Status: %d, Response: %s\n", resp.StatusCode, string(body))
		return "", fmt.Errorf("JIRA API returned status %d: %s", resp.StatusCode, string(body))
	}

	var jiraResp JiraResponse
	if err := json.NewDecoder(resp.Body).Decode(&jiraResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return jiraResp.Key, nil
}

func (app *App) testJiraConnection() error {
	// First, test authentication by getting all accessible projects
	infoColor.Println("🔍 Testing JIRA authentication and listing accessible projects...")

	url := fmt.Sprintf("%s/rest/api/2/project", app.config.Jira.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create projects request: %v", err)
	}

	req.SetBasicAuth(app.config.Jira.Username, app.config.Jira.APIToken)

	client := &http.Client{Timeout: time.Duration(app.config.Jira.Timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("projects request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authentication failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the projects response
	var projects []struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return fmt.Errorf("failed to decode projects response: %v", err)
	}

	successColor.Printf("✅ Authentication successful! Found %d accessible projects:\n", len(projects))

	projectFound := false
	for _, project := range projects {
		if project.Key == app.config.Jira.ProjectKey {
			successColor.Printf("  ✅ %s (%s) - This is your configured project\n", project.Key, project.Name)
			projectFound = true
		} else {
			infoColor.Printf("  📋 %s (%s)\n", project.Key, project.Name)
		}
	}

	if !projectFound {
		warningColor.Printf("⚠️  Project key '%s' not found in accessible projects!\n", app.config.Jira.ProjectKey)
		warningColor.Println("Please update your config.yaml with one of the available project keys above.")
		return fmt.Errorf("project key '%s' not found in accessible projects", app.config.Jira.ProjectKey)
	}

	// Test specific project access
	infoColor.Printf("🔍 Testing access to project '%s'...\n", app.config.Jira.ProjectKey)
	url = fmt.Sprintf("%s/rest/api/2/project/%s", app.config.Jira.BaseURL, app.config.Jira.ProjectKey)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create project test request: %v", err)
	}

	req.SetBasicAuth(app.config.Jira.Username, app.config.Jira.APIToken)
	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("project test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("project access test failed with status %d: %s", resp.StatusCode, string(body))
	}

	successColor.Printf("✅ Successfully accessed project '%s'\n", app.config.Jira.ProjectKey)

	return nil
}
