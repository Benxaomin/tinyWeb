// TinyWeb1 主程序入口文件
// =============================================
// 作用：
//   程序的启动入口，负责以下工作：
//
//   1. 加载配置：从环境变量读取数据库连接信息、端口等配置
//   2. 初始化数据库：建立连接池，自动创建所需的数据表
//   3. 注册路由：绑定 URL 路径到对应的处理函数（handler）
//   4. 配置 CORS：设置跨域访问策略，允许前端页面调用后端 API
//   5. 启动 HTTP 服务器：同时提供静态文件服务（index.html 等）和 REST API 服务
//
// 项目架构说明：
//   - config/    : 配置管理模块（环境变量 → 结构体）
//   - model/     : 数据结构定义（请求/响应格式）
//   - db/        : 数据库连接和表初始化
//   - handler/   : 各业务模块的 API 处理函数
//   - main.go    : 本文件，路由组装和服务启动
//
// 启动方式：
//   cd server && go run main.go
//
// 环境变量配置（可选）：
//   DB_HOST=localhost DB_PASS=your_password go run main.go
// =============================================

package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"tinyweb1/config"
	"tinyweb1/db"
	"tinyweb1/handler"
)

func main() {
	// ============================================================
	// 步骤 1: 加载配置
	// ============================================================
	config.Load() // 从环境变量读取配置，未设置的用默认值
	fmt.Println("📋 配置加载完成")
	fmt.Printf("   数据库: %s:%s/%s\n", config.GetDBHost(), config.GetDBPort(), config.GetDBName())
	fmt.Printf("   端口: %s\n", config.GetServerPort())

	// ============================================================
	// 步骤 2: 初始化数据库连接池并自动建表
	// ============================================================
	db.Initialize()

	// ============================================================
	// 步骤 3: 注册静态文件服务和 API 路由
	// ============================================================

	// 获取项目根目录（server/ 的上一级），用于提供 index.html 等静态文件
	_, currentFile, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(filepath.Dir(currentFile))

	// 静态文件服务：处理 / 开头的请求（优先级低于精确匹配的路由）
	fs := http.FileServer(http.Dir(rootDir))

	// 使用自定义的多路复用器 (mux) 替代默认的 DefaultServeMux
	// 这样可以更灵活地控制路由匹配顺序和中间件
	mux := http.NewServeMux()

	// ---- 备忘录 API 路由 ----
	mux.HandleFunc("GET /api/todos", handler.GetTodos)              // 获取待办列表 ?category=life|study|important
	mux.HandleFunc("POST /api/todos", handler.CreateTodo)            // 新增待办任务
	mux.HandleFunc("PUT /api/todos/", handler.UpdateTodo)            // 更新待办任务 /api/todos/:id
	mux.HandleFunc("DELETE /api/todos/", handler.DeleteTodo)         // 删除待办任务 /api/todos/:id
	mux.HandleFunc("POST /api/todos/archive", handler.ArchiveTodos)  // 归档当天任务

	// ---- 历史归档 API 路由 ----
	mux.HandleFunc("GET /api/todos/history", handler.GetTodoHistoryByDate) // 按日期查询归档 ?date=2026-04-05
	mux.HandleFunc("GET /api/todos/history/dates", handler.GetTodoHistoryDates) // 获取有归档的日期列表

	// ---- 设置 API 路由 ----
	mux.HandleFunc("GET /api/settings/theme", handler.GetTheme)     // 获取主题偏好
	mux.HandleFunc("PUT /api/settings/theme", handler.UpdateTheme)  // 更新主题偏好

	// ---- 留言板 API 路由 ----
	mux.HandleFunc("GET /api/guestbook", handler.GetGuestbookMessages)    // 获取留言列表（分页）
	mux.HandleFunc("POST /api/guestbook", handler.CreateGuestbookMessage) // 发布新留言

	// ---- 静态文件兜底路由 ----
	// 所有未被 API 路由匹配的请求都交给静态文件服务器处理
	mux.Handle("/", fs)

	// ============================================================
	// 步骤 4: 包装 CORS 中间件 + 启动 HTTP 服务
	// ============================================================
	addr := config.GetServerPort()
	fmt.Println("")
	fmt.Println("========================================")
	fmt.Println("  🚀 TinyWeb1 Server is running!")
	fmt.Printf("  📍 访问地址: http://localhost%s\n", addr)
	fmt.Printf("  📂 静态文件: %s\n", rootDir)
	fmt.Println("  🔗 API 接口:")
	fmt.Println("     GET    /api/todos?category=life")
	fmt.Println("     POST   /api/todos")
	fmt.Println("     PUT    /api/todos/:id")
	fmt.Println("     DELETE /api/todos/:id")
	fmt.Println("     POST   /api/todos/archive")
	fmt.Println("     GET    /api/todos/history?date=...")
	fmt.Println("     GET    /api/todos/history/dates")
	fmt.Println("     GET    /api/settings/theme")
	fmt.Println("     PUT    /api/settings/theme")
	fmt.Println("     GET    /api/guestbook")
	fmt.Println("     POST   /api/guestbook")
	fmt.Println("========================================")

	// 使用 corsMiddleware 包装 mux，添加跨域支持
	if err := http.ListenAndServe(addr, corsMiddleware(mux)); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}

// corsMiddleware CORS 跨域中间件
// =============================================
// 作用：
//   在每个 HTTP 响应中添加 CORS 相关的头部，
//   允许前端 JavaScript 从不同域名/端口调用此 API。
//
// 为什么需要 CORS？
//   前端页面可能部署在 example.com:80，
//   后端 API 运行在 api.example.com:8081，
//   浏览器的同源策略会阻止这种跨域请求。
//   通过添加 Access-Control-Allow-* 头部来允许合法的跨域访问。
//
// 安全注意事项：
//   生产环境应将 ALLOWED_ORIGINS 设为具体的域名（如 https://yourdomain.com），
//   仅开发环境使用 "*" 允许所有来源。
// =============================================
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigins := config.GetAllowedOrigins()

		// 检查请求来源是否在允许列表中
		allowOrigin := ""
		if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
			// 开发模式：允许所有来源
			allowOrigin = "*"
		} else {
			// 生产模式：逐一比对允许的来源
			for _, allowed := range allowedOrigins {
				if origin == allowed {
					allowOrigin = origin
					break
				}
			}
		}

		// 设置 CORS 响应头
		if allowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)           // 允许的来源
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // 允许的方法
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")      // 允许的请求头
			w.Header().Set("Access-Control-Max-Age", "86400") // 预检请求缓存时间（24小时）
		}

		// 处理预检请求 (OPTIONS)：浏览器在非简单请求前会先发送 OPTIONS 探测
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent) // 204 No Content
			return
		}

		// 继续处理实际请求
		next.ServeHTTP(w, r)
	})
}
