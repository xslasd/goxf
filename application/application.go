package application

import (
	"fmt"
	"time"
)

const (
	//TraceContextHeaderName trace-id
	TraceContextHeaderName = "x-trace-id"
	//KeyServiceInfo service_info
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

func GetVersion() string {
	return buildVersion
}
func GetBuildUser() string {
	return buildUser
}

func GetBuildTime() string {
	return buildTime
}
func GetStartTime() time.Time {
	return runtime.startTime
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
	fmt.Printf("title: %s\n", title)
	fmt.Printf("description: %s\n", description)
	fmt.Printf("buildVersion: %s\n", buildVersion)
	fmt.Printf("buildUser: %s\n", buildUser)
	fmt.Printf("buildTime: %s\n", buildTime)
}
