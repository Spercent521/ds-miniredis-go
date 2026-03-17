package main

import (
	"flag"
	"log"

	"goredis/internal/app"
	"goredis/internal/config"
)

func main() {
	cfg := config.Default()
	flag.StringVar(&cfg.Addr, "addr", cfg.Addr, "TCP listen address")
	flag.IntVar(&cfg.Workers, "workers", cfg.Workers, "worker pool size")
	flag.Int64Var(&cfg.MaxMemoryBytes, "maxmemory", cfg.MaxMemoryBytes, "max memory in bytes (0 = unlimited)")
	flag.StringVar(&cfg.AppendOnlyPath, "appendonly", cfg.AppendOnlyPath, "AOF file path")
	flag.Parse()

	log.Printf("go-redis starting on %s (workers=%d, maxmemory=%d bytes)",
		cfg.Addr, cfg.Workers, cfg.MaxMemoryBytes)

	if err := app.RunWithConfig(cfg); err != nil {
		log.Fatal(err)
	}
}
