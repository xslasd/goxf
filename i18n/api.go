package i18n

import (
	"fmt"

	"github.com/xslasd/goxf/application"
	"github.com/xslasd/goxf/conf"
)

type Instance interface {
	Language(lang string) ILanguage
	ILanguage
}

type ILanguage interface {
	LanguageStr() string
	T(key string) string
	Tf(key string, values ...any) string
	Translate(key string) string
	TranslateFormat(key string, values ...any) string
}

func NewI18N(opts ...Option) (Instance, error) {
	application.CheckStartupGoxf()
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	key := opt.confPrefix + "." + opt.confName
	if err := conf.UnmarshalKey(key, opt.config); err != nil {
		return nil, err
	}

	return newManager(opt)
}

func (m *manager) T(key string) string {
	return m.Translate(key)
}
func (m *manager) Tf(key string, values ...any) string {
	return m.TranslateFormat(key, values...)
}

func (m *manager) Translate(key string) string {
	return getTranslateWords(m.data, m.language, key)
}
func (m *manager) TranslateFormat(key string, values ...any) string {
	format := getTranslateWords(m.data, m.language, key)
	return fmt.Sprintf(format, values...)
}

func (m *manager) Language(lang string) ILanguage {
	return &language{
		m:    m,
		lang: lang,
	}
}
func (m *manager) AllLanguageData() map[string]map[string]string {
	return m.data
}

type language struct {
	m    *manager
	lang string
}

func (i *language) LanguageStr() string {
	return i.lang
}

func (i *language) T(key string) string {
	return i.Translate(key)
}
func (i *language) Tf(key string, values ...any) string {
	return i.TranslateFormat(key, values...)
}

func (i *language) Translate(key string) string {
	return getTranslateWords(i.m.data, i.lang, key)
}
func (i *language) TranslateFormat(key string, values ...any) string {
	format := getTranslateWords(i.m.data, i.lang, key)
	return fmt.Sprintf(format, values...)
}
