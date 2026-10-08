package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/hunterhug/fafacms/core/util/mail"
	"github.com/hunterhug/fafacms/core/util/oss"
	"github.com/hunterhug/fafacms/core/util/rdb"
)

// EnvFileEncrypt 用于覆盖本地文件加密开关的环境变量。
const EnvFileEncrypt = "FAFACMS_FILE_ENCRYPT"

var (
	// FaFaConfig Global config!
	FaFaConfig *Config
)

type Config struct {
	DefaultConfig MyConfig       `yaml:"DefaultConfig"` // default config
	OssConfig     oss.Key        `yaml:"OssConfig"`     // oss like aws s3
	DbConfig      rdb.MyDbConfig `yaml:"DbConfig"`      // mysql config
	SessionConfig MyRedisConf    `yaml:"SessionConfig"` // redis config for user session
	MailConfig    mail.Sender    `yaml:"MailConfig"`    // email config
}

type MyRedisConf struct {
	RedisHost        string `yaml:"RedisHost"`
	RedisMaxIdle     int    `yaml:"RedisMaxIdle"`
	RedisMaxActive   int    `yaml:"RedisMaxActive"`
	RedisIdleTimeout int    `yaml:"RedisIdleTimeout"`
	RedisDB          int    `yaml:"RedisDB"`
	RedisPass        string `yaml:"RedisPass"`
}

// MyConfig Some especial my config
type MyConfig struct {
	WebPort       string `yaml:"WebPort"`
	LogPath       string `yaml:"LogPath"`
	StoragePath   string `yaml:"StoragePath"`
	LogDebug      bool   `yaml:"LogDebug"`
	StorageOss    bool   `yaml:"StorageOss"`
	CloseRegister bool   `yaml:"CloseRegister"`
	// EncryptFile 是否加密存储本地文件（仅本地模式生效；OSS 模式强制不加密）。
	// 用指针以区分"未配置（默认开启）"与"显式关闭"。
	EncryptFile *bool `yaml:"EncryptFile"`
}

// FileEncryptEnabled 返回本地文件是否应加密存储。
//
// 规则（已确认决策）：
//   - OSS 模式：强制不加密 —— OSS 模式下文件由对象存储直链提供，加密会让浏览器拿到密文；
//   - 本地模式：默认加密，可用 `EncryptFile: false` 或环境变量 `FAFACMS_FILE_ENCRYPT=0` 关闭。
func FileEncryptEnabled() bool {
	if FaFaConfig == nil {
		return false
	}
	if FaFaConfig.DefaultConfig.StorageOss {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv(EnvFileEncrypt))) {
	case "0", "false", "off", "no":
		return false
	case "1", "true", "on", "yes":
		return true
	}
	if FaFaConfig.DefaultConfig.EncryptFile != nil {
		return *FaFaConfig.DefaultConfig.EncryptFile
	}
	return true
}

// JsonOutConfig Let the config struct to json file, just for test
func JsonOutConfig(config Config) (string, error) {
	raw, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	back := string(raw)
	return back, nil
}
