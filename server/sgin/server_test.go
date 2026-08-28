package sgin

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/conf"
	_ "github.com/xslasd/goxf/conf/filesource"
	"github.com/xslasd/goxf/log"
)

func setupTestEnv(t *testing.T) {
	application.NewRuntime("test_app", "test_service", true, false, false, false, false)
	log.SetLogger(log.NewLogger())

	cfgFile := "test_sgin_config.yaml"
	content := []byte(`
server:
  gin:
    default:
      Addr: "0.0.0.0:8080"
      SlowQueryThresholdInMilli: 500
      MaxMultipartMemory: 67108864
`)
	require.NoError(t, os.WriteFile(cfgFile, content, 0644))
	t.Cleanup(func() { _ = os.Remove(cfgFile) })
	require.NoError(t, conf.LoadFromSource(cfgFile))
}

func TestNewGinServer_MaxMultipartMemory(t *testing.T) {
	setupTestEnv(t)

	// Case 1: 从配置文件中读取 MaxMultipartMemory (64MB = 67108864)
	s1, err := NewGinServer()
	require.NoError(t, err)
	assert.Equal(t, int64(67108864), s1.Engine.MaxMultipartMemory)

	// Case 2: 通过 Option 覆盖 MaxMultipartMemory (128MB)
	customMemory := int64(128 << 20)
	s2, err := NewGinServer(WithMaxMultipartMemory(customMemory))
	require.NoError(t, err)
	assert.Equal(t, customMemory, s2.Engine.MaxMultipartMemory)
}
