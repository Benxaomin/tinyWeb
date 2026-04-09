# Day 1 学习指南 — Go + GORM + MySQL 数据库集成

> **完成日期**：2026-04-07  
> **目标**：从零搭建一个带数据库功能的 Go Web 服务，能看懂所有代码

---

## 📋 今天完成了什么

| # | 任务 | 涉及文件 | 状态 |
|---|------|---------|------|
| 1 | 配置阿里云 Go 代理，解决依赖下载超时 | 环境配置 | ✅ |
| 2 | 创建 Go 项目，引入 GORM ORM 框架 | `go.mod` | ✅ |
| 3 | 编写配置管理模块（环境变量 → 结构体） | `config/config.go` | ✅ |
| 4 | 编写数据模型定义（VisitStats 结构体） | `model/model.go` | ✅ |
| 5 | 编写数据库连接 + 自动建表 + 自动建库 | `db/db.go` | ✅ |
| 6 | 编写主程序入口（初始化流程 + HTTP服务器） | `main.go` | ✅ |
| 7 | 编写数据库 CRUD 测试代码 | `main.go` testVisitStats() | ✅ |
| 8 | 实现健康检查 API 接口 | `main.go` healthCheckHandler() | ✅ |
| 9 | 实现 CORS 跨域中间件 | `main.go` corsMiddleware() | ✅ |

---

## 🎯 最终效果

```
浏览器访问 http://localhost:8081/api/health
返回 → {"code":0,"message":"success","data":{"status":"connected","env":"development",...}}
```

程序启动后自动：
- 连接 MySQL 主库 (`tinyweb1`) 和测试库 (`tinyweb1_test`)
- 如果数据库不存在则**自动创建**
- 如果表不存在则**自动建表** (AutoMigrate)
- 执行 5 项数据库读写测试验证功能正常

---

## 📚 学习路线图（推荐顺序）

### 第一阶段：先搞懂这些基础概念（1-2天）

> ⚠️ **不要急着看代码！** 先理解这些"词汇"，再看代码就像看拼音读物。

#### 1️⃣ 什么是 ORM？
```
没有 ORM 时你这样操作数据库：
  ❌ 手写 SQL: "INSERT INTO visit_stats (ip, count) VALUES ('192.168.1.1', 1)"
  ❌ 手动解析结果集到变量

有了 ORM（GORM）后你这样操作：
  ✅ db.Create(&record)        → 自动生成 INSERT SQL
  ✅ db.Where("ip = ?", ip).First(&result)  → 自动映射到结构体
  ✅ db.Model(&obj).Update("count", 5)      → 自动生成 UPDATE SQL
```
**一句话总结**：ORM 让你用 Go 代码操作数据库，不用手写 SQL。

🔗 推荐阅读：[GORM 中文文档 - 快速开始](https://gorm.io/zh/docs/)

#### 2️⃣ 什么是 DSN（数据源名称）？
```go
// 这就是 DSN，一行字符串包含连接数据库所需的所有信息
"root:password@tcp(localhost:3306)/tinyweb1?charset=utf8mb4&parseTime=True&loc=Local"
// │    │        │   │            │      ││              │            │           └─ 时区
// │    │        │   │            │      │└── 数据库名    └─ 参数列表   └─ 自动解析时间
// │    │        │   │            │      └─ 端口
// 用户名 密码     协议  MySQL地址
```
**对应代码位置**：`config/config.go` → `buildDSN()` 函数（第 210-213 行）

#### 3️⃣ 什么是 AutoMigrate（自动迁移）？
```go
// 你只需要定义一个 Go 结构体
type VisitStats struct {
    ID         uint   // GORM 自动加主键
    VisitorIP  string // 对应 VARCHAR 字段
    VisitCount int    // 对应 INT 字段
}

// 调用这一行，GORM 自动帮你建表！
db.AutoMigrate(&VisitStats{})
// 相当于执行了: CREATE TABLE visit_stats (id INT AUTO_INCREMENT, visitor_ip VARCHAR(255), ...)
```
**对应代码位置**：`db/db.go` → `autoMigrateMainDB()` 函数（第 191-201 行）

#### 4️⃣ 什么是环境变量？为什么要用它？
```bash
# 在终端设置环境变量（不写在代码里）
set DB_PASS=123456
set APP_ENV=production

# 好处：
# ✅ 密码不会提交到 Git 仓库（安全）
# ✅ 开发环境和生产环境用不同配置（灵活）
# ✅ 不改代码就能切换数据库（方便）
```
**对应代码位置**：`config/config.go` → `Load()` 和 `getEnv()` 函数

#### 5️⃣ 什么是 CORS（跨域资源共享）？
```
前端运行在 http://localhost:3000
后端运行在 http://localhost:8081
                    ↑ 端口不同 = 不同源

浏览器会阻止前端直接请求后端（安全策略）
CORS 就是告诉浏览器："这个后端允许前端调用"
```
**对应代码位置**：`main.go` → `corsMiddleware()` 函数（第 286-323 行）

---

### 第二阶段：读懂项目结构（2-3天）

> 理解项目目录为什么这么划分，每个文件干什么。

#### 项目目录树

```
server(数据库代码)/
├── main.go              ← 🚀 程序入口（启动一切）
├── go.mod               ← 📦 依赖声明（用了哪些第三方库）
├── config/
│   └── config.go        ← ⚙️  配置管理（读环境变量）
├── model/
│   └── model.go         ← 📐 数据结构定义（数据库表 ↔ Go 结构体）
├── db/
│   └── db.go            ← 🔗 数据库连接（连库、建库、建表）
└── handler/
    ├── guestbook.go     ← 📝 留言板接口（待迁移到 GORM）
    ├── todo.go          ← ✅ 备忘录接口（待迁移到 GORM）
    └── setting.go       ← ⚙️ 设置接口（待迁移到 GORM）
```

#### 各模块职责速查表

| 文件 | 核心函数 | 干了什么 | 代码行数参考 |
|------|---------|---------|-------------|
| `config.go` | `Load()`, `GetDSN()`, `buildDSN()` | 读环境变量，构建连接字符串 | 第 70-214 行 |
| `model.go` | `VisitStats struct` | 定义数据库表的字段和类型 | 全文 |
| `db.go` | `Initialize()`, `connectDB()`, `autoMigrateMainDB()` | 连接数据库、自动建库、自动建表 | 第 66-220 行 |
| `main.go` | `main()`, `startServer()`, `testVisitStats()`, `healthCheckHandler()`, `corsMiddleware()` | 启动流程、HTTP服务、测试、健康检查、跨域 | 第 50-323 行 |

#### 程序启动顺序（按执行顺序理解）

```
main()
  │
  ├─① config.Load()          → 加载所有配置到内存
  │
  ├─② db.Initialize()         → 连接主库 tinyweb1
  │     └→ connectDB()
  │         ├→ 先 CREATE DATABASE IF NOT EXISTS
  │         ├→ gorm.Open() 连接数据库
  │         └→ autoMigrateMainDB() 建表
  │
  ├─③ db.InitializeTestDB()   → 连接测试库 tinyweb1_test（同上流程）
  │
  ├─④ testVisitStats()        → 执行 5 个 CRUD 测试
  │     ├→ 测试1: Create 插入
  │     ├→ 测试2: First 查询
  │     ├→ 测试3: Updates 更新
  │     ├→ 测试4: Count 统计
  │     └→ 测试5: 主库统计
  │
  └─⑤ startServer()           → 启动 HTTP 服务器
       ├→ 注册路由 /api/health
       ├→ 注册静态文件 /
       └→ ListenAndServe 监听 8081 端口
```

---

### 第三阶段：精读核心代码（3-5天）

> 按以下顺序逐个文件阅读，每读完一个在下面打勾 ✓

#### 📖 阅读步骤 1：`config/config.go`（最简单，从这里开始）

**必懂知识点**：
- [ ] Go 的 `struct` 结构体是什么？（类似 JavaScript 的 Object 或 Python 的 Dict）
- [ ] Go 的 `os.Getenv()` 怎么读取环境变量？
- [ ] 字符串拼接怎么写？（用 `+` 号或 `fmt.Sprintf`）
- [ ] `strings.Split()` 怎么分割字符串？

**关键代码段**：

```go
// 第 70-98 行：Load() 函数 —— 所有配置在这里初始化
appConfig = &AppConfig{
    AppEnv: getEnv("APP_ENV", "development"),  // 读 APP_ENV，没有就用默认值 "development"
    MainDB: DBConfig{
        Host: getEnv("DB_HOST", "localhost"),
        Port: getEnv("DB_PORT", "3306"),
        User: getEnv("DB_USER", "root"),
        Pass: getEnv("DB_PASS", ""),
        Name: getEnv("DB_NAME", "tinyweb1"),
    },
    // ...测试库配置类似...
}
```

```go
// 第 210-213 行：buildDSN() 函数 —— 组装数据库连接字符串
func buildDSN(db DBConfig) string {
    return db.User + ":" + db.Pass +
        "@tcp(" + db.Host + ":" + db.Port + ")/" +
        db.Name + "?charset=utf8mb4&parseTime=True&loc=Local"
}
```

---

#### 📖 阅读步骤 2：`model/model.go`（理解数据模型）

**必懂知识点**：
- [ ] GORM 的 `gorm.Model` 嵌入了什么字段？（ID, CreatedAt, UpdatedAt, DeletedAt）
- [ ] GORM Tag（标签）是什么？`gorm:"..."` 怎么控制数据库行为？
- [ ] `uniqueIndex` 是干什么的？（唯一索引，防止重复数据）
- [ ] Go 的指针类型 `*string` vs 值类型 `string` 有什么区别？

**关键代码段**：

```go
type VisitStats struct {
    gorm.Model                          // 嵌入 GORM 基础模型（自带 ID, 时间, 软删除）
    VisitorIP    string `gorm:"size:45;uniqueIndex;not null"`   // 访客IP，唯一索引
    VisitCount   int    `gorm:"default:0"`                      // 访问次数，默认0
    FirstVisitAt time.Time                                        // 首次访问时间
    LastVisitAt  time.Time                                        // 最后访问时间
    UserAgent    *string `gorm:"size:500"`                        // *string 可以为 NULL
    DeviceType   string `gorm="size:20;default:'unknown'"`        // 设备类型
    Browser      string `gorm="size:50"`                          // 浏览器
    OS           string `gorm="size:30"`                          // 操作系统
    Referrer     string `gorm="size:500"`                         // 来源页面
}
```

---

#### 📖 阅读步骤 3：`db/db.go`（核心！数据库操作都在这里）

**必懂知识点**：
- [ ] `sql.Open()` 和 `gorm.Open()` 有什么区别？
- [ ] 连接池是什么？为什么要设 MaxOpenConns、MaxIdleConns？
- [ ] `defer` 关键字做什么？（函数结束时执行清理）
- [ ] `log.Fatalf()` 和 `fmt.Println()` 区别？

**关键代码段**：

```go
// connectDB() 函数 —— 数据库连接的核心逻辑
func connectDB(dsn, host, port, name, label string) *gorm.DB {
    // ① 设置日志级别（开发模式看 SQL，生产模式只看错误）
    gormConfig := &gorm.Config{}
    if config.IsDevelopment() {
        gormConfig.Logger = logger.Default.LogMode(logger.Info)  // 开发：显示 SQL
    } else {
        gormConfig.Logger = logger.Default.LogMode(logger.Error)  // 生产：只显示错误
    }

    // ② 建立 GORM 连接
    db, err := gorm.Open(mysql.Open(dsn), gormConfig)
    if err != nil {
        log.Fatalf("❌ 无法连接%s: %v", label, err)  // 出错就终止程序
    }

    // ③ 获取底层 *sql.DB 来配置连接池
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(10)                 // 最多10个同时活跃连接
    sqlDB.SetMaxIdleConns(5)                  // 保持5个空闲连接备用
    sqlDB.SetConnMaxLifetime(30 * time.Minute)// 连接最多存活30分钟

    // ④ Ping 测试连通性
    sqlDB.Ping()

    return db
}
```

```go
// 新增的自动建库功能
// 先不指定数据库名连接 MySQL，执行 CREATE DATABASE IF NOT EXISTS
createDBDSN := fmt.Sprintf("%s@tcp(%s:%s)/?...", user+pass, host, port)
rawDB, _ := sql.Open("mysql", createDBDSN)
rawDB.Exec("CREATE DATABASE IF NOT EXISTS `tinyweb1`")
rawDB.Close()  // 关闭临时连接
```

```go
// autoMigrate() —— 一行代码搞定建表
err := db.AutoMigrate(&model.VisitStats{})
```

---

#### 📖 阅读步骤 4：`main.go`（最长，分块读）

##### 4a. 启动流程 `main()` （第 50-76 行）

```go
func main() {
    config.Load()          // 1. 加载配置
    db.Initialize()        // 2. 连接主库
    db.InitializeTestDB()  // 3. 连接测试库
    testVisitStats()       // 4. 运行测试
    startServer()          // 5. 启动 Web 服务器
}
```

##### 4b. 数据库测试 `testVisitStats()` （第 103-171 行）

```go
// 测试 1：插入
database.Create(&testRecord)

// 测试 2：查询
database.Where("visitor_ip = ?", testIP).First(&found)

// 测试 3：更新（使用 gorm.Expr 执行 SQL 表达式）
database.Model(&found).Updates(map[string]interface{}{
    "visit_count":  gorm.Expr("visit_count + 1"),  // SQL: SET visit_count = visit_count + 1
    "last_visit_at": time.Now(),
})

// 测试 4：计数
database.Model(&model.VisitStats{}).Count(&totalCount)
```

##### 4c. HTTP 服务器 `startServer()` （第 187-219 行）

```go
mux := http.NewServeMux()           // 创建路由器
mux.HandleFunc("/api/health", ...)   // 注册路由
mux.Handle("/", fs)                  // 静态文件兜底
http.ListenAndServe(":8081", corsMiddleware(mux))  // 启动监听
```

##### 4d. 中间件模式 `corsMiddleware()` （第 286-323 行）

```go
// 这是经典的中间件模式：包装 Handler，在每个请求前后添加处理逻辑
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 请求前：添加 CORS 头
        w.Header().Set("Access-Control-Allow-Origin", "*")
        
        // 处理 OPTIONS 预检请求
        if r.Method == http.MethodOptions { return }
        
        // 传递给下一个处理器
        next.ServeHTTP(w, r)
    })
}
```

---

### 第四阶段：动手练习（建议做这些）

> 光看不练假把式，试试修改代码观察变化。

#### 练习 1：新增一个数据表
```go
// 在 model/model.go 中新增一个 GuestMessage 结构体
type GuestMessage struct {
    gorm.Model
    Nickname  string `gorm:"size:50;not null"`
    Content   string `gorm:"size:500;not null"`
    IP        string `gorm:"size:45"`
}

// 在 db/db.go 的 autoMigrateMainDB() 中加入
&model.GuestMessage{},
```
编译运行，观察数据库是否自动创建了 `guest_messages` 表。

#### 练习 2：修改配置默认值
把 `config/config.go` 中的默认端口改成 `9090`，重新编译运行，访问 `http://localhost:9090/api/health`。

#### 练习 3：新增一个 API 接口
在 `main.go` 的 `startServer()` 中模仿 `healthCheckHandler` 写一个新的 `/api/version` 接口，返回版本号。

#### 练习 4：查看数据库
用 Navicat/DBeaver 打开 MySQL，看看 `tinyweb1` 库中的 `visit_stats` 表结构和里面的数据。

---

## 📖 推荐学习资源（免费）

| 资源 | 地址 | 说明 |
|------|------|------|
| Go 语言圣经（中文版） | https://gopl-zh.github.io/ | 最权威的 Go 入门书 |
| Go by Example | https://gobyexample.com/ | 代码示例驱动，边看边练 |
| GORM 中文文档 | https://gorm.io/zh/docs/ | 本项目用的 ORM 框架官方文档 |
| A Tour of Go | https://tour.go-zh.org/ | Go 官方交互式教程 |

---

## 💡 常见问题 FAQ

### Q1：为什么用 GORM 不直接写 SQL？
**A**：GORM 可以自动建表、防注入、类型转换、关联查询，开发效率高很多。复杂查询仍然可以用原生 SQL。

### Q2：`*string` 和 `string` 有什么区别？
**A**：`string` 不能为 NULL（空字符串），`*string` 可以为 NULL（nil）。数据库中有些字段需要区分"没填"和"填了空值"。

### Q3：什么是软删除（Soft Delete）？
**A**：`gorm.Model` 自带 `DeletedAt` 字段，调用 `Delete()` 不是真删数据，而是设置 `Deleted_at = 当前时间`。查询时自动过滤已删除的记录。

### Q4：连接池参数怎么调？
**A**：
- 小型应用：MaxOpenConns=5~10 够用
- 大型应用：根据 CPU 核数 × 2 来设
- 太大会耗尽 MySQL 连接数，太小会排队等待

### Q5：测试时那个 Duplicate entry 错误怎么回事？
**A**：因为 `VisitorIP` 设了 `uniqueIndex` 唯一索引，第二次插入相同 IP 就报错。实际业务中应该用 Upsert（存在就更新，不存在才插入）。

---

## 🗺️ 下一步学习方向（Day 2 预告）

完成 Day 1 后，你已经掌握了：
- ✅ Go 基础语法（struct、函数、错误处理）
- ✅ GORM 基本操作（CRUD、AutoMigrate）
- ✅ HTTP 服务器和路由
- ✅ 中间件模式

**Day 2 建议**：
1. 把现有的 `handler/todo.go`（备忘录功能）迁移到 GORM
2. 学习 RESTful API 设计规范
3. 了解 JSON 序列化/反序列化

---

*本指南由 CodeBuddy AI 生成，如有疑问随时提问！*
