//go:build packfile
// +build packfile

package packfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:generate go-bindata -o=staticFile.go -pkg=packfile -tags=packfile ../resource/... ../config.yaml

func writeFile(path string, data []byte) {
	if lastSeparator := strings.LastIndex(path, "/"); lastSeparator != -1 {
		dirPath := path[:lastSeparator]
		if _, err := os.Stat(dirPath); err != nil && os.IsNotExist(err) {
			os.MkdirAll(dirPath, os.ModePerm)
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err2 := os.WriteFile(path, data, os.ModePerm); err2 != nil {
			fmt.Printf("Write file failed: %s\n", path)
		}
	} else {
		fmt.Printf("File exist, skip: %s\n", path)
	}
}

func init() {
	for key := range _bindata {
		filePath, _ := filepath.Abs(strings.TrimPrefix(key, "."))
		data, err := Asset(key)
		if err != nil {
			// Asset was not found.
			fmt.Printf("Fail to find: %s\n", filePath)
		} else {
			writeFile(filePath, data)
		}
	}
}
