package panel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sun-panel/api/api_v1/common/apiData/commonApiStructs"
	"sun-panel/api/api_v1/common/apiData/panelApiStructs"
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

type ItemIcon struct {
}

func (a *ItemIcon) Edit(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	req := models.ItemIcon{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	if req.ItemIconGroupId == 0 {
		// apiReturn.Error(c, "Group is mandatory")
		apiReturn.ErrorParamFomat(c, "Group is mandatory")
		return
	}

	req.UserId = userInfo.ID

	// json转字符串
	if j, err := json.Marshal(req.Icon); err == nil {
		req.IconJson = string(j)
	}

	if req.ID != 0 {
		// 修改
		updateField := []string{"IconJson", "Icon", "Title", "Url", "LanUrl", "Description", "OpenMethod", "GroupId", "UserId", "ItemIconGroupId"}
		if req.Sort != 0 {
			updateField = append(updateField, "Sort")
		}
		global.Db.Model(&models.ItemIcon{}).
			Select(updateField).
			Where("id=?", req.ID).Updates(&req)
	} else {
		req.Sort = 9999
		// 创建
		global.Db.Create(&req)
	}

	apiReturn.SuccessData(c, req)
}

// 添加多个图标
func (a *ItemIcon) AddMultiple(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	// type Request
	req := []models.ItemIcon{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	for i := 0; i < len(req); i++ {
		if req[i].ItemIconGroupId == 0 {
			apiReturn.ErrorParamFomat(c, "Group is mandatory")
			return
		}
		req[i].UserId = userInfo.ID
		// json转字符串
		if j, err := json.Marshal(req[i].Icon); err == nil {
			req[i].IconJson = string(j)
		}
	}

	global.Db.Create(&req)

	apiReturn.SuccessData(c, req)
}

// // 获取详情
// func (a *ItemIcon) GetInfo(c *gin.Context) {
// 	req := systemApiStructs.AiDrawGetInfoReq{}

// 	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
// 		apiReturn.ErrorParamFomat(c, err.Error())
// 		return
// 	}

// 	userInfo, _ := base.GetCurrentUserInfo(c)

// 	aiDraw := models.AiDraw{}
// 	aiDraw.ID = req.ID
// 	if err := aiDraw.GetInfo(global.Db); err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			apiReturn.Error(c, "不存在记录")
// 			return
// 		}
// 		apiReturn.ErrorDatabase(c, err.Error())
// 		return
// 	}

// 	if userInfo.ID != aiDraw.UserID {
// 		apiReturn.ErrorNoAccess(c)
// 		return
// 	}

// 	apiReturn.SuccessData(c, aiDraw)
// }

func (a *ItemIcon) GetListByGroupId(c *gin.Context) {
	req := models.ItemIcon{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)
	itemIcons := []models.ItemIcon{}

	if err := global.Db.Order("sort ,created_at").Find(&itemIcons, "item_icon_group_id = ? AND user_id=?", req.ItemIconGroupId, userInfo.ID).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	for k, v := range itemIcons {
		json.Unmarshal([]byte(v.IconJson), &itemIcons[k].Icon)
	}

	apiReturn.SuccessListData(c, itemIcons, 0)
}

func (a *ItemIcon) Deletes(c *gin.Context) {
	req := commonApiStructs.RequestDeleteIds[uint]{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)
	if err := global.Db.Delete(&models.ItemIcon{}, "id in ? AND user_id=?", req.Ids, userInfo.ID).Error; err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}

	apiReturn.Success(c)
}

// 保存排序
func (a *ItemIcon) SaveSort(c *gin.Context) {
	req := panelApiStructs.ItemIconSaveSortRequest{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}

	userInfo, _ := base.GetCurrentUserInfo(c)

	transactionErr := global.Db.Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		for _, v := range req.SortItems {
			if err := tx.Model(&models.ItemIcon{}).Where("user_id=? AND id=? AND item_icon_group_id=?", userInfo.ID, v.Id, req.ItemIconGroupId).Update("sort", v.Sort).Error; err != nil {
				// 返回任何错误都会回滚事务
				return err
			}
		}

		// 返回 nil 提交事务
		return nil
	})

	if transactionErr != nil {
		apiReturn.ErrorDatabase(c, transactionErr.Error())
		return
	}

	apiReturn.Success(c)
}

// 支持获取并直接下载对方网站图标到服务器，同时获取标题和描述信息
// 现在通过调用内部Hello Favicon服务获取网站信息
func (a *ItemIcon) GetSiteFavicon(c *gin.Context) {
	userInfo, _ := base.GetCurrentUserInfo(c)
	req := panelApiStructs.ItemIconGetSiteFaviconReq{}

	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		apiReturn.ErrorParamFomat(c, err.Error())
		return
	}
	resp := panelApiStructs.ItemIconGetSiteFaviconResp{}
	
	// 首先检查Hello Favicon服务是否可用
	if err := checkHelloFaviconService(); err != nil {
		global.Logger.Debug("Hello Favicon服务不可用: ", err.Error())
		apiReturn.Error(c, "acquisition failed: hello favicon service unavailable: "+err.Error())
		return
	}
	
	parsedURL, err := url.Parse(req.Url)
	if err != nil {
		apiReturn.Error(c, "acquisition failed: parse URL error: "+err.Error())
		return
	}
	
	
	// 调用Hello Favicon POST API获取完整的favicon信息，使用重试机制
	faviconInfo, err := getFaviconFromHelloFaviconWithRetry(req.Url, 2)
	if err != nil {
		global.Logger.Debug("获取favicon信息失败: ", err.Error())
		apiReturn.Error(c, "acquisition failed: get favicon info error: "+err.Error())
		return
	}
	
	global.Logger.Debug("成功获取favicon信息: ", faviconInfo.Title)
	
	// 设置标题和描述
	resp.Title = faviconInfo.Title
	resp.Description = faviconInfo.Description
	
	// 选择最大尺寸的favicon以保证清晰度
	largestFavicon := getLargestFavicon(faviconInfo.Favicons)
	if largestFavicon == "" {
		apiReturn.Error(c, "acquisition failed: no favicon data found")
		return
	}
	
	global.Logger.Debug("选择最大尺寸的favicon")
	
	// 保存base64数据到本地文件
	iconPath, err := saveFaviconToCache(parsedURL.Host, largestFavicon)
	if err != nil {
		global.Logger.Debug("保存favicon失败: ", err.Error())
		apiReturn.Error(c, "acquisition failed: save favicon error: "+err.Error())
		return
	}
	
	// 保存到数据库
	mFile := models.File{}
	if _, err := mFile.AddFile(userInfo.ID, parsedURL.Host, ".png", iconPath); err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	
	// 去除./conf前缀，只保留/uploads/年/月/日/文件名
	resp.IconUrl = strings.Replace(iconPath, "./conf", "", 1)
	global.Logger.Debug("favicon处理完成，保存路径: ", resp.IconUrl)
	apiReturn.SuccessData(c, resp)
}

// HelloFaviconFullResponse Hello Favicon POST API的完整响应结构
type HelloFaviconFullResponse struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	URL         string            `json:"url"`
	FaviconURL  string            `json:"faviconUrl"`
	Favicons    map[string]string `json:"favicons"` // 尺寸 -> base64编码
}

// 创建配置好的HTTP客户端
var helloFaviconClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
		DisableKeepAlives:   false,
	},
}

// checkHelloFaviconService 检查Hello Favicon服务是否可用
func checkHelloFaviconService() error {
	global.Logger.Debug("检查Hello Favicon服务健康状态")
	
	// 创建简单的健康检查请求
	req, err := http.NewRequest("GET", "http://127.0.0.1:3000/", nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %v", err)
	}
	
	// 设置较短的超时时间用于健康检查
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("hello favicon service is not available: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hello favicon service health check failed with status: %d", resp.StatusCode)
	}
	
	global.Logger.Debug("Hello Favicon服务健康检查通过")
	return nil
}



// getFaviconFromHelloFaviconWithRetry 带重试机制的favicon信息获取（使用POST API）
func getFaviconFromHelloFaviconWithRetry(targetURL string, maxRetries int) (*HelloFaviconFullResponse, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			global.Logger.Debug(fmt.Sprintf("重试获取favicon信息，第%d次尝试", attempt+1))
			// 重试前等待一段时间
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		
		resp, err := getFaviconFromHelloFavicon(targetURL)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		global.Logger.Debug(fmt.Sprintf("获取favicon信息失败 (尝试 %d/%d): %v", attempt+1, maxRetries+1, err))
	}
	
	return nil, fmt.Errorf("所有重试都失败了: %v", lastErr)
}

// getFaviconFromHelloFavicon 从Hello Favicon内部服务获取完整的favicon信息（使用POST API）
func getFaviconFromHelloFavicon(targetURL string) (*HelloFaviconFullResponse, error) {
	// 构建POST请求体
	requestBody := map[string]string{
		"url": targetURL,
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}
	
	// 创建HTTP POST请求
	req, err := http.NewRequest("POST", "http://127.0.0.1:3000/api/favicon", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Sun-Panel/1.0")
	
	global.Logger.Debug("调用Hello Favicon POST API: ", targetURL)
	
	// 发送请求
	resp, err := helloFaviconClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call hello favicon service: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("hello favicon service returned status code: %d", resp.StatusCode)
	}
	
	// 解析响应
	var helloFaviconResp HelloFaviconFullResponse
	if err := json.NewDecoder(resp.Body).Decode(&helloFaviconResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	
	global.Logger.Debug("成功获取favicon数据，包含尺寸数量: ", len(helloFaviconResp.Favicons))
	return &helloFaviconResp, nil
}



// saveFaviconToCache 保存favicon的base64数据到本地缓存文件
func saveFaviconToCache(domain, base64Data string) (string, error) {
	// 解析base64数据
	if !strings.HasPrefix(base64Data, "data:image/png;base64,") {
		return "", fmt.Errorf("invalid base64 data format")
	}
	
	// 提取base64编码部分
	base64Str := strings.TrimPrefix(base64Data, "data:image/png;base64,")
	imgData, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %v", err)
	}
	
	// 生成保存目录 - 统一保存到auto-icons文件夹
	configUpload := global.Config.GetValueString("base", "source_path")
	savePath := fmt.Sprintf("%s/auto-icons/", configUpload)
	isExist, _ := cmn.PathExists(savePath)
	if !isExist {
		os.MkdirAll(savePath, os.ModePerm)
	}
	
	// 生成文件名 - 使用domain的MD5值确保同一域名的图标不会重复保存
	fileName := fmt.Sprintf("%s.png", cmn.Md5(domain))
	filePath := savePath + fileName
	
	// 检查文件是否已存在，如果存在则直接返回路径，避免重复保存
	if fileExists, _ := cmn.PathExists(filePath); fileExists {
		global.Logger.Debug("favicon文件已存在，直接使用: ", filePath)
		return filePath, nil
	}
	
	// 写入文件
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()
	
	if _, err := file.Write(imgData); err != nil {
		return "", fmt.Errorf("failed to write file: %v", err)
	}
	
	global.Logger.Debug("favicon缓存保存成功: ", filePath)
	return filePath, nil
}

// getLargestFavicon 从favicon映射中选择最大尺寸的favicon
func getLargestFavicon(favicons map[string]string) string {
	var selectedBase64 string
	maxSize := 0
	
	for size, base64Data := range favicons {
		if sizeInt, err := strconv.Atoi(size); err == nil && sizeInt > maxSize {
			maxSize = sizeInt
			selectedBase64 = base64Data
		}
	}
	
	return selectedBase64
}
