# 05 String

**基本要素与解决的问题** : String 是 Redis 中最基本的数据结构 , 最大能存储 512MB 的数据 , 它是二进制安全的 , 可以包含任何数据 ( 如文本、图片 )  . 它主要用于处理简单的缓存需求 , 如网页缓存、会话存储、计数器以及分布式锁等 . 

```bash
SET key value [EX seconds] [NX|XX]
# key : 要设置的键名 ( 区分大小写 ) 
# value : 要设置的具体内容
# EX : 可选 , 指定过期时间 ( 秒 ) 
# NX/XX : 可选 , NX 表示键不存在时才设置 , XX 表示键存在时才设置

GET key
# key : 要获取值的键名

DEL key [key ...]
# key : 要删除的键名

EXISTS key [key ...]
# key : 要检查存在性的键名

KEYS pattern
# pattern : 用于匹配键名的模式字符串 , 如 *

FLUSHALL
# 清空整个 Redis 服务器的数据

TTL key
# key : 查询剩余过期时间 ( 秒 , -1表示未设置 , -2表示已过期 ) 

EXPIRE key seconds
# key : 要设置过期的键名
# seconds : 过期的秒数

SETEX key seconds value
# key : 要设置的键名
# seconds : 过期秒数
# value : 具体内容

SETNX key value
# key : 要设置的键名
# value : 具体内容 , 仅不存在时设置
```

# 06 List

**基本要素与解决的问题** : List 是简单的字符串列表 , 按插入顺序排序 , 底层为双向链表 . 主要用于实现时间线、消息队列或者最新动态等场景 , 解决了需要保持数据顺序并且频繁在两端进行增删操作的问题 . 

```bash
LPUSH key element [element ...]
# key : 目标列表的键名
# element : 要插入到左侧头部的元素

RPUSH key element [element ...]
# key : 目标列表的键名
# element : 要插入到右侧尾部的元素

LPOP key [count]
# key : 目标列表的键名
# count : 可选 , 要左侧弹出的元素数量

RPOP key [count]
# key : 目标列表的键名
# count : 可选 , 要右侧弹出的元素数量

LRANGE key start stop
# key : 目标列表键名
# start/stop : 起止索引 , 0 是第一个 , -1 是最后一个

LLEN key
# key : 获取目标列表长度

LTRIM key start stop
# key : 保留设定范围内的元素 , 其余删除
```

# 07 Set

不重复 无序

```bash
SADD key member

SMEMBERS key # 展示key中所有元素

SISMEMBER key member # 判断member是否在集合中

SREM key value # 删除 集合 中的元素
```

## sinter 交集

```bash
# 准备数据
127.0.0.1:6379> SADD set1 a b c d
(integer) 4
127.0.0.1:6379> SADD set2 c d e f
(integer) 4
127.0.0.1:6379> SADD set3 d f g h
(integer) 4

# SINTER 基础用法 : 取多个集合的交集
127.0.0.1:6379> SINTER set1 set2 set3
1) "d"  # d 同时在三个集合中存在

127.0.0.1:6379> SINTER set1 set2
1) "c"
2) "d"  # c 和 d 同时在 set1 和 set2 中存在

# SINTER 特殊情况 : 只给一个 key
127.0.0.1:6379> SINTER set1
1) "a"
2) "b"
3) "c"
4) "d"  # 返回 set1 自身的所有元素

127.0.0.1:6379> SINTER non_exist_key
(empty array)  # key 不存在 , 返回空数组

# SINTERSTORE : 将交集结果保存到新集合
127.0.0.1:6379> SINTERSTORE result_set set1 set2
(integer) 2  # 返回结果集的元素个数
127.0.0.1:6379> SMEMBERS result_set
1) "c"
2) "d"  # result_set 中保存了 set1 和 set2 的交集

# 验证原集合没有被修改
127.0.0.1:6379> SMEMBERS set1
1) "a"
2) "b"
3) "c"
4) "d"  # set1 内容不变
```

## sunion 并集

```bash
# 准备数据
127.0.0.1:6379> SADD setA 1 2 3
(integer) 3
127.0.0.1:6379> SADD setB 3 4 5
(integer) 3

# SUNION 基础用法 : 取多个集合的并集
127.0.0.1:6379> SUNION setA setB
1) "1"
2) "2"
3) "3"
4) "4"
5) "5"  # 返回所有元素 , 自动去重

# 验证 SUNION 不会修改原集合
127.0.0.1:6379> SMEMBERS setA
1) "1"
2) "2"
3) "3"  # setA 内容不变

127.0.0.1:6379> SMEMBERS setB
1) "3"
2) "4"
3) "5"  # setB 内容不变

# SUNIONSTORE : 将并集结果保存起来
127.0.0.1:6379> SUNIONSTORE saved_union setA setB
(integer) 5  # 返回结果集的元素个数
127.0.0.1:6379> SMEMBERS saved_union
1) "1"
2) "2"
3) "3"
4) "4"
5) "5"  # saved_union 中保存了并集

# 方法1 : 用 SUNIONSTORE 覆盖第一个集合
127.0.0.1:6379> SUNIONSTORE setA setA setB
(integer) 5
127.0.0.1:6379> SMEMBERS setA
1) "1"
2) "2"
3) "3"
4) "4"
5) "5"  # setA 现在变成了并集

# 方法2 : 先存临时集合 , 再删除原集合 , 最后重命名
127.0.0.1:6379> SUNIONSTORE temp setX setY
(integer) 5
127.0.0.1:6379> DEL setX
(integer) 1
127.0.0.1:6379> RENAME temp setX
OK
127.0.0.1:6379> SMEMBERS setX
1) "a"
2) "b"
3) "c"
4) "d"
5) "e"  # setX 现在变成了并集
```

## sdiff 差集

```bash
# 准备数据
127.0.0.1:6379> SADD group1 apple banana cherry durian
(integer) 4
127.0.0.1:6379> SADD group2 banana cherry fig grape
(integer) 4
127.0.0.1:6379> SADD group3 cherry durian honey ice
(integer) 4

# SDIFF 基础用法 : 注意顺序很重要！
127.0.0.1:6379> SDIFF group1 group2
1) "apple"
2) "durian"  # group1 中有 , 但 group2 中没有的元素

127.0.0.1:6379> SDIFF group2 group1
1) "fig"
2) "grape"  # group2 中有 , 但 group1 中没有的元素

# SDIFF 多个集合 : 从第一个集合中减去后面所有集合的并集
127.0.0.1:6379> SDIFF group1 group2 group3
1) "apple"  # group1 中有 , 但 group2 和 group3 中都没有的元素
# 分析 : group1 有 {apple, banana, cherry, durian}
# group2 有 {banana, cherry, fig, grape}
# group3 有 {cherry, durian, honey, ice}
# 减去 {banana, cherry, fig, grape, durian, honey, ice} 后只剩 apple

# 特殊情况 : 空集合
127.0.0.1:6379> SDIFF group1 non_exist_set
1) "apple"
2) "banana"
3) "cherry"
4) "durian"  # 不存在的 key 视为空集合 , 所以返回 group1 全部

127.0.0.1:6379> SDIFF non_exist_set group1
(empty array)  # 第一个集合是空集合 , 结果为空

# SDIFFSTORE : 将差集结果保存到新集合
127.0.0.1:6379> SDIFFSTORE diff_result group1 group2
(integer) 2
127.0.0.1:6379> SMEMBERS diff_result
1) "apple"
2) "durian"  # diff_result 中保存了差集结果

# 覆盖第一个集合的方法
127.0.0.1:6379> SDIFFSTORE group1 group1 group2
(integer) 2
127.0.0.1:6379> SMEMBERS group1
1) "apple"
2) "durian"  # group1 现在变成了它自己和 group2 的差集
```

# 08 SortedSet/ZSet

- 每个元素都有一个**分数 score ** 用于排序 本质是浮点数
- 元素唯一 但是分数可以重复
- 自动按分数**从小到大排序**
- 底层实现 : **跳表 ( skiplist )  + 哈希表**

```bash
ZADD key score member [score member ...] 		# 添加或更新元素 , score 为分数 ( 用于从小到大排序 ) 

ZRANGE key start stop [WITHSCORES] 				# 按索引范围获取元素 ( 按分数从小到大排序 )  , 加上 WITHSCORES 可以同时获取分数
ZREVRANGE key start stop [WITHSCORES] 			# 反向按索引范围获取 ( 按分数从大到小排序 ) 

ZRANGEBYSCORE key min max [WITHSCORES] [LIMIT offset count] 		# 按分数范围获取元素
ZREVRANGEBYSCORE key max min [WITHSCORES] [LIMIT offset count] 		# 反向按分数范围获取 ( 注意 max 和 min 顺序 ) 

ZRANK key member 		# 获取元素的正序排名 ( 从 0 开始 , 分数越低排名越高 ) 
ZREVRANK key member 	# 获取元素的逆序排名 ( 从 0 开始 , 分数越高排名越高 ) 

ZSCORE key member 		# 获取指定元素的分数
ZCARD key 				# 获取集合中元素的总数量
ZCOUNT key min max	 	# 统计分数在 min 到 max 之间的元素数量

ZREM key member [member ...] 		# 移除一个或多个元素
ZREMRANGEBYRANK key start stop 		# 移除指定排名范围内的元素
ZREMRANGEBYSCORE key min max 		# 移除指定分数范围内的元素

ZINCRBY key increment member # 为元素的分数加上指定的增量 increment

ZPOPMAX key [count] # 弹出并返回分数最高的一个或多个元素
ZPOPMIN key [count] # 弹出并返回分数最低的一个或多个元素

ZINTER numkeys key [key ...] # 计算多个有序集合的交集
ZUNION numkeys key [key ...] # 计算多个有序集合的并集
ZDIFF numkeys key [key ...]  # 计算多个有序集合的差集

ZRANGEBYLEX key min max [LIMIT offset count] # 按字典序获取范围内的元素 ( 假定所有元素分数相同 ) 
```

# 09 Hash

-   `Redis hash` 是一个 `string` 类型的 `field`(字段) 和 `value`(值) 的映射表 , hash 特别适合用于存储对象

-   Redis 中每个 hash 可以存储 $2^{32} - 1$ 键值对 ( 40多亿 )  . 

```bash
## 基础操作
HSET key field value [field value ...]           # 设置一个或多个字段的值 ( 存在则覆盖 ) 
HGET key field                                   # 获取指定字段的值
HSETNX key field value                           # 只在字段不存在时设置值 ( 存在则失败 ) 
HMGET key field [field ...]                      # 获取多个指定字段的值
HGETALL key                                       # 获取所有字段和值
HDEL key field [field ...]                        # 删除一个或多个字段
HEXISTS key field                                 # 判断字段是否存在 ( 返回1或0 ) 
HLEN key                                          # 获取字段数量

## 批量获取
HKEYS key           # 获取所有field
HVALS key           # 获取所有value
```

# 10 发布订阅 (Pub/Sub)

**基本要素与解决的问题** : 发布订阅是一种消息通信模式 , 包含发布者 ( 发送消息 ) 、订阅者 ( 接收消息 ) 和频道 ( 消息传递的通道 )  . 它主要用于解决系统间的消息路由和实时解耦问题 , 如实时聊天、通知推送等 . 注意 : Pub/Sub 的消息不会持久化 , 离线的订阅者会丢失在这期间发布的消息 . 

```bash
PUBLISH channel message
# channel : 要发布消息的目标频道名称
# message : 要发布的具体消息内容

SUBSCRIBE channel [channel ...]
# channel : 客户端想要订阅的一个或多个频道名称

UNSUBSCRIBE [channel [channel ...]]
# 退订指定的频道 . 如果不提供频道名 , 则退订所有已订阅的频道

PSUBSCRIBE pattern [pattern ...]
# 按模式订阅频道 ( 支持通配符 , 如 news.* ) 
# pattern : 匹配频道的模式字符串
```

# 11 消息队列 Stream

**基本要素与解决的问题** : Stream 是 Redis 5.0 引入的全新数据结构 , 表示一个**可持久化的追加消息列表** . 它包含消息 ID、字段和值 , 并支持消费者组 ( Consumer Group )  . 主要解决 Pub/Sub 无法持久化消息、无法记录消费进度 ( 游标 ) 、以及缺少多消费者负载均衡的问题 , 非常适合用作高可靠的消息队列 . 

```bash
XADD key [NOMKSTREAM] [MAXLEN|MINID [=|~] threshold [LIMIT count]] *|ID field value [field value ...]
# key : 流的名称
# no-make-stream 如果这个stream不存在 则不会自动创建 返回 nil
# MAXLEN | MINID 
	# 修剪策略 二选一 防止stream无限增长
		# MAXLEN threshold : 按长度修剪 消息数量不超过阈值
		# MINID threshold : 按ID修剪 移除所有小于阈值的id
	# =/~ 修剪精度 默认是精确修剪(=) 可以使用模糊修剪来保证速度
	# LIMIT count : 限制修剪工作量
		# 作用 : 这个选项通常与近似修剪 ~ 一起使用 , 用于限制一次修剪操作中最多移除的消息数量 ( count )  . 
		# 这可以防止一次修剪操作耗时过长 , 导致 Redis 阻塞 . 
		# 如果不指定 , Redis 会有一个默认的限制值 . 
# ID : 消息的唯一标识符 , 通常用 * 让 Redis 自动生成 ( 基于时间戳 ) 
# field value : 消息中包含的键值对内容

XRANGE key start end [COUNT count]
# start / end : 需查询消息的 ID 范围 , - 表示最小 ID , + 表示最大 ID
# COUNT count : 限制返回的消息数量

XREAD [COUNT count] [BLOCK milliseconds] STREAMS key [key ...] id [id ...]
# BLOCK milliseconds : 阻塞等待的毫秒数 , 0 表示一直阻塞
# STREAMS : 后接需要读取的流名称列表和对应的起始消息 ID ( $ 表示只接收最新的消息 ) 

XGROUP CREATE key groupname id|$ [MKSTREAM]
# groupname : 消费者组的名称
# id|$ : 开始消费的消息 ID , $ 表示从尾部 ( 最新 ) 开始消费
# MKSTREAM : 如果流不存在 , 自动创建

XREADGROUP GROUP group consumer [COUNT count] [BLOCK milliseconds] [NOACK] STREAMS key [key ...] id [id ...]
# group : 消费者组的名称
# consumer : 当前消费者的名称
# STREAMS ... : 作用同 XREAD , 这里 ID 通常用 > , 表示读取从未派发给其他消费者的最新消息
```

# 12 地理空间 Geospatial

**基本要素与解决的问题** : Geospatial 提供了一种存储地理位置 ( 经度和纬度 ) 的方法 , 底层通过 Sorted Set 实现 . 主要解决 LBS ( Location-Based Services , 基于位置的服务 ) 场景中的地理位置推算和检索问题 , 如计算两点间距离、查找附近的商家或用户 . 

```bash
GEOADD key [NX|XX] [CH] longitude latitude member [longitude latitude member ...]
# key : 地理空间集合的键
# longitude : 经度 ( 有效范围 : -180 到 180 度 ) 
# latitude : 纬度 ( 有效范围 : -85.05112878 到 85.05112878 度 ) 
# member : 该位置对应的元素/对象名称

GEOPOS key member [member ...]
# member : 要查询坐标的元素名称 , 返回这些元素的经纬度

GEODIST key member1 member2 [m|km|ft|mi]
# member1/member2 : 要计算距离的两个元素
# [m|km|ft|mi] : 指定返回距离的单位 , 分别是米(m)、千米(km)、英尺(ft)、英里(mi) , 默认是米

GEOSEARCH key [FROMMEMBER member] [FROMLONLAT longitude latitude] [BYRADIUS radius m|km|ft|mi] [BYBOX width height m|km|ft|mi] [ASC|DESC] [COUNT count]
# FROMMEMBER/FROMLONLAT : 指定搜索的中心点 , 可以是已有的元素或者具体的经纬度
# BYRADIUS/BYBOX : 指定搜索范围是圆形区域或矩形区域
# ASC|DESC : 按距离中心点的远近排序返回结果
```

# 13 HyperLogLog

**基本要素与解决的问题** : HyperLogLog 是一种概率数据结构 , 用于估算集合中的唯一元素数量 ( 基数计数 )  . 它的核心优势是在统计海量数据时 , 只需要固定且极小的内存空间 ( 约 12KB )  , 而标准误差率仅约为 0.81% . 它完美取代了使用 Set 结构统计网站日活跃用户 ( UV ) 、独立 IP 数量时造成的巨大内存开销问题 . 

```bash
PFADD key element [element ...]
# key : HyperLogLog 结构的键
# element : 要添加的一个或多个元素

PFCOUNT key [key ...]
# key : 要统计基数 ( 估算的唯一元素数量 ) 的键 , 可以指定多个键求它们的并集基数

PFMERGE destkey sourcekey [sourcekey ...]
# destkey : 合并后保存目标结果的键
# sourcekey : 需要被合并的一个或多个来源 HyperLogLog 键
```

# 14 Bitmap (位图)

**基本要素与解决的问题** : Bitmap ( 位图 ) 不是一种独立的数据结构 , 而是建立在 String 上的按位 ( bit ) 操作接口 . 由于字符串最大为 512MB , 因此一个位图可以存储多达 $2^{32}$ 个不同的数据状态 . 它通过极高的空间利用率 , 解决大量只包含 0 和 1 ( 布尔值 ) 状态的记录问题 , 例如用户的每日签到、活跃记录、用户在线状态追踪等 . 

```bash
SETBIT key offset value
# key : 位图的键
# offset : 位的偏移量 ( 从 0 开始 ) 
# value : 设置的值 , 只能是 0 或 1

GETBIT key offset
# offset : 要获取数值的位偏移量

BITCOUNT key [start end [BYTE|BIT]]
# start/end : 指定统计范围 ( 默认单位为字节 BYTE , 可显式指定定位为 BIT ) 
# 统计整个位图 ( 或指定范围内 ) 值为 1 的二进制位的数量

BITOP operation destkey key [key ...]
# operation : 进行的位逻辑运算操作 , 支持 AND(与)、OR(或)、XOR(异或)、NOT(非)
# destkey : 保存位运算结果的目标键
# key : 参与运算的一个或多个源位图键
```

# 15 Bitfield 位域

**基本要素与解决的问题** : Bitfield 允许将一个 Redis 字符串看作是由多个被编码为整数的特定位长区域组成的数组 . 不同于 Bitmap 只能操作单个布尔值 ( 1位 )  , Bitfield 可以对指定位置和长度的位段进行按整数的赋值、读取或自增操作 . 它常用于在单个键中高度压缩地存储多个有长度限制的独立小数值参数 , 如游戏中在单个对象内存储等级、拥有的物品数量和经验值 . 

```bash
BITFIELD key [GET type offset] [SET type offset value] [INCRBY type offset increment] [OVERFLOW WRAP|SAT|FAIL]
# key : 位域操作的键
# type : 指定的整数类型和位数 , 例如 i8 表示 8 位有符号整数 , u4 表示 4 位无符号整数
# offset : 数据偏移量 , 可以在前加 # 号表示基于类型的数组索引来进行等长位移
# value / increment : 要设置或增加的具体数值
# OVERFLOW : 定义超出设定数值范围时的溢出控制行为 , WRAP为折返(默认) , SAT为饱和截断 , FAIL为失败拒绝修改
```

# 16 Redis 事务

**基本要素与解决的问题** : Redis 事务允许将一组命令打包 , 并按照串行化顺序一次性执行 . 事务中的命令在执行时不会由于其他客户端的请求被打断 . 不过需注意 Redis 事务不保证原子性的回滚 ( **如果某条指令执行报错 , 其他指令通常仍会执行** )  . 它主要用于将多步操作打包发送以保障操作连贯性和减少网络延时 , 以及结合 WATCH 命令实现乐观锁 , 防范并发修改问题 . 

```bash
MULTI
# 标记一个事务块的开始 , 之后的命令会被放入队列中排队 , 不会立即执行

EXEC
# 按照先后顺序 , 执行事务队列中积累的所有命令 , 并恢复正常连接状态

DISCARD
# 取消当前事务 , 清空由 MULTI 开始积累的事务队列 , 并恢复正常连接状态

WATCH key [key ...]
# key : 要监视的一个或多个键
# 开启乐观锁机制 , 如果被监视的键在 EXEC 执行前被其他客户端修改 , 则事务将会被打断 ( 执行失败并返回空 ) 

UNWATCH
# 取消当前客户端对所有被 WATCH 命令监视的键的监控锁
```

```bash
# 设置初始余额
> SET account:A 1000
OK
> SET account:B 500
OK

# 开启事务
> MULTI
OK

# 事务队列中的命令（还没有真正执行）
> DECRBY account:A 100
QUEUED
> INCRBY account:B 100
QUEUED

# 执行事务
> EXEC
1) (integer) 900    # A账户扣减后的余额
2) (integer) 600    # B账户增加后的余额
```

# 17 持久化 (Persistence)

**基本要素与解决的问题** : 持久化机制将 Redis 在内存中的热点数据不定期或实时地备份到硬盘中 , 主要分为 RDB ( 数据全量快照 ) 和 AOF ( 追加命令日志记录 ) 两种机制 . 它解决了由于服务器进程意外退出、系统崩溃或断电等导致的内存数据丢失问题 , 保证了系统的高可靠性和灾难数据恢复能力 . 

```bash
SAVE
# 强制在主线程同步阻塞地生成 RDB 快照保存到磁盘中 ( 会阻塞所有客户端的读写请求 , 不推荐在生产环境直接使用 ) 

BGSAVE
# 异步在后台 fork 出一个子进程将数据保存到磁盘 , 不会阻塞客户端的正常使用请求 ( 推荐的 RDB 快照方式 ) 

BGREWRITEAOF
# 异步在后台执行 AOF 文件的重写操作 , 合并冗余的操作指令 , 用来压缩 AOF 文件体积并优化将来系统的加载恢复速度

LASTSAVE
# 获取最近一次成功执行 SAVE 或 BGSAVE 操作保存数据到磁盘的时间点 ( 返回 Unix 时间戳形式 ) 
```

# 18 主从复制 (Replication)

**基本要素与解决的问题** : 主从复制允许一个 Redis 实例 ( 即 Master 主节点 ) 将自身的数据状态完美复制到一个或多个其他实例 ( Replica 从节点 ) 上 . 主节点可读可写 , 而从节点通常只接收读请求 . 这种机制解决了单机单点的故障问题 , 实现了数据冗余和热备份；同时能够通过读写分离的架构 , 利用从节点负担海量的读请求进而提升整个系统的吞吐量 . 

```bash
REPLICAOF host port
# host / port : 目标主节点 ( Master ) 的 IP 地址和端口号
# 将当前 Redis 实例转变成指定主节点的从节点 . 如果指定为 REPLICAOF NO ONE , 则主动断开与主节点的同步 , 将自己提升为独立节点/主节点

ROLE
# 返回当前实例本身在复制系统架构中的角色 ( master 还是 replica )  , 以及关于其复制偏移量状态、主从网络连接延迟的信息
```

# 19 哨兵模式 (Sentinel)

**基本要素与解决的问题** : Sentinel ( 哨兵模式 ) 是官方推荐的高可用性 ( HA ) 解决方案 , 由一个或多个 Sentinel 实例组成的独立集群去监控任意多个主从 Redis 架构群 . 它主要解决了在主从模式中 , 主节点发生故障宕机时需要人工手动干预执行故障转移 ( 如晋升新的主节点 ) 的痛点 , 哨兵会自动发现故障、自动从从节点中选举出新的主节点并接管通知客户端的工作 , 实现架构的自我修复与高度可用 . 

```bash
SENTINEL masters
# 获取当前哨兵集群正在监控的所有主节点的信息的状态与参数列表详情

SENTINEL get-master-addr-by-name master-name
# master-name : 在哨兵配置文件中定义的被监控的主节点名称
# 获取当前生效的且正常运行的主节点的 IP 和端口 , 大部分客户端 SDK 使用此命令进行服务动态发现感知

SENTINEL failover master-name
# 强制让哨兵对此特定的主节点执行一次人工干预下的故障转移操作 , 触发新的选主流程 ( 无需等待原主节点真的宕机 ) 
```



































