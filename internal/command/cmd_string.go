package command

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"goredis/internal/db"
	"goredis/internal/persistence"
	"goredis/internal/protocol"
)

// encodeRESP encodes a command+args back to RESP format for AOF storage.
func encodeRESP(argv []string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "*%d\r\n", len(argv))
	for _, a := range argv {
		fmt.Fprintf(&sb, "$%d\r\n%s\r\n", len(a), a)
	}
	return sb.String()
}

func RegisterStringCommands(reg *Registry, engine *db.DB, aof *persistence.AOF) {
	reg.Register("SET", func(args []string) (string, error) {
		if len(args) < 2 {
			return protocol.EncodeError("wrong number of arguments for 'set'"), nil
		}
		key, value := args[0], args[1]
		var expireAtMs int64

		// Parse optional EX / PX
		for i := 2; i < len(args)-1; i++ {
			opt := strings.ToUpper(args[i])
			n, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil || n <= 0 {
				return protocol.EncodeError("invalid expire time in 'set'"), nil
			}
			switch opt {
			case "EX":
				expireAtMs = time.Now().Add(time.Duration(n) * time.Second).UnixMilli()
				i++
			case "PX":
				expireAtMs = time.Now().Add(time.Duration(n) * time.Millisecond).UnixMilli()
				i++
			}
		}

		engine.SetString(key, value, expireAtMs)

		if aof != nil {
			_ = aof.Append(encodeRESP(append([]string{"SET"}, args...)))
		}
		return protocol.EncodeSimpleString("OK"), nil
	})

	reg.Register("GET", func(args []string) (string, error) {
		if len(args) != 1 {
			return protocol.EncodeError("wrong number of arguments for 'get'"), nil
		}
		v, ok := engine.GetString(args[0])
		if !ok {
			return protocol.EncodeBulkString(nil), nil
		}
		return protocol.EncodeBulkString(&v), nil
	})

	reg.Register("DEL", func(args []string) (string, error) {
		if len(args) == 0 {
			return protocol.EncodeError("wrong number of arguments for 'del'"), nil
		}
		deleted := int64(engine.Del(args...))
		if aof != nil {
			_ = aof.Append(encodeRESP(append([]string{"DEL"}, args...)))
		}
		return protocol.EncodeInteger(deleted), nil
	})
}
