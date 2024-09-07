package internal

import "time"

// Post 文章属性
type Post struct {
	Title           string    // 文章标题
	FileBasename    string    // 文件名
	Link            string    // 链接
	Category        string    // 分类
	CategoryAlias   string    // 分类别名
	Tags            []string  // 标签
	CreatedAt       time.Time // 创建时间
	CreatedAtFormat string    // 创建时间格式化
	UpdatedAtFormat string    // 更新时间
	MD              string    // 文章内容
	TOC             string    // 文章toc
}

// Page 页面属性
type Page struct {
	Title string // 展示名
	Link  string // 链接名
	MD    string // 页面内容
}
