package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"plugin"
	"syscall"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/data"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database/adapters"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm/openai"
	"github.com/google/uuid"

	"github.com/spf13/viper"
)

type Config struct {
	Character struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"character"`

	Database struct {
		Type string `mapstructure:"type"`
		Path string `mapstructure:"path"`
	} `mapstructure:"database"`

	LLM struct {
		Provider string `mapstructure:"provider"`
		APIKey   string `mapstructure:"api_key"`
		BaseURL  string `mapstructure:"base_url"`
	} `mapstructure:"llm"`

	Data struct {
		CarvID struct {
			URL    string `mapstructure:"url"`
			APIKey string `mapstructure:"api_key"`
		} `mapstructure:"carvid"`
	} `mapstructure:"data"`
}

func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Load .env file
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	// Environment variables take precedence
	viper.AutomaticEnv()
	viper.SetEnvPrefix("DATA")

	// Default values
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "./data/data.db")
	viper.SetDefault("llm.provider", "openai")
	viper.SetDefault("llm.base_url", "https://api.openai.com/v1")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	ctx := context.Background()

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup database
	store := adapters.NewSQLiteStore(config.Database.Path)
	if err := store.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Setup LLM client
	llmClient := openai.NewClient(config.LLM.APIKey)

	dataManager := data.NewManager(llmClient)
	memoryManager := memory.NewManager(store)
	tokenManager := token.NewTokenManager(dataManager)

	stakeholderManager := token.NewStakeholderManager(memoryManager, tokenManager)

	// Load character
	character, err := characters.LoadFromFile(config.Character.Path)
	if err != nil {
		log.Fatalf("Failed to load character: %v", err)
	}

	analysisPlugin := &plugin.Plugin{
		// Plugin implements new capabilities
	}

	// Initialize system
	system := core.NewAgentSystem(core.AgentSystemConfig{
		TokenManager:       tokenManager,
		StakeholderManager: stakeholderManager,
	})

	// Add agents
	system.AddAgent(core.AgentConfig{
		ID: uuid.New(),

		Character: character,
		// Goals: []Goal{
		// 	{ID: "profit", Weight: 0.7},
		// 	{ID: "risk_management", Weight: 0.3},
		// },
	})

	system.RegisterPlugin(analysisPlugin)

	// Start system
	if err := system.Start(); err != nil {
		log.Fatal(err)
	}

	// Wait for shutdown signal
	<-handleShutdown(ctx, system)
}

func handleShutdown(ctx context.Context, system *core.AgentSystem) chan struct{} {
	done := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		system.Shutdown(ctx)
		close(done)
	}()

	return done
}
