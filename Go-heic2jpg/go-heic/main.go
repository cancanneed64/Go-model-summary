package main

import (
	"flag"
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/gen2brain/heic"
)

func main() {
	in := flag.String("in", ".", "HEIC 文件或目录")
	out := flag.String("out", "", "输出目录，默认与输入同目录")
	quality := flag.Int("q", 95, "JPEG 质量 1-100")
	jobs := flag.Int("j", 0, "并发数，0 表示 CPU 核数")
	deleteSource := flag.Bool("delete", false, "成功后删除源文件")
	flag.Parse()

	if *quality < 1 || *quality > 100 || *jobs < 0 {
		fmt.Fprintln(os.Stderr, "-q 必须在 1-100 之间，-j 不能为负数")
		os.Exit(2)
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
		case ".heic", ".heif", ".hif":
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
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	img, err := heic.Decode(f)
	if err != nil {
		return fmt.Errorf("解码失败: %w", err)
	}
	if img == nil || img.Bounds().Empty() {
		return fmt.Errorf("解码结果为空")
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	// 避免覆盖既有文件，失败时也不会留下损坏的 JPEG。
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	if err := jpeg.Encode(out, img, &jpeg.Options{Quality: quality}); err != nil {
		out.Close()
		os.Remove(dst)
		return fmt.Errorf("编码失败: %w", err)
	}
	if err := out.Close(); err != nil {
		os.Remove(dst)
		return err
	}
	return nil
}
