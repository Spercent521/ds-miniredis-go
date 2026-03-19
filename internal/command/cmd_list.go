package command

import (
	"goredis/internal/db"
	"goredis/internal/persistence"
	"goredis/internal/protocol"
)

// RegisterListCommands 注册所有与 List 相关的命令
func RegisterListCommands(reg *Registry, engine *db.DB, aof *persistence.AOF) {
	
	// 1. 注册 LPUSH 命令
	reg.Register("LPUSH", func(args []string) (string, error) {
		// LPUSH 至少需要 2 个参数：key 和 至少一个 value
		if len(args) < 2 {
			return protocol.EncodeError("wrong number of arguments for 'lpush' command"), nil
		}
		
		key := args[0]
		values := args[1:] // 剩下的全是 value

		// 调用你刚刚在 db 层面写的 LPush 方法
		newLen := engine.LPush(key, values...)

		// 如果开启了 AOF 持久化，把这条命令记录到磁盘中
		if aof != nil {
			// 将 "LPUSH" 和参数合并，复用 cmd_string.go 中的 encodeRESP 方法
			fullArgs := append([]string{"LPUSH"}, args...)
			_ = aof.Append(encodeRESP(fullArgs))
		}

		// 返回 RESP 协议格式的整数（即插入后列表的长度）
		return protocol.EncodeInteger(int64(newLen)), nil
	})

	// 2. 注册 LPOP 命令
	reg.Register("LPOP", func(args []string) (string, error) {
		// LPOP 只需要 1 个参数：key
		if len(args) != 1 {
			return protocol.EncodeError("wrong number of arguments for 'lpop' command"), nil
		}
		
		key := args[0]

		// 调用你写的 LPop 方法
		val, ok := engine.LPop(key)

		// 如果 ok 为 false，说明 Key 不存在或者已经弹空了
		if !ok {
			// 返回 RESP 中的空值 (Null Bulk String: $-1\r\n)
			return protocol.EncodeBulkString(nil), nil
		}

		// 记录 AOF 持久化
		if aof != nil {
			_ = aof.Append(encodeRESP(append([]string{"LPOP"}, args...)))
		}

		// 返回弹出的具体字符串数据
		return protocol.EncodeBulkString(&val), nil
	})
}