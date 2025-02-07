package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"plugin"
	"strings"
	"syscall"

	"github.com/carv-protocol/d.a.t.a/src/characters"
	"github.com/carv-protocol/d.a.t.a/src/internal/actions"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
	"github.com/carv-protocol/d.a.t.a/src/internal/tasks"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
	"github.com/carv-protocol/d.a.t.a/src/internal/tools"
	"github.com/carv-protocol/d.a.t.a/src/pkg/carv"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database/adapters"
	"github.com/carv-protocol/d.a.t.a/src/pkg/llm"
	customTools "github.com/carv-protocol/d.a.t.a/src/tools"

	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./src/config")

	// Load .env file
	// TODO: override config with .env
	// viper.SetConfigName(".env")
	// viper.SetConfigType("env")
	// viper.AddConfigPath("./src")

	// Default values
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "./data/data.db")
	viper.SetDefault("llm_config.provider", "openai")
	viper.SetDefault("llm_config.base_url", "https://api.openai.com/v1")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Load .env file
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath("./src")
	if err := viper.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("error reading .env: %w", err)
	}

	viper.Set("llm_config.api_key", viper.Get("LLM_API_KEY"))
	viper.Set("social.twitter.bearer_token", viper.Get("TWITTER_BEARER_TOKEN"))
	viper.Set("social.twitter.api_key", viper.Get("TWITTER_API_KEY"))
	viper.Set("social.twitter.api_key_secret", viper.Get("TWITTER_API_KEY_SECRET"))
	viper.Set("social.twitter.access_token", viper.Get("TWITTER_ACCESS_TOKEN"))
	viper.Set("social.twitter.token_secret", viper.Get("TWITTER_TOKEN_SECRET"))
	viper.Set("social.discord.api_token", viper.Get("DISCORD_API_TOKEN"))
	viper.Set("social.telegram.bot_token", viper.Get("TELEGRAM_BOT_TOKEN"))
	if channelID := viper.GetString("TELEGRAM_CHANNEL_ID"); channelID != "" {
		// Remove any potential "${}" wrapper if present
		channelID = strings.Trim(channelID, "${}")
		viper.Set("social.telegram.channel_id", channelID)
	}
	viper.Set("data.carv.base_url", viper.Get("CARV_DATA_BASE_URL"))
	viper.Set("data.carv.api_key", viper.Get("CARV_DATA_API_KEY"))

	// Environment variables take precedence
	viper.AutomaticEnv()

	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, err
	}

	return &conf, nil
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
	llmClient := llm.NewClient((*llm.LLMConfig)(&config.LLMConfig))

	carvClient := carv.NewClient(config.Data.CarvConfig.APIKey, config.Data.CarvConfig.BaseURL)

	// dataManager := data.NewManager(llmClient, carvClient)
	memoryManager := memory.NewManager(store)
	tokenManager := token.NewTokenManager(carvClient, &token.TokenInfo{
		Network: config.Token.Network,
		Ticker:  config.Token.Ticker,
	})

	stakeholderManager := token.NewStakeholderManager(memoryManager)

	// Load character
	character, err := characters.LoadFromFile(config.Character.Path)
	if err != nil {
		log.Fatalf("Failed to load character: %v", err)
	}

	analysisPlugin := &plugin.Plugin{
		// Plugin implements new capabilities
	}

	taskStore := tasks.NewTaskStore(store)
	toolsManager := tools.NewManager()
	toolsManager.Register(&customTools.TwitterTool{})
	toolsManager.Register(&customTools.CARVDataTool{})
	toolsManager.Register(&customTools.WalletTool{})

	actionManager := actions.NewManager()

	// Initialize system
	agent := core.NewAgent(core.AgentConfig{
		ID:        uuid.New(),
		Character: character,
		LLMClient: llmClient,
		// DataManager:   dataManager,
		MemoryManager: memoryManager,
		Stakeholders:  stakeholderManager,
		ToolsManager:  toolsManager,
		SocialClient:  core.NewSocialClient(nil, &config.Social.DiscordConfig, &config.Social.TelegramConfig),
		TaskManager:   tasks.NewManager(taskStore),
		ActionManager: actionManager,
		TokenManager:  tokenManager,
	})

	agent.RegisterPlugin(analysisPlugin)

	if err := agent.Start(); err != nil {
		log.Fatal(err)
	}

	// Wait for shutdown signal
	<-handleShutdown(ctx, agent)
}

func handleShutdown(ctx context.Context, agent *core.Agent) chan struct{} {
	done := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		agent.Shutdown(ctx)
		close(done)
	}()

	return done
}
