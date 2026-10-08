package controllers

import (
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hunterhug/fafacms/core/config"
	"github.com/hunterhug/fafacms/core/model"
	myutil "github.com/hunterhug/fafacms/core/util"
	"github.com/hunterhug/fafacms/core/util/secure"
	log "github.com/hunterhug/golog"
)

// 受控文件下载：
//
// 原先由 engine.Static 直接以静态文件方式提供；由于文件内容可能是密文
// （EncVersion=1），必须由后端解密后再返回，因此改为受控 handler。
//
// 行为对齐原静态服务：
//   - EncVersion=0（未加密，含加密开关关闭时上传的文件）→ 明文直接返回，Range/缓存语义不变；
//   - EncVersion=1 → 解密后返回，并支持 Range（音视频拖动），只解密覆盖区间的块。
func ServeStorage(c *gin.Context)  { serveStorageFile(c, false) }
func ServeStorageX(c *gin.Context) { serveStorageFile(c, true) }

func serveStorageFile(c *gin.Context, thumb bool) {
	// 仅本地模式需要后端提供文件；OSS 模式的文件 URL 直指对象存储，不会走到这里
	if config.FaFaConfig.DefaultConfig.StorageOss {
		c.Status(http.StatusNotFound)
		return
	}

	urlPath := c.Request.URL.Path

	var root, prefix string
	if thumb {
		root, prefix = config.FaFaConfig.DefaultConfig.StoragePath+"_x", "/storage_x/"
	} else {
		root, prefix = config.FaFaConfig.DefaultConfig.StoragePath, "/storage/"
	}

	rel := strings.TrimPrefix(urlPath, prefix)
	// 防御路径穿越
	if rel == urlPath || rel == "" || strings.Contains(rel, "..") {
		c.Status(http.StatusNotFound)
		return
	}
	diskPath := filepath.Join(root, filepath.FromSlash(rel))

	// 按 URL 摘要定位文件记录。
	// 缩略图没有独立记录：它属于原图那条记录（url 存的是 /storage/... 形式），
	// 因此查询时要把 /storage_x/ 还原为 /storage/ 再算摘要。
	lookupPath := urlPath
	if thumb {
		lookupPath = "/storage/" + strings.TrimPrefix(urlPath, prefix)
	}
	urlHash, herr := myutil.Sha256([]byte(lookupPath))
	if herr != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	f := &model.File{UrlHashCode: urlHash}
	exist, err := f.Get()
	if err != nil {
		log.Errorf("ServeFile err: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if !exist {
		c.Status(http.StatusNotFound)
		return
	}

	contentType := contentTypeByPath(rel)

	// 未加密：等价原静态服务（含 Range、Last-Modified）
	if f.EncVersion != 1 {
		servePlainFile(c, diskPath, contentType)
		return
	}

	// 加密：解密后返回
	cipherBytes, err := ioutil.ReadFile(diskPath)
	if err != nil {
		log.Errorf("ServeFile read err: %s", err.Error())
		c.Status(http.StatusNotFound)
		return
	}

	keyring, err := secure.Get()
	if err != nil {
		log.Errorf("ServeFile keyring err: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	fek, err := keyring.UnwrapKey(f.FekWrapped)
	if err != nil {
		log.Errorf("ServeFile unwrap err: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	nonce, err := secure.ParseFileNonce(f.EncNonce)
	if err != nil {
		log.Errorf("ServeFile nonce err: %s", err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	if thumb {
		// 缩略图用由原图 nonce 派生的独立 nonce（避免密钥流复用）
		if nonce, err = secure.ThumbNonce(nonce); err != nil {
			log.Errorf("ServeFile thumb nonce err: %s", err.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
	}

	plainLen := f.Size
	if thumb {
		plainLen = f.SizeX
	}
	if plainLen <= 0 {
		c.Status(http.StatusNotFound)
		return
	}

	// AAD 前缀与上传时一致：去掉前导斜杠的访问路径
	aadPrefix := strings.TrimPrefix(urlPath, "/")

	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")

	rangeHeader := c.GetHeader("Range")
	if rangeHeader == "" {
		plain, derr := secure.DecryptFileBytes(fek, nonce, aadPrefix, cipherBytes, plainLen)
		if derr != nil {
			log.Errorf("ServeFile decrypt err: %s", derr.Error())
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Header("Content-Length", strconv.Itoa(len(plain)))
		c.Status(http.StatusOK)
		_, _ = c.Writer.Write(plain)
		return
	}

	start, end, ok := parseRangeHeader(rangeHeader, plainLen)
	if !ok {
		c.Header("Content-Range", fmt.Sprintf("bytes */%d", plainLen))
		c.Status(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	part, derr := secure.DecryptFileRange(fek, nonce, aadPrefix, cipherBytes, plainLen, start, end)
	if derr != nil {
		log.Errorf("ServeFile decrypt range err: %s", derr.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, plainLen))
	c.Header("Content-Length", strconv.Itoa(len(part)))
	c.Status(http.StatusPartialContent)
	_, _ = c.Writer.Write(part)
}

// servePlainFile 以静态文件方式返回（未加密文件），保留 Range 等原生语义。
func servePlainFile(c *gin.Context, diskPath, contentType string) {
	if _, err := ioutil.ReadFile(diskPath); err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	if contentType != "" {
		c.Header("Content-Type", contentType)
	}
	http.ServeFile(c.Writer, c.Request, diskPath)
}

// contentTypeByPath 由扩展名推断 Content-Type。
func contentTypeByPath(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".mp4":
		return "video/mp4"
	case ".mp3":
		return "audio/mpeg"
	}
	return "application/octet-stream"
}

// parseRangeHeader 解析单区间 Range 头：bytes=start-end / bytes=start- / bytes=-suffix。
// 多区间请求只取第一段（足以满足音视频拖动）；解析失败返回 ok=false。
func parseRangeHeader(header string, size int64) (int64, int64, bool) {
	const prefix = "bytes="
	if size <= 0 || !strings.HasPrefix(header, prefix) {
		return 0, 0, false
	}
	spec := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if idx := strings.Index(spec, ","); idx >= 0 {
		spec = strings.TrimSpace(spec[:idx])
	}
	dash := strings.Index(spec, "-")
	if dash < 0 {
		return 0, 0, false
	}

	startStr := strings.TrimSpace(spec[:dash])
	endStr := strings.TrimSpace(spec[dash+1:])

	var start, end int64
	if startStr == "" {
		// 后缀区间：最后 n 字节
		n, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil || n <= 0 {
			return 0, 0, false
		}
		if n > size {
			n = size
		}
		start, end = size-n, size-1
	} else {
		v, err := strconv.ParseInt(startStr, 10, 64)
		if err != nil || v < 0 {
			return 0, 0, false
		}
		start = v
		if endStr == "" {
			end = size - 1
		} else {
			e, err := strconv.ParseInt(endStr, 10, 64)
			if err != nil {
				return 0, 0, false
			}
			if e > size-1 {
				e = size - 1
			}
			end = e
		}
	}

	if start >= size || start > end {
		return 0, 0, false
	}
	return start, end, true
}
