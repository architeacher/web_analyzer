```bash
 _       __     __       ___                __
| |     / /__  / /_     /   |  ____  ____ _/ /_  ______  ___  _____
| | /| / / _ \/ __ \   / /| | / __ \/ __ `/ / / / /_  / / _ \/ ___/
| |/ |/ /  __/ /_/ /  / ___ |/ / / / /_/ / / /_/ / / /_/  __/ /
|__/|__/\___/_.___/  /_/  |_/_/ /_/\__,_/_/\__, / /___/\___/_/
                                          /____/
```

A comprehensive web application that analyzes web pages and provides detailed insights about HTML structure, links, and forms.

## Features

- **Web Page Analysis**: HTML version detection, title extraction, heading analysis, and form detection
- **Link Analysis**: Internal/external link identification with accessibility checking
- **Real-time Updates**: Server-Sent Events for live progress tracking
- **Secure API**: [PASETO](https://paseto.io/) token authentication with comprehensive security headers
- **Multiple API Versioning**: URL path, header, and content type versioning strategies

For complete feature documentation, see [Features Documentation](docs/features.md).

## Documentation

### Architecture & Design
- **[Architecture Decisions](docs/architecture-decisions.md)**: Comprehensive ADRs documenting all major architectural choices and their rationale
- **[Features Documentation](docs/features.md)**: Detailed documentation of all implemented features, APIs, and capabilities

### API Documentation
- **[OpenAPI Specification](docs/openapi-spec/web-analyzer-api.yaml)**: Complete OpenAPI 3.0.3 specification
- **[Generated Documentation](https://docs.web-analyzer.dev)**: Interactive API documentation (available after running `make init`)

## Architecture

This project implements a **code-first API design** approach with comprehensive OpenAPI specification and generated server code.

### Project Structure

```
web-analyzer/
├── cmd/web-analyzer/             # Application entry point
├── internal/                     # Private application packages
│   ├── adapters/                # Interface implementations
│   │   └── middleware/          # HTTP middleware components
│   ├── config/                  # Configuration management
│   ├── domain/                  # Core business logic and entities
│   ├── handlers/                # Generated HTTP handlers
│   ├── infrastructure/          # External dependencies (storage, logging)
│   ├── ports/                   # Service interfaces
│   ├── runtime/                 # Application bootstrap and DI
│   ├── service/                 # Business services
│   ├── shared/decorator/        # Cross-cutting concerns decorators
│   ├── tools/                   # Code generation tools
│   └── usecases/                # Application use cases (commands/queries)
│       ├── command/             # Command handlers
│       ├── query/               # Query handlers
│       └── sse/                 # Server-sent events
├── docs/                         # Documentation
│   ├── architecture-decisions.md # Architectural decision records
│   ├── features.md              # Feature documentation
│   └── openapi-spec/            # OpenAPI 3.0.3 specification
│       ├── web-analyzer-api.yaml # Main API specification
│       ├── schemas/             # Schema definitions and examples
│       └── public/              # Generated API documentation
├── deployments/docker/           # Docker deployment configuration
├── build/                        # Build system and configuration
│   ├── mk/                      # Make-based build system
│   └── oapi/                    # OpenAPI code generation config
├── assets/                       # Project assets and branding
├── compose.yaml                  # Docker Compose configuration
├── CHANGELOG.md                  # Project changelog
└── go.mod                        # Go module definition
```

## Technology Stack

### Backend
- **Language**: Go 1.25
- **Code Generation**: oapi-codegen for OpenAPI-to-Go conversion
- **Authentication**: PASETO tokens with enhanced security validation
- **API Specification**: OpenAPI 3.0.3 with comprehensive examples
- **Build System**: Make with modular build configuration

### API Design
- **Specification**: OpenAPI 3.0.3 with detailed schemas and examples
- **Versioning**: Multiple strategies (URL path `/v1/`, headers, content type)
- **Real-time**: Server-Sent Events for analysis progress
- **Security**: Complete security headers and PASETO authentication
- **Documentation**: Auto-generated from OpenAPI specification

### Infrastructure
- **Containerization**: Docker with multi-stage builds
- **Reverse Proxy**: Traefik with automatic SSL/TLS
- **Local Development**: SSL certificate generation with mkcert
- **Documentation**: Redocly CLI for API bundling and validation

## Quick Start

### Prerequisites
- Go 1.25+
- Docker and Docker Compose
- mkcert (for SSL certificates)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/architeacher/svc-web-analyzer.git
   cd svc-web-analyzer
   ```

2. **Initialize and start the development environment**
   ```bash
   make init start
   ```
   This will:
   - Copy `.envrc.dist` to `.envrc` (edit as needed)
   - Add local domains to `/etc/hosts`
   - Generate SSL certificates with mkcert
   - Download Go dependencies with `go mod vendor`
   - Generate API code from OpenAPI specification
   - Start all services with Docker Compose

3. **Access the application**
   - **API**: https://api.web-analyzer.dev/v1/ (TBD: API documentation)
   - **Documentation**: https://docs.web-analyzer.dev
   - **Traefik Dashboard**: https://traefik.web-analyzer.dev (admin/admin)

### Development Commands

```bash
# Initialize project (hosts, SSL certs, API generation)
make init

# Start development services
make start

# Stop and remove development services
make destroy

# Generate SSL certificates for local development
make certify

# Generate API code from OpenAPI specification
make generate-api

# Run all tests
make test

# Update local hosts
make set-hosts

# View all available targets
make help

# List all targets
make list
```

## API Documentation

The API is fully documented using OpenAPI 3.0.3 specification with comprehensive examples.

- **API Specification**: [docs/openapi-spec/web-analyzer-api.yaml](docs/openapi-spec/web-analyzer-api.yaml)
- **Generated Bundle**: [docs/openapi-spec/public/web-analyzer-swagger-v1.json](docs/openapi-spec/public/web-analyzer-swagger-v1.json)
- **Documentation**: https://docs.web-analyzer.dev (after running `make init`)
- **API Endpoint**: https://api.web-analyzer.dev/v1/

### Core Endpoints

- `POST /v1/analyze` - Submit URL for analysis
- `GET /v1/analysis/{analysisId}` - Get analysis result
- `GET /v1/analysis/{analysisId}/events` - Real-time progress (SSE)
- `GET /v1/health` - Health check endpoint

### API Examples

#### Health Check

```bash
curl -s https://api.web-analyzer.dev/v1/health | jq
```

**Output:**
```json
{
  "dependencies": {
    "external_services": "healthy",
    "storage": "healthy"
  },
  "status": "healthy",
  "timestamp": "2025-09-22T16:45:11.572777759Z",
  "version": "1.0.0"
}
```

#### Submit URL for Analysis

**Testing GitHub Login Page with PASETO v4:**
```bash
curl -v https://api.web-analyzer.dev/v1/analyze \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer v4.public.eyJhdWQiOiJ3ZWItYW5hbHl6ZXItYXBpIiwiZXhwIjoiMjA2My0wOS0xOFQwMjoyMDoxNyswMjowMCIsImlhdCI6IjIwMjUtMDktMjdUMDI6MjA6MTcrMDI6MDAiLCJpc3MiOiJ3ZWItYW5hbHl6ZXItc2VydmljZSIsImp0aSI6InByb3Blci1wYXNldG8tdjQtdG9rZW4iLCJuYmYiOiIyMDI1LTA5LTI3VDAyOjIwOjE3KzAyOjAwIiwic2NvcGVzIjpbImFuYWx5emUiLCJyZWFkIl0sInN1YiI6InRlc3QtdXNlciJ9MVH2eMTu9jMw6ZUIB538m-4gUoonWUbkHPDReqzD_2lojhtO2d1l3FXc6RCOozfW3fIdbU9y9SWAzBBamKydAQ" \
  -d '{"url": "https://github.com/login"}' | jq
```

**Successful Response:**
```json
{
  "analysis_id": "f1c67ce5-0236-433f-99c1-be1e465af446",
  "created_at": "2025-09-22T17:39:32.231180065Z",
  "status": "requested",
  "url": "https://github.com/login"
}
```

#### Get Analysis Result

```bash
curl -s https://api.web-analyzer.dev/v1/analysis/f1c67ce5-0236-433f-99c1-be1e465af446 \
  -H "Authorization: Bearer v4.public.eyJhdWQiOiJ3ZWItYW5hbHl6ZXItYXBpIiwiZXhwIjoiMjA2My0wOS0xOFQwMjoyMDoxNyswMjowMCIsImlhdCI6IjIwMjUtMDktMjdUMDI6MjA6MTcrMDI6MDAiLCJpc3MiOiJ3ZWItYW5hbHl6ZXItc2VydmljZSIsImp0aSI6InByb3Blci1wYXNldG8tdjQtdG9rZW4iLCJuYmYiOiIyMDI1LTA5LTI3VDAyOjIwOjE3KzAyOjAwIiwic2NvcGVzIjpbImFuYWx5emUiLCJyZWFkIl0sInN1YiI6InRlc3QtdXNlciJ9MVH2eMTu9jMw6ZUIB538m-4gUoonWUbkHPDReqzD_2lojhtO2d1l3FXc6RCOozfW3fIdbU9y9SWAzBBamKydAQ" | jq
```

**Current Response (Analysis In Progress):**
```json
{
  "analysis_id": "f1c67ce5-0236-433f-99c1-be1e465af446",
  "status": "in_progress"
}
```

> **Note:** The analysis is processing in the background. The endpoint successfully authenticates with PASETO v4 tokens and queues the analysis. Processing time depends on the complexity of the target website.

#### Real-time Progress (Server-Sent Events)

```bash
curl -s https://api.web-analyzer.dev/v1/analysis/f1c67ce5-0236-433f-99c1-be1e465af446/events \
  -H "Authorization: Bearer v4.public.eyJhdWQiOiJ3ZWItYW5hbHl6ZXItYXBpIiwiZXhwIjoiMjA2My0wOS0xOFQwMjoyMDoxNyswMjowMCIsImlhdCI6IjIwMjUtMDktMjdUMDI6MjA6MTcrMDI6MDAiLCJpc3MiOiJ3ZWItYW5hbHl6ZXItc2VydmljZSIsImp0aSI6InByb3Blci1wYXNldG8tdjQtdG9rZW4iLCJuYmYiOiIyMDI1LTA5LTI3VDAyOjIwOjE3KzAyOjAwIiwic2NvcGVzIjpbImFuYWx5emUiLCJyZWFkIl0sInN1YiI6InRlc3QtdXNlciJ9MVH2eMTu9jMw6ZUIB538m-4gUoonWUbkHPDReqzD_2lojhtO2d1l3FXc6RCOozfW3fIdbU9y9SWAzBBamKydAQ" \
  -H "Accept: text/event-stream"
```

**Expected SSE Stream:**
```
data: {"stage": "queued", "progress": 0, "message": "Analysis queued"}

data: {"stage": "fetching", "progress": 20, "message": "Fetching web page"}

data: {"stage": "parsing_html", "progress": 50, "message": "Parsing HTML structure"}

data: {"stage": "analyzing_links", "progress": 75, "message": "Analyzing links"}

data: {"stage": "completed", "progress": 100, "message": "Analysis completed"}
```

### Authentication

> **Note:** Authentication is required for all endpoints except `/v1/health`. The API supports both PASETO v4 and custom token formats.

**Authorization Header (Bearer Token)**
```bash
-H "Authorization: Bearer v4.public.{base64url-payload}{base64url-signature}"
```
```

**Working PASETO v4 Example:**
The following token is valid for 38 years and includes `analyze` and `read` scopes:
```
v4.public.eyJhdWQiOiJ3ZWItYW5hbHl6ZXItYXBpIiwiZXhwIjoiMjA2My0wOS0xOFQwMjoyMDoxNyswMjowMCIsImlhdCI6IjIwMjUtMDktMjdUMDI6MjA6MTcrMDI6MDAiLCJpc3MiOiJ3ZWItYW5hbHl6ZXItc2VydmljZSIsImp0aSI6InByb3Blci1wYXNldG8tdjQtdG9rZW4iLCJuYmYiOiIyMDI1LTA5LTI3VDAyOjIwOjE3KzAyOjAwIiwic2NvcGVzIjpbImFuYWx5emUiLCJyZWFkIl0sInN1YiI6InRlc3QtdXNlciJ9MVH2eMTu9jMw6ZUIB538m-4gUoonWUbkHPDReqzD_2lojhtO2d1l3FXc6RCOozfW3fIdbU9y9SWAzBBamKydAQ
```

**About PASETO:**
[PASETO (Platform-Agnostic Security Tokens)](https://paseto.io/) provides secure, authenticated tokens with Ed25519 signatures for v4 public tokens. The implementation supports both standard PASETO v4 tokens and backward-compatible custom formats.

## Configuration

The application is configured using environment variables. See `.envrc.dist` for available configuration options.

### Local Development
The project includes a complete local development setup:
- **SSL Certificates**: Automatic generation with mkcert
- **Local Domains**: `*.web-analyzer.dev` configured in `/etc/hosts`
- **Reverse Proxy**: Traefik configuration for service routing
- **Docker Compose**: Multi-service development environment

## Code Generation

The project uses a code-first approach with OpenAPI specification:

### API Generation Process
1. **Define**: Write OpenAPI 3.0.3 specification in `docs/openapi-spec/`
2. **Bundle**: Use Redocly CLI to create a unified specification
3. **Generate**: Use oapi-codegen to create Go server interfaces
4. **Implement**: Write business logic implementing the generated interfaces

### Generated Code
- **HTTP Server**: Generated interfaces and types in `internal/httpserver/`
- **API Bundle**: Single JSON specification for documentation
- **Examples**: Comprehensive request/response examples

## Security Features

- **PASETO Authentication**: Enhanced security tokens with issuer validation
- **Security Headers**: Complete set of standard security headers
- **CORS Configuration**: Configurable cross-origin resource sharing
- **Input Validation**: Schema-based validation from OpenAPI specification

## Development Tools

- **Make Targets**: Comprehensive build automation
- **Docker Integration**: Multi-stage builds and development containers
- **SSL/TLS**: Local development with valid certificates
- **API Documentation**: Auto-generated from OpenAPI specification

## Running the Application

### Prerequisites

Before running the application, ensure you have the following installed:

- **Go 1.25+**: Required for building and running the application
- **Docker & Docker Compose**: For containerized development environment
- **mkcert**: For generating local SSL certificates
- **Make**: For build automation (usually pre-installed on macOS/Linux)

### Installation Steps

1. **Install Go 1.25+**
   ```bash
   # macOS with Homebrew
   brew install go

   # Or download from https://golang.org/dl/
   ```

2. **Install Docker Desktop**
   - Download from [https://docker.com/products/docker-desktop](https://docker.com/products/docker-desktop)
   - Ensure Docker Compose is included (it comes with Docker Desktop)

3. **Install mkcert**
   ```bash
   # macOS with Homebrew
   brew install mkcert

   # Ubuntu/Debian
   sudo apt install libnss3-tools
   curl -JLO "https://dl.filippo.io/mkcert/latest?for=linux/amd64"
   chmod +x mkcert-v*-linux-amd64
   sudo cp mkcert-v*-linux-amd64 /usr/local/bin/mkcert
   ```

### Quick Start

1. **Clone and setup**
   ```bash
   git clone https://github.com/architeacher/svc-web-analyzer.git
   cd svc-web-analyzer
   make init start
   ```

2. **Verify installation**
   ```bash
   # Check health endpoint
   curl -s https://api.web-analyzer.dev/v1/health | jq

   # Should return status: "healthy"
   ```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/domain/
```

### Building the Application

```bash
# Build for current platform
go build -o bin/web-analyzer ./cmd/web-analyzer

# Build for Linux (useful for containers)
GOOS=linux GOARCH=amd64 go build -o bin/web-analyzer-linux ./cmd/web-analyzer

# Build with optimizations for production
go build -ldflags="-w -s" -o bin/web-analyzer ./cmd/web-analyzer
```

### Environment Variables

The application uses environment variables for configuration. Copy `.envrc.dist` to `.envrc` and modify as needed:

```bash
cp .envrc.dist .envrc
# Edit .envrc with your preferred editor
```

Key environment variables:
- `PORT`: Server port (default: 8080)
- `ENVIRONMENT`: Application environment (development, staging, production)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `KEYDB_HOST`: KeyDB server host
- `KEYDB_PORT`: KeyDB server port

### Manual Setup (without Docker)

If you prefer to run without Docker:

1. **Install KeyDB**
   ```bash
   # macOS with Homebrew
   brew install keydb
   keydb-server

   # Ubuntu/Debian
   curl -s --compressed "https://download.keydb.dev/keydb-ppa/KEY.gpg" | sudo apt-key add -
   sudo curl -s --compressed -o /etc/apt/sources.list.d/keydb.list https://download.keydb.dev/keydb-ppa/bionic.list
   sudo apt update
   sudo apt install keydb
   ```

2. **Generate API code**
   ```bash
   make generate-api
   ```

3. **Run the application**
   ```bash
   go run ./cmd/web-analyzer
   ```

### Troubleshooting

#### Common Issues

1. **Permission denied on `/etc/hosts`**
   ```bash
   # The init command requires sudo for hosts file modification
   sudo make init
   ```

2. **Port already in use**
   ```bash
   # Check what's using the cache
   lsof -i :8080  # API port
   lsof -i :80    # Traefik HTTP
   lsof -i :443   # Traefik HTTPS

   # Stop conflicting services or change cache in .envrc
   ```

3. **SSL certificate issues**
   ```bash
   # Regenerate certificates
   make certify

   # Ensure mkcert CA is installed
   mkcert -install
   ```

4. **Docker build fails**
   ```bash
   # Clean Docker cache
   docker system prune -a

   # Rebuild without cache
   docker-compose build --no-cache
   ```

#### Logs and Debugging

```bash
# View application logs
docker-compose logs -f web-analyzer

# View all service logs
docker-compose logs -f

# Check container status
docker-compose ps

# Access container shell
docker-compose exec web-analyzer sh
```

### Development Workflow

1. **Start development environment**
   ```bash
   make start
   ```

2. **Make code changes** - The application will automatically reload with Air

3. **Test your changes**
   ```bash
   # Run tests
   go test ./...

   # Test API endpoints
   curl -s https://api.web-analyzer.dev/v1/health
   ```

4. **Generate new API code** (if OpenAPI spec changed)
   ```bash
   make generate-api
   ```

5. **Stop environment when done**
   ```bash
   make destroy
   ```

This comprehensive setup ensures a smooth development experience with automatic reloading, SSL certificates, and proper service orchestration.

## TODOs

### Testing & Quality
- **Increase Test Coverage**: Expand unit and integration test coverage across all layers
  - Add comprehensive tests for domain entities and business logic
  - Implement integration tests for API endpoints
  - Add middleware and adapter layer testing
  - Target 80%+ code coverage across the codebase

### Performance & Scalability
- **Background Job Processing**: Implement queue-based analysis processing
  - Integrate message queue system (Redis Queue, RabbitMQ, or similar)
  - Move web page analysis to background workers
  - Implement job status tracking and progress updates
  - Add retry mechanisms for failed analysis jobs

- **Content Deduplication**: Optimize analysis efficiency with content hashing
  - Calculate SHA-256 hash of HTML content
  - Store hash-to-analysis mapping to avoid duplicate processing
  - Implement cache lookup before initiating new analysis
  - Return cached results for identical content

### Code Quality & Architecture
- **Refactoring Improvements**
  - Extract common patterns into reusable components
  - Simplify complex adapter implementations
  - Improve error handling consistency across layers
  - Optimize dependency injection and configuration management
  - Review and consolidate middleware implementations
  - Adding linting rules

### Observability & Maintainability
- **Enhanced Analysis Features**
  - Metrics and monitoring integration
  - Performance metrics collection
  - Security vulnerability detection

- **Operational Improvements**
  - CI/CD pipeline
  - K8s deployment

These improvements will enhance the application's reliability, performance, and maintainability while reducing operational overhead.
