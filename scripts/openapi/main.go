package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
)

// cliConfig 描述命令行可配置的 OpenAPI 生成参数。
type cliConfig struct {
	RootDir string // 仓库根目录

	ModuleOutputDir string // 模块文档输出目录

	OverlayDir string // overlay 目录

	BundleOutputPath string // 聚合 bundle 输出文件路径

	BufBin string // buf 可执行文件路径
}

// main 负责解析命令行参数并执行 OpenAPI 生成流程。
func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// run 运行 OpenAPI 文档生成命令。
func run(ctx context.Context, args []string, stdout io.Writer, stderr io.Writer) error {
	fs := flag.NewFlagSet("openapi", flag.ContinueOnError)
	fs.SetOutput(stderr)

	cfg := cliConfig{}
	fs.StringVar(&cfg.RootDir, "root", ".", "仓库根目录")
	fs.StringVar(&cfg.ModuleOutputDir, "modules", "docs/openapi/modules", "模块文档输出目录")
	fs.StringVar(&cfg.OverlayDir, "overlays", "docs/openapi/overlays", "overlay 目录")
	fs.StringVar(&cfg.BundleOutputPath, "bundle", "docs/openapi/bundles/openapi.json", "bundle 输出路径")
	fs.StringVar(&cfg.BufBin, "buf", defaultBufBin(), "buf 可执行文件路径")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected arguments: %v", fs.Args())
	}

	generator, err := newGenerator(cfg)
	if err != nil {
		return err
	}
	result, err := generator.Generate(ctx)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(
		stdout,
		"generated %d module docs under %s and bundle %s\n",
		len(result.Modules),
		result.ModuleOutputDir,
		result.BundleOutputPath,
	)
	return err
}

// defaultBufBin 返回默认的 buf 可执行文件路径。
func defaultBufBin() string {
	if value := os.Getenv("BUF_BIN"); value != "" {
		return value
	}
	return "buf"
}
