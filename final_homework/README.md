# 智能招聘系统

## 1. 项目结构

```text
final_homework/
  hr-frontend/            HR 管理端静态前端
  user-frontend/          候选人用户端静态前端
  web-gin-service/        Gin Web 网关服务
  logic-grpc-service/     Logic gRPC 核心业务服务
  api.md                  接口说明
  db.md                   数据库设计
  answer.md               三方 AI Agent 平台接入设计
```

## 2. 技术架构

1. 前端只负责页面渲染、交互和 HTTP 请求。
2. Web 服务使用 Gin，负责跨域、JWT、参数校验和 HTTP 接入。
3. Web 服务通过 gRPC 调用 Logic 服务，不直接调用业务函数。gRPC 合约见 `logic-grpc-service/proto/logic.proto`。
4. Logic 服务承载账号、岗位、档案、简历、投递、AI 问答和历史记录。
5. MySQL 表结构见 `db.md`；默认连接本机 MySQL，配置 `MYSQL_DSN` 后可覆盖默认连接串。
6. 私有 OSS 和 Eino 均保留配置入口，生产部署时接入真实服务。

## 3. 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `LOGIC_ADDR` | Logic: `:9001`；Web: `127.0.0.1:9001` | Logic 监听地址或 Web 连接地址 |
| `WEB_ADDR` | `:8080` | Web 服务监听地址 |
| `JWT_SECRET` | `dev-secret-change-me` | JWT 签名密钥 |
| `MYSQL_DSN` | `root:123456@tcp(127.0.0.1:3306)/smart_recruitment?parseTime=true&charset=utf8mb4&loc=Local` | MySQL 连接串；为空时使用本机默认 MySQL |
| `COS_BUCKET` | 配置文件读取 | 腾讯云 COS 私有 Bucket 名称，可用环境变量覆盖 |
| `COS_REGION` | 配置文件读取 | 腾讯云 COS 地域，可用环境变量覆盖 |
| `COS_ENDPOINT` | 配置文件读取 | 腾讯云 COS Bucket 请求域名，可用环境变量覆盖 |
| `TENCENT_SECRET_ID` | 配置文件读取 | 腾讯云 API SecretId，可用环境变量覆盖，禁止提交真实值 |
| `TENCENT_SECRET_KEY` | 配置文件读取 | 腾讯云 API SecretKey，可用环境变量覆盖，禁止提交真实值 |
| `DEEPSEEK_API_KEY` | 配置文件读取 | DeepSeek API Key，可用环境变量覆盖，禁止提交真实值 |
| `DEEPSEEK_MODEL` | 配置文件读取 | DeepSeek 模型名，可用环境变量覆盖 |
| `DEEPSEEK_BASE_URL` | 配置文件读取 | DeepSeek OpenAI 兼容接口地址，可用环境变量覆盖 |

## 4. 启动顺序

### 4.1 启动 Logic 服务

```powershell
cd final_homework/logic-grpc-service
go run ./cmd/logic
```

默认使用本机 MySQL：

```text
root:123456@tcp(127.0.0.1:3306)/smart_recruitment?parseTime=true&charset=utf8mb4&loc=Local
```

启动前需先创建数据库：

```sql
CREATE DATABASE IF NOT EXISTS smart_recruitment DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

如需覆盖默认连接串，可设置 `MYSQL_DSN`：

```powershell
$env:MYSQL_DSN="root:password@tcp(127.0.0.1:3306)/smart_recruitment?parseTime=true&charset=utf8mb4&loc=Local"
go run ./cmd/logic
```

Logic 服务启动时会自动创建 `users`、`jobs`、`candidate_profiles`、`resumes`、`applications`、`ai_chat_histories` 表。生产环境建议使用专用数据库账号，并只授予当前库的必要权限。

### 4.2 配置腾讯云 COS

当前项目使用腾讯云 COS 作为私有 OSS。Bucket 必须保持“私有读写”，不能开启匿名访问或公开读。

按作业要求，OSS 敏感配置写入独立配置文件。先复制模板：

```powershell
cd final_homework/logic-grpc-service
copy config\config.example.yaml config\config.yaml
```

然后编辑 `config/config.yaml`，填写自己的 MySQL 与 COS 配置：

```yaml
mysql:
  dsn: "root:123456@tcp(127.0.0.1:3306)/smart_recruitment?parseTime=true&charset=utf8mb4&loc=Local"

cos:
  bucket: "ai-1418276225"
  region: "ap-guangzhou"
  endpoint: "https://ai-1418276225.cos.ap-guangzhou.myqcloud.com"
  secret_id: "你的 SecretId"
  secret_key: "你的 SecretKey"

ai:
  base_url: "https://api.deepseek.com"
  model: "deepseek-v4-flash"
  api_key: "你的 DeepSeek API Key"
```

`config/config.yaml` 已加入 `.gitignore`，不要提交真实密钥。环境变量 `MYSQL_DSN`、`COS_BUCKET`、`COS_REGION`、`COS_ENDPOINT`、`TENCENT_SECRET_ID`、`TENCENT_SECRET_KEY`、`DEEPSEEK_API_KEY`、`DEEPSEEK_MODEL`、`DEEPSEEK_BASE_URL` 仍可临时覆盖配置文件。

如果不配置 `endpoint`，系统会按 `bucket` 和 `region` 自动拼接。简历上传时，Logic 服务会先校验文件后缀和文件头，再上传到私有 COS，并将对象 Key 写入 MySQL。HR 查看台账时，后端会动态生成短有效期签名 URL。

AI 问答使用 Eino 的 OpenAI ChatModel 组件接入 DeepSeek OpenAI 兼容接口。Logic 服务会先查询 MySQL 真实业务数据，拼接业务上下文和 HR 问题，再调用 Eino 生成回答，并将问答历史写入 MySQL。

### 4.3 启动 Web 服务

另开一个终端：

```powershell
cd final_homework/web-gin-service
go run ./cmd/web
```

### 4.4 启动前端

HR 管理端：

```powershell
cd final_homework/hr-frontend
npm install
npm run dev -- --host
```

候选人用户端：

```powershell
cd final_homework/user-frontend
npm install
npm run dev -- --host
```

默认 API 地址为 `http://127.0.0.1:8080/api`。如需修改，可在浏览器控制台执行：

```js
localStorage.setItem("apiBase", "http://127.0.0.1:8080/api")
```

## 5. 演示流程

1. 打开 HR 管理端，注册 HR 账号并登录。
2. 发布一个岗位。
3. 打开候选人端，游客模式查看公开岗位。
4. 注册候选人账号，填写完整档案。
5. 上传 PDF、DOC 或 DOCX 简历。
6. 投递岗位。
7. 回到 HR 管理端，查看候选人台账。
8. 在 AI 对话窗口提问“投递总人数是多少”。
9. 刷新页面后重新加载 AI 历史。

## 6. 项目亮点

1. 两层服务架构清晰，Web 到 Logic 强制走 gRPC。
2. HR 与候选人角色隔离，岗位操作校验归属。
3. 简历上传校验文件后缀和文件头，防止伪造格式。
4. 简历文件上传到私有腾讯云 COS，HR 访问时使用短有效期签名 URL。
5. AI 问答通过 Eino 调用 DeepSeek，并基于 MySQL 业务数据生成回答，持久化历史上下文。

## 7. 测试

```powershell
cd final_homework/logic-grpc-service
go test ./...

cd ../web-gin-service
go test ./...

cd ../user-frontend
npm run build

cd ../hr-frontend
npm run build
```
