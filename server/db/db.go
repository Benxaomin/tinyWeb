// Package db 提供数据库连接管理和初始化功能
// =============================================
// 作用：
//   1. 建立和管理与 MySQL 的连接池（避免频繁创建/销毁连接）
//   2. 应用启动时自动创建所需的数据库表（如果不存在）
//   3. 提供全局的数据库访问入口 (*sql.DB)
//
// 使用方式：
//   在 main.go 中调用 db.Initialize() 初始化数据库连接和建表，
//   在各 handler 文件中调用 db.GetDB() 获取数据库连接执行 SQL。
//
// 连接池配置：
//   - 最大打开连接数：10（同时最多10个活跃数据库连接）
//   - 最大空闲连接数：5（保持5个空闲连接以备复用）
//   - 连接最大生命周期：30分钟（防止长时间占用连接）
//
// 自动建表：
//   首次运行或表不存在时，自动创建以下4张表：
//   - todos: 当前待办任务表
//   - todo_history: 历史归档表
//   - settings: 用户设置表
//   - guestbook: 留言板表
// =============================================

package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // 导入 MySQL 驱动（仅注册驱动，不直接使用）
	"tinyweb1/config"
)

// db 全局数据库连接实例（包内私有）
var dbInstance *sql.DB

// Initialize 初始化数据库连接池并创建所需的数据表
// 此函数应在程序启动时（main.go 中）调用一次
// 执行步骤：
//   1. 使用 config.GetDSN() 获取连接字符串
//   2. 建立 MySQL 连接并配置连接池参数
//   3. 测试连接是否正常
//   4. 创建所需的数据表（IF NOT EXISTS，不会覆盖已有数据）
//
// 错误处理：
//   连接失败或建表失败时会 log.Fatal 终止程序
func Initialize() {
	dsn := config.GetDSN()

	var err error
	// 使用 sql.Open 打开数据库连接（此时并未真正连接，只是初始化）
	dbInstance, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("❌ 无法打开数据库连接:", err)
	}

	// 配置连接池参数
	dbInstance.SetMaxOpenConns(10)                    // 最大同时打开的连接数
	dbInstance.SetMaxIdleConns(5)                     // 最大空闲连接数
	dbInstance.SetConnMaxLifetime(30 * time.Minute)   // 连接最大生存时间（防长连接失效）

	// 验证数据库连接是否可用
	if err = dbInstance.Ping(); err != nil {
		log.Fatal("❌ 无法连接到数据库:", err)
	}

	fmt.Println("✅ 数据库连接成功!")

	// 自动创建所有需要的数据表
	createTables()
}

// GetDB 获取全局数据库连接实例
// 供各个 handler 使用来执行 SQL 查询
// 注意：必须在 Initialize() 之后调用
func GetDB() *sql.DB {
	return dbInstance
}

// createTables 创建所有必需的数据表
// 使用 CREATE TABLE IF NOT EXISTS 语句，确保不会覆盖已有数据
// 建表顺序无依赖关系，可以并行执行
func createTables() {
	// 1. 创建 todos 表 - 存储当前待办任务
	createTodosTable()
	// 2. 创建 todo_history 表 - 存储历史归档的任务
	createTodoHistoryTable()
	// 3. 创建 settings 表 - 存储用户设置
	createSettingsTable()
	// 4. 创建 guestbook 表 - 存储留言板留言
	createGuestbookTable()

	fmt.Println("✅ 数据库表检查完成!")
}

// createTodosTable 创建 todos 表（如果不存在）
// 表结构说明：
//   - id: 自增主键
//   - user_id: 用户标识（索引，支持多用户查询）
//   - category: 任务分类（ENUM类型限制取值范围）
//   - text: 任务内容（VARCHAR(200)限制长度）
//   - done: 完成状态（TINYINT(1) 即 BOOL）
//   - sort_order: 排序权重（默认0）
//   - created_at / updated_at: 时间戳（自动记录）
func createTodosTable() {
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
		user_id VARCHAR(64) NOT NULL DEFAULT 'default' COMMENT '用户标识',
		category ENUM('life', 'study', 'important') NOT NULL COMMENT '分类',
		text VARCHAR(200) NOT NULL COMMENT '任务内容',
		done TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否完成',
		sort_order INT NOT NULL DEFAULT 0 COMMENT '排序序号',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	 INDEX idx_user_category (user_id, category) COMMENT '用户-分类联合索引'
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='待办任务表';
	`
	if _, err := dbInstance.Exec(query); err != nil {
		log.Fatal("❌ 创建 todos 表失败:", err)
	}
}

// createTodoHistoryTable 创建 todo_history 表（如果不存在）
// 表结构说明：
//   - 与 todos 类似，但增加了 archive_date 归档日期字段
//   - 无 sort_order 和 updated_at（历史数据不需要修改）
//   - 联合索引支持按日期+分类快速查询
func createTodoHistoryTable() {
	query := `
	CREATE TABLE IF NOT EXISTS todo_history (
		id INT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
		user_id VARCHAR(64) NOT NULL DEFAULT 'default' COMMENT '用户标识',
		archive_date DATE NOT NULL COMMENT '归档日期',
		category ENUM('life', 'study', 'important') NOT NULL COMMENT '分类',
		text VARCHAR(200) NOT NULL COMMENT '任务内容',
		done TINYINT(1) NOT NULL DEFAULT 0 COMMENT '完成状态',
		INDEX idx_user_date (user_id, archive_date) COMMENT '用户-日期联合索引'
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='待办历史归档表';
	`
	if _, err := dbInstance.Exec(query); err != nil {
		log.Fatal("❌ 创建 todo_history 表失败:", err)
	}
}

// createSettingsTable 创建 settings 表（如果不存在）
// 表结构说明：
//   - user_id 作为主键（一个用户只有一行设置记录）
//   - theme 字段存储主题偏好
//   - 使用 INSERT ... ON DUPLICATE KEY UPDATE 实现 Upsert 语义
func createSettingsTable() {
	query := `
	CREATE TABLE IF NOT EXISTS settings (
		user_id VARCHAR(64) PRIMARY KEY COMMENT '用户标识（主键）',
		theme VARCHAR(16) NOT NULL DEFAULT 'light' COMMENT '主题偏好：light/dark',
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户设置表';
	`
	if _, err := dbInstance.Exec(query); err != nil {
		log.Fatal("❌ 创建 settings 表失败:", err)
	}
}

// createGuestbookTable 创建 guestbook 表（如果不存在）
// 表结构说明：
//   - nickname: 昵称（可选，允许为空字符串表示匿名）
//   - content: 留言内容（TEXT 类型支持较长内容）
//   - id 自增主键，按插入顺序排列
//   - created_at 记录发布时间
func createGuestbookTable() {
	query := `
	CREATE TABLE IF NOT EXISTS guestbook (
		id INT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
		nickname VARCHAR(64) NOT NULL DEFAULT '' COMMENT '昵称（匿名时为空）',
		content TEXT NOT NULL COMMENT '留言内容',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间'
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='留言板表';
	`
	if _, err := dbInstance.Exec(query); err != nil {
		log.Fatal("❌ 创建 guestbook 表失败:", err)
	}
}
