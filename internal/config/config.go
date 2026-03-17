package config

type Config struct {
	Addr            string
	Workers         int
	MaxMemoryBytes  int64
	AppendOnlyPath  string
	SnapshotRDBPath string
}

func Default() Config {
	return Config{
		Addr:            ":6380",
		Workers:         4,
		MaxMemoryBytes:  64 << 20,
		AppendOnlyPath:  "appendonly.aof",
		SnapshotRDBPath: "dump.rdb",
	}
}
