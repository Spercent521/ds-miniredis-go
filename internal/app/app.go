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

	// Replay should not append commands back into AOF, otherwise startup can loop forever.
	replayRegistry := command.NewRegistry()
	command.RegisterStringCommands(replayRegistry, engine, nil)
	command.RegisterGenericCommands(replayRegistry, engine, nil)
	command.RegisterListCommands(replayRegistry, engine, nil)
	replayDispatcher := command.NewDispatcher(replayRegistry)

	// runtime dispatcher：write AOF
	runtimeRegistry := command.NewRegistry()
	command.RegisterStringCommands(runtimeRegistry, engine, aof)
	command.RegisterGenericCommands(runtimeRegistry, engine, aof)
	command.RegisterListCommands(runtimeRegistry, engine, aof)
	runtimeDispatcher := command.NewDispatcher(runtimeRegistry)

	// Restore data from AOF before accepting connections.
	if err := aof.Replay(replayDispatcher); err != nil {
		log.Printf("[WARN] AOF replay error: %v", err)
	}

	parser := protocol.NewRESPParser()
	srv := server.New(cfg, parser, runtimeDispatcher)
	return srv.Start()
}
