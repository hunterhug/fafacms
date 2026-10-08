package gosession

// TokenManage token manage
// token will be put in cache database such redis and user info relate with that token will cache too
type TokenManage interface {
	SetToken(userId string, tokenValidTimes int64) (token string, err error)                           // Set token, expire after some second
	RefreshToken(token string, tokenValidTimes int64) error                                            // Refresh token，token expire will be again after some second
	DeleteToken(token string) error                                                                    // Delete token when you do action such logout
	CheckToken(token string) (user *User, exist bool, err error)                                       // Check the token,  but not refresh user info cache
	CheckTokenOrUpdateUser(token string, userInfoValidTimes int64) (user *User, exist bool, err error) // Check the token, when cache database exist return user info directly, others hit the persistent database and save newest user in cache database then return. such redis check, not check load from mysql.
	ListUserToken(userId string) ([]string, error)                                                     // List all token of one user
	DeleteUserToken(userId string) error                                                               // Delete all token of this user
	RefreshUser(userId []string, userInfoValidTimes int64) error                                       // Refresh cache of user info batch
	DeleteUser(userId string) error                                                                    // Delete user info in cache
	AddUser(userId string, userInfoValidTimes int64) (user *User, exist bool, err error)               // Add the user info to cache，expire after some second
	ConfigTokenKeyPrefix(tokenKey string) TokenManage                                                  // Config chain, just cache key prefix
	ConfigUserKeyPrefix(userKey string) TokenManage                                                    // Config chain, just cache key prefix
	ConfigDefaultExpireTime(second int64) TokenManage                                                  // Config chain, token expire after second
	ConfigGetUserInfoFunc(fn GetUserInfoFunc) TokenManage                                              // Config chain, when cache not found user info, will load from this func
	SetSingleMode() TokenManage                                                                        // Can set single mode, before one new token gen, will destroy other token
}

// User core user info, it's Id will be the primary key store in cache database such redis
type User struct {
	Id                  string      `json:"id"`     // unique mark
	TokenRemainLiveTime int64       `json:"-"`      // token remain live time in cache
	TokenExpireTime     int64       `json:"-"`      // when token expire
	Token               string      `json:"-"`      // this token
	Detail              interface{} `json:"detail"` // can diy your real user info by config ConfigGetUserInfoFunc()
}
