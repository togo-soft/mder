package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/minio/selfupdate"
	"github.com/spf13/cobra"
)

const (
	repoAPIBase     = "https://codeberg.org/api/v1/repos/mder/mder"
	updateUserAgent = "mder/self-update"
)

func updateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Short:   "mder auto update",
		Example: "mder update",
		Aliases: []string{"u"},
		Run: func(cmd *cobra.Command, args []string) {
			release, err := getRepoLatestRelease()
			if err != nil {
				logger.Error("get latest release failed", "reason", err)
				return
			}

			current := getCurrentVersion()
			if current == release.TagName {
				logger.Info("already latest version", "version", current)
				return
			}

			asset, err := selectReleaseAsset(release.Assets)
			if err != nil {
				logger.Error("select release asset failed", "reason", err, "version", release.TagName)
				return
			}

			logger.Info("updating mder", "from", current, "to", release.TagName, "asset", asset.Name)
			if err := applyBinaryUpdate(asset.BrowserDownloadURL, asset.Name); err != nil {
				logger.Error("update failed", "reason", err, "asset", asset.Name)
				return
			}

			logger.Info("mder update success", "version", release.TagName)
		},
	}

	return cmd
}

type RepoRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func getRepoLatestRelease() (*RepoRelease, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	req, err := http.NewRequest(http.MethodGet, repoAPIBase+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", updateUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var release RepoRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	if release.TagName == "" {
		return nil, fmt.Errorf("empty release tag")
	}
	if len(release.Assets) == 0 {
		return nil, fmt.Errorf("release has no assets")
	}
	return &release, nil
}

func selectReleaseAsset(assets []ReleaseAsset) (*ReleaseAsset, error) {
	targetOS := strings.ToLower(runtime.GOOS)
	targetArch := strings.ToLower(runtime.GOARCH)
	var archiveCandidate *ReleaseAsset

	for i := range assets {
		name := strings.ToLower(assets[i].Name)
		if !strings.Contains(name, targetOS) || !strings.Contains(name, targetArch) {
			continue
		}
		if isArchiveAsset(name) {
			if archiveCandidate == nil {
				archiveCandidate = &assets[i]
			}
			continue
		}
		if targetOS == "windows" && !strings.HasSuffix(name, ".exe") {
			continue
		}
		return &assets[i], nil
	}
	if archiveCandidate != nil {
		return archiveCandidate, nil
	}

	var names []string
	for _, asset := range assets {
		names = append(names, asset.Name)
	}
	return nil, fmt.Errorf("no binary asset matches %s/%s in [%s]", targetOS, targetArch, strings.Join(names, ", "))
}

func applyBinaryUpdate(downloadURL, assetName string) error {
	client := &http.Client{Timeout: 2 * time.Minute}
	req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", updateUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download asset failed: %s", resp.Status)
	}

	switch name := strings.ToLower(assetName); {
	case strings.HasSuffix(name, ".zip"):
		return applyFromZip(resp.Body)
	case strings.HasSuffix(name, ".tar.gz"), strings.HasSuffix(name, ".tgz"):
		return applyFromTarGz(resp.Body)
	default:
		return applyUpdate(resp.Body)
	}
}

func applyUpdate(binary io.Reader) error {
	if err := selfupdate.Apply(binary, selfupdate.Options{}); err != nil {
		if rollbackErr := selfupdate.RollbackError(err); rollbackErr != nil {
			return fmt.Errorf("apply failed: %v, rollback failed: %v", err, rollbackErr)
		}
		return err
	}
	return nil
}

func applyFromZip(archive io.Reader) error {
	data, err := io.ReadAll(archive)
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}

	expected := currentBinaryName()
	for _, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if !strings.EqualFold(path.Base(f.Name), expected) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		return applyUpdate(rc)
	}
	return fmt.Errorf("binary %s not found in zip", expected)
}

func applyFromTarGz(archive io.Reader) error {
	gz, err := gzip.NewReader(archive)
	if err != nil {
		return err
	}
	defer gz.Close()

	expected := currentBinaryName()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeRegA {
			continue
		}
		if !strings.EqualFold(path.Base(h.Name), expected) {
			continue
		}
		return applyUpdate(tr)
	}
	return fmt.Errorf("binary %s not found in tar.gz", expected)
}

func isArchiveAsset(name string) bool {
	return strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz")
}

func currentBinaryName() string {
	if runtime.GOOS == "windows" {
		return "mder.exe"
	}
	return "mder"
}

func getCurrentVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" {
		return "unknown"
	}
	return info.Main.Version
}
