package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"scrum-master/internal/config"
	"scrum-master/internal/helpers"
	"scrum-master/internal/models"
	"scrum-master/internal/services"

	"github.com/spf13/cobra"
)

var (
	configFile string
	dryRun     bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "scrum-master",
		Short: "Scrum Master - AI-powered project breakdown and JIRA integration",
		Long: `Scrum Master is a tool that uses AI to analyze project descriptions 
and automatically create JIRA epics and stories with proper linking.`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "config.yaml", "Configuration file path")

	// Process command
	var processCmd = &cobra.Command{
		Use:   "process",
		Short: "Process a project description file",
		Long:  "Analyze a project description and create a breakdown of epics and stories",
		Args:  cobra.ExactArgs(1),
		RunE:  runProcess,
	}
	processCmd.Flags().StringP("mode", "m", "full", "Processing mode (analyze-only, full)")
	rootCmd.AddCommand(processCmd)

	// Create from analysis command
	var createFromAnalysisCmd = &cobra.Command{
		Use:   "create-from-analysis",
		Short: "Create JIRA tickets from an analysis file",
		Long:  "Load an analysis file and create JIRA epics and stories",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreateFromAnalysis,
	}
	createFromAnalysisCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Show what would be created without actually creating JIRA tickets")
	rootCmd.AddCommand(createFromAnalysisCmd)

	if err := rootCmd.Execute(); err != nil {
		helpers.PrintError("Error: %v", err)
		os.Exit(1)
	}
}

func runProcess(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	mode, _ := cmd.Flags().GetString("mode")

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	helpers.PrintTitle("Processing Project Description")
	helpers.PrintInfo("Input file: %s", inputFile)
	helpers.PrintInfo("Mode: %s", mode)

	// Create analysis service
	analysisService := services.NewAnalysisService(cfg)

	// Process the project with AI
	breakdown, err := analysisService.ProcessProject(inputFile)
	if err != nil {
		return fmt.Errorf("failed to process project: %w", err)
	}

	// Display breakdown
	analysisService.DisplayProjectBreakdown(breakdown)

	// Save results
	if err := analysisService.SaveAnalysisResult(breakdown, cfg.Processing.OutputDir); err != nil {
		return fmt.Errorf("failed to save analysis result: %w", err)
	}

	helpers.PrintSuccess("Processing completed successfully!")
	return nil
}

func runCreateFromAnalysis(cmd *cobra.Command, args []string) error {
	analysisFile := args[0]

	// Load configuration
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	helpers.PrintTitle("Creating JIRA Tickets from Analysis")
	helpers.PrintInfo("Analysis file: %s", analysisFile)

	// Load analysis result
	var result models.AnalysisResult
	if err := helpers.LoadJSON(analysisFile, &result); err != nil {
		return fmt.Errorf("failed to load analysis file: %w", err)
	}

	helpers.PrintSuccess("Loaded analysis for project: %s", result.ProjectBreakdown.ProjectName)

	// Display breakdown
	analysisService := services.NewAnalysisService(cfg)
	analysisService.DisplayProjectBreakdown(&result.ProjectBreakdown)

	if dryRun {
		helpers.PrintInfo("Dry run mode - no JIRA tickets will be created")
		return nil
	}

	// Confirm with user
	if !confirmCreation() {
		helpers.PrintInfo("Operation cancelled by user")
		return nil
	}

	// Test JIRA connection
	jiraService := services.NewJiraService(&cfg.Jira)
	if err := jiraService.TestConnection(); err != nil {
		return fmt.Errorf("failed to create JIRA tickets: %w", err)
	}

	// Create tickets
	if err := jiraService.CreateTicketsFromBreakdown(&result.ProjectBreakdown); err != nil {
		return fmt.Errorf("failed to create JIRA tickets: %w", err)
	}

	return nil
}

func confirmCreation() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to create these tickets in JIRA? (y/N): ")
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}
