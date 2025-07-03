package models

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"gorm.io/gorm"
)

type File struct {
	BaseModel
	Src      string `json:"src"`
	UserId   uint   `json:"userId"`
	FileName string `json:"fileName" gorm:"varchar(255)"` // 文件名
	Method   int    `gorm:"int(5)" json:"method"`         // 上传方式
	Ext      string `gorm:"varchar(255)" json:"ext"`      // 扩展名
}

// 检查文件是否已存在
func (m *File) FileExists(userId uint, src string) (bool, File, error) {
	var file File
	err := Db.Where("user_id = ? AND src = ?", userId, src).First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, file, nil
		}
		return false, file, err
	}
	return true, file, nil
}

// 计算文件内容的MD5值
func (m *File) GetFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// 基于文件内容MD5检查文件是否已存在
func (m *File) FileExistsByContentMD5(userId uint, fileMD5 string) (bool, File, error) {
	var file File
	err := Db.Where("user_id = ? AND src LIKE ?", userId, "%/uploads/uploads-icons/"+fileMD5+".%").First(&file).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, file, nil
		}
		return false, file, err
	}
	return true, file, nil
}

// 添加一个文件记录
func (m *File) AddFile(userId uint, fileName, ext, src string) (File, error) {
	file := File{
		UserId:   userId,
		FileName: fileName,
		Src:      src,
		Ext:      ext,
	}
	err := Db.Create(&file).Error

	return file, err
}

// 添加或获取已存在的文件记录
func (m *File) AddOrGetFile(userId uint, fileName, ext, src string) (File, error) {
	// 先检查文件是否已存在
	exists, existingFile, err := m.FileExists(userId, src)
	if err != nil {
		return File{}, err
	}
	
	if exists {
		// 文件已存在，直接返回现有记录
		return existingFile, nil
	}
	
	// 文件不存在，创建新记录
	return m.AddFile(userId, fileName, ext, src)
}

// 基于文件内容MD5添加或获取已存在的文件记录（用于手动上传去重）
func (m *File) AddOrGetFileByContentMD5(userId uint, fileName, ext, tempFilePath, finalDir string) (File, string, error) {
	// 计算文件内容的MD5
	fileMD5, err := m.GetFileMD5(tempFilePath)
	if err != nil {
		return File{}, "", err
	}
	
	// 检查是否已存在相同内容的文件
	exists, existingFile, err := m.FileExistsByContentMD5(userId, fileMD5)
	if err != nil {
		return File{}, "", err
	}
	
	if exists {
		// 文件已存在，删除临时文件，返回现有记录
		os.Remove(tempFilePath)
		return existingFile, existingFile.Src, nil
	}
	
	// 文件不存在，移动到最终位置并创建新记录
	finalPath := fmt.Sprintf("%s%s%s", finalDir, fileMD5, ext)
	if err := os.Rename(tempFilePath, finalPath); err != nil {
		return File{}, "", err
	}
	
	file, err := m.AddFile(userId, fileName, ext, finalPath)
	return file, finalPath, err
}
