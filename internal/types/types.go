package types

// Session represents a single OpenCode session for export/import.
type Session struct {
	ID           string  `json:"id"`
	ProjectID    string  `json:"project_id"`
	ParentID     string  `json:"parent_id,omitempty"`
	Slug         string  `json:"slug"`
	Directory    string  `json:"directory"`
	Path         string  `json:"path,omitempty"`
	Title        string  `json:"title"`
	Agent        string  `json:"agent,omitempty"`
	Model        string  `json:"model,omitempty"`
	Cost         float64 `json:"cost"`
	TokensInput  int64   `json:"tokens_input"`
	TokensOutput int64   `json:"tokens_output"`
	TimeCreated  int64   `json:"time_created"`
	TimeUpdated  int64   `json:"time_updated"`
}

type Message struct {
	ID          string `json:"id"`
	SessionID   string `json:"session_id"`
	TimeCreated int64  `json:"time_created"`
	TimeUpdated int64  `json:"time_updated"`
	Data        string `json:"data"`
}

// Part is a part of a message (text, tool call, etc.).
type Part struct {
	ID          string `json:"id"`
	MessageID   string `json:"message_id"`
	SessionID   string `json:"session_id"`
	TimeCreated int64  `json:"time_created"`
	TimeUpdated int64  `json:"time_updated"`
	Data        string `json:"data"`
}

type Todo struct {
	ID          string `json:"id"`
	SessionID   string `json:"session_id"`
	Content     string `json:"content"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	Position    int    `json:"position"`
	TimeCreated int64  `json:"time_created"`
	TimeUpdated int64  `json:"time_updated"`
}

// Project describes the project a session belongs to.
type Project struct {
	ID          string `json:"id"`
	Worktree    string `json:"worktree"`
	VCS         string `json:"vcs,omitempty"`
	Name        string `json:"name,omitempty"`
	TimeCreated int64  `json:"time_created"`
	TimeUpdated int64  `json:"time_updated"`
}

const ExportVersion = 1

type SessionExport struct {
	Version  int       `json:"version"`
	Session  Session   `json:"session"`
	Project  Project   `json:"project"`
	Messages []Message `json:"messages"`
	Parts    []Part    `json:"parts"`
	Todos    []Todo    `json:"todos,omitempty"`
	Source   string    `json:"source"`
}
