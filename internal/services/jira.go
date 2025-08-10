package services

import (
	"fmt"
	"strings"
	"time"

	"scrum-master/internal/config"
	"scrum-master/internal/helpers"
	"scrum-master/internal/models"
	"scrum-master/internal/repositories"
)

// JiraService handles JIRA business logic
type JiraService struct {
	repo   *repositories.JiraRepository
	config *config.JiraConfig
}

// NewJiraService creates a new JIRA service
func NewJiraService(jiraConfig *config.JiraConfig) *JiraService {
	return &JiraService{
		repo:   repositories.NewJiraRepository(jiraConfig),
		config: jiraConfig,
	}
}

// TestConnection tests the JIRA connection and validates project access
func (s *JiraService) TestConnection() error {
	helpers.PrintInfo("Testing JIRA authentication and listing accessible projects...")

	projects, err := s.repo.TestConnection()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	helpers.PrintSuccess("Authentication successful! Found %d accessible projects:", len(projects))

	projectFound := false
	for _, project := range projects {
		marker := "📋"
		if project.Key == s.config.ProjectKey {
			marker = "✅"
			projectFound = true
		}
		helpers.PrintInfo("  %s %s (%s)", marker, project.Key, project.Name)
	}

	if !projectFound {
		helpers.PrintWarning("Project key '%s' not found in accessible projects!", s.config.ProjectKey)
		helpers.PrintInfo("Please update your config.yaml with one of the available project keys above.")
		return fmt.Errorf("project key '%s' not found in accessible projects", s.config.ProjectKey)
	}

	helpers.PrintInfo("Testing access to project '%s'...", s.config.ProjectKey)
	if _, err := s.repo.GetProjectInfo(s.config.ProjectKey); err != nil {
		return fmt.Errorf("failed to access project: %w", err)
	}

	helpers.PrintSuccess("Successfully accessed project '%s'", s.config.ProjectKey)
	helpers.PrintSuccess("JIRA connection successful")
	return nil
}

// CreateIssueWithRetry creates a JIRA issue with retry logic
func (s *JiraService) CreateIssueWithRetry(title, description, issueType, priority, epicLink string) (string, error) {
	var lastErr error

	for attempt := 1; attempt <= 3; attempt++ {
		key, err := s.CreateIssue(title, description, issueType, priority, epicLink)
		if err == nil {
			return key, nil
		}

		lastErr = err
		helpers.PrintWarning("Attempt %d failed: %v", attempt, err)

		if attempt < 3 {
			time.Sleep(2 * time.Second)
		}
	}

	return "", fmt.Errorf("failed after 3 attempts: %w", lastErr)
}

// CreateIssue creates a single JIRA issue
func (s *JiraService) CreateIssue(title, description, issueType, priority, epicLink string) (string, error) {
	helpers.PrintInfo("Making JIRA API request to: %s/rest/api/2/issue", s.config.BaseURL)
	helpers.PrintInfo("Project Key: %s, Issue Type: %s", s.config.ProjectKey, issueType)

	issue := &models.JiraIssue{
		Fields: models.JiraFields{
			Project: models.JiraProject{
				Key: s.config.ProjectKey,
			},
			Summary:     title,
			Description: description,
			IssueType: models.JiraIssueType{
				Name: issueType,
			},
		},
	}

	// Set parent (epic) if provided and issue type is not Epic
	if epicLink != "" && issueType != "Epic" {
		issue.Fields.Parent = &models.JiraParent{Key: epicLink}
	}

	resp, err := s.repo.CreateIssue(issue)
	if err != nil {
		helpers.PrintError("JIRA API Error - Status: %v", err)
		return "", err
	}

	return resp.Key, nil
}

// CreateEpic creates an epic in JIRA
func (s *JiraService) CreateEpic(title, description, priority string) (string, error) {
	return s.CreateIssueWithRetry(title, description, "Epic", priority, "")
}

// CreateTask creates a task in JIRA
func (s *JiraService) CreateTask(title, description, priority, epicLink string) (string, error) {
	return s.CreateIssueWithRetry(title, description, "Task", priority, epicLink)
}

// CreateTicketsFromBreakdown creates JIRA tickets from a project breakdown
func (s *JiraService) CreateTicketsFromBreakdown(breakdown *models.ProjectBreakdown) error {
	createdEpics := make(map[string]string) // epic title -> JIRA key

	// Create epics first
	for i, epic := range breakdown.Epics {
		helpers.PrintProgress(i+1, len(breakdown.Epics), fmt.Sprintf("Creating epic: %s", epic.Title))

		epicKey, err := s.CreateEpic(epic.Title, epic.Description, epic.Priority)
		if err != nil {
			return fmt.Errorf("failed to create epic '%s': %w", epic.Title, err)
		}

		createdEpics[epic.Title] = epicKey
		helpers.PrintSuccess("Created epic: %s", epicKey)

		// Create stories for this epic
		for j, story := range epic.Stories {
			helpers.PrintProgress(j+1, len(epic.Stories), fmt.Sprintf("Creating story: %s", story.Title))

			// Format story description with acceptance criteria
			fullDescription := story.Description + "\n\n*Acceptance Criteria:*\n"
			for _, criteria := range story.AcceptanceCriteria {
				fullDescription += "• " + criteria + "\n"
			}

			if len(story.Dependencies) > 0 {
				fullDescription += "\n*Dependencies:* " + strings.Join(story.Dependencies, ", ")
			}

			storyKey, err := s.CreateTask(story.Title, fullDescription, story.Priority, epicKey)
			if err != nil {
				helpers.PrintWarning("Failed to create story '%s': %v", story.Title, err)
				continue
			}

			helpers.PrintSuccess("Created story: %s", storyKey)
		}
	}

	helpers.PrintSuccess("JIRA tickets created successfully!")
	return nil
}
