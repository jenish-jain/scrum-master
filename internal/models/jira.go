package models

// JiraIssue represents a JIRA issue
type JiraIssue struct {
	Fields JiraFields `json:"fields"`
}

// JiraFields represents JIRA issue fields
type JiraFields struct {
	Project     JiraProject   `json:"project"`
	Summary     string        `json:"summary"`
	Description string        `json:"description"`
	IssueType   JiraIssueType `json:"issuetype"`
	Parent      *JiraParent   `json:"parent,omitempty"`
}

// JiraProject represents a JIRA project
type JiraProject struct {
	Key string `json:"key"`
}

// JiraIssueType represents a JIRA issue type
type JiraIssueType struct {
	Name string `json:"name"`
}

// JiraParent represents a JIRA parent issue
type JiraParent struct {
	Key string `json:"key"`
}

// JiraResponse represents a JIRA API response
type JiraResponse struct {
	ID  string `json:"id"`
	Key string `json:"key"`
}

// JiraProjectInfo represents JIRA project information
type JiraProjectInfo struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// JiraIssueTypeInfo represents JIRA issue type information
type JiraIssueTypeInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
