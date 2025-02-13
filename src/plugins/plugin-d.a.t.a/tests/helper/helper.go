package helper

import (
	"os"
	"testing"

	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/plugins/d.a.t.a/actions"
	"github.com/carv-protocol/d.a.t.a/src/plugins/d.a.t.a/providers"
)

const defaultModel = "gpt-4"

// SetupTestProvider creates a new database provider for testing
func SetupTestProvider(t *testing.T) actions.DatabaseProvider {
	t.Helper()
	apiURL := os.Getenv("DATA_API_KEY")
	if apiURL == "" {
		t.Skip("DATA_API_KEY not set")
	}
	authToken := os.Getenv("DATA_AUTH_TOKEN")
	if authToken == "" {
		t.Skip("DATA_AUTH_TOKEN not set")
	}
	return providers.NewDatabaseProvider(apiURL, authToken, "ethereum-mainnet", nil, defaultModel)
}

// SetupTestProviderWithLLM creates a new database provider with LLM client for testing
func SetupTestProviderWithLLM(t *testing.T, llmClient llm.Client) actions.DatabaseProvider {
	t.Helper()
	apiURL := os.Getenv("DATA_API_KEY")
	if apiURL == "" {
		t.Skip("DATA_API_KEY not set")
	}
	authToken := os.Getenv("DATA_AUTH_TOKEN")
	if authToken == "" {
		t.Skip("DATA_AUTH_TOKEN not set")
	}
	return providers.NewDatabaseProvider(apiURL, authToken, "ethereum-mainnet", llmClient, defaultModel)
}

// StrPtr returns a pointer to the given string
func StrPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to the given integer
func IntPtr(i int) *int {
	return &i
}
