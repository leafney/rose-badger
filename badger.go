package rbadger

import (
	"bytes"
	"encoding/gob"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// BadgerDB 结构体封装了 badger 的基本操作
type BadgerDB struct {
	db *badger.DB
	mu sync.Mutex // 添加互斥锁
}

// NewBadgerDB 创建一个新的 BadgerDB 实例
// 示例：
//
//	db, err := NewBadgerDB("./data")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
func NewBadgerDB(dbPath string) (*BadgerDB, error) {
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerDB{db: db, mu: sync.Mutex{}}, nil
}

// NewBadgerDBWithOptions 创建一个带自定义选项的 BadgerDB 实例
// 示例：
//
//	opts := badger.DefaultOptions("./data")
//	opts.SyncWrites = true
//	db, err := NewBadgerDBWithOptions(opts)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
func NewBadgerDBWithOptions(opts badger.Options) (*BadgerDB, error) {
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &BadgerDB{db: db, mu: sync.Mutex{}}, nil
}

// Get 获取指定key的值
// 示例：
//
//	value, err := db.Get("key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("值: %s\n", value)
func (b *BadgerDB) Get(key string) ([]byte, error) {
	var valCopy []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		err = item.Value(func(val []byte) error {
			// 复制值，因为在事务外部使用值需要复制
			valCopy = append([]byte{}, val...)
			return nil
		})
		return err
	})
	return valCopy, err
}

// GetS 获取指定key的字符串值
// 示例：
//
//	value, err := db.GetS("key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("字符串值: %s\n", value)
func (b *BadgerDB) GetS(key string) (string, error) {
	value, err := b.Get(key)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

// Set 设置key的值
// 示例：
//
//	err := db.Set("key", []byte("value"))
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) Set(key string, value []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), value)
	})
}

// SetS 设置key的字符串值
// 示例：
//
//	err := db.SetS("key", "value")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) SetS(key string, value string) error {
	return b.Set(key, []byte(value))
}

// Exists 检查key是否存在
// 示例：
//
//	if db.Exists("key") {
//	    fmt.Println("key存在")
//	} else {
//	    fmt.Println("key不存在")
//	}
func (b *BadgerDB) Exists(key string) bool {
	err := b.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		return err
	})
	return err == nil
}

// Del 删除指定的key
// 示例：
//
//	err := db.Del("key")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) Del(key string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
}

// Close 关闭数据库连接
// 示例：
//
//	defer db.Close()
func (b *BadgerDB) Close() error {
	return b.db.Close()
}

// *******************

// CacheType 定义缓存数据结构
type CacheType struct {
	Data   []byte
	Expire int64 // Unix timestamp 表示过期时间点
}

// XGet 获取带过期时间的缓存数据
// 当数据过期时会自动删除并返回nil
// 示例：
//
//	value, err := db.XGet("key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if value == nil {
//	    fmt.Println("key不存在或已过期")
//	} else {
//	    fmt.Printf("值: %s\n", value)
//	}
func (b *BadgerDB) XGet(key string) ([]byte, error) {
	var valCopy []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			var cache CacheType
			decoder := gob.NewDecoder(bytes.NewReader(val))
			if err := decoder.Decode(&cache); err != nil {
				return err
			}

			// 检查是否过期
			if cache.Expire > 0 && cache.Expire <= time.Now().Unix() {
				// 过期了，但在只读事务中无法删除，所以在外部删除
				return badger.ErrKeyNotFound
			}

			// 复制值，因为在事务外部使用值需要复制
			valCopy = append([]byte{}, cache.Data...)
			return nil
		})
	})

	if err == badger.ErrKeyNotFound {
		// 如果是过期或不存在，尝试删除（如果是过期的情况）
		b.Del(key)
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return valCopy, nil
}

// XGetS 获取带过期时间的字符串数据
// 示例：
//
//	value, err := db.XGetS("key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if value == "" {
//	    fmt.Println("key不存在或已过期")
//	} else {
//	    fmt.Printf("字符串值: %s\n", value)
//	}
func (b *BadgerDB) XGetS(key string) (string, error) {
	data, err := b.XGet(key)
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", nil
	}
	return string(data), nil
}

// XSet 设置带过期时间的缓存数据
// 示例：
//
//	err := db.XSet("key", []byte("value"))
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) XSet(key string, value []byte) error {
	cache := CacheType{
		Data:   value,
		Expire: 0,
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(cache); err != nil {
		return err
	}

	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), buf.Bytes())
	})
}

// XSetS 设置带过期时间的字符串数据
// 示例：
//
//	err := db.XSetS("key", "value")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) XSetS(key string, value string) error {
	return b.XSet(key, []byte(value))
}

// XSetEx 设置带过期时间的缓存数据 （使用 time.Duration）
// 示例：
//
//	err := db.XSetEx("key", []byte("value"), time.Hour)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) XSetEx(key string, value []byte, expires time.Duration) error {
	cache := CacheType{
		Data:   value,
		Expire: time.Now().Add(expires).Unix(),
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(cache); err != nil {
		return err
	}

	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), buf.Bytes())
	})
}

// XSetExS 设置带过期时间的字符串数据 （使用 time.Duration）
// 示例：
//
//	err := db.XSetExS("key", "value", time.Hour)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) XSetExS(key string, value string, expires time.Duration) error {
	return b.XSetEx(key, []byte(value), expires)
}

// XSetExSec 设置带过期时间的缓存数据（使用秒数）
// 示例：
//
//	err := db.XSetExSec("key", []byte("value"), 3600)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) XSetExSec(key string, value []byte, seconds int64) error {
	return b.XSetEx(key, value, time.Duration(seconds)*time.Second)
}

// XSetExSecS 设置带过期时间的字符串数据（使用秒数）
// 示例：
//
//	err := db.XSetExSecS("key", "value", 3600)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) XSetExSecS(key string, value string, seconds int64) error {
	return b.XSetExSec(key, []byte(value), seconds)
}

// XTTL 返回key的剩余生存时间(秒)
// 返回值说明：
//
//	-2: key不存在（包括已过期的情况）
//	-1: key存在但未设置过期时间
//	>=0: key的剩余生存时间(秒)
//
// 示例：
//
//	ttl, err := db.XTTL("key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	switch ttl {
//	case -2:
//	    fmt.Println("key不存在")
//	case -1:
//	    fmt.Println("key未设置过期时间")
//	default:
//	    fmt.Printf("剩余生存时间: %d秒\n", ttl)
//	}
func (b *BadgerDB) XTTL(key string) (int64, error) {
	var ttl int64 = -2 // 默认为不存在

	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			var cache CacheType
			decoder := gob.NewDecoder(bytes.NewReader(val))
			if err := decoder.Decode(&cache); err != nil {
				return err
			}

			// key 存在但未设置过期时间
			if cache.Expire == 0 {
				ttl = -1
				return nil
			}

			// 计算剩余生存时间
			remaining := cache.Expire - time.Now().Unix()
			if remaining <= 0 {
				// 已过期，但在只读事务中无法删除
				ttl = -2
				return badger.ErrKeyNotFound
			}

			ttl = remaining
			return nil
		})
	})

	if err == badger.ErrKeyNotFound {
		// 如果是过期或不存在，尝试删除（如果是过期的情况）
		b.Del(key)
		return -2, nil
	}

	if err != nil {
		return -2, err
	}

	return ttl, nil
}

// XExpire 设置key的过期时间
// 示例：
//
//	err := db.XExpire("key", time.Hour)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) XExpire(key string, expires time.Duration) error {
	return b.XExpireAt(key, time.Now().Add(expires))
}

// XExpireSec 设置key的过期时间(秒)
// 示例：
//
//	err := db.XExpireSec("key", 3600)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) XExpireSec(key string, seconds int64) error {
	return b.XExpire(key, time.Duration(seconds)*time.Second)
}

// XExpireAt 设置key的过期时间点
// 该方法是并发安全的
// 示例：
//
//	err := db.XExpireAt("key", time.Now().Add(time.Hour))
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) XExpireAt(key string, tm time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var cache CacheType

	// 先获取当前值
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			decoder := gob.NewDecoder(bytes.NewReader(val))
			return decoder.Decode(&cache)
		})
	})

	if err != nil {
		return err
	}

	// 设置新的过期时间
	cache.Expire = tm.Unix()

	// 保存回数据库
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(cache); err != nil {
		return err
	}

	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), buf.Bytes())
	})
}

// XIncrBy 将key中存储的数字值增加指定的值
// 该方法是并发安全的
// 示例：
//
//	value, err := db.XIncrBy("counter", 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("新值: %d\n", value)
func (b *BadgerDB) XIncrBy(key string, increment int64) (int64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var cache CacheType
	var value int64

	// 先获取当前值
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			// key不存在，初始化为0
			cache = CacheType{
				Expire: 0,
			}
			value = 0
			return nil
		}
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			decoder := gob.NewDecoder(bytes.NewReader(val))
			if err := decoder.Decode(&cache); err != nil {
				return err
			}

			// 解析当前值
			value, err = strconv.ParseInt(string(cache.Data), 10, 64)
			return err
		})
	})

	if err != nil {
		return 0, err
	}

	// 增加值
	value += increment
	cache.Data = []byte(strconv.FormatInt(value, 10))

	// 保存新值
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(cache); err != nil {
		return 0, err
	}

	err = b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), buf.Bytes())
	})

	if err != nil {
		return 0, err
	}

	return value, nil
}

// XIncr 将key中存储的数字值加1
// 该方法是并发安全的
// 示例：
//
//	value, err := db.XIncr("counter")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("新值: %d\n", value)
func (b *BadgerDB) XIncr(key string) (int64, error) {
	return b.XIncrBy(key, 1)
}

// XDecr 将key中存储的数字值减1
// 该方法是并发安全的
// 示例：
//
//	value, err := db.XDecr("counter")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("新值: %d\n", value)
func (b *BadgerDB) XDecr(key string) (int64, error) {
	return b.XDecrBy(key, 1)
}

// XDecrBy 将key中存储的数字值减少指定的值
// 该方法是并发安全的
// 示例：
//
//	value, err := db.XDecrBy("counter", 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("新值: %d\n", value)
func (b *BadgerDB) XDecrBy(key string, decrement int64) (int64, error) {
	return b.XIncrBy(key, -decrement)
}

// RunGC 运行垃圾回收以清理过期的值日志
// 示例：
//
//	err := db.RunGC(0.5) // 丢弃比例为0.5
//	if err != nil {
//	    log.Fatal(err)
//	}
func (b *BadgerDB) RunGC(discardRatio float64) error {
	return b.db.RunValueLogGC(discardRatio)
}
