package repositories

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"scrum-master/internal/config"
	"scrum-master/internal/models"
)

// JiraRepository handles JIRA API interactions
type JiraRepository struct {
	config *config.JiraConfig
	client *http.Client
}

// NewJiraRepository creates a new JIRA repository
func NewJiraRepository(jiraConfig *config.JiraConfig) *JiraRepository {
	return &JiraRepository{
		config: jiraConfig,
		client: &http.Client{
			Timeout: time.Duration(jiraConfig.Timeout) * time.Second,
		},
	}
}

// TestConnection tests the JIRA connection and returns accessible projects
func (r *JiraRepository) TestConnection() ([]models.JiraProjectInfo, error) {
	url := fmt.Sprintf("%s/rest/api/2/project", r.config.BaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(r.config.Username, r.config.APIToken)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("JIRA API returned status %d: %s", resp.StatusCode, string(body))
	}

	var projects []models.JiraProjectInfo
	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return projects, nil
}

// GetProjectInfo gets information about a specific project
func (r *JiraRepository) GetProjectInfo(projectKey string) (*models.JiraProjectInfo, error) {
	url := fmt.Sprintf("%s/rest/api/2/project/%s", r.config.BaseURL, projectKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(r.config.Username, r.config.APIToken)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("project test failed with status %d: %s", resp.StatusCode, string(body))
	}

	var project models.JiraProjectInfo
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &project, nil
}

// GetIssueTypes gets available issue types for a project
func (r *JiraRepository) GetIssueTypes(projectKey string) ([]models.JiraIssueTypeInfo, error) {
	url := fmt.Sprintf("%s/rest/api/2/project/%s", r.config.BaseURL, projectKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(r.config.Username, r.config.APIToken)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("issue types test failed with status %d: %s", resp.StatusCode, string(body))
	}

	var projectInfo struct {
		IssueTypes []models.JiraIssueTypeInfo `json:"issueTypes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&projectInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return projectInfo.IssueTypes, nil
}

// CreateIssue creates a new JIRA issue
func (r *JiraRepository) CreateIssue(issue *models.JiraIssue) (*models.JiraResponse, error) {
	jsonData, err := json.Marshal(issue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal issue: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/2/issue", r.config.BaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(r.config.Username, r.config.APIToken)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("JIRA API returned status %d: %s", resp.StatusCode, string(body))
	}

	var jiraResp models.JiraResponse
	if err := json.NewDecoder(resp.Body).Decode(&jiraResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &jiraResp, nil
}
