package main

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// 参数
const (
	FileSavePath string = "./data"
	MaxFileSize  int64  = 128 << 20
)

// 记录文件数据 查找的时候使用
type FileData struct {
	Name      string
	Size      int64
	TimeStamp int64
}

// 返回消息
type Response struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

const UploadHtml string = `
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

const DownloadHtml string = `
<!DOCTYPE html>
<head>
	<meta charset="UTF-8">
	<title>下载文件</title>
</head>
<body>
<form action="/v1/download" method="post" enctype="multipart/form-data">
	<input type="text" name="file"/>
	<input type="submit" value="下载"/>
</form>
</body>
</html>`

// 上传页面
func HandlerUploadPage(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte(UploadHtml))
	}
}

// 下载页面
func HandlerDownloadPage(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte(DownloadHtml))
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
	fmt.Printf("上传文件：文件名[%s],大小[%d],上传时间[%d] \n", uploadHeader.Filename, uploadHeader.Size, time.Now().Unix())
	savePath = FileSavePath + "/" + uploadHeader.Filename
	fmt.Printf("上传文件：文件名[%s]", savePath)

	//文件是否存在
	if _, err = os.Stat(savePath); err == nil {
		fmt.Printf("文件名重复")
		return
	}
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

	// 返回结果
	response := Response{"上传成功!", 200}
	// 使用 json 包将结构体序列化
	jsonResponse, _ := json.Marshal(response)
	// 设置Content-Type，强制浏览器显示内容为JSON格式
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// 发送响应给客户端
	writer.Write(jsonResponse)
}

// 下载文件
func HandlerDownload(writer http.ResponseWriter, request *http.Request) {
	var (
		err          error
		fileName     string
		downloadFile *os.File
		fileinfo     os.FileInfo
	)

	fileName = request.FormValue("file")
	if len(fileName) == 0 {
		fmt.Printf("文件名不存在")
		return
	}

	fileName = FileSavePath + "/" + fileName
	if downloadFile, err = os.Open(fileName); err != nil {
		fmt.Printf("文件异常")
		return
	}

	if fileinfo, err = os.Stat(fileName); err != nil {
		fmt.Printf("文件异常")
		return
	}

	defer downloadFile.Close()
	if downloadFile == nil {
		fmt.Printf("文件不存在")
		return
	}

	writer.Header().Set("Content-Type", "application/octet-stream")
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", fileinfo.Size))

	http.ServeContent(writer, request, fileName, fileinfo.ModTime(), downloadFile)
	// buffer := make([]byte, 1024)
	// if _, err = io.CopyBuffer(writer, downloadFile, buffer); err != nil {
	// 	fmt.Printf("下载文件失败")
	// 	return
	// }
}

func main() {
	absPath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	absPath = absPath + FileSavePath
	http.Handle("/", http.FileServer(http.Dir(absPath)))
	http.HandleFunc("/upload.html", HandlerUploadPage)
	http.HandleFunc("/v1/upload", HandlerUpload)
	http.HandleFunc("/download.html", HandlerDownloadPage)
	http.HandleFunc("/v1/download", HandlerDownload)

	err := http.ListenAndServe(":8088", nil)
	if err != nil {
		fmt.Println(err)
	}

}
