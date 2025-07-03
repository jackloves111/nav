package models

import "gorm.io/gorm"

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
