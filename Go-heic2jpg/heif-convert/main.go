package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func main() {
	in := flag.String("in", ".", "HEIC 文件或目录")
	out := flag.String("out", "", "输出目录，默认与输入同目录")
	quality := flag.Int("q", 95, "JPEG 质量 0-100")
	jobs := flag.Int("j", 0, "并发数，0 表示 CPU 核数")
	deleteSource := flag.Bool("delete", false, "成功后删除源文件")
	flag.Parse()

	if *quality < 0 || *quality > 100 || *jobs < 0 {
		fmt.Fprintln(os.Stderr, "-q 必须在 0-100 之间，-j 不能为负数")
		os.Exit(2)
	}
	if _, err := exec.LookPath("heif-convert"); err != nil {
		fmt.Fprintln(os.Stderr, "未找到 heif-convert；请安装 libheif-examples (Linux) 或 libheif (macOS)")
		os.Exit(1)
	}
	files, err := collect(*in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("没有找到 HEIC/HEIF 文件")
		return
	}

	n := *jobs
	if n == 0 {
		n = runtime.NumCPU()
	}
	sem := make(chan struct{}, n)
	var wg sync.WaitGroup
	var mu sync.Mutex
	failed := 0
	for _, src := range files {
		sem <- struct{}{}
		wg.Add(1)
		go func(src string) {
			defer wg.Done()
			defer func() { <-sem }()
			dst := destination(src, *out)
			err := convert(src, dst, *quality)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failed++
				fmt.Fprintf(os.Stderr, "失败 %s: %v\n", src, err)
				return
			}
			fmt.Printf("完成 %s -> %s\n", src, dst)
			if *deleteSource {
				if err := os.Remove(src); err != nil {
					failed++
					fmt.Fprintf(os.Stderr, "无法删除 %s: %v\n", src, err)
				}
			}
		}(src)
	}
	wg.Wait()
	if failed != 0 {
		os.Exit(1)
	}
}

func collect(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(entry.Name())) {
		case ".heic", ".heif", ".hif", ".avif":
			files = append(files, filepath.Join(path, entry.Name()))
		}
	}
	return files, nil
}

func destination(src, dir string) string {
	if dir == "" {
		dir = filepath.Dir(src)
	}
	return filepath.Join(dir, strings.TrimSuffix(filepath.Base(src), filepath.Ext(src))+".jpg")
}

func convert(src, dst string, quality int) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("输出文件已存在: %s", dst)
	} else if !os.IsNotExist(err) {
		return err
	}
	cmd := exec.Command("heif-convert", "-q", fmt.Sprint(quality), src, dst)
	if output, err := cmd.CombinedOutput(); err != nil {
		os.Remove(dst)
		return fmt.Errorf("heif-convert: %w: %s", err, strings.TrimSpace(string(output)))
	}
	if info, err := os.Stat(dst); err != nil || info.Size() == 0 {
		os.Remove(dst)
		return fmt.Errorf("输出文件为空或不存在: %s", dst)
	}
	return nil
}
