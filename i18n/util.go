package i18n

import (
	"github.com/fatih/color"
	"github.com/xslasd/goxf/log"
)

func merge(data map[string]map[string]string, lang string, configuration map[string]string) {
	langData, ok := data[lang]
	if !ok {
		data[lang] = make(map[string]string)
		langData = data[lang]
	}
	for k, v := range configuration {
		_, ok := langData[k]
		if ok {
			continue
		}
		langData[k] = v
	}
}

func traverse(data map[string]any, keyDelimiter string, base string) map[string]string {
	res := make(map[string]string)
	for k, v := range data {
		switch v.(type) {
		case string:
			res[base+k] = v.(string)
		case map[string]any:
			m := v.(map[string]any)
			mp := traverse(m, base+k+keyDelimiter, keyDelimiter)
			for k1, v1 := range mp {
				res[base+k+k1] = v1
			}
		case []map[string]any:
			m := v.([]map[string]any)
			for _, item := range m {
				mp := traverse(item, base+k+keyDelimiter, keyDelimiter)
				for k1, v1 := range mp {
					res[base+k+k1] = v1
				}
			}
		}
	}
	return res
}

func getTranslateWords(data map[string]map[string]string, language, key string) string {
	tData, ok := data[language]
	if !ok {
		log.Warnf(color.YellowString("i18n: language [%s] not find!"), color.RedString(language))
		return ""
	}
	v, ok := tData[key]
	if !ok {
		log.Warnf(color.YellowString("i18n: language [%s] translated words  %s  not find!"), color.GreenString(language), color.RedString(key))
		return ""
	}
	return v
}
