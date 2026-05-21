# 数据库设计文档

本系统采用 MySQL 存储双角色账号、岗位、候选人档案、简历元信息、投递记录和 AI 对话历史。Logic 服务启动时会自动创建以下数据表；本地单元测试仍保留内存仓库用于快速验证业务规则。

## 1. users

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT PRIMARY KEY AUTO_INCREMENT | 用户 ID |
| role | VARCHAR(20) NOT NULL | `hr` 或 `candidate` |
| email | VARCHAR(120) NOT NULL UNIQUE | 登录邮箱 |
| password_hash | VARCHAR(255) NOT NULL | 密码摘要 |
| created_at | DATETIME NOT NULL | 创建时间 |
| updated_at | DATETIME NOT NULL | 更新时间 |

## 2. jobs

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT PRIMARY KEY AUTO_INCREMENT | 岗位 ID |
| owner_hr_id | BIGINT NOT NULL | 创建岗位的 HR ID |
| title | VARCHAR(120) NOT NULL | 岗位名称 |
| description | TEXT NOT NULL | 岗位描述 |
| status | VARCHAR(20) NOT NULL | `open`、`closed` 或 `deleted` |
| created_at | DATETIME NOT NULL | 创建时间 |
| updated_at | DATETIME NOT NULL | 更新时间 |

索引：`idx_jobs_owner_hr_id(owner_hr_id)`。

## 3. candidate_profiles

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| candidate_id | BIGINT PRIMARY KEY | 候选人用户 ID |
| name | VARCHAR(80) NOT NULL | 姓名 |
| phone | VARCHAR(40) NOT NULL | 联系电话 |
| education | VARCHAR(80) NOT NULL | 最高学历 |
| school | VARCHAR(120) NOT NULL | 毕业院校 |
| experience | TEXT NOT NULL | 工作/项目经历 |
| skills | VARCHAR(255) NOT NULL | 核心技能标签 |
| created_at | DATETIME NOT NULL | 创建时间 |
| updated_at | DATETIME NOT NULL | 更新时间 |

## 4. resumes

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT PRIMARY KEY AUTO_INCREMENT | 简历记录 ID |
| candidate_id | BIGINT NOT NULL | 候选人用户 ID |
| file_name | VARCHAR(255) NOT NULL | 原始文件名 |
| file_ext | VARCHAR(20) NOT NULL | 文件扩展名 |
| oss_object_key | VARCHAR(500) NOT NULL | OSS 对象 Key |
| uploaded_at | DATETIME NOT NULL | 上传时间 |

索引：`idx_resumes_candidate_id(candidate_id)`。

## 5. applications

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT PRIMARY KEY AUTO_INCREMENT | 投递 ID |
| job_id | BIGINT NOT NULL | 岗位 ID |
| candidate_id | BIGINT NOT NULL | 候选人 ID |
| resume_id | BIGINT NOT NULL | 简历 ID |
| created_at | DATETIME NOT NULL | 投递时间 |

唯一约束：`uk_applications_job_candidate(job_id, candidate_id)`，防止重复投递。

## 6. ai_chat_histories

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | BIGINT PRIMARY KEY AUTO_INCREMENT | 消息 ID |
| hr_id | BIGINT NOT NULL | HR 用户 ID |
| question | TEXT NOT NULL | HR 提问 |
| answer | TEXT NOT NULL | AI 回复 |
| created_at | DATETIME NOT NULL | 创建时间 |

索引：`idx_ai_chat_histories_hr_id(hr_id)`。

## 7. 关系说明

1. `jobs.owner_hr_id` 关联 `users.id`，且用户角色必须为 HR。
2. `candidate_profiles.candidate_id` 关联 `users.id`，且用户角色必须为候选人。
3. `resumes.candidate_id` 关联 `users.id`。
4. `applications.job_id` 关联 `jobs.id`。
5. `applications.candidate_id` 关联 `users.id`。
6. `applications.resume_id` 关联 `resumes.id`。
7. `ai_chat_histories.hr_id` 关联 `users.id`，且用户角色必须为 HR。
