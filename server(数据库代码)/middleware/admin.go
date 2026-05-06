// Package middleware 提供管理员权限中间件
// =============================================
// 作用：
//   AdminMiddleware 是管理员权限中间件，用于保护只有管理员才能访问的 API 接口。
//   工作流程：
//   1. 从 context 中获取用户角色（由 AuthMiddleware 注入）
//   2. 检查角色是否为 "admin"
//   3. 如果是管理员，放行继续处理
//   4. 如果不是管理员，返回 403 禁止访问错误
//
// 使用方式：
//   必须在 AuthMiddleware 之后使用：
//   mux.HandleFunc("/api/admin/pages", middleware.AuthMiddleware(middleware.AdminMiddleware(handler)))
// =============================================

package middleware

import (
	"net/http"
)

// AdminMiddleware 管理员权限中间件
// 检查用户是否为管理员角色，只有 admin 角色才能访问受保护的接口
// 注意：此中间件必须在 AuthMiddleware 之后使用，因为需要从 context 读取角色信息
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从 context 获取用户角色（由 AuthMiddleware 注入）
		role, ok := GetRole(r.Context())
		if !ok {
			http.Error(w, `{"code":500,"message":"无法获取用户角色信息"}`, http.StatusInternalServerError)
			return
		}

		// 检查是否为管理员
		if role != "admin" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"code":403,"message":"权限不足，需要管理员权限"}`))
			return
		}

		// 是管理员，继续执行后续处理器
		next.ServeHTTP(w, r)
	}
}
