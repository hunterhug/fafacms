package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/controllers"
)

type HttpHandle struct {
	Name   string
	Func   gin.HandlerFunc
	Method []string
	Admin  bool
}

var (
	POST = []string{"POST"}
	GET  = []string{"GET"}
	GP   = []string{"POST", "GET"}
)

// Router
var (
	HomeRouter = map[string]HttpHandle{
		// Home Router, not need auth
		"/": {"Home", controllers.Home, GP, false},

		"/u":               {"List Peoples", controllers.Peoples, GP, false},                    // 列出用户
		"/u/node":          {"List User Nodes One", controllers.NodeInfo, GP, false},            // 查找某一个节点
		"/u/nodes":         {"List User Nodes", controllers.NodesInfo, GP, false},               // 列出某用户下的节点
		"/u/info":          {"List User Info", controllers.UserInfo, GP, false},                 // 获取某用户信息
		"/u/count":         {"Count User Content", controllers.UserCount, GP, false},            // 统计某用户文章情况（某用户可留空）
		"/u/content":       {"List User Content", controllers.Contents, GP, false},              // 列出某用户下文章（某用户可留空）
		"/content":         {"Get Content", controllers.Content, GP, false},                     // 获取文章
		"/content/comment": {"List Comment of Content", controllers.ListHomeComment, GP, false}, // 列出文章下的评论
		"/captcha":         {"Captcha Image", controllers.Captcha, GP, false},                   // 图形验证码
		"/captcha/unblock": {"Unblock IP", controllers.Unblock, GP, false},                      // 限流解封（验证码通过后解除该 IP 拉黑）
		"/site/config":     {"Site Config Public", controllers.SiteConfig, GP, false},           // 站点配置 + 可见友情链接

		"/user/token/get":       {"User Token get", controllers.Login, GP, false},
		"/user/token/2fa":       {"User Token 2FA", controllers.LoginTwoFa, GP, false}, // 两步验证登录
		"/user/token/refresh":   {"User Token refresh", controllers.Refresh, GP, false},
		"/user/token/delete":    {"User Token delete", controllers.Logout, GP, false},
		"/user/register":        {"User Register", controllers.RegisterUser, GP, false},
		"/user/activate":        {"User Verify Email To Activate", controllers.ActivateUser, GP, false},               // 用户自己激活
		"/user/activate/code":   {"User Resend Email Activate Code", controllers.ResendActivateCodeToUser, GP, false}, // 激活码过期重新获取
		"/user/password/forget": {"User Forget Password Gen Code", controllers.ForgetPasswordOfUser, GP, false},       // 忘记密码，验证码发往邮箱
		"/user/password/change": {"User Change Password", controllers.ChangePasswordOfUser, GP, false},                // 根据邮箱验证码修改密码
	}

	// /v1/user/create
	// need login group auth
	V1Router = map[string]HttpHandle{
		// 用户组操作
		"/group/create":        {"Create Group", controllers.CreateGroup, POST, true},
		"/group/update":        {"Update Group", controllers.UpdateGroup, POST, true},
		"/group/delete":        {"Delete Group", controllers.DeleteGroup, POST, true},
		"/group/take":          {"Take Group", controllers.TakeGroup, GP, true},
		"/group/list":          {"List Group", controllers.ListGroup, GP, true},
		"/group/user/list":     {"Group List User", controllers.ListGroupUser, GP, true},         // 超级管理员列出组下的用户
		"/group/resource/list": {"Group List Resource", controllers.ListGroupResource, GP, true}, // 超级管理员列出组下的资源

		// 用户操作
		"/user/list":            {"User List All", controllers.ListUser, GP, true},               // 超级管理员列出用户列表
		"/user/create":          {"User Create", controllers.CreateUser, GP, true},               // 超级管理员创建用户，默认激活
		"/user/assign":          {"User Assign Group", controllers.AssignGroupToUser, GP, true},  // 超级管理员给用户分配用户组
		"/user/update":          {"User Update Self", controllers.UpdateUser, GP, false},         // 更新自己的信息
		"/user/admin/update":    {"User Update Admin", controllers.UpdateUserAdmin, GP, true},    // 管理员修改其他用户信息，可以修改用户密码，以及将用户加入黑名单，禁止使用等
		"/user/info":            {"User Info Self", controllers.TakeUser, GP, false},             // 获取自己的信息
		"/user/perm":            {"User Admin Perm", controllers.UserPerm, GP, false},            // 获取自己管理后台可见资源（组权限）
		"/user/2fa/secret":      {"User 2FA Secret", controllers.TwoFaSecret, GP, false},         // 生成 2FA 秘钥 + otpauth URI（未开启时）
		"/user/2fa/enable":      {"User 2FA Enable", controllers.EnableTwoFa, GP, false},         // 绑定 2FA
		"/user/2fa/disable":     {"User 2FA Disable", controllers.DisableTwoFa, GP, false},       // 关闭 2FA
		"/user/admin/2fa/reset": {"Admin Reset User 2FA", controllers.ResetTwoFaAdmin, GP, true}, // 管理员重置用户 2FA
		"/user/bad":             {"Bad User", controllers.BadUser, POST, false},                  // 举报用户
		"/user/admin/bad/list":  {"List User Bad Admin", controllers.ListUserBadAdmin, GP, true}, // 管理员列出用户举报

		// 资源操作
		"/resource/list":   {"Resource List All", controllers.ListResource, GP, true},              // 列出资源
		"/resource/assign": {"Resource Assign Group", controllers.AssignResourceToGroup, GP, true}, // 资源分配给组

		// 文件操作
		"/file/upload":       {"File Upload", controllers.UploadFile, POST, false},
		"/file/list":         {"File List Self", controllers.ListFile, POST, false},
		"/file/admin/list":   {"File List All", controllers.ListFileAdmin, POST, true}, // 管理员查看所有文件
		"/file/update":       {"File Update Self", controllers.UpdateFile, POST, false},
		"/file/admin/update": {"File Update All", controllers.UpdateFileAdmin, POST, true}, // 管理员修改文件

		// 比较重要的, 节点和文章都应该支持拖曳，文章首页排序还是按照创建时间，但是后台使用排序字段
		// 内容节点操作
		"/node/create":        {"Create Node Self", controllers.CreateNode, POST, false},
		"/node/update/seo":    {"Update Node Self Seo", controllers.UpdateSeoOfNode, POST, false},          // 更新节点SEO
		"/node/update/info":   {"Update Node Self Info", controllers.UpdateInfoOfNode, POST, false},        // 更新节点名字和描述
		"/node/update/image":  {"Update Node Self Info Image", controllers.UpdateImageOfNode, POST, false}, // 更新图片地址
		"/node/update/status": {"Update Node Self Status", controllers.UpdateStatusOfNode, POST, false},    // 更新状态，可以设置隐藏
		"/node/update/parent": {"Update Node Self Parent", controllers.UpdateParentOfNode, POST, false},    // 这个接口不如下面这个全功能的接口
		"/node/sort":          {"Sort Node Self", controllers.SortNode, POST, false},                       // 拖曳超级函数
		"/node/delete":        {"Delete Node Self", controllers.DeleteNode, POST, false},

		"/node/take":                {"Take Node Self", controllers.TakeNode, GP, false}, //  和前端的那部分一毛一样
		"/node/list":                {"List Node Self", controllers.ListNode, GP, false},
		"/node/admin/list":          {"List Node All", controllers.ListNodeAdmin, GP, true},                      // 管理员查看其他用户节点
		"/node/admin/update/status": {"Update Node All Status", controllers.UpdateStatusOfNodeAdmin, POST, true}, // 管理员隐藏/显示任意用户的节点

		// 内容操作
		"/content/create":              {"Create Content Self", controllers.CreateContent, POST, false},                             // 创建文章内容(必须归属一个节点)
		"/content/update/seo":          {"Update Content Self Seo", controllers.UpdateSeoOfContent, POST, false},                    // 更新内容SEO
		"/content/update/image":        {"Update Content Self Image", controllers.UpdateImageOfContent, POST, false},                // 更新内容图片
		"/content/update/status":       {"Update Content Self Status", controllers.UpdateStatusOfContent, POST, false},              // 更新内容的状态，如设置隐藏
		"/content/admin/update/status": {"Update Content All Status", controllers.UpdateStatusOfContentAdmin, POST, true},           // 超级管理员修改文章，比如禁用或者逻辑删除/恢复文章
		"/content/update/node":         {"Update Content Self Node", controllers.UpdateNodeOfContent, POST, false},                  // 更改内容的节点，顺便需要重新排序
		"/content/update/top":          {"Update Content Self Top", controllers.UpdateTopOfContent, POST, false},                    // 设置内容的置顶与否
		"/content/update/comment":      {"Update Content Self Comment", controllers.UpdateCommentOfContent, POST, false},            // 设置内容可以评论与否
		"/content/update/password":     {"Update Content Self Password", controllers.UpdatePasswordOfContent, POST, false},          // 更改内容的密码保护
		"/content/update/info":         {"Update Content Self Info", controllers.UpdateInfoOfContent, POST, false},                  // 更新内容标题和内容
		"/content/sort":                {"Sort Content Self", controllers.SortContent, POST, false},                                 // 对内容进行拖曳排序
		"/content/publish":             {"Publish Content Self", controllers.PublishContent, POST, false},                           // 将预览刷进另外一个字段
		"/content/restore":             {"Restore Content Self", controllers.RestoreContent, POST, false},                           // 恢复历史，刷回来
		"/content/rubbish":             {"Sent Content Self To Rubbish", controllers.SentContentToRubbish, POST, false},             // 软删除：移入回收站（status=3）
		"/content/recycle":             {"Sent Rubbish Content Self To Origin", controllers.ReCycleOfContentInRubbish, POST, false}, // 回收站恢复
		"/content/delete":              {"Delete Content Self Real", controllers.ReallyDeleteContent, POST, false},                  // 彻底删除（软删 status=4，仅用户世界消失，管理员可见）
		"/content/take":                {"Take Content Self", controllers.TakeContent, GP, false},                                   // 获取文章内容
		"/content/admin/take":          {"Take Content Admin", controllers.TakeContentAdmin, GP, true},                              // 管理员获取文章内容
		"/content/history/take":        {"Take Content History Self", controllers.TakeContentHistory, GP, false},                    // 获取文章历史内容
		"/content/history/admin/take":  {"Take Content History Admin", controllers.TakeContentHistoryAdmin, GP, true},               // 管理员获取文章历史内容
		"/content/list":                {"List Content Self", controllers.ListContent, GP, false},                                   // 列出文章
		"/content/admin/list":          {"List Content All", controllers.ListContentAdmin, GP, true},                                // 管理员列出文章，什么类型都可以
		"/content/history/list":        {"List Content History Self", controllers.ListContentHistory, GP, false},                    // 列出文章的历史记录
		"/content/history/admin/list":  {"List Content History All", controllers.ListContentHistoryAdmin, GP, true},                 // 管理员列出文章的历史纪录
		"/content/history/delete":      {"Delete Content History Self Real", controllers.ReallyDeleteHistoryContent, POST, false},   // 真删除历史内容
		"/content/cool":                {"Cool the Content Self", controllers.CoolContent, GP, false},                               // 点赞内容
		"/content/bad":                 {"Bad the Content Self", controllers.BadContent, GP, false},                                 // 举报内容
		"/comment/create":              {"Create the Comment Self", controllers.CreateComment, POST, false},                         // 创建评论
		"/comment/real/name":           {"Real Name the Comment Self", controllers.RealNameComment, POST, false},                    // 评论取消匿名
		"/comment/delete":              {"Delete the Comment Self", controllers.DeleteComment, POST, false},                         // 删除评论，逻辑删除
		"/comment/take":                {"Take the Comment Self", controllers.TakeComment, GP, false},                               // 获取评论
		"/comment/cool":                {"Cool the Comment Self", controllers.CoolComment, GP, false},                               // 点赞评论
		"/comment/bad":                 {"Bad the Comment Self", controllers.BadComment, GP, false},                                 // 举报评论
		"/comment/admin/list":          {"List the Comment Admin", controllers.ListComment, GP, true},                               // 管理员列出评论
		"/comment/admin/update/status": {"Update the Comment Status Admin", controllers.UpdateComment, GP, true},                    // 管理员评论违禁处理
		"/content/admin/bad/list":      {"List Content Bad Admin", controllers.ListContentBadAdmin, GP, true},                       // 管理员列出内容举报
		"/comment/admin/bad/list":      {"List Comment Bad Admin", controllers.ListCommentBadAdmin, GP, true},                       // 管理员列出评论举报

		"/relation/follow/add":     {"Follow add Who", controllers.AddRelation, GP, false},                    // 关注
		"/relation/follow/minute":  {"Follow Minute Who", controllers.MinuteRelation, GP, false},              // 关注解除
		"/relation/followed/me":    {"List Who Follow You", controllers.ListFollowedRelationOfMe, GP, false},  // 查看谁关注了你
		"/relation/following/me":   {"List You Follow Who", controllers.ListFollowingRelationOfMe, GP, false}, // 查看你关注了谁
		"/relation/followed/list":  {"List Who Follow You", controllers.ListFollowedRelation, GP, false},      // 查看谁关注了用户B，然后你和谁的关系
		"/relation/following/list": {"List You Follow Who", controllers.ListFollowingRelation, GP, false},     // 查看用户A关注了谁，然后你和谁的关系
		"/relation/admin/list":     {"List Who Follow Who Admin", controllers.ListAllRelation, GP, true},      // 查看所有关系

		"/message/list":                       {"List Your Message Can Include Private Message", controllers.ListMessage, GP, false}, // 列出自己的系统消息，包括与其他用户间的私信
		"/message/admin/list":                 {"List All Message", controllers.ListAllMessage, GP, true},                            // 管理员列出所有系统消息
		"/message/read":                       {"Read Your Message", controllers.ReadMessage, GP, false},                             // 读取系统消息
		"/message/delete":                     {"Delete Your Message", controllers.DeleteMessage, GP, false},                         // 删除系统消息
		"/message/admin/global/create":        {"Admin Create Global Message", controllers.CreateGlobalMessage, GP, true},            // 管理员创建全局站内信
		"/message/admin/global/list":          {"Admin List Global Message", controllers.ListGlobalMessage, GP, true},                // 管理员列出全局站内信
		"/message/admin/global/update/status": {"Admin Change Global Message Status", controllers.UpdateGlobalMessage, GP, true},     // 管理员更改全局站内信状态

		"/message/private/send":   {"Send Message To Private People", controllers.SendPrivateMessage, GP, false}, // 私信
		"/message/private/delete": {"Delete Message Has Sent", controllers.DeletePrivateMessage, GP, false},      // 删除自己发出的私信，但收件方还是可以看到

		// 站点配置 + 友情链接（管理）
		"/site/config/update": {"Update Site Config", controllers.UpdateSiteConfig, GP, true},
		"/friend/list":        {"List Friend Link", controllers.ListFriendLinkAdmin, GP, true},
		"/friend/create":      {"Create Friend Link", controllers.CreateFriendLink, POST, true},
		"/friend/update":      {"Update Friend Link", controllers.UpdateFriendLink, POST, true},
		"/friend/delete":      {"Delete Friend Link", controllers.DeleteFriendLink, POST, true},
		"/friend/sort":        {"Sort Friend Link", controllers.SortFriendLink, POST, true},
	}
)

// SetRouter home end
func SetRouter(router *gin.Engine) {
	for url, app := range HomeRouter {
		for _, method := range app.Method {
			router.Handle(method, url, app.Func)
		}
	}
}

func SetAPIRouter(router *gin.RouterGroup, handles map[string]HttpHandle) {
	for url, app := range handles {
		for _, method := range app.Method {
			router.Handle(method, url, app.Func)
		}
	}
}
