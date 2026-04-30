package models

import "time"

type CreateRepoRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
}

type CreateRepoResponse struct {
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	CloneURL  string    `json:"clone_url"`
	CreatedAt time.Time `json:"created_at"`
	Files     []string  `json:"files_created"`
	Message   string    `json:"message"`
}

type RepoRecord struct {
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	CloneURL    string    `json:"clone_url"`
	Description string    `json:"description"`
	Private     bool      `json:"private"`
	CreatedAt   time.Time `json:"created_at"`
}

type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Uptime  string `json:"uptime"`
}

type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}
