package config

import (
	"os"
)

type Config struct {
	Port                string
	PineconeAPIKey      string
	PineconeEnvironment string
	PineconeIndexName   string
	PineconeHost        string
	OpenAIAPIKey        string
	MCPSecretToken      string
	GitHubClientID      string
	GitHubClientSecret  string
	GitHubOAuthRedirectURL string
}

func Load() *Config {
	return &Config{
		Port:                getEnv("PORT", "8081"),
		PineconeAPIKey:      os.Getenv("PINECONE_API_KEY"),
		PineconeEnvironment: os.Getenv("PINECONE_ENVIRONMENT"),
		PineconeIndexName:   os.Getenv("PINECONE_INDEX_NAME"),
		PineconeHost:        os.Getenv("PINECONE_HOST"),
		OpenAIAPIKey:        os.Getenv("OPENAI_API_KEY"),
		MCPSecretToken:      os.Getenv("MCP_SECRET_TOKEN"),
		GitHubClientID:      os.Getenv("GITHUB_CLIENT_ID"),
		GitHubClientSecret:  os.Getenv("GITHUB_CLIENT_SECRET"),
		GitHubOAuthRedirectURL: os.Getenv("GITHUB_OAUTH_REDIRECT_URL"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}