// Package config 提供配置管理功能
// =============================================
// 作用：
//   统一管理应用程序的所有配置项，包括数据库连接参数、服务器端口、CORS 策略等。
//   配置优先级：环境变量 > 默认值
//
// 使用方式：
//   在 main.go 中调用 config.Load() 加载配置，通过 config.Get*() 函数获取各项配置值。
//
// 环境变量说明：
//   DB_HOST     - MySQL 服务器地址（默认 localhost）
//   DB_PORT     - MySQL 端口（默认 3306）
//   DB_USER     - MySQL 用户名（默认 root）
//   DB_PASS     - MySQL 密码（默认空字符串）
//   DB_NAME     - 数据库名称（默认 tinyweb1）
//   SERVER_PORT - HTTP 服务端口（默认 :8081）
//   ALLOWED_ORIGINS - 允许的跨域来源，逗号分隔（开发环境默认 *）
// =============================================

package config

import (
	"os"
	"strings"
)

// AppConfig 应用程序全局配置结构体
// 包含所有运行时需要的配置项
type AppConfig struct {
	// 数据库相关配置
	DBHost string // MySQL 服务器地址
	DBPort string // MySQL 端口
	DBUser string // MySQL 用户名
	DBPass string // MySQL 密码
	DBName string // 数据库名称

	// 服务器配置
	ServerPort string      // HTTP 监听端口（含冒号，如 ":8081"）
	AllowedOrigins []string // 允许的 CORS 跨域来源列表
}

// appConfig 全局配置实例（包内私有，通过函数访问）
var appConfig *AppConfig

// Load 加载并初始化所有配置项
// 从环境变量读取配置，如果环境变量未设置则使用默认值
// 应在程序启动时（main.go 中）首先调用此函数
func Load() {
	appConfig = &AppConfig{
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "3306"),
		DBUser:         getEnv("DB_USER", "root"),
		DBPass:         getEnv("DB_PASS", ""),
		DBName:         getEnv("DB_NAME", "tinyweb1"),
		ServerPort:     ":" + getEnv("SERVER_PORT", "8081"),
		AllowedOrigins: parseOrigins(getEnv("ALLOWED_ORIGINS", "*")),
	}
}

// GetDBHost 返回 MySQL 服务器地址
func GetDBHost() string {
	return appConfig.DBHost
}

// GetDBPort 返回 MySQL 端口
func GetDBPort() string {
	return appConfig.DBPort
}

// GetDBUser 返回 MySQL 用户名
func GetDBUser() string {
	return appConfig.DBUser
}

// GetDBPass 返回 MySQL 密码
func GetDBPass() string {
	return appConfig.DBPass
}

// GetDBName 返回数据库名称
func GetDBName() string {
	return appConfig.DBName
}

// GetServerPort 返回 HTTP 服务监听端口（格式如 ":8081"）
func GetServerPort() string {
	return appConfig.ServerPort
}

// GetAllowedOrigins 返回允许的 CORS 跨域来源列表
func GetAllowedOrigins() []string {
	return appConfig.AllowedOrigins
}

// GetDSN 构建并返回 MySQL 数据源名称 (Data Source Name)
// 格式：用户名:密码@tcp(地址:端口)/数据库名?参数
// 用于 database/sql 的 Open() 方法连接数据库
func GetDSN() string {
	return appConfig.DBUser + ":" + appConfig.DBPass +
		"@tcp(" + appConfig.DBHost + ":" + appConfig.DBPort + ")/" +
		appConfig.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// getEnv 获取环境变量值，如果未设置则返回默认值
// 辅助函数，供内部使用
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseOrigins 解析逗号分隔的允许来源字符串为字符串数组
// 特殊处理：如果输入为 "*"，返回包含单个 "*" 的切片（表示允许所有来源）
func parseOrigins(originsStr string) []string {
	if originsStr == "*" {
		return []string{"*"}
	}
	origins := strings.Split(originsStr, ",")
	// 去除每个元素的首尾空格
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}
