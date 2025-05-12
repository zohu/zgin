package zgin

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/zohu/zgin/zlog"
	"github.com/zohu/zgin/zutil"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func init() {
	_ = godotenv.Load()
}

type Options struct {
	Addr   string
	Domain string
}

func (o *Options) Validate() {
	o.Addr = zutil.FirstTruth(o.Addr, ":8080")
}

type Option func(*Options)

func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}
func WithDomain(domain string) Option {
	return func(o *Options) {
		o.Domain = domain
	}
}

type App struct {
	options  *Options
	server   *http.Server
	tcp      http.Handler
	grpc     http.Handler
	preload  []func(*Options) error
	shutdown []func()
}

func NewApp(opts ...Option) *App {
	options := new(Options)
	for _, opt := range opts {
		opt(options)
	}
	options.Validate()

	return &App{
		options: options,
		server: &http.Server{
			Addr: options.Addr,
		},
	}
}

func (app *App) WithPreload(preload ...func(*Options) error) *App {
	app.preload = append(app.preload, preload...)
	return app
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
	app.server.Handler = h2c.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && app.grpc != nil && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
				app.grpc.ServeHTTP(w, r)
			} else if app.tcp != nil {
				app.tcp.ServeHTTP(w, r)
			} else {
				zlog.Fatalf("tcp and grpc must set one")
			}
		}),
		&http2.Server{},
	)

	if app.tcp == nil && app.grpc == nil {
		zlog.Fatalf("tcp and grpc must set one")
	}

	// 初始化依赖
	for _, f := range app.preload {
		if err := f(app.options); err != nil {
			zlog.Fatalf("preload %s failed: %v", zutil.GetFunctionName(f), err)
		}
	}

	// 启动服务
	go func() {
		if err := app.server.ListenAndServe(); err != nil {
			zlog.Fatalf("starting serve failed: %v", err)
			return
		}
		zlog.Infof("serve is listening on :%s", app.server.Addr)
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
	if err := app.server.Shutdown(ctx); err != nil {
		zlog.Fatalf("serve closing failed: %v", err)
	}
	zlog.Infof("serve closed")
}
