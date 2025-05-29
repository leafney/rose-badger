package rbadger

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// BadgerDBExample 展示了如何使用BadgerDB封装
func TestBadgerDBExample(t *testing.T) {
	// 创建一个新的BadgerDB实例
	db, err := NewBadgerDB("./badger")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 基本的Set和Get操作
	fmt.Println("=== 基本的Set和Get操作 ===")
	err = db.SetS("key1", "value1")
	if err != nil {
		log.Fatal(err)
	}

	val, err := db.GetS("key1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("key1的值: %s\n", val)

	// 检查key是否存在
	fmt.Println("\n=== 检查key是否存在 ===")
	if db.Exists("key1") {
		fmt.Println("key1存在")
	} else {
		fmt.Println("key1不存在")
	}

	if db.Exists("key2") {
		fmt.Println("key2存在")
	} else {
		fmt.Println("key2不存在")
	}

	// 带过期时间的操作
	fmt.Println("\n=== 带过期时间的操作 ===")
	err = db.XSetExSecS("key2", "value2", 5) // 5秒后过期
	if err != nil {
		log.Fatal(err)
	}

	val, err = db.XGetS("key2")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("key2的值: %s\n", val)

	// 获取TTL
	ttl, err := db.XTTL("key2")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("key2的TTL: %d秒\n", ttl)

	// 等待3秒
	fmt.Println("等待3秒...")
	time.Sleep(3 * time.Second)

	// 再次获取TTL
	ttl, err = db.XTTL("key2")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("3秒后key2的TTL: %d秒\n", ttl)

	// 等待3秒（使key2过期）
	fmt.Println("再等待3秒（使key2过期）...")
	time.Sleep(3 * time.Second)

	// 尝试获取已过期的key
	val, err = db.XGetS("key2")
	if err != nil {
		log.Fatal(err)
	}
	if val == "" {
		fmt.Println("key2已过期")
	} else {
		fmt.Printf("key2的值: %s\n", val)
	}

	// 计数器操作
	fmt.Println("\n=== 计数器操作 ===")
	count, err := db.XIncr("counter")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("计数器初始值: %d\n", count)

	count, err = db.XIncrBy("counter", 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("增加10后: %d\n", count)

	count, err = db.XDecrBy("counter", 5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("减少5后: %d\n", count)

	// 删除操作
	fmt.Println("\n=== 删除操作 ===")
	err = db.Del("key1")
	if err != nil {
		log.Fatal(err)
	}

	if db.Exists("key1") {
		fmt.Println("key1存在")
	} else {
		fmt.Println("key1已被删除")
	}

	// 运行垃圾回收
	fmt.Println("\n=== 运行垃圾回收 ===")
	err = db.RunGC(0.5)
	if err != nil {
		log.Printf("垃圾回收错误: %v\n", err)
	} else {
		fmt.Println("垃圾回收完成")
	}
	// 注意：RunGC 方法已经处理了 "没有清理任何数据" 的提示性信息，
	// 如果返回 nil，表示垃圾回收成功或没有需要清理的数据
}

func TestInMemory(t *testing.T) {
	opts := badger.DefaultOptions("").WithInMemory(true).WithLogger(nil)
	opts.WithIndexCacheSize(100 << 20) // 100 MB for index cache

	db, err := NewBadgerDBWithOptions(opts)
	if err != nil {
		log.Fatal(err)
	}

	db.SetS("key", "value")
	t.Log(db.GetS("key"))
}
