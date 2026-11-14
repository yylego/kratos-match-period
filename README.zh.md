<!-- TEMPLATE (ZH) BEGIN: BADGES -->

[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/yylego/kratos-match-period/release.yml?branch=main&label=BUILD)](https://github.com/yylego/kratos-match-period/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/yylego/kratos-match-period)](https://pkg.go.dev/github.com/yylego/kratos-match-period)
[![Coverage Status](https://img.shields.io/coveralls/github/yylego/kratos-match-period/main.svg)](https://coveralls.io/github/yylego/kratos-match-period?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.26%2B-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/yylego/kratos-match-period.svg)](https://github.com/yylego/kratos-match-period/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/yylego/kratos-match-period)](https://goreportcard.com/report/github.com/yylego/kratos-match-period)

<!-- TEMPLATE (ZH) CLOSE: BADGES -->

# kratos-match-period

周期性请求采样匹配函数，基于计数器的 N 次放行控制。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->

## 英文文档

[ENGLISH README](README.md)

<!-- TEMPLATE (ZH) CLOSE: LANGUAGE NAVIGATION -->

## 核心特性

- 🔢 **计数采样**: 每 N 个请求匹配一个，每个路由各自计数
- 🏁 **首个或末个**: 默认匹配每个周期的第一个请求，用 `WithMatchFirst(false)` 改成最后一个
- 🗺️ **路由范围**: INCLUDE/EXCLUDE 模式选择采样路由
- 🪵 **调试日志**: 每次匹配结果以 debug 级别写进日志，显示与否由日志级别决定

## 安装

```bash
go get github.com/yylego/kratos-match-period
```

## 使用方法

```go
import (
    "log/slog"

    "github.com/go-kratos/kratos/v3/middleware"
    "github.com/go-kratos/kratos/v3/middleware/selector"
    "github.com/go-kratos/kratos/v3/transport/http"
    "github.com/yylego/kratos-auth/authkratos"
    "github.com/yylego/kratos-match-period/matchkratosperiod"
)

// newHTTPServer 让被包裹的中间件（比如请求日志）每 10 个 Search 请求只对 1 个生效：第 1、11、21 个……
func newHTTPServer(wrapped middleware.Middleware, logger *slog.Logger) *http.Server {
    routeScope := authkratos.NewInclude("/api.Article/Search")
    cfg := matchkratosperiod.NewConfig(routeScope, 10) // 加上 WithMatchFirst(false) 就改成第 10、20、30 个……
    matchFunc := matchkratosperiod.NewMatchFunc(cfg, logger)
    return http.NewServer(http.Middleware(selector.Server(wrapped).Match(matchFunc).Build()))
}
```

范围外的路由不匹配，被包裹的中间件不会对它们生效。

---

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 许可证类型

MIT 许可证 - 详见 [LICENSE](LICENSE)。

---

## 💬 联系与反馈

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **问题报告？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **新颖思路？** 创建 issue 讨论
- 📖 **文档疑惑？** 报告问题，帮助我们完善文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，协助解决性能问题
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：面向用户的更改需要更新文档
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来贡献此项目。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) CLOSE: STANDARD PROJECT FOOTER -->
