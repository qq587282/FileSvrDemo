package handler

import (
	"FileSvrDemo/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 参数
const (
	FileSavePath string = "./data"
	MaxFileSize  int64  = 128 << 20
)

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

const searchHtml string = `
<!DOCTYPE html>
<head>
	<meta charset="UTF-8">
	<title>查找文件</title>
</head>
<body>
<form action="/v1/search" method="post" enctype="multipart/form-data">
	<input type="text" name="file"/>
	<input type="submit" value="查找"/>
</form>
</body>
</html>`

const delHtml string = `
<!DOCTYPE html>
<head>
	<meta charset="UTF-8">
	<title>删除文件</title>
</head>
<body>
<form action="/v1/del" method="post" enctype="multipart/form-data">
	<input type="text" name="file"/>
	<input type="submit" value="删除"/>
</form>
</body>
</html>`

// 记录文件数据 查找的时候使用
type FileData struct {
	Name      string //文件名
	Size      int64  //大小
	Sha1      string // 哈希
	Location  string //位置
	TimeStamp int64  //时间
}

// 保存文件信息映射
var fileDatas map[string]FileData

// 初始化
func Init() {
	fileDatas = make(map[string]FileData)
}

// UpdateFileData 更新文件信息
func UpdateFileData(fileData FileData) {
	fileDatas[fileData.Sha1] = fileData
}

// GetFileData 获取文件信息对象
func GetFileData(Sha1 string) FileData {
	return fileDatas[Sha1]
}

// RemoveFileData 删除文件信息
func RemoveFileData(Sha1 string) {
	delete(fileDatas, Sha1)
}

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

// 查找页面
func HandlerSearchPage(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte(searchHtml))
	}
}

// 删除页面
func HandlerDelPage(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPost {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte(delHtml))
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
		retinfo      string
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

	buffer := make([]byte, 1024)
	if _, err = io.CopyBuffer(saveFile, uploadFile, buffer); err != nil {
		fmt.Printf("写入文件失败")
		return
	}

	defer saveFile.Close()
	defer uploadFile.Close()

	//保存数据查找
	filedata.Name = uploadHeader.Filename      //文件名
	filedata.Size = request.ContentLength      //大小
	filedata.TimeStamp = time.Now().Unix()     //时间
	filedata.Sha1 = utils.GetFileMD5(saveFile) // 哈希
	filedata.Location = savePath               //位置
	UpdateFileData(filedata)
	retinfo += filedata.Sha1

	// 返回结果
	response := Response{"上传成功!\n" + retinfo, 200}
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

// 查找文件
func HandlerSearch(writer http.ResponseWriter, request *http.Request) {
	var (
		err        error
		fileName   string
		retinfo    string
		buffer     bytes.Buffer
		targetFile *os.File
		fileSha1   string // 哈希
	)

	fileName = request.FormValue("file")
	if len(fileName) == 0 {
		fmt.Printf("查找文件名为空")
		return
	}

	files, err1 := filepath.Glob(FileSavePath + "/" + fileName + ".*")
	if err1 != nil {
		fmt.Println(err1)
	}

	if len(files) > 0 {
		for _, file := range files {
			buffer.WriteString(file)
			fmt.Println(file)
			retinfo = buffer.String()
		}

		if targetFile, err = os.Open(buffer.String()); err != nil {
			fmt.Printf("文件异常")
			return
		}
		fileSha1 = utils.GetFileMD5(targetFile) // 哈希
	} else {
		retinfo = "没有找到"
	}

	defer targetFile.Close()
	// 返回结果
	response := Response{retinfo + " " + fileSha1, 200}
	// 使用 json 包将结构体序列化
	jsonResponse, _ := json.Marshal(response)
	// 设置Content-Type，强制浏览器显示内容为JSON格式
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// 发送响应给客户端
	writer.Write(jsonResponse)
}

// 删除文件
func HandlerdDel(writer http.ResponseWriter, request *http.Request) {
	var (
		err      error
		fileName string
		retinfo  string
		delFile  *os.File
		fileSha1 string // 哈希
	)

	fileName = request.FormValue("file")
	if len(fileName) == 0 {
		fmt.Printf("删除文件名为空")
		return
	}

	fileName = FileSavePath + "/" + fileName
	if delFile, err = os.Open(fileName); err != nil {
		fmt.Printf("文件异常")
		return
	}

	if _, err = os.Stat(fileName); err != nil {
		fmt.Printf("文件异常")
		return
	}

	fileSha1 = utils.GetFileMD5(delFile) // 哈希
	defer delFile.Close()

	RemoveFileData(fileSha1)
	if err = os.Remove(fileName); err != nil {
		fmt.Printf("文件删除失败")
		return
	}

	// 返回结果
	response := Response{retinfo, 200}
	// 使用 json 包将结构体序列化
	jsonResponse, _ := json.Marshal(response)
	// 设置Content-Type，强制浏览器显示内容为JSON格式
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	// 发送响应给客户端
	writer.Write(jsonResponse)
}

func SearchFiles() error {
	fileSystem := os.DirFS(FileSavePath)
	if err2 := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err3 error) error {
		if err3 != nil {
			// 错误处理
			return err3
		}
		if strings.Contains(path, "*.*") {
			// 处理文件
			fmt.Println(path)
		}
		return nil
	}); err2 != nil {
		fmt.Println(err2)
	}
	return nil
}

func WalkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		// 错误处理
		return err
	}
	if !info.IsDir() && strings.Contains(path, ".*") {
		// 处理文件
		fmt.Println(path)
	}
	return nil
}

func ShowFile() {
	fileList, err := utils.ListFiles(FileSavePath)
	if err == nil {
		for i := 0; i < len(fileList); i++ {
			savePath := fileList[i]
			fmt.Printf(savePath + "\n")
		}
	}
}

// 读取配置文件
func LoadCfg() {
	// 解析，读取配置文件内容

	return
}
