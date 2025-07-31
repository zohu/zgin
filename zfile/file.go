package zfile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/zohu/zgin/zdb"
	"github.com/zohu/zgin/zid"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"gorm.io/gorm"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
)

var opts *Options
var svr iService
var useDatabase bool

func New(options *Options) {
	if err := validator.New().Struct(options); err != nil {
		zlog.Fatalf("config error: %s", err)
		return
	}
	opts = options

	// 如果可能，同步数据库表
	if err := zdb.NewDB(context.TODO()).AutoMigrate(&ZfileRecord{}); err == nil {
		useDatabase = true
	}
	switch opts.Provider {
	case ProviderTypeOss:
		svr = newOssService()
	case ProviderTypeS3:
		svr = news3Service()
	default:
		zlog.Fatalf("unknown file provider type: %s", opts.Provider)
	}
	zlog.Infof("init file provider success: %s", opts.Provider)
}

type ReqUpload struct {
	Fid      string `json:"fid" note:"文件ID"`
	Path     string `json:"path" validate:"required" note:"文件存储路径"`
	Name     string `json:"name" validate:"required" note:"文件名"`
	IdleDays int64  `json:"idle_days" note:"最长闲置时间，0则永久, 使用Forward时会更新最后使用时间，否则按上传时间算起"`
	Progress Progress
}

type RespUpload struct {
	Fid  string `json:"fid" note:"文件ID"`
	Name string `json:"name" note:"文件名"`
	Url  string `json:"url" note:"文件地址"`
	Md5  string `json:"md5" note:"文件MD5"`
}

func Upload(ctx context.Context, h *ReqUpload, rs io.ReadSeeker) (*RespUpload, error) {
	if err := validator.New().Struct(h); err != nil {
		return nil, err
	}
	h.Fid = zutil.FirstTruth(h.Fid, zid.NextIdShort())
	// 计算文件哈希
	hash := sha256.New()
	if _, err := io.Copy(hash, rs); err != nil {
		return nil, fmt.Errorf("failed to calculate file MD5: %w", err)
	}
	md5 := hex.EncodeToString(hash.Sum(nil))

	ext := path.Ext(h.Name)
	if ext == "" {
		buf := make([]byte, 512)
		_, err := io.ReadFull(rs, buf)
		if err != nil && err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, err
		}
		_, _ = rs.Seek(0, io.SeekStart)
		extensions, _ := mime.ExtensionsByType(http.DetectContentType(buf))
		if len(extensions) > 0 {
			ext = extensions[0]
		}
	}

	name := opts.FullName(h.Path, h.Fid, ext)
	if useDatabase {
		// 检查文件是否已存在
		var exist ZfileRecord
		zdb.NewDB(ctx).Where("md5=?", md5).First(&exist)
		if exist.Fid != "" {
			return &RespUpload{
				Fid:  exist.Fid,
				Name: opts.FullName(h.Path, exist.Fid, ext),
				Url:  opts.HTTPDomain(exist.Name),
				Md5:  md5,
			}, nil
		}
	}
	// 上传文件
	_, _ = rs.Seek(0, io.SeekStart)
	if err := svr.upload(ctx, rs, name, h.Progress); err != nil {
		return nil, err
	}
	if useDatabase {
		zdb.NewDB(ctx).Create(&ZfileRecord{
			Fid:    h.Fid,
			Md5:    md5,
			Bucket: opts.Bucket,
			Name:   name,
			Expire: zutil.FirstTruth(h.IdleDays, opts.IdleDays),
		})
	}
	return &RespUpload{
		Fid:  h.Fid,
		Name: name,
		Url:  opts.HTTPDomain(name),
		Md5:  md5,
	}, nil
}

func UploadTransfer(ctx context.Context, h *ReqUpload, uri string) (*RespUpload, error) {
	tmpFile, err := os.CreateTemp("", "zfile-url-*.tmp")
	if err != nil {
		return nil, err
	}
	defer tmpFile.Close()
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(tmpPath)
	if err != nil {
		return nil, err
	}
	h.Name = uri
	return Upload(ctx, h, file)
}

func RouteForward(g *gin.RouterGroup) {
	g.GET("static/*fids", func(c *gin.Context) {
		fids := c.Param("fids")
		arr := strings.Split(fids, "/")
		fid := strings.Split(arr[len(arr)-1], ".")[0]
		var ext ZfileRecord
		err := zdb.NewDB(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("fid=?", fid).First(&ext).Error; err != nil {
				return err
			}
			ext.Pv += 1
			zdb.NewDB(c.Request.Context()).Updates(&ext)
			return nil
		})
		if err != nil {
			c.AbortWithStatus(404)
			return
		}
		c.Redirect(http.StatusFound, opts.HTTPDomain(ext.Name))
	})
}
