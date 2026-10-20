package refile

import (
	"fmt"

	"github.com/disintegration/imaging"
)

/*
func Resize_jpg(inpath string, outpath string) {
	file, err := os.Open(inpath)
	if err != nil {
		fmt.Println("无法打开图像文件:", err)
		return
	}
	defer file.Close()

	// 解码JPEG图像
	img, err := jpeg.Decode(file)
	if err != nil {
		fmt.Println("无法解码JPEG图像:", err)
		return
	}

	// 将图像调整为1024x960的大小
	resizedImg := imaging.Resize(img, 1920, 1080, imaging.Lanczos)

	// 创建输出文件
	outputFile, err := os.Create(outpath)
	if err != nil {
		fmt.Println("无法创建输出文件:", err)
		return
	}
	defer outputFile.Close()

	// 将调整后的图像保存为JPEG格式
	err = jpeg.Encode(outputFile, resizedImg, nil)
	if err != nil {
		fmt.Println("无法保存图像:", err)
		return
	}

	fmt.Println("图像转换完成，已保存为output.jpg")

}
*/

// ResizeImage 调整图片大小并保存为 JPG
// 优化点：
// 1. 返回 error，让上层函数知道是否成功
// 2. 支持多种输入格式 (jpg, png, gif 等)
// 3. 使用 imaging 库的高级封装，大幅减少代码量
func ResizeImage(inPath string, outPath string) error {
	// 1. 打开图片
	// imaging.Open 会自动检测格式 (jpg/png/gif) 并处理 EXIF 旋转问题
	src, err := imaging.Open(inPath)
	if err != nil {
		return fmt.Errorf("无法打开或解码图片: %w", err)
	}

	// 2. 调整大小
	// 原始逻辑：强制拉伸到 1920x1080 (如果原图比例不同，图片会变形)
	dst := imaging.Resize(src, 1920, 1080, imaging.Lanczos)
	
	// --- 进阶选项 (根据需求二选一) ---
	// 选项 A: 保持比例缩放，不超出 1920x1080 (不会变形，但可能填不满)
	// dst := imaging.Fit(src, 1920, 1080, imaging.Lanczos)
	
	// 选项 B: 保持比例裁剪，填满 1920x1080 (不会变形，但边缘会被切掉)
	// dst := imaging.Fill(src, 1920, 1080, imaging.Center, imaging.Lanczos)
	// -------------------------------

	// 3. 保存图片
	// imaging.Save 会根据 outPath 的后缀名自动选择编码格式
	// 这里我们可以指定 JPEGQuality 来控制压缩质量 (默认 95，建议 80 以减小体积)
	err = imaging.Save(dst, outPath, imaging.JPEGQuality(95))
	if err != nil {
		return fmt.Errorf("无法保存图片: %w", err)
	}

	// 成功，返回 nil
	return nil
}