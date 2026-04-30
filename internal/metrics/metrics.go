package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ReposCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "repository_tool_repos_created_total",
			Help: "Total number of repositories created",
		},
		[]string{"status"},
	)

	RepoCreationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "repository_tool_repo_creation_duration_seconds",
			Help:    "Time taken to create a repository",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "repository_tool_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "repository_tool_http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	ActiveRepoCreations = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "repository_tool_active_repo_creations",
			Help: "Number of repository creations currently in progress",
		},
	)
)
