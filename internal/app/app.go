package app

import (
	"log"

	"goredis/internal/command"
	"goredis/internal/config"
	"goredis/internal/db"
	"goredis/internal/persistence"
	"goredis/internal/protocol"
	"goredis/internal/server"
)

func Run() error {
	return RunWithConfig(config.Default())
}

func RunWithConfig(cfg config.Config) error {
	// DB with optional LRU eviction (MaxMemoryBytes == 0 means unlimited).
	engine := db.NewWithOptions(cfg.MaxMemoryBytes)
	aof := persistence.NewAOF(cfg.AppendOnlyPath)

	registry := command.NewRegistry()
	command.RegisterStringCommands(registry, engine, aof)
	command.RegisterGenericCommands(registry, engine, aof)
	dispatcher := command.NewDispatcher(registry)

	// Restore data from AOF before accepting connections.
	if err := aof.Replay(dispatcher); err != nil {
		log.Printf("[WARN] AOF replay error: %v", err)
	}

	parser := protocol.NewRESPParser()
	srv := server.New(cfg, parser, dispatcher)
	return srv.Start()
}
