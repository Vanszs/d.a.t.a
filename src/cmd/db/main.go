package main

import (
	"context"
	"encoding/json"
	"github.com/carv-protocol/d.a.t.a/src/internal/core"
	"github.com/carv-protocol/d.a.t.a/src/internal/memory"
	"github.com/carv-protocol/d.a.t.a/src/internal/token"
	"github.com/carv-protocol/d.a.t.a/src/pkg/carv"
	"github.com/carv-protocol/d.a.t.a/src/pkg/database/adapters"
	"github.com/carv-protocol/d.a.t.a/src/pkg/logger"
)

func main() {
	logger.Init()

	ctx := context.Background()

	// Initialize database connection
	store := adapters.NewSQLiteStore("./data/agent.db") // You might want to make this configurable
	if err := store.Connect(ctx); err != nil {
		logger.GetLogger().Fatalf("Failed to connect to database: %v", err)
	}
	defer store.Close()

	// Initialize memory manager
	memoryManager := memory.NewManager(store)

	// Fetch all memories
	memories, err := memoryManager.ListMemories(ctx, memory.MemoryFilter{})
	if err != nil {
		logger.GetLogger().Fatalf("Failed to fetch memories: %v", err)
	}

	logger.GetLogger().Infoln(memories)

	var stakeholder *core.Stakeholder
	for _, memory := range memories {
		// logger.GetLogger().Infoln(memory.Content)
		carvClient := carv.NewClient("89fa0b9c-4b1e-42a9-b5f3-d4c47f69b4f6", "https://interface.carv.io/ai-agent-backend")
		tokenManager := token.NewTokenManager(carvClient, &core.TokenInfo{
			Network:      "base",
			Ticker:       "carv",
			ContractAddr: "0xc08cd26474722ce93f4d0c34d16201461c10aa8c",
		})
		err = json.Unmarshal(memory.Content, &stakeholder)
		if err != nil {
			logger.GetLogger().Fatalf("Failed to unmarshal memory: %v", err)
		}
		balance, _ := tokenManager.FetchNativeTokenBalance(ctx, stakeholder.ID, stakeholder.Platform)
		if balance != nil {
			logger.GetLogger().Infoln(stakeholder.ID, stakeholder.Platform, len(stakeholder.HistoricalMsgs), balance.Balance)
		}

		// logger.GetLogger().Infoln(stakeholder.ID, stakeholder.Platform, stakeholder.TokenBalance.Balance)
	}

	// Print summary
	logger.GetLogger().Infof("Total memories: %d", len(memories))
}
