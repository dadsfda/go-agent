# 最终录屏前验收清单

## 1. 录屏前准备

- [ ] 确认 MySQL 已启动。
- [ ] 确认数据库 `smart_recruitment` 已存在。
- [ ] 确认腾讯云 COS Bucket `ai-1418276225` 是“私有读写”。
- [ ] 确认 `final_homework/logic-grpc-service/config/config.yaml` 已配置 MySQL、COS、DeepSeek。
- [ ] 确认真实密钥不会放进提交材料或公开文档。
- [ ] 准备一个合规简历文件：PDF、DOC 或 DOCX。
- [ ] 准备一个非法简历文件：TXT、图片或压缩包，用于演示拦截。

## 2. 启动服务

### 2.1 启动 Logic 服务

```powershell
cd C:\Users\91414\Desktop\code\project\金山agent\final_homework\logic-grpc-service
go run ./cmd/logic
```

录屏要点：

- [ ] 展示 Logic 服务启动成功。
- [ ] 口述说明：Logic 服务承载权限管控、岗位业务、MySQL、COS、Eino AI 封装。

### 2.2 启动 Web 服务

另开 PowerShell：

```powershell
cd C:\Users\91414\Desktop\code\project\金山agent\final_homework\web-gin-service
go run ./cmd/web
```

录屏要点：

- [ ] 展示 Web 服务启动成功。
- [ ] 口述说明：Web 服务只做 HTTP 接入、跨域、JWT、参数校验，并通过 gRPC 调用 Logic。

## 3. 打开前端

前端已升级为 Vue 3 + Vite，需要启动 dev server：

另开 PowerShell：

```powershell
cd C:\Users\91414\Desktop\code\project\金山agent\final_homework\hr-frontend
npm run dev
```

再开 PowerShell：

```powershell
cd C:\Users\91414\Desktop\code\project\金山agent\final_homework\user-frontend
npm run dev
```

- [ ] 打开 HR 管理端：http://localhost:5173
- [ ] 打开候选人端：http://localhost:5174
- [ ] 确认页面没有 `Failed to fetch`。

## 4. HR 管理端演示

### 4.1 HR 注册/登录

- [ ] 注册一个 HR 账号。
- [ ] 登录 HR 管理端。
- [ ] 展示登录后进入岗位管理、候选人台账、AI 问答区域。

讲解点：

- [ ] 后端签发 JWT。
- [ ] 无令牌不能访问 HR 后台接口。

### 4.2 岗位新增、编辑、下架

- [ ] 新增一个岗位。
- [ ] 点击“编辑”，修改岗位名称和描述。
- [ ] 保存修改，确认列表中展示新内容。
- [ ] 可选：下架一个岗位。

讲解点：

- [ ] HR 只能管理自己创建的岗位。
- [ ] 后端 Logic 服务做岗位归属校验。

## 5. 候选人端演示

### 5.1 游客浏览岗位

- [ ] 不登录候选人端。
- [ ] 展示公开岗位列表。
- [ ] 展示游客不能投递，只能浏览。

### 5.2 候选人注册/登录

- [ ] 注册候选人账号。
- [ ] 登录候选人端。
- [ ] 展示登录后岗位出现投递按钮。

### 5.3 未完善资料投递拦截

- [ ] 不填写档案或不上传简历时点击投递。
- [ ] 展示前端提示：请先保存完整资料或上传合规简历。

### 5.4 编辑结构化档案

- [ ] 填写姓名、联系电话、最高学历、毕业院校、工作/项目经历、核心技能标签。
- [ ] 保存档案。

讲解点：

- [ ] 档案数据写入 MySQL。

### 5.5 简历格式校验

- [ ] 上传非法格式文件，例如 TXT、图片或压缩包。
- [ ] 展示后端拦截提示。
- [ ] 上传 PDF、DOC 或 DOCX 合规简历。
- [ ] 展示上传成功。

讲解点：

- [ ] 后端同时校验文件后缀和文件头。
- [ ] 简历源文件不落本地。

### 5.6 岗位投递

- [ ] 点击“一键投递”。
- [ ] 展示投递成功。
- [ ] 刷新候选人页面，展示该岗位按钮变为“已投递”。

## 6. 私有 COS 验证

- [ ] 打开腾讯云 COS 控制台。
- [ ] 进入 Bucket：`ai-1418276225`。
- [ ] 展示访问权限是“私有读写”。
- [ ] 进入 `resumes/` 目录，展示刚上传的简历对象。
- [ ] 复制普通对象 URL，不带签名参数访问。
- [ ] 展示普通 URL 访问失败或 `AccessDenied`。

讲解点：

- [ ] Bucket 关闭匿名访问和公开读。
- [ ] 简历只存到私有 COS。
- [ ] 访问必须依赖签名 URL。

## 7. HR 候选人台账

- [ ] 回到 HR 管理端。
- [ ] 刷新候选人台账。
- [ ] 展示投递候选人的档案信息。
- [ ] 展示简历文件名和签名访问链接。
- [ ] 点击签名访问链接，确认可以打开或下载简历。
- [ ] 展示分页控件。

讲解点：

- [ ] HR 台账从 MySQL 查询真实投递数据。
- [ ] 简历访问链接由后端动态生成短有效期签名 URL。

## 8. AI 业务问答

- [ ] 在 AI 问答窗口提问：`投递总人数是多少？`
- [ ] 展示 AI 返回统计回答。
- [ ] 再提问：`哪个岗位最热门？`
- [ ] 展示 AI 返回岗位热度。
- [ ] 刷新 HR 页面。
- [ ] 展示历史 AI 对话仍然存在。

讲解点：

- [ ] HR 问题先进入 Web 服务。
- [ ] Web 服务通过 gRPC 调用 Logic 服务。
- [ ] Logic 查询 MySQL 真实业务数据。
- [ ] Logic 拼接业务上下文和用户问题。
- [ ] Logic 使用 Eino 基础 Chat 组件调用 DeepSeek。
- [ ] HR 提问和 AI 回复成对写入 MySQL。
- [ ] 本系统不做向量、不做 RAG、不做简历智能匹配。

## 9. 两层 gRPC 架构讲解

录屏中必须口述：

- [ ] `web-gin-service` 是 Gin 网关层。
- [ ] `logic-grpc-service` 是核心业务层。
- [ ] 前端只请求 Web 服务。
- [ ] Web 服务不写核心业务逻辑。
- [ ] Web 与 Logic 之间只通过 gRPC 远程调用。
- [ ] Logic 统一处理权限、岗位、档案、简历、投递、COS、MySQL、Eino。

## 10. 提交材料检查

- [ ] 所有代码在 `final_homework/` 下。
- [ ] 存在 `final_homework/hr-frontend/`。
- [ ] 存在 `final_homework/user-frontend/`。
- [ ] 存在 `final_homework/web-gin-service/`。
- [ ] 存在 `final_homework/logic-grpc-service/`。
- [ ] 存在 `final_homework/api.md`。
- [ ] 存在 `final_homework/db.md`。
- [ ] 存在 `final_homework/README.md`。
- [ ] 存在 `final_homework/answer.md`。
- [ ] `config/config.yaml` 不提交。
- [ ] `config/config.example.yaml` 可提交。

## 11. 录屏命名与提交

- [ ] 视频命名为：`姓名_学号_全栈大作业.mp4`。
- [ ] 视频不要剪辑拼接或加后期特效。
- [ ] 全程有页面操作和语音讲解。
- [ ] 提交到 WPS 表单入口。

