# rose-badger

这个项目是对 [hypermodeinc/badger](https://github.com/hypermodeinc/badger) 库的封装，提供了简单易用的 API 接口，使得在项目中可以方便地使用 BadgerDB 作为底层存储。

## 特性

- 简单易用的 API
- 支持基本的键值对操作：Get、Set、Del、Exists
- 支持带过期时间的缓存操作
- 支持原子计数器操作
- 支持垃圾回收
- 线程安全

## 安装

```bash
go get github.com/leafney/rose-badger
```

依赖包：

```bash
go get github.com/dgraph-io/badger/v4
```

## 使用示例

### 基本操作

```go
// 创建一个新的BadgerDB实例
db, err := NewBadgerDB("./data")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// 设置值
err = db.SetS("key", "value")
if err != nil {
    log.Fatal(err)
}

// 获取值
val, err := db.GetS("key")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("值: %s\n", val)

// 检查键是否存在
if db.Exists("key") {
    fmt.Println("键存在")
} else {
    fmt.Println("键不存在")
}

// 删除键
err = db.Del("key")
if err != nil {
    log.Fatal(err)
}
```

### 带过期时间的操作

```go
// 设置带过期时间的值（1小时后过期）
err = db.XSetExS("key", "value", time.Hour)
if err != nil {
    log.Fatal(err)
}

// 设置带过期时间的值（3600秒后过期）
err = db.XSetExSecS("key", "value", 3600)
if err != nil {
    log.Fatal(err)
}

// 获取带过期时间的值
val, err := db.XGetS("key")
if err != nil {
    log.Fatal(err)
}
if val == "" {
    fmt.Println("键不存在或已过期")
} else {
    fmt.Printf("值: %s\n", val)
}

// 获取剩余生存时间（TTL）
ttl, err := db.XTTL("key")
if err != nil {
    log.Fatal(err)
}
switch ttl {
case -2:
    fmt.Println("键不存在")
case -1:
    fmt.Println("键未设置过期时间")
default:
    fmt.Printf("剩余生存时间: %d秒\n", ttl)
}

// 设置已存在键的过期时间
err = db.XExpire("key", time.Hour)
if err != nil {
    log.Fatal(err)
}

// 设置已存在键的过期时间（秒）
err = db.XExpireSec("key", 3600)
if err != nil {
    log.Fatal(err)
}

// 设置已存在键的过期时间点
err = db.XExpireAt("key", time.Now().Add(time.Hour))
if err != nil {
    log.Fatal(err)
}
```

### 计数器操作

```go
// 将计数器加1
count, err := db.XIncr("counter")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("新值: %d\n", count)

// 将计数器加10
count, err = db.XIncrBy("counter", 10)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("新值: %d\n", count)

// 将计数器减1
count, err = db.XDecr("counter")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("新值: %d\n", count)

// 将计数器减5
count, err = db.XDecrBy("counter", 5)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("新值: %d\n", count)
```

### 扫描操作

```go
// 扫描所有以"user:"为前缀的key
keys, err := db.FindKeys("user:")
if err != nil {
    log.Fatal(err)
}
for _, key := range keys {
    fmt.Printf("找到key: %s\n", key)
}

// 扫描所有以"cache:"为前缀且未过期的key
validKeys, err := db.FindXKeys("cache:")
if err != nil {
    log.Fatal(err)
}
for _, key := range validKeys {
    fmt.Printf("找到未过期的key: %s\n", key)
}
```

### 垃圾回收

```go
// 运行垃圾回收以清理过期的值日志
err = db.RunGC(0.5) // 丢弃比例为0.5
if err != nil {
    log.Fatal(err)
}
```

## API 参考

### 基本操作

- `NewBadgerDB(dbPath string) (*BadgerDB, error)` - 创建一个新的 BadgerDB 实例
- `NewBadgerDBWithOptions(opts badger.Options) (*BadgerDB, error)` - 创建一个带自定义选项的 BadgerDB 实例
- `Get(key string) ([]byte, error)` - 获取指定键的值
- `GetS(key string) (string, error)` - 获取指定键的字符串值
- `Set(key string, value []byte) error` - 设置键的值
- `SetS(key string, value string) error` - 设置键的字符串值
- `Exists(key string) bool` - 检查键是否存在
- `Del(key string) error` - 删除指定的键
- `Close() error` - 关闭数据库连接

### 带过期时间的操作

- `XGet(key string) ([]byte, error)` - 获取带过期时间的缓存数据
- `XGetS(key string) (string, error)` - 获取带过期时间的字符串数据
- `XSet(key string, value []byte) error` - 设置带过期时间的缓存数据
- `XSetS(key string, value string) error` - 设置带过期时间的字符串数据
- `XSetEx(key string, value []byte, expires time.Duration) error` - 设置带过期时间的缓存数据
- `XSetExS(key string, value string, expires time.Duration) error` - 设置带过期时间的字符串数据
- `XSetExSec(key string, value []byte, seconds int64) error` - 设置带过期时间的缓存数据（秒）
- `XSetExSecS(key string, value string, seconds int64) error` - 设置带过期时间的字符串数据（秒）
- `XTTL(key string) (int64, error)` - 返回键的剩余生存时间（秒）
- `XExpire(key string, expires time.Duration) error` - 设置键的过期时间
- `XExpireSec(key string, seconds int64) error` - 设置键的过期时间（秒）
- `XExpireAt(key string, tm time.Time) error` - 设置键的过期时间点

### 计数器操作

- `XIncrBy(key string, increment int64) (int64, error)` - 将键中存储的数字值增加指定的值
- `XIncr(key string) (int64, error)` - 将键中存储的数字值加1
- `XDecr(key string) (int64, error)` - 将键中存储的数字值减1
- `XDecrBy(key string, decrement int64) (int64, error)` - 将键中存储的数字值减少指定的值

### 扫描操作

- `FindKeys(prefix string) ([]string, error)` - 扫描所有匹配指定前缀的key列表
- `FindXKeys(prefix string) ([]string, error)` - 扫描所有匹配指定前缀且未过期的key列表

### 其他操作

- `RunGC(discardRatio float64) error` - 运行垃圾回收以清理过期的值日志

## 实现说明

- 使用 `badger.DB` 作为底层存储
- 使用 `gob` 编码和解码 `CacheType` 结构体来存储数据和过期时间
- 使用互斥锁 `sync.Mutex` 确保并发安全
- 过期时间的处理：在读取时检查过期时间，如果已过期则删除并返回 nil
- 计数器操作：使用 `strconv` 包进行字符串和整数之间的转换



## 注意事项

- 在使用完数据库后，务必调用 `Close()` 方法关闭数据库连接
- 对于大量写入操作，可以考虑定期调用 `RunGC()` 方法进行垃圾回收
- 对于需要频繁更新的键，可以使用计数器操作来避免读取-修改-写入的竞争条件