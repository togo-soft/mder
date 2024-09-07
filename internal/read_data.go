package internal

import (
	"os"
	"path/filepath"

	"github.com/samber/lo"
)

type DataItem struct {
	Name string `yaml:"name"`
	Link string `yaml:"link"`
	Desc string `yaml:"desc"`
}

type DataSource map[string][]*DataItem

// GetDataSource 读取数据目录 数据目录不支持嵌套读入
func GetDataSource(dir string) (DataSource, error) {
	dir = filepath.Join(dir, "data")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var datasource = make(DataSource)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		// 不是json文件
		ext := filepath.Ext(entry.Name())
		// 没有指定扩展名
		if ext == "" {
			continue
		}
		if !lo.Contains([]string{".yaml", ".yml"}, ext) {
			continue
		}
		filename := entry.Name()[:len(entry.Name())-len(filepath.Ext(entry.Name()))]
		dataItems, err := ReadYamlToDataItems(filepath.Join(dir, entry.Name()))
		if err != nil {
			panic("read data source file failed: " + err.Error())
		}
		datasource[filename] = dataItems
	}
	return datasource, nil
}
