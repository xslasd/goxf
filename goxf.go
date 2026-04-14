package goxf

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
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
	"github.com/google/uuid"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/flag"
	"github.com/xslasd/goxf/hooks"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/server"
	"github.com/xslasd/goxf/utils/xfmt"
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
}

func NewService(opts ...Option) *Service {
	s := new(Service)
	s.appID = strings.ToLower(strings.ReplaceAll(uuid.New().String(), "-", ""))
	for _, o := range opts {
		o(s)
	}
	if s.confAddr == "" {
		s.confAddr = "config.yaml"
	}
	err := parseFlags(s.appID, s.confAddr)
	if err != nil {
		panic(fmt.Errorf("parse flags error:%w", err))
	}
	s.printBanner()
	hooks.Do(hooks.Stage_BeforeLoadConfig)
	s.initConfig()
	s.loadBaseConfig()
	s.initLogger()
	s.initTracer()
	s.initRegistry()
	return s
}

// initConfig init
func (s *Service) initConfig() {
	configAddr := flag.String("config")
	if !s.isWatchConf {
		s.isWatchConf = flag.Bool("watch")
	}
	absPath, err := filepath.Abs(configAddr)
	if err != nil {
		panic(fmt.Errorf("get absolute path error:%w", err))
	}

	xfmt.Printf("goxf intends to read config from: %s", absPath)
	conf.SetCryptFilePath(absPath)
	provider, err := conf.NewDataSource(absPath, s.isWatchConf)
	if err != nil {
		panic(fmt.Errorf("create config data source error:%w", err))
	}
	if s.confUnmarshal == nil {
		s.confUnmarshal, err = conf.ExtToUnmarshal(configAddr)
		if err != nil {
			xfmt.Printf("unsupported config unmarshal type: %s", filepath.Ext(configAddr))
			os.Exit(1)
		}
	}
	err = conf.LoadFromDataSource(provider, s.confUnmarshal)
	if err != nil {
		panic(err)
	}
	if s.isWatchConf {
		hooks.Register(hooks.Stage_AfterStop, func() {
			err = provider.Close()
			if err != nil {
				panic(err)
			}
		})
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

// Run 开启所有的server，并且监听退出信号
func (s *Service) Run(servers ...server.Server) error {
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
