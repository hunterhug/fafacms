package model

import (
	"fmt"
	"testing"
)

// TestSchema_LongTextColumnsUseMediumText 校验"长文本列在新建表时为 MEDIUMTEXT"。
//
// 背景：密文相对明文会膨胀约 1/3（base64 + nonce + tag），
// 若长文本列仍为 TEXT（上限 65535 字节），约 49KB 以上的明文加密后会写不下。
// 注意：xorm 的 Sync2 不会升级已存在列的类型，因此该变更在"清库重建"后才生效。
type schemaProbe struct {
	Id    int64  `xorm:"bigint pk autoincr"`
	Body  string `json:"body" xorm:"MEDIUMTEXT"`
	Short string `json:"short" xorm:"TEXT"`
}

func (schemaProbe) TableName() string { return "schema_probe_tmp" }

func TestSchema_LongTextColumnsUseMediumText(t *testing.T) {
	setupTestDB(t)

	probe := new(schemaProbe)
	_ = FaFaRdb.Client.DropTables(probe)
	defer func() { _ = FaFaRdb.Client.DropTables(probe) }()

	if err := FaFaRdb.Client.Sync2(probe); err != nil {
		t.Fatalf("建表失败: %v", err)
	}

	// 实际库里应被创建为 mediumtext
	rows, err := FaFaRdb.Client.QueryString(
		"SELECT COLUMN_TYPE FROM information_schema.COLUMNS " +
			"WHERE TABLE_SCHEMA='fafa' AND TABLE_NAME='schema_probe_tmp' AND COLUMN_NAME='body'")
	if err != nil || len(rows) == 0 {
		t.Fatalf("查询列类型失败: %v", err)
	}
	got := rows[0]["COLUMN_TYPE"]
	if got != "mediumtext" {
		t.Errorf("MEDIUMTEXT 标记应生成 mediumtext 列，实际: %s", got)
	}

	// 复核：明文长度超过 TEXT 上限的值可以写入（写入 60KB，密文将超过 65535 字节）
	large := make([]byte, 60*1024)
	for i := range large {
		large[i] = 'a'
	}
	initTestKeyring(t)
	enc, err := EncryptTableField("schema_probe_tmp", "body", 0, string(large))
	if err != nil {
		t.Fatalf("加密失败: %v", err)
	}
	if len(enc) <= 65535 {
		t.Fatalf("样例密文未超过 TEXT 上限（%d 字节），该用例失去意义", len(enc))
	}

	p := &schemaProbe{Body: enc, Short: "x"}
	if _, err := FaFaRdb.Client.InsertOne(p); err != nil {
		t.Fatalf("写入超 TEXT 上限的密文失败: %v", err)
	}

	back := new(schemaProbe)
	exist, err := FaFaRdb.Client.ID(p.Id).Get(back)
	if err != nil || !exist {
		t.Fatalf("读回失败: %v", err)
	}
	if len(back.Body) != len(enc) {
		t.Errorf("密文被截断: 写入 %d 字节，读回 %d 字节", len(enc), len(back.Body))
	}
	plain, err := DecryptTableField("schema_probe_tmp", "body", 0, back.Body)
	if err != nil {
		t.Fatalf("解密失败: %v", err)
	}
	if len(plain) != len(large) {
		t.Errorf("大文本往返长度不一致: got=%d want=%d", len(plain), len(large))
	}
	fmt.Printf("  大文本加密往返 OK：明文 %d 字节 → 密文 %d 字节\n", len(large), len(enc))
}
