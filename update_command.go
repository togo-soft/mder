package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/guonaihong/gout"
	"github.com/spf13/cobra"
)

func updateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "mder auto update",
		Example: "mder update",
		Run: func(cmd *cobra.Command, args []string) {
			commit, err := GetRepoLatestCommit()
			if err != nil {
				logger.Errorf("get mder version failed: %v", err)
				return
			}
			logger.Infof("-----\nmder latest version: %s\n更新中...\n", commit.Sha)
			var url = fmt.Sprintf("gitter.top/mder/mder@%s", commit.Sha)
			_, err = exec.Command("go", "install", url).CombinedOutput()
			if err != nil {
				logger.Errorf("install: go install %s\nupdate failed: %v\n", url, err)
				return
			}
			logger.Infof("mder %s update success", commit.Sha)
		},
	}

	return cmd
}

type RepoCommits []*RepoCommit

type RepoCommit struct {
	Author struct {
		Active            bool      `json:"active"`
		AvatarURL         string    `json:"avatar_url"`
		Created           time.Time `json:"created"`
		Description       string    `json:"description"`
		Email             string    `json:"email"`
		FollowersCount    int       `json:"followers_count"`
		FollowingCount    int       `json:"following_count"`
		FullName          string    `json:"full_name"`
		ID                int       `json:"id"`
		IsAdmin           bool      `json:"is_admin"`
		Language          string    `json:"language"`
		LastLogin         time.Time `json:"last_login"`
		Location          string    `json:"location"`
		Login             string    `json:"login"`
		ProhibitLogin     bool      `json:"prohibit_login"`
		Restricted        bool      `json:"restricted"`
		StarredReposCount int       `json:"starred_repos_count"`
		Visibility        string    `json:"visibility"`
		Website           string    `json:"website"`
	} `json:"author"`
	Commit struct {
		Author struct {
			Date  string `json:"date"`
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"author"`
		Committer struct {
			Date  string `json:"date"`
			Email string `json:"email"`
			Name  string `json:"name"`
		} `json:"committer"`
		Message string `json:"message"`
		Tree    struct {
			Created time.Time `json:"created"`
			Sha     string    `json:"sha"`
			URL     string    `json:"url"`
		} `json:"tree"`
		URL          string `json:"url"`
		Verification struct {
			Payload   string `json:"payload"`
			Reason    string `json:"reason"`
			Signature string `json:"signature"`
			Signer    struct {
				Email    string `json:"email"`
				Name     string `json:"name"`
				Username string `json:"username"`
			} `json:"signer"`
			Verified bool `json:"verified"`
		} `json:"verification"`
	} `json:"commit"`
	Committer struct {
		Active            bool      `json:"active"`
		AvatarURL         string    `json:"avatar_url"`
		Created           time.Time `json:"created"`
		Description       string    `json:"description"`
		Email             string    `json:"email"`
		FollowersCount    int       `json:"followers_count"`
		FollowingCount    int       `json:"following_count"`
		FullName          string    `json:"full_name"`
		ID                int       `json:"id"`
		IsAdmin           bool      `json:"is_admin"`
		Language          string    `json:"language"`
		LastLogin         time.Time `json:"last_login"`
		Location          string    `json:"location"`
		Login             string    `json:"login"`
		ProhibitLogin     bool      `json:"prohibit_login"`
		Restricted        bool      `json:"restricted"`
		StarredReposCount int       `json:"starred_repos_count"`
		Visibility        string    `json:"visibility"`
		Website           string    `json:"website"`
	} `json:"committer"`
	Created time.Time `json:"created"`
	Files   []struct {
		Filename string `json:"filename"`
	} `json:"files"`
	HTMLURL string `json:"html_url"`
	Parents []struct {
		Created time.Time `json:"created"`
		Sha     string    `json:"sha"`
		URL     string    `json:"url"`
	} `json:"parents"`
	Sha   string `json:"sha"`
	Stats struct {
		Additions int `json:"additions"`
		Deletions int `json:"deletions"`
		Total     int `json:"total"`
	} `json:"stats"`
	URL string `json:"url"`
}

// GetRepoLatestCommit 获取仓库最新提交记录
func GetRepoLatestCommit() (*RepoCommit, error) {
	var commits RepoCommits

	api := "https://gitter.top/api/v1/repos/mder/mder/commits"

	if err := gout.GET(api).SetHeader(gout.H{
		"Accept":     "application/json",
		"User-Agent": "mder/beta",
	}).SetQuery(gout.H{
		"page":  1,
		"limit": 5,
	}).SetTimeout(time.Second * 3).BindJSON(&commits).Do(); err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("empty commit")
	}
	return commits[0], nil
}
