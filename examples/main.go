package main

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cobra"
	"github.com/xslasd/goxf"
	"github.com/xslasd/goxf/auth"
	"github.com/xslasd/goxf/conf"
	"github.com/xslasd/goxf/ecode"
	"github.com/xslasd/goxf/flag"
	"github.com/xslasd/goxf/i18n"
	"github.com/xslasd/goxf/log"
	"github.com/xslasd/goxf/server/sgin"
	"github.com/xslasd/goxf/utils/xfmt"
)

type MyClaims struct {
	jwt.RegisteredClaims
}

type CustomConfig struct {
	AppName  string `mapstructure:"appName"`
	Version  string `mapstructure:"version"`
	Debug    bool   `mapstructure:"debug"`
	Database struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		Username string `mapstructure:"username"`
	} `mapstructure:"database"`
}

func main() {
	// 注册一个不需要 InitBase 的离线子命令 (如生成代码、纯离线工具)
	flag.AddCommand(&cobra.Command{
		Use:   "offline",
		Short: "纯离线子命令示例，不加载配置文件和运行时",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🚀 Offline command executed! No config loaded.")
		},
	})

	// 注册一个带有参数排除机制的子命令 (默认自动 InitBase，但带 -g 时跳过 InitBase)
	var genName string
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "数据库迁移示例",
		Run: func(cmd *cobra.Command, args []string) {
			if genName != "" {
				fmt.Printf("✨ Generate migration file: %s (No config loaded!)\n", genName)
				return
			}
			log.Infof("🚀 Executing DB migration... App serviceName: %s", conf.GetString("goxf.serviceName"))
		},
	}
	migrateCmd.Flags().StringVarP(&genName, "gen", "g", "", "生成迁移文件")
	flag.AddCommandWithBaseExcept(migrateCmd, "gen", "g")

	// 注册一个需要 InitBase 的子命令 (如 DB迁移、定时任务、数据导入)
	flag.AddCommandWithBase(&cobra.Command{
		Use:   "task",
		Short: "数据任务子命令示例，自动加载配置并初始化运行时",
		Run: func(cmd *cobra.Command, args []string) {
			log.Infof("🚀 Task command executed! App serviceName from config: %s", conf.GetString("goxf.serviceName"))
		},
	})

	srv := goxf.NewService(
		goxf.WithDefaultConfAddr("examples/config.yaml"),
		// 若要体验配置文件加密，取消注释下行并在启动时携带 --crypt-conf 生成 system.enc
		// goxf.WithConfigPassword("123456"),
	)

	// 方式一：加载新配置文件并合并到全局 defaultConfiguration
	if err := conf.LoadFromSource("examples/custom.yaml", conf.WithIgnoreGlobalPassword()); err != nil {
		log.Errorf("failed to load custom.yaml: %v", err)
	} else {
		// 1. 通过 key 读取单个值
		appName := conf.GetString("custom.appName")
		version := conf.GetString("custom.version")
		xfmt.Printf("Loaded from custom.yaml (Global) -> appName: %s, version: %s", appName, version)

		// 2. 将配置映射/反序列化到结构体
		var customCfg CustomConfig
		if err := conf.UnmarshalKey("custom", &customCfg); err != nil {
			log.Errorf("unmarshal custom config error: %v", err)
		} else {
			xfmt.Printf("CustomConfig struct (Global): %+v", customCfg)
		}

		// 3. 注册 Watcher 监听配置动态更新
		conf.SetWatcher("custom", func(c *conf.Conf) {
			xfmt.Printf("[Watcher] global custom config changed! New appName: %s", c.GetString("custom.appName"))
		})
	}

	// 方式二：创建一个全新的独立 Conf 实例（支持 WithKeyDelimiter, WithPassword, WithWatch 等）
	isolatedConf, err := conf.NewConfFromSource("examples/custom.yaml", conf.WithIgnoreGlobalPassword(), conf.WithWatch(true), conf.WithKeyDelimiter("."))
	if err != nil {
		log.Errorf("failed to create isolated conf: %v", err)
	} else {
		xfmt.Printf("Isolated Conf -> appName: %s", isolatedConf.GetString("custom.appName"))
	}

	lang, err := i18n.NewI18N()
	if err != nil {
		panic(err)
	}
	lang_en := lang.Language("en")
	xfmt.Printf("demo.test:%s", lang_en.T("demo.test"))
	xfmt.Printf("user.name:%s", lang.T("user.name"))

	certifier, err := auth.NewJWTCertifier()
	if err != nil {
		fmt.Println(ecode.As(err).Message(), "--")
		panic(err)
	}
	claims := &MyClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "dddd",
			ExpiresAt: certifier.NextExpiresAt(),
		},
	}

	token, err := certifier.GenerateToken(claims)

	if err != nil {
		panic(err.Error())
	}

	claimsData, err := certifier.ParseToken(token)
	if err != nil {
		fmt.Println(ecode.As(err).Message())
		panic(err)
	}
	xfmt.Printf("data:%v", claimsData)

	ginSrv, err := sgin.NewGinServer()
	if err != nil {
		panic(err)
	}

	if err := srv.Run(ginSrv); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
