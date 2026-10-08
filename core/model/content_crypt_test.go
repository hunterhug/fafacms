package model

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// 本文件覆盖"内容类表"的加密：文章正文与标题、历史版本、举报理由、
// 评论正文与各级名称副本、专栏、关注关系、用户组、举报用户。

func TestContent_EncryptedAtRest(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	title := "文章标题-e2e-秘密-" + suffix
	preTitle := "草稿标题-秘密-" + suffix
	body := "这是文章正文 e2e-secret-body 内容"
	preBody := "草稿正文 e2e-secret-prebody"

	c := &Content{
		UserId:      testSendUid,
		NodeId:      1,
		Seo:         "e2e-seo-" + suffix,
		Title:       title,
		PreTitle:    preTitle,
		Describe:    body,
		PreDescribe: preBody,
		UserName:    "__e2e_login",
		Status:      0,
		CreateTime:  time.Now().Unix(),
	}
	if _, err := FaFaRdb.Client.InsertOne(c); err != nil {
		t.Fatalf("插入文章失败: %v", err)
	}
	defer func() {
		_, _ = FaFaRdb.Client.Exec(fmt.Sprintf("DELETE FROM fafacms_content WHERE id=%d", c.Id))
	}()
	if c.Id == 0 {
		t.Fatal("应拿到自增 Id")
	}

	rows, err := FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT title, pre_title, `describe`, pre_describe, user_name, title_bidx, pre_title_bidx "+
			"FROM fafacms_content WHERE id=%d", c.Id))
	if err != nil || len(rows) == 0 {
		t.Fatalf("原始查询失败: %v", err)
	}
	raw := rows[0]

	for _, col := range []string{"title", "pre_title", "describe", "pre_describe", "user_name"} {
		if !strings.HasPrefix(raw[col], "v1:") {
			t.Errorf("%s 应为密文，实际: %q", col, raw[col])
		}
	}
	if strings.Contains(raw["title"], "e2e-秘密") || strings.Contains(raw["describe"], "e2e-secret-body") {
		t.Error("密文中不应残留明文片段")
	}
	if len(raw["title_bidx"]) != 64 || len(raw["pre_title_bidx"]) != 64 {
		t.Errorf("标题盲索引应已填充: %q / %q", raw["title_bidx"], raw["pre_title_bidx"])
	}

	// 盲索引与独立计算一致（保证按标题精确搜索能命中）
	wantBidx, _ := BlindIndexOf(title)
	if raw["title_bidx"] != wantBidx {
		t.Error("title_bidx 与独立计算不一致")
	}

	// 读回解密
	got := new(Content)
	exist, err := FaFaRdb.Client.ID(c.Id).Get(got)
	if err != nil || !exist {
		t.Fatalf("读取失败: exist=%v err=%v", exist, err)
	}
	if got.Title != title || got.PreTitle != preTitle {
		t.Errorf("标题解密失败: %q / %q", got.Title, got.PreTitle)
	}
	if got.Describe != body || got.PreDescribe != preBody {
		t.Errorf("正文解密失败: %q / %q", got.Describe, got.PreDescribe)
	}
	if got.UserName != "__e2e_login" {
		t.Errorf("冗余副本解密失败: %q", got.UserName)
	}
}

func TestContent_ExactTitleSearchViaBlindIndex(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	title := "可搜索标题-" + suffix

	c := &Content{
		UserId: testSendUid, NodeId: 1, Seo: "e2e-search-" + suffix,
		Title: title, Describe: "x", Status: 0, CreateTime: time.Now().Unix(),
	}
	if _, err := FaFaRdb.Client.InsertOne(c); err != nil {
		t.Fatalf("插入失败: %v", err)
	}
	defer func() {
		_, _ = FaFaRdb.Client.Exec(fmt.Sprintf("DELETE FROM fafacms_content WHERE id=%d", c.Id))
	}()

	// 精确标题：应命中
	bidx, _ := BlindIndexOf(title)
	rows, err := FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT COUNT(*) AS c FROM fafacms_content WHERE title_bidx='%s'", bidx))
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if rows[0]["c"] == "0" {
		t.Error("精确标题应能命中")
	}

	// 标题片段（模糊）：不应命中——模糊搜索已按既定决策取消
	partBidx, _ := BlindIndexOf(title[:5])
	rows, err = FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT COUNT(*) AS c FROM fafacms_content WHERE title_bidx='%s'", partBidx))
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if rows[0]["c"] != "0" {
		t.Error("标题片段不应命中（模糊搜索已取消）")
	}
}

func TestContentHistory_Encrypted(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	h := &ContentHistory{
		ContentId: 999000, UserId: testSendUid, NodeId: 1,
		Title: "历史标题-e2e-" + suffix, Describe: "历史正文 e2e-history-secret",
		CreateTime: time.Now().Unix(),
	}
	if _, err := FaFaRdb.Client.InsertOne(h); err != nil {
		t.Fatalf("插入历史版本失败: %v", err)
	}
	defer func() {
		_, _ = FaFaRdb.Client.Exec(fmt.Sprintf("DELETE FROM fafacms_content_history WHERE id=%d", h.Id))
	}()

	rows, err := FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT title, `describe` FROM fafacms_content_history WHERE id=%d", h.Id))
	if err != nil || len(rows) == 0 {
		t.Fatalf("查询失败: %v", err)
	}
	if !strings.HasPrefix(rows[0]["title"], "v1:") || !strings.HasPrefix(rows[0]["describe"], "v1:") {
		t.Errorf("历史版本应加密: title=%q describe=%q", rows[0]["title"], rows[0]["describe"])
	}

	back := new(ContentHistory)
	exist, err := FaFaRdb.Client.ID(h.Id).Get(back)
	if err != nil || !exist {
		t.Fatalf("读回失败: %v", err)
	}
	if back.Title != h.Title || back.Describe != h.Describe {
		t.Errorf("历史版本解密失败: %q / %q", back.Title, back.Describe)
	}
}

func TestComment_EncryptedWithCopies(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	c := &Comment{
		UserId:              testSendUid,
		UserName:            "评论者-" + suffix,
		ContentId:           123456,
		ContentTitle:        "被评论的文章标题-e2e",
		ContentUserId:       777,
		ContentUserName:     "文章作者-e2e",
		CommentUserName:     "被回复者-e2e",
		RootCommentUserName: "根评论者-e2e",
		Describe:            "评论正文 e2e-comment-secret",
		CreateTime:          time.Now().Unix(),
	}
	if _, err := FaFaRdb.Client.InsertOne(c); err != nil {
		t.Fatalf("插入评论失败: %v", err)
	}
	defer func() {
		_, _ = FaFaRdb.Client.Exec(fmt.Sprintf("DELETE FROM fafacms_comment WHERE id=%d", c.Id))
	}()

	rows, err := FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT user_name, content_title, content_user_name, comment_user_name, "+
			"root_comment_user_name, `describe` FROM fafacms_comment WHERE id=%d", c.Id))
	if err != nil || len(rows) == 0 {
		t.Fatalf("查询失败: %v", err)
	}
	raw := rows[0]
	for _, col := range []string{"user_name", "content_title", "content_user_name",
		"comment_user_name", "root_comment_user_name", "describe"} {
		if !strings.HasPrefix(raw[col], "v1:") {
			t.Errorf("%s 应为密文（冗余副本最易漏），实际: %q", col, raw[col])
		}
	}

	back := new(Comment)
	exist, err := FaFaRdb.Client.ID(c.Id).Get(back)
	if err != nil || !exist {
		t.Fatalf("读回失败: %v", err)
	}
	if back.Describe != c.Describe || back.ContentTitle != c.ContentTitle ||
		back.UserName != c.UserName || back.CommentUserName != c.CommentUserName {
		t.Errorf("评论解密失败: %+v", back)
	}
}

func TestNode_Encrypted(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	n := &ContentNode{
		UserId: testSendUid, Seo: "e2e-node-" + suffix,
		Name: "专栏名-e2e-" + suffix, Describe: "专栏描述 e2e-node-secret",
		UserName: "专栏主人-e2e", Status: 1, CreateTime: time.Now().Unix(),
	}
	if _, err := FaFaRdb.Client.InsertOne(n); err != nil {
		t.Fatalf("插入专栏失败: %v", err)
	}
	defer func() {
		_, _ = FaFaRdb.Client.Exec(fmt.Sprintf("DELETE FROM fafacms_content_node WHERE id=%d", n.Id))
	}()

	rows, err := FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT name, `describe`, user_name FROM fafacms_content_node WHERE id=%d", n.Id))
	if err != nil || len(rows) == 0 {
		t.Fatalf("查询失败: %v", err)
	}
	for _, col := range []string{"name", "describe", "user_name"} {
		if !strings.HasPrefix(rows[0][col], "v1:") {
			t.Errorf("%s 应为密文: %q", col, rows[0][col])
		}
	}

	back := new(ContentNode)
	if exist, err := FaFaRdb.Client.ID(n.Id).Get(back); err != nil || !exist {
		t.Fatalf("读回失败: %v", err)
	}
	if back.Name != n.Name || back.Describe != n.Describe || back.UserName != n.UserName {
		t.Errorf("专栏解密失败: %q / %q / %q", back.Name, back.Describe, back.UserName)
	}
}

func TestRelation_Encrypted(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	r := &Relation{
		UserAId: testSendUid, UserBId: testReceiveUid,
		UserAName: "关注者-e2e", UserBName: "被关注者-e2e",
		CreateTime: time.Now().Unix(),
	}
	if _, err := FaFaRdb.Client.InsertOne(r); err != nil {
		t.Fatalf("插入关注关系失败: %v", err)
	}
	defer func() {
		_, _ = FaFaRdb.Client.Exec(fmt.Sprintf("DELETE FROM fafacms_relation WHERE id=%d", r.Id))
	}()

	rows, err := FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT user_a_name, user_b_name FROM fafacms_relation WHERE id=%d", r.Id))
	if err != nil || len(rows) == 0 {
		t.Fatalf("查询失败: %v", err)
	}
	if !strings.HasPrefix(rows[0]["user_a_name"], "v1:") || !strings.HasPrefix(rows[0]["user_b_name"], "v1:") {
		t.Errorf("关系名称副本应加密: %q / %q", rows[0]["user_a_name"], rows[0]["user_b_name"])
	}

	back := new(Relation)
	if exist, err := FaFaRdb.Client.ID(r.Id).Get(back); err != nil || !exist {
		t.Fatalf("读回失败: %v", err)
	}
	if back.UserAName != r.UserAName || back.UserBName != r.UserBName {
		t.Errorf("关系解密失败: %q / %q", back.UserAName, back.UserBName)
	}
}

func TestGroup_EncryptedWithBlindIndex(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	name := "组名-e2e-" + suffix
	g := &Group{Name: name, Describe: "组描述 e2e-group-secret", CreateTime: time.Now().Unix()}
	if _, err := FaFaRdb.Client.InsertOne(g); err != nil {
		t.Fatalf("插入用户组失败: %v", err)
	}
	defer func() {
		_, _ = FaFaRdb.Client.Exec(fmt.Sprintf("DELETE FROM fafacms_group WHERE id=%d", g.Id))
	}()

	rows, err := FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT name, `describe`, name_bidx FROM fafacms_group WHERE id=%d", g.Id))
	if err != nil || len(rows) == 0 {
		t.Fatalf("查询失败: %v", err)
	}
	if !strings.HasPrefix(rows[0]["name"], "v1:") || !strings.HasPrefix(rows[0]["describe"], "v1:") {
		t.Errorf("组名与描述应加密: %q / %q", rows[0]["name"], rows[0]["describe"])
	}
	wantBidx, _ := BlindIndexOf(name)
	if rows[0]["name_bidx"] != wantBidx {
		t.Error("组名盲索引不一致")
	}

	// 按名称查询（走盲索引）
	found := new(Group)
	found.Name = name
	if err := found.PrepareSearch(); err != nil {
		t.Fatal(err)
	}
	exist, err := FaFaRdb.Client.Get(found)
	if err != nil || !exist {
		t.Fatalf("按组名查询失败: exist=%v err=%v", exist, err)
	}
	if found.Id != g.Id {
		t.Errorf("查到的组 id 不符: %d vs %d", found.Id, g.Id)
	}
	if found.Describe != g.Describe {
		t.Errorf("组描述解密失败: %q", found.Describe)
	}
}

func TestUserBad_ReasonEncrypted(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	b := &UserBad{
		UserId: testSendUid, BadUserId: testReceiveUid,
		Reason: "举报理由 e2e-userbad-secret", CreateTime: time.Now().Unix(),
	}
	if _, err := FaFaRdb.Client.InsertOne(b); err != nil {
		t.Fatalf("插入举报失败: %v", err)
	}
	defer func() {
		_, _ = FaFaRdb.Client.Exec(fmt.Sprintf("DELETE FROM fafacms_user_bad WHERE id=%d", b.Id))
	}()

	rows, err := FaFaRdb.Client.QueryString(fmt.Sprintf(
		"SELECT reason FROM fafacms_user_bad WHERE id=%d", b.Id))
	if err != nil || len(rows) == 0 {
		t.Fatalf("查询失败: %v", err)
	}
	if !strings.HasPrefix(rows[0]["reason"], "v1:") {
		t.Errorf("举报理由应加密: %q", rows[0]["reason"])
	}

	back := new(UserBad)
	if exist, err := FaFaRdb.Client.ID(b.Id).Get(back); err != nil || !exist {
		t.Fatalf("读回失败: %v", err)
	}
	if back.Reason != b.Reason {
		t.Errorf("举报理由解密失败: %q", back.Reason)
	}
}
