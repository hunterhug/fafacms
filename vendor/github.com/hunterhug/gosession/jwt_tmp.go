package gosession

import (
	"time"
)

const (
	// JwtPayloadClientVersionName Todo client jwtPayload version field
	JwtPayloadClientVersionName = "clientVersion"
	// JwtPayloadClientVersion client jwtPayload version value
	JwtPayloadClientVersion = 1.0
)

// JwtManage Todo 另外一种实现的 Session
type JwtManage interface {
	// Config 配置令牌有效时间
	Config(refreshTokenExpireTime time.Duration, accessTokenExpireTime time.Duration)
	// CreateNewLogInToken 创建令牌对，可传入客户端或服务端数据
	CreateNewLogInToken(userId string, clientJwtPayload map[string]interface{}, serverSessionData map[string]interface{}) (accessToken, refreshToken string, err error)
	// GetSessionInfoByAccessToken 获取某令牌对信息，先走JWT无状态解码数据，force则强制查询服务端
	GetSessionInfoByAccessToken(accessToken string, force bool) (jwtData JwtData, err error)
	// GetAllSessionsForUser 获取某用户所有令牌对信息
	GetAllSessionsForUser(userId string) (jwtDataList []JwtData, err error)
	// RevokeLogInTokenByUserId 撤销某用户所有令牌对
	RevokeLogInTokenByUserId(userId string) error
	// RevokeLogInTokenByAccessToken 撤销用户某令牌对
	RevokeLogInTokenByAccessToken(accessToken string) (revoke bool, err error)
	// RefreshLogInToken 刷新令牌对
	RefreshLogInToken(refreshToken string) (newAccessToken string, newRefreshToken string, err error)
	// GetSignPublicKey 获取服务端签名key，客户端使用这个来进行解密，可以是空
	GetSignPublicKey() (publicKey string, err error)
	// ParseAccessToken 解析和验证JWT信息，客户端方法
	ParseAccessToken(accessToken string, publicKey string) (jwtData JwtData, err error)
}

// JwtData 数据
type JwtData struct {
	// 用户ID
	UserId string `json:"user_id,omitempty"`
	// 创建时间，毫秒
	CreateMSTime int64 `json:"create_ms_time,omitempty"`
	// 辅助
	CreateString string `json:"create_string,omitempty"`
	// 过期时间，毫秒
	ExpiryMSTime int64 `json:"expiry_ms_time,omitempty"`
	// 辅助
	ExpiryString string `json:"expiry_string,omitempty"`
	// 是否已过期
	IsExpire bool `json:"is_expire,omitempty"`
	// 服务端数据库令牌对ID，可忽视
	ServerTokenHandle string `json:"server_token_handle,omitempty"`
	// 客户端Payload数据，返回给客户端的数据，客户端解密可看
	ClientPayload map[string]interface{} `json:"client_payload,omitempty"`
	// 服务端Session数据，保存在服务器端的数据，客户端不可看
	ServerSessionData map[string]interface{} `json:"server_session_data,omitempty"`
}
