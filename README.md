# Repository-Tool - Automated GitHub Repository Setup Platform

A platform microservice that automates the creation of new GitHub repositories with production-ready defaults: standard files, branch protection rules, CI/CD workflows, and a working generated microservice.

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                Repository-Tool Platform              │
│                                                     │
│  ┌──────────┐   ┌──────────────┐   ┌─────────────┐ │
│  │ Web UI   │──▸│  REST API    │──▸│   GitHub     │ │
│  │ (embed)  │   │  /api/v1/*   │   │   Service    │ │
│  └──────────┘   └──────┬───────┘   └──────┬──────┘ │
│                        │                   │        │
│  ┌──────────┐   ┌──────┴───────┐   ┌──────┴──────┐ │
│  │ /metrics │   │  Middleware   │   │  Templates  │ │
│  │ Prometheus│  │  Log+Metrics │   │  (Go files) │ │
│  └──────────┘   └──────────────┘   └─────────────┘ │
│                                                     │
│                    kind cluster                      │
│                  (Kubernetes)                        │
└─────────────────────────────────────────────────────┘
                         │
                         ▼
              ┌──────────────────┐
              │   GitHub API     │
              │                  │
              │  • Create repo   │
              │  • Commit files  │
              │  • Set protection│
              └──────────────────┘
```

### Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Go stdlib router** (`net/http.ServeMux`) | Go 1.22+ supports method-based routing natively (`"POST /api/v1/repos"`). No framework dependency needed. |
| **In-memory state** | For a local platform tool, tracking created repos in-memory is sufficient. Production would use PostgreSQL/Redis. |
| **Sequential file commits** | Uses GitHub Contents API (one commit per file) instead of Git Trees API (batch commit). Simpler code, acceptable for bootstrapping. |
| **Embedded HTML** | `//go:embed` bundles the web UI into the binary. Single artifact, no separate frontend build. |
| **NodePort service** | Simplest way to expose a service from kind. Production would use Ingress + TLS. |
| **No authentication on platform API** | Local-only deployment in kind. Production would require API keys or OAuth2. |

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [kind](https://kind.sigs.k8s.io/) (`brew install kind`)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) (`brew install kubectl`)
- [Helm](https://helm.sh/docs/intro/install/) (`brew install helm`)
- A GitHub Personal Access Token (see below)

## GitHub Personal Access Token Setup

1. Go to [GitHub Settings > Developer settings > Personal access tokens > Tokens (classic)](https://github.com/settings/tokens)
2. Click **Generate new token (classic)**
3. Give it a descriptive name (e.g., "Repository-Tool Platform")
4. Select these scopes:
   - `repo` (Full control of private repositories)
   - `workflow` (Update GitHub Action workflows)
5. Click **Generate token** and copy it immediately

> **Security note**: Never commit your token. The deployment uses Kubernetes Secrets.

## Quick Start

### Option A: Deploy on Kubernetes (Recommended)

```bash
# 1. Create the kind cluster
make cluster

# 2. Set your GitHub PAT and deploy via Helm
export GITHUB_TOKEN="ghp_your_token_here"
make deploy

# 3. Wait for the pod to be ready
kubectl -n repository-tool get pods -w

# 4. Access the service
open http://localhost:8080
```

Or do it all in one step:
```bash
export GITHUB_TOKEN="ghp_your_token_here"
make all
```

### Option B: Run Locally

```bash
# Set your GitHub token
export GITHUB_TOKEN="ghp_your_token_here"

# Run the server
make run

# Or build and run the binary
make build
./bin/repository-tool
```

### Option C: Run with Docker

```bash
docker build -t repository-tool:latest .
docker run -p 8080:8080 -e GITHUB_TOKEN="ghp_your_token_here" repository-tool:latest
```

## API Usage

### Create a Repository

```bash
curl -X POST http://localhost:8080/api/v1/repos \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-new-service",
    "description": "A shiny new microservice",
    "private": false
  }'
```

**Response:**
```json
{
  "name": "my-new-service",
  "url": "https://github.com/youruser/my-new-service",
  "clone_url": "https://github.com/youruser/my-new-service.git",
  "created_at": "2026-04-30T12:00:00Z",
  "files_created": [
    "README.md", ".gitignore", "LICENSE", "go.mod",
    "main.go", "Dockerfile", ".github/workflows/ci.yml"
  ],
  "message": "Repository created successfully with branch protection and CI/CD workflow"
}
```

### List Created Repositories

```bash
curl http://localhost:8080/api/v1/repos
```

### Health Check

```bash
curl http://localhost:8080/health
```

### Prometheus Metrics

```bash
curl http://localhost:8080/metrics
```

### Swagger UI

Interactive API documentation is available at [http://localhost:8080/swagger/](http://localhost:8080/swagger/).

### Web Interface

Open [http://localhost:8080](http://localhost:8080) in your browser.

## Using the Generated Microservice

After creating a repository via the API:

```bash
# 1. Clone the generated repository
git clone https://github.com/youruser/my-new-service.git
cd my-new-service

# 2. Run the service
go run main.go

# 3. Test it
curl http://localhost:8080/
# {"message":"Hello from my-new-service!","service":"my-new-service"}

curl http://localhost:8080/health
# {"status":"ok","uptime":"5s"}
```

The generated repository includes:
- **main.go** - Working Go HTTP server with `/` and `/health` endpoints
- **go.mod** - Go module definition
- **Dockerfile** - Multi-stage Docker build
- **README.md** - Documentation with run instructions
- **.gitignore** - Standard Go gitignore
- **LICENSE** - MIT license
- **.github/workflows/ci.yml** - GitHub Actions CI that runs golangci-lint on PRs

### Branch Protection

Each generated repository has branch protection on `main`:
- Requires 1 approving pull request review
- Requires the `lint` status check to pass
- Direct pushes to main are discouraged (PRs required)

## Observability

### Structured Logging

All requests and operations are logged with structured fields (using `log/slog`):

```json
{
  "time": "2026-04-30T12:00:00Z",
  "level": "INFO",
  "msg": "request",
  "method": "POST",
  "path": "/api/v1/repos",
  "status": 201,
  "duration_ms": 3200,
  "remote_addr": "10.244.0.1:45032"
}
```

Set `LOG_FORMAT=json` for JSON output (default in K8s), or leave unset for human-readable text.

### Prometheus Metrics

The `/metrics` endpoint exposes:

| Metric | Type | Description |
|--------|------|-------------|
| `repository_tool_repos_created_total` | Counter | Total repositories created (by status) |
| `repository_tool_repo_creation_duration_seconds` | Histogram | Time to create a repository |
| `repository_tool_http_requests_total` | Counter | HTTP requests (by method, path, status) |
| `repository_tool_http_request_duration_seconds` | Histogram | HTTP request latency |
| `repository_tool_active_repo_creations` | Gauge | In-progress repository creations |

These metrics enable:
- Alerting on error rates or slow creation times
- Dashboard for repo creation trends (Business Intelligence)
- SLO tracking for the platform service

### Business Intelligence

The service tracks repository creation data that can inform developer experience decisions:
- **Creation volume**: How many repos are being created and when (visible via Prometheus counters)
- **Error rates**: Which failures occur most frequently (API errors, validation errors)
- **Creation latency**: How long the provisioning takes (histogram for p50/p95/p99)
- **Repository history**: GET `/api/v1/repos` returns all repos created in the current session

In production, this data would be persisted to a database and exported to a BI tool for trend analysis.

## Security

### Credential Management
- GitHub PAT is stored as a **Kubernetes Secret** (created by Helm) and injected via environment variable
- The token is passed via `--set github.token=...` at deploy time — never committed to values.yaml
- The token is never logged or returned in API responses
- In production, use sealed-secrets or an external secret manager

### Input Validation
- Repository names are validated against GitHub naming rules (alphanumeric, hyphens, dots, underscores; 1-100 chars)
- Description length is capped at 350 characters
- Request body size is bounded by Go's default HTTP limits

### Generated Repository Security
- `.gitignore` excludes common secret patterns (`.env`, credentials)
- README includes security guidance
- CI workflow enforces code quality via linting

### Production Considerations
- Add API authentication (API keys, OAuth2, or mTLS)
- Use HTTPS (TLS termination at Ingress)
- Implement rate limiting
- Rotate GitHub PATs regularly
- Use GitHub App instead of PAT for better audit trails
- Scan generated repos with tools like Trivy or Snyk

## GitOps Readiness

The deployment is structured for GitOps adoption:

- **Helm chart** in `deploy/repository-tool/` — compatible with Flux and ArgoCD
- All manifests are **declarative** and templated via Helm values
- Image tag and all configuration can be overridden per environment via values files
- Generated repositories could include their own Helm chart for GitOps-managed deployment

To use with ArgoCD:
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: repository-tool
spec:
  source:
    repoURL: https://github.com/youruser/repository-tool.git
    path: deploy/repository-tool
    helm:
      valueFiles:
        - values.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: repository-tool
```

## Project Structure

```
repository-tool/
├── cmd/server/main.go           # Application entrypoint
├── internal/
│   ├── api/
│   │   ├── handler.go           # HTTP request handlers
│   │   ├── middleware.go        # Logging, metrics, recovery
│   │   └── routes.go            # Route definitions
│   ├── github/
│   │   ├── client.go            # GitHub API client
│   │   ├── repo.go              # Repository creation logic
│   │   ├── protection.go        # Branch protection rules
│   │   └── files.go             # File content commits
│   ├── templates/               # Generated file templates
│   │   ├── templates.go         # Data structures
│   │   ├── readme.go            # README template
│   │   ├── gitignore.go         # .gitignore template
│   │   ├── license.go           # MIT license template
│   │   ├── ciworkflow.go        # GitHub Actions CI
│   │   ├── microservice.go      # Go HTTP server template
│   │   └── dockerfile.go        # Dockerfile template
│   ├── metrics/metrics.go       # Prometheus metrics
│   ├── models/models.go         # Data models
│   └── web/
│       ├── embed.go             # Embedded UI handler
│       └── static/index.html    # Web interface
├── deploy/
│   ├── kind-config.yaml         # kind cluster port mapping
│   └── repository-tool/         # Helm chart
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
│           ├── _helpers.tpl
│           ├── namespace.yaml
│           ├── secret.yaml
│           ├── deployment.yaml
│           └── service.yaml
├── Dockerfile                   # Platform service container
├── Makefile                     # Build & deploy commands
└── README.md                    # This file
```

## Assumptions & Trade-offs

| Assumption | Impact |
|-----------|--------|
| Single user / local deployment | No auth, in-memory state, NodePort access |
| Go-only generated services | Template system generates Go microservices; extensible to other languages |
| GitHub.com (not Enterprise) | Uses public GitHub API; configurable for GHE |
| Sequential file commits | Multiple commits instead of one atomic commit; simpler but noisier git history |
| No persistent storage | Repository history lost on pod restart; production would use a database |

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the Go binary |
| `make run` | Run locally (needs `GITHUB_TOKEN` env var) |
| `make docker-build` | Build Docker image |
| `make cluster` | Create kind cluster |
| `make deploy` | Build + load + Helm install to kind |
| `make all` | Full setup (cluster + deploy) |
| `make template` | Render Helm templates (dry-run) |
| `make lint` | Lint the Helm chart |
| `make status` | Check pod and service status |
| `make logs` | Stream pod logs |
| `make uninstall` | Uninstall the Helm release |
| `make clean` | Delete kind cluster + cleanup |
# repoistory-tool
