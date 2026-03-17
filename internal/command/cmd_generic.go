package command

import (
	"strconv"
	"time"

	"goredis/internal/db"
	"goredis/internal/persistence"
	"goredis/internal/protocol"
)

func RegisterGenericCommands(reg *Registry, engine *db.DB, aof *persistence.AOF) {
	// PING [message] → +PONG or bulk message
	reg.Register("PING", func(args []string) (string, error) {
		if len(args) == 0 {
			return protocol.EncodeSimpleString("PONG"), nil
		}
		return protocol.EncodeBulkString(&args[0]), nil
	})

	// EXISTS key [key ...] → integer (count of existing keys)
	reg.Register("EXISTS", func(args []string) (string, error) {
		if len(args) == 0 {
			return protocol.EncodeError("wrong number of arguments for 'exists'"), nil
		}
		return protocol.EncodeInteger(int64(engine.Exists(args...))), nil
	})

	// EXPIRE key seconds → 1 if set, 0 if key not found
	reg.Register("EXPIRE", func(args []string) (string, error) {
		if len(args) != 2 {
			return protocol.EncodeError("wrong number of arguments for 'expire'"), nil
		}
		secs, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil || secs < 0 {
			return protocol.EncodeError("invalid expire time in 'expire'"), nil
		}
		expireAtMs := time.Now().Add(time.Duration(secs) * time.Second).UnixMilli()
		ok := engine.Expire(args[0], expireAtMs)
		if ok && aof != nil {
			_ = aof.Append(encodeRESP(append([]string{"EXPIRE"}, args...)))
		}
		result := int64(0)
		if ok {
			result = 1
		}
		return protocol.EncodeInteger(result), nil
	})

	// PEXPIRE key milliseconds → 1 if set, 0 if key not found
	reg.Register("PEXPIRE", func(args []string) (string, error) {
		if len(args) != 2 {
			return protocol.EncodeError("wrong number of arguments for 'pexpire'"), nil
		}
		ms, err := strconv.ParseInt(args[1], 10, 64)
		if err != nil || ms < 0 {
			return protocol.EncodeError("invalid expire time in 'pexpire'"), nil
		}
		expireAtMs := time.Now().Add(time.Duration(ms) * time.Millisecond).UnixMilli()
		ok := engine.Expire(args[0], expireAtMs)
		if ok && aof != nil {
			_ = aof.Append(encodeRESP(append([]string{"PEXPIRE"}, args...)))
		}
		result := int64(0)
		if ok {
			result = 1
		}
		return protocol.EncodeInteger(result), nil
	})

	// TTL key → remaining seconds (-1 no expiry, -2 not exist/expired)
	reg.Register("TTL", func(args []string) (string, error) {
		if len(args) != 1 {
			return protocol.EncodeError("wrong number of arguments for 'ttl'"), nil
		}
		ms := engine.TTLMs(args[0])
		if ms < 0 {
			return protocol.EncodeInteger(ms), nil
		}
		return protocol.EncodeInteger(ms / 1000), nil
	})

	// PTTL key → remaining milliseconds (-1 no expiry, -2 not exist/expired)
	reg.Register("PTTL", func(args []string) (string, error) {
		if len(args) != 1 {
			return protocol.EncodeError("wrong number of arguments for 'pttl'"), nil
		}
		return protocol.EncodeInteger(engine.TTLMs(args[0])), nil
	})

	// KEYS pattern → array of matching key names
	reg.Register("KEYS", func(args []string) (string, error) {
		if len(args) != 1 {
			return protocol.EncodeError("wrong number of arguments for 'keys'"), nil
		}
		return protocol.EncodeArray(engine.Keys(args[0])), nil
	})

	// DBSIZE → total number of keys
	reg.Register("DBSIZE", func(args []string) (string, error) {
		return protocol.EncodeInteger(int64(engine.DBSize())), nil
	})

	// FLUSHDB → removes all keys and clears the AOF
	reg.Register("FLUSHDB", func(args []string) (string, error) {
		engine.FlushDB()
		if aof != nil {
			_ = aof.Truncate()
		}
		return protocol.EncodeSimpleString("OK"), nil
	})

	// COMMAND (and sub-commands) – minimal stub for redis-cli compatibility
	reg.Register("COMMAND", func(args []string) (string, error) {
		return protocol.EncodeArray(nil), nil
	})
}
