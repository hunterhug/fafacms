package model

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// newBidxTestUser 创建一个带全部可查询字段的临时用户。
func newBidxTestUser(t *testing.T, suffix string) (*User, func()) {
	t.Helper()
	u := &User{
		Name:     "__bidx_name_" + suffix,
		NickName: "__bidx_nick_" + suffix,
		Email:    "__bidx_" + suffix + "@Example.com", // 故意含大写，验证规范化
		Password: "test123456",
		Status:   1,
		WeChat:   "wx_" + suffix,
		WeiBo:    "wb_" + suffix,
		Github:   "gh_" + suffix,
		QQ:       "1000" + suffix,
	}
	if err := u.InsertOne(); err != nil {
		t.Fatalf("插入用户失败: %v", err)
	}
	return u, func() {
		_, _ = FaFaRdb.Client.Exec(fmt.Sprintf("DELETE FROM fafacms_user WHERE id=%d", u.Id))
	}
}

func TestUser_SearchableFieldsEncryptedWithBlindIndex(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	u, cleanup := newBidxTestUser(t, suffix)
	defer cleanup()

	rows, err := FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT name, nick_name, email, we_chat, wei_bo, github, q_q, "+
			"name_bidx, nick_name_bidx, email_bidx, we_chat_bidx, wei_bo_bidx, github_bidx, q_q_bidx "+
			"FROM fafacms_user WHERE id=%d", u.Id))
	if err != nil || len(rows) == 0 {
		t.Fatalf("原始查询失败: %v", err)
	}
	raw := rows[0]

	// ① 明文列必须是密文
	for _, col := range []string{"name", "nick_name", "email", "we_chat", "wei_bo", "github", "q_q"} {
		got := raw[col]
		if !strings.HasPrefix(got, "v1:") {
			t.Errorf("%s 应为密文，实际: %q", col, got)
		}
	}
	// 明文片段不应出现在任何列
	for _, col := range []string{"name", "nick_name", "email", "we_chat", "wei_bo", "github", "q_q"} {
		if strings.Contains(raw[col], "__bidx_") {
			t.Errorf("%s 密文中残留明文片段", col)
		}
	}

	// ② 盲索引列必须已填充且为 64 位 hex
	for _, col := range []string{"name_bidx", "nick_name_bidx", "email_bidx", "we_chat_bidx", "wei_bo_bidx", "github_bidx", "q_q_bidx"} {
		got := raw[col]
		if len(got) != 64 {
			t.Errorf("%s 盲索引应为 64 位 hex，实际(%d): %q", col, len(got), got)
		}
	}

	// ③ 盲索引必须与独立计算一致（保证查询侧能命中）
	wantEmailBidx, _ := BlindIndexOf(u.Email)
	if raw["email_bidx"] != wantEmailBidx {
		t.Errorf("email_bidx 不一致: db=%s calc=%s", raw["email_bidx"], wantEmailBidx)
	}

	// ④ 读回后应解密
	back := reloadUser(t, u.Id)
	if back.Name != u.Name || back.Email != u.Email || back.NickName != u.NickName {
		t.Errorf("解密失败: name=%q email=%q nick=%q", back.Name, back.Email, back.NickName)
	}
	if back.WeChat != u.WeChat || back.WeiBo != u.WeiBo || back.Github != u.Github || back.QQ != u.QQ {
		t.Errorf("社交字段解密失败: %q %q %q %q", back.WeChat, back.WeiBo, back.Github, back.QQ)
	}
}

func TestUser_LookupByEmailAndNameStillWorks(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	u, cleanup := newBidxTestUser(t, suffix)
	defer cleanup()

	// 按邮箱查（登录路径）
	byEmail := &User{Email: u.Email}
	exist, err := byEmail.GetRaw()
	if err != nil {
		t.Fatalf("按邮箱查询失败: %v", err)
	}
	if !exist {
		t.Fatal("按邮箱应能查到用户")
	}
	if byEmail.Id != u.Id {
		t.Errorf("查到的用户 id 不符: got=%d want=%d", byEmail.Id, u.Id)
	}
	if byEmail.Name != u.Name {
		t.Errorf("命中后应解密回填: got=%q want=%q", byEmail.Name, u.Name)
	}

	// 按用户名查
	byName := &User{Name: u.Name}
	if exist, err = byName.GetRaw(); err != nil || !exist {
		t.Fatalf("按用户名查询失败: exist=%v err=%v", exist, err)
	}
	if byName.Id != u.Id {
		t.Errorf("按用户名查到的 id 不符: got=%d want=%d", byName.Id, u.Id)
	}

	// 邮箱大小写/空格不同也应命中（规范化一致性）
	byTypo := &User{Email: "  " + strings.ToUpper(u.Email) + "  "}
	if exist, err = byTypo.GetRaw(); err != nil || !exist {
		t.Errorf("大小写/空格不同应命中（规范化）: exist=%v err=%v", exist, err)
	}

	// 不存在的邮箱不应命中
	miss := &User{Email: "__no_such_user_" + suffix + "@example.com"}
	if exist, _ := miss.GetRaw(); exist {
		t.Error("不存在的邮箱不应命中")
	}
}

func TestUser_DuplicateDetectionViaBlindIndex(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	u, cleanup := newBidxTestUser(t, suffix)
	defer cleanup()

	// 同名 / 同昵称 / 同邮箱（含大小写差异）都应被识别为重复
	dupName := &User{Name: u.Name}
	if repeat, err := dupName.IsNameRepeat(); err != nil || !repeat {
		t.Errorf("同名应判定重复: repeat=%v err=%v", repeat, err)
	}
	dupNick := &User{NickName: u.NickName}
	if repeat, err := dupNick.IsNickNameRepeat(); err != nil || !repeat {
		t.Errorf("同昵称应判定重复: repeat=%v err=%v", repeat, err)
	}
	dupMail := &User{Email: strings.ToUpper(u.Email)}
	if repeat, err := dupMail.IsEmailRepeat(); err != nil || !repeat {
		t.Errorf("同邮箱（大小写不同）应判定重复: repeat=%v err=%v", repeat, err)
	}

	// 唯一的新值不应判定重复
	fresh := &User{Name: "__fresh_" + suffix, NickName: "__fresh_nick_" + suffix, Email: "__fresh_" + suffix + "@example.com"}
	if repeat, _ := fresh.IsNameRepeat(); repeat {
		t.Error("新用户名不应判定重复")
	}
	if repeat, _ := fresh.IsEmailRepeat(); repeat {
		t.Error("新邮箱不应判定重复")
	}
}

func TestUser_PrepareSearch(t *testing.T) {
	initTestKeyring(t)

	u := &User{Name: "Alice", Email: "  ALICE@Example.com  "}
	if err := u.PrepareSearch(); err != nil {
		t.Fatalf("PrepareSearch: %v", err)
	}
	// 明文被清空，避免 xorm 生成错误的 WHERE
	if u.Name != "" || u.Email != "" {
		t.Errorf("PrepareSearch 后应清空明文字段: name=%q email=%q", u.Name, u.Email)
	}
	// 盲索引已生成
	if len(u.NameBidx) != 64 || len(u.EmailBidx) != 64 {
		t.Errorf("应生成盲索引: name=%q email=%q", u.NameBidx, u.EmailBidx)
	}
	// 与独立计算一致
	wantName, _ := BlindIndexOf("Alice")
	if u.NameBidx != wantName {
		t.Error("name 盲索引不一致")
	}
	wantEmail, _ := BlindIndexOf("  ALICE@Example.com  ")
	if u.EmailBidx != wantEmail {
		t.Error("email 盲索引不一致（应使用同一套规范化）")
	}

	// 幂等：再次调用不应改变结果
	before := u.EmailBidx
	if err := u.PrepareSearch(); err != nil {
		t.Fatal(err)
	}
	if u.EmailBidx != before {
		t.Error("PrepareSearch 应幂等")
	}
}
