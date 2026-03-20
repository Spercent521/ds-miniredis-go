# ds-miniredis-go

> 用 Go 实现的教学版 miniRedis。当前版本已支持 `string / list / set / zset / hash` 五类数据结构，
> 并将全局锁升级为 **16 分片局部锁（sharded lock）**，提升并发访问能力。

---

## 1. 当前版本特性

- 数据结构：`String`、`List`、`Set`、`Hash`、`ZSet`
- 协议：RESP2（可直接使用 `redis-cli` / `redis-benchmark`）
- 并发：按 key 路由到 16 个分片，每个分片独立 `RWMutex`
- 过期：`EX/PX`（SET 内）、`EXPIRE/PEXPIRE`、`TTL/PTTL`
- 内存淘汰：基础 LRU（主要在 String / List 路径生效）
- 持久化：启动时支持 AOF Replay（用于恢复历史数据）

---

## 2. 启动方式对比（miniRedis vs 原生 Redis）

### 2.1 启动 miniRedis（本项目）

```bash
# 进入项目目录
cd ds-miniredis-go

# 直接运行
go run ./cmd/goredis

# 或编译后运行
go build -o goredis ./cmd/goredis
./goredis
```

默认参数：

- `-addr :6380`
- `-workers 4`
- `-maxmemory 67108864`
- `-appendonly appendonly.aof`

自定义示例：

```bash
./goredis -addr :6380 -workers 8 -maxmemory 104857600 -appendonly appendonly.aof
```

启动日志示例：

```text
go-redis starting on :6380 (workers=8, maxmemory=104857600 bytes)
```

### 2.2 启动原生 Redis

```bash
# 默认启动（常见默认端口 6379）
redis-server

# 指定端口
redis-server --port 6379
```

### 2.3 客户端连接对照

```bash
# 连接 miniRedis
redis-cli -h 127.0.0.1 -p 6380

# 连接原生 Redis
redis-cli -h 127.0.0.1 -p 6379
```

---

## 3. 命令支持总览（按数据结构）

> 说明：以下为“当前代码已实现”命令，不含 Redis 全量指令集。

### 3.1 通用命令

| 命令 | 示例 | 返回类型 | 说明 |
|---|---|---|---|
| `PING [msg]` | `PING` / `PING hi` | SimpleString / BulkString | 心跳或回显 |
| `EXISTS key [key...]` | `EXISTS a b` | Integer | 返回存在 key 个数 |
| `DEL key [key...]` | `DEL a b` | Integer | 返回删除成功个数 |
| `KEYS pattern` | `KEYS user:*` | Array | 模式匹配 key |
| `DBSIZE` | `DBSIZE` | Integer | 当前 key 总数 |
| `FLUSHDB` | `FLUSHDB` | SimpleString | 清空当前 DB |
| `COMMAND` | `COMMAND` | Array | 兼容桩实现（空数组） |

### 3.2 过期时间相关

| 命令 | 示例 | 返回类型 | 说明 |
|---|---|---|---|
| `SET key val EX seconds` | `SET token abc EX 10` | SimpleString | 写入并设置秒级过期 |
| `SET key val PX ms` | `SET token abc PX 5000` | SimpleString | 写入并设置毫秒过期 |
| `EXPIRE key seconds` | `EXPIRE token 20` | Integer | 设置过期成功=1，失败=0 |
| `PEXPIRE key ms` | `PEXPIRE token 8000` | Integer | 设置过期成功=1，失败=0 |
| `TTL key` | `TTL token` | Integer | 秒级 TTL（`-1/-2` 含义见下） |
| `PTTL key` | `PTTL token` | Integer | 毫秒级 TTL（`-1/-2` 含义见下） |

### 3.3 String

| 命令 | 示例 | 返回类型 |
|---|---|---|
| `SET key value` | `SET name alice` | SimpleString |
| `GET key` | `GET name` | BulkString / NullBulk |

### 3.4 List

| 命令 | 示例 | 返回类型 | 说明 |
|---|---|---|---|
| `LPUSH key v1 [v2...]` | `LPUSH q a b c` | Integer | 返回新长度 |
| `LPOP key` | `LPOP q` | BulkString / NullBulk | 弹出并返回头元素 |

### 3.5 Set

| 命令 | 示例 | 返回类型 | 说明 |
|---|---|---|---|
| `SADD key m1 [m2...]` | `SADD s a b c` | Integer | 返回新增成员数 |
| `SMEMBERS key` | `SMEMBERS s` | Array | 返回全部成员（无序） |
| `SISMEMBER key member` | `SISMEMBER s a` | Integer | 是成员=1，否则0 |

### 3.6 Hash

| 命令 | 示例 | 返回类型 | 说明 |
|---|---|---|---|
| `HSET key field value` | `HSET user:1 name alice` | Integer | 新字段=1，覆盖=0 |
| `HGET key field` | `HGET user:1 name` | BulkString / NullBulk | 读取字段值 |
| `HGETALL key` | `HGETALL user:1` | Array | 平铺返回 `field value ...` |

### 3.7 ZSet

| 命令 | 示例 | 返回类型 | 说明 |
|---|---|---|---|
| `ZADD key score member [score member...]` | `ZADD rank 100 tom 98 bob` | Integer | 返回新增成员数 |
| `ZRANGE key start stop [WITHSCORES]` | `ZRANGE rank 0 -1 WITHSCORES` | Array | 按分值升序返回 |

---

## 4. 输出结果解释（RESP2）

项目遵循 RESP2 编码，`redis-cli` 已自动帮你把协议格式“翻译成人类可读输出”。

### 4.1 原始 RESP 含义

- `+OK`：Simple String，通常表示命令成功
- `:1`：Integer，常用于计数、布尔语义（1/0）
- `$5\r\nhello`：Bulk String（长度 + 内容）
- `$-1`：Null Bulk String（不存在）
- `*3 ...`：Array（数组，多元素返回）
- `-ERR ...`：错误返回

### 4.2 常见返回语义

- `GET` 不存在 key → `(nil)`（协议层是 `$-1`）
- `TTL/PTTL`：
  - `-1`：key 存在，但未设置过期
  - `-2`：key 不存在（或已过期）
- `EXPIRE/PEXPIRE`：
  - `1`：设置成功
  - `0`：key 不存在，设置失败

---

## 5. miniRedis 与原生 Redis 命令对比示例

### 5.1 基础 string 对比

```bash
# miniRedis (6380)
redis-cli -p 6380 SET name alice
redis-cli -p 6380 GET name

# Redis (6379)
redis-cli -p 6379 SET name alice
redis-cli -p 6379 GET name
```

说明：对于上述基础命令，交互体验和返回形式基本一致。

### 5.2 list 对比（你终端中的压测场景）

```bash
# miniRedis
redis-benchmark -h 127.0.0.1 -p 6380 -c 1 -n 100000 -t lpush,lpop -q

# Redis
redis-benchmark -h 127.0.0.1 -p 6379 -c 1 -n 100000 -t lpush,lpop -q
```

说明：

- `-q` 只输出摘要吞吐结果（如 `LPUSH: xxxx requests per second`）
- 在相同机器和参数下，数值可用于横向比较；不同机器不建议直接对比绝对值

### 5.3 set/get 对比（你终端中的并发场景）

```bash
# miniRedis
redis-benchmark -h 127.0.0.1 -p 6380 -c 10 -n 100000 -t set,get -q

# Redis
redis-benchmark -h 127.0.0.1 -p 6379 -c 10 -n 100000 -t set,get -q
```

---

## 6. 并发模型说明（本次优化重点）

### 6.1 从全局锁到局部锁

- 旧模型：单把全局锁，所有 key 竞争同一临界区
- 新模型：16 个分片，每个分片独立 `RWMutex`
- 路由方式：对 key 做 FNV 哈希后取模，落到固定分片

这意味着：

- 不同分片上的 key 可以并行读写，锁冲突显著下降
- 热点 key 仍会在同分片内竞争（符合分片锁预期）

### 6.2 路由与聚合行为

- 单 key 命令（如 `GET/HGET/LPUSH`）只访问对应分片
- 多 key 命令（如 `DEL/EXISTS`）会逐 key 路由并汇总结果
- 全局命令（如 `KEYS/DBSIZE/FLUSHDB`）会遍历所有分片并聚合

---

## 7. 边界与注意事项

- 本项目是课程/教学 miniRedis，不追求与官方 Redis 全量一致
- 某些命令在类型冲突场景下采用简化处理（返回值而非完整 Redis 错误语义）
- `SMEMBERS/HGETALL` 的元素顺序受 Go map 迭代影响，不保证稳定顺序
- `DBSIZE` 为当前 map 长度，可能包含尚未惰性清理的过期 key
- 若要做严格对比实验，建议固定：端口、并发数、请求总数、测试命令、机器负载

---

## 8. 目录结构（核心）

```text
ds-miniredis-go/
  cmd/goredis/main.go
  internal/
    app/app.go
    command/
      cmd_string.go
      cmd_list.go
      cmd_set.go
      cmd_hash.go
      cmd_zset.go
      cmd_generic.go
    db/
      db.go           # 16 分片 + 局部锁
      router.go       # 命令路由到分片
      string.go
      list.go
      set.go
      hash.go
      zset.go
      generic.go
    protocol/
      resp_parser.go
      resp_writer.go
    server/
      server.go
```

---
