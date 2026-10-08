package model

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// rawColumn 绕过 xorm 钩子直接读原始列值，用于确认"落库是密文"。
func rawColumn(t *testing.T, sqlStr string) string {
	t.Helper()
	rows, err := FaFaRdb.Client.QueryString(sqlStr)
	if err != nil {
		t.Fatalf("原始查询失败: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("原始查询无结果: %s", sqlStr)
	}
	for _, v := range rows[0] {
		return v
	}
	return ""
}

// cleanupMessages 清理测试产生的消息行（条件为测试内可控的字面量）。
func cleanupMessages(t *testing.T, where string) {
	t.Helper()
	if _, err := FaFaRdb.Client.Exec("DELETE FROM fafacms_message WHERE " + where); err != nil {
		t.Logf("清理消息失败: %v", err)
	}
}

const (
	testSendUid    = int64(900001)
	testReceiveUid = int64(900002)
)

// ---------------------------------------------------------------------------
// 私信正文
// ---------------------------------------------------------------------------

func TestMessage_PrivateBodyEncryptedAtRest(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	body := "这是一条私密消息 secret-body-abcdef"
	if err := Private(testSendUid, testReceiveUid, body); err != nil {
		t.Fatalf("Private: %v", err)
	}
	defer cleanupMessages(t, fmt.Sprintf("send_user_id=%d AND receive_user_id=%d", testSendUid, testReceiveUid))

	// ① 直接看库：必须是密文
	raw := rawColumn(t, fmt.Sprintf("SELECT send_message FROM fafacms_message WHERE send_user_id=%d AND receive_user_id=%d ORDER BY id DESC LIMIT 1", testSendUid, testReceiveUid))
	if raw == body {
		t.Fatal("私信正文不应以明文落库")
	}
	if !strings.HasPrefix(raw, "v1:") {
		t.Fatalf("私信正文应为密文，实际: %q", raw)
	}
	if strings.Contains(raw, "secret-body") {
		t.Error("密文中不应残留明文片段")
	}

	// ② 通过模型读：钩子应自动解密
	m := new(Message)
	exist, err := FaFaRdb.Client.Where("send_user_id=?", testSendUid).And("receive_user_id=?", testReceiveUid).Desc("id").Get(m)
	if err != nil || !exist {
		t.Fatalf("读取私信失败: exist=%v err=%v", exist, err)
	}
	if m.SendMessage != body {
		t.Errorf("读取后应自动解密: got=%q want=%q", m.SendMessage, body)
	}
	if m.PrivateChanel == "" {
		t.Error("会话标识应保持明文可查")
	}

	// ③ Find 批量读取也要能解密
	list := make([]Message, 0)
	if err := FaFaRdb.Client.Where("send_user_id=?", testSendUid).Find(&list); err != nil {
		t.Fatalf("批量读取失败: %v", err)
	}
	for _, it := range list {
		if it.SendMessage != body {
			t.Errorf("批量读取未解密: got=%q", it.SendMessage)
		}
	}
}

func TestMessage_NotificationCopiesEncrypted(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	title := "文章标题-明文副本检测"
	comment := "评论摘要-明文副本检测"

	// 通知类：UserId 为触发者
	if err := CommentAbout(testSendUid, testReceiveUid, 12345, title, 67890, comment, 0, MessageTypeCommentForContent, false); err != nil {
		t.Fatalf("CommentAbout: %v", err)
	}
	defer cleanupMessages(t, fmt.Sprintf("user_id=%d AND content_id=%d", testSendUid, 12345))

	rawTitle := rawColumn(t, fmt.Sprintf("SELECT content_title FROM fafacms_message WHERE user_id=%d AND content_id=%d ORDER BY id DESC LIMIT 1", testSendUid, 12345))
	rawComment := rawColumn(t, fmt.Sprintf("SELECT comment_describe FROM fafacms_message WHERE user_id=%d AND content_id=%d ORDER BY id DESC LIMIT 1", testSendUid, 12345))

	if rawTitle == title {
		t.Error("冗余副本 content_title 不应明文落库")
	}
	if rawComment == comment {
		t.Error("冗余副本 comment_describe 不应明文落库")
	}
	if !strings.HasPrefix(rawTitle, "v1:") || !strings.HasPrefix(rawComment, "v1:") {
		t.Errorf("冗余副本应为密文: title=%q comment=%q", rawTitle, rawComment)
	}

	// 读取后应还原
	m := new(Message)
	exist, err := FaFaRdb.Client.Where("user_id=?", testSendUid).And("content_id=?", 12345).Desc("id").Get(m)
	if err != nil || !exist {
		t.Fatalf("读取通知失败: %v", err)
	}
	if m.ContentTitle != title || m.CommentDescribe != comment {
		t.Errorf("通知冗余副本未正确解密: title=%q comment=%q", m.ContentTitle, m.CommentDescribe)
	}
}

// ---------------------------------------------------------------------------
// 群发公告：两条链路 AAD 不同，需保证扇出后仍可解密
// ---------------------------------------------------------------------------

func TestGlobalMessage_EncryptedAndFanoutDecryptable(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	body := "系统公告：服务将于今晚维护 global-msg-body"

	gm := new(GlobalMessage)
	gm.CreateTime = time.Now().Unix()
	gm.Status = 1
	gm.SendMessage = body
	if _, err := FaFaRdb.Client.InsertOne(gm); err != nil {
		t.Fatalf("插入群发公告失败: %v", err)
	}
	defer func() {
		_, _ = FaFaRdb.Client.Exec("DELETE FROM fafacms_global_message WHERE id=?", gm.Id)
		cleanupMessages(t, fmt.Sprintf("global_message_id=%d", gm.Id))
	}()

	// ① 群发公告表：密文落库
	rawGM := rawColumn(t, fmt.Sprintf("SELECT send_message FROM fafacms_global_message WHERE id=%d LIMIT 1", gm.Id))
	if rawGM == body {
		t.Fatal("群发公告正文不应明文落库")
	}
	if !strings.HasPrefix(rawGM, "v1:") {
		t.Fatalf("群发公告应为密文: %q", rawGM)
	}

	// ② 扇出：调用方传入的是"刚插入后的结构体字段"（密文），
	//    实现需先按公告 AAD 解回明文，再按消息表 AAD 重新加密。
	if _, err := FanoutGlobalMessage(gm.Id, gm.SendMessage); err != nil {
		t.Fatalf("FanoutGlobalMessage: %v", err)
	}

	// ③ 投递到每个用户的 Message 行应能用「消息表」AAD 正确解密
	m := new(Message)
	exist, err := FaFaRdb.Client.Where("global_message_id=?", gm.Id).Desc("id").Get(m)
	if err != nil || !exist {
		t.Fatalf("未找到扇出后的消息行: exist=%v err=%v", exist, err)
	}
	if m.SendMessage != body {
		t.Errorf("扇出消息解密失败: got=%q want=%q", m.SendMessage, body)
	}
}

func TestInsertGlobalMessageToUser_Encrypted(t *testing.T) {
	setupTestDB(t)
	initTestKeyring(t)

	body := "按需投递的公告 insert-global-msg"
	gm := new(GlobalMessage)
	gm.CreateTime = time.Now().Unix()
	gm.Status = 1
	gm.SendMessage = body
	if _, err := FaFaRdb.Client.InsertOne(gm); err != nil {
		t.Fatalf("插入群发公告失败: %v", err)
	}
	defer func() {
		_, _ = FaFaRdb.Client.Exec("DELETE FROM fafacms_global_message WHERE id=?", gm.Id)
		cleanupMessages(t, fmt.Sprintf("global_message_id=%d", gm.Id))
	}()

	if err := InsertGlobalMessageToUser(testReceiveUid); err != nil {
		t.Fatalf("InsertGlobalMessageToUser: %v", err)
	}

	m := new(Message)
	exist, err := FaFaRdb.Client.Where("global_message_id=?", gm.Id).And("receive_user_id=?", testReceiveUid).Get(m)
	if err != nil || !exist {
		t.Fatalf("未找到投递消息: exist=%v err=%v", exist, err)
	}
	if m.SendMessage != body {
		t.Errorf("投递消息解密失败: got=%q want=%q", m.SendMessage, body)
	}
}

// ---------------------------------------------------------------------------
// AAD 归属判定：三类消息归属不同，读写得保持一致
// ---------------------------------------------------------------------------

func TestMessage_DataOwnerId(t *testing.T) {
	cases := []struct {
		name string
		msg  Message
		want int64
	}{
		{"通知类用触发者", Message{UserId: 11, ReceiveUserId: 22}, 11},
		{"私信用发送者", Message{SendUserId: 33, ReceiveUserId: 22}, 33},
		{"全局公告用接收者", Message{ReceiveUserId: 44}, 44},
		{"全空为 0", Message{}, 0},
	}
	for _, c := range cases {
		if got := c.msg.dataOwnerId(); got != c.want {
			t.Errorf("%s: got=%d want=%d", c.name, got, c.want)
		}
	}
}
