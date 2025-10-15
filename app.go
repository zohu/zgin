package zgin

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func init() {
	_ = godotenv.Load()
}

type Options struct {
	Addr   string `yaml:"addr" binding:"required" note:"监听地址"`
	Domain string `yaml:"domain" binding:"required" note:"域名"`
}

func (o *Options) Validate() {
	o.Addr = zutil.FirstTruth(o.Addr, ":8080")
	if err := validator.New().Struct(o); err != nil {
		zlog.Fatalf("validate options failed: %v", err)
	}
}

type App struct {
	options  *Options
	server   *http.Server
	tcp      http.Handler
	grpc     http.Handler
	shutdown []func()
}

func NewApp(options *Options) *App {
	options = zutil.FirstTruth(options, &Options{})
	options.Validate()

	return &App{
		options: options,
		server: &http.Server{
			Addr: options.Addr,
		},
	}
}

func (app *App) WithShutdown(shutdown ...func()) *App {
	app.shutdown = append(app.shutdown, shutdown...)
	return app
}
func (app *App) WithGin(e *gin.Engine) *App {
	app.tcp = e
	return app
}
func (app *App) WithGrpc(h http.Handler) *App {
	app.grpc = h
	return app
}

func (app *App) Listen() {
	if app.tcp == nil && app.grpc == nil {
		zlog.Fatalf("tcp and grpc must set one")
	}

	app.server.Handler = h2c.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if app.grpc != nil && r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
				app.grpc.ServeHTTP(w, r)
			} else {
				app.tcp.ServeHTTP(w, r)
			}
		}),
		&http2.Server{},
	)

	// 启动服务
	go func() {
		zlog.Infof("serve is listening on %s", app.server.Addr)
		_ = app.server.ListenAndServe()
	}()

	// 优雅关闭服务
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-quit
	zlog.Infof("serve closing...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for _, f := range app.shutdown {
		f()
	}
	_ = app.server.Shutdown(ctx)
	zlog.Infof("serve closed")
}

func Bind[T any](fn func(*gin.Context, *T) *RespBean) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.ContentType(), gin.MIMEJSON) {
			var params T
			if err := c.ShouldBindJSON(&params); err != nil && !errors.Is(err, io.EOF) {
				AbortHttpCode(c, http.StatusBadRequest, MessageParamInvalid.Resp(c).WithValidateErrs(c, params, err))
			} else {
				Abort(c, fn(c, &params))
			}
			return
		}
		Abort(c, fn(c, new(T)))
	}
}
