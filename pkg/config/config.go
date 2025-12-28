package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 应用配置结构
type Config struct {
	// 扫描配置
	Scan ScanConfig `json:"scan"`

	// 输出配置
	Output OutputConfig `json:"output"`

	// ShowDoc配置
	ShowDoc ShowDocConfig `json:"showdoc"`
}

// ScanConfig 扫描配置
type ScanConfig struct {
	Dir           string   `json:"dir"`            // 根扫描路径（用于结构体解析等）
	Scan          string   `json:"scan"`           // 带文档注释的文件扫描路径（可选，默认同dir）
	ExtraDirs     []string `json:"extra_dirs"`     // 额外的扫描目录
	IncludeVendor bool     `json:"include_vendor"` // 是否包含vendor目录
}

// OutputConfig 输出配置
type OutputConfig struct {
	File string `json:"file"` // 输出文件路径
}

// ShowDocConfig ShowDoc配置
type ShowDocConfig struct {
	URL      string `json:"url"`       // ShowDoc基础URL
	APIKey   string `json:"api_key"`   // API密钥
	APIToken string `json:"api_token"` // API令牌
	Enabled  bool   `json:"enabled"`   // 是否启用ShowDoc推送
}

// LoadConfig 加载配置文件，支持多级覆盖
func LoadConfig(currentDir, configPath string) (*Config, error) {
	config := &Config{
		Scan: ScanConfig{
			Dir: ".",
		},
		Output: OutputConfig{
			File: "api-docs.json",
		},
		ShowDoc: ShowDocConfig{
			Enabled: false,
		},
	}

	// 1. 加载当前运行目录的配置文件
	if err := loadConfigFile(filepath.Join(currentDir, "runapi.json"), config); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("加载当前目录配置文件失败: %v", err)
		}
	}

	// 2. 加载指定的配置文件
	if configPath != "" {
		if err := loadConfigFile(configPath, config); err != nil {
			return nil, fmt.Errorf("加载指定配置文件失败: %v", err)
		}
	}

	// 3. 设置默认值
	// 如果dir路径没指定，则默认同当前运行路径
	if config.Scan.Dir == "" || config.Scan.Dir == "." {
		config.Scan.Dir = currentDir
	} else if !filepath.IsAbs(config.Scan.Dir) {
		// 如果是相对路径，基于当前目录
		config.Scan.Dir = filepath.Join(currentDir, config.Scan.Dir)
	}

	// 如果scan没指定，则默认同dir路径
	if config.Scan.Scan == "" {
		config.Scan.Scan = config.Scan.Dir
	} else if !filepath.IsAbs(config.Scan.Scan) {
		// 如果是相对路径，基于当前目录
		config.Scan.Scan = filepath.Join(currentDir, config.Scan.Scan)
	}

	// 处理额外目录
	for i, extraDir := range config.Scan.ExtraDirs {
		if !filepath.IsAbs(extraDir) {
			config.Scan.ExtraDirs[i] = filepath.Join(currentDir, extraDir)
		}
	}

	return config, nil
}

// loadConfigFile 加载单个配置文件
func loadConfigFile(filePath string, config *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// 创建临时配置来解析
	var tempConfig Config
	if err := json.Unmarshal(data, &tempConfig); err != nil {
		return fmt.Errorf("解析配置文件 %s 失败: %v", filePath, err)
	}

	// 合并配置，后者覆盖前者
	if tempConfig.Scan.Dir != "" {
		config.Scan.Dir = tempConfig.Scan.Dir
	}
	if tempConfig.Scan.Scan != "" {
		config.Scan.Scan = tempConfig.Scan.Scan
	}
	if tempConfig.Scan.ExtraDirs != nil {
		config.Scan.ExtraDirs = tempConfig.Scan.ExtraDirs
	}
	if tempConfig.Output.File != "" {
		config.Output.File = tempConfig.Output.File
	}
	if tempConfig.ShowDoc.URL != "" {
		config.ShowDoc.URL = tempConfig.ShowDoc.URL
	}
	if tempConfig.ShowDoc.APIKey != "" {
		config.ShowDoc.APIKey = tempConfig.ShowDoc.APIKey
	}
	if tempConfig.ShowDoc.APIToken != "" {
		config.ShowDoc.APIToken = tempConfig.ShowDoc.APIToken
	}
	// 布尔值直接覆盖
	config.ShowDoc.Enabled = tempConfig.ShowDoc.Enabled
	config.Scan.IncludeVendor = tempConfig.Scan.IncludeVendor

	return nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config, filePath string) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// CreateDefaultConfig 创建默认配置文件
func CreateDefaultConfig(filePath string) error {
	config := &Config{
		Scan: ScanConfig{
			Dir:           "./example",
			Scan:          "", // 将在LoadConfig中自动设置为同dir
			ExtraDirs:     []string{},
			IncludeVendor: false,
		},
		Output: OutputConfig{
			File: "api-docs.json",
		},
		ShowDoc: ShowDocConfig{
			URL:      "https://www.showdoc.cc/server/api/open",
			APIKey:   "",
			APIToken: "",
			Enabled:  false,
		},
	}

	return SaveConfig(config, filePath)
}
