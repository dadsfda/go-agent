# 智能招聘系统接口说明

基础地址：`http://127.0.0.1:8080/api`

鉴权方式：登录或注册成功后返回 `token`，受保护接口使用请求头：

```http
Authorization: Bearer <token>
```

## 1. 账号接口

### POST `/auth/register`

注册 HR 或候选人。

```json
{
  "role": "hr",
  "email": "hr@example.com",
  "password": "pass"
}
```

### POST `/auth/login`

账号密码登录。

```json
{
  "role": "candidate",
  "email": "candidate@example.com",
  "password": "pass"
}
```

响应：

```json
{
  "token": "jwt-like-token",
  "user": {
    "id": 1,
    "role": "hr",
    "email": "hr@example.com"
  }
}
```

## 2. 公开岗位

### GET `/jobs`

游客、候选人均可访问，返回所有公开岗位。

## 3. 候选人接口

### GET `/candidate/profile`

获取当前候选人档案。

### POST `/candidate/profile`

保存当前候选人结构化档案。

```json
{
  "name": "张三",
  "phone": "13800000000",
  "education": "本科",
  "school": "测试大学",
  "experience": "有 Go 项目经验",
  "skills": "Go,gRPC,MySQL"
}
```

### POST `/candidate/resume`

上传简历，使用 `multipart/form-data`。

字段：

- `resume`：PDF、DOC 或 DOCX 文件。

后端校验文件后缀和文件头，合规文件上传至私有 COS，并返回短有效期签名访问链接。

### POST `/candidate/resume/parse`

可选能力：上传文字版 PDF 简历后，后端通过 Eino 调用大模型解析结构化档案字段。该能力不作为本期核心验收依赖。

字段：

- `resume`：PDF 文件。

### POST `/candidate/applications`

投递岗位。候选人必须已完善档案并上传合规简历。

```json
{
  "jobId": 2
}
```

### GET `/candidate/applications`

获取当前候选人的投递记录，用于候选人端展示“已投递”状态。

响应：

```json
{
  "items": [],
  "applications": [],
  "total": 0,
  "page": 1,
  "pageSize": 0
}
```

## 4. HR 接口

### GET `/hr/jobs`

查看当前 HR 创建的岗位。

### POST `/hr/jobs`

创建岗位。

```json
{
  "title": "后端工程师",
  "description": "负责 Gin 与 gRPC 服务开发"
}
```

### PUT `/hr/jobs/:id`

编辑或下架当前 HR 本人创建的岗位。

```json
{
  "title": "后端工程师",
  "description": "负责 Logic 服务开发",
  "status": "closed"
}
```

### DELETE `/hr/jobs/:id`

删除当前 HR 本人创建的岗位。后端实际将岗位状态标记为 `deleted`，保留历史关联数据。

### GET `/hr/applications?page=1&pageSize=10`

分页返回当前 HR 岗位的投递记录。

响应：

```json
{
  "items": [],
  "applications": [],
  "total": 0,
  "page": 1,
  "pageSize": 10
}
```

`applications` 字段用于兼容旧前端，推荐使用 `items`。

### POST `/hr/ai`

发起 AI 业务问答。

```json
{
  "question": "投递总人数是多少"
}
```

响应包含基于业务数据生成的回答，并写入对话历史。

### GET `/hr/ai/history`

获取当前 HR 的历史 AI 对话上下文。
