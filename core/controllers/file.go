package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/hunterhug/fafacms/core/config"
	"github.com/hunterhug/fafacms/core/model"
	myutil "github.com/hunterhug/fafacms/core/util"
	"github.com/hunterhug/fafacms/core/util/oss"
	"github.com/hunterhug/fafacms/core/util/secure"
	"github.com/hunterhug/go_image"
	log "github.com/hunterhug/golog"
	"io/ioutil"
	"math"
	"path/filepath"
	"strings"
	"time"
)

// randomStorageName 生成随机存储文件名（16 字节随机 → 32 位 hex + 后缀）。
//
// 不再用内容哈希命名：旧命名把 sha256(明文) 直接暴露在文件名与 URL 中，
// 可被用于"确认攻击"（拿到候选文件即可判断某用户是否上传过该文件）。
func randomStorageName(suffix string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b) + "." + suffix, nil
}

var scaleType = []string{"jpg", "jpeg", "png"}
var FileAllow = map[string][]string{
	"image": {
		"jpg", "jpeg", "png", "gif", "webp", "heic", "heif", "bmp"},
	"flash": {
		"swf", "flv"},
	"media": {
		"swf", "flv", "mp3", "wav", "wma", "wmv", "mid", "avi", "mpg", "asf", "rm", "rmvb"},
	"file": {
		"doc", "docx", "xls", "xlsx", "ppt", "htm", "html", "txt", "zip", "rar", "gz", "bz2", "pdf"},
	"other": {
		"jpg", "jpeg", "png", "bmp", "gif", "swf", "flv", "mp3",
		"wav", "wma", "wmv", "mid", "avi", "mpg", "asf", "rm", "rmvb",
		"doc", "docx", "xls", "xlsx", "ppt", "htm", "html", "txt", "zip", "rar", "gz", "bz2"}}

var (
	FileBytes  = 1 << 25 // (1<<25)/1000.0/1000.0 33.54 size can not beyond 33M
	CanScale   = true
	ScaleWidth = 500
)

type UploadResponse struct {
	FileName       string `json:"file_name"`
	ReallyFileName string `json:"really_file_name"`
	Size           int64  `json:"size"`
	Url            string `json:"url"`
	UrlX           string `json:"url_x"`
	IsPicture      bool   `json:"is_picture"`
	Addon          string `json:"addon"`
	Oss            bool   `json:"oss"`
}

/*
file: the binary file of HTML form's name
type: can be: image、flash、media、file、other
describe: some describe of file
*/
func UploadFile(c *gin.Context) {
	resp := new(Resp)
	data := UploadResponse{}
	defer func() {
		JSONL(c, 200, nil, resp)
	}()

	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("upload err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		return
	}

	uName := uu.Name

	fileType := c.DefaultPostForm("type", "other")
	if fileType == "" {
		fileType = "other"
	}

	tag := c.DefaultPostForm("tag", "other")
	if tag == "" {
		tag = "other"
	}

	describe := c.DefaultPostForm("describe", "")

	// Read binary file
	h, err := c.FormFile("file")
	if err != nil {
		log.Errorf("upload err:%s", err.Error())
		resp.Error = Error(UploadFileError, err.Error())
		return
	}

	fileAllowArray, ok := FileAllow[fileType]
	if !ok {
		log.Errorf("upload err: type not permit")
		resp.Error = Error(UploadFileTypeNotPermit, "")
		return
	}

	fileSuffix := myutil.GetFileSuffix(h.Filename)
	if !myutil.InArray(fileAllowArray, fileSuffix) {
		log.Errorf("upload err: file suffix: %s not permit", fileSuffix)
		resp.Error = Error(UploadFileTypeNotPermit, fmt.Sprintf("file suffix: %s not permit", fileSuffix))
		return
	}

	if h.Size > int64(FileBytes) {
		log.Errorf("upload err: file size too big: %d", h.Size)
		resp.Error = Error(UploadFileTooMaxLimit, fmt.Sprintf(" file size too big: %d", h.Size))
		return
	}

	// Open file
	f, err := h.Open()
	if err != nil {
		log.Errorf("upload err:%s", err.Error())
		resp.Error = Error(UploadFileError, err.Error())
		return
	}

	defer f.Close()

	// Read binary
	raw, err := ioutil.ReadAll(f)
	if err != nil {
		log.Errorf("upload err:%s", err.Error())
		resp.Error = Error(UploadFileError, err.Error())
		return
	}

	// When raw bytes empty will occur err
	fileSize := len(raw)
	if fileSize == 0 {
		log.Errorf("upload err:%s", "file empty")
		resp.Error = Error(UploadFileError, "file empty")
		return
	}

	// 去重键：盲索引（同用户 + 同内容 → 同 bidx）。
	// 语义与旧的 hash_code 完全一致（同一用户重复上传同一文件只保留一份），
	// 但不再把明文内容的 sha256 直接写进库与 URL，避免"内容指纹"泄漏。
	rawHash, err := myutil.Sha256(raw)
	if err != nil {
		log.Errorf("upload err:%s", err.Error())
		resp.Error = Error(UploadFileError, err.Error())
		return
	}
	fileBidx, err := model.BlindIndexOf(uName + ":" + rawHash)
	if err != nil {
		log.Errorf("upload err:%s", err.Error())
		resp.Error = Error(UploadFileError, err.Error())
		return
	}

	// 随机存储名：不再用内容哈希命名（旧命名会泄漏明文指纹，路径也更易被猜）
	fileName, err := randomStorageName(fileSuffix)
	if err != nil {
		log.Errorf("upload err:%s", err.Error())
		resp.Error = Error(UploadFileError, err.Error())
		return
	}

	// 查重：命中的复用已有记录与文件，跳过存储
	var responseFile *model.File
	lookup := new(model.File)
	lookup.Bidx = fileBidx
	exist, err := lookup.Get()
	if err != nil {
		resp.Error = Error(DBError, err.Error())
		return
	}

	helpPath := fmt.Sprintf("storage/%s/%s", uName, fileType)
	if !exist {
		p := new(model.File)
		p.Bidx = fileBidx

		// 是否加密文件内容：OSS 模式强制不加密；本地模式默认加密（可配置关闭）
		encryptContent := config.FileEncryptEnabled()
		var (
			fek       []byte
			baseNonce []byte
		)
		storeBytes := raw
		if encryptContent {
			if fek, err = secure.NewKey(); err != nil {
				log.Errorf("upload err:%s", err.Error())
				resp.Error = Error(UploadFileError, err.Error())
				return
			}
			if baseNonce, err = secure.NewFileNonce(); err != nil {
				log.Errorf("upload err:%s", err.Error())
				resp.Error = Error(UploadFileError, err.Error())
				return
			}
			keyring, kerr := secure.Get()
			if kerr != nil {
				log.Errorf("upload err:%s", kerr.Error())
				resp.Error = Error(UploadFileError, kerr.Error())
				return
			}
			wrapped, werr := keyring.WrapKey(fek)
			if werr != nil {
				log.Errorf("upload err:%s", werr.Error())
				resp.Error = Error(UploadFileError, werr.Error())
				return
			}
			if storeBytes, err = secure.EncryptFileBytes(fek, baseNonce, helpPath+"/"+fileName, raw); err != nil {
				log.Errorf("upload err:%s", err.Error())
				resp.Error = Error(UploadFileError, err.Error())
				return
			}
			p.EncVersion = 1
			p.FekWrapped = wrapped
			p.EncNonce = secure.FileNonceHex(baseNonce)
		}

		// File not exist, save it
		fileDir := filepath.Join(config.FaFaConfig.DefaultConfig.StoragePath, uName, fileType)
		fileAbName := filepath.Join(fileDir, fileName)

		// Local mode will save in disk
		if !config.FaFaConfig.DefaultConfig.StorageOss {
			// disk mode first make dir
			err := myutil.MakeDir(fileDir)
			if err != nil {
				log.Errorf("upload err:%s", err.Error())
				resp.Error = Error(UploadFileError, err.Error())
				return
			}

			err = myutil.SaveToFile(fileAbName, storeBytes)
			if err != nil {
				log.Errorf("upload err:%s", err.Error())
				resp.Error = Error(UploadFileError, err.Error())
				return
			}

			p.Url = fmt.Sprintf("/%s/%s", helpPath, fileName)
		} else {
			// Oss mode
			p.StoreType = 1
			p.Url = fmt.Sprintf("%s.%s/%s/%s", config.FaFaConfig.OssConfig.BucketName, config.FaFaConfig.OssConfig.Endpoint, helpPath, fileName)
			err = oss.SaveFile(config.FaFaConfig.OssConfig, helpPath+"/"+fileName, storeBytes)
			if err != nil {
				log.Errorf("upload err:%s", err.Error())
				resp.Error = Error(UploadFileError, err.Error())
				return
			}
		}

		p.UrlHashCode, _ = myutil.Sha256([]byte(p.Url))

		// If is picture, cut the size
		if myutil.InArray(scaleType, fileSuffix) {
			p.IsPicture = 1

			if CanScale {
				// 统一从内存缩放：原图可能已是密文，不能再从磁盘读明文
				outRaw, serr := go_image.ScaleB2B(raw, ScaleWidth)
				if serr != nil {
					log.Errorf("upload err:%s", serr.Error())
					resp.Error = Error(UploadFileError, serr.Error())
					return
				}

				xPath := strings.Replace(helpPath, "storage/", "storage_x/", 1) + "/" + fileName
				p.SizeX = int64(len(outRaw))
				thumbBytes := outRaw
				if encryptContent {
					// 缩略图必须用"派生出的另一条 nonce"：复用 (FEK, nonce) 加密不同内容会泄漏明文
					thumbNonce, nerr := secure.ThumbNonce(baseNonce)
					if nerr != nil {
						log.Errorf("upload err:%s", nerr.Error())
						resp.Error = Error(UploadFileError, nerr.Error())
						return
					}
					if thumbBytes, serr = secure.EncryptFileBytes(fek, thumbNonce, xPath, outRaw); serr != nil {
						log.Errorf("upload err:%s", serr.Error())
						resp.Error = Error(UploadFileError, serr.Error())
						return
					}
				}

				if !config.FaFaConfig.DefaultConfig.StorageOss {
					// Local disk mode，cut the picture and save in  /storage_x
					fileScaleDir := filepath.Join(config.FaFaConfig.DefaultConfig.StoragePath+"_x", uName, fileType)
					fileScaleAbName := filepath.Join(fileScaleDir, fileName)

					if err = myutil.MakeDir(fileScaleDir); err != nil {
						log.Errorf("upload err:%s", err.Error())
						resp.Error = Error(UploadFileError, err.Error())
						return
					}
					if err = myutil.SaveToFile(fileScaleAbName, thumbBytes); err != nil {
						log.Errorf("upload err:%s", err.Error())
						resp.Error = Error(UploadFileError, err.Error())
						return
					}
				} else {
					// OSS again
					if err = oss.SaveFile(config.FaFaConfig.OssConfig, xPath, thumbBytes); err != nil {
						log.Errorf("upload err:%s", err.Error())
						resp.Error = Error(UploadFileError, err.Error())
						return
					}
				}
			}
		}

		p.Type = fileType
		p.FileName = fileName
		p.ReallyFileName = h.Filename
		p.CreateTime = time.Now().Unix()
		p.Describe = describe
		p.UserId = uu.Id
		p.UserName = uName
		p.Tag = tag
		p.Size = int64(fileSize)
		_, err = model.FaFaRdb.InsertOne(p)
		if err != nil {
			log.Errorf("upload err:%s", err.Error())
			resp.Error = Error(DBError, err.Error())
			return
		}
		responseFile = p
	} else {
		// File exist（去重命中，复用已有记录）
		data.Addon = "file the same in server"
		if lookup.Status != 0 {
			// If file is hide must change back
			lookup.Status = 0
			_, _ = lookup.UpdateStatus()
		}
		responseFile = lookup
	}

	// Return all basic info
	data.FileName = responseFile.FileName
	data.ReallyFileName = responseFile.ReallyFileName
	data.IsPicture = responseFile.IsPicture == 1
	data.Size = responseFile.Size
	data.Url = responseFile.Url
	data.Oss = responseFile.StoreType == 1
	if data.IsPicture && CanScale {
		data.UrlX = strings.Replace(responseFile.Url, "/storage", "/storage_x", -1)
	}

	resp.Data = data
	resp.Flag = true
	return
}

type ListFileAdminRequest struct {
	CreateTimeBegin int64    `json:"create_time_begin"`
	CreateTimeEnd   int64    `json:"create_time_end"`
	UpdateTimeBegin int64    `json:"update_time_begin"`
	UpdateTimeEnd   int64    `json:"update_time_end"`
	SizeBegin       int64    `json:"size_begin"`
	SizeEnd         int64    `json:"size_end"`
	Sort            []string `json:"sort"`
	HashCode        string   `json:"hash_code"`
	Url             string   `json:"url"`
	StoreType       int      `json:"store_type" validate:"oneof=-1 0 1"`
	Status          int      `json:"status" validate:"oneof=-1 0 1"`
	Type            string   `json:"type"`
	Tag             string   `json:"tag"`
	UserId          int64    `json:"user_id"`
	Id              int64    `json:"id"`
	IsPicture       int      `json:"is_picture" validate:"oneof=-1 0 1"`
	PageHelp
}

type ListFileAdminResponse struct {
	Files []model.File `json:"files"`
	PageHelp
}

func ListFileAdminHelper(c *gin.Context, userId int64) {
	resp := new(Resp)

	respResult := new(ListFileAdminResponse)
	req := new(ListFileAdminRequest)
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
		log.Errorf("ListFileAdmin err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	// new query list session
	session := model.FaFaRdb.Client.NewSession()
	defer session.Close()

	// group list where prepare
	session.Table(new(model.File)).Where("1=1")

	// query prepare
	if req.Id != 0 {
		session.And("id=?", req.Id)
	}

	if req.Status != -1 {
		// this not expose out, all people well see the file in show, those hide will emm, hide.
		session.And("status=?", req.Status)
	}

	if req.Url != "" {
		urlHashCode, _ := myutil.Sha256([]byte(req.Url))
		session.And("url_hash_code=?", urlHashCode)
	}

	if req.IsPicture != -1 {
		session.And("is_picture=?", req.IsPicture)
	}

	if req.Type != "" {
		session.And("type=?", req.Type)
	}

	if req.StoreType != -1 {
		session.And("store_type=?", req.StoreType)
	}

	if req.Tag != "" {
		session.And("tag=?", req.Tag)
	}

	if userId != 0 {
		session.And("user_id=?", userId)
	} else {
		if req.UserId != 0 {
			session.And("user_id=?", req.UserId)
		}
	}

	if req.CreateTimeBegin > 0 {
		session.And("create_time>=?", req.CreateTimeBegin)
	}

	if req.CreateTimeEnd > 0 {
		session.And("create_time<?", req.CreateTimeEnd)
	}

	if req.UpdateTimeBegin > 0 {
		session.And("update_time>=?", req.UpdateTimeBegin)
	}

	if req.UpdateTimeEnd > 0 {
		session.And("update_time<?", req.UpdateTimeEnd)
	}

	if req.SizeBegin > 0 {
		session.And("size>=?", req.SizeBegin)
	}

	if req.SizeEnd > 0 {
		session.And("size<?", req.SizeEnd)
	}

	files := make([]model.File, 0)
	p := &req.PageHelp

	// sql build
	p.build(session, req.Sort, model.FileSortName)

	// do query
	total, err := session.FindAndCount(&files)
	if err != nil {
		log.Errorf("ListFileAdmin err:%s", err.Error())
		resp.Error = Error(DBError, err.Error())
		return
	}

	// result
	respResult.Files = files
	p.Pages = int(math.Ceil(float64(total) / float64(p.Limit)))
	p.Total = int(total)
	respResult.PageHelp = *p
	resp.Data = respResult
	resp.Flag = true
}

// List all file info of all user, admin url
func ListFileAdmin(c *gin.Context) {
	ListFileAdminHelper(c, 0)
}

// list file of oneself
func ListFile(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("ListFile err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	ListFileAdminHelper(c, uid)
}

type UpdateFileRequest struct {
	Id       int64  `json:"id" validate:"required"`
	Tag      string `json:"tag"`
	Hide     *bool  `json:"hide"` // nil=不改隐藏状态；true=隐藏；false=取消隐藏
	Describe string `json:"describe"`
}

func UpdateFileAdminHelper(c *gin.Context, userId int64) {
	resp := new(Resp)
	req := new(UpdateFileRequest)
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
		log.Errorf("UpdateFileAdmin err: %s", err.Error())
		resp.Error = Error(ParasError, err.Error())
		return
	}

	f := new(model.File)
	f.Id = req.Id

	// can change file tag so can group out
	f.Tag = req.Tag
	f.Describe = req.Describe
	f.UserId = userId

	var ok bool
	var uerr error
	if req.Hide != nil {
		// 显式切换隐藏状态：只动 status，不影响 tag/describe 之外的字段
		ok, uerr = f.UpdateHide(*req.Hide)
	} else {
		// 仅更新描述/标签：不得触碰隐藏状态（避免"改描述"把隐藏文件悄悄恢复）
		ok, uerr = f.UpdateInfo()
	}
	if uerr != nil {
		log.Errorf("UpdateFileAdmin err:%s", uerr.Error())
		resp.Error = Error(DBError, uerr.Error())
		return
	}

	resp.Data = ok
	resp.Flag = true
}

// update file info or every user, admin url
func UpdateFileAdmin(c *gin.Context) {
	UpdateFileAdminHelper(c, 0)
}

// update file info of oneself
func UpdateFile(c *gin.Context) {
	resp := new(Resp)
	uu, err := GetUserSession(c)
	if err != nil {
		log.Errorf("UpdateFile err: %s", err.Error())
		resp.Error = Error(GetUserSessionError, err.Error())
		JSONL(c, 200, nil, resp)
		return
	}

	uid := uu.Id
	UpdateFileAdminHelper(c, uid)
}
