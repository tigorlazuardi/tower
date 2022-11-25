package loader

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func LoadEnv() {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".env" {
			setEnv(path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func setEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		if text[0] == '#' {
			continue
		}
		if index := strings.Index(text, "#"); index != -1 {
			text = text[:index]
		}
		parts := strings.SplitN(text, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		_ = os.Setenv(key, value)
	}
}
