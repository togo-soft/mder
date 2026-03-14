package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func isExist(fp string) bool {
	_, err := os.Stat(fp)
	return !os.IsNotExist(err)
}

func mkdir(fp string) error {
	return os.MkdirAll(fp, os.ModePerm)
}

func cloneTemplate(base string) error {
	_, err := exec.Command("git", "clone", "https://codeberg.org/mder/template", base).Output()
	if err != nil {
		return err
	}
	return nil
}

func datetimeStringToTime(datetime string) (time.Time, error) {
	if datetime == "" {
		return time.Now(), nil
	}
	t, err := time.ParseInLocation(time.DateTime, datetime, time.Local)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

func toString(i interface{}) string {
	s, ok := i.(string)
	if !ok {
		return ""
	}
	return s
}

func toStringSlice(i interface{}) []string {
	var arr []string
	iarr, ok := i.([]interface{})
	if !ok {
		return []string{}
	}
	for _, str := range iarr {
		r, ok := str.(string)
		if !ok {
			continue
		}
		arr = append(arr, r)
	}
	return arr
}

func isCommandExist(name string) bool {
	_, err := exec.LookPath(name)
	if err != nil {
		return false
	}
	return true
}

func goInstall(pkg string) error {
	url := fmt.Sprintf("%s@latest", pkg)
	_, err := exec.Command("go", "install", url).Output()
	return err
}

func uploadToUpyun(auth, dir string) error {
	if strings.HasSuffix(dir, "/") {
		dir = strings.TrimSuffix(dir, "/")
	}
	_, err := exec.Command("upx", "--auth", auth, "rm", "-d", "-a", "/*").Output()
	if err != nil {
		return fmt.Errorf("remove old data failed: %v", err)
	}
	_, err = exec.Command("upx", "--auth", auth, "put", dir+"/dist/.", "/").Output()
	if err != nil {
		return fmt.Errorf("deploy data failed: %v", err)
	}
	return nil
}

func randString(length int) string {
	b := make([]byte, (length+1)/2)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)[:length]
}

func isDraft(tags []string) bool {
	for _, tag := range tags {
		if strings.ToLower(tag) == "draft" {
			return true
		}
	}
	return false
}

func copyDir(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)
		if d.IsDir() {
			return os.MkdirAll(targetPath, os.ModePerm)
		}
		return copyFile(path, targetPath)
	})
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, os.ModePerm)
}
