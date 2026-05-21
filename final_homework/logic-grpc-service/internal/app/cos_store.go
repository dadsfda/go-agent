package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

const (
	defaultCOSBucket = "ai-1418276225"
	defaultCOSRegion = "ap-guangzhou"
)

type COSConfig struct {
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	Endpoint  string `yaml:"endpoint"`
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
}

type COSStore struct {
	cfg    COSConfig
	client *cos.Client
}

func COSConfigFromEnv() (COSConfig, error) {
	cfg, err := ConfigFromFileAndEnv("")
	if err != nil {
		return COSConfig{}, err
	}
	return NormalizeCOSConfig(cfg.COS)
}

func NormalizeCOSConfig(cfg COSConfig) (COSConfig, error) {
	if cfg.Bucket == "" {
		cfg.Bucket = defaultCOSBucket
	}
	if cfg.Region == "" {
		cfg.Region = defaultCOSRegion
	}
	if cfg.Endpoint == "" && cfg.Bucket != "" && cfg.Region != "" {
		cfg.Endpoint = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region)
	}
	if strings.TrimSpace(cfg.SecretID) == "" || strings.TrimSpace(cfg.SecretKey) == "" {
		return COSConfig{}, errors.New("请在 config/config.yaml 或环境变量中配置 TENCENT_SECRET_ID 和 TENCENT_SECRET_KEY 后再上传简历到私有 COS")
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return COSConfig{}, errors.New("请配置 COS_ENDPOINT，或同时配置 COS_BUCKET 与 COS_REGION")
	}
	return cfg, nil
}

func NewCOSStoreFromEnv() (*COSStore, error) {
	cfg, err := COSConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return NewCOSStore(cfg)
}

func NewCOSStore(cfg COSConfig) (*COSStore, error) {
	bucketURL, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("COS_ENDPOINT 格式错误: %w", err)
	}
	baseURL := &cos.BaseURL{BucketURL: bucketURL}
	client := cos.NewClient(baseURL, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.SecretID,
			SecretKey: cfg.SecretKey,
		},
	})
	return &COSStore{cfg: cfg, client: client}, nil
}

func (s *COSStore) PutObject(ctx context.Context, key string, content []byte) error {
	_, err := s.client.Object.Put(ctx, key, bytes.NewReader(content), nil)
	return err
}

func (s *COSStore) SignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := s.client.Object.GetPresignedURL(ctx, http.MethodGet, key, s.cfg.SecretID, s.cfg.SecretKey, ttl, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
