package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Status int

const (
	Downloading      Status = 0
	DownloadComplete Status = 1
	DownloadFail     Status = 2
	Installing       Status = 3
	InstallComplete  Status = 4
	InstallFail      Status = 5
)

type Manage struct {
	ID         string `json:"id"`
	FileName   string `json:"file_name"`
	Status     Status `json:"status"`
	Msg        string `json:"msg"`
	Downloader *Downloader
}

type Downloader struct {
	io.Reader       `json:"_"`
	Total           int64                `json:"total"`
	Current         int64                `json:"current"`
	DownloadSuccess func(manage *Manage) `json:"_"`
}

var DownloadManage []*Manage

func (d *Downloader) Read(p []byte) (n int, err error) {
	n, err = d.Reader.Read(p)
	d.Current += int64(n)
	//log.Printf("\r正在下载，下载进度：%.2f%%", float64(d.Current*10000/d.Total)/100)
	//if d.Current == d.Total {
	//	log.Printf("\r下载完成，下载进度：%.2f%%", float64(d.Current*10000/d.Total)/100)
	//}
	return
}
func DownloadFile(url, filePath string, timing int) (*Manage, error) {
	manage := &Manage{
		ID:         strconv.Itoa(int(time.Now().UnixNano())),
		FileName:   "",
		Status:     Downloading,
		Msg:        "",
		Downloader: nil,
	}
	resp, err := http.Get(url)
	if err != nil {
		manage.Status = DownloadFail
		manage.Msg = "请求文件下载出错"
		log.Printf("请求文件下载出错: %v", err.Error())
		return nil, err
	}

	downloader := &Downloader{
		Reader: resp.Body,
		Total:  resp.ContentLength,
	}
	manage.Downloader = downloader

	DownloadManage = append(DownloadManage, manage)

	var filename string
	Disposition := resp.Header.Get("Content-Disposition")
	r := regexp.MustCompile(`filename=(.+)`)
	if r.MatchString(Disposition) {
		filename = r.FindStringSubmatch(Disposition)[1]
	} else {
		filename = manage.ID

		if path.Base(url) != "" {
			filename = path.Base(url)
			filename = strings.ReplaceAll(filename, "?", "_")
			filename = strings.ReplaceAll(filename, "&", "_")
		}
	}
	manage.FileName = filename
	fmt.Println("  Disposition :", Disposition)
	fmt.Println("  fileName :", manage.FileName)
	go func(manage *Manage) {
		os.Remove(path.Join(filePath, filename))

		file, _ := os.Create(path.Join(filePath, filename))

		defer func() {
			_ = resp.Body.Close()
			_ = file.Close()
			if downloader.DownloadSuccess != nil {
				downloader.DownloadSuccess(manage)
			}

			go func() {
				time.Sleep(time.Minute * time.Duration(timing))
				err := os.Remove(path.Join(filePath, filename))
				if err != nil {
					log.Printf("%v 删除文件出错: %v", filename, err.Error())
				}
			}()
		}()

		if _, err := io.Copy(file, downloader); err != nil {
			log.Printf("下载文件出错: %v", err.Error())
		}

		manage.Status = DownloadComplete

	}(manage)

	return manage, nil
}

func download(c *gin.Context) {
	url, _ := c.GetQuery("url")
	t, _ := c.GetQuery("t")
	timing, _ := strconv.Atoi(t)
	manage, _ := DownloadFile(url, FileDir, timing)
	//if err != nil {
	//	c.JSON(200, gin.H{ "ID": "" })
	//}
	//c.JSON(200, gin.H{ "ID": manage.ID })
	c.JSON(200, manage)
}
func progress(c *gin.Context) {
	id, _ := c.GetQuery("id")

	for i := 0; i < len(DownloadManage); i++ {
		if DownloadManage[i].ID == id {
			c.JSON(200, DownloadManage[i])
			return
		}
	}

	c.JSON(200, nil)
}

func upload(c *gin.Context) {
	file, _ := c.FormFile("file")
	c.SaveUploadedFile(file, path.Join(UploadDir, file.Filename))
	c.String(200, fmt.Sprintf("%s/files/upload/%s",c.Request.Host, file.Filename))
}

const FileDir = "./static/files"
const UploadDir = FileDir + "/upload"

func main() {

	_, err := os.Stat(FileDir)
	if err != nil {
		_ = os.MkdirAll(FileDir, os.ModePerm)
	}
	_ = os.MkdirAll(UploadDir, os.ModePerm)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {

		router := gin.Default()
		router.GET("/", func(c *gin.Context) {
			c.File("./static/index.html")
		})

		rApi := router.Group("/api")
		rApi.GET("/download", download)
		rApi.GET("/progress", progress)
		rApi.POST("/upload", upload)

		router.StaticFS("/files", http.Dir("./static/files"))

		router.Run(":1280")
	}()

	wg.Wait()
}
