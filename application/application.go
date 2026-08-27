package application

import (
	"fmt"
	"time"
)

const (
	// GoxfVersion 当前 goxf 框架核心版本号
	GoxfVersion = "v0.5.0"

	// TraceContextHeaderName trace-id
	TraceContextHeaderName = "x-trace-id"
	// KeyServiceInfo service_info
	KeyServiceInfo = "__service_info_"
)

var (
	title       string
	description string

	buildVersion string
	buildUser    string
	buildTime    string
)

type application struct {
	appId       string
	serviceName string

	enableConsole  bool
	enableTrace    bool
	enableMetric   bool
	enableRegister bool
	enablePprof    bool

	startTime time.Time
}

var runtime = new(application)

// GetGoxfVersion 获取 goxf 框架核心版本号
func GetGoxfVersion() string {
	return GoxfVersion
}

func GetAppId() string {
	return runtime.appId
}
func GetServiceName() string {
	return runtime.serviceName
}
func GetEnableConsole() bool {
	return runtime.enableConsole
}
func GetEnableTrace() bool {
	return runtime.enableTrace
}
func GetEnableMetric() bool {
	return runtime.enableMetric
}
func GetEnableRegister() bool {
	return runtime.enableRegister
}
func GetEnablePprof() bool {
	return runtime.enablePprof
}
func CheckStartupGoxf() {
	if runtime.appId == "" {
		panic("Important: Please execute the `goxf.NewService()` function before starting the client and server.")
	}
}

func GetTitle() string {
	return title
}
func SetTitle(t string) {
	title = t
}

func GetDescription() string {
	return description
}
func SetDescription(d string) {
	description = d
}

func GetVersion() string {
	return buildVersion
}
func SetVersion(v string) {
	buildVersion = v
}

func GetBuildUser() string {
	return buildUser
}
func SetBuildUser(u string) {
	buildUser = u
}

func GetBuildTime() string {
	return buildTime
}
func SetBuildTime(t string) {
	buildTime = t
}

func GetStartTime() time.Time {
	return runtime.startTime
}

// InitAppInfo 初始化业务应用版本与元数据信息
func InitAppInfo(t, d, v, u, bTime string) {
	title = t
	description = d
	buildVersion = v
	buildUser = u
	buildTime = bTime
}

func NewRuntime(appId, serviceName string, enableConsole, enableTrace, enableMetric, enableRegister, enablePprof bool) {
	runtime.appId = appId
	runtime.serviceName = serviceName
	runtime.enableConsole = enableConsole
	runtime.enableTrace = enableTrace
	runtime.enableMetric = enableMetric
	runtime.enableRegister = enableRegister
	runtime.enablePprof = enablePprof
	runtime.startTime = time.Now()
}

func PrintVersion() {
	fmt.Printf("goxfVersion: %s\n", GoxfVersion)
	fmt.Printf("title: %s\n", title)
	fmt.Printf("description: %s\n", description)
	fmt.Printf("buildVersion: %s\n", buildVersion)
	fmt.Printf("buildUser: %s\n", buildUser)
	fmt.Printf("buildTime: %s\n", buildTime)
}
