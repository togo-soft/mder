package main

import (
	"os"

	"github.com/goccy/go-yaml"
)

type Logo struct {
	Enabled bool   `yaml:"enabled"`
	Width   int64  `yaml:"width"`
	Height  int64  `yaml:"height"`
	Url     string `yaml:"url"`
	Favicon string `yaml:"favicon"`
}

type SocialLinks struct {
	Github   string `yaml:"github"`
	Email    string `yaml:"email"`
	QQ       string `yaml:"qq"`
	Wechat   string `yaml:"wechat"`
	Twitter  string `yaml:"twitter"`
	Telegram string `yaml:"telegram"`
}

type ICP struct {
	Enabled bool   `yaml:"enabled"`
	Url     string `yaml:"url"`
	Text    string `yaml:"text"`
}

type CDN struct {
	Enabled bool   `yaml:"enabled"`
	Url     string `yaml:"url"`
	Image   string `yaml:"image"`
	Text    string `yaml:"text"`
}

type Comment struct {
	Enabled bool `yaml:"enabled"`
}

type Site struct {
	Title       string `yaml:"title"`
	Subtitle    string `yaml:"subtitle"`
	Description string `yaml:"description"`
	Keywords    string `yaml:"keywords"`
	Author      string `yaml:"author"`
	Summary     string `yaml:"summary"`
	Theme       string `yaml:"theme"`
}

type PageConfig struct {
	Paginate    bool  `yaml:"paginate"`
	Size        int64 `yaml:"size"`
	Total       int
	CurrentSize int
}

type Config struct {
	Logo          Logo        `yaml:"logo"`
	SocialLinks   SocialLinks `yaml:"social_links"`
	ICP           ICP         `yaml:"icp"`
	CDN           CDN         `yaml:"cdn"`
	Comment       Comment     `yaml:"comment"`
	PageConfig    PageConfig  `yaml:"page"`
	Site          Site        `yaml:"site"`
	SourceVersion string      `yaml:"-"`
}

var BaseDir string

func (c *Config) load() error {
	file := "config.yaml"
	if BaseDir != "" {
		file = BaseDir + "/config.yaml"
	}
	configBuffer, err := os.ReadFile(file)
	if err != nil {
		logger.Error("read config file failed", "reason", err)
		return err
	}
	if err := yaml.Unmarshal(configBuffer, c); err != nil {
		logger.Error("read config file failed", "reason", err)
		return err
	}
	if c.Logo.Favicon == "" {
		c.Logo.Favicon = "favicon.ico"
	}
	return nil
}
