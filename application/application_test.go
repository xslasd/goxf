package application

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplication_VersionInfo(t *testing.T) {
	// 验证 Goxf 核心版本号
	assert.NotEmpty(t, GetGoxfVersion())
	assert.Equal(t, GoxfVersion, GetGoxfVersion())

	// 设置与获取应用元数据
	InitAppInfo("goxf_demo", "demo description", "v1.0.0", "admin", "2026-08-27")
	assert.Equal(t, "goxf_demo", GetTitle())
	assert.Equal(t, "demo description", GetDescription())
	assert.Equal(t, "v1.0.0", GetVersion())
	assert.Equal(t, "admin", GetBuildUser())
	assert.Equal(t, "2026-08-27", GetBuildTime())

	// 单独 Setter 测试
	SetTitle("custom_title")
	assert.Equal(t, "custom_title", GetTitle())

	SetDescription("custom_desc")
	assert.Equal(t, "custom_desc", GetDescription())

	SetVersion("v2.0.0")
	assert.Equal(t, "v2.0.0", GetVersion())

	SetBuildUser("deploy_bot")
	assert.Equal(t, "deploy_bot", GetBuildUser())

	SetBuildTime("2026-08-28")
	assert.Equal(t, "2026-08-28", GetBuildTime())

	// 验证 PrintVersion 正常运行无 panic
	assert.NotPanics(t, func() {
		PrintVersion()
	})
}
