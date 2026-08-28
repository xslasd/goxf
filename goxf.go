package goxf

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/client/cetcd"
	"github.com/xslasd/goxf/governor"
	"github.com/xslasd/goxf/registry"
	etcdv3Registry "github.com/xslasd/goxf/registry/etcdsource"
	"github.com/xslasd/goxf/registry/resolver"
	"github.com/xslasd/goxf/tracer/jaeger"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/flag"
	"github.com/xslasd/goxf/hooks"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/server"
	"github.com/xslasd/goxf/utils/xfmt"
	"github.com/xslasd/goxf/utils/xrand"
	"golang.org/x/sync/errgroup"

	_ "github.com/xslasd/goxf/conf/filesource"
)

type Service struct {
	isWatchConf bool   // 是否监听配置文件变化
	loggerKey   string // 日志config key
	appID       string

	confAddr      string
	confUnmarshal conf.Unmarshal
	servers       []server.Server
	registry      registry.Registry

	baseInited   bool // 是否已初始化基础依赖 (config, application runtime, logger)
	bootstrapped bool // 是否已经初始化过全部基础设施
}

var (
	currentServiceMu sync.RWMutex
	currentService   *Service
)

// RequireInitBase 标记一个子命令在执行前需要 goxf 自动完成基础依赖初始化 (配置加载、运行时元数据、日志系统)
func RequireInitBase(cmd *cobra.Command, require ...bool) *cobra.Command {
	return flag.RequireInitBase(cmd, require...)
}

// InitBase 供任何代码或子命令随时按需手动初始化基础依赖 (配置加载、运行时元数据、日志系统)，幂等安全
func InitBase(opts ...Option) error {
	currentServiceMu.RLock()
	s := currentService
	currentServiceMu.RUnlock()

	if s == nil {
		s = &Service{
			appID:    xrand.UUIDv7Simple(),
			confAddr: "config.yaml",
		}
		for _, o := range opts {
			o(s)
		}
		registerDefaultFlags(s.appID, s.confAddr)
	}
	return s.InitBase()
}

func NewService(opts ...Option) *Service {
	s := new(Service)
	s.appID = xrand.UUIDv7Simple()
	for _, o := range opts {
		o(s)
	}
	if s.confAddr == "" {
		s.confAddr = "config.yaml"
	}
	currentServiceMu.Lock()
	currentService = s
	currentServiceMu.Unlock()
	s.printBanner()
	// 注册默认参数标志与子命令
	registerDefaultFlags(s.appID, s.confAddr)

	// 1. 如果命令行命中了注册的子命令（例如 migrate, scaffold, version 等）
	if matchedCmd := flag.MatchedSubCommand(); matchedCmd != nil {
		// 若该子命令声明了需要 InitBase，则在执行前自动初始化配置、运行时和日志
		if flag.ShouldInitBase(matchedCmd) {
			if err := s.InitBase(); err != nil {
				fmt.Fprintf(os.Stderr, "init base error for command [%s]: %v\n", matchedCmd.Name(), err)
				os.Exit(1)
			}
		}

		// 直接分发执行该子命令逻辑并安全退出
		if err := flag.Execute(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 2. 正常微服务启动流程：执行完整 Bootstrap
	if err := s.Bootstrap(); err != nil {
		panic(err)
	}
	return s
}

// InitBase 初始化核心基础依赖（配置加载、运行时元数据、日志系统），不启动网络服务器与注册中心
func (s *Service) InitBase() error {
	if s.baseInited {
		return nil
	}
	s.baseInited = true

	// 解析全局命令行参数 (支持识别 -c, -w 等全局标志)
	if err := flag.Parse(); err != nil {
		return fmt.Errorf("parse flags error: %w", err)
	}

	hooks.Do(hooks.Stage_BeforeLoadConfig)
	s.initConfig()
	s.loadBaseConfig()
	s.initLogger()
	return nil
}

// Bootstrap 初始化全部微服务基础设施（配置、日志、Tracer 和 Registry 等）
func (s *Service) Bootstrap() error {
	if s.bootstrapped {
		return nil
	}
	s.bootstrapped = true

	if err := s.InitBase(); err != nil {
		return err
	}
	s.initTracer()
	s.initRegistry()
	return nil
}

// initConfig init
func (s *Service) initConfig() {
	configAddr := flag.String("config")
	if !s.isWatchConf {
		s.isWatchConf = flag.Bool("watch")
	}
	xfmt.Printf("goxf intends to read config from: %s", configAddr)
	err := conf.LoadFromSource(configAddr, conf.WithUnmarshal(s.confUnmarshal), conf.WithWatch(s.isWatchConf))
	if err != nil {
		panic(fmt.Errorf("create config data source error:%w", err))
	}
}
func (s *Service) loadBaseConfig() {
	c := new(config)
	if err := conf.UnmarshalKey("goxf", c); err != nil {
		panic(fmt.Sprintf("load goxf config error:%s", err))
	}
	application.NewRuntime(s.appID, c.ServiceName, c.EnableConsole, c.EnableTrace, c.EnableMetric, c.EnableRegister, c.EnablePprof)
}
func (s *Service) initLogger() {
	var logger *log.Logger
	configHandle := func(key string, c *log.Config) error {
		return conf.UnmarshalKey(key, c)
	}
	log.SetLoggerConfigHandle(configHandle)
	if s.loggerKey != "" {
		logger = log.NewLogger(log.ConfName(s.loggerKey))
	} else {
		logger = log.NewLogger()
	}
	log.SetLogger(logger)

	if s.isWatchConf {
		// AutoLevel will dynamically set level when config source is changed
		confKey := logger.GetConfLevelKey()
		conf.SetWatcher(confKey, func(config *conf.Conf) {
			lvText := strings.ToLower(config.GetString(confKey))
			if lvText != "" {
				log.Info("update level", log.String("level", lvText), log.String("name", logger.ConfName()))
				logger.Level(lvText)
			}
		})
	}
}

func (s *Service) initTracer() {
	if !application.GetEnableTrace() {
		return
	}
	tracerSource, closer, err := jaeger.NewTrace(jaeger.WithTags(opentracing.Tag{
		Key:   "appid",
		Value: s.appID,
	}))
	if err != nil {
		panic(err)
	}
	hooks.Register(hooks.Stage_AfterStop, func() {
		closer.Close()
	})
	opentracing.InitGlobalTracer(tracerSource)
}
func (s *Service) initRegistry() {
	if !application.GetEnableRegister() {
		return
	}
	if s.registry == nil {
		eCli, err := cetcd.NewClient()
		if err != nil {
			panic(err)
		}
		etcdRegistry := etcdv3Registry.NewRegistry(eCli, &etcdv3Registry.Config{
			ReadTimeout: 5 * time.Second,
			ServiceTTL:  10,
			Prefix:      "goxf",
		})
		s.registry = etcdRegistry
	}
	resolver.RegisterBuilder(s.registry.Kind(), s.registry)
}

// Run 开启所有的server，并且监听退出信号；若命中了子命令则自动路由执行子命令
func (s *Service) Run(servers ...server.Server) error {
	// 1. 如果命令行命中了注册的子命令（例如 migrate, scaffold, version 等）
	// 直接分发执行该子命令逻辑并退出，不触发微服务基础设施初始化与 Server 启动
	if flag.HasSubCommand() {
		return flag.Execute()
	}

	// 2. 正常启动流程：初始化微服务基础设施
	if err := s.Bootstrap(); err != nil {
		return err
	}

	hooks.Do(hooks.Stage_BeforeRun)
	done := make(chan error)
	//governor server
	if conf.GetString("server.gin.governor.addr") != "" {
		s.servers = append(s.servers, governor.CreateGovernor(application.GetEnablePprof(), application.GetEnableMetric()))
	}
	s.runServers(done, servers...)
	select {
	case sig := <-s.waitSignals():
		log.Info("received signal, starting to exit.", log.Any("signal", sig))
		return s.gracefulStop()
	case err := <-done:
		return err
	}
}

func (s *Service) runServers(done chan<- error, servers ...server.Server) {
	var eg errgroup.Group
	s.servers = append(s.servers, servers...)
	for _, srv := range s.servers {
		err := srv.Init()
		if err != nil {
			done <- err
			return
		}
		info := srv.Info()
		if srv.IsRegister() {
			err = s.registry.Register(context.Background(), info)
			if err != nil {
				done <- err
				return
			}
		}
		log.Infof("%s.%s: start server[http://%s],The server process will goroutine.", info.Scheme, info.Name, info.Address)
		// you must do this, or all goroutines will get the same last reference of srv
		srvTmp := srv
		eg.Go(func() error {
			return srvTmp.Serve()
		})
	}
	hooks.Do(hooks.Stage_AfterRun)
	log.Info("★★★ goxf start complete ★★★. block until done or return error.")
	go func() {
		if err := eg.Wait(); err != nil {
			done <- err
		}
	}()
}

func (s *Service) gracefulStop() error {
	hooks.Do(hooks.Stage_BeforeStop)
	var wg sync.WaitGroup
	errChan := make(chan error, len(s.servers)*2)
	for _, srv := range s.servers {
		wg.Add(1)
		go func(childSrv server.Server) {
			defer wg.Done()
			srvInfo := childSrv.Info()
			// 1. 并发注销逻辑
			if childSrv.IsRegister() {
				uCtx, uCancel := context.WithTimeout(context.Background(), 3*time.Second)
				err := s.registry.UnRegister(uCtx, srvInfo)
				uCancel()
				if err != nil {
					log.Error("unregister service error", log.Any("service", srvInfo), log.FieldErr(err))
					errChan <- fmt.Errorf("unregister %s error: %w", srvInfo.Name, err)
				}
			}
			// 2. 优雅停止 Server
			sCtx, sCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer sCancel()
			if err := childSrv.Stop(sCtx); err != nil {
				log.Error("stop server error", log.Any("service", srvInfo), log.FieldErr(err))
				errChan <- fmt.Errorf("stop %s error: %w", srvInfo.Name, err)
			}
		}(srv)
	}

	wg.Wait()
	close(errChan)
	hooks.Do(hooks.Stage_AfterStop)
	if len(errChan) > 0 {
		var errMsgs []string
		for err := range errChan {
			errMsgs = append(errMsgs, err.Error())
		}
		return fmt.Errorf("graceful stop finished with errors: [%s]", strings.Join(errMsgs, "; "))
	}
	return nil
}

// AddServer 提供Server服务加入到servers 执行队列中
func (s *Service) AddServer(srv server.Server) {
	s.servers = append(s.servers, srv)
}

// waitSignals 等待关闭信号
func (s *Service) waitSignals() <-chan os.Signal {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	return signals
}

// printBanner initPrint
func (s *Service) printBanner() {
	const banner = `
                      __ 
                     / _|
   __ _   ___ __  __| |_ 
  / _' | / _ \\ \/ /|  _|
 | (_| || (_) |>  < | |  
  \__, | \___//_/\_\|_|  
   __/ |                 
  |___/

 Welcome to goxf, starting service ...
`
	fmt.Println(color.GreenString(banner))
}
