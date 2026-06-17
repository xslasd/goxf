package conf

import (
	"crypto/md5"
	"encoding/hex"
	"os"
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
	res := verifyPassword(true)
	if res {
		t.Error("verifyPassword should return false when --crypt-conf is not provided and isEnc is true")
	}
}
