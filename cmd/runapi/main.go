package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cheivin/go-runapi/pkg/config"
	"github.com/cheivin/go-runapi/pkg/generator"
	"github.com/cheivin/go-runapi/pkg/showdoc"
	"github.com/cheivin/go-runapi/pkg/types"
)

// Mode 运行模式
type Mode string

const (
	ModeGenerate     Mode = "generate" // 仅生成文档
	ModePush         Mode = "push"     // 仅推送文档
	ModeGeneratePush Mode = "genpush"  // 生成并推送变更文档
)

func main() {
	var (
		configFile string
		mode       string
		help       bool
		initConfig bool
	)

	flag.StringVar(&configFile, "config", "", "指定配置文件路径")
	flag.StringVar(&mode, "mode", "generate", "运行模式: generate(仅生成), push(仅推送), genpush(生成并推送)")
	flag.BoolVar(&help, "help", false, "显示帮助信息")
	flag.BoolVar(&initConfig, "init", false, "初始化配置文件")
	flag.Parse()

	if help {
		showHelp()
		return
	}

	if initConfig {
		initConfigFile()
		return
	}

	// 验证运行模式
	runMode := Mode(mode)
	if runMode != ModeGenerate && runMode != ModePush && runMode != ModeGeneratePush {
		fmt.Printf("错误: 无效的运行模式 '%s'\n", mode)
		showHelp()
		os.Exit(1)
	}

	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取当前目录失败: %v", err)
	}

	// 加载配置
	cfg, err := config.LoadConfig(currentDir, configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 验证根扫描目录
	if _, err := os.Stat(cfg.Scan.Dir); os.IsNotExist(err) {
		log.Fatalf("根扫描目录不存在: %s", cfg.Scan.Dir)
	}

	// 验证文档注释扫描目录
	if _, err := os.Stat(cfg.Scan.Scan); os.IsNotExist(err) {
		log.Fatalf("文档注释扫描目录不存在: %s", cfg.Scan.Scan)
	}

	fmt.Printf("当前目录: %s\n", currentDir)
	fmt.Printf("根扫描目录: %s\n", cfg.Scan.Dir)
	fmt.Printf("文档注释扫描目录: %s\n", cfg.Scan.Scan)
	if len(cfg.Scan.ExtraDirs) > 0 {
		fmt.Printf("额外扫描目录: %v\n", cfg.Scan.ExtraDirs)
	}
	fmt.Printf("输出文件: %s\n", cfg.Output.File)
	fmt.Printf("运行模式: %s\n", mode)

	// 创建文档生成器
	gen := generator.NewGenerator(cfg)

	switch runMode {
	case ModeGenerate:
		err = runGenerateMode(gen)
	case ModePush:
		err = runPushMode(gen, cfg)
	case ModeGeneratePush:
		err = runGeneratePushMode(gen, cfg)
	}

	if err != nil {
		log.Fatalf("执行失败: %v", err)
	}

	fmt.Println("执行完成")
}

// runGenerateMode 仅生成文档模式
func runGenerateMode(gen *generator.Generator) error {
	fmt.Println("\n=== 生成文档模式 ===")

	changed, err := gen.GenerateDocuments()
	if err != nil {
		return err
	}

	if changed {
		fmt.Println("文档已更新")
	} else {
		fmt.Println("文档无变化")
	}

	return nil
}

// runPushMode 仅推送模式
func runPushMode(gen *generator.Generator, cfg *config.Config) error {
	fmt.Println("\n=== 推送文档模式 ===")

	if !cfg.ShowDoc.Enabled {
		return fmt.Errorf("ShowDoc推送未启用，请在配置文件中设置 showdoc.enabled = true")
	}

	// 加载现有文档
	docs, err := gen.LoadExistingDocuments()
	if err != nil {
		return fmt.Errorf("加载现有文档失败: %v", err)
	}

	// 创建推送器
	pusher := showdoc.NewPusher(&cfg.ShowDoc)

	// 推送文档
	if err := pusher.PushDocuments(docs); err != nil {
		return err
	}

	return nil
}

// runGeneratePushMode 生成并推送模式
func runGeneratePushMode(gen *generator.Generator, cfg *config.Config) error {
	fmt.Println("\n=== 生成并推送模式 ===")

	// 1. 生成新文档
	newDocs, _, err := gen.GetGeneratedDocuments()
	if err != nil {
		return fmt.Errorf("生成新文档失败: %v", err)
	}

	// 2. 加载现有文档
	var oldDocs []types.APIDoc
	if _, err := os.Stat(cfg.Output.File); err == nil {
		oldDocs, err = gen.LoadExistingDocuments()
		if err != nil {
			return fmt.Errorf("加载现有文档失败: %v", err)
		}
	}

	// 3. 比较文档差异
	diff := gen.CompareDocuments(oldDocs, newDocs)
	if !diff.HasChanges() {
		fmt.Println("文档无变化，跳过生成和推送")
		return nil
	}

	fmt.Printf("检测到文档变更: %s\n", diff.GetSummary())

	// 4. 生成文档文件
	changed, err := gen.GenerateDocuments()
	if err != nil {
		return fmt.Errorf("生成文档失败: %v", err)
	}

	if !changed {
		fmt.Println("警告: 检测到变更但文件未更新")
	}

	// 5. 推送文档到ShowDoc（如果启用）
	if cfg.ShowDoc.Enabled {
		pusher := showdoc.NewPusher(&cfg.ShowDoc)
		if err := pusher.PushChangedDocuments(diff); err != nil {
			return fmt.Errorf("推送文档失败: %v", err)
		}
	} else {
		fmt.Println("ShowDoc推送未启用，仅生成本地文档")
	}

	return nil
}

// initConfigFile 初始化配置文件
func initConfigFile() {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("获取当前目录失败: %v", err)
	}

	configPath := filepath.Join(currentDir, "runapi.json")

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("配置文件已存在: %s\n", configPath)
		return
	}

	if err := config.CreateDefaultConfig(configPath); err != nil {
		log.Fatalf("创建配置文件失败: %v", err)
	}

	fmt.Printf("已创建默认配置文件: %s\n", configPath)
	fmt.Println("请根据需要修改配置文件中的参数")
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Println("runapi - Go代码注释API文档生成器")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  runapi [选项]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -config string  指定配置文件路径")
	fmt.Println("  -mode string    运行模式 (默认: generate)")
	fmt.Println("                  generate  - 仅生成文档")
	fmt.Println("                  push      - 仅推送文档到ShowDoc")
	fmt.Println("                  genpush   - 生成并推送变更文档")
	fmt.Println("  -init           初始化配置文件")
	fmt.Println("  -help           显示帮助信息")
	fmt.Println()
	fmt.Println("配置文件说明:")
	fmt.Println("  scan.dir        - 根扫描路径（用于结构体解析等）")
	fmt.Println("  scan.scan       - 带文档注释的文件扫描路径（可选，默认同dir）")
	fmt.Println("  scan.extra_dirs - 额外的扫描目录")
	fmt.Println()
	fmt.Println("配置文件查找顺序:")
	fmt.Println("  1. 当前运行目录的 runapi.json")
	fmt.Println("  2. 指定的配置文件路径")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  runapi                           # 使用默认配置生成文档")
	fmt.Println("  runapi -mode push                # 仅推送现有文档到ShowDoc")
	fmt.Println("  runapi -mode genpush             # 生成并推送变更文档")
	fmt.Println("  runapi -config ./custom.json     # 使用指定配置文件")
	fmt.Println("  runapi -init                     # 初始化配置文件")
}
