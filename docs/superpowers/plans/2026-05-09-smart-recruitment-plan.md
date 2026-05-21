# 智能招聘系统 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现智能招聘系统的最小可演示闭环。

**Architecture:** 两套独立静态前端调用 Gin Web 网关；Web 网关通过 gRPC 调用 Logic 服务；Logic 服务承载核心业务与数据访问。默认内存仓库用于本地演示，文档说明 MySQL 与 OSS 配置。

**Tech Stack:** Go 1.25、Gin、gRPC、JWT、HTML/CSS/JavaScript、MySQL 设计文档。

---

### Task 1: 后端核心模型与业务测试

**Files:**
- Create: `final_homework/logic-grpc-service/go.mod`
- Create: `final_homework/logic-grpc-service/internal/app/service_test.go`
- Create: `final_homework/logic-grpc-service/internal/app/service.go`

- [ ] 编写候选人资料、简历格式、投递校验、HR 岗位权限和 AI 问答测试。
- [ ] 运行 `go test ./...`，确认测试因实现缺失失败。
- [ ] 实现最小业务服务，让测试通过。

### Task 2: Logic gRPC 服务

**Files:**
- Create: `final_homework/logic-grpc-service/internal/rpc/contracts.go`
- Create: `final_homework/logic-grpc-service/internal/rpc/server.go`
- Create: `final_homework/logic-grpc-service/cmd/logic/main.go`

- [ ] 定义 JSON gRPC 编解码、请求响应结构和服务描述。
- [ ] 将业务服务挂载为 gRPC 服务。
- [ ] 提供 Logic 服务启动入口。

### Task 3: Gin Web 网关

**Files:**
- Create: `final_homework/web-gin-service/go.mod`
- Create: `final_homework/web-gin-service/internal/rpcclient/client.go`
- Create: `final_homework/web-gin-service/internal/httpapi/router.go`
- Create: `final_homework/web-gin-service/cmd/web/main.go`

- [ ] 实现 gRPC 客户端。
- [ ] 实现 HTTP 路由、JWT、跨域、角色鉴权和参数校验。
- [ ] Web 所有核心能力通过 gRPC 调 Logic。

### Task 4: 两套前端

**Files:**
- Create: `final_homework/hr-frontend/index.html`
- Create: `final_homework/hr-frontend/styles.css`
- Create: `final_homework/hr-frontend/app.js`
- Create: `final_homework/user-frontend/index.html`
- Create: `final_homework/user-frontend/styles.css`
- Create: `final_homework/user-frontend/app.js`

- [ ] 实现 HR 登录、岗位管理、候选人台账、AI 问答。
- [ ] 实现候选人岗位浏览、登录注册、档案编辑、简历上传、岗位投递。
- [ ] 前端不写核心业务逻辑，仅调用 Web API。

### Task 5: 交付文档与验证

**Files:**
- Create: `final_homework/api.md`
- Create: `final_homework/db.md`
- Create: `final_homework/README.md`
- Create: `final_homework/answer.md`

- [ ] 写明接口、数据库、启动方式和三方 Agent 接入设计。
- [ ] 运行后端测试。
- [ ] 检查 PRD 中 P0 要求是否都有对应实现或文档说明。

