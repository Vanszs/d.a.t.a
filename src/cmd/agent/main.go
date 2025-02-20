package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/carv-protocol/d.a.t.a/src/web"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
	"github.com/carv-protocol/d.a.t.a/src/internal/social"
	"github.com/carv-protocol/d.a.t.a/src/internal/tasks"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
	"github.com/carv-protocol/d.a.t.a/src/internal/tools"
	"github.com/carv-protocol/d.a.t.a/src/pkg/carv"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database/adapters"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	customTools "github.com/carv-protocol/d.a.t.a/src/tools"
	data "github.com/carv-protocol/d.a.t.a/src/tools/d.a.t.a"
	"github.com/carv-protocol/d.a.t.a/src/tools/wallet"

	"github.com/google/uuid"
)

// Config validation errors
var (
	ErrInvalidLLMConfig = errors.New("invalid LLM configuration")
	ErrInvalidDBConfig  = errors.New("invalid database configuration")
	FlagConfig          string
)

func init() {
	flag.StringVar(&FlagConfig, "conf", "./src/config", "config path, eg: -conf config.yaml")
}

func main() {
	flag.Parse()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	config, err := loadConfig(FlagConfig)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize components
	agent, err := initializeAgent(ctx, config)
	if err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}

	// Start the agent
	if err := agent.Start(); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	web.Start(config.Web.Port)

	// Wait for shutdown signal
	<-handleShutdown(ctx, agent, config.Settings.ShutdownTimeout)
}

func initializeAgent(ctx context.Context, config *Config) (*core.Agent, error) {
	// Setup database
	store := adapters.NewPostgresStore(config.Database.Path)
	if err := store.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Initialize components
	llmClient := llm.NewClient((*llm.LLMConfig)(&config.LLMConfig))
	carvClient := carv.NewClient(config.Data.CarvConfig.APIKey, config.Data.CarvConfig.BaseURL)
	memoryManager := memory.NewManager(store)
	tokenManager := token.NewTokenManager(carvClient, &core.TokenInfo{
		Network:      config.Token.Network,
		Ticker:       config.Token.Ticker,
		ContractAddr: config.Token.ContractAddr,
	})
	stakeholderManager := token.NewStakeholderManager(memoryManager)

	// Load character
	character, err := characters.LoadFromFile(config.Character.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to load character: %w", err)
	}

	// Initialize tools
	toolsManager := initializeTools(config)

	// Initialize action manager and register actions
	actionManager := actions.NewManager()
	// dbProvider := actions.NewDatabaseProvider(
	// 	config.Data.CarvConfig.BaseURL,
	// 	config.Data.CarvConfig.APIKey,
	// 	config.Token.Network,
	// 	llmClient,
	// )
	// fetchAction := actions.NewFetchTransactionAction(dbProvider)
	// if err := actionManager.Register(fetchAction); err != nil {
	// 	return nil, fmt.Errorf("failed to register fetch transaction action: %w", err)
	// }

	// Create agent
	agentConfig := core.AgentConfig{
		ID:           uuid.New(),
		Character:    character,
		LLMClient:    llmClient,
		Model:        config.LLMConfig.Model,
		Stakeholders: stakeholderManager,
		ToolsManager: toolsManager,
		SocialClient: social.NewSocialClient(
			&config.Social.TwitterConfig,
			&config.Social.DiscordConfig,
			&config.Social.TelegramConfig,
		),
		TaskManager:   tasks.NewManager(tasks.NewTaskStore(store)),
		ActionManager: actionManager,
		TokenManager:  tokenManager,
	}

	agent, err := core.NewAgent(agentConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
}

func initializeTools(config *Config) *tools.Manager {

	walletTool, err := wallet.NewWalletTool(&config.Wallet)
	if err != nil {
		log.Fatalf("Failed to create wallet tool: %v", err)
	}

	toolsManager := tools.NewManager()
	toolsManager.Register(&customTools.TwitterTool{})
	toolsManager.Register(&data.CARVDataTool{})
	toolsManager.Register(walletTool)
	return toolsManager
}

func handleShutdown(ctx context.Context, agent *core.Agent, timeoutSeconds int) chan struct{} {
	done := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, initiating graceful shutdown...")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
		defer cancel()

		web.Stop()

		if err := agent.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}

		close(done)
		log.Println("Shutdown completed")
	}()

	return done
}
