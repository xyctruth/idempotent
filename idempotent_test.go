package idempotent

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/xyctruth/idempotent/drive/drive_gorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"sync"
	"testing"
	"time"
)

const DefaultMysqlConnStr = "root:@tcp(localhost:3306)/test?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4"

func TestAcquire(t *testing.T) {
	mysqlConnStr := os.Getenv("MYSQL_CONN_STR")
	if mysqlConnStr == "" {
		mysqlConnStr = DefaultMysqlConnStr
	}
	db := NewDB(mysqlConnStr)
	i, err := New(drive_gorm.New(db))
	if err != nil {
		panic(err)
	}

	tests := []struct {
		name  string
		key   string
		ttl   time.Duration
		want  bool
		sleep time.Duration
	}{
		{name: "1 success", key: "TestAcquire", ttl: time.Second * 1, want: true, sleep: 0},
		{name: "2 repeated failure", key: "TestAcquire", ttl: time.Second * 1, want: false, sleep: 0},
		{name: "3 overdue success", key: "TestAcquire", ttl: time.Second * 1, want: true, sleep: time.Second * 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bool
			time.Sleep(tt.sleep)
			_ = db.Transaction(func(tx *gorm.DB) error {
				got = i.Acquire(tt.key, drive_gorm.New(tx), WithTTL(tt.ttl))
				return nil
			})
			require.Equal(t, tt.want, got)
		})
	}
}

func TestConcurrentAcquire(t *testing.T) {
	mysqlConnStr := os.Getenv("MYSQL_CONN_STR")
	if mysqlConnStr == "" {
		mysqlConnStr = DefaultMysqlConnStr
	}
	db := NewDB(mysqlConnStr)
	client, err := New(drive_gorm.New(db))
	if err != nil {
		panic(err)
	}
	wg := sync.WaitGroup{}
	count := make(map[string]int)
	var l sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		key := fmt.Sprintf("key:%d", i)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = db.Transaction(func(tx *gorm.DB) error {
						if client.Acquire(key, drive_gorm.New(tx), WithTTL(time.Minute)) {
							l.Lock()
							defer l.Unlock()
							count[key] = count[key] + 1
						}
						return nil
					})
				}()
			}
		}()
	}
	wg.Wait()
	for _, v := range count {
		require.Equal(t, 1, v)
	}
}

func NewDB(dns string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dns), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}
