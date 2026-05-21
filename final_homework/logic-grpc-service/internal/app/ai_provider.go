package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

type AIConfig struct {
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	BaseURL string `yaml:"base_url"`
}

type ChatMessage struct {
	Role    string
	Content string
}

type AIProvider interface {
	Answer(ctx context.Context, businessContext, question string, history []ChatMessage) (string, error)
	ParseFields(ctx context.Context, resumeText string) (Profile, error)
}

type EinoAIProvider struct {
	model *openai.ChatModel
}

func NormalizeAIConfig(cfg AIConfig) (AIConfig, error) {
	if cfg.Model == "" {
		cfg.Model = "deepseek-v4-flash"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com"
	}
	if strings.TrimSpace(cfg.APIKey) == "" {
		return AIConfig{}, errors.New("请在 config/config.yaml 或 DEEPSEEK_API_KEY 中配置 DeepSeek API Key 后再使用 Eino AI 问答")
	}
	return cfg, nil
}

func NewEinoAIProvider(ctx context.Context, cfg AIConfig) (*EinoAIProvider, error) {
	cfg, err := NormalizeAIConfig(cfg)
	if err != nil {
		return nil, err
	}
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		BaseURL: cfg.BaseURL,
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &EinoAIProvider{model: chatModel}, nil
}

func (p *EinoAIProvider) Answer(ctx context.Context, businessContext, question string, history []ChatMessage) (string, error) {
	messages := []*schema.Message{
		{
			Role: schema.System,
			Content: fmt.Sprintf(`你是智能招聘系统的 HR 数据助手。你必须只根据给定业务上下文回答，不能编造不存在的数据。
如果用户问某类数据，从业务上下文中查找并列出具体信息（如候选人姓名、电话、技能等）。
回答要简洁、中文、适合直接展示给 HR。

业务上下文：
%s`, businessContext),
		},
	}

	// 加入历史对话（最近 5 轮）
	for _, msg := range history {
		role := schema.User
		if msg.Role == "assistant" {
			role = schema.Assistant
		}
		messages = append(messages, &schema.Message{Role: role, Content: msg.Content})
	}

	// 当前问题
	messages = append(messages, &schema.Message{Role: schema.User, Content: question})

	response, err := p.model.Generate(ctx, messages)
	if err != nil {
		return "", err
	}
	if response == nil || strings.TrimSpace(response.Content) == "" {
		return "", errors.New("Eino 模型返回为空")
	}
	return response.Content, nil
}

func (p *EinoAIProvider) ParseFields(ctx context.Context, resumeText string) (Profile, error) {
	if strings.TrimSpace(resumeText) == "" {
		return Profile{}, errors.New("简历内容为空，无法解析")
	}

	response, err := p.model.Generate(ctx, []*schema.Message{
		{
			Role: schema.System,
			Content: `你是一个简历信息提取助手。从给定的简历文本中提取以下 6 个字段，严格以 JSON 格式返回，不要添加任何其他文字：
{
  "name": "姓名",
  "phone": "联系电话",
  "education": "最高学历（如本科、硕士、博士）",
  "school": "毕业院校",
  "experience": "工作或项目经历，尽量完整保留简历原文，不要总结、不要概括、不要压缩，多段经历用换行分隔",
  "skills": "核心技能标签（逗号分隔）"
}
如果某个字段无法从简历中提取，填空字符串。`,
		},
		{
			Role:    schema.User,
			Content: fmt.Sprintf("以下是简历内容：\n\n%s", resumeText),
		},
	})
	if err != nil {
		return Profile{}, fmt.Errorf("AI 解析简历失败: %w", err)
	}
	if response == nil || strings.TrimSpace(response.Content) == "" {
		return Profile{}, errors.New("AI 返回为空")
	}

	content := response.Content
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || end <= start {
		return Profile{}, errors.New("AI 返回格式异常")
	}
	content = content[start : end+1]

	var fields struct {
		Name       string `json:"name"`
		Phone      string `json:"phone"`
		Education  string `json:"education"`
		School     string `json:"school"`
		Experience string `json:"experience"`
		Skills     string `json:"skills"`
	}
	if err := json.Unmarshal([]byte(content), &fields); err != nil {
		return Profile{}, fmt.Errorf("AI 返回 JSON 解析失败: %w", err)
	}

	profile := Profile{
		Name:       strings.TrimSpace(fields.Name),
		Phone:      strings.TrimSpace(fields.Phone),
		Education:  strings.TrimSpace(fields.Education),
		School:     strings.TrimSpace(fields.School),
		Experience: strings.TrimSpace(fields.Experience),
		Skills:     strings.TrimSpace(fields.Skills),
	}
	if !hasExtractedResumeFields(profile) {
		return Profile{}, errors.New("AI 未能从简历中提取到有效字段")
	}
	return profile, nil
}

func hasExtractedResumeFields(profile Profile) bool {
	return strings.TrimSpace(profile.Name) != "" ||
		strings.TrimSpace(profile.Phone) != "" ||
		strings.TrimSpace(profile.Education) != "" ||
		strings.TrimSpace(profile.School) != "" ||
		strings.TrimSpace(profile.Experience) != "" ||
		strings.TrimSpace(profile.Skills) != ""
}
