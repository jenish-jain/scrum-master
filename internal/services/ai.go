package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"scrum-master/internal/config"
	"scrum-master/internal/helpers"
	"scrum-master/internal/models"
)

// AIService handles AI-powered project analysis
type AIService struct {
	config *config.AnthropicConfig
	client *http.Client
}

// NewAIService creates a new AI service
func NewAIService(anthropicConfig *config.AnthropicConfig) *AIService {
	return &AIService{
		config: anthropicConfig,
		client: &http.Client{
			Timeout: time.Duration(anthropicConfig.TimeoutSeconds) * time.Second,
		},
	}
}

// ProcessWithAI analyzes project content and returns a breakdown
func (s *AIService) ProcessWithAI(content string, chunkIndex, totalChunks int) (*models.ProjectBreakdown, error) {
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
		"model":      s.config.Model,
		"max_tokens": s.config.MaxTokens,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
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
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	if len(apiResponse.Content) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	// Parse the AI response
	var breakdown models.ProjectBreakdown
	responseText := strings.TrimSpace(apiResponse.Content[0].Text)

	// Remove any potential markdown formatting
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	if err := json.Unmarshal([]byte(responseText), &breakdown); err != nil {
		return nil, fmt.Errorf("failed to parse AI response as JSON: %w\nResponse: %s", err, responseText)
	}

	return &breakdown, nil
}

// ProcessWithRetry processes content with retry logic
func (s *AIService) ProcessWithRetry(content string, chunkIndex, totalChunks int) (*models.ProjectBreakdown, error) {
	var lastErr error

	for attempt := 1; attempt <= s.config.RetryCount; attempt++ {
		helpers.PrintInfo("Processing chunk %d/%d (attempt %d/%d)...", chunkIndex, totalChunks, attempt, s.config.RetryCount)

		breakdown, err := s.ProcessWithAI(content, chunkIndex, totalChunks)
		if err == nil {
			return breakdown, nil
		}

		lastErr = err
		helpers.PrintWarning("Attempt %d failed: %v", attempt, err)

		if attempt < s.config.RetryCount {
			helpers.PrintInfo("Retrying in %d seconds...", s.config.RetryDelaySeconds)
			time.Sleep(time.Duration(s.config.RetryDelaySeconds) * time.Second)
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", s.config.RetryCount, lastErr)
}

// MergeEpics merges multiple epics, deduplicating and combining stories
func (s *AIService) MergeEpics(epics []models.Epic) []models.Epic {
	epicMap := make(map[string]*models.Epic)

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
	var result []models.Epic
	for _, epic := range epicMap {
		// Deduplicate stories within each epic
		epic.Stories = s.deduplicateStories(epic.Stories)
		result = append(result, *epic)
	}

	return result
}

// deduplicateStories removes duplicate stories based on title
func (s *AIService) deduplicateStories(stories []models.Story) []models.Story {
	storyMap := make(map[string]models.Story)

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
	var result []models.Story
	for _, story := range storyMap {
		result = append(result, story)
	}

	return result
}
