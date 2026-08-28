package conf

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func md5Crypt(pwd string) string {
	h := md5.New()
	h.Write([]byte(pwd))
	return hex.EncodeToString(h.Sum(nil))
}

func TestSetPasswordWithCrypt(t *testing.T) {
	// 重置全局状态
	configPassword = ""
	passwordCryptFunc = nil

	plain := "my_secure_pwd"
	hashed := md5Crypt(plain)

	// 测试：带上加密方法的 SetPassword
	SetPassword(hashed, md5Crypt)

	if configPassword != hashed {
		t.Errorf("expected configPassword to be %s, got %s", hashed, configPassword)
	}
	if passwordCryptFunc == nil {
		t.Fatal("passwordCryptFunc should not be nil")
	}

	// 验证哈希后的明文输入是否与密文匹配
	if passwordCryptFunc(plain) != configPassword {
		t.Error("hash of plain text did not match the hashed configPassword")
	}
}

func TestSetPasswordWithoutCrypt(t *testing.T) {
	// 重置全局状态
	configPassword = ""
	passwordCryptFunc = nil

	plain := "my_secure_pwd"

	// 测试：不带加密方法的 SetPassword（向后兼容测试）
	SetPassword(plain)

	if configPassword != plain {
		t.Errorf("expected configPassword to be %s, got %s", plain, configPassword)
	}
	if passwordCryptFunc != nil {
		t.Error("passwordCryptFunc should be nil")
	}
}

func TestVerifyPasswordLogic(t *testing.T) {
	// 备份并还原 os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// 1. 设置命令行参数（不带有 --crypt-conf 标志）
	os.Args = []string{"test_app"}

	configPassword = "hashed_or_plain_password"
	passwordCryptFunc = nil

	// 当 configPassword 不为空，且没有 --crypt-conf 参数，并且 isEnc 是 true 时：
	// verifyPassword 应直接返回 false
	res := verifyPassword("test_app.yaml", configPassword, passwordCryptFunc, true)
	if res {
		t.Error("verifyPassword should return false when --crypt-conf is not provided and isEnc is true")
	}
}

type mockFileSource struct {
	path string
}

func (m *mockFileSource) ReadConfig() ([]byte, string, error) {
	content, err := os.ReadFile(m.path)
	if err != nil {
		return nil, "", err
	}
	return content, filepath.Ext(m.path), nil
}

func (m *mockFileSource) Changed() <-chan struct{} {
	return nil
}

func (m *mockFileSource) Close() error {
	return nil
}

func TestNewConfFromSource(t *testing.T) {
	// 重置全局密码状态，避免前面测试的影响
	configPassword = ""
	passwordCryptFunc = nil

	Register(FileScheme, func(path string, watch bool) ConfigSource {
		return &mockFileSource{path: path}
	})

	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "test_config.yaml")
	yamlContent := []byte("app:\n  name: isolated-service\n  port: 8080\n")
	if err := os.WriteFile(cfgPath, yamlContent, 0644); err != nil {
		t.Fatal(err)
	}

	// 1. 创建独立 Conf 实例（使用 Options）
	customConf, err := NewConfFromSource(cfgPath, WithKeyDelimiter("/"), WithWatch(false))
	if err != nil {
		t.Fatalf("failed to create conf from source: %v", err)
	}

	// 2. 验证独立实例中的配置值（使用自定义的分隔符 "/"）
	if customConf.GetString("app/name") != "isolated-service" {
		t.Errorf("expected 'isolated-service', got '%s'", customConf.GetString("app/name"))
	}
	if customConf.GetInt("app/port") != 8080 {
		t.Errorf("expected 8080, got %d", customConf.GetInt("app/port"))
	}

	// 3. 验证全局 defaultConfiguration 未受污染
	if defaultConfiguration.GetString("app.name") == "isolated-service" {
		t.Errorf("defaultConfiguration should not contain isolated-service")
	}

	// 4. 测试 NewFileConfSource
	fileConf, err := NewFileConfSource(cfgPath, false)
	if err != nil {
		t.Fatalf("failed to create file conf: %v", err)
	}
	if fileConf.GetString("app.name") != "isolated-service" {
		t.Errorf("expected 'isolated-service', got '%s'", fileConf.GetString("app.name"))
	}
}

func TestConfOptions_Password(t *testing.T) {
	// 测试 WithPassword 与 WithIgnoreGlobalPassword
	configPassword = "global_password"
	passwordCryptFunc = nil
	defer func() {
		configPassword = ""
	}()

	opts := defaultOptions()
	WithPassword("custom_pwd")(opts)
	pwd, _ := opts.getEffectivePassword()
	if pwd != "custom_pwd" {
		t.Errorf("expected 'custom_pwd', got '%s'", pwd)
	}

	opts2 := defaultOptions()
	WithIgnoreGlobalPassword()(opts2)
	pwd2, _ := opts2.getEffectivePassword()
	if pwd2 != "" {
		t.Errorf("expected empty password when ignoring global password, got '%s'", pwd2)
	}
}

