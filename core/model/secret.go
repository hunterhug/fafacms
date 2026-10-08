package model

import (
	"github.com/hunterhug/fafacms/core/util/secure"
)

// DigestSecretCode 计算短凭据（邮箱激活码 / 忘记密码验证码）的摘要。
//
// 入库与校验一律使用摘要，绝不存明文：数据库泄漏时无法据此激活他人账号
// 或重置他人密码。摘要不可逆且确定性，因此仍可按值比对。
func DigestSecretCode(plain string) (string, error) {
	k, err := secure.Get()
	if err != nil {
		return "", err
	}
	return k.CredentialDigest(plain), nil
}

// BlindIndexOf 计算盲索引：HMAC(K_index, 规范化(明文))。
//
// 用于「需要等值查询或唯一约束、但列本身要加密」的字段。
// 确定性：同一明文（规范化后）永远得到同一结果，因此可以建索引、可以做 = 查询。
func BlindIndexOf(plain string) (string, error) {
	k, err := secure.Get()
	if err != nil {
		return "", err
	}
	return k.BlindIndex(plain), nil
}

// BlindIndexOptional 对可空字段求盲索引：空值返回空字符串（不建索引）。
func BlindIndexOptional(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	return BlindIndexOf(plain)
}

// SecretCodeMatch 判断明文凭据是否与库中摘要一致。
func SecretCodeMatch(digest, plain string) bool {
	if digest == "" || plain == "" {
		return false
	}
	got, err := DigestSecretCode(plain)
	if err != nil {
		return false
	}
	return got == digest
}

// EncryptUserField 加密 user 表的某个字段（AAD 绑定表名/字段名/用户 ID）。
func EncryptUserField(field string, userId int64, plain string) (string, error) {
	return EncryptTableField("user", field, userId, plain)
}

// DecryptUserField 解密 user 表的某个字段。
func DecryptUserField(field string, userId int64, stored string) (string, error) {
	return DecryptTableField("user", field, userId, stored)
}

// EncryptTableField 加密任意表的字段，AAD 绑定「表名:字段名:归属用户ID」。
// 归属用户 ID 用于把密文与该行绑定：把密文搬到别的行/别的字段后解密会失败。
func EncryptTableField(table, field string, ownerId int64, plain string) (string, error) {
	k, err := secure.Get()
	if err != nil {
		return "", err
	}
	return k.EncryptField(table, field, ownerId, plain)
}

// DecryptTableField 解密任意表的字段。
func DecryptTableField(table, field string, ownerId int64, stored string) (string, error) {
	k, err := secure.Get()
	if err != nil {
		return "", err
	}
	return k.DecryptField(table, field, ownerId, stored)
}

// EncryptOptional 加密可空字段：空值原样返回（不产生密文），避免把"未填写"变成密文。
func EncryptOptional(table, field string, ownerId int64, plain string) (string, error) {
	if plain == "" || secure.IsCipher(plain) {
		return plain, nil
	}
	return EncryptTableField(table, field, ownerId, plain)
}

// DecryptOptional 解密可空字段：空值原样返回。
func DecryptOptional(table, field string, ownerId int64, stored string) (string, error) {
	if stored == "" {
		return "", nil
	}
	return DecryptTableField(table, field, ownerId, stored)
}
