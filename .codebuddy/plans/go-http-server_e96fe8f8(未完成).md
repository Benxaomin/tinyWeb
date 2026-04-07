---
name: go-http-server
overview: 用 Go 语言编写一个后端 HTTP 服务器，监听本地 8080 端口，提供 index.html 及静态资源（图片等）的访问服务，使得浏览器通过 http://localhost:8080 即可访问个人主页。
todos:
  - id: create-go-server
    content: 创建 server/main.go 和 server/go.mod，编写 Go HTTP 静态文件服务器
    status: completed
  - id: test-server
    content: 在本地启动服务器并验证 http://localhost:8080 能正常访问页面
    status: completed
    dependencies:
      - create-go-server
  - id: commit-pr
    content: 创建新分支，提交 Go 后端代码并推送，提交 PR 到 dev/renhan0328
    status: in_progress
    dependencies:
      - test-server
---

## Product Overview

为 tinyWeb1 个人主页项目编写一个 Go 语言后端 HTTP 服务器，将当前纯静态的前端页面通过 HTTP 服务的方式在本地运行，使用户可以通过浏览器访问 http://localhost:8080 打开个人主页。

## Core Features

- 使用 Go 标准库 `net/http` 编写轻量级 HTTP 服务器
- 监听本地 8080 端口，提供服务
- 访问 `/` 时返回 `index.html` 页面
- 支持静态资源访问（webp、jpg、png、gif 等图片文件）
- 创建新分支并提交 PR 到 dev/renhan0328

## Tech Stack

- 语言: Go (Golang)
- 依赖: 仅使用 Go 标准库 `net/http`、`log`、`fmt`，无第三方依赖
- 模块名: `tinyweb1`

## Implementation Approach

使用 Go 标准库的 `http.FileServer` 作为静态文件服务器，这是 Go 官方推荐的静态文件服务方式。将 Go 服务器代码放在项目根目录的 `server/` 子文件夹中，保持与前端代码的分离。服务器从项目根目录读取静态文件，访问 `/` 自动映射到 `index.html`。

关键设计决策:

- **纯标准库**: 不引入任何第三方依赖（如 gin、echo），保持项目轻量
- **目录分离**: Go 代码放 `server/` 子目录，前端文件保持在项目根目录不动
- **`go.mod` 在 server/ 下**: 独立的 Go module，不影响前端项目结构

## Implementation Notes

- `http.FileServer` 会自动处理 `/` 路径到 `index.html` 的映射（index 默认行为）
- 静态文件路径使用 `filepath.Join` 保证跨平台兼容（Windows/Linux/macOS）
- 服务器启动后在终端打印访问地址，方便用户确认
- `go run main.go` 即可启动，无需额外构建步骤

## Directory Structure

```
d:\workspace\tinyWeb1\
├── server/
│   ├── main.go     # [NEW] Go HTTP 服务器入口。监听 8080 端口，使用 http.FileServer 提供静态文件服务，从项目根目录读取 index.html 和图片资源
│   └── go.mod      # [NEW] Go module 定义文件。模块名 tinyweb1，Go 版本 1.21+，无第三方依赖
├── index.html      # [EXISTING] 前端主页面，不做修改
├── background1.webp # [EXISTING] 静态资源
├── background2.webp # [EXISTING] 静态资源
├── background3.webp # [EXISTING] 静态资源
├── background7.webp # [EXISTING] 静态资源
├── wallpaper.gif    # [EXISTING] 静态资源
├── image.png        # [EXISTING] 静态资源
└── README.md        # [EXISTING] 项目说明
```