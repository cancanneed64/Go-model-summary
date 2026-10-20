package main

import (
	"fmt"
	"ncm" // 引入你的包，路径取决于你的 go.mod 定义
	"runtime"
)

func main() {
	// 1. 设置配置
	config := ncm.Config{
		InputDir:           "../ncm",     // 输入目录
		OutputDirectory:    "../ncm/converted_music", // 输出目录
		DeleteOnCompletion: false,             // 是否删除源文件
		MaxWorkers:         runtime.NumCPU(),  // 并发数
	}

	fmt.Println("正在开始转换...")

	// 2. 调用转换函数
	err := ncm.Convert(&config)
	
	if err != nil {
		fmt.Printf("转换过程中发生错误: %v\n", err)
	} else {
		fmt.Println("转换完成！")
	}
}