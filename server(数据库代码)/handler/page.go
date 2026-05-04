// Package handler 处理 HTML 页面管理相关请求
// =============================================
// 作用：
//   提供管理员上传、查看、删除 HTML 页面的接口
//   所有页面存储在 uploads/pages/ 目录下，通过 /pages/:slug 访问
//
// 路由设计：
//   POST   /api/admin/pages    - 上传新页面（需要管理员权限）
//   GET    /api/admin/pages    - 获取页面列表（需要管理员权限）
//   DELETE /api/admin/pages/:id - 删除页面（需要管理员权限）
//   GET    /pages/:slug        - 访问页面（公开访问）
// =============================================

package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tinyweb1/middleware"
	"tinyweb1/model"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// 文件存储路径
const (
	UploadDir     = "uploads/pages"
	MaxUploadSize = 10 << 20 // 10 MB
)

// 初始化：确保上传目录存在
func init() {
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		fmt.Printf("创建上传目录失败: %v\n", err)
	}
}

// CreatePage 创建/上传 HTML 页面
// POST /api/admin/pages
func CreatePage(database *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 限制上传大小
		r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)

		// 解析 multipart 表单
		if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
			sendJSON(w, http.StatusBadRequest, model.APIResponse{
				Code:    400,
				Message: "文件过大或格式错误",
			})
			return
		}

		// 获取表单字段
		title := r.FormValue("title")
		slug := r.FormValue("slug")
		if title == "" || slug == "" {
			sendJSON(w, http.StatusBadRequest, model.APIResponse{
				Code:    400,
				Message: "标题和标识不能为空",
			})
			return
		}

		// 验证 slug 格式（只允许字母、数字、横线）
		slug = strings.ToLower(slug)
		for _, c := range slug {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
				sendJSON(w, http.StatusBadRequest, model.APIResponse{
					Code:    400,
					Message: "标识只能包含小写字母、数字和横线",
				})
				return
			}
		}

		// 获取上传的文件
		file, header, err := r.FormFile("file")
		if err != nil {
			sendJSON(w, http.StatusBadRequest, model.APIResponse{
				Code:    400,
				Message: "请上传文件",
			})
			return
		}
		defer file.Close()

		// 验证文件类型
		if !strings.HasSuffix(strings.ToLower(header.Filename), ".html") {
			sendJSON(w, http.StatusBadRequest, model.APIResponse{
				Code:    400,
				Message: "只支持 .html 文件",
			})
			return
		}

		// 检查 slug 是否已存在
		var existing model.Page
		if err := database.Where("slug = ?", slug).First(&existing).Error; err == nil {
			sendJSON(w, http.StatusConflict, model.APIResponse{
				Code:    409,
				Message: "该标识已被使用",
			})
			return
		}

		// 生成文件名：时间戳_随机数.html
		timestamp := time.Now().Unix()
		filename := fmt.Sprintf("%d_%s.html", timestamp, slug)
		filepath := filepath.Join(UploadDir, filename)

		// 保存文件
		out, err := os.Create(filepath)
		if err != nil {
			sendJSON(w, http.StatusInternalServerError, model.APIResponse{
				Code:    500,
				Message: "保存文件失败",
			})
			return
		}
		defer out.Close()

		size, err := io.Copy(out, file)
		if err != nil {
			os.Remove(filepath)
			sendJSON(w, http.StatusInternalServerError, model.APIResponse{
				Code:    500,
				Message: "写入文件失败",
			})
			return
		}

		// 获取上传者用户名
		username, _ := middleware.GetUsername(r.Context())

		// 保存到数据库
		page := model.Page{
			Title:    title,
			Slug:     slug,
			FileName: filename,
			Size:     size,
			UploadBy: username,
		}

		if err := database.Create(&page).Error; err != nil {
			os.Remove(filepath)
			sendJSON(w, http.StatusInternalServerError, model.APIResponse{
				Code:    500,
				Message: "保存记录失败",
			})
			return
		}

		sendJSON(w, http.StatusOK, model.APIResponse{
			Code:    0,
			Message: "上传成功",
			Data:    toPageResponse(page),
		})
	}
}

// GetPages 获取所有页面列表
// GET /api/admin/pages
func GetPages(database *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var pages []model.Page
		if err := database.Order("created_at DESC").Find(&pages).Error; err != nil {
			sendJSON(w, http.StatusInternalServerError, model.APIResponse{
				Code:    500,
				Message: "获取列表失败",
			})
			return
		}

		// 转换为响应格式
		responses := make([]model.PageResponse, len(pages))
		for i, p := range pages {
			responses[i] = toPageResponse(p)
		}

		sendJSON(w, http.StatusOK, model.APIResponse{
			Code:    0,
			Message: "success",
			Data:    responses,
		})
	}
}

// DeletePage 删除页面
// DELETE /api/admin/pages/:id
func DeletePage(database *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			sendJSON(w, http.StatusBadRequest, model.APIResponse{
				Code:    400,
				Message: "无效的文件ID",
			})
			return
		}

		// 查询页面
		var page model.Page
		if err := database.First(&page, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				sendJSON(w, http.StatusNotFound, model.APIResponse{
					Code:    404,
					Message: "页面不存在",
				})
				return
			}
			sendJSON(w, http.StatusInternalServerError, model.APIResponse{
				Code:    500,
				Message: "查询失败",
			})
			return
		}

		// 删除文件
		filepath := filepath.Join(UploadDir, page.FileName)
		os.Remove(filepath) // 忽略错误，文件可能不存在

		// 删除数据库记录
		if err := database.Delete(&page).Error; err != nil {
			sendJSON(w, http.StatusInternalServerError, model.APIResponse{
				Code:    500,
				Message: "删除记录失败",
			})
			return
		}

		sendJSON(w, http.StatusOK, model.APIResponse{
			Code:    0,
			Message: "删除成功",
		})
	}
}

// ServePage 提供页面访问（公开）
// GET /pages/:slug
func ServePage(database *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		slug := vars["slug"]

		// 查询页面
		var page model.Page
		if err := database.Where("slug = ?", slug).First(&page).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "服务器错误", http.StatusInternalServerError)
			return
		}

		// 读取并返回文件
		filepath := filepath.Join(UploadDir, page.FileName)
		http.ServeFile(w, r, filepath)
	}
}

// toPageResponse 转换为响应格式
func toPageResponse(page model.Page) model.PageResponse {
	return model.PageResponse{
		ID:        page.ID,
		Title:     page.Title,
		Slug:      page.Slug,
		FileName:  page.FileName,
		Size:      page.Size,
		SizeHuman: formatFileSize(page.Size),
		UploadBy:  page.UploadBy,
		CreatedAt: page.CreatedAt,
	}
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
