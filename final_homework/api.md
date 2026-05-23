# 智能招聘系统接口文档

## 1. 通用说明

基础地址：

```text
http://127.0.0.1:8080/api
```

接口返回格式：

- 成功：返回业务 JSON。
- 失败：统一返回 `{"error": "错误说明"}`。

鉴权方式：

- 登录或注册成功后，Web 服务签发 token。
- 受保护接口需要在请求头中携带：

```http
Authorization: Bearer <token>
```

角色说明：

- `hr`：HR 管理端账号，可管理岗位、查看投递、使用 AI 问答。
- `candidate`：候选人账号，可维护档案、上传简历、投递岗位。

通用错误码：

| HTTP 状态码 | 含义 |
| --- | --- |
| 400 | 请求参数错误、业务规则校验失败、gRPC Logic 服务返回业务错误 |
| 401 | 未登录、token 无效、token 过期、角色不匹配 |
| 500 | Web 服务签发 token 等内部异常 |

通用错误示例：

```json
{
  "error": "请求参数错误"
}
```

## 2. 账号接口

### 2.1 注册账号

接口路径：`/auth/register`

请求方法：`POST`

鉴权说明：无需登录。

请求 Body：

```json
{
  "role": "hr",
  "email": "hr@example.com",
  "password": "pass123456"
}
```

参数说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| role | string | 是 | 用户角色，只能是 `hr` 或 `candidate` |
| email | string | 是 | 登录邮箱，系统内唯一 |
| password | string | 是 | 登录密码 |

成功返回：

```json
{
  "token": "base64-payload.signature",
  "user": {
    "id": 1,
    "role": "hr",
    "email": "hr@example.com"
  }
}
```

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 400 | 角色不是 `hr` 或 `candidate` |
| 400 | 邮箱或密码为空 |
| 400 | 邮箱已存在 |
| 500 | token 签发失败 |

失败示例：

```json
{
  "error": "角色必须是 hr 或 candidate"
}
```

### 2.2 登录账号

接口路径：`/auth/login`

请求方法：`POST`

鉴权说明：无需登录。

请求 Body：

```json
{
  "role": "candidate",
  "email": "candidate@example.com",
  "password": "pass123456"
}
```

参数说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| role | string | 是 | 登录角色，只能是 `hr` 或 `candidate` |
| email | string | 是 | 登录邮箱 |
| password | string | 是 | 登录密码 |

成功返回：

```json
{
  "token": "base64-payload.signature",
  "user": {
    "id": 2,
    "role": "candidate",
    "email": "candidate@example.com"
  }
}
```

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 400 | 请求 JSON 结构错误 |
| 400 | 账号不存在、密码错误或角色不匹配 |
| 500 | token 签发失败 |

失败示例：

```json
{
  "error": "账号或密码错误"
}
```

## 3. 公开岗位接口

### 3.1 游客查看公开岗位

接口路径：`/jobs`

请求方法：`GET`

鉴权说明：无需登录。游客、HR、候选人均可访问。

请求参数：无。

成功返回：

```json
{
  "jobs": [
    {
      "id": 1,
      "ownerHrId": 10,
      "title": "后端工程师",
      "description": "负责 Gin、gRPC、MySQL 后端服务开发",
      "status": "open",
      "createdAt": "2026-05-23T15:00:00Z"
    }
  ]
}
```

返回说明：

- 只返回 `status = open` 的公开岗位。
- `ownerHrId` 是创建该岗位的 HR 用户 ID。

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 400 | Logic 服务查询岗位失败 |

## 4. 候选人接口

### 4.1 获取候选人档案

接口路径：`/candidate/profile`

请求方法：`GET`

鉴权说明：需要候选人 token。

请求参数：无。

成功返回：

```json
{
  "candidateId": 2,
  "name": "张三",
  "phone": "13800000000",
  "education": "本科",
  "school": "测试大学",
  "experience": "有 Go 项目经验",
  "skills": "Go,gRPC,MySQL"
}
```

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带候选人 token 或角色不是候选人 |
| 400 | 候选人档案不存在 |

失败示例：

```json
{
  "error": "候选人档案不存在"
}
```

### 4.2 保存候选人档案

接口路径：`/candidate/profile`

请求方法：`POST`

鉴权说明：需要候选人 token。

请求 Body：

```json
{
  "name": "张三",
  "phone": "13800000000",
  "education": "本科",
  "school": "测试大学",
  "experience": "负责 Go 后端项目开发",
  "skills": "Go,gRPC,MySQL"
}
```

参数说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| name | string | 是 | 候选人姓名 |
| phone | string | 是 | 联系电话 |
| education | string | 是 | 最高学历 |
| school | string | 是 | 毕业院校 |
| experience | string | 是 | 工作或项目经历 |
| skills | string | 是 | 核心技能标签，最长 255 个字符 |

成功返回：

```json
{
  "ok": true
}
```

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带候选人 token 或角色不是候选人 |
| 400 | 请求 JSON 结构错误 |
| 400 | 必填字段为空 |
| 400 | 技能标签超过 255 个字符 |

失败示例：

```json
{
  "error": "候选人档案必填字段不能为空"
}
```

### 4.3 上传简历到私有 OSS

接口路径：`/candidate/resume`

请求方法：`POST`

鉴权说明：需要候选人 token。

请求类型：`multipart/form-data`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| resume | file | 是 | 简历文件，仅支持 PDF、DOC、DOCX |

业务规则：

- Web 服务读取文件内容后通过 gRPC 转发给 Logic 服务。
- Logic 服务同时校验文件后缀和文件头。
- 合规简历上传到私有 COS Bucket。
- 数据库只保存简历元信息和 OSS 对象 Key，不保存本地源文件。
- 返回的 `signedUrl` 是私有 OSS 的短有效期签名访问链接。

成功返回：

```json
{
  "id": 1,
  "candidateId": 2,
  "fileName": "resume.pdf",
  "objectKey": "resumes/2/1716450000000000000-resume.pdf",
  "signedUrl": "https://example.cos.ap-guangzhou.myqcloud.com/resumes/2/resume.pdf?sign=xxx",
  "uploadedAt": "2026-05-23T15:00:00Z"
}
```

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带候选人 token 或角色不是候选人 |
| 400 | 未选择文件 |
| 400 | 文件读取失败 |
| 400 | 文件不是真实 PDF、DOC 或 DOCX |
| 400 | COS 配置缺失或上传失败 |

失败示例：

```json
{
  "error": "简历仅支持真实 PDF、DOC、DOCX 文件"
}
```

### 4.4 解析简历并自动填充档案

接口路径：`/candidate/resume/parse`

请求方法：`POST`

鉴权说明：需要候选人 token。

请求类型：`multipart/form-data`

请求参数：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| resume | file | 是 | 文字版 PDF、DOC 或 DOCX 简历 |

业务流程：

1. Web 服务接收简历文件。
2. Web 服务通过 gRPC 调用 Logic 服务。
3. Logic 服务提取简历文本：PDF 使用 PyMuPDF，DOCX 使用 OpenXML，DOC 使用 `antiword`。
4. Logic 服务通过 Eino 调用大模型提取结构化字段。
5. 前端拿到字段后自动填入档案表单，由用户确认后再调用 `POST /candidate/profile` 保存。

成功返回：

```json
{
  "candidateId": 0,
  "name": "张三",
  "phone": "13800000000",
  "education": "本科",
  "school": "测试大学",
  "experience": "负责 Go 后端服务开发，参与 Gin、gRPC、MySQL 项目",
  "skills": "Go,gRPC,MySQL,Vue"
}
```

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带候选人 token 或角色不是候选人 |
| 400 | 未选择文件 |
| 400 | 文件读取失败 |
| 400 | 文件文本提取失败 |
| 400 | AI 服务未配置，无法解析简历 |
| 400 | Eino 大模型返回为空或返回格式异常 |

失败示例：

```json
{
  "error": "AI 服务未配置，无法解析简历"
}
```

### 4.5 投递岗位

接口路径：`/candidate/applications`

请求方法：`POST`

鉴权说明：需要候选人 token。

请求 Body：

```json
{
  "jobId": 1
}
```

参数说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| jobId | number | 是 | 要投递的岗位 ID |

业务规则：

- 只能投递 `status = open` 的岗位。
- 候选人必须已保存完整结构化档案。
- 候选人必须已上传合规简历。
- 同一个候选人不能重复投递同一个岗位。

成功返回：

```json
{
  "id": 1,
  "jobId": 1,
  "candidateId": 2,
  "resumeId": 1,
  "createdAt": "2026-05-23T15:00:00Z",
  "job": {
    "id": 1,
    "ownerHrId": 10,
    "title": "后端工程师",
    "description": "负责 Gin、gRPC、MySQL 后端服务开发",
    "status": "open",
    "createdAt": "2026-05-23T14:00:00Z"
  },
  "candidate": {
    "id": 2,
    "role": "candidate",
    "email": "candidate@example.com"
  },
  "profile": {
    "candidateId": 2,
    "name": "张三",
    "phone": "13800000000",
    "education": "本科",
    "school": "测试大学",
    "experience": "负责 Go 后端项目开发",
    "skills": "Go,gRPC,MySQL"
  },
  "resume": {
    "id": 1,
    "candidateId": 2,
    "fileName": "resume.pdf",
    "objectKey": "resumes/2/1716450000000000000-resume.pdf",
    "signedUrl": "https://example.cos.ap-guangzhou.myqcloud.com/resumes/2/resume.pdf?sign=xxx",
    "uploadedAt": "2026-05-23T15:00:00Z"
  }
}
```

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带候选人 token 或角色不是候选人 |
| 400 | 请求 JSON 结构错误 |
| 400 | 岗位不存在或已下架 |
| 400 | 未完善结构化个人档案 |
| 400 | 未上传合规简历 |
| 400 | 重复投递同一岗位 |

失败示例：

```json
{
  "error": "请先上传合规简历"
}
```

### 4.6 查看候选人投递记录

接口路径：`/candidate/applications`

请求方法：`GET`

鉴权说明：需要候选人 token。

请求参数：无。

成功返回：

```json
{
  "items": [
    {
      "id": 1,
      "jobId": 1,
      "candidateId": 2,
      "resumeId": 1,
      "createdAt": "2026-05-23T15:00:00Z",
      "job": {
        "id": 1,
        "ownerHrId": 10,
        "title": "后端工程师",
        "description": "负责 Gin、gRPC、MySQL 后端服务开发",
        "status": "open",
        "createdAt": "2026-05-23T14:00:00Z"
      },
      "candidate": {
        "id": 0,
        "role": "",
        "email": ""
      },
      "profile": {
        "candidateId": 0,
        "name": "",
        "phone": "",
        "education": "",
        "school": "",
        "experience": "",
        "skills": ""
      },
      "resume": {
        "id": 1,
        "candidateId": 2,
        "fileName": "resume.pdf",
        "objectKey": "resumes/2/1716450000000000000-resume.pdf",
        "signedUrl": "",
        "uploadedAt": "2026-05-23T15:00:00Z"
      }
    }
  ],
  "applications": [],
  "total": 1,
  "page": 1,
  "pageSize": 1
}
```

说明：

- `items` 和 `applications` 都是兼容字段，前端优先使用 `items`。
- 候选人端主要使用该接口判断岗位是否已投递。

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带候选人 token 或角色不是候选人 |
| 400 | Logic 服务查询投递记录失败 |

## 5. HR 接口

### 5.1 查看 HR 岗位列表

接口路径：`/hr/jobs`

请求方法：`GET`

鉴权说明：需要 HR token。

请求参数：无。

成功返回：

```json
{
  "jobs": [
    {
      "id": 1,
      "ownerHrId": 10,
      "title": "后端工程师",
      "description": "负责 Gin、gRPC、MySQL 后端服务开发",
      "status": "open",
      "createdAt": "2026-05-23T14:00:00Z"
    }
  ]
}
```

说明：

- 返回所有未删除岗位，即 `status != deleted`。
- 前端可展示全部岗位。
- 编辑、下架、删除操作仍由后端校验岗位是否属于当前 HR。

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带 HR token 或角色不是 HR |
| 400 | Logic 服务查询岗位失败 |

### 5.2 创建岗位

接口路径：`/hr/jobs`

请求方法：`POST`

鉴权说明：需要 HR token。

请求 Body：

```json
{
  "title": "后端工程师",
  "description": "负责 Gin、gRPC、MySQL 后端服务开发"
}
```

参数说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| title | string | 是 | 岗位名称 |
| description | string | 是 | 岗位描述 |

成功返回：

```json
{
  "id": 1,
  "ownerHrId": 10,
  "title": "后端工程师",
  "description": "负责 Gin、gRPC、MySQL 后端服务开发",
  "status": "open",
  "createdAt": "2026-05-23T14:00:00Z"
}
```

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带 HR token 或角色不是 HR |
| 400 | 请求 JSON 结构错误 |
| 400 | 岗位名称或描述为空 |

失败示例：

```json
{
  "error": "岗位名称和描述不能为空"
}
```

### 5.3 编辑或下架岗位

接口路径：`/hr/jobs/:id`

请求方法：`PUT`

鉴权说明：需要 HR token，只能操作当前 HR 自己创建的岗位。

路径参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| id | number | 是 | 岗位 ID |

请求 Body：

```json
{
  "title": "高级后端工程师",
  "description": "负责 Logic gRPC 服务和 MySQL 数据建模",
  "status": "closed"
}
```

参数说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| title | string | 是 | 岗位名称；当 `status` 不是 `deleted` 时不能为空 |
| description | string | 是 | 岗位描述；当 `status` 不是 `deleted` 时不能为空 |
| status | string | 是 | 岗位状态：`open`、`closed`、`deleted` |

成功返回：

```json
{
  "ok": true
}
```

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带 HR token 或角色不是 HR |
| 400 | 岗位 ID 非法 |
| 400 | 请求 JSON 结构错误 |
| 400 | 岗位状态不是 `open`、`closed` 或 `deleted` |
| 400 | 岗位名称或描述为空 |
| 400 | 岗位不存在或无权操作他人岗位 |

失败示例：

```json
{
  "error": "岗位不存在或无权操作他人岗位"
}
```

### 5.4 删除岗位

接口路径：`/hr/jobs/:id`

请求方法：`DELETE`

鉴权说明：需要 HR token，只能删除当前 HR 自己创建的岗位。

路径参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| id | number | 是 | 岗位 ID |

业务说明：

- 后端不物理删除岗位。
- 删除操作实际将岗位 `status` 更新为 `deleted`。
- 历史投递关联数据保留。

成功返回：

```json
{
  "ok": true
}
```

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带 HR token 或角色不是 HR |
| 400 | 岗位 ID 非法 |
| 400 | 岗位不存在或无权操作他人岗位 |

### 5.5 分页查看岗位投递台账

接口路径：`/hr/applications`

请求方法：`GET`

鉴权说明：需要 HR token。

Query 参数：

| 参数 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| page | number | 否 | 1 | 页码，小于 1 时按 1 处理 |
| pageSize | number | 否 | 10 | 每页数量，小于 1 时按 10 处理，大于 100 时按 100 处理 |

成功返回：

```json
{
  "items": [
    {
      "id": 1,
      "jobId": 1,
      "candidateId": 2,
      "resumeId": 1,
      "createdAt": "2026-05-23T15:00:00Z",
      "job": {
        "id": 1,
        "ownerHrId": 10,
        "title": "后端工程师",
        "description": "负责 Gin、gRPC、MySQL 后端服务开发",
        "status": "open",
        "createdAt": "2026-05-23T14:00:00Z"
      },
      "candidate": {
        "id": 2,
        "role": "candidate",
        "email": "candidate@example.com"
      },
      "profile": {
        "candidateId": 2,
        "name": "张三",
        "phone": "13800000000",
        "education": "本科",
        "school": "测试大学",
        "experience": "负责 Go 后端项目开发",
        "skills": "Go,gRPC,MySQL"
      },
      "resume": {
        "id": 1,
        "candidateId": 2,
        "fileName": "resume.pdf",
        "objectKey": "resumes/2/1716450000000000000-resume.pdf",
        "signedUrl": "https://example.cos.ap-guangzhou.myqcloud.com/resumes/2/resume.pdf?sign=xxx",
        "uploadedAt": "2026-05-23T15:00:00Z"
      }
    }
  ],
  "applications": [
    {
      "id": 1,
      "jobId": 1,
      "candidateId": 2,
      "resumeId": 1,
      "createdAt": "2026-05-23T15:00:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "pageSize": 10
}
```

说明：

- 只返回当前 HR 创建岗位下的投递记录。
- 返回内容包含岗位、候选人账号、结构化档案和简历记录。
- HR 访问简历时，后端动态生成私有 OSS 签名 URL。
- `applications` 字段用于兼容旧前端，推荐使用 `items`。

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带 HR token 或角色不是 HR |
| 400 | Logic 服务查询投递台账失败 |

### 5.6 AI 业务问答

接口路径：`/hr/ai`

请求方法：`POST`

鉴权说明：需要 HR token。

请求 Body：

```json
{
  "question": "投递总人数是多少"
}
```

参数说明：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| question | string | 是 | HR 自然语言问题 |

业务流程：

1. Logic 服务根据当前 HR ID 查询 MySQL 真实业务数据。
2. 生成业务上下文，包括投递总人数、岗位热度、各岗位投递统计、最近候选人明细。
3. 通过 Eino 调用大模型生成自然语言回答。
4. 将 HR 问题和 AI 回答成对写入 `ai_chat_histories` 表。

成功返回：

```json
{
  "id": 1,
  "hrId": 10,
  "question": "投递总人数是多少",
  "answer": "当前你的岗位投递总人数为 1 人，岗位热度最高的是后端工程师。",
  "createdAt": "2026-05-23T15:00:00Z"
}
```

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带 HR token 或角色不是 HR |
| 400 | 请求 JSON 结构错误 |
| 400 | 问题为空 |
| 400 | AI 服务未配置或 Eino 调用失败 |
| 400 | MySQL 业务数据查询失败 |

失败示例：

```json
{
  "error": "问题不能为空"
}
```

### 5.7 获取 AI 历史对话

接口路径：`/hr/ai/history`

请求方法：`GET`

鉴权说明：需要 HR token。

请求参数：无。

成功返回：

```json
{
  "messages": [
    {
      "id": 1,
      "hrId": 10,
      "question": "投递总人数是多少",
      "answer": "当前你的岗位投递总人数为 1 人。",
      "createdAt": "2026-05-23T15:00:00Z"
    }
  ]
}
```

说明：

- 只返回当前登录 HR 自己的 AI 历史。
- 前端进入 AI 页面时调用该接口恢复历史上下文。

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带 HR token 或角色不是 HR |
| 400 | Logic 服务查询 AI 历史失败 |

### 5.8 清空 AI 历史对话

接口路径：`/hr/ai/history`

请求方法：`DELETE`

鉴权说明：需要 HR token。

请求参数：无。

成功返回：

```json
{
  "ok": true
}
```

说明：

- 只删除当前登录 HR 账号自己的聊天历史。
- 不影响其他 HR 的聊天记录。

失败错误码：

| HTTP 状态码 | 错误含义 |
| --- | --- |
| 401 | 未携带 HR token 或角色不是 HR |
| 400 | Logic 服务删除 AI 历史失败 |
