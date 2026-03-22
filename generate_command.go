package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/toc"

	"codeberg.org/mder/mder/internal"
)

func generateCmd() *cobra.Command {
	var startAt time.Time
	var gen = newGenerator()
	var path string
	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"g"},
		Short:   "generate project to dist folder",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if path != "" {
				BaseDir = strings.TrimSuffix(path, "/")
			}
			startAt = time.Now()
			// 读取配置文件
			if err := gen.loadConfig(); err != nil {
				logger.Error("load config file failed", "reason", err)
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			// 读取资源文件
			dataSource, err := internal.GetDataSource(BaseDir)
			if err != nil {
				logger.Error("get data source failed", "reason", err)
				return
			}
			gen.DataSource = dataSource
			// 读取主题模板文件
			if err := gen.readTheme(gen.Config.Site.Theme); err != nil {
				logger.Error("read theme source failed", "reason", err)
				return
			}
			// 读取页面列表
			if err := gen.readAllPages(); err != nil {
				logger.Error("read page source failed", "reason", err)
				return
			}
			// 读取文章列表
			if err := gen.readAllPosts(); err != nil {
				logger.Error("read post source failed", "reason", err)
				return
			}
			// 数据写入模板文件
			gen.generate()
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			endAt := time.Since(startAt)
			logger.Info("generate success", "cost", endAt.String())
		},
	}
	cmd.Flags().StringVar(&path, "path", ".", "mder project path")
	return cmd
}

// readAllPosts 读取所有post文章
func (g *Generator) readAllPosts() error {
	var postDir = BaseDir + "/posts"
	dirs, err := os.ReadDir(postDir)
	if err != nil {
		logger.Error("read directory failed", "reason", err)
		return err
	}
	for _, info := range dirs {
		// 获取文件信息
		fsInfo, err := info.Info()
		if err != nil {
			logger.Error("read file info failed", "reason", err)
			continue
		}
		// 不是目录 没有分类
		if info.IsDir() {
			g.Posts = append(g.Posts, g.readPosts(postDir, info.Name())...)
			continue
		}
		if !strings.HasSuffix(info.Name(), ".md") {
			continue
		}
		post, err := g.readPost(fmt.Sprintf("%s/%s", postDir, info.Name()))
		if err != nil {
			logger.Error("read post file failed", "reason", err)
			continue
		}
		post.UpdatedAtFormat = fsInfo.ModTime().Format(time.DateOnly)
		post.FileBasename = strings.ReplaceAll(fsInfo.Name(), ".md", "")
		post.Link = fmt.Sprintf("/default/%s.html", post.FileBasename)
		post.Category = "default"
		if isDraft(post.Tags) {
			post.Link = fmt.Sprintf("/draft/%s.html", post.FileBasename)
			g.DraftPosts = append(g.DraftPosts, post)
		} else {
			g.Posts = append(g.Posts, post)
		}
	}
	sort.Slice(g.Posts, func(i, j int) bool {
		return g.Posts[i].CreatedAt.Unix() > g.Posts[j].CreatedAt.Unix()
	})
	return nil
}

func (g *Generator) readPost(fp string) (*Post, error) {
	var post = new(Post)
	content, err := os.ReadFile(fp)
	if err != nil {
		logger.Error("read file failed", "file", fp, "reason", err)
		return nil, err
	}
	var buf bytes.Buffer
	var parserContext = parser.NewContext()
	doc := markdown.Parser().Parse(text.NewReader(content), parser.WithContext(parserContext))

	tocTree, err := toc.Inspect(doc, content)
	if err != nil {
		logger.Error("read content toc tree failed", "reason", err)
		return nil, err
	}

	list := toc.RenderList(tocTree)
	if list != nil {
		if err := markdown.Renderer().Render(&buf, content, list); err != nil {
			logger.Error("render toc content failed", "reason", err)
			return nil, err
		}
		post.TOC = buf.String()
		buf.Reset()
	}
	if err := markdown.Renderer().Render(&buf, content, doc); err != nil {
		logger.Error("render content failed", "reason", err)
		return nil, err
	}

	metaData := meta.Get(parserContext)
	post.CreatedAt, err = datetimeStringToTime(toString(metaData["date"]))
	if err != nil {
		logger.Error("get post created_at time failed, please check file content", "name", post.FileBasename)
		return nil, err
	}
	post.CreatedAtFormat = post.CreatedAt.Format("2006-01-02")
	post.Title = toString(metaData["title"])
	post.Tags = toStringSlice(metaData["tags"])
	post.Category = toString(metaData["category"])
	post.MD = buf.String()
	return post, nil
}

func (g *Generator) readPage(fp string) (*Page, error) {
	var page = new(Page)
	content, err := os.ReadFile(fp)
	if err != nil {
		logger.Error("read page file failed", "reason", err)
		return nil, err
	}
	var buf bytes.Buffer
	context := parser.NewContext()
	if err := markdown.Convert(content, &buf, parser.WithContext(context)); err != nil {
		logger.Error("render content failed", "reason", err)
		return nil, err
	}
	metaData := meta.Get(context)
	page.Title = toString(metaData["title"])
	page.MD = buf.String()
	return page, nil
}

func (g *Generator) readPosts(base, category string) []*Post {
	var dir = fmt.Sprintf("%s/%s", base, category)
	dirs, err := os.ReadDir(dir)
	if err != nil {
		logger.Error("read directory failed", "reason", err)
		return nil
	}
	var posts []*Post
	for _, info := range dirs {
		if info.IsDir() {
			continue
		}
		if !strings.HasSuffix(info.Name(), ".md") {
			continue
		}
		// 获取文件信息
		fsInfo, err := info.Info()
		if err != nil {
			logger.Error("read file info failed", "reason", err)
			return nil
		}
		post, err := g.readPost(fmt.Sprintf("%s/%s", dir, info.Name()))
		if err != nil {
			logger.Error("read post failed", "reason", err)
			continue
		}
		post.UpdatedAtFormat = fsInfo.ModTime().Format(time.DateOnly)
		post.FileBasename = strings.ReplaceAll(fsInfo.Name(), ".md", "")
		post.Link = fmt.Sprintf("/%s/%s.html", strings.ToLower(category), post.FileBasename)
		post.Category = category
		if isDraft(post.Tags) {
			post.Link = fmt.Sprintf("/draft/%s.html", post.FileBasename)
			g.DraftPosts = append(g.DraftPosts, post)
		} else {
			posts = append(posts, post)
		}
	}
	return posts
}

func (g *Generator) readAllPages() error {
	var dir = BaseDir + "/pages"
	dirs, err := os.ReadDir(dir)
	if err != nil {
		logger.Error("read directory failed", "reason", err)
		return err
	}
	for _, info := range dirs {
		if info.IsDir() {
			continue
		}
		if !strings.HasSuffix(info.Name(), ".md") {
			continue
		}
		page, err := g.readPage(fmt.Sprintf("%s/%s", dir, info.Name()))
		if err != nil {
			logger.Error("read page file failed", "reason", err)
			continue
		}
		page.Link = strings.ReplaceAll(info.Name(), ".md", "")
		g.Pages = append(g.Pages, page)
	}
	return nil
}

func (g *Generator) readTheme(themeName string) error {
	g.Theme = new(Theme)
	dir := fmt.Sprintf("%s/themes/%s", BaseDir, themeName)

	type layoutEntry struct {
		name   string
		target *[]byte
	}
	entries := []layoutEntry{
		{"base", &g.Theme.BaseLayout},
		{"index", &g.Theme.IndexLayout},
		{"page", &g.Theme.PageLayout},
		{"post", &g.Theme.PostLayout},
		{"archive", &g.Theme.ArchiveLayout},
		{"tag", &g.Theme.TagLayout},
		{"category", &g.Theme.CategoryLayout},
	}

	for _, e := range entries {
		data, err := os.ReadFile(fmt.Sprintf("%s/%s.html", dir, e.name))
		if err != nil {
			logger.Error(fmt.Sprintf("read %s.html theme file failed", e.name), "reason", err)
			return err
		}
		*e.target = data
	}

	for _, e := range entries[1:] {
		*e.target = bytes.ReplaceAll(g.Theme.BaseLayout, []byte("{{layout_placeholder}}"), *e.target)
	}

	return nil
}

// Generator holds all data needed for site generation
type Generator struct {
	SourceVersion string              // 资源号 防缓存
	Config        *Config             // 全局配置
	DataSource    internal.DataSource // 记录source/data下的所有文件
	Posts         []*Post             // 记录文章
	DraftPosts    []*Post             // 不宜发布的草稿文章
	Pages         []*Page             // 记录页面
	Theme         *Theme              // 记录主题模板文件
	Now           time.Time           // 当前时间
}

type Theme struct {
	Name           string // 主题名
	BaseLayout     []byte // 基本布局
	PostLayout     []byte // 文章布局
	PageLayout     []byte // 页面布局
	IndexLayout    []byte // 首页布局
	ArchiveLayout  []byte // 文章归档布局
	TagLayout      []byte // 标签归档布局
	CategoryLayout []byte // 分类归档布局
}

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

func newGenerator() *Generator {
	return &Generator{
		SourceVersion: randString(8),
		Config:        new(Config),
		Now:           time.Now(),
	}
}

// loadConfig 加载配置文件
func (g *Generator) loadConfig() error {
	if g.Config == nil {
		g.Config = new(Config)
	}
	if err := g.Config.load(); err != nil {
		return err
	}
	g.Config.SourceVersion = g.SourceVersion
	return nil
}

// generate 文件生成
func (g *Generator) generate() {
	// 清理dist目录
	if err := g.clearDistDirectory(); err != nil {
		logger.Error("clear dist directory failed", "reason", err)
		return
	}
	if err := g.sourceCopy(); err != nil {
		logger.Error("copy theme source to dist failed", "reason", err)
		return
	}
	_ = g.generateIndex()
	g.generatePosts(g.Posts, "")
	g.generatePosts(g.DraftPosts, "draft")
	g.generatePage()
	g.generateArchives()
	g.generateTags()
	g.generateCategories()
}

// clearDistDirectory 清理dist目录
func (g *Generator) clearDistDirectory() error {
	if err := os.RemoveAll(BaseDir + "/dist"); err != nil {
		logger.Error("remove old dist directory failed", "reason", err)
		return err
	}

	if err := mkdir(BaseDir + "/dist"); err != nil {
		logger.Error("create dist directory failed", "reason", err)
		return err
	}
	return nil
}

// sourceCopy 资源拷贝 将主题的资源拷贝到目标文件夹中
func (g *Generator) sourceCopy() error {
	sourceDir := fmt.Sprintf("%s/themes/%s", BaseDir, g.Config.Site.Theme)
	destDir := BaseDir + "/dist/" + g.SourceVersion

	for _, subDir := range []string{"css", "js", "images"} {
		src := sourceDir + "/" + subDir
		dst := destDir + "/" + subDir
		if err := copyDir(src, dst); err != nil {
			logger.Error("copy theme "+subDir+" source failed", "reason", err)
			return err
		}
	}
	return nil
}

// generateIndex 首页生成
func (g *Generator) generateIndex() error {
	indexTemplate := template.New("index")

	funcMap["title"] = func() string {
		return g.Config.Site.Title
	}

	indexTemplate.Funcs(funcMap)

	indexTemplate, err := indexTemplate.Parse(string(g.Theme.IndexLayout))
	if err != nil {
		logger.Error("parse index layout failed", "reason", err)
		return err
	}

	// 首页文件
	var buffer = bytes.Buffer{}

	// 没开分页
	if !g.Config.PageConfig.Paginate {
		err = indexTemplate.Execute(&buffer, g)
		if err != nil {
			logger.Error("generate index page failed", "reason", err)
			return err
		}
		var filename = BaseDir + "/dist/index.html"
		if err := os.WriteFile(filename, buffer.Bytes(), os.ModePerm); err != nil {
			logger.Error("write index file failed", "reason", err)
			return err
		}
		logger.Info("index generate success")
		return nil
	}

	// 分页数据不规范
	if g.Config.PageConfig.Size < 1 {
		logger.Error("page size must > 1")
		return err
	}

	var pageSize = int(math.Ceil(float64(len(g.Posts)) / float64(g.Config.PageConfig.Size)))
	g.Config.PageConfig.Total = pageSize
	for i := 0; i < pageSize; i++ {
		rightBorder := int(g.Config.PageConfig.Size) * (i + 1)
		if rightBorder > len(g.Posts) {
			rightBorder = len(g.Posts)
		}
		leftBorder := int(g.Config.PageConfig.Size) * i
		g.Config.PageConfig.CurrentSize = i + 1

		pageView := *g
		pageView.Posts = g.Posts[leftBorder:rightBorder]

		err = indexTemplate.Execute(&buffer, &pageView)
		if err != nil {
			logger.Error("generate index page failed", "reason", err)
			return err
		}
		// 第一页 主页
		if i == 0 {
			// 第一次的时候创建目录
			dir := BaseDir + "/dist/page"
			if err := g.createDir(dir); err != nil {
				logger.Error("create index page failed", "reason", err)
				return err
			}
			var filename = BaseDir + "/dist/index.html"
			if err := os.WriteFile(filename, buffer.Bytes(), os.ModePerm); err != nil {
				logger.Error("write index file failed", "reason", err)
				return err
			}
		}
		fPage := fmt.Sprintf(BaseDir+"/dist/page/%d.html", i+1)
		if err := os.WriteFile(fPage, buffer.Bytes(), os.ModePerm); err != nil {
			logger.Error("write index file failed", "reason", err)
			return err
		}
		buffer.Reset()
	}
	return nil
}

func (g *Generator) generatePosts(posts []*Post, fixedOutputDir string) {
	postTemplate := template.New("post")
	var postBuffer = new(bytes.Buffer)
	for _, post := range posts {
		instance := PostContext{
			Post:   post,
			Config: g.Config,
		}
		instance.Generator = g
		funcMap["title"] = func() string {
			return instance.Post.Title
		}
		funcMap["post_name"] = func() string {
			return g.Config.Site.Title
		}
		postTemplate.Funcs(funcMap)
		postTemplate, err := postTemplate.Parse(string(g.Theme.PostLayout))
		if err != nil {
			logger.Error("generate post page failed", "reason", err)
			return
		}
		// 入buffer
		if err := postTemplate.Execute(postBuffer, instance); err != nil {
			logger.Error("generate post page failed", "reason", err)
			return
		}
		// buffer写文件
		outputDir := fixedOutputDir
		if outputDir == "" {
			outputDir = post.Category
		}
		var catDir = fmt.Sprintf(BaseDir+"/dist/%s", outputDir)
		if err := g.createDir(catDir); err != nil {
			return
		}
		var filename = fmt.Sprintf("%s/%s.html", catDir, post.FileBasename)
		if err := os.WriteFile(filename, postBuffer.Bytes(), os.ModePerm); err != nil {
			logger.Error("write file failed", "reason", err)
			return
		}
		postBuffer.Reset()
	}
}

func (g *Generator) generatePage() {
	pageTemplate := template.New("page")
	var pageBuffer = new(bytes.Buffer)
	for _, page := range g.Pages {
		instance := PageContext{
			Page:   page,
			Config: g.Config,
		}
		instance.Generator = g
		funcMap["title"] = func() string {
			return page.Title
		}
		funcMap["page_name"] = func() string {
			return g.Config.Site.Title
		}
		pageTemplate.Funcs(funcMap)
		pageTemplate, err := pageTemplate.Parse(string(g.Theme.PageLayout))
		if err != nil {
			logger.Error("generate page page failed", "reason", err)
			return
		}
		// 入buffer
		if err := pageTemplate.Execute(pageBuffer, instance); err != nil {
			logger.Error("generate page page failed", "reason", err)
			return
		}
		// buffer写文件
		var filename = fmt.Sprintf(BaseDir+"/dist/%s.html", page.Link)
		if err := os.WriteFile(filename, pageBuffer.Bytes(), os.ModePerm); err != nil {
			logger.Error("write page content failed", "reason", err)
			return
		}
		pageBuffer.Reset()
	}
}

func (g *Generator) generateArchives() {
	archiveTemplate := template.New("archive")
	var archiveBuffer = new(bytes.Buffer)
	// 按时间归档
	var m = make(map[string][]*Post)
	for _, post := range g.Posts {
		newPost := post
		newPost.MD = ""
		m[strconv.Itoa(post.CreatedAt.Year())] = append(m[strconv.Itoa(post.CreatedAt.Year())], newPost)
	}

	var postData []*PostData
	for year, posts := range m {
		postData = append(postData, &PostData{
			Key:   year,
			Posts: posts,
		})
	}

	sort.Slice(postData, func(i, j int) bool {
		return postData[i].Key > postData[j].Key
	})

	var instance = PageContext{
		Page: &Page{
			Title: "Archives",
			Link:  "archives",
		},
		Config:   g.Config,
		PostData: postData,
	}
	instance.Generator = g
	funcMap["title"] = func() string {
		return instance.Page.Title
	}
	funcMap["page_name"] = func() string {
		return g.Config.Site.Title
	}
	archiveTemplate.Funcs(funcMap)
	archiveTemplate, err := archiveTemplate.Parse(string(g.Theme.ArchiveLayout))
	if err != nil {
		logger.Error("generate archive page failed", "reason", err)
		return
	}

	// 入buffer
	if err := archiveTemplate.Execute(archiveBuffer, instance); err != nil {
		logger.Error("generate archive page failed", "reason", err)
		return
	}
	// buffer写文件
	var filename = fmt.Sprintf(BaseDir+"/dist/%s.html", instance.Page.Link)
	if err := os.WriteFile(filename, archiveBuffer.Bytes(), os.ModePerm); err != nil {
		logger.Error("write archive page failed", "reason", err)
		return
	}
	archiveBuffer.Reset()
}

func (g *Generator) generateTags() {
	tagGroups := make(map[string][]*Post)
	for _, post := range g.Posts {
		for _, tag := range post.Tags {
			tagGroups[tag] = append(tagGroups[tag], post)
		}
	}
	g.generateGroupedArchive("tags", g.Theme.TagLayout, tagGroups, "tags")
}

func (g *Generator) generateCategories() {
	catGroups := make(map[string][]*Post)
	for _, post := range g.Posts {
		catGroups[post.Category] = append(catGroups[post.Category], post)
	}
	catGroups["draft"] = g.DraftPosts
	g.generateGroupedArchive("categories", g.Theme.CategoryLayout, catGroups, "category")
}

func (g *Generator) generateGroupedArchive(templateName string, layout []byte, groups map[string][]*Post, outputDir string) {
	tmpl := template.New(templateName)
	var buf bytes.Buffer

	for name, posts := range groups {
		// 按时间归档
		yearMap := make(map[string][]*Post)
		for _, post := range posts {
			newPost := post
			newPost.MD = ""
			yearMap[strconv.Itoa(post.CreatedAt.Year())] = append(yearMap[strconv.Itoa(post.CreatedAt.Year())], newPost)
		}

		var postData []*PostData
		for year, yearPosts := range yearMap {
			postData = append(postData, &PostData{
				Key:   year,
				Posts: yearPosts,
			})
		}

		sort.Slice(postData, func(i, j int) bool {
			return postData[i].Key > postData[j].Key
		})

		instance := PageContext{
			Page: &Page{
				Title: name,
				Link:  name,
			},
			Config:   g.Config,
			PostData: postData,
		}
		instance.Generator = g
		funcMap["title"] = func() string {
			return instance.Page.Title
		}
		funcMap["page_name"] = func() string {
			return g.Config.Site.Title
		}
		tmpl.Funcs(funcMap)
		tmpl, err := tmpl.Parse(string(layout))
		if err != nil {
			logger.Error("generate "+templateName+" page failed", "reason", err)
			return
		}

		// 入buffer
		if err := tmpl.Execute(&buf, instance); err != nil {
			logger.Error("generate "+templateName+" page failed", "reason", err)
			return
		}
		// 创建目录
		dir := BaseDir + "/dist/" + outputDir
		if err := g.createDir(dir); err != nil {
			return
		}
		// buffer写文件
		filename := fmt.Sprintf("%s/%s.html", dir, instance.Page.Link)
		if err := os.WriteFile(filename, buf.Bytes(), os.ModePerm); err != nil {
			logger.Error("write "+templateName+" content failed", "reason", err)
			return
		}
		buf.Reset()
	}
}

type PostContext struct {
	Post   *Post
	Config *Config
	*Generator
}

type PageContext struct {
	Page     *Page
	Config   *Config
	PostData []*PostData
	*Generator
}

type PostData struct {
	Key   string
	Posts []*Post
}

func (g *Generator) createDir(fp string) error {
	if !isExist(fp) {
		if err := mkdir(fp); err != nil {
			logger.Error("create directory failed", "dir", fp, "reason", err)
			return err
		}
	}
	return nil
}

var funcMap = template.FuncMap{
	"getSource": getSource,
	"add":       add,
	"sub":       sub,
}

func getSource(data internal.DataSource, key string) []*internal.DataItem {
	return data[key]
}

func add(a, b int) int {
	return a + b
}

func sub(a, b int) int {
	return a - b
}
