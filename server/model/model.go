// Package model 定义应用程序中使用的所有数据结构
// =============================================
// 作用：
//   定义前后端交互的数据模型，包括：
//   - Todo / TodoHistory: 备忘录待办任务和历史归档
//   - Setting: 用户设置（主题偏好等）
//   - Guestbook: 留言板留言
//   - APIResponse: 统一的 API 响应格式
//
// 使用方式：
//   handler 层使用这些结构体进行 JSON 序列化/反序列化，
//   db 层使用这些结构体与数据库表进行映射。
//
// 设计原则：
//   所有字段使用 json tag 支持 JSON 序列化/反序列化
//   时间字段使用 time.Time 类型配合 parseTime 参数自动解析
// =============================================

package model

import "time"

// ============================================================
// 备忘录相关模型
// ============================================================

// Todo 待办任务结构体
// 对应数据库 todos 表，存储用户当前的待办事项
type Todo struct {
	ID        int        `json:"id"`                  // 任务唯一标识（自增主键）
	UserID    string     `json:"user_id"`             // 用户标识（当前固定为 "default"，预留多用户扩展）
	Category  string     `json:"category"`            // 分类："life"(生活) / "study"(学习) / "important"(重要)
	Text      string     `json:"text"`                // 任务内容文本（最长200字符）
	Done      bool       `json:"done"`               // 是否已完成：true=完成, false=未完成
	SortOrder int        `json:"sort_order"`          // 排序序号（数值越小越靠前）
	CreatedAt time.Time  `json:"created_at"`          // 创建时间
	UpdatedAt time.Time  `json:"updated_at"`          // 最后更新时间
}

// TodoCreateRequest 新增任务的请求体结构
// 前端 POST /api/todos 时提交的 JSON 数据
type TodoCreateRequest struct {
	Category string `json:"category"` // 必填，分类：life/study/important
	Text     string `json:"text"`     // 必填，任务内容
}

// TodoUpdateRequest 更新任务的请求体结构
// 前端 PUT /api/todos/:id 时提交的 JSON 数据
// 字段均为可选，只更新提供的字段
type TodoUpdateRequest struct {
	Text *string `json:"text,omitempty"` // 可选，更新的任务内容
	Done *bool   `json:"done,omitempty"` // 可选，更新的完成状态
}

// TodoHistory 历史归档结构体
// 对应数据库 todo_history 表，存储已归档的过期任务
type TodoHistory struct {
	ID          int       `json:"id"`           // 记录唯一标识（自增主键）
	UserID      string    `json:"user_id"`      // 用户标识
	ArchiveDate string    `json:"archive_date"` // 归档日期（格式 YYYY-MM-DD）
	Category    string    `json:"category"`     // 归档时的分类
	Text        string    `json:"text"`         // 任务内容
	Done        bool      `json:"done"`         // 归档时的完成状态
}

// TodoHistoryByDate 按日期分组的历史归档响应结构
// 前端 GET /api/todo/history?date=2026-04-05 的返回数据
type TodoHistoryByDate struct {
	Date  string             `json:"date"`  // 归档日期
	Todos map[string][]TodoItem `json:"todos"` // 按 category 分组的任务列表
}

// TodoItem 简化的待办项（用于历史归档展示）
type TodoItem struct {
	Text string `json:"text"` // 任务内容
	Done bool   `json:"done"` // 完成状态
}

// ============================================================
// 设置相关模型
// ============================================================

// Setting 用户设置结构体
// 对应数据库 settings 表，存储用户的个性化偏好设置
type Setting struct {
	UserID    string    `json:"user_id"`    // 用户标识（主键）
	Theme     string    `json:"theme"`      // 主题偏好："light"(亮色) / "dark"(暗色)
	UpdatedAt time.Time `json:"updated_at"` // 最后更新时间
}

// ThemeUpdateRequest 主题更新的请求体结构
// 前端 PUT /api/settings/theme 时提交的 JSON 数据
type ThemeUpdateRequest struct {
	Theme string `json:"theme"` // 必填，目标主题："light" 或 "dark"
}

// ============================================================
// 留言板相关模型
// ============================================================

// Guestbook 留言板留言结构体
// 对应数据库 guestbook 表，存储访客的留言
type Guestbook struct {
	ID        int       `json:"id"`         // 留言唯一标识（自增主键）
	Nickname  string    `json:"nickname"`   // 留言者昵称（可选，为空时显示"匿名访客"）
	Content   string    `json:"content"`    // 留言内容（最长500字符）
	CreatedAt time.Time `json:"created_at"` // 发布时间
}

// GuestbookCreateRequest 发布留言的请求体结构
// 前端 POST /api/guestbook 时提交的 JSON 数据
type GuestbookCreateRequest struct {
	Nickname string `json:"nickname"` // 可选，留言者昵称
	Content  string `json:"content"`  // 必填，留言内容
}

// GuestbookListResponse 留言列表的分页响应结构
// 前端 GET /api/guestbook?page=1&size=20 的返回数据
type GuestbookListResponse struct {
	List      []Guestbook `json:"list"`        // 当前页的留言列表
	Total     int64       `json:"total"`       // 留言总数
	Page      int         `json:"page"`        // 当前页码
	Size      int         `json:"size"`        // 每页条数
	TotalPages int        `json:"total_pages"` // 总页数
}

// ============================================================
// API 统一响应模型
// ============================================================

// APIResponse 统一的 API 响应格式
// 所有 API 接口都使用此结构返回数据，便于前端统一处理
// 成功时 code=0，失败时 code>0 并附带错误信息
type APIResponse struct {
	Code    int         `json:"code"`              // 状态码：0=成功, 其他=错误码
	Message string      `json:"message"`           // 响应消息：成功时为 "success"，失败时为错误描述
	Data    interface{} `json:"data,omitempty"`    // 响应数据（可选，查询接口有值）
}

// SuccessResponse 快速创建成功响应的辅助函数
// code=0, message="success", data 为传入的数据
func SuccessResponse(data interface{}) APIResponse {
	return APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// ErrorResponse 快速创建错误响应的辅助函数
// code>0, message 为错误描述, data 为 nil
func ErrorResponse(code int, message string) APIResponse {
	return APIResponse{
		Code:    code,
		Message: message,
	}
}
