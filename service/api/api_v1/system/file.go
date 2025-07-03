package system

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sun-panel/api/api_v1/common/apiData/commonApiStructs"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/cmn"
	"sun-panel/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type FileApi struct{}

func (a *FileApi) UploadImg(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	configUpload := global.Config.GetValueString("base", "source_path")
	f, err := c.FormFile("imgfile")
	if err != nil {
		apiReturn.ErrorByCode(c, 1300)
		return
	} else {
		fileExt := strings.ToLower(path.Ext(f.Filename))
		agreeExts := []string{
			".png",
			".jpg",
			".gif",
			".jpeg",
			".webp",
			".svg",
			".ico",
		}

		if !cmn.InArray(agreeExts, fileExt) {
			apiReturn.ErrorByCode(c, 1301)
			return
		}
		
		// 创建手动上传文件目录
		uploadsDir := fmt.Sprintf("%s/uploads/uploads-icons/", configUpload)
		isExist, _ := cmn.PathExists(uploadsDir)
		if !isExist {
			os.MkdirAll(uploadsDir, os.ModePerm)
		}
		
		// 先保存到临时文件
		tempFileName := cmn.Md5(fmt.Sprintf("%s%s", f.Filename, time.Now().String()))
		tempFilePath := fmt.Sprintf("%stemp_%s%s", uploadsDir, tempFileName, fileExt)
		c.SaveUploadedFile(f, tempFilePath)

		// 使用基于内容MD5的去重方法
		mFile := models.File{}
		file, finalPath, err := mFile.AddOrGetFileByContentMD5(userInfo.ID, f.Filename, fileExt, tempFilePath, uploadsDir)
		if err != nil {
			apiReturn.ErrorDatabase(c, err.Error())
			return
		}
		
		// 去除./conf前缀，只保留/uploads/uploads-icons/文件名
		urlPath := strings.Replace(finalPath, "./conf", "", 1)
		apiReturn.SuccessData(c, gin.H{
			"imageUrl": urlPath,
		})
	}
}

func (a *FileApi) UploadFiles(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	configUpload := global.Config.GetValueString("base", "source_path")

	form, err := c.MultipartForm()
	if err != nil {
		apiReturn.ErrorByCode(c, 1300)
		return
	}
	
	// 创建手动上传文件目录
	uploadsDir := fmt.Sprintf("%s/uploads/uploads-icons/", configUpload)
	isExist, _ := cmn.PathExists(uploadsDir)
	if !isExist {
		os.MkdirAll(uploadsDir, os.ModePerm)
	}
	
	files := form.File["files[]"]
	errFiles := []string{}
	succMap := map[string]string{}
	for _, f := range files {
		fileExt := strings.ToLower(path.Ext(f.Filename))
		
		// 先保存到临时文件
		tempFileName := cmn.Md5(fmt.Sprintf("%s%s", f.Filename, time.Now().String()))
		tempFilePath := fmt.Sprintf("%stemp_%s%s", uploadsDir, tempFileName, fileExt)
		
		if c.SaveUploadedFile(f, tempFilePath) != nil {
			errFiles = append(errFiles, f.Filename)
		} else {
			// 使用基于内容MD5的去重方法
			mFile := models.File{}
			file, finalPath, err := mFile.AddOrGetFileByContentMD5(userInfo.ID, f.Filename, fileExt, tempFilePath, uploadsDir)
			if err != nil {
				errFiles = append(errFiles, f.Filename)
			} else {
				// 去除./conf前缀，只保留/uploads/uploads-icons/文件名
				urlPath := strings.Replace(finalPath, "./conf", "", 1)
				succMap[f.Filename] = urlPath
			}
		}
	}

	apiReturn.SuccessData(c, gin.H{
		"succMap":  succMap,
		"errFiles": errFiles,
	})
}

func (a *FileApi) GetList(c *gin.Context) {
	list := []models.File{}
	userInfo, _ := base.GetCurrentUserInfo(c)
	var count int64
	if err := global.Db.Order("created_at desc").Find(&list, "user_id=?", userInfo.ID).Count(&count).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	data := []map[string]interface{}{}
	for _, v := range list {
		// 去除./conf前缀，只保留/uploads/uploads-icons/文件名
		urlPath := strings.Replace(v.Src, "./conf", "", 1)
		data = append(data, map[string]interface{}{
			"src":        urlPath,
			"fileName":   v.FileName,
			"id":         v.ID,
			"createTime": v.CreatedAt,
			"updateTime": v.UpdatedAt,
			"path":       v.Src,
		})
	}
	apiReturn.SuccessListData(c, data, count)
}

func (a *FileApi) Deletes(c *gin.Context) {
	req := commonApiStructs.RequestDeleteIds[uint]{}
	userInfo, _ := base.GetCurrentUserInfo(c)
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	global.Db.Transaction(func(tx *gorm.DB) error {
		files := []models.File{}

		if err := tx.Order("created_at desc").Find(&files, "user_id=? AND id in ?", userInfo.ID, req.Ids).Error; err != nil {
			return err
		}

		for _, v := range files {
			os.Remove(v.Src)
		}

		if err := tx.Order("created_at desc").Delete(&files, "user_id=? AND id in ?", userInfo.ID, req.Ids).Error; err != nil {
			return err
		}

		return nil
	})

	apiReturn.Success(c)

}
