package controllers

import (
	"github.com/hunterhug/fafacms/core/model"
)

// 说明：用户名（user.name）与各表里的用户名冗余副本均已加密存储，无法再用
// `user_name = ?` 做等值过滤。因此这里统一把"按用户名过滤"解析为该用户的 ID，
// 再改按 *_id 列过滤——功能等价，且这些表本来就有对应的 ID 列。

// resolveUserIdByName 按用户名解析用户 ID（用户名已加密，走盲索引精确匹配）。
// 返回 0 表示用户不存在。
func resolveUserIdByName(name string) (int64, error) {
	if name == "" {
		return 0, nil
	}
	u := new(model.User)
	u.Name = name
	exist, err := u.GetRaw()
	if err != nil {
		return 0, err
	}
	if !exist {
		return 0, nil
	}
	return u.Id, nil
}

// userFilterId 返回按用户名过滤时应使用的用户 ID。
// 用户不存在时返回 -1（不可能匹配任何行），使查询结果为空，避免退化为"忽略该过滤"。
func userFilterId(name string) (int64, error) {
	id, err := resolveUserIdByName(name)
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return -1, nil
	}
	return id, nil
}
