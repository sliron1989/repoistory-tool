package api

import (
	"net/http"

	"github.com/sliron1989/repoistory-tool/internal/web"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/repos", h.CreateRepo)
	mux.HandleFunc("GET /api/v1/repos", h.ListRepos)
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /readyz", h.Ready)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
	mux.HandleFunc("GET /", web.ServeUI)

	return recoveryMiddleware(loggingMiddleware(mux))
}
