# go-redis（学习版）复刻计划

> 目标：在 **2 小时可理解 + 可动手** 的前提下，做一个你自己的简化 Redis（参考 miniredis 思路，但不复制实现）。

## 你要完成的功能（按优先级）

### 第一阶段：先跑起来（MVP）
- [ ] 实现 Redis 中的基本数据结构（先支持 String，结构上可扩展）
- [ ] 支持字符串命令：`SET`、`GET`、`DEL`
- [ ] 支持 RESP 协议解析（至少 RESP2 的 Array + Bulk String）
- [ ] 实现简易网络服务器（TCP）

### 第二阶段：进阶功能
- [ ] 数据持久化（建议先做 AOF 简易版）
- [ ] 支持过期时间（`EX` / `PX`）
- [ ] 线程池处理并发请求（worker pool）
- [ ] 内存淘汰策略（LRU）

---

## 和 miniredis 的关系（非常重要）

- miniredis 是 **参考项目**：看它如何组织命令、协议和测试。
- 你的项目要自己写：
  - 自己定义 `DB`、`RESP parser`、`Server`、`Command dispatcher`
  - 不直接拷贝 miniredis 源码
- 建议“借鉴接口和拆分思路”，不借鉴具体实现细节。

---

## 2 小时冲刺学习安排

### 0 ~ 20 分钟：搭骨架
1. 创建项目目录和 Go module
2. 跑一个最小 TCP echo server（确认网络代码可跑）
3. 定义核心结构：`Entry`、`DB`、`Server`

### 20 ~ 50 分钟：RESP + SET/GET/DEL
1. 先只解析：`*<n>\r\n$<len>\r\n<arg>\r\n...`
2. 命令分发：`switch strings.ToUpper(cmd)`
3. 完成：`SET key value`、`GET key`、`DEL key`

### 50 ~ 80 分钟：TTL + AOF
1. `SET key value EX seconds`（或 `PX ms`）
2. 读请求时惰性删除过期键
3. 每次写命令追加到 `appendonly.aof`
4. 启动时重放 AOF

### 80 ~ 120 分钟：并发 + LRU
1. worker pool 处理连接任务
2. 加 `maxmemory`（按 key 数量或近似字节数都可）
3. 内存超限时淘汰最久未访问 key（LRU）
4. 最后用 `redis-cli` 做回归验证

---

## 推荐项目结构（你可以直接照抄这个结构）

```text
go-redis/
  cmd/server/main.go
  internal/
    server/
      server.go        # 监听、accept、连接处理、worker pool
    protocol/
      resp.go          # RESP 解析与响应编码
    command/
      dispatcher.go    # SET/GET/DEL/EXPIRE 解析分发
    db/
      db.go            # 核心存储、TTL、LRU
      aof.go           # AOF 追加与重放
      lru.go           # LRU 双向链表 + map（或 container/list）
  go.mod
  README.md
```

---

## 核心设计（简化但可扩展）

### 1) 数据结构

```go
type ValueType int

const (
    TypeString ValueType = iota
)

type Entry struct {
    Type       ValueType
    Str        string
    ExpireAtMs int64 // 0 表示不过期
    LastAccess int64 // LRU 使用
}

type DB struct {
    mu    sync.RWMutex
    data  map[string]*Entry
    lru   *LRU
    aof   *AOF
    maxKB int64
}
```

> 先只做 String，但 `Type` 保留，后续扩展 List/Hash/Set 时不推翻结构。

### 2) RESP 协议（最小实现）

只需要支持客户端常用命令格式：
- 请求：Array（`*`）
- 参数：Bulk String（`$`）

返回支持：
- `+OK\r\n`
- `$<len>\r\n<val>\r\n`
- `$-1\r\n`（nil）
- `:<n>\r\n`（整数）
- `-ERR msg\r\n`

### 3) 命令语义（本项目最小子集）
- `SET key value [EX seconds|PX milliseconds]`
- `GET key`
- `DEL key [key ...]`

### 4) 过期策略
- 惰性删除：`GET/SET/DEL` 访问 key 时检查过期
- 可选主动删除：后台 ticker 每秒随机扫描部分 key

### 5) AOF（建议优先于 RDB）
- 写命令成功后，原始 RESP 命令追加到 AOF
- 启动时逐条解析并执行（关闭再次写 AOF，避免重复）

### 6) 线程池并发模型
- 一个 goroutine 负责 `Accept`
- 连接任务投递到 `jobs chan net.Conn`
- N 个 worker 从 `jobs` 消费并处理请求
- `DB` 用 `sync.RWMutex` 保护

### 7) LRU 淘汰（简化版）
- 每次 `GET` 或命中 `SET` 更新访问时间/链表位置
- 超过 `maxmemory`（你可按 key 数近似）时：
  - 从 LRU 尾部淘汰
  - 同步删除 `data` 中的 key

---

## 开发顺序（建议严格按这个顺序）

1. `server + RESP parser` 跑通
2. `SET/GET/DEL` 正确返回 RESP
3. `TTL`（`EX/PX`）
4. `AOF` append + replay
5. `worker pool`
6. `LRU`

> 每完成一步都用 `redis-cli` 验证，不要等全部写完再调试。

---

## 验收用例（可手工）

```bash
# 连接
redis-cli -p 6380

# 字符串
SET name alice
GET name
DEL name
GET name

# 过期
SET token abc EX 2
GET token
# 2秒后再查应为 nil
GET token
```

AOF 验证：
1. 启动服务，执行若干 `SET`
2. 停服务
3. 重启服务
4. `GET` 仍能取到数据

LRU 验证：
1. 设置小内存上限（比如最多 3 个 key）
2. 访问顺序：`k1 k2 k3`，再访问 `k1`
3. 插入 `k4`，应优先淘汰 `k2`

---

## 参考 miniredis 的阅读路径（只看思路）

你当前仓库里可优先阅读：
- `miniredis/miniredis.go`（服务主结构）
- `miniredis/redis.go`（协议相关）
- `miniredis/cmd_string.go`（字符串命令实现）
- `miniredis/db.go`（数据组织方式）

阅读目标：
- 看“模块边界”
- 看“命令分发方式”
- 看“测试如何覆盖行为”

不要目标：
- 不要逐行仿写
- 不要一开始追求全命令集

---

## 你这次作业的交付建议

最小可交付（老师能验收）建议包含：
- `SET/GET/DEL`
- RESP2 基础解析
- TCP server
- TTL
- AOF
- worker pool
- LRU（简化版）
- 一份你自己的测试说明（哪怕是手工命令清单）

---

## 最后建议

你目前 Go 学习时间不长，最关键是：
- 先把主流程打通（能连、能存、能查）
- 再补高级特性（TTL/AOF/LRU）
- 每一步都“可运行、可验证”

如果你愿意，下一步我可以直接在根目录给你生成一个 **最小 go-redis 项目骨架代码**（包含可运行的 `SET/GET/DEL + RESP + TCP`），你只需要继续补 TTL/AOF/LRU。