package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"sort"
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
	var outter = newOutter()
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
			if err := outter.loadConfig(); err != nil {
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
			outter.DataSource = dataSource
			// 读取主题模板文件
			if err := outter.readTheme(outter.Config.Site.Theme); err != nil {
				logger.Error("read theme source failed", "reason", err)
				return
			}
			// 读取页面列表
			if err := outter.readAllPages(); err != nil {
				logger.Error("read page source failed", "reason", err)
				return
			}
			// 读取文章列表
			if err := outter.readAllPosts(); err != nil {
				logger.Error("read post source failed", "reason", err)
				return
			}
			// 数据写入模板文件
			outter.generate()
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
func (o *Outter) readAllPosts() error {
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
			o.Posts = append(o.Posts, o.readPosts(postDir, info.Name())...)
			continue
		}
		if !strings.HasSuffix(info.Name(), ".md") {
			continue
		}
		post, err := o.readPost(fmt.Sprintf("%s/%s", postDir, info.Name()))
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
			o.DraftPosts = append(o.DraftPosts, post)
		} else {
			o.Posts = append(o.Posts, post)
		}
	}
	sort.Slice(o.Posts, func(i, j int) bool {
		return o.Posts[i].CreatedAt.Unix() > o.Posts[j].CreatedAt.Unix()
	})
	return nil
}

func (o *Outter) readPost(fp string) (*Post, error) {
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
	post.CreatedAt, err = datetimeStringToTime(mustString(metaData["date"]))
	if err != nil {
		logger.Error("get post created_at time failed, please check file content", "name", post.FileBasename)
		return nil, err
	}
	post.CreatedAtFormat = post.CreatedAt.Format("2006-01-02")
	post.Title = mustString(metaData["title"])
	post.Tags = mustStringSlice(metaData["tags"])
	post.Category = mustString(metaData["category"])
	post.MD = buf.String()
	return post, nil
}

func (o *Outter) readPage(fp string) (*Page, error) {
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
	page.Title = mustString(metaData["title"])
	page.MD = buf.String()
	return page, nil
}

func (o *Outter) readPosts(base, category string) []*Post {
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
		post, err := o.readPost(fmt.Sprintf("%s/%s", dir, info.Name()))
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
			o.DraftPosts = append(o.DraftPosts, post)
		} else {
			posts = append(posts, post)
		}
	}
	return posts
}

func (o *Outter) readAllPages() error {
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
		page, err := o.readPage(fmt.Sprintf("%s/%s", dir, info.Name()))
		if err != nil {
			logger.Error("read page file failed", "reason", err)
			continue
		}
		page.Link = strings.ReplaceAll(info.Name(), ".md", "")
		o.Pages = append(o.Pages, page)
	}
	return nil
}

func (o *Outter) readTheme(themeName string) (err error) {
	o.Theme = new(Theme)
	var dir = fmt.Sprintf("%s/themes/%s", BaseDir, themeName)
	o.Theme.BaseLayout, err = os.ReadFile(fmt.Sprintf("%s/base.html", dir))
	if err != nil {
		logger.Error("read base.html theme file failed", "reason", err)
		return err
	}
	o.Theme.IndexLayout, err = os.ReadFile(fmt.Sprintf("%s/index.html", dir))
	if err != nil {
		logger.Error("read index.html theme file failed", "reason", err)
		return err
	}
	o.Theme.PageLayout, err = os.ReadFile(fmt.Sprintf("%s/page.html", dir))
	if err != nil {
		logger.Error("read page.html theme file failed", "reason", err)
		return err
	}
	o.Theme.PostLayout, err = os.ReadFile(fmt.Sprintf("%s/post.html", dir))
	if err != nil {
		logger.Error("read pose.html theme file failed", "reason", err)
		return err
	}
	o.Theme.ArchiveLayout, err = os.ReadFile(fmt.Sprintf("%s/archive.html", dir))
	if err != nil {
		logger.Error("read archive.html theme file failed", "reason", err)
		return err
	}
	o.Theme.TagLayout, err = os.ReadFile(fmt.Sprintf("%s/tag.html", dir))
	if err != nil {
		logger.Error("read tag.html theme file failed", "reason", err)
		return err
	}
	o.Theme.CategoryLayout, err = os.ReadFile(fmt.Sprintf("%s/category.html", dir))
	if err != nil {
		logger.Error("read category.html theme file failed", "reason", err)
		return err
	}

	o.Theme.IndexLayout = bytes.ReplaceAll(o.Theme.BaseLayout, []byte("{{layout_placeholder}}"), o.Theme.IndexLayout)
	o.Theme.PageLayout = bytes.ReplaceAll(o.Theme.BaseLayout, []byte("{{layout_placeholder}}"), o.Theme.PageLayout)
	o.Theme.PostLayout = bytes.ReplaceAll(o.Theme.BaseLayout, []byte("{{layout_placeholder}}"), o.Theme.PostLayout)
	o.Theme.ArchiveLayout = bytes.ReplaceAll(o.Theme.BaseLayout, []byte("{{layout_placeholder}}"), o.Theme.ArchiveLayout)
	o.Theme.TagLayout = bytes.ReplaceAll(o.Theme.BaseLayout, []byte("{{layout_placeholder}}"), o.Theme.TagLayout)
	o.Theme.CategoryLayout = bytes.ReplaceAll(o.Theme.BaseLayout, []byte("{{layout_placeholder}}"), o.Theme.TagLayout)
	return nil
}

type Outter struct {
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

func newOutter() *Outter {
	return &Outter{
		SourceVersion: randString(8),
		Config:        new(Config),
		Now:           time.Now(),
	}
}

// loadConfig 加载配置文件
func (o *Outter) loadConfig() error {
	if o.Config == nil {
		o.Config = new(Config)
	}
	return o.Config.load()
}

// generate 文件生成
func (o *Outter) generate() {
	// 清理dist目录
	if err := o.clearDistDirectory(); err != nil {
		logger.Error("clear dist directory failed", "reason", err)
		return
	}
	if err := o.sourceCopy(); err != nil {
		logger.Error("copy theme source to dist failed", "reason", err)
		return
	}
	_ = o.generateIndex()
	o.generatePost()
	o.generateDraftPost()
	o.generatePage()
	o.generateArchives()
	o.generateTags()
	o.generateCategories()
}

// clearDistDirectory 清理dist目录
func (o *Outter) clearDistDirectory() error {
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
func (o *Outter) sourceCopy() error {
	sourcePath := slash(fmt.Sprintf(BaseDir+"/themes/%s/", o.Config.Site.Theme))
	destPath := slash(BaseDir + "/dist/" + o.SourceVersion + "/")
	// mkdir
	cssDir := fmt.Sprintf("%s/css", destPath)
	jsDir := fmt.Sprintf("%s/js", destPath)
	imagesDir := fmt.Sprintf("%s/images", destPath)
	if err := mkdir(cssDir); err != nil {
		logger.Error("mkdir css directory failed", "reason", err)
		return err
	}
	if err := mkdir(jsDir); err != nil {
		logger.Error("mkdir js directory failed", "reason", err)
	}
	if err := mkdir(imagesDir); err != nil {
		logger.Error("mkdir images directory failed", "reason", err)
	}
	cmd := exec.Command("cp", "-r", sourcePath+"css", destPath)
	if err := cmd.Run(); err != nil {
		logger.Info(cmd.String())
		logger.Error("copy theme css source failed", "reason", err)
		return err
	}
	cmd = exec.Command("cp", "-r", sourcePath+"js", destPath)
	if err := cmd.Run(); err != nil {
		logger.Info(cmd.String())
		logger.Error("copy theme js source failed", "reason", err)
		return err
	}
	cmd = exec.Command("cp", "-r", sourcePath+"images", destPath)
	if err := cmd.Run(); err != nil {
		logger.Info(cmd.String())
		logger.Error("copy theme images source failed", "reason", err)
		return err
	}
	return nil
}

// generateIndex 首页生成
func (o *Outter) generateIndex() error {
	indexTemplate := template.New("index")

	funcMap["title"] = func() string {
		return o.Config.Site.Title
	}

	indexTemplate.Funcs(funcMap)

	indexTemplate, err := indexTemplate.Parse(string(o.Theme.IndexLayout))
	if err != nil {
		logger.Error("parse index layout failed", "reason", err)
		return err
	}

	// 首页文件
	var buffer = bytes.Buffer{}

	// 没开分页
	if !o.Config.PageConfig.Paginate {
		err = indexTemplate.Execute(&buffer, o)
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
	if o.Config.PageConfig.Size < 1 {
		logger.Error("page size must > 1")
		return err
	}

	var pageSize = int(math.Ceil(float64(len(o.Posts)) / float64(o.Config.PageConfig.Size)))
	o.Config.PageConfig.Total = pageSize
	var posts = make([]*Post, len(o.Posts))
	copy(posts, o.Posts)
	for i := 0; i < pageSize; i++ {
		rightBorder := int(o.Config.PageConfig.Size) * (i + 1)
		if rightBorder > len(posts) {
			rightBorder = len(posts)
		}
		leftBorder := int(o.Config.PageConfig.Size) * (i)
		o.Posts = make([]*Post, rightBorder-leftBorder)
		copy(o.Posts, posts[leftBorder:rightBorder])
		o.Config.PageConfig.CurrentSize = i + 1

		err = indexTemplate.Execute(&buffer, o)
		if err != nil {
			logger.Error("generate index page failed", "reason", err)
			return err
		}
		// 第一页 主页
		if i == 0 {
			// 第一次的时候创建目录
			dir := BaseDir + "/dist/page"
			if err := o.createDir(dir); err != nil {
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
	o.Posts = make([]*Post, len(posts))
	copy(o.Posts, posts)
	return nil
}

func (o *Outter) generatePost() {
	postTemplate := template.New("post")
	var postBuffer = new(bytes.Buffer)
	for _, post := range o.Posts {
		instance := PostOutter{
			Post:   post,
			Config: o.Config,
		}
		instance.Outter = o
		funcMap["title"] = func() string {
			return instance.Post.Title
		}
		funcMap["post_name"] = func() string {
			return o.Config.Site.Title
		}
		postTemplate.Funcs(funcMap)
		postTemplate, err := postTemplate.Parse(string(o.Theme.PostLayout))
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
		var catDir = fmt.Sprintf(BaseDir+"/dist/%s", post.Category)
		if !isExist(catDir) {
			if err := mkdir(catDir); err != nil {
				logger.Error("create directory failed", "reason", err)
				return
			}
		}
		var filename = fmt.Sprintf("%s/%s.html", catDir, post.FileBasename)
		if err := os.WriteFile(filename, postBuffer.Bytes(), os.ModePerm); err != nil {
			logger.Error("write file failed", "reason", err)
			return
		}
		postBuffer.Reset()
	}
}

func (o *Outter) generateDraftPost() {
	postTemplate := template.New("post")
	var postBuffer = new(bytes.Buffer)
	for _, post := range o.DraftPosts {
		instance := PostOutter{
			Post:   post,
			Config: o.Config,
		}
		instance.Outter = o
		funcMap["title"] = func() string {
			return instance.Post.Title
		}
		funcMap["post_name"] = func() string {
			return o.Config.Site.Title
		}
		postTemplate.Funcs(funcMap)
		postTemplate, err := postTemplate.Parse(string(o.Theme.PostLayout))
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
		var catDir = BaseDir + "/dist/draft"
		if !isExist(catDir) {
			if err := mkdir(catDir); err != nil {
				logger.Error("create directory failed", "reason", err)
				return
			}
		}
		var filename = fmt.Sprintf("%s/%s.html", catDir, post.FileBasename)
		if err := os.WriteFile(filename, postBuffer.Bytes(), os.ModePerm); err != nil {
			logger.Error("write file failed", "reason", err)
			return
		}
		postBuffer.Reset()
	}
}

func (o *Outter) generatePage() {
	pageTemplate := template.New("page")
	var pageBuffer = new(bytes.Buffer)
	for _, page := range o.Pages {
		instance := PageOutter{
			Page:   page,
			Config: o.Config,
		}
		instance.Outter = o
		funcMap["title"] = func() string {
			return page.Title
		}
		funcMap["page_name"] = func() string {
			return o.Config.Site.Title
		}
		pageTemplate.Funcs(funcMap)
		pageTemplate, err := pageTemplate.Parse(string(o.Theme.PageLayout))
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

func (o *Outter) generateArchives() {
	archiveTemplate := template.New("archive")
	var archiveBuffer = new(bytes.Buffer)
	// 按时间归档
	var m = make(map[string][]*Post)
	for _, post := range o.Posts {
		newPost := post
		newPost.MD = ""
		m[int2String(post.CreatedAt.Year())] = append(m[int2String(post.CreatedAt.Year())], newPost)
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

	var instance = PageOutter{
		Page: &Page{
			Title: "Archives",
			Link:  "archives",
		},
		Config:   o.Config,
		PostData: postData,
	}
	instance.Outter = o
	funcMap["title"] = func() string {
		return instance.Page.Title
	}
	funcMap["page_name"] = func() string {
		return o.Config.Site.Title
	}
	archiveTemplate.Funcs(funcMap)
	archiveTemplate, err := archiveTemplate.Parse(string(o.Theme.ArchiveLayout))
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

func (o *Outter) generateTags() {
	tagsTemplate := template.New("tags")
	var tagsBuffer = new(bytes.Buffer)

	// 按标签归档
	var mTag = make(map[string][]*Post)
	for _, post := range o.Posts {
		for _, tag := range post.Tags {
			newPost := post
			newPost.MD = ""
			mTag[tag] = append(mTag[tag], newPost)
		}
	}

	for tag, posts := range mTag {
		// 按时间归档
		var m = make(map[string][]*Post)
		for _, post := range posts {
			newPost := post
			newPost.MD = ""
			m[int2String(post.CreatedAt.Year())] = append(m[int2String(post.CreatedAt.Year())], newPost)
		}

		var postData []*PostData
		for year, _posts := range m {
			postData = append(postData, &PostData{
				Key:   year,
				Posts: _posts,
			})
		}

		sort.Slice(postData, func(i, j int) bool {
			return postData[i].Key > postData[j].Key
		})

		var instance = PageOutter{
			Page: &Page{
				Title: tag,
				Link:  tag,
			},
			Config:   o.Config,
			PostData: postData,
		}
		instance.Outter = o
		funcMap["title"] = func() string {
			return instance.Page.Title
		}
		funcMap["page_name"] = func() string {
			return o.Config.Site.Title
		}
		tagsTemplate.Funcs(funcMap)
		tagsTemplate, err := tagsTemplate.Parse(string(o.Theme.TagLayout))
		if err != nil {
			logger.Error("generate tags page failed", "reason", err)
			return
		}

		// 入buffer
		if err := tagsTemplate.Execute(tagsBuffer, instance); err != nil {
			logger.Error("generate tags page failed", "reason", err)
			return
		}
		// 创建tag文件夹
		var tagDir = BaseDir + "/dist/tags"
		if !isExist(tagDir) {
			if err := mkdir(tagDir); err != nil {
				logger.Error("create tag directory failed", "reason", err)
				return
			}
		}
		// buffer写文件
		var filename = fmt.Sprintf("%s/%s.html", tagDir, instance.Page.Link)
		if err := os.WriteFile(filename, tagsBuffer.Bytes(), os.ModePerm); err != nil {
			logger.Error("write tag content failed", "reason", err)
			return
		}
		tagsBuffer.Reset()
	}
}

func (o *Outter) generateCategories() {
	tagsTemplate := template.New("categories")
	var tagsBuffer = new(bytes.Buffer)

	// 按标签归档
	var cats = make(map[string][]*Post)
	for _, post := range o.Posts {
		newPost := post
		cats[post.Category] = append(cats[post.Category], newPost)
	}
	cats["draft"] = o.DraftPosts

	for tag, posts := range cats {
		// 按时间归档
		var m = make(map[string][]*Post)
		for _, post := range posts {
			newPost := post
			newPost.MD = ""
			m[int2String(post.CreatedAt.Year())] = append(m[int2String(post.CreatedAt.Year())], newPost)
		}

		var postData []*PostData
		for year, _posts := range m {
			postData = append(postData, &PostData{
				Key:   year,
				Posts: _posts,
			})
		}

		sort.Slice(postData, func(i, j int) bool {
			return postData[i].Key > postData[j].Key
		})

		var instance = PageOutter{
			Page: &Page{
				Title: tag,
				Link:  tag,
			},
			Config:   o.Config,
			PostData: postData,
		}
		instance.Outter = o
		funcMap["title"] = func() string {
			return instance.Page.Title
		}
		funcMap["page_name"] = func() string {
			return o.Config.Site.Title
		}
		tagsTemplate.Funcs(funcMap)
		tagsTemplate, err := tagsTemplate.Parse(string(o.Theme.TagLayout))
		if err != nil {
			logger.Error("generate tags page failed", "reason", err)
			return
		}

		// 入buffer
		if err := tagsTemplate.Execute(tagsBuffer, instance); err != nil {
			logger.Error("generate tags page failed", "reason", err)
			return
		}
		// 创建tag文件夹
		var tagDir = BaseDir + "/dist/category"
		if !isExist(tagDir) {
			if err := mkdir(tagDir); err != nil {
				logger.Error("create tag directory failed", "reason", err)
				return
			}
		}
		// buffer写文件
		var filename = fmt.Sprintf("%s/%s.html", tagDir, instance.Page.Link)
		if err := os.WriteFile(filename, tagsBuffer.Bytes(), os.ModePerm); err != nil {
			logger.Error("write tag content failed", "reason", err)
			return
		}
		tagsBuffer.Reset()
	}
}

type PostOutter struct {
	Post   *Post
	Config *Config
	*Outter
}

type PageOutter struct {
	Page     *Page
	Config   *Config
	PostData []*PostData
	*Outter
}

type PostData struct {
	Key   string
	Posts []*Post
}

func (o *Outter) createDir(fp string) error {
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
	"sum":       sum,
}

func getSource(data internal.DataSource, key string) []*internal.DataItem {
	return data[key]
}

func add(a, b int) int {
	return a + b
}

func sum(a, b int) int {
	return a - b
}
