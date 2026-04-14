package i18n

import (
	"path/filepath"
	"strings"

	"github.com/xslasd/goxf/conf"
	fileconfsource "github.com/xslasd/goxf/conf/filesource"
	"github.com/xslasd/goxf/utils/xfile"
)

type manager struct {
	data        map[string]map[string]string //Translating Data
	creatorFunc conf.DataSourceCreatorFunc

	language     string
	path         string
	exts         []string
	keyDelimiter string
}

func (m *manager) LanguageStr() string {
	return m.language
}

func newManager(opt *options) (*manager, error) {
	m := &manager{
		data:         make(map[string]map[string]string),
		language:     opt.config.Language,
		path:         opt.config.Path,
		keyDelimiter: opt.keyDelimiter,
		exts:         opt.config.Exts,
	}
	err := m.loadConf()
	if err != nil {
		return nil, err
	}
	return m, nil
}
func (m *manager) loadConf() error {
	files, err := xfile.LookupFilesByDirs(m.path, m.exts)
	if err != nil {
		return err
	}
	for _, name := range files {
		rel, err := filepath.Rel(m.path, name)
		if err != nil {
			return err
		}
		array := strings.Split(filepath.ToSlash(rel), "/")
		var lang string
		if len(array) > 1 {
			lang = array[0]
		} else {
			lang = strings.TrimSuffix(rel, filepath.Ext(name))
		}
		configuration, err := m.readConf(name)
		if err != nil {
			return err
		}
		_, ok := m.data[lang]
		if !ok {
			m.data[lang] = make(map[string]string)
		}
		merge(m.data, lang, configuration)
	}
	return nil
}
func (m *manager) readConf(fileName string) (map[string]string, error) {
	ds := fileconfsource.NewConfigSource(fileName, false)
	content, err := ds.ReadConfig()
	if err != nil {
		return nil, err
	}
	unmarshal, err := conf.ExtToUnmarshal(fileName)
	if err != nil {
		return nil, err
	}
	configuration := make(map[string]any)
	err = unmarshal(content, &configuration)
	if err != nil {
		return nil, err
	}
	return traverse(configuration, m.keyDelimiter, ""), nil
}
