package model

import (
	"fmt"
	"github.com/hunterhug/fafacms/core/util/rdb"
	"os"
	"testing"
)

// setupTestDB 连接测试数据库；默认跳过，需显式设置 FAFA_TEST_DB=1 才运行
// （这些是依赖真实 MySQL 的集成测试，且会写库，不应在普通 go test 里默认执行）
func setupTestDB(t *testing.T) {
	if os.Getenv("FAFA_TEST_DB") != "1" {
		t.Skip("skipping DB integration test; set FAFA_TEST_DB=1 to run against a test database")
	}
	host := os.Getenv("FAFA_TEST_DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("FAFA_TEST_DB_PORT")
	if port == "" {
		port = "13307"
	}
	c := rdb.MyDbConfig{}
	c.User = "root"
	c.Host = host
	c.Port = port
	c.Pass = "123456789"
	c.DriverName = "mysql"
	c.Prefix = "fafacms_"
	c.Name = "fafa"
	c.Debug = false
	db, err := rdb.NewDb(c)
	if err != nil {
		t.Fatalf("connect db: %v", err)
	}

	FaFaRdb = db

	// 确保测试库结构与模型一致：新增列/加宽列（如密文列、*_bidx 盲索引列）在集成测试前生效。
	// 注意：若库里已有历史数据，唯一索引可能因历史空值而创建失败，此时只记录告警——
	// 列本身仍会被补上（正式环境走"清库重建"，不存在该问题）。
	if err := db.Client.Sync2(new(User), new(Message), new(GlobalMessage),
		new(Content), new(ContentHistory), new(ContentBad),
		new(Comment), new(CommentBad), new(CommentCool),
		new(ContentNode), new(Group), new(Resource), new(Relation),
		new(UserBad), new(File)); err != nil {
		t.Logf("同步测试库表结构告警（可忽略索引类失败）: %v", err)
	}
}

func TestContentNode_CheckSeoValid(t *testing.T) {
	setupTestDB(t)
	n := new(ContentNode)
	n.Seo = "sss"
	n.UserId = 2
	n.Level = 1

	exist, err := n.CheckSeoValid()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(exist)

}

func TestContentNode_InsertOne(t *testing.T) {
	setupTestDB(t)

	n := new(ContentNode)
	n.UserId = 1
	n.Seo = ""
	n.Status = 1
	n.SortNum = 3
	c, err := FaFaRdb.Client.MustCols("seo", "status").Omit("user_id").Where("user_id=?", n.UserId).Update(n)
	fmt.Println(c, err)

}

func TestGroup_Delete(t *testing.T) {
	setupTestDB(t)

	g := new(Group)
	g.Id = 1000
	err := g.Delete()
	fmt.Println(err)
}

func TestResource_Get(t *testing.T) {
	setupTestDB(t)

	g := new(Resource)
	g.Id = 1000
	err := g.Get()
	fmt.Println(err)
}

func TestContent_Get(t *testing.T) {
	setupTestDB(t)

	c := new(Content)
	c.Id = 2
	c.UserId = 2
	c.Status = 1
	c.PreFlush = 1
	c.Get()
}
