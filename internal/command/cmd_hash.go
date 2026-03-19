package command

import (
	"goredis/internal/db"
	"goredis/internal/persistence"
	"goredis/internal/protocol"
)

// RegisterHashCommands 注册 Hash 相关命令
func RegisterHashCommands(reg *Registry, engine *db.DB, aof *persistence.AOF) {
	
	reg.Register("HSET", func(args []string) (string, error) {
		if len(args) != 3 {
			return protocol.EncodeError("wrong number of arguments for 'hset' command"), nil
		}
		key, field, value := args[0], args[1], args[2]
		
		res := engine.HSet(key, field, value)
		
		if aof != nil {
			_ = aof.Append(encodeRESP(append([]string{"HSET"}, args...)))
		}
		return protocol.EncodeInteger(int64(res)), nil
	})

	reg.Register("HGET", func(args []string) (string, error) {
		if len(args) != 2 {
			return protocol.EncodeError("wrong number of arguments for 'hget' command"), nil
		}
		
		val, ok := engine.HGet(args[0], args[1])
		if !ok {
			return protocol.EncodeBulkString(nil), nil
		}
		return protocol.EncodeBulkString(&val), nil
	})

	reg.Register("HGETALL", func(args []string) (string, error) {
		if len(args) != 1 {
			return protocol.EncodeError("wrong number of arguments for 'hgetall' command"), nil
		}
		
		result := engine.HGetAll(args[0])
		if len(result) == 0 {
			// 返回空的 Array: *0\r\n
			return "*0\r\n", nil
		}
		return protocol.EncodeArray(result), nil
	})
}