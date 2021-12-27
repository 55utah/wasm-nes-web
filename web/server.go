package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

func main() {
	// http://localhost:8003/web/src/
	addr := "127.0.0.1:8003"
	dir := "."
	http.ListenAndServe(addr, http.FileServer(http.Dir(dir)))
}

// 构建出所有游戏名称的json文件，提供给web端使用
func BuildJson() {
	filePath, _ := filepath.Abs("./web/src/nes-roms")
	targetPath, _ := filepath.Abs("./web/src/nes-roms.json")
	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		panic(err)
	}
	list := []map[string]interface{}{}
	for _, file := range files {
		name := file.Name()
		if file.IsDir() || path.Ext(name) != ".nes" {
			continue
		}
		target := map[string]interface{}{"name": name}
		list = append(list, target)
	}
	bytes, err := json.Marshal(list)
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(targetPath, bytes, os.FileMode(0444))
}
