package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
)

func updateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "mder auto update",
		Example: "mder update",
		Aliases: []string{"u"},
		Run: func(cmd *cobra.Command, args []string) {
			commit, err := GetRepoLatestCommit()
			if err != nil {
				logger.Error("get mder version failed", "reason", err)
				return
			}
			logger.Info(fmt.Sprintf("-----\nmder latest version: %s\nupdating...\n", commit.Sha))
			var url = fmt.Sprintf("codeberg.org/mder/mder@%s", commit.Sha)
			_, err = exec.Command("go", "install", url).CombinedOutput()
			if err != nil {
				logger.Error("update failed", "install", url, "reason", err)
				return
			}
			logger.Info("mder update success", "hash", commit.Sha)
		},
	}

	return cmd
}

type RepoCommit struct {
	Sha string `json:"sha"`
}

// GetRepoLatestCommit 获取仓库最新提交记录
func GetRepoLatestCommit() (*RepoCommit, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("GET", "https://codeberg.org/api/v1/repos/mder/mder/commits?page=1&limit=5", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "mder/beta")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var commits []*RepoCommit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("empty commit")
	}
	return commits[0], nil
}
