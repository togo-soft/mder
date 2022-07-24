package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
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
	_, err := exec.Command("git", "clone", "https://gitter.top/mder/template", base).Output()
	if err != nil {
		return err
	}
	return nil
}

func datetimeStringToTime(datetime string) (time.Time, error) {
	if datetime == "" {
		return time.Now(), nil
	}
	var tpl = "2006-01-02 15:04:05"
	t, err := time.ParseInLocation(tpl, datetime, time.Local)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

func mustString(i interface{}) string {
	s, ok := i.(string)
	if !ok {
		return ""
	}
	return s
}

func mustStringSlice(i interface{}) []string {
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

func int2String(i int) string {
	return strconv.FormatInt(int64(i), 10)
}

func slash(str string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(str, "/", "\\")
	}
	return str
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
	s := sha256.New()
	s.Write([]byte(time.Now().String()))
	res := s.Sum(nil)
	return hex.EncodeToString(res)[:length]
}

func isDraft(tags []string) bool {
	for _, tag := range tags {
		if strings.ToLower(tag) == "draft" {
			return true
		}
	}
	return false
}
