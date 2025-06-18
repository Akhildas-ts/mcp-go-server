package main

import (
	"log"
	"net/http"
	"os"
	"mcpserver/internal/config"
	"mcpserver/internal/handler"
	"mcpserver/internal/service"
	"mcpserver/internal/storage"
	"mcpserver/internal/middleware"

	"github.com/joho/godotenv"
)

type Server struct {
	Config   *config.Config
	Services *Services
	Handlers *Handlers
}

type Services struct {
	VectorSearch *service.VectorSearchService
	RepoIndexer  *service.RepoIndexerService
	MCPServer    *service.MCPServerService
}

type Handlers struct {
	Health       *handler.HealthHandler
	VectorSearch *handler.VectorSearchHandler
	RepoIndexer  *handler.RepoIndexerHandler
	MCP          *handler.MCPHandler
}
type OAuthConfig struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
}

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Printf("Warning: Error loading .env file: %v", err)
    }

    handler.InitOAuthConfig() // Initialize OAuth config after env is loaded

    // Create OAuth configuration
    oauthConfig := &OAuthConfig{
        ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
        ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
        RedirectURL:  os.Getenv("GITHUB_OAUTH_REDIRECT_URL"),
    }

    // Validate OAuth configuration
    if oauthConfig.ClientID == "" || oauthConfig.ClientSecret == "" || oauthConfig.RedirectURL == "" {
        log.Fatal("Missing required OAuth environment variables")
    }

    // Initialize server
    server, err := initializeServer()
    if err != nil {
        log.Fatalf("Failed to initialize server: %v", err)
    }

    // Setup routes (pass OAuth config to routes)
    router := setupRoutes(server.Handlers)

    // Start server
    log.Printf("MCP Server starting on port %s...", server.Config.Port)
    if err := http.ListenAndServe(":"+server.Config.Port, router); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}
func initializeServer() (*Server, error) {
	// Load configuration
	cfg := config.Load()

	// Initialize storage layer
	pineconeStore, err := storage.NewPineconeStore(
		cfg.PineconeAPIKey,
		cfg.PineconeEnvironment,
		cfg.PineconeIndexName,
		cfg.PineconeHost,
	)
	if err != nil {
		return nil, err
	}

	openaiClient := storage.NewOpenAIClient(cfg.OpenAIAPIKey)

	// Initialize services
	services := &Services{
		VectorSearch: service.NewVectorSearchService(pineconeStore, openaiClient),
		RepoIndexer:  service.NewRepoIndexerService(pineconeStore, openaiClient),
		MCPServer:    service.NewMCPServerService(pineconeStore, openaiClient),
	}

	// Initialize handlers
	handlers := &Handlers{
		Health:       handler.NewHealthHandler(),
		VectorSearch: handler.NewVectorSearchHandler(services.VectorSearch),
		RepoIndexer:  handler.NewRepoIndexerHandler(services.RepoIndexer),
		MCP:          handler.NewMCPHandler(services.MCPServer),
	}

	return &Server{
		Config:   cfg,
		Services: services,
		Handlers: handlers,
	}, nil
}

func setupRoutes(h *Handlers) *http.ServeMux {
	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("/health", h.Health.HandleHealthCheck)
	mux.HandleFunc("/auth/login", handler.HandleGitHubLogin)
	mux.HandleFunc("/auth/callback", handler.HandleGitHubCallback)

	// Protected endpoints
	mux.Handle("/mcp-info", middleware.AuthMiddleware(http.HandlerFunc(h.MCP.HandleMCPRegistration)))
	mux.Handle("/cursor", middleware.AuthMiddleware(http.HandlerFunc(h.MCP.HandleCursorConnection)))
	mux.Handle("/chat", middleware.AuthMiddleware(http.HandlerFunc(h.MCP.HandleChat)))
	mux.Handle("/github-config", middleware.AuthMiddleware(http.HandlerFunc(h.MCP.HandleGitHubConfig)))
	mux.Handle("/vector-search", middleware.AuthMiddleware(http.HandlerFunc(h.VectorSearch.HandleVectorSearch)))
	mux.Handle("/index-repository", middleware.AuthMiddleware(http.HandlerFunc(h.RepoIndexer.HandleRepositoryIndexing)))

	return mux
}