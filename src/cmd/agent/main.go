package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	character "github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/data"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"
)

func main() {
	// Parse command line flags
	characterPath := flag.String("character", "", "Path to character config file")
	llmEndpoint := flag.String("llm-endpoint", "", "LLM service endpoint")
	dataEndpoint := flag.String("data-endpoint", "", "Data service endpoint")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize components
	log := logger.GetLogger()

	// Load character
	character, err := character.LoadFromFile(*characterPath)
	if err != nil {
		log.Fatalf("Failed to load character: %v", err)
	}

	// Initialize services
	llmClient := llm.NewClient(*llmEndpoint)
	dataManager := data.NewManager(*dataEndpoint)

	// Create agent
	agent, err := core.NewAgent(core.AgentConfig{
		Character:   character,
		LLMClient:   llmClient,
		DataManager: dataManager,
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Start autonomous loop
	agent.StartAutonomousLoop(ctx)
	log.Info("Agent started autonomous operations")

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Info("Shutdown signal received")
		cancel()
	case <-ctx.Done():
		log.Info("Context cancelled")
	}

	log.Info("Shutting down gracefully...")
	if err := agent.Shutdown(ctx); err != nil {
		log.Errorf("Error during shutdown: %v", err)
	}
}
