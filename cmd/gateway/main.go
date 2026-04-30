package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/okuc/llm-bridge-gateway/internal/config"
	"github.com/okuc/llm-bridge-gateway/internal/converter"
	"github.com/okuc/llm-bridge-gateway/internal/middleware"
	"github.com/okuc/llm-bridge-gateway/internal/proxy"
	"github.com/okuc/llm-bridge-gateway/internal/router"
	"github.com/okuc/llm-bridge-gateway/pkg/logger"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "./config/config.yaml", "Path to config file")
	flag.Parse()

	// 加载配置
	loader := config.NewLoader()
	cfg, err := loader.Load(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 验证配置
	validator := config.NewValidator()
	if err := validator.Validate(cfg); err != nil {
		fmt.Printf("Invalid config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger.Init(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.Output, cfg.Logging.FilePath)
	defer logger.Close()
	logger.Info("Starting LLM Bridge Gateway",
		"version", "1.0.0",
		"config", *configPath,
	)

	// 初始化协议转换器
	protocolConverter := converter.New()
	protocolConverter.Init()
	logger.Info("Protocol converter initialized")

	// 初始化HTTP代理
	httpProxy := proxy.New()
	defer httpProxy.Close()

	// 创建Gin引擎
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// 注册中间件
	engine.Use(middleware.Recovery())
	engine.Use(middleware.Logger())
	engine.Use(middleware.CORS())

	// 注册路由
	routerManager := router.New(cfg, protocolConverter, httpProxy)
	routerManager.RegisterRoutes(engine)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 启动服务器
	errCh := make(chan error, 1)
	go func() {
		logger.Info("Server starting", "address", cfg.GetAddress())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
			errCh <- err
		}
		close(errCh)
	}()

	// 等待中断信号或启动错误
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Received shutdown signal")
	case err := <-errCh:
		if err != nil {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}
