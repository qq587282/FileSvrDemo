package main

import (
	"fmt"
	"net/http"
	"os"
	"mime/multipart"
	"io"
	"time"
	"path/filepath"
)

// 参数
const (
	FileSavePath   string = "./data"
	MaxFileSize    int64  = 128 << 20
)

//记录文件数据 查找的时候使用
type FileData struct {
	Name      string  
	Size      int64  
	TimeStamp int64 
}

// 上传页面
func HandlerUploadPage(writer http.ResponseWriter, request *http.Request) {
	var UploadHtml string = `
	<!DOCTYPE html>
	<head>
		<meta charset="UTF-8">
		<title>上传文件</title>
	</head>
	<body>
	<form action="/v1/upload" method="post" enctype="multipart/form-data">
		<input type="file" name="file"/>
		<input type="submit" value="上传"/>
	</form>
	</body>
	</html>`
	
	if request.Method == http.MethodGet {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte(UploadHtml))
	}
}

// 上传文件
func HandlerUpload(writer http.ResponseWriter, request *http.Request) {
	var (
		err          error
		filedata     FileData
		uploadFile   multipart.File
		uploadHeader *multipart.FileHeader
		savePath     string 
		saveFile     *os.File
	)

	if request.Method != "POST" {
		fmt.Printf("不是提交")
		return
	}

	if err = request.ParseMultipartForm(32 << 20); err != nil {
		fmt.Printf("上传文件失败")
		return
	}
	//nFrom := request.MultipartForm
	if request.ContentLength > MaxFileSize {
		fmt.Printf("文件过大")
		return
	}

	if uploadFile, uploadHeader, err = request.FormFile("file"); err != nil {
		fmt.Printf("上传文件异常错误")
		return
	}
	defer uploadFile.Close()
	fmt.Printf("上传文件：文件名[%s],大小[%d],上传时间[%d] \n",uploadHeader.Filename,uploadHeader.Size,time.Now().Unix())
	savePath = FileSavePath +"/" + uploadHeader.Filename
	fmt.Printf("上传文件：文件名[%s]",savePath)
	//保存文件到本地
	if saveFile, err = os.Create(savePath); err != nil {
		fmt.Printf("保存文件失败")
		return
	}
	defer saveFile.Close()	

	buffer := make([]byte, 1024)
	if _, err = io.CopyBuffer(saveFile, uploadFile, buffer); err != nil {
		fmt.Printf("写入文件失败")
		return
	}

	//to do 保持数据查找
	filedata.Name = uploadHeader.Filename
	filedata.Size = request.ContentLength
	filedata.TimeStamp = time.Now().Unix()

}

// 下载文件
func HandlerDownload(writer http.ResponseWriter, request *http.Request) {
}

func main() {
	absPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	absPath = absPath + FileSavePath;
	http.Handle("/", http.FileServer(http.Dir(absPath)))
	http.HandleFunc("/upload.html", HandlerUploadPage)
	http.HandleFunc("/v1/upload", HandlerUpload)
	http.HandleFunc("/v1/download", HandlerDownload)

	err := http.ListenAndServe(":8088", nil)
	if err != nil  {
		fmt.Println(err)
	}

}