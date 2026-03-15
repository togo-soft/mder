package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Logo 头像配置
type Logo struct {
	Enabled bool   `yaml:"enabled"` // 显示或隐藏
	Width   int64  `yaml:"width"`   // 宽度控制
	Height  int64  `yaml:"height"`  // 高度控制
	URL     string `yaml:"url"`     // 源url
}

// Favicon 自定义favicon
type Favicon struct {
	URL string `yaml:"url"`
}

// SocialLinks 社交媒体链接
type SocialLinks struct {
	Github   string `yaml:"github"`
	Email    string `yaml:"email"`
	QQ       string `yaml:"qq"`
	Wechat   string `yaml:"wechat"`
	Twitter  string `yaml:"twitter"`
	Telegram string `yaml:"telegram"`
}

// ICP 备案配置
type ICP struct {
	Enabled bool   `yaml:"enabled"` // 显示或隐藏
	URL     string `yaml:"url"`     // 重定向地址
	Text    string `yaml:"text"`    // 内容
}

// CDN 厂商信息
type CDN struct {
	Enabled bool   `yaml:"enabled"`
	URL     string `yaml:"url"`
	Image   string `yaml:"image"`
	Text    string `yaml:"text"`
}

// Comment 评论功能
type Comment struct {
	Enabled bool `yaml:"enabled"`
}

// Site 站点配置
type Site struct {
	Title       string `yaml:"title"`       // 网站标题
	Subtitle    string `yaml:"subtitle"`    // 副标题
	Description string `yaml:"description"` // 描述
	Keywords    string `yaml:"keywords"`    // 关键字
	Author      string `yaml:"author"`      // 您的名字
	Summary     string `yaml:"summary"`     // 个人总结
	Theme       string `yaml:"theme"`       // 主题
}

// PageConfig 页面配置
type PageConfig struct {
	Paginate    bool  `yaml:"paginate"` // 是否开启分页
	Size        int64 `yaml:"size"`     // 每页数
	Total       int   // 页总数
	CurrentSize int   // 当前页数
}

// Config 配置文件
type Config struct {
	Logo        Logo        `yaml:"logo"`
	Favicon     Favicon     `yaml:"favicon"`
	SocialLinks SocialLinks `yaml:"social_links"`
	ICP         ICP         `yaml:"icp"`
	CDN         CDN         `yaml:"cdn"`
	Comment     Comment     `yaml:"comment"`
	PageConfig  PageConfig  `yaml:"page"`
	Site        Site        `yaml:"site"` // 站点配置信息
}

var BaseDir string

func (c *Config) load() error {
	var file = "config.yaml"
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
	return nil
}
