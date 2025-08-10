package models

import "time"

// ProjectBreakdown represents the complete project structure
type ProjectBreakdown struct {
	ProjectName      string `json:"project_name"`
	Overview         string `json:"overview"`
	Epics            []Epic `json:"epics"`
	TotalEpics       int    `json:"total_epics"`
	TotalStories     int    `json:"total_stories"`
	TotalStoryPoints int    `json:"total_story_points"`
	ProcessedChunks  int    `json:"processed_chunks"`
}

// Epic represents a project epic
type Epic struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	Chunk       int     `json:"chunk"`
	Stories     []Story `json:"stories"`
}

// Story represents a user story
type Story struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	StoryPoints        int      `json:"story_points"`
	Priority           string   `json:"priority"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	Dependencies       []string `json:"dependencies"`
}

// AnalysisResult represents the analysis output
type AnalysisResult struct {
	ProjectBreakdown ProjectBreakdown `json:"project_breakdown"`
	AnalysisTime     time.Time        `json:"analysis_time"`
	ProcessingMode   string           `json:"processing_mode"`
}
