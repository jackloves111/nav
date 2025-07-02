package panel

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sun-panel/api/api_v1/common/apiData/commonApiStructs"
	"sun-panel/api/api_v1/common/apiData/panelApiStructs"
	"sun-panel/api/api_v1/common/apiReturn"
	"sun-panel/api/api_v1/common/base"
	"sun-panel/global"
	"sun-panel/lib/cmn"
	"sun-panel/lib/siteFavicon"
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
	fullUrl := ""
	
	// 首先检查Hello Favicon服务是否可用
	if err := checkHelloFaviconService(); err != nil {
		global.Logger.Debug("Hello Favicon服务不可用: ", err.Error())
		apiReturn.Error(c, "acquisition failed: hello favicon service unavailable: "+err.Error())
		return
	}
	
	// 调用内部Hello Favicon服务获取网站信息，使用重试机制
	webInfo, err := getWebsiteInfoFromHelloFaviconWithRetry(req.Url, 2)
	if err != nil {
		global.Logger.Debug("获取网站信息失败: ", err.Error())
		apiReturn.Error(c, "acquisition failed: get website info error: "+err.Error())
		return
	}
	
	global.Logger.Debug("成功获取网站信息: ", webInfo.Title)
	
	// 设置标题和描述
	resp.Title = webInfo.Title
	resp.Description = webInfo.Description
	
	// 如果有图标URL，使用它
	if webInfo.FaviconURL != "" {
		fullUrl = webInfo.FaviconURL
	} else {
		apiReturn.Error(c, "acquisition failed: no favicon found")
		return
	}

	parsedURL, err := url.Parse(req.Url)
	if err != nil {
		apiReturn.Error(c, "acquisition failed:"+err.Error())
		return
	}

	protocol := parsedURL.Scheme
	global.Logger.Debug("protocol:", protocol)
	global.Logger.Debug("fullUrl:", fullUrl)

	// 如果URL以双斜杠（//）开头，则使用当前页面协议
	if strings.HasPrefix(fullUrl, "//") {
		fullUrl = protocol + "://" + fullUrl[2:]
	} else if !strings.HasPrefix(fullUrl, "http://") && !strings.HasPrefix(fullUrl, "https://") {
		// 如果URL既不以http://开头也不以https://开头，则默认为http协议
		fullUrl = "http://" + fullUrl
	}
	global.Logger.Debug("fullUrl:", fullUrl)
	// 去除图标的get参数
	{
		parsedIcoURL, err := url.Parse(fullUrl)
		if err != nil {
			apiReturn.Error(c, "acquisition failed: parsed ico URL :"+err.Error())
			return
		}
		fullUrl = parsedIcoURL.Scheme + "://" + parsedIcoURL.Host + parsedIcoURL.Path
	}
	global.Logger.Debug("fullUrl:", fullUrl)

	// 生成保存目录
	configUpload := global.Config.GetValueString("base", "source_path")
	savePath := fmt.Sprintf("%s/%d/%d/%d/", configUpload, time.Now().Year(), time.Now().Month(), time.Now().Day())
	isExist, _ := cmn.PathExists(savePath)
	if !isExist {
		os.MkdirAll(savePath, os.ModePerm)
	}

	// 下载
	var imgInfo *os.File
	{
		var err error
		if imgInfo, err = siteFavicon.DownloadImage(fullUrl, savePath, 1024*1024); err != nil {
			apiReturn.Error(c, "acquisition failed: download"+err.Error())
			return
		}
	}

	// 保存到数据库
	ext := path.Ext(fullUrl)
	mFile := models.File{}
	if _, err := mFile.AddFile(userInfo.ID, parsedURL.Host, ext, imgInfo.Name()); err != nil {
		apiReturn.ErrorDatabase(c, err.Error())
		return
	}
	// 去除./conf前缀，只保留/uploads/年/月/日/文件名
	resp.IconUrl = strings.Replace(imgInfo.Name(), "./conf", "", 1)
	apiReturn.SuccessData(c, resp)
}

// HelloFaviconResponse Hello Favicon服务的响应结构
type HelloFaviconResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	FaviconURL  string `json:"faviconUrl"`
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

// getWebsiteInfoFromHelloFaviconWithRetry 带重试机制的网站信息获取
func getWebsiteInfoFromHelloFaviconWithRetry(targetURL string, maxRetries int) (*HelloFaviconResponse, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			global.Logger.Debug(fmt.Sprintf("重试获取网站信息，第%d次尝试", attempt+1))
			// 重试前等待一段时间
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		
		resp, err := getWebsiteInfoFromHelloFavicon(targetURL)
		if err == nil {
			return resp, nil
		}
		
		lastErr = err
		global.Logger.Debug(fmt.Sprintf("获取网站信息失败 (尝试 %d/%d): %v", attempt+1, maxRetries+1, err))
	}
	
	return nil, fmt.Errorf("所有重试都失败了: %v", lastErr)
}

// getWebsiteInfoFromHelloFavicon 从Hello Favicon内部服务获取网站信息
func getWebsiteInfoFromHelloFavicon(targetURL string) (*HelloFaviconResponse, error) {
	// 构建内部API请求URL
	apiURL := fmt.Sprintf("http://127.0.0.1:3000/api?=%s", url.QueryEscape(targetURL))
	
	// 创建HTTP请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	// 设置User-Agent
	req.Header.Set("User-Agent", "Sun-Panel/1.0")
	
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
	var helloFaviconResp HelloFaviconResponse
	if err := json.NewDecoder(resp.Body).Decode(&helloFaviconResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	
	return &helloFaviconResp, nil
}
