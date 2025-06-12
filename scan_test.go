package rbadger

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// TestKeys 测试Keys方法
func TestKeys(t *testing.T) {
	// 创建临时数据库
	dbPath := "./test_scan_db"
	defer os.RemoveAll(dbPath)

	db, err := NewBadgerDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 设置一些测试数据
	testData := map[string]string{
		"user:1": "alice",
		"user:2": "bob",
		"user:3": "charlie",
		"order:1": "order1",
		"order:2": "order2",
		"product:1": "product1",
	}

	for key, value := range testData {
		err := db.SetS(key, value)
		if err != nil {
			t.Fatal(err)
		}
	}

	// 测试扫描user前缀的key
	userKeys, err := db.FindKeys("user:")
	if err != nil {
		t.Fatal(err)
	}

	if len(userKeys) != 3 {
		t.Errorf("期望找到3个user key，实际找到%d个", len(userKeys))
	}

	fmt.Printf("找到的user keys: %v\n", userKeys)

	// 测试扫描order前缀的key
	orderKeys, err := db.FindKeys("order:")
	if err != nil {
		t.Fatal(err)
	}

	if len(orderKeys) != 2 {
		t.Errorf("期望找到2个order key，实际找到%d个", len(orderKeys))
	}

	fmt.Printf("找到的order keys: %v\n", orderKeys)
}

// TestXKeys 测试XKeys方法
func TestXKeys(t *testing.T) {
	// 创建临时数据库
	dbPath := "./test_xscan_db"
	defer os.RemoveAll(dbPath)

	db, err := NewBadgerDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 设置一些带过期时间的测试数据
	err = db.XSetExS("cache:valid1", "value1", time.Hour) // 1小时后过期
	if err != nil {
		t.Fatal(err)
	}

	err = db.XSetExS("cache:valid2", "value2", time.Hour) // 1小时后过期
	if err != nil {
		t.Fatal(err)
	}

	err = db.XSetExSecS("cache:expired", "value3", 1) // 1秒后过期
	if err != nil {
		t.Fatal(err)
	}

	err = db.XSetS("cache:permanent", "value4") // 永不过期
	if err != nil {
		t.Fatal(err)
	}

	// 等待过期key过期
	time.Sleep(2 * time.Second)

	// 测试扫描cache前缀的未过期key
	validKeys, err := db.FindXKeys("cache:")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("找到的未过期cache keys: %v\n", validKeys)

	// 应该找到3个未过期的key（valid1, valid2, permanent）
	if len(validKeys) != 3 {
		t.Errorf("期望找到3个未过期的cache key，实际找到%d个", len(validKeys))
	}

	// 验证过期的key是否被自动删除
	expiredValue, err := db.XGetS("cache:expired")
	if err != nil {
		t.Fatal(err)
	}
	if expiredValue != "" {
		t.Error("过期的key应该已被删除")
	}
}

// TestMixedKeys 测试混合存储的key扫描
func TestMixedKeys(t *testing.T) {
	// 创建临时数据库
	dbPath := "./test_mixed_db"
	defer os.RemoveAll(dbPath)

	db, err := NewBadgerDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// 设置混合类型的数据
	err = db.SetS("test:normal", "normal_value") // 普通存储
	if err != nil {
		t.Fatal(err)
	}

	err = db.XSetS("test:cache", "cache_value") // 缓存存储（永不过期）
	if err != nil {
		t.Fatal(err)
	}

	err = db.XSetExS("test:temp", "temp_value", time.Hour) // 缓存存储（1小时后过期）
	if err != nil {
		t.Fatal(err)
	}

	// 使用FindKeys方法应该找到所有key
	allKeys, err := db.FindKeys("test:")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("所有test keys: %v\n", allKeys)
	if len(allKeys) != 3 {
		t.Errorf("期望找到3个test key，实际找到%d个", len(allKeys))
	}

	// 使用FindXKeys方法应该只找到缓存类型的key
	cacheKeys, err := db.FindXKeys("test:")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("缓存类型的test keys: %v\n", cacheKeys)
	if len(cacheKeys) != 2 {
		t.Errorf("期望找到2个缓存类型的test key，实际找到%d个", len(cacheKeys))
	}
}