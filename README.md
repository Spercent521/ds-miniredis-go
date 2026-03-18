# go-redis

> 用 Go 实现的简化 Redis 服务器，功能特性对齐 Redis C 源码思路。  
> 支持 RESP2 协议、TCP 网络、AOF 持久化、TTL 过期、LRU 内存淘汰、worker pool 并发。

---

## 目录

- [功能特性](#功能特性)
- [构建与启动](#构建与启动)
- [启动参数](#启动参数)
- [命令参考](#命令参考)
- [功能演示](#功能演示)
  - [基础 SET / GET / DEL](#基础-set--get--del)
  - [过期时间 TTL](#过期时间-ttl)
  - [AOF 持久化](#aof-持久化)
  - [LRU 内存淘汰](#lru-内存淘汰)
  - [并发验证](#并发验证)
- [架构说明](#架构说明)

---

## 功能特性

| 特性 | 说明 |
|------|------|
| **基本数据结构** | String 类型，Object 模型可扩展到 List/Hash/Set |
| **字符串命令** | SET / GET / DEL / EXISTS / KEYS / DBSIZE / FLUSHDB |
| **RESP2 协议** | Array + Bulk String 解析；五种响应类型编码 |
| **TCP 网络服务** | 每个连接绑定到 worker pool，默认监听 `:6380` |
| **过期时间** | `EX`（秒）/ `PX`（毫秒）/ `EXPIRE` / `PEXPIRE` / `TTL` / `PTTL` |
| **AOF 持久化** | 写命令追加到文件；启动时自动重放恢复数据 |
| **LRU 内存淘汰** | 超过 `--maxmemory` 时淘汰最久未访问的 key |
| **Worker Pool** | 固定大小 goroutine 池处理并发连接（`--workers`） |

---

## 构建与启动

**前置条件**：Go 1.22+

```bash
# 进入项目目录
cd utils-ds

# 直接运行（使用默认配置）
go run ./cmd/goredis

# 编译为可执行文件
go build -o goredis ./cmd/goredis

# 运行可执行文件
./goredis
```

启动成功后输出：

```
go-redis starting on :6380 (workers=4, maxmemory=67108864 bytes)
```

---

## 启动参数

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-addr` | `:6380` | TCP 监听地址 |
| `-workers` | `4` | worker pool 大小（并发处理连接数） |
| `-maxmemory` | `67108864`（64MB） | 最大内存字节数，`0` 表示不限制 |
| `-appendonly` | `appendonly.aof` | AOF 文件路径 |

```bash
# 示例：自定义端口、2个worker、8MB 内存上限
./goredis -addr :6399 -workers 2 -maxmemory 8388608
```

---

## 命令参考

可使用任意 Redis 协议客户端连接本服务；`redis-cli` 只是调试时的便捷工具，不是运行依赖。

例如：

- 使用 `redis-cli -p 6380`
- 使用 `nc` 直接发送 RESP2 报文

```bash
# PING
printf '*1\r\n$4\r\nPING\r\n' | nc 127.0.0.1 6380

# SET name alice
printf '*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$5\r\nalice\r\n' | nc 127.0.0.1 6380

# GET name
printf '*2\r\n$3\r\nGET\r\n$4\r\nname\r\n' | nc 127.0.0.1 6380
```

下面命令语义与返回说明对所有客户端一致。

### 连接测试

| 命令 | 返回 | 说明 |
|------|------|------|
| `PING` | `+PONG` | 心跳检测 |
| `PING hello` | `$5\r\nhello` | 带消息的 PING |

### 字符串操作

| 命令 | 返回 | 说明 |
|------|------|------|
| `SET key value` | `+OK` | 设置键值 |
| `SET key value EX 60` | `+OK` | 设置键值并在 60 秒后过期 |
| `SET key value PX 5000` | `+OK` | 设置键值并在 5000 毫秒后过期 |
| `GET key` | `$N\r\nvalue` 或 `$-1` | 获取值，不存在返回 nil |
| `DEL key [key ...]` | `:N` | 删除键，返回成功删除数 |

### Key 管理

| 命令 | 返回 | 说明 |
|------|------|------|
| `EXISTS key [key ...]` | `:N` | 返回存在的 key 数量 |
| `EXPIRE key seconds` | `:1` 或 `:0` | 设置过期（秒），key 不存在返回 0 |
| `PEXPIRE key ms` | `:1` 或 `:0` | 设置过期（毫秒） |
| `TTL key` | `:N` 或 `:-1` 或 `:-2` | 剩余秒数；-1 无过期；-2 不存在 |
| `PTTL key` | `:N` 或 `:-1` 或 `:-2` | 剩余毫秒数 |
| `KEYS pattern` | 数组 | 返回匹配 `pattern` 的所有 key（支持 `*` `?` `[...]`） |
| `DBSIZE` | `:N` | 返回当前 key 总数 |
| `FLUSHDB` | `+OK` | 清空所有数据并清空 AOF 文件 |

---

## 功能演示

### 基础 SET / GET / DEL

```bash
redis-cli -p 6380

> PING
PONG

> SET name "alice"
OK

> GET name
"alice"

> SET age 25
OK

> KEYS *
1) "age"
2) "name"

> DBSIZE
(integer) 2

> DEL name
(integer) 1

> GET name
(nil)

> EXISTS name age
(integer) 1
```

---

### 过期时间 TTL

```bash
> SET token "abc123" EX 5
OK

> TTL token
(integer) 5

> PTTL token
(integer) 4873

# 等待 5 秒后...
> GET token
(nil)

> TTL token
(integer) -2
```

用 EXPIRE 给已有 key 追加过期：

```bash
> SET session "xyz"
OK

> EXPIRE session 10
(integer) 1

> TTL session
(integer) 10

# 对不存在的 key 执行 EXPIRE
> EXPIRE no_such_key 10
(integer) 0
```

---

### AOF 持久化

AOF 文件（默认 `appendonly.aof`）在每次写命令成功后追加 RESP 格式记录。

**演示步骤：**

```bash
# 1. 启动服务
./goredis

# 2. 写入数据
redis-cli -p 6380 SET user:1 alice
redis-cli -p 6380 SET user:2 bob
redis-cli -p 6380 SET counter 100

# 3. 停止服务（Ctrl+C）

# 4. 查看 AOF 文件内容
cat appendonly.aof
# 输出 RESP 格式的命令记录：
# *3
# $3
# SET
# $6
# user:1
# $5
# alice
# ...

# 5. 重新启动服务（会自动 replay AOF）
./goredis

# 6. 数据恢复验证
redis-cli -p 6380 GET user:1    # → alice
redis-cli -p 6380 GET user:2    # → bob
redis-cli -p 6380 GET counter   # → 100
```

---

### LRU 内存淘汰

设置很小的内存上限来触发淘汰：

```bash
# 启动时限制内存为 300 字节（演示用；每个 key 约 64+len 字节开销）
./goredis -maxmemory 300

redis-cli -p 6380

> SET k1 "value1"
OK
> SET k2 "value2"
OK
> SET k3 "value3"
OK

# 访问 k1，刷新它的 LRU 位置（使它不是最旧的）
> GET k1
"value1"

# 插入新 key，内存超限，k2 或 k3（最久未访问的）被自动淘汰
> SET k4 "value4"
OK

# k2 被淘汰（最久未访问），k1 和 k3 保留
> GET k2
(nil)
> GET k1
"value1"
```

**淘汰规则**：每次 `SET` 后如果 `usedBytes > maxmemory`，从 LRU 链表尾部（最久未访问）依次驱逐，直到内存回到限制以下。

---

### 并发验证

worker pool 默认大小为 4，可通过 `-workers` 调整。

```bash
# 启动服务，8 个 worker
./goredis -workers 8

# 使用 redis-benchmark 发送并发请求（需要安装 redis-tools）
redis-benchmark -p 6380 -n 10000 -c 20 -t set,get

# 或者用 Shell 并发脚本
for i in $(seq 1 20); do
  redis-cli -p 6380 SET "key$i" "val$i" &
done
wait
redis-cli -p 6380 DBSIZE
```

---

## 架构说明

```
请求链路：TCP连接 → Worker Pool → RESP Parser → Command Dispatcher → DB Engine
持久化  ：写命令  → AOF.Append()；启动时 AOF.Replay()
淘汰链路：DB.SetString() → LRU.Touch() → 超限时 LRU.Evict()
```

### 包职责（对齐 Redis C 子系统）

| Go 包 | 对应 Redis C 文件 | 职责 |
|-------|-----------------|------|
| `server` | `server.c` `networking.c` | Accept、连接生命周期、worker pool 分发 |
| `protocol` | 协议层 | RESP2 解析与编码 |
| `command` | `t_string.c` 等 | 命令注册表、参数校验、分发 |
| `db` | `db.c` | 键空间、对象模型、LRU、过期惰性删除 |
| `expire` | `expire.c` | TTL 工具方法 |
| `persistence` | `aof.c` `rdb.c` | AOF 追加与重放、RDB 接口占位 |
| `eviction` | `evict.c` | LRU 双向链表 + map 实现 |
| `pool` | — | Worker goroutine 池 |
| `config` / `app` | — | 启动装配与依赖注入 |

### 目录结构

```text
utils-ds/
  cmd/goredis/main.go            ← 启动入口，flag 解析
  internal/
    app/app.go                   ← 装配所有组件
    config/config.go             ← 配置结构与默认值
    server/
      server.go                  ← TCP Accept + worker pool
      connection.go              ← ConnHandler 接口
    protocol/
      types.go                   ← RESP 类型常量
      resp_parser.go             ← 请求解析
      resp_writer.go             ← 响应编码
    command/
      registry.go                ← 命令注册表
      dispatcher.go              ← 命令分发
      cmd_string.go              ← SET / GET / DEL
      cmd_generic.go             ← PING / EXPIRE / TTL / KEYS / FLUSHDB ...
    db/
      object.go                  ← Object / ObjectType 定义
      db.go                      ← DB 结构（含 LRU + 内存计数）
      string.go                  ← String 类型读写 + 淘汰
      generic.go                 ← Expire / TTL / Keys / FlushDB
    expire/expire.go             ← TTL 工具
    persistence/
      aof.go                     ← AOF 追加 / 重放 / 清空
      rdb.go                     ← RDB Snapshotter 接口（待实现）
    eviction/
      policy.go                  ← Policy 接口
      lru.go                     ← LRU 双向链表实现
    pool/worker_pool.go          ← Worker Pool
  go.mod
  README.md
```
