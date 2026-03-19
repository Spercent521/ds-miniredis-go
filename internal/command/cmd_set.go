package command

import (
	"goredis/internal/db"
	"goredis/internal/persistence"
	"goredis/internal/protocol"
)

// RegisterSetCommands 注册 Set 相关命令
func RegisterSetCommands(reg *Registry, engine *db.DB, aof *persistence.AOF) {
	
	// SADD key member [member ...]
	reg.Register("SADD", func(args []string) (string, error) {
		if len(args) < 2 {
			return protocol.EncodeError("wrong number of arguments for 'sadd' command"), nil
		}
		
		key := args[0]
		members := args[1:]
		
		added := engine.SAdd(key, members...)
		
		if aof != nil {
			_ = aof.Append(encodeRESP(append([]string{"SADD"}, args...)))
		}
		return protocol.EncodeInteger(int64(added)), nil
	})

	// SMEMBERS key
	reg.Register("SMEMBERS", func(args []string) (string, error) {
		if len(args) != 1 {
			return protocol.EncodeError("wrong number of arguments for 'smembers' command"), nil
		}
		
		result := engine.SMembers(args[0])
		if len(result) == 0 {
			return "*0\r\n", nil
		}
		return protocol.EncodeArray(result), nil
	})

	// SISMEMBER key member
	reg.Register("SISMEMBER", func(args []string) (string, error) {
		if len(args) != 2 {
			return protocol.EncodeError("wrong number of arguments for 'sismember' command"), nil
		}
		
		isMember := engine.SIsMember(args[0], args[1])
		return protocol.EncodeInteger(int64(isMember)), nil
	})
}