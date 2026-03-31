package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
)

func main() {
	// 获取项目根目录（server/ 的上一级）
	_, currentFile, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(currentFile))

	// 设置静态文件目录为项目根目录
	fs := http.FileServer(http.Dir(rootDir))

	http.Handle("/", fs)

	addr := ":8081"
	fmt.Println("========================================")
	fmt.Println("  TinyWeb1 Server is running!")
	fmt.Printf("  Access: http://localhost%s\n", addr)
	fmt.Printf("  Serving files from: %s\n", rootDir)
	fmt.Println("========================================")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
