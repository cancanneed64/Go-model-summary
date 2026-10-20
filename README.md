# Go-model-summary

一组 Go 格式转换与表格处理小工具。每个子项目使用独立的 `go.mod`，需要进入对应目录运行。

| 项目                                                                                    | 作用                                          | 状态   |
| --------------------------------------------------------------------------------------- | --------------------------------------------- | ------ |
| [Go-heic2jpg](Go-heic2jpg/)                                                             | HEIC/HEIF → JPG，提供外部工具和 Go 库两种实现 | 可用   |
| [Go-ncm2mp3](https://github.com/cancanneed64/Go-model-summary/tree/main/Go-ncm2mp3)     | 批量转换 `.ncm` 为 MP3/FLAC，并写入标签与封面 | 可用   |
| [Go-resizejpg](https://github.com/cancanneed64/Go-model-summary/tree/main/Go-resizejpg) | 图片缩放到 1920×1080                          | 可用   |
| [Go-ppt2jpg](https://github.com/cancanneed64/Go-model-summary/tree/main/Go-ppt2jpg)     | 读取和改写 `.pptx` 内部文件，不导出 JPG       | 实验性 |
| [Go-transgrid](https://github.com/cancanneed64/Go-model-summary/tree/main/Go-transgrid) | 从压缩成字符串的表格中按坐标取值              | 实验性 |

仓库包含五个子项目；`Go-heic2jpg` 的两种实现各自是独立模块。

## ncm2mp3

`Go-ncm2mp3/ncm.go`（package `ncm`）

批量转换目录下的 `.ncm` 文件：解析 NCM 头部 → AES-128-ECB 解出音频密钥 → RC4 类变体逐块解密音频流 → 根据元数据写 ID3v2（mp3）或 Vorbis Comment + Picture（flac）。

```go
import "ncm"

err := ncm.Convert(&ncm.Config{
    InputDir:           "./music",
    OutputDirectory:    "./out",
    DeleteOnCompletion: false,
    MaxWorkers:         0, // <=0 时取 runtime.NumCPU()
})
```

- `Convert` 并发处理（信号量控制并发数），**单个文件失败只会被静默跳过**，返回值只反映目录级错误。
- 格式取元数据里的 `format`：`flac` → `.flac`，其余 → `.mp3`。
- 示例调用见 `Go-ncm2mp3/test/main.go`。示例模块的 `replace ncm => ..\ncm` 是 Windows 路径写法，在 Linux/macOS 上需改成 `replace ncm => ..`。

## resizejpg

`Go-resizejpg/resizejpg.go`（package `refile`）

```go
err := refile.ResizeImage("in.png", "out.jpg")
```

基于 `disintegration/imaging`：强制拉伸到 1920×1080（会变形，原脚本同样如此），JPEG 质量 95。注释里保留了两种不变形方案：`imaging.Fit`（等比缩放留白）与 `imaging.Fill`（等比裁剪填满）。

## ppt2jpg

`Go-ppt2jpg/ppttojpg.go`（package `main`）

> 名字沿用自同目录的 Python 原版 `p.py`（通过 COM 调 PowerPoint 导出 JPG）。**Go 版并未实现导出 JPG**，做的是 pptx 内部结构读写。

```go
pp, err := ReadPowerPoint("new.pptx")
pp.ReplaceSlideContent("旧文本", "新文本", -1) // -1 = 全部替换
pp.ReplaceNotesSlideContent("旧", "新", -1)
pp.ReplaceThemeName("旧", "新", -1)
pp.ReplaceImage("logo.png", 1)          // 替换 ppt/media/image1.*
pp.DeletePassWord()                     // 移除 p:modifyVerifier（仅文本编辑密码）
pp.DeleteSlide(2)                       // 置空并移到末尾
pp.WriteToFile("out.pptx")
```

实现方式：`archive/zip` 读入，把 `ppt/slides/slide*`、`ppt/notesSlides/*`、`ppt/theme/*`、`ppt/media/image*`、`ppt/presentation.xml` 以字符串保存，改完再打包。

已知限制：

- 以字符串保存 XML，不做结构校验，替换内容需自行保证转义完整。
- `DeleteSlide` 不删除 `slide*.xml` 与 rels，只清空正文，故不会减少页数。
- 幻灯片内容以 map 保存，`GetSlidesContent()` 顺序不固定。
- `go.mod` 里的 `gooffice` 依赖实际未使用。

## transgrid

`Go-transgrid/`（package `trans`）与本目录下的 `gemini3version/`（package `geminitrans`）是同一功能的两个版本：把一份被压成单串字符的表格（例如考试/证件信息矩阵），按 `A1`、`B3` 这类坐标取值。

```go
ans := trans.SearchData(st, "A1", c_if, r_if, cc, rr)
// st   表格字符串（标点会被过滤，只保留 \w）
// q    查询坐标串，如 "A1B3"
// c_if 是否带列头序列  "1"/"0"
// r_if 是否带行头序列  "1"/"0"
// cc   列数（0 默认 12）
// rr   行数（0 默认 7）
```

单元格宽度自动判断：数据长度大于 `(cols+1)*rows` 时按 2 字符一格，否则 1 字符一格；数据不足时补 `0`。

两个版本的差异：

|          | `trans`（旧）                     | `geminitrans`（新）    |
| -------- | --------------------------------- | ---------------------- |
| 返回     | 仅字符串                          | 字符串 + `error`       |
| 参数解析 | 忽略 `Atoi` 错误，`rr` 变量名写错 | 校验并返回错误         |
| 越界查询 | 直接 panic                        | 返回空串并打日志       |
| 解析逻辑 | `goto` + 反复截断 `data`          | 游标 `cursor` 顺序读取 |

`test/main.go` 引用新版；其 `go.mod` 的 `replace` 写的是 Windows 路径，其他平台运行示例前需将 `geminitrans` 的替换路径改成 `../gemini3version`。

## heic2jpg

`Go-heic2jpg/` 提供两种独立模块：

| 实现     | 位置                                         | 依赖                                                                                                    |
| -------- | -------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| 外部工具 | [`heif-convert/`](Go-heic2jpg/heif-convert/) | Go 1.21+，系统安装 `heif-convert`（Debian/Ubuntu: `libheif-examples`；macOS: `brew install libheif`）   |
| Go 库    | [`go-heic/`](Go-heic2jpg/go-heic/)           | Go 1.25+，`github.com/gen2brain/heic`；无需 cgo，优先动态使用可用的 libheif，否则回退到内置 WASM 解码器 |

从仓库根目录进入一种实现目录（也可以选择 `Go-heic2jpg/heif-convert`）；两版参数相同：

```bash
cd Go-heic2jpg/go-heic
go run . -in ./photos -out ./jpg -q 90 -j 4
go run . -in ./photo.heic              # 输出同目录的 photo.jpg
go run . -in ./photos -delete           # 仅在成功转换后删除源文件
```

注意：`-in` 和 `-out` 相对于当前工作目录。默认保留源文件，也不会覆盖已有 JPG；输入目录只扫描当前一层，不递归。两版都支持 `.heic`、`.heif`、`.hif`，外部工具版还接受 `.avif`。Go 库版只解码一张图，不导出序列中的全部帧；不能保证保留 EXIF。不同扩展名但同名的输入可能映射到同一个 JPG，批量转换时请避免这种情况。默认仅生成 JPEG，不做尺寸调整。

## 构建

各子目录独立构建。Go 库版从仓库根目录执行：

```bash
cd Go-heic2jpg/go-heic && CGO_ENABLED=0 go build -o heic2jpg .
```

外部工具版在 `Go-heic2jpg/heif-convert` 目录执行 `go build -o heic2jpg .`，运行时仍需安装 `heif-convert`。

`Go-ppt2jpg` 的示例可在该目录运行 `go run ppttojpg.go`，默认读取 `new.pptx` 并打印幻灯片 XML。
