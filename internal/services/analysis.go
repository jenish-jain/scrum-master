package services

import (
	"fmt"
	"strings"
	"time"

	"scrum-master/internal/config"
	"scrum-master/internal/helpers"
	"scrum-master/internal/models"
)

// AnalysisService handles project analysis and breakdown
type AnalysisService struct {
	config    *config.Config
	aiService *AIService
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(config *config.Config) *AnalysisService {
	return &AnalysisService{
		config:    config,
		aiService: NewAIService(&config.Anthropic),
	}
}

// DisplayProjectBreakdown displays the project breakdown in a formatted way
func (s *AnalysisService) DisplayProjectBreakdown(breakdown *models.ProjectBreakdown) {
	helpers.PrintTitle("Project Breakdown: %s", breakdown.ProjectName)
	helpers.PrintInfo("Overview: %s", breakdown.Overview)
	helpers.PrintInfo("Processed in %d chunks", breakdown.ProcessedChunks)
	helpers.PrintSeparator()

	for i, epic := range breakdown.Epics {
		helpers.PrintInfo("Epic %d: %s", i+1, epic.Title)
		helpers.PrintInfo("Priority: %s | Chunk: %d", epic.Priority, epic.Chunk)
		helpers.PrintInfo("Description: %s", epic.Description)
		helpers.PrintSeparator()

		for j, story := range epic.Stories {
			helpers.PrintInfo("  Story %d.%d: %s", i+1, j+1, story.Title)
			helpers.PrintInfo("    Points: %d | Priority: %s", story.StoryPoints, story.Priority)
			helpers.PrintInfo("    Description: %s", story.Description)
			helpers.PrintSeparator()

			if len(story.AcceptanceCriteria) > 0 {
				helpers.PrintInfo("    Acceptance Criteria:")
				for _, criteria := range story.AcceptanceCriteria {
					helpers.PrintInfo("      • %s", criteria)
				}
			}

			if len(story.Dependencies) > 0 {
				helpers.PrintInfo("    Dependencies: %s", strings.Join(story.Dependencies, ", "))
			}
			helpers.PrintSeparator()
		}
	}

	helpers.PrintInfo("Summary: %d epics, %d stories, %d story points total",
		breakdown.TotalEpics, breakdown.TotalStories, breakdown.TotalStoryPoints)
}

// SaveAnalysisResult saves the analysis result to files
func (s *AnalysisService) SaveAnalysisResult(breakdown *models.ProjectBreakdown, outputDir string) error {
	if err := helpers.EnsureDir(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create analysis result
	result := &models.AnalysisResult{
		ProjectBreakdown: *breakdown,
		AnalysisTime:     time.Now(),
		ProcessingMode:   s.config.Processing.Mode,
	}

	// Save full analysis
	fullAnalysisFilename := helpers.GenerateOutputFilename("project-desc-analysis", "json")
	fullAnalysisPath := helpers.GetOutputPath(outputDir, fullAnalysisFilename)

	if err := helpers.SaveJSON(result, fullAnalysisPath); err != nil {
		return fmt.Errorf("failed to save full analysis: %w", err)
	}

	helpers.PrintSuccess("Saved full analysis to: %s", fullAnalysisPath)

	// Save summary
	summaryFilename := helpers.GenerateOutputFilename("project-desc-summary", "md")
	summaryPath := helpers.GetOutputPath(outputDir, summaryFilename)

	if err := s.saveSummary(breakdown, summaryPath); err != nil {
		return fmt.Errorf("failed to save summary: %w", err)
	}

	helpers.PrintSuccess("Saved summary to: %s", summaryPath)
	return nil
}

// saveSummary saves a markdown summary of the analysis
func (s *AnalysisService) saveSummary(breakdown *models.ProjectBreakdown, filepath string) error {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("# %s\n\n", breakdown.ProjectName))
	summary.WriteString(fmt.Sprintf("**Overview:** %s\n\n", breakdown.Overview))
	summary.WriteString(fmt.Sprintf("**Total Epics:** %d\n", breakdown.TotalEpics))
	summary.WriteString(fmt.Sprintf("**Total Stories:** %d\n", breakdown.TotalStories))
	summary.WriteString(fmt.Sprintf("**Total Story Points:** %d\n\n", breakdown.TotalStoryPoints))

	for i, epic := range breakdown.Epics {
		summary.WriteString(fmt.Sprintf("## Epic %d: %s\n\n", i+1, epic.Title))
		summary.WriteString(fmt.Sprintf("**Priority:** %s | **Chunk:** %d\n\n", epic.Priority, epic.Chunk))
		summary.WriteString(fmt.Sprintf("%s\n\n", epic.Description))

		for j, story := range epic.Stories {
			summary.WriteString(fmt.Sprintf("### Story %d.%d: %s\n\n", i+1, j+1, story.Title))
			summary.WriteString(fmt.Sprintf("**Points:** %d | **Priority:** %s\n\n", story.StoryPoints, story.Priority))
			summary.WriteString(fmt.Sprintf("%s\n\n", story.Description))

			if len(story.AcceptanceCriteria) > 0 {
				summary.WriteString("**Acceptance Criteria:**\n")
				for _, criteria := range story.AcceptanceCriteria {
					summary.WriteString(fmt.Sprintf("- %s\n", criteria))
				}
				summary.WriteString("\n")
			}

			if len(story.Dependencies) > 0 {
				summary.WriteString(fmt.Sprintf("**Dependencies:** %s\n\n", strings.Join(story.Dependencies, ", ")))
			}
		}
	}

	return helpers.SaveJSON(summary.String(), filepath)
}

// ProcessProject processes a project description file with AI analysis
func (s *AnalysisService) ProcessProject(inputFile string) (*models.ProjectBreakdown, error) {
	// Read the input file
	content, err := helpers.ReadFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	helpers.PrintInfo("Read %d bytes from input file", len(content))

	// Determine if we need to chunk the content
	chunks := s.chunkContent(content)
	helpers.PrintInfo("Processing with AI (%d chunks)...", len(chunks))

	var allEpics []models.Epic
	var projectName string
	var overview string
	totalStories := 0
	totalStoryPoints := 0

	// Process each chunk
	for i, chunk := range chunks {
		helpers.PrintProgress(i+1, len(chunks), fmt.Sprintf("Processing chunk %d", i+1))

		// Process chunk with AI
		breakdown, err := s.aiService.ProcessWithRetry(chunk, i+1, len(chunks))
		if err != nil {
			return nil, fmt.Errorf("failed to process chunk %d: %w", i+1, err)
		}

		// Save intermediate results if enabled
		if s.config.Processing.SaveIntermediate {
			intermediateFilename := helpers.GenerateOutputFilename(fmt.Sprintf("chunk-%d", i+1), "json")
			intermediatePath := helpers.GetOutputPath(s.config.Processing.OutputDir, intermediateFilename)

			if err := helpers.SaveJSON(breakdown, intermediatePath); err != nil {
				helpers.PrintWarning("Failed to save intermediate result: %v", err)
			}
		}

		// Collect epics and update totals
		allEpics = append(allEpics, breakdown.Epics...)
		totalStories += breakdown.TotalStories
		totalStoryPoints += breakdown.TotalStoryPoints

		// Use the first chunk's project name and overview, or merge if needed
		if i == 0 {
			projectName = breakdown.ProjectName
			overview = breakdown.Overview
		} else if breakdown.ProjectName != projectName {
			// If project names differ, create a merged name
			projectName = fmt.Sprintf("%s (Merged)", projectName)
		}
	}

	// Merge and deduplicate epics
	mergedEpics := s.aiService.MergeEpics(allEpics)

	// Calculate final totals
	finalTotalStories := 0
	finalTotalStoryPoints := 0
	for _, epic := range mergedEpics {
		finalTotalStories += len(epic.Stories)
		for _, story := range epic.Stories {
			finalTotalStoryPoints += story.StoryPoints
		}
	}

	// Create final breakdown
	finalBreakdown := &models.ProjectBreakdown{
		ProjectName:      projectName,
		Overview:         overview,
		Epics:            mergedEpics,
		TotalEpics:       len(mergedEpics),
		TotalStories:     finalTotalStories,
		TotalStoryPoints: finalTotalStoryPoints,
		ProcessedChunks:  len(chunks),
	}

	helpers.PrintSuccess("AI processing complete - %d chunks processed, %d epics found", len(chunks), len(mergedEpics))
	return finalBreakdown, nil
}

// chunkContent splits content into chunks if it's too large
func (s *AnalysisService) chunkContent(content string) []string {
	if len(content) <= s.config.Anthropic.ChunkSizeChars {
		return []string{content}
	}

	var chunks []string
	chunkSize := s.config.Anthropic.ChunkSizeChars
	overlap := chunkSize / 4 // 25% overlap

	for i := 0; i < len(content); i += chunkSize - overlap {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		chunks = append(chunks, content[i:end])
	}

	return chunks
}
