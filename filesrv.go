package main

import (
	"FileSvrDemo/handler"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	absPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	absPath = absPath + handler.FileSavePath
	filepath.Walk(handler.FileSavePath, handler.WalkFunc)
	http.Handle("/", http.FileServer(http.Dir(absPath)))
	http.HandleFunc("/upload.html", handler.HandlerUploadPage)
	http.HandleFunc("/v1/upload", handler.HandlerUpload)
	http.HandleFunc("/download.html", handler.HandlerDownloadPage)
	http.HandleFunc("/v1/download", handler.HandlerDownload)
	http.HandleFunc("/search.html", handler.HandlerSearchPage)
	http.HandleFunc("/v1/search", handler.HandlerSearch)

	err := http.ListenAndServe(":8088", nil)
	if err != nil {
		fmt.Println(err)
	}

}
