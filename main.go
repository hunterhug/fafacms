package main

import (
	"flag"
	"fmt"
	"github.com/hunterhug/fafacms/core/config"
	"github.com/hunterhug/fafacms/core/controllers"
	"github.com/hunterhug/fafacms/core/flog"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/router"
	"github.com/hunterhug/fafacms/core/server"
	"github.com/hunterhug/fafacms/core/session"
	"github.com/hunterhug/fafacms/core/util"
	"github.com/hunterhug/fafacms/core/util/mail"
	"github.com/hunterhug/fafacms/core/util/rds"
	"github.com/hunterhug/fafacms/core/util/secure"
	log "github.com/hunterhug/golog"
	"time"
)

var (
	// Global path of config file
	configFile string

	// Auto create database and tables when set true
	createTable bool

	// Email debug will not send email
	mailDebug bool

	// Skip some admin Auth when debug
	canSkipAuth bool

	// Record the history of content when edit or publish
	historyRecord bool

	// Time zone offset the utc, default is 8,Beijing
	timeZone int64

	// Auto ban the content or comment
	autoBan bool

	// Beyond and autoBan is true will Ban it!
	banTime int64

	// Single Login
	singleLogin bool

	// Login Session expire time
	sessionExpireTime int64

	// Scale the picture auto
	canScale   bool
	scaleWidth int

	// Generate a new root key for FAFACMS_KEK_ROOT then exit
	genKek bool
)

// Parse flag when init
// Those variables will not config in file
func init() {
	// Default read ./config.yaml
	flag.StringVar(&configFile, "config", "./config.yaml", "Config file")

	// Auto init db, the second time can set false
	flag.BoolVar(&createTable, "init_db", true, "Init create db table")

	// Some important config
	flag.BoolVar(&canScale, "can_scale", true, "Can scale the picture auto")
	flag.IntVar(&scaleWidth, "scale_width", 500, "The width of scale size of picture")
	flag.Int64Var(&timeZone, "time_zone", 8, "Time zone offset the utc")
	flag.BoolVar(&autoBan, "auto_ban", false, "Auto ban the content or comment")
	flag.Int64Var(&banTime, "ban_time", 10, "Content or comment will be ban in how much bad's time")
	flag.BoolVar(&historyRecord, "history_record", true, "Content history can be record")
	flag.BoolVar(&singleLogin, "single_login", false, "User can only single point login")
	flag.Int64Var(&sessionExpireTime, "session_expire_time", 7*3600*24, "Login session expire second time, token will destroy after this time")

	// When in production, please set to all false
	flag.BoolVar(&mailDebug, "email_debug", false, "Email debug")
	flag.BoolVar(&canSkipAuth, "auth_skip_debug", false, "Auth skip debug")

	// 生成一把新的根密钥（写入环境变量 FAFACMS_KEK_ROOT 用），打印后退出
	flag.BoolVar(&genKek, "gen_kek", false, "Generate a new root key for FAFACMS_KEK_ROOT and exit")

	flag.Parse()
}

// Init the URL resource, some admin url put inside a map will save a lot of time
func initResource() (adminUrl map[string]int64) {
	adminUrl = make(map[string]int64)
	for url, handler := range router.V1Router {
		if !handler.Admin {
			continue
		}
		r := new(model.Resource)
		url1 := fmt.Sprintf("/v1%s", url)
		r.UrlHash, _ = util.Sha256([]byte(url1))
		r.Admin = true
		exist, err := r.GetRaw()
		if err != nil {
			panic(err)
		}

		// Exist will put in map, otherwise save in db then put in map
		if exist {
			adminUrl[url1] = r.Id
			continue
		} else {
			r := new(model.Resource)
			r.Url = url1
			r.UrlHash, _ = util.Sha256([]byte(url1))
			r.Name = handler.Name
			r.Describe = handler.Name
			r.Admin = handler.Admin
			r.CreateTime = time.Now().Unix()
			err := r.InsertOne()
			if err != nil {
				panic(err)
			}
			adminUrl[url1] = r.Id
		}
	}
	return adminUrl
}

// The Beauty Main
// I'm FaFa
func main() {
	welcome()

	// 生成根密钥后直接退出（用于初始化环境变量 FAFACMS_KEK_ROOT）
	if genKek {
		key, err := secure.GenerateKey()
		if err != nil {
			log.Panicf("GenKek err: %s", err.Error())
			return
		}
		fmt.Printf("%s=%s\n", secure.EnvKEKRoot, key)
		return
	}

	// Package var init
	mail.Debug = mailDebug
	controllers.AuthDebug = canSkipAuth
	controllers.TimeZone = timeZone
	controllers.BadTime = banTime
	controllers.AutoBan = autoBan
	controllers.SessionExpireTime = sessionExpireTime
	controllers.CanScale = canScale
	controllers.ScaleWidth = scaleWidth
	model.HistoryRecord = historyRecord

	var err error

	// Init global config
	err = server.InitYamlConfig(configFile)
	if err != nil {
		log.Panicf("InitYamlConfig err: %s", err.Error())
		return
	}

	// 只打印非敏感摘要：绝不输出整份配置（含数据库/Redis/OSS 凭据）
	log.Infof("Hi! Config loaded: web_port=%s storage_oss=%v storage_path=%s log_path=%s db_host=%s db_name=%s db_port=%s",
		config.FaFaConfig.DefaultConfig.WebPort,
		config.FaFaConfig.DefaultConfig.StorageOss,
		config.FaFaConfig.DefaultConfig.StoragePath,
		config.FaFaConfig.DefaultConfig.LogPath,
		config.FaFaConfig.DbConfig.Host,
		config.FaFaConfig.DbConfig.Name,
		config.FaFaConfig.DbConfig.Port,
	)

	// Init 数据加密密钥：优先环境变量 FAFACMS_KEK_ROOT。
	// 未设置时回退到内置默认密钥并打出醒目警告（详见 secure.DefaultKEKRoot 的说明）；
	// 设置但非法则拒绝启动，避免用错误密钥写入数据导致永久无法解密。
	var usedDefaultKek bool
	if usedDefaultKek, err = secure.InitFromEnv(); err != nil {
		log.Panicf("Init secure keyring err: %s", err.Error())
		return
	}
	if usedDefaultKek {
		log.Warnf("!!! SECURITY WARNING: %s is NOT set, using the BUILT-IN DEFAULT key. "+
			"The default key is public with the source code, so database dumps/backups are effectively UNENCRYPTED. "+
			"Set %s (32 bytes, base64 or hex) for any real deployment; `./install/deploy.sh` generates and persists one automatically.",
			secure.EnvKEKRoot, secure.EnvKEKRoot)
	}
	if fp, ferr := secure.Fingerprint(); ferr == nil {
		// 只打印指纹，不打印密钥本身
		log.Infof("Data encryption key ready (fingerprint=%s, builtin_default=%v)", fp, usedDefaultKek)
	}

	// Init log
	flog.InitLog(config.FaFaConfig.DefaultConfig.LogPath, config.FaFaConfig.DefaultConfig.LogDebug)

	// Init db
	err = server.InitRdb(config.FaFaConfig.DbConfig)
	if err != nil {
		log.Panicf("InitRdb err: %s", err.Error())
		return
	}

	// Init session
	err = session.InitSession(config.FaFaConfig.SessionConfig, singleLogin)
	if err != nil {
		log.Panicf("InitSession err: %s", err.Error())
		return
	}

	// Init Redis cache for shared security state (rate limit/sign/risk control/2FA/captcha)
	rc := config.FaFaConfig.SessionConfig
	err = rds.Init(rc.RedisHost, rc.RedisDB, rc.RedisPass, rc.RedisMaxIdle, rc.RedisMaxActive, rc.RedisIdleTimeout)
	if err != nil {
		log.Panicf("InitRds err: %s", err.Error())
		return
	}

	// Auto create db table
	if createTable {
		model.CreateTable([]interface{}{
			model.User{},           // User Table
			model.Group{},          // User Group, every user can assign a group
			model.Resource{},       // Url Resource, if user not own those will be refuse to auth
			model.GroupResource{},  // Resource will be assign to group
			model.Content{},        // Content Table, very import
			model.ContentCool{},    // Content Cool, user can cool your content
			model.ContentBad{},     // Content Bad, user can bad your content and if auto ban, your content will be ban
			model.ContentHistory{}, // Content History, when publish or edit a content, and you set history record, emm, save it
			model.ContentNode{},    // Contents' Node, every content must belong to a node
			model.File{},           // File Table, your picture file and some will save in.
			model.Comment{},        // Comment Table, comment for content, comment for comment
			model.CommentCool{},    // Like the Content Cool
			model.CommentBad{},     // Like the Content Bad
			model.UserBad{},        // Bad User (report user)
			model.Relation{},       // Who follow who
			model.Message{},        // Message inside
			model.GlobalMessage{},  // Global Message helper
			model.SiteConfig{},     // Site config (title/footer)
			model.FriendLink{},     // Friend link in footer
			//model.Log{},            // Log Table, not use
		})
	}

	controllers.AdminUrl = initResource()

	// Count ticker
	go controllers.LoopCount()

	// Server Run
	engine := server.Server()
	// 本地存储文件由后端受控提供（原先用 engine.Static 直接静态服务）：
	// 文件内容可能是密文，必须解密后再返回；未加密文件行为与原静态服务一致。
	engine.GET("/storage/*path", controllers.ServeStorage)
	engine.HEAD("/storage/*path", controllers.ServeStorage)
	engine.GET("/storage_x/*path", controllers.ServeStorageX)
	engine.HEAD("/storage_x/*path", controllers.ServeStorageX)

	// Web welcome home!
	router.SetRouter(engine)

	// V1 API, will maybe change to V2...
	v1 := engine.Group("/v1")
	v1.Use(controllers.AuthFilter)

	// Router Set
	router.SetAPIRouter(v1, router.V1Router)

	log.Infof("Server run in %s", config.FaFaConfig.DefaultConfig.WebPort)

	err = engine.Run(config.FaFaConfig.DefaultConfig.WebPort)
	if err != nil {
		log.Errorf("Server run err: %s", err.Error())
		return
	}
}

func welcome() {
	log.Infof("Hi! %s! A Nice CMS.", config.Title)
	s := `
███████╗ █████╗ ███████╗ █████╗  ██████╗███╗   ███╗███████╗
██╔════╝██╔══██╗██╔════╝██╔══██╗██╔════╝████╗ ████║██╔════╝
█████╗  ███████║█████╗  ███████║██║     ██╔████╔██║███████╗
██╔══╝  ██╔══██║██╔══╝  ██╔══██║██║     ██║╚██╔╝██║╚════██║
██║     ██║  ██║██║     ██║  ██║╚██████╗██║ ╚═╝ ██║███████║
╚═╝     ╚═╝  ╚═╝╚═╝     ╚═╝  ╚═╝ ╚═════╝╚═╝     ╚═╝╚══════╝`
	//log.Infof("\n%s-v%s\n", s, config.Version, util.BuildTime())
	log.Infof("\n%s-%s\n", s, config.Version)
}
