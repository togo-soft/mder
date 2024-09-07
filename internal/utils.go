package internal

import (
	"os"

	"gopkg.in/yaml.v3"
)

func ReadYamlToDataItems(filename string) ([]*DataItem, error) {
	readFile, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var dataItems []*DataItem
	err = yaml.Unmarshal(readFile, &dataItems)
	return dataItems, err
}
