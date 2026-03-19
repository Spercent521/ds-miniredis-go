package command

import (
	"strconv"
	"strings"

	"goredis/internal/db"
	"goredis/internal/persistence"
	"goredis/internal/protocol"
)

func RegisterZSetCommands(reg *Registry, engine *db.DB, aof *persistence.AOF) {

	// ZADD key score member [score member ...]
	reg.Register("ZADD", func(args []string) (string, error) {
		// 参数数量必须是奇数（1个key + N对 score和member）
		if len(args) < 3 || len(args)%2 == 0 {
			return protocol.EncodeError("wrong number of arguments for 'zadd' command"), nil
		}

		key := args[0]
		var pairs []db.ZSetMember

		// 解析 score 和 member 键值对
		for i := 1; i < len(args); i += 2 {
			score, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				return protocol.EncodeError("value is not a valid float"), nil
			}
			pairs = append(pairs, db.ZSetMember{
				Score:  score,
				Member: args[i+1],
			})
		}

		added := engine.ZAdd(key, pairs...)

		if aof != nil {
			_ = aof.Append(encodeRESP(append([]string{"ZADD"}, args...)))
		}
		return protocol.EncodeInteger(int64(added)), nil
	})

	// ZRANGE key start stop [WITHSCORES]
	reg.Register("ZRANGE", func(args []string) (string, error) {
		if len(args) < 3 {
			return protocol.EncodeError("wrong number of arguments for 'zrange' command"), nil
		}

		key := args[0]
		start, err1 := strconv.Atoi(args[1])
		stop, err2 := strconv.Atoi(args[2])
		if err1 != nil || err2 != nil {
			return protocol.EncodeError("value is not an integer or out of range"), nil
		}

		withScores := false
		if len(args) == 4 && strings.ToUpper(args[3]) == "WITHSCORES" {
			withScores = true
		}

		result := engine.ZRange(key, start, stop, withScores)
		if len(result) == 0 {
			return "*0\r\n", nil
		}
		return protocol.EncodeArray(result), nil
	})
}