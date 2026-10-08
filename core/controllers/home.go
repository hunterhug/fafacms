package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hunterhug/fafacms/core/config"
	"github.com/hunterhug/fafacms/core/model"
	"github.com/hunterhug/fafacms/core/util"
	log "github.com/hunterhug/golog"
	"math"
	"sort"
	"strings"
	"time"
)

var TimeZone int64 = 0

// ExcerptRunes 列表页正文摘要按 rune 截断的最大字符数
const ExcerptRunes = 200

// BuildExcerpt Make list page content excerpt.
// Truncate by rune instead of byte, otherwise multi-byte characters (Chinese etc.)
// would be cut into broken bytes. And drop the dangling half HTML tag left by
// truncation (such as `<img src="...` without closing `>`), otherwise clients
// would render half of an image url as plain text.
func BuildExcerpt(describe string, maxRunes int) string {
	if maxRunes <= 0 {
		return describe
	}

	runes := []rune(describe)
	if len(runes) <= maxRunes {
		return describe
	}

	s := string(runes[:maxRunes])
	if i := strings.LastIndex(s, "<"); i > strings.LastIndex(s, ">") {
		s = s[:i]
	}

	return s
}

// GetSecond2DateTimes Local time format you know
func GetSecond2DateTimes(second int64) string {
	second = second + 3600*TimeZone
	tm := time.Unix(second, 0)
	return tm.UTC().Format("2006-01-02 15:04:05")

}

// optionalRequester 宽松取当前请求者：公开接口不要求登录，取不到就当作游客。
func optionalRequester(c *gin.Context) *model.User {
	u, err := GetUserSession(c)
	if err != nil {
		return nil
	}
	return u
}

// isAuthorViewer 判断当前请求者是否就是该作者本人。
// 节点隐藏只对作者本人放行，管理员不特殊对待（管理员要看内容用后台列表）。
func isAuthorViewer(c *gin.Context, authorId int64) bool {
	if authorId == 0 {
		return false
	}
	u := optionalRequester(c)
	return u != nil && u.Id == authorId
}

// nodeStatusRow 只取节点表里不加密的列，避免为了统计把加密的 name/describe 也解密一遍。
type nodeStatusRow struct {
	Id           int64 `xorm:"id"`
	ParentNodeId int64 `xorm:"parent_node_id"`
	Status       int   `xorm:"status"`
}

// hiddenNodeIds 返回「隐藏节点」（status != 0）及其所有后代节点的 id。
// authorId 为 0 表示不限用户（首页/发现页这类不带作者过滤的公开列表）。
// 语义：这些节点下的文章退出公开列表，只有作者本人仍能看到。
func hiddenNodeIds(authorId int64) ([]int64, error) {
	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	rows := make([]nodeStatusRow, 0)
	q := session.Table(new(model.ContentNode)).Select("id, parent_node_id, status")
	if authorId != 0 {
		q = q.Where("user_id=?", authorId)
	}
	if err := q.Find(&rows); err != nil {
		return nil, err
	}

	hidden := make(map[int64]struct{})
	for _, r := range rows {
		if r.Status != 0 {
			hidden[r.Id] = struct{}{}
		}
	}

	// 父节点被隐藏时，其子节点（以及更深层级）下的文章同样隐藏。
	// 节点层级很浅（当前最多两级），循环收敛很快。
	for i := 0; i < 8; i++ {
		changed := false
		for _, r := range rows {
			if _, ok := hidden[r.Id]; ok {
				continue
			}
			if _, ok := hidden[r.ParentNodeId]; ok {
				hidden[r.Id] = struct{}{}
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	ids := make([]int64, 0, len(hidden))
	for id := range hidden {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids, nil
}

// countVisibleContentOfUser 按公开列表口径统计某作者对当前访问者可见的文章数：
// 与 Contents 完全一致（文章状态过滤 + 排除隐藏节点及其子节点）。
func countVisibleContentOfUser(authorId int64, hidden []int64) (int64, error) {
	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	q := session.Table(new(model.Content)).Where("user_id=?", authorId).
		And("status!=?", 1).And("status!=?", 2).And("status!=?", 3).And("status!=?", 4).And("version>?", 0)
	if len(hidden) > 0 {
		q = q.NotIn("node_id", hidden)
	}
	return q.Count()
}

func Home(c *gin.Context) {
	resp := new(Resp)
	resp.Flag = true
	resp.Data = "FaFa CMS: https://github.com/hunterhug/fafacms Version:" + config.Version
	defer func() {
		c.JSON(200, resp)
	}()
}

type People struct {
	Id                    int64  `json:"id"`
	Name                  string `json:"name"`
	NickName              string `json:"nick_name"`
	Email                 string `json:"email,omitempty"` // 仅本人接口（/user/info）返回；公开接口不填充
	WeChat                string `json:"wechat"`
	WeiBo                 string `json:"weibo"`
	Github                string `json:"github"`
	QQ                    string `json:"qq"`
	Gender                int    `json:"gender"`
	Describe              string `json:"describe"`
	ShortDescribe         string `json:"short_describe"`
	HeadPhoto             string `json:"head_photo"`
	CreateTime            string `json:"create_time"`
	CreateTimeInt         int64  `json:"create_time_int"`
	UpdateTime            string `json:"update_time,omitempty"`
	UpdateTimeInt         int64  `json:"update_time_int,omitempty"`
	ActivateTime          string `json:"activate_time,omitempty"`
	ActivateTimeInt       int64  `json:"activate_time_int,omitempty"`
	LoginTime             string `json:"login_time,omitempty"`
	LoginTimeInt          int64  `json:"login_time_int,omitempty"`
	LoginIp               string `json:"login_ip,omitempty"`
	NickNameUpdateTimeInt int64  `json:"nick_name_update_time,omitempty"`
	NickNameUpdateTime    string `json:"nick_name_update_time,omitempty"`
	IsInBlack             bool   `json:"is_in_black"`
	IsVip                 bool   `json:"is_vip"`
	TwoFa                 bool   `json:"two_fa,omitempty"`
	FollowedNum           int64  `json:"followed_num"`
	FollowingNum          int64  `json:"following_num"`
	ContentNum            int64  `json:"content_num"`      // normal publish content num
	ContentCoolNum        int64  `json:"content_cool_num"` // normal content cool num
}

type PeoplesRequest struct {
	Vip      int      `json:"vip" validate:"oneof=-1 0 1"`
	NickName string   `json:"nick_name"` // 公开搜用户：只按昵称模糊匹配
	Sort     []string `json:"sort"`
	PageHelp
}

type PeoplesResponse struct {
	Users []People `json:"users"`
	PageHelp
}

func Peoples(c *gin.Context) {
	resp := new(Resp)

	defer func() {
		JSON(c, 200, resp)
	}()

	respResult := new(PeoplesResponse)
	req := new(PeoplesRequest)
	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		log.Errorf("Peoples err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	session.Table(new(model.User)).Where("1=1").And("status!=?", 0)
	// 推荐列表不显示超管，但搜索时（nick_name 非空）允许搜到
	if req.NickName == "" {
		session.And("name!=?", "admin")
	}

	if req.Vip != -1 {
		if req.Vip == 0 {
			session.And("vip=?", 0)
		} else {
			session.And("vip=?", 1)
		}
	}

	// 昵称列已加密，无法模糊匹配：改为按盲索引精确匹配
	if req.NickName != "" {
		bidx, berr := model.BlindIndexOf(req.NickName)
		if berr != nil {
			log.Errorf("Peoples err: %s", berr.Error())
			resp.Error = Error(DBError, berr.Error())
			return
		}
		session.And("nick_name_bidx=?", bidx)
	}

	users := make([]model.User, 0)
	p := &req.PageHelp

	p.build(session, req.Sort, model.UserSortName)
	total, err := session.FindAndCount(&users)
	if err != nil {
		log.Errorf("Peoples err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	peoples := make([]People, 0, len(users))
	for _, v := range users {
		p := People{}
		p.Id = v.Id
		p.ShortDescribe = v.ShortDescribe
		p.Describe = v.Describe
		p.CreateTimeInt = v.CreateTime
		p.CreateTime = GetSecond2DateTimes(v.CreateTime)

		p.UpdateTimeInt = v.UpdateTime
		if v.UpdateTime > 0 {
			p.UpdateTime = GetSecond2DateTimes(v.UpdateTime)
		}

		p.ActivateTimeInt = v.ActivateTime
		if v.ActivateTime > 0 {
			p.ActivateTime = GetSecond2DateTimes(v.ActivateTime)
		}

		p.LoginTimeInt = v.LoginTime
		if v.LoginTime > 0 {
			p.LoginTime = GetSecond2DateTimes(v.LoginTime)
		}

		if v.Status == 2 {
			p.IsInBlack = true
		}
		// 安全：公开接口不返回邮箱（仅本人可通过 /user/info 查看自己的邮箱）
		p.Github = v.Github
		p.Name = v.Name
		p.NickName = v.NickName
		p.HeadPhoto = v.HeadPhoto
		p.QQ = v.QQ
		p.WeChat = v.WeChat
		p.WeiBo = v.WeiBo
		p.Gender = v.Gender
		p.IsVip = v.Vip == 1
		p.FollowedNum = v.FollowedNum
		p.FollowingNum = v.FollowingNum
		p.ContentNum = v.ContentNum
		p.ContentCoolNum = v.ContentCoolNum
		peoples = append(peoples, p)
	}
	respResult.Users = peoples
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type Node struct {
	Id            int64  `json:"id"`
	Seo           string `json:"seo"`
	Name          string `json:"name"`
	Describe      string `json:"describe"`
	ImagePath     string `json:"image_path"`
	CreateTime    string `json:"create_time"`
	CreateTimeInt int64  `json:"create_time_int"`
	UpdateTime    string `json:"update_time,omitempty"`
	UpdateTimeInt int64  `json:"update_time_int,omitempty"`
	UserId        int64  `json:"user_id"`
	UserName      string `json:"user_name"`
	SortNum       int64  `json:"sort_num"`
	Level         int    `json:"level"`
	Status        int    `json:"status"`
	ParentNodeId  int64  `json:"parent_node_id"`
	Son           []Node `json:"son,omitempty"`
	ContentNum    int64  `json:"content_num"` // normal publish content num
}

type NodesInfoRequest struct {
	UserId   int64    `json:"user_id"`
	UserName string   `json:"user_name"`
	Sort     []string `json:"sort"`
}

type NodesResponse struct {
	Nodes []Node `json:"nodes"`
}

func NodesInfo(c *gin.Context) {
	resp := new(Resp)

	defer func() {
		JSON(c, 200, resp)
	}()

	respResult := new(NodesResponse)
	req := new(NodesInfoRequest)
	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.UserId == 0 && req.UserName == "" {
		log.Errorf("NodesInfo err:%s", "")
		resp.Error = Error(ParasError, "user info empty")
		return
	}

	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	session.Table(new(model.ContentNode)).Where("1=1").And("status=?", 0)

	if req.UserId != 0 {
		session.And("user_id=?", req.UserId)
	}

	if req.UserName != "" {
		uid, uerr := userFilterId(req.UserName)
		if uerr != nil {
			log.Errorf("NodesInfo err: %s", uerr.Error())
			resp.Error = Error(DBError, uerr.Error())
			return
		}
		session.And("user_id=?", uid)
	}

	nodes := make([]model.ContentNode, 0)
	Build(session, req.Sort, model.ContentNodeSortName)
	err := session.Find(&nodes)
	if err != nil {
		log.Errorf("NodesInfo err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	father := make([]model.ContentNode, 0)
	son := make([]model.ContentNode, 0)
	for _, v := range nodes {
		if v.Level == 0 {
			father = append(father, v)
		} else {
			son = append(son, v)
		}
	}

	n := make([]Node, 0)
	for _, v := range father {
		f := Node{}
		f.Id = v.Id
		f.Seo = v.Seo
		f.Describe = v.Describe
		f.ImagePath = v.ImagePath
		f.Name = v.Name
		if v.UpdateTime > 0 {
			f.UpdateTime = GetSecond2DateTimes(v.UpdateTime)
			f.UpdateTimeInt = v.UpdateTime
		}
		f.CreateTime = GetSecond2DateTimes(v.CreateTime)
		f.CreateTimeInt = v.CreateTime
		f.SortNum = v.SortNum
		f.UserName = v.UserName
		f.UserId = v.UserId
		f.Level = v.Level
		f.ParentNodeId = v.ParentNodeId
		f.ContentNum = CountContentNumOfNode(v.UserId, v.Id, true)
		for _, vv := range son {
			if vv.ParentNodeId == f.Id {
				s := Node{}
				s.Id = vv.Id
				s.Seo = vv.Seo
				s.Describe = vv.Describe
				s.ImagePath = vv.ImagePath
				s.Name = vv.Name
				if vv.UpdateTime > 0 {
					s.UpdateTimeInt = vv.UpdateTime
					s.UpdateTime = GetSecond2DateTimes(vv.UpdateTime)
				}
				s.CreateTime = GetSecond2DateTimes(vv.CreateTime)
				s.CreateTimeInt = vv.CreateTime
				s.SortNum = vv.SortNum
				s.UserId = vv.UserId
				s.UserName = vv.UserName
				s.Level = vv.Level
				s.ContentNum = CountContentNumOfNode(vv.UserId, vv.Id, true)
				s.ParentNodeId = vv.ParentNodeId
				f.Son = append(f.Son, s)
			}
		}

		n = append(n, f)
	}

	respResult.Nodes = n
	resp.Flag = true
	resp.Data = respResult
}

type NodeInfoRequest struct {
	Id       int    `json:"id"`
	UserId   int    `json:"user_id"`
	UserName string `json:"user_name"`
	Seo      string `json:"seo"`
	ListSon  bool   `json:"list_son"`
}

func NodeInfo(c *gin.Context) {
	resp := new(Resp)

	defer func() {
		JSON(c, 200, resp)
	}()

	req := new(NodeInfoRequest)
	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.Id == 0 && req.Seo == "" {
		log.Errorf("NodeInfo err: %s", "content node id or seo empty")
		resp.Error = Error(ParasError, "content node id or seo empty")
		return
	}

	if req.Id == 0 && req.Seo != "" {
		if req.UserId == 0 && req.UserName == "" {
			log.Errorf("NodeInfo err: %s", "content node seo exist but user info empty")
			resp.Error = Error(ParasError, "content node seo exist but user info empty")
			return
		}
	}

	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	// 归属用户：节点 SEO 只在同一用户内唯一，所以按 SEO 查询必须带用户名/ID
	ownerId := int64(req.UserId)
	if ownerId == 0 && req.UserName != "" {
		uid, uerr := userFilterId(req.UserName)
		if uerr != nil {
			log.Errorf("NodeInfo err: %s", uerr.Error())
			resp.Error = Error(DBError, uerr.Error())
			return
		}
		ownerId = uid
	}

	// 隐藏节点（status=1）对公众不可见；节点作者本人仍能取到自己的节点，
	// 用于文章详情页显示节点名字（否则只能退回显示 SEO 字符串）。管理员不特殊对待。
	requester, _ := GetUserSession(c)
	isOwner := requester != nil && ownerId != 0 && requester.Id == ownerId

	session.Table(new(model.ContentNode)).Where("1=1")
	if !isOwner {
		session.And("status=?", 0)
	}

	if ownerId != 0 {
		session.And("user_id=?", ownerId)
	}

	if req.Id != 0 {
		session.And("id=?", req.Id)
	}

	if req.Seo != "" {
		session.And("seo=?", req.Seo)
	}

	v := new(model.ContentNode)
	exist, err := session.Get(v)
	if err != nil {
		log.Errorf("NodeInfo err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("NodeInfo err:%s", "content node not found")
		resp.Error = Error(ContentNodeNotFound, "")
		return
	}

	f := Node{}
	f.Id = v.Id
	f.Seo = v.Seo
	f.Describe = v.Describe
	f.ImagePath = v.ImagePath
	f.Name = v.Name
	if v.UpdateTime > 0 {
		f.UpdateTime = GetSecond2DateTimes(v.UpdateTime)
		f.UpdateTimeInt = v.UpdateTime
	}
	f.CreateTime = GetSecond2DateTimes(v.CreateTime)
	f.CreateTimeInt = v.CreateTime
	f.SortNum = v.SortNum
	f.UserName = v.UserName
	f.UserId = v.UserId
	f.Level = v.Level
	f.ParentNodeId = v.ParentNodeId
	f.ContentNum = v.ContentNum

	// 是顶层且需要列出儿子
	if f.Level == 0 && req.ListSon {
		ns := make([]model.ContentNode, 0)
		err = model.FaFaRdb.Client.Where("parent_node_id=?", f.Id).And("status=?", 0).Find(&ns)
		if err != nil {
			log.Errorf("NodeInfo err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}

		for _, vv := range ns {
			ff := Node{}
			ff.Id = vv.Id
			ff.Seo = vv.Seo
			ff.Describe = vv.Describe
			ff.ImagePath = vv.ImagePath
			ff.Name = vv.Name
			if vv.UpdateTime > 0 {
				ff.UpdateTime = GetSecond2DateTimes(vv.UpdateTime)
				ff.UpdateTimeInt = vv.UpdateTime
			}
			ff.CreateTime = GetSecond2DateTimes(vv.CreateTime)
			ff.CreateTimeInt = vv.CreateTime
			ff.SortNum = vv.SortNum
			ff.UserName = vv.UserName
			ff.UserId = vv.UserId
			ff.Level = vv.Level
			ff.ParentNodeId = vv.ParentNodeId
			ff.ContentNum = vv.ContentNum
			f.Son = append(f.Son, ff)
		}
	}
	resp.Flag = true
	resp.Data = f
}

type UserInfoRequest struct {
	Id   int64  `json:"user_id"`
	Name string `json:"user_name"`
}

func UserInfo(c *gin.Context) {
	resp := new(Resp)

	defer func() {
		JSON(c, 200, resp)
	}()

	req := new(UserInfoRequest)
	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.Id == 0 && req.Name == "" {
		resp.Error = Error(ParasError, "where is empty")
		return
	}

	user := new(model.User)
	user.Id = req.Id
	user.Name = req.Name
	// name 列已加密：必须走 GetActivateRaw（内部先 PrepareSearch 把明文转成盲索引条件），
	// 直接用明文列做条件将永远匹配不上。
	exist, err := user.GetActivateRaw()
	if err != nil {
		log.Errorf("UserInfo err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("UserInfo err:%s", "user  not found")
		resp.Error = Error(UserNotFound, "")
		return
	}

	v := user
	p := People{}
	p.Id = v.Id
	p.Describe = v.Describe
	p.ShortDescribe = v.ShortDescribe
	p.CreateTime = GetSecond2DateTimes(v.CreateTime)
	p.CreateTimeInt = v.CreateTime

	if v.Status == 2 {
		p.IsInBlack = true
	}

	p.UpdateTimeInt = v.UpdateTime
	if v.UpdateTime > 0 {
		p.UpdateTime = GetSecond2DateTimes(v.UpdateTime)
	}

	p.LoginTimeInt = v.LoginTime
	if v.LoginTime > 0 {
		p.LoginTime = GetSecond2DateTimes(v.LoginTime)
	}

	p.ActivateTimeInt = v.ActivateTime
	if v.ActivateTime > 0 {
		p.ActivateTime = GetSecond2DateTimes(v.ActivateTime)
	}
	// 安全：公开接口不返回邮箱（仅本人可通过 /user/info 查看自己的邮箱）
	p.Github = v.Github
	p.Name = v.Name
	p.NickName = v.NickName
	p.HeadPhoto = v.HeadPhoto
	p.QQ = v.QQ
	p.WeChat = v.WeChat
	p.WeiBo = v.WeiBo
	p.Gender = v.Gender
	p.FollowingNum = v.FollowingNum
	p.FollowedNum = v.FollowedNum
	p.ContentNum = v.ContentNum
	// 文章数口径与公开列表一致，并且不能直接用用户表的缓存列：
	// 缓存列由 CountContentAll 重算，而它只统计「未隐藏节点」下的文章，
	// 所以隐藏节点后作者自己看到的数字也会跟着变小。
	// 这里实时统计：非作者视角排除隐藏节点（与列表一致），
	// 作者本人看自己的全部已发布文章（含隐藏节点，与自己内容管理页一致）。
	filterHidden := []int64(nil)
	countOk := true
	if !isAuthorViewer(c, v.Id) {
		if h, herr := hiddenNodeIds(v.Id); herr != nil {
			countOk = false
		} else {
			filterHidden = h
		}
	}
	if countOk {
		if num, cerr := countVisibleContentOfUser(v.Id, filterHidden); cerr == nil {
			p.ContentNum = num
		}
	}
	p.ContentCoolNum = v.ContentCoolNum
	p.IsVip = v.Vip == 1
	resp.Flag = true
	resp.Data = p
}

type UserCountRequest struct {
	UserId   int64  `json:"user_id"`
	UserName string `json:"user_name"`
}

type UserCountX struct {
	Count           int    `json:"count"`
	Days            string `json:"days"`
	CreateTimeBegin int64  `json:"first_publish_time_begin"`
	CreateTimeEnd   int64  `json:"first_publish_time_end"`
}
type UserCountResponse struct {
	Info     []UserCountX `json:"info"`
	UserId   int64        `json:"user_id"`
	UserName string       `json:"user_name"`
}

func UserCount(c *gin.Context) {
	resp := new(Resp)

	defer func() {
		JSON(c, 200, resp)
	}()

	req := new(UserCountRequest)
	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	if req.UserId == 0 && req.UserName == "" {
		resp.Error = Error(ParasError, "where is empty")
		return
	}

	user := new(model.User)
	user.Id = req.UserId
	user.Name = req.UserName
	user.Status = 1
	exist, err := user.GetRaw()
	if err != nil {
		log.Errorf("UserCount err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("UserCount err:%s", "user not found")
		resp.Error = Error(UserNotFound, "")
		return
	}

	req.UserId = user.Id

	// 与公开列表口径一致：排除 1 隐藏 / 2 封禁 / 3 回收站 / 4 彻底删除，
	// 并且非作者本人视角还要排除隐藏节点（含其子节点）下的文章。
	hiddenSql := ""
	args := make([]interface{}, 0, 8)
	if !isAuthorViewer(c, user.Id) {
		hidden, herr := hiddenNodeIds(user.Id)
		if herr != nil {
			log.Errorf("UserCount err:%s", herr.Error())
			resp.Error = Error(DBError, herr.Error())
			return
		}
		if len(hidden) > 0 {
			ph := make([]string, 0, len(hidden))
			for _, id := range hidden {
				ph = append(ph, "?")
				args = append(args, id)
			}
			hiddenSql = " and node_id not in (" + strings.Join(ph, ",") + ")"
		}
	}

	args = append(args, req.UserId)
	sql := fmt.Sprintf("SELECT DATE_FORMAT(from_unixtime(first_publish_time + %d * 3600)", TimeZone) + ",'%Y%m%d') as days,count(id) as count FROM `fafacms_content` WHERE first_publish_time!=0 and version>0 and status!=1 and status!=2 and status!=3 and status!=4" + hiddenSql + " and user_id=? group by days;"
	result, err := model.FaFaRdb.Client.QueryString(sql, args)
	if err != nil {
		log.Errorf("UserCount err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	back := make([]UserCountX, 0)
	for _, v := range result {
		t := UserCountX{}
		t.Count, _ = util.SI(v["count"])
		t.Days = v["days"]
		begin, _ := time.ParseInLocation("20060102", t.Days, time.UTC)
		begin = begin.Add(time.Second * time.Duration(3600*TimeZone))
		end := begin.AddDate(0, 0, 1)
		t.CreateTimeBegin = begin.Unix()
		t.CreateTimeEnd = end.Unix()
		back = append(back, t)
	}

	resp.Flag = true
	resp.Data = UserCountResponse{
		Info:     back,
		UserId:   user.Id,
		UserName: user.Name,
	}
}

type ContentsRequest struct {
	NodeId                int64    `json:"node_id"`
	IncludeChildren       bool     `json:"include_children"`
	NodeSeo               string   `json:"node_seo"`
	UserId                int64    `json:"user_id"`
	UserName              string   `json:"user_name"`
	Title                 string   `json:"title"`
	FirstPublishTimeBegin int64    `json:"first_publish_time_begin"`
	FirstPublishTimeEnd   int64    `json:"first_publish_time_end"`
	PublishTimeBegin      int64    `json:"publish_time_begin"`
	PublishTimeEnd        int64    `json:"publish_time_end"`
	Sort                  []string `json:"sort"`
	PageHelp
}

type ContentsX struct {
	Id                  int64      `json:"id"`
	Seo                 string     `json:"seo"`
	Title               string     `json:"title"`
	UserId              int64      `json:"user_id"`
	UserName            string     `json:"user_name"`
	UserNickName        string     `json:"user_nick_name,omitempty"`
	UserHeadPhoto       string     `json:"user_head_photo,omitempty"`
	NodeId              int64      `json:"node_id"`
	NodeSeo             string     `json:"node_seo"`
	NodeHidden          bool       `json:"node_hidden,omitempty"` // 所属节点已隐藏（仅作者本人能看到这类文章）
	Top                 int        `json:"top"`
	FirstPublishTime    string     `json:"first_publish_time"`
	PublishTime         string     `json:"publish_time,omitempty"`
	FirstPublishTimeInt int64      `json:"first_publish_time_int"`
	PublishTimeInt      int64      `json:"publish_time_int"`
	ImagePath           string     `json:"image_path"`
	Views               int64      `json:"views"`
	IsLock              bool       `json:"is_lock"`
	Describe            string     `json:"describe"`
	Next                *ContentsX `json:"next,omitempty"`
	Pre                 *ContentsX `json:"pre,omitempty"`
	SortNum             int64      `json:"sort_num"`
	Bad                 int64      `json:"bad"`
	Cool                int64      `json:"cool"`
	CommentNum          int64      `json:"comment_num"`
	CloseComment        int        `json:"close_comment"`
	IsBan               bool       `json:"is_ban"`
}

type ContentsResponse struct {
	Contents []ContentsX `json:"contents"`
	PageHelp
}

func Contents(c *gin.Context) {
	resp := new(Resp)

	respResult := new(ContentsResponse)
	req := new(ContentsRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		log.Errorf("Contents err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.Content)).Where("1=1")

	// 作者（被查看的那个人），用于隐藏节点过滤与作者本人视角判断
	authorId := req.UserId

	if req.UserId != 0 {
		session.And("user_id=?", req.UserId)
	}

	if req.UserName != "" {
		uid, uerr := userFilterId(req.UserName)
		if uerr != nil {
			log.Errorf("Contents err: %s", uerr.Error())
			resp.Error = Error(DBError, uerr.Error())
			return
		}
		authorId = uid
		session.And("user_id=?", uid)
	}

	// 公开列表的口径：排除 1 隐藏 / 2 封禁 / 3 回收站 / 4 用户彻底删除。
	// status=4 是软删除，仅从用户世界消失，管理员后台仍可见（见 content.go 状态机）；
	// 这里若漏掉 4，已删除的文章会继续出现在首页，点进去却是空的。
	session.And("status!=?", 1).And("status!=?", 2).And("status!=?", 3).And("status!=?", 4).And("version>?", 0)

	// 隐藏节点（含其子节点）下的文章退出公开列表，只有作者本人仍能看到。
	// authorId 为 0（首页/发现页）时按全站过滤。
	hidden, herr := hiddenNodeIds(authorId)
	if herr != nil {
		log.Errorf("Contents err: %s", herr.Error())
		resp.Error = Error(DBError, herr.Error())
		return
	}
	hiddenSet := make(map[int64]struct{}, len(hidden))
	for _, id := range hidden {
		hiddenSet[id] = struct{}{}
	}
	if !isAuthorViewer(c, authorId) && len(hidden) > 0 {
		// NotIn 已把隐藏节点（含其子节点）下的文章排除干净；
		// 指定节点本身被隐藏时，这里的结果自然就是空列表。
		session.NotIn("node_id", hidden)
	}

	if req.NodeId != 0 {
		if req.IncludeChildren {
			// 一级节点下钻：列出该节点及其二级子节点的文章
			var sons []model.ContentNode
			err := model.FaFaRdb.Client.Cols("id").Where("parent_node_id=?", req.NodeId).Find(&sons)
			if err != nil {
				log.Errorf("Contents err: %s", err.Error())
				resp.Error = Error(DBError, err.Error())
				return
			}
			nodeIds := make([]int64, 0, len(sons)+1)
			nodeIds = append(nodeIds, req.NodeId)
			for _, s := range sons {
				nodeIds = append(nodeIds, s.Id)
			}
			session.In("node_id", nodeIds)
		} else {
			session.And("node_id=?", req.NodeId)
		}
	}

	if req.NodeSeo != "" {
		session.And("node_seo=?", req.NodeSeo)
	}

	if req.Title != "" {
		// 标题列已加密，无法模糊匹配：改为按标题盲索引做精确匹配（模糊搜索已按既定决策取消）
		bidx, berr := model.BlindIndexOf(req.Title)
		if berr != nil {
			log.Errorf("Contents err: %s", berr.Error())
			resp.Error = Error(DBError, berr.Error())
			return
		}
		session.And("title_bidx=?", bidx)
	}

	if req.FirstPublishTimeBegin > 0 {
		session.And("first_publish_time>=?", req.FirstPublishTimeBegin)
	}

	if req.FirstPublishTimeEnd > 0 {
		session.And("first_publish_time<?", req.FirstPublishTimeEnd)
	}

	if req.PublishTimeBegin > 0 {
		session.And("publish_time>=?", req.PublishTimeBegin)
	}

	if req.PublishTimeEnd > 0 {
		session.And("publish_time<?", req.PublishTimeEnd)
	}

	// if count>0 start list
	cs := make([]model.Content, 0)
	p := &req.PageHelp

	// sql build
	p.build(session, req.Sort, model.ContentSortName2)

	// do query
	total, err := session.Omit("pre_describe", "pre_title").FindAndCount(&cs)
	if err != nil {
		log.Errorf("Contents err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// result
	bcs := make([]ContentsX, 0, len(cs))
	for _, c := range cs {
		temp := ContentsX{}
		temp.UserId = c.UserId
		temp.Seo = c.Seo
		temp.SortNum = c.SortNum
		temp.NodeSeo = c.NodeSeo
		temp.UserName = c.UserName
		temp.Id = c.Id
		temp.Top = c.Top
		temp.Title = c.Title
		temp.NodeId = c.NodeId
		temp.Views = c.Views
		temp.ImagePath = c.ImagePath
		temp.FirstPublishTime = GetSecond2DateTimes(c.FirstPublishTime)
		temp.PublishTime = GetSecond2DateTimes(c.PublishTime)
		temp.FirstPublishTimeInt = c.FirstPublishTime
		temp.PublishTimeInt = c.PublishTime
		temp.CommentNum = c.CommentNum
		temp.Bad = c.Bad
		temp.Cool = c.Cool
		temp.CloseComment = c.CloseComment
		if c.Status == 2 {
			temp.IsBan = true
		}

		if c.Password != "" {
			temp.IsLock = true
		}

		// 该文章挂在隐藏节点下：只有作者本人能看到（前端据此打「已隐藏」标记）
		if _, ok := hiddenSet[c.NodeId]; ok {
			temp.NodeHidden = true
		}

		// 正文摘要：按 rune 截断（不能按字节截断，否则中文等多字节字符会被切成乱码），
		// 并丢掉截断处残留的半截 HTML 标签
		temp.Describe = BuildExcerpt(c.Describe, ExcerptRunes)
		bcs = append(bcs, temp)
	}

	// 批量填充作者昵称/头像，避免前端逐作者 N+1 调用 /u/info
	authorIds := make([]int64, 0, len(cs))
	seen := make(map[int64]struct{}, len(cs))
	for _, c := range cs {
		if c.UserId != 0 {
			if _, ok := seen[c.UserId]; !ok {
				seen[c.UserId] = struct{}{}
				authorIds = append(authorIds, c.UserId)
			}
		}
	}
	if len(authorIds) > 0 {
		if userMap, err := model.GetUser(authorIds); err == nil {
			for i := range bcs {
				if u, ok := userMap[bcs[i].UserId]; ok {
					bcs[i].UserNickName = u.NickName
					bcs[i].UserHeadPhoto = u.HeadPhoto
				}
			}
		}
	}

	respResult.Contents = bcs
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

type ContentRequest struct {
	Id       int64  `json:"id"`
	UserId   int64  `json:"user_id"`
	UserName string `json:"user_name"`
	NodeSeo  string `json:"node_seo"`
	Seo      string `json:"seo"`
	Password string `json:"password"`
	More     bool   `json:"more"`
}

func Content(c *gin.Context) {
	resp := new(Resp)
	req := new(ContentRequest)
	defer func() {
		JSONL(c, 200, req, resp)
	}()

	if errResp := ParseJSON(c, req); errResp != nil {
		resp.Error = errResp
		return
	}

	var validate = validator.New()
	err := validate.Struct(req)
	if err != nil {
		log.Errorf("Content err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	if req.Id == 0 && req.Seo == "" {
		log.Errorf("Content err: %s", "content id or seo empty")
		resp.Error = Error(ParasError, "content id or seo empty")
		return
	}

	if req.Id == 0 && req.Seo != "" {
		if req.UserId == 0 && req.UserName == "" {
			log.Errorf("Content err: %s", "content seo exist but user info empty")
			resp.Error = Error(ParasError, "content seo exist but user info empty")
			return
		}
		// 文章 SEO 只在所属节点内唯一，所以按 SEO 查询必须同时给出节点 SEO。
		// 地址是 /u/<用户名>/node/<节点SEO>/<文章SEO>，三者一起才能唯一定位。
		if req.NodeSeo == "" {
			log.Errorf("Content err: %s", "content node_seo empty")
			resp.Error = Error(ParasError, "content node_seo empty")
			return
		}
	}

	content := new(model.Content)
	content.Id = req.Id
	content.UserId = req.UserId
	// content.user_name 列已加密，不能直接作为等值条件：把用户名解析成 user_id 再过滤。
	if content.UserId == 0 && req.UserName != "" {
		uid, uerr := userFilterId(req.UserName)
		if uerr != nil {
			log.Errorf("Content err: %s", uerr.Error())
			resp.Error = Error(DBError, uerr.Error())
			return
		}
		content.UserId = uid
	}

	// 节点 SEO 在同一用户内唯一，据此把节点 SEO 解析成 node_id
	if req.Id == 0 && req.NodeSeo != "" {
		node := new(model.ContentNode)
		exist, nerr := model.FaFaRdb.Client.Where("user_id=?", content.UserId).And("seo=?", req.NodeSeo).Get(node)
		if nerr != nil {
			log.Errorf("Content err: %s", nerr.Error())
			resp.Error = Error(DBError, nerr.Error())
			return
		}
		if !exist {
			log.Errorf("Content err: %s", "content node not found")
			resp.Error = Error(ContentNodeNotFound, "")
			return
		}
		content.NodeId = node.Id
	}

	content.Seo = req.Seo
	exist, err := content.GetByRawAll()
	if err != nil {
		log.Errorf("Content err: %s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	if !exist {
		log.Errorf("Content err: %s", "content not found")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if content.Status == 0 {

	} else if content.Status == 2 {
		log.Errorf("Content err: %s", "content ban")
		resp.Error = Error(ContentBanPermit, "")
		return
	} else {
		log.Errorf("Content err: %s", "content not found for it hide")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	if content.Version == 0 {
		log.Errorf("Content err: %s", "content not found for it not publish")
		resp.Error = Error(ContentNotFound, "")
		return
	}

	// 访问密码以 bcrypt 哈希存储，比对时用哈希校验
	if content.Password != "" {
		if ok, _ := model.CheckPassword(content.Password, req.Password); !ok {
			log.Errorf("Content err: %s", "content password")
			resp.Error = Error(ContentPasswordWrong, "")
			return
		}
	}

	cx := content
	temp := ContentsX{}
	temp.UserId = cx.UserId
	temp.Seo = cx.Seo
	temp.NodeSeo = cx.NodeSeo
	temp.UserName = cx.UserName
	temp.Id = cx.Id
	temp.Top = cx.Top
	temp.Title = cx.Title
	temp.NodeId = cx.NodeId
	temp.Views = cx.Views
	temp.SortNum = cx.SortNum
	temp.FirstPublishTime = GetSecond2DateTimes(cx.FirstPublishTime)
	temp.PublishTime = GetSecond2DateTimes(cx.PublishTime)
	temp.FirstPublishTimeInt = cx.FirstPublishTime
	temp.PublishTimeInt = cx.PublishTime
	temp.ImagePath = cx.ImagePath
	temp.CommentNum = cx.CommentNum
	temp.Bad = cx.Bad
	temp.Cool = cx.Cool
	temp.CloseComment = cx.CloseComment
	if cx.Password != "" {
		temp.IsLock = true
	}

	temp.Describe = cx.Describe

	cx.UpdateView()

	if req.More {
		cxx := new(model.Content)
		cxx.SortNum = cx.SortNum
		cxx.NodeId = cx.NodeId
		cxx.Id = cx.Id
		pre, next, err := cxx.GetBrotherContent()

		if err != nil {
			log.Errorf("Content err: %s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		if pre.Id != 0 {
			temp1 := new(ContentsX)
			temp1.UserId = pre.UserId
			temp1.Seo = pre.Seo
			temp1.NodeSeo = pre.NodeSeo
			temp1.UserName = pre.UserName
			temp1.Id = pre.Id
			temp1.Top = pre.Top
			temp1.Title = pre.Title
			temp1.NodeId = pre.NodeId
			temp1.Views = pre.Views
			temp1.SortNum = pre.SortNum
			temp1.FirstPublishTime = GetSecond2DateTimes(pre.FirstPublishTime)
			temp1.PublishTime = GetSecond2DateTimes(pre.PublishTime)
			temp1.FirstPublishTimeInt = pre.FirstPublishTime
			temp1.PublishTimeInt = pre.PublishTime
			temp1.ImagePath = pre.ImagePath
			temp1.CommentNum = pre.CommentNum
			temp1.Bad = pre.Bad
			temp1.Cool = pre.Cool
			if pre.Password != "" {
				temp1.IsLock = true
			}
			temp.Pre = temp1
		}
		if next.Id != 0 {
			temp2 := new(ContentsX)
			temp2.UserId = next.UserId
			temp2.Seo = next.Seo
			temp2.NodeSeo = next.NodeSeo
			temp2.UserName = next.UserName
			temp2.Id = next.Id
			temp2.Top = next.Top
			temp2.Title = next.Title
			temp2.NodeId = next.NodeId
			temp2.Views = next.Views
			temp2.SortNum = next.SortNum
			temp2.FirstPublishTime = GetSecond2DateTimes(next.FirstPublishTime)
			temp2.PublishTime = GetSecond2DateTimes(next.PublishTime)
			temp2.FirstPublishTimeInt = next.FirstPublishTime
			temp2.PublishTimeInt = next.PublishTime
			temp2.ImagePath = next.ImagePath
			temp2.CommentNum = next.CommentNum
			temp2.Bad = next.Bad
			temp2.Cool = next.Cool
			if next.Password != "" {
				temp2.IsLock = true
			}
			temp.Next = temp2
		}
	}
	resp.Flag = true
	resp.Data = temp
}
