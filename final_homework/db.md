# 智能招聘系统数据库设计文档

## 1. 设计概述

本项目使用 MySQL 存储智能招聘系统核心业务数据。Logic gRPC 服务启动时会自动执行建表逻辑，创建以下 6 张业务表：

1. `users`：双角色账号表，统一存储 HR 和候选人账号。
2. `jobs`：招聘岗位表，存储 HR 发布的岗位。
3. `candidate_profiles`：候选人结构化档案表。
4. `resumes`：简历元信息表，存储 OSS 对象 Key，不保存本地简历源文件。
5. `applications`：岗位投递表，关联候选人、岗位和简历。
6. `ai_chat_histories`：AI 问答历史表，保存 HR 问题和 AI 回答。

说明：

- 当前建表 SQL 未使用物理 `FOREIGN KEY` 约束，外键关系由业务逻辑和查询条件保证。
- 文档中的“外键”表示业务外键关联。
- 所有时间字段由 MySQL 默认值或业务代码生成。
- 服务启动时使用 `CREATE TABLE IF NOT EXISTS`，不会删除已有数据。

## 2. 表结构设计

### 2.1 users 用户账号表

表名：`users`

业务含义：统一保存 HR 和候选人账号，用 `role` 区分用户角色。登录、鉴权、岗位归属、档案归属、投递归属都基于该表用户 ID。

建表结构：

```sql
CREATE TABLE IF NOT EXISTS users (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  role VARCHAR(20) NOT NULL,
  email VARCHAR(120) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

字段说明：

| 字段名 | 数据类型 | 主键 | 外键 | 可空 | 索引 | 业务含义 |
| --- | --- | --- | --- | --- | --- | --- |
| id | BIGINT AUTO_INCREMENT | 是 | 否 | 否 | PRIMARY KEY | 用户唯一 ID，其他表通过该 ID 关联用户 |
| role | VARCHAR(20) | 否 | 否 | 否 | 无 | 用户角色，只允许业务层写入 `hr` 或 `candidate` |
| email | VARCHAR(120) | 否 | 否 | 否 | UNIQUE | 登录邮箱，全系统唯一 |
| password_hash | VARCHAR(255) | 否 | 否 | 否 | 无 | 密码哈希值，不保存明文密码 |
| created_at | DATETIME | 否 | 否 | 否 | 无 | 账号创建时间 |
| updated_at | DATETIME | 否 | 否 | 否 | 无 | 账号最近更新时间 |

索引设置：

| 索引名 | 类型 | 字段 | 用途 |
| --- | --- | --- | --- |
| PRIMARY | 主键索引 | id | 按用户 ID 查询账号 |
| email | 唯一索引 | email | 保证邮箱唯一，用于登录查询 |

业务约束：

- `role = hr` 的用户可以创建和管理岗位。
- `role = candidate` 的用户可以维护档案、上传简历和投递岗位。
- 登录时同时校验 `email`、`password_hash` 和 `role`。

### 2.2 jobs 招聘岗位表

表名：`jobs`

业务含义：保存 HR 创建的岗位。岗位有公开、下架、删除三种状态，游客只看到公开岗位，HR 管理端可以看到未删除岗位。

建表结构：

```sql
CREATE TABLE IF NOT EXISTS jobs (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  owner_hr_id BIGINT NOT NULL,
  title VARCHAR(120) NOT NULL,
  description TEXT NOT NULL,
  status VARCHAR(20) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_jobs_owner_hr_id (owner_hr_id)
);
```

字段说明：

| 字段名 | 数据类型 | 主键 | 外键 | 可空 | 索引 | 业务含义 |
| --- | --- | --- | --- | --- | --- | --- |
| id | BIGINT AUTO_INCREMENT | 是 | 否 | 否 | PRIMARY KEY | 岗位唯一 ID |
| owner_hr_id | BIGINT | 否 | 是，关联 `users.id` | 否 | idx_jobs_owner_hr_id | 创建该岗位的 HR 用户 ID |
| title | VARCHAR(120) | 否 | 否 | 否 | 无 | 岗位名称 |
| description | TEXT | 否 | 否 | 否 | 无 | 岗位描述 |
| status | VARCHAR(20) | 否 | 否 | 否 | 无 | 岗位状态：`open`、`closed`、`deleted` |
| created_at | DATETIME | 否 | 否 | 否 | 无 | 岗位创建时间 |
| updated_at | DATETIME | 否 | 否 | 否 | 无 | 岗位最近更新时间 |

索引设置：

| 索引名 | 类型 | 字段 | 用途 |
| --- | --- | --- | --- |
| PRIMARY | 主键索引 | id | 按岗位 ID 查询、更新、投递 |
| idx_jobs_owner_hr_id | 普通索引 | owner_hr_id | 查询某个 HR 创建的岗位和投递统计 |

业务约束：

- 创建岗位时，`owner_hr_id` 必须是角色为 `hr` 的用户。
- `status = open`：公开岗位，游客和候选人可见，可投递。
- `status = closed`：下架岗位，不在公开岗位列表中，不可投递。
- `status = deleted`：逻辑删除岗位，不在 HR 岗位管理列表中展示。
- HR 只能编辑、下架、删除自己创建的岗位。

### 2.3 candidate_profiles 候选人档案表

表名：`candidate_profiles`

业务含义：保存候选人的结构化个人档案。候选人投递前必须先完善该表数据。

建表结构：

```sql
CREATE TABLE IF NOT EXISTS candidate_profiles (
  candidate_id BIGINT PRIMARY KEY,
  name VARCHAR(80) NOT NULL,
  phone VARCHAR(40) NOT NULL,
  education VARCHAR(80) NOT NULL,
  school VARCHAR(120) NOT NULL,
  experience TEXT NOT NULL,
  skills VARCHAR(255) NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

字段说明：

| 字段名 | 数据类型 | 主键 | 外键 | 可空 | 索引 | 业务含义 |
| --- | --- | --- | --- | --- | --- | --- |
| candidate_id | BIGINT | 是 | 是，关联 `users.id` | 否 | PRIMARY KEY | 候选人用户 ID，一个候选人只有一份结构化档案 |
| name | VARCHAR(80) | 否 | 否 | 否 | 无 | 候选人姓名 |
| phone | VARCHAR(40) | 否 | 否 | 否 | 无 | 候选人联系电话 |
| education | VARCHAR(80) | 否 | 否 | 否 | 无 | 候选人最高学历 |
| school | VARCHAR(120) | 否 | 否 | 否 | 无 | 候选人毕业院校 |
| experience | TEXT | 否 | 否 | 否 | 无 | 候选人工作或项目经历 |
| skills | VARCHAR(255) | 否 | 否 | 否 | 无 | 候选人核心技能标签，业务层限制最长 255 个字符 |
| created_at | DATETIME | 否 | 否 | 否 | 无 | 档案创建时间 |
| updated_at | DATETIME | 否 | 否 | 否 | 无 | 档案最近更新时间 |

索引设置：

| 索引名 | 类型 | 字段 | 用途 |
| --- | --- | --- | --- |
| PRIMARY | 主键索引 | candidate_id | 按候选人 ID 获取和更新档案 |

业务约束：

- `candidate_id` 必须对应 `users.role = candidate`。
- 所有档案字段均为必填。
- 保存档案使用 `INSERT ... ON DUPLICATE KEY UPDATE`，支持候选人反复修改。
- 投递岗位时，如果该表没有候选人记录，则拒绝投递。

### 2.4 resumes 简历元信息表

表名：`resumes`

业务含义：保存候选人简历的元信息和私有 OSS 对象 Key。系统不在本地保存简历源文件。

建表结构：

```sql
CREATE TABLE IF NOT EXISTS resumes (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  candidate_id BIGINT NOT NULL,
  file_name VARCHAR(255) NOT NULL,
  file_ext VARCHAR(20) NOT NULL,
  oss_object_key VARCHAR(500) NOT NULL,
  uploaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_resumes_candidate_id (candidate_id)
);
```

字段说明：

| 字段名 | 数据类型 | 主键 | 外键 | 可空 | 索引 | 业务含义 |
| --- | --- | --- | --- | --- | --- | --- |
| id | BIGINT AUTO_INCREMENT | 是 | 否 | 否 | PRIMARY KEY | 简历记录唯一 ID |
| candidate_id | BIGINT | 否 | 是，关联 `users.id` | 否 | idx_resumes_candidate_id | 上传简历的候选人用户 ID |
| file_name | VARCHAR(255) | 否 | 否 | 否 | 无 | 候选人上传时的原始文件名 |
| file_ext | VARCHAR(20) | 否 | 否 | 否 | 无 | 文件扩展名，如 `.pdf`、`.doc`、`.docx` |
| oss_object_key | VARCHAR(500) | 否 | 否 | 否 | 无 | 私有 OSS 中的对象 Key |
| uploaded_at | DATETIME | 否 | 否 | 否 | 无 | 简历上传时间 |

索引设置：

| 索引名 | 类型 | 字段 | 用途 |
| --- | --- | --- | --- |
| PRIMARY | 主键索引 | id | 按简历 ID 关联投递记录 |
| idx_resumes_candidate_id | 普通索引 | candidate_id | 查询候选人最近上传的简历 |

业务约束：

- `candidate_id` 必须对应 `users.role = candidate`。
- 简历文件必须通过后端扩展名和文件头双重校验。
- 仅允许真实 PDF、DOC、DOCX 文件。
- 简历上传后，文件内容直接写入私有 OSS。
- 投递时使用候选人最近上传的一份简历。
- HR 查看投递台账时，后端根据 `oss_object_key` 动态生成签名 URL。

### 2.5 applications 岗位投递表

表名：`applications`

业务含义：记录候选人对岗位的投递行为，同时绑定投递时使用的简历。

建表结构：

```sql
CREATE TABLE IF NOT EXISTS applications (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  job_id BIGINT NOT NULL,
  candidate_id BIGINT NOT NULL,
  resume_id BIGINT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY uk_applications_job_candidate (job_id, candidate_id)
);
```

字段说明：

| 字段名 | 数据类型 | 主键 | 外键 | 可空 | 索引 | 业务含义 |
| --- | --- | --- | --- | --- | --- | --- |
| id | BIGINT AUTO_INCREMENT | 是 | 否 | 否 | PRIMARY KEY | 投递记录唯一 ID |
| job_id | BIGINT | 否 | 是，关联 `jobs.id` | 否 | uk_applications_job_candidate | 被投递的岗位 ID |
| candidate_id | BIGINT | 否 | 是，关联 `users.id` | 否 | uk_applications_job_candidate | 发起投递的候选人用户 ID |
| resume_id | BIGINT | 否 | 是，关联 `resumes.id` | 否 | 无 | 本次投递绑定的简历记录 ID |
| created_at | DATETIME | 否 | 否 | 否 | 无 | 投递创建时间 |

索引设置：

| 索引名 | 类型 | 字段 | 用途 |
| --- | --- | --- | --- |
| PRIMARY | 主键索引 | id | 按投递 ID 查询投递详情 |
| uk_applications_job_candidate | 唯一索引 | job_id, candidate_id | 防止同一候选人重复投递同一岗位 |

业务约束：

- `job_id` 必须对应 `jobs.status = open` 的岗位。
- `candidate_id` 必须对应 `users.role = candidate`。
- 候选人投递前必须存在完整 `candidate_profiles` 记录。
- 候选人投递前必须存在至少一条合规 `resumes` 记录。
- 同一候选人对同一岗位只能投递一次。
- HR 投递台账只查询自己创建岗位下的投递记录。

### 2.6 ai_chat_histories AI 对话历史表

表名：`ai_chat_histories`

业务含义：保存 HR 与 AI 招聘数据助手的历史对话。每条记录保存一轮 HR 问题和 AI 回答。

建表结构：

```sql
CREATE TABLE IF NOT EXISTS ai_chat_histories (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  hr_id BIGINT NOT NULL,
  question TEXT NOT NULL,
  answer TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_ai_chat_histories_hr_id (hr_id)
);
```

字段说明：

| 字段名 | 数据类型 | 主键 | 外键 | 可空 | 索引 | 业务含义 |
| --- | --- | --- | --- | --- | --- | --- |
| id | BIGINT AUTO_INCREMENT | 是 | 否 | 否 | PRIMARY KEY | AI 对话记录唯一 ID |
| hr_id | BIGINT | 否 | 是，关联 `users.id` | 否 | idx_ai_chat_histories_hr_id | 发起 AI 问答的 HR 用户 ID |
| question | TEXT | 否 | 否 | 否 | 无 | HR 输入的自然语言问题 |
| answer | TEXT | 否 | 否 | 否 | 无 | AI 根据 MySQL 业务数据生成的回答 |
| created_at | DATETIME | 否 | 否 | 否 | 无 | 对话记录创建时间 |

索引设置：

| 索引名 | 类型 | 字段 | 用途 |
| --- | --- | --- | --- |
| PRIMARY | 主键索引 | id | 按对话 ID 排序和定位记录 |
| idx_ai_chat_histories_hr_id | 普通索引 | hr_id | 加载某个 HR 的全部历史对话 |

业务约束：

- `hr_id` 必须对应 `users.role = hr`。
- 每次 AI 问答成功后，自动写入一条历史记录。
- 查询历史时只返回当前 HR 自己的记录。
- 清空历史时只删除当前 HR 自己的记录，不影响其他 HR。
- Eino 调用时会取最近若干轮历史按时间正序拼接上下文。

## 3. 关联关系设计

### 3.1 用户与岗位

关系：

```text
users.id (hr) 1 ---- N jobs.owner_hr_id
```

说明：

- 一个 HR 可以创建多个岗位。
- 一个岗位只属于一个 HR。
- HR 只能修改、下架、删除自己创建的岗位。

### 3.2 用户与候选人档案

关系：

```text
users.id (candidate) 1 ---- 1 candidate_profiles.candidate_id
```

说明：

- 一个候选人最多一份结构化档案。
- `candidate_profiles.candidate_id` 同时作为主键和业务外键。
- 保存档案时使用 upsert，同一候选人重复保存会更新原档案。

### 3.3 用户与简历

关系：

```text
users.id (candidate) 1 ---- N resumes.candidate_id
```

说明：

- 一个候选人可以多次上传简历。
- 投递岗位时使用该候选人最新上传的简历记录。
- 简历源文件存储在私有 OSS，MySQL 只保存元信息。

### 3.4 岗位、候选人与投递

关系：

```text
jobs.id 1 ---- N applications.job_id
users.id (candidate) 1 ---- N applications.candidate_id
resumes.id 1 ---- N applications.resume_id
```

说明：

- 一条投递记录绑定一个岗位、一个候选人和一份简历。
- `uk_applications_job_candidate(job_id, candidate_id)` 保证同一候选人不能重复投递同一岗位。
- HR 查看投递台账时，系统通过 `applications.job_id -> jobs.id -> jobs.owner_hr_id` 过滤当前 HR 的岗位数据。

### 3.5 HR 与 AI 对话历史

关系：

```text
users.id (hr) 1 ---- N ai_chat_histories.hr_id
```

说明：

- 一个 HR 可以有多条 AI 问答历史。
- AI 历史按 `id` 或 `created_at` 时间顺序展示。
- 问答历史只绑定 HR，不绑定候选人。
- AI 回答基于当前 HR 自己岗位下的 MySQL 投递数据生成。

## 4. 典型查询场景

### 4.1 游客查看公开岗位

使用表：`jobs`

查询逻辑：

```sql
SELECT id, owner_hr_id, title, description, status, created_at
FROM jobs
WHERE status = 'open'
ORDER BY id;
```

### 4.2 候选人投递岗位

使用表：`users`、`jobs`、`candidate_profiles`、`resumes`、`applications`

校验流程：

1. 校验当前用户是候选人。
2. 查询岗位是否存在且 `status = open`。
3. 查询候选人是否已有完整档案。
4. 查询候选人最近一份合规简历。
5. 写入 `applications` 表。
6. 依赖唯一索引阻止重复投递。

### 4.3 HR 查看投递台账

使用表：`applications`、`jobs`、`users`、`candidate_profiles`、`resumes`

查询逻辑：

```sql
SELECT ...
FROM applications a
JOIN jobs j ON a.job_id = j.id
JOIN users u ON a.candidate_id = u.id
JOIN candidate_profiles p ON a.candidate_id = p.candidate_id
JOIN resumes r ON a.resume_id = r.id
WHERE j.owner_hr_id = ?
ORDER BY a.id
LIMIT ? OFFSET ?;
```

### 4.4 AI 问答统计业务数据

使用表：`applications`、`jobs`、`users`、`candidate_profiles`、`ai_chat_histories`

查询内容：

- 当前 HR 岗位下的投递总人数。
- 当前 HR 各岗位投递数量。
- 当前 HR 的热门岗位。
- 最近投递候选人的姓名、电话、学历、学校和技能。
- 最近 AI 问答历史上下文。

## 5. 数据安全设计

1. 密码只保存哈希值，不保存明文。
2. JWT 只保存用户 ID、角色和过期时间，不保存密码。
3. 简历源文件不落本地磁盘，上传后直接进入私有 OSS。
4. MySQL 中只保存简历对象 Key，HR 查看时动态生成签名 URL。
5. HR 只能管理本人岗位，只能查看本人岗位下的投递数据。
6. 候选人只能查看和修改自己的档案、简历和投递记录。
7. AI 历史按 HR ID 隔离，清空历史也只删除当前 HR 的记录。
