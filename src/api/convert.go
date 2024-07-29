package api

import (
	"XBS/xbstools"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	task map[string]string = make(map[string]string, 1000)
	look                   = sync.Mutex{}
)

const (
	cacheDir string = "middlefile"
)

type BindFile struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
	FileName string                `form:"filename"` // 文件名
}

// func ConvertFile(c *gin.Context) {
// 	var bindFile BindFile
// 	if err := c.ShouldBind(&bindFile); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"message": "upload failed", "success": false})
// 		return
// 	}
// 	file := bindFile.File

// 	src, err := file.Open()
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"message": "upload failed", "success": false})
// 		return
// 	}
// 	defer src.Close()

// 	buffer, err := io.ReadAll(src)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"message": "upload failed", "success": false})
// 		return
// 	}
// 	flag := uuid.New().String()
// 	task[flag] = filepath.Join(cacheDir, flag)
// 	go func(flag string, buffer []byte) {
// 		var out []byte
// 		var err error
// 		outpath := task[flag]
// 		if json.Valid(buffer) {
// 			out, err = xbstools.Json2XBS(buffer)
// 			look.Lock()
// 			task[flag] = "xbs"
// 			look.Unlock()
// 		} else {
// 			out, err = xbstools.XBS2Json(buffer)
// 			look.Lock()
// 			task[flag] = "json"
// 			look.Unlock()
// 		}
// 		if err != nil {
// 			look.Lock()
// 			delete(task, flag)
// 			look.Unlock()
// 			return
// 		}
// 		err = os.WriteFile(outpath, out, os.ModePerm)
// 		if err != nil {
// 			look.Lock()
// 			delete(task, flag)
// 			look.Unlock()
// 			return
// 		}
// 	}(flag, buffer)
// 	c.JSON(http.StatusOK, gin.H{"message": "upload success", "success": true, "flag": flag})
// }
func ConvertFile(c *gin.Context) {
    var bindFile BindFile
    if err := c.ShouldBind(&bindFile); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "upload failed", "success": false})
        return
    }
    file := bindFile.File
    originalFileName := bindFile.FileName // 获取原始文件名

    src, err := file.Open()
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "upload failed", "success": false})
        return
    }
    defer src.Close()

    buffer, err := io.ReadAll(src)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"message": "upload failed", "success": false})
        return
    }

    flag := uuid.New().String() // 生成唯一标识符
    task[flag] = filepath.Join(cacheDir, originalFileName) // 使用原始文件名作为转换后的文件名

    go func(flag string, buffer []byte) {
        var out []byte
        var err error
        outpath := task[flag]
        if json.Valid(buffer) {
            out, err = xbstools.Json2XBS(buffer)
        } else {
            out, err = xbstools.XBS2Json(buffer)
        }
        
        look.Lock()
        if err != nil {
            delete(task, flag)
            look.Unlock()
            return
        }
        task[flag] = "done" // 标记任务完成
        look.Unlock()

        err = os.WriteFile(outpath, out, os.ModePerm)
        if err != nil {
            look.Lock()
            delete(task, flag)
            look.Unlock()
        }
    }(flag, buffer)

    c.JSON(http.StatusOK, gin.H{"message": "upload success", "success": true, "flag": flag})
}


func ConversionStatus(c *gin.Context) {
	flag := c.PostForm("flag")
	if v, ok := task[flag]; ok {
		if v == "xbs" || v == "json" {
			c.JSON(http.StatusOK, gin.H{"message": "convert success", "status": "done"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "converting", "status": "converting"})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "convert failed", "status": false})
	}
}

func DownloadFile(c *gin.Context) {
	flag := c.Param("flag")
	if v, ok := task[flag]; ok {
		path := filepath.Join(cacheDir, flag)
		c.FileAttachment(path, fmt.Sprintf("%s.%s", flag, v))
	} else {
		c.String(http.StatusNotFound, "file not found")
	}
}
