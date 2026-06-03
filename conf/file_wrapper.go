package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/xslasd/goxf/hooks"
	"github.com/xslasd/goxf/utils/gmsm/sm4"
	"github.com/xslasd/goxf/utils/xfmt"
	"github.com/xslasd/goxf/utils/xmap"
	"gopkg.in/yaml.v3"
)

type fileSourceWrapper struct {
	mainDs          ConfigSource
	localDs         ConfigSource
	isEnc           bool
	isLocal         bool
	vPwd            bool
	ext             string
	absPath         string
	localPath       string
	dir             string
	customUnmarshal Unmarshal
	changed         chan struct{}
	closeOnce       sync.Once
}

func (s *fileSourceWrapper) ReadConfig() ([]byte, string, error) {
	var content []byte
	var format string

	if s.isEnc {
		c, _, err := s.mainDs.ReadConfig()
		if err != nil {
			return nil, "", err
		}
		parts := strings.Split(string(c), "|")
		if len(parts) != 2 {
			return nil, "", InvalidConfigData
		}
		hexStr, ext := parts[0], parts[1]
		pwd := configPassword
		if pwd == "" {
			return nil, "", errors.New("password cannot be empty for system.enc")
		}
		decrypted, err := sm4.Sm4DecryptFromHex([]byte(pwd), hexStr)
		if err != nil {
			return nil, "", err
		}
		content = decrypted
		format = ext

		if s.vPwd {
			fName := filepath.Join(s.dir, "config"+ext)
			os.WriteFile(fName, content, 0666)
			hooks.Register(hooks.Stage_AfterRun, func() {
				fmt.Println()
				relName := formatRelPath(fName)
				xfmt.Printf(color.YellowString("Generated plaintext to: %s"), relName)
				xfmt.Printf(color.RedString("Attention: After editing the configuration file, please delete the plaintext configuration file!"))
				xfmt.Printf(color.YellowString("Start command: %s --config=%s"), filepath.Base(os.Args[0]), relName)
			})
			s.vPwd = false // 避免重复生成
		}
	} else if s.isLocal {
		c, fm, err := s.mainDs.ReadConfig()
		if err != nil {
			return nil, "", err
		}
		content = c
		format = fm
		if format == "" {
			format = s.ext
		}

		if s.vPwd {
			pwd := configPassword
			enc, err := sm4.Sm4EncryptToHex([]byte(pwd), content)
			if err == nil {
				fName := filepath.Join(s.dir, "system.enc")
				finalData := fmt.Sprintf("%s|%s", enc, s.ext)
				os.WriteFile(fName, []byte(finalData), 0666)
				hooks.Register(hooks.Stage_AfterRun, func() {
					fmt.Println()
					relName := formatRelPath(fName)
					xfmt.Printf(color.YellowString("Generated ciphertext to: %s"), fName)
					xfmt.Printf(color.RedString("Attention: Please delete the plaintext configuration files to prevent security leaks!"))
					xfmt.Printf(color.YellowString("Start command: %s --config=%s"), filepath.Base(os.Args[0]), relName)
				})
			}
			s.vPwd = false
		}
	} else {
		// 主配置文件
		c1, fm1, err1 := s.mainDs.ReadConfig()
		if err1 != nil && !os.IsNotExist(err1) {
			return nil, "", err1
		}
		format = fm1
		if format == "" {
			format = s.ext
		}

		unmarshal := s.customUnmarshal
		if unmarshal == nil {
			unmarshal = ExtToUnmarshal(format)
		}

		var mainMap = make(map[string]any)
		if len(c1) > 0 && unmarshal != nil {
			unmarshal(c1, &mainMap)
		}

		if s.localDs != nil {
			c2, fm2, err2 := s.localDs.ReadConfig()
			if err2 == nil && len(c2) > 0 {
				umLocal := s.customUnmarshal
				if umLocal == nil && fm2 != "" {
					umLocal = ExtToUnmarshal(fm2)
				}
				if umLocal == nil {
					umLocal = ExtToUnmarshal(filepath.Ext(s.localPath))
				}
				if umLocal != nil {
					var localMap = make(map[string]any)
					umLocal(c2, &localMap)
					var changes = make(map[string]any)
					xmap.MergeStringMapWithChanged(mainMap, localMap, changes, "")
				}
			}
		}

		mergedBytes, err := yaml.Marshal(mainMap)
		if err != nil {
			return nil, "", err
		}
		content = mergedBytes
		format = "yaml"

		if s.vPwd {
			pwd := configPassword
			enc, err := sm4.Sm4EncryptToHex([]byte(pwd), content)
			if err == nil {
				fName := filepath.Join(s.dir, "system.enc")
				finalData := fmt.Sprintf("%s|%s", enc, ".yaml")
				os.WriteFile(fName, []byte(finalData), 0666)
				hooks.Register(hooks.Stage_AfterRun, func() {
					fmt.Println()
					relName := formatRelPath(fName)
					xfmt.Printf(color.YellowString("Generated ciphertext to: %s"), fName)
					xfmt.Printf(color.RedString("Attention: Please delete the plaintext configuration files to prevent security leaks!"))
					xfmt.Printf(color.YellowString("Start command: %s --config=%s"), filepath.Base(os.Args[0]), relName)
				})
			}
			s.vPwd = false
		}
	}
	return content, format, nil
}

func (s *fileSourceWrapper) startWatch() {
	if s.mainDs == nil {
		return
	}
	// 如果都没有被初始化监听，就不启动 goroutine
	if s.mainDs.Changed() == nil && (s.localDs == nil || s.localDs.Changed() == nil) {
		return
	}

	go func() {
		for {
			if s.localDs != nil && s.localDs.Changed() != nil {
				select {
				case _, ok := <-s.mainDs.Changed():
					if !ok {
						return
					}
					select {
					case s.changed <- struct{}{}:
					default:
					}
				case _, ok := <-s.localDs.Changed():
					if !ok {
						return
					}
					select {
					case s.changed <- struct{}{}:
					default:
					}
				}
			} else {
				if s.mainDs.Changed() == nil {
					return
				}
				select {
				case _, ok := <-s.mainDs.Changed():
					if !ok {
						return
					}
					select {
					case s.changed <- struct{}{}:
					default:
					}
				}
			}
		}
	}()
}

func (s *fileSourceWrapper) Changed() <-chan struct{} {
	return s.changed
}

func (s *fileSourceWrapper) Close() error {
	s.closeOnce.Do(func() {
		if s.mainDs != nil {
			s.mainDs.Close()
		}
		if s.localDs != nil {
			s.localDs.Close()
		}
		close(s.changed)
	})
	return nil
}

func formatRelPath(p string) string {
	cwd, err := os.Getwd()
	if err != nil {
		return p
	}
	rel, err := filepath.Rel(cwd, p)
	if err != nil {
		return p
	}
	if !strings.HasPrefix(rel, ".") && !strings.HasPrefix(rel, "/") {
		return "./" + filepath.ToSlash(rel)
	}
	return filepath.ToSlash(rel)
}
