package ncm

import (
	"bytes"
	"crypto/aes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/bogem/id3v2"
	"github.com/go-flac/go-flac/v2"
)

// --- 全局常量 (私有) ---
var (
	coreKey   = []byte{0x68, 0x7A, 0x48, 0x52, 0x41, 0x6D, 0x73, 0x6F, 0x35, 0x6B, 0x49, 0x6E, 0x62, 0x61, 0x78, 0x57}
	modifyKey = []byte{0x23, 0x31, 0x34, 0x6C, 0x6A, 0x6B, 0x5F, 0x21, 0x5C, 0x5D, 0x26, 0x30, 0x55, 0x3C, 0x27, 0x28}
)

// Config 配置结构 (公开)
// 用于在调用时传递参数
type Config struct {
	InputDir           string // 输入目录路径
	OutputDirectory    string // 输出目录路径
	DeleteOnCompletion bool   // 是否删除原文件
	MaxWorkers         int    // 最大并发数 (0 表示自动检测)
}

// metaData 对应 NCM 内部 JSON 结构 (私有)
type metaData struct {
	MusicName string          `json:"musicName"`
	Album     string          `json:"album"`
	Artist    [][]interface{} `json:"artist"` // [["Name", id], ...]
	Format    string          `json:"format"`
}

// Convert 是包的入口函数 (公开)
// 接收 Config 配置并在后台执行转换
func Convert(cfg *Config) error {
	// 1. 参数校验与默认值处理
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = runtime.NumCPU()
	}
	if cfg.InputDir == "" {
		cfg.InputDir = "./"
	}
	if cfg.OutputDirectory == "" {
		cfg.OutputDirectory = "converted_music"
	}

	// 2. 创建输出目录
	if err := os.MkdirAll(cfg.OutputDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// 3. 读取输入目录
	files, err := os.ReadDir(cfg.InputDir)
	if err != nil {
		return fmt.Errorf("failed to read input directory: %v", err)
	}

	// 4. 筛选 .ncm 文件
	var ncmFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".ncm") {
			ncmFiles = append(ncmFiles, filepath.Join(cfg.InputDir, file.Name()))
		}
	}

	if len(ncmFiles) == 0 {
		return nil // 没有文件需要处理，直接返回
	}

	// 5. 并发处理
	var wg sync.WaitGroup
	sem := make(chan struct{}, cfg.MaxWorkers) // 信号量控制并发

	for _, file := range ncmFiles {
		wg.Add(1)
		sem <- struct{}{} // 获取令牌
		go func(fPath string) {
			processNCM(fPath, cfg, &wg, sem)
		}(file)
	}

	wg.Wait()
	return nil
}

// --- 以下均为私有辅助函数 ---

func aesECBDecrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	if len(data)%blockSize != 0 {
		return nil, fmt.Errorf("data length not a multiple of block size")
	}
	decrypted := make([]byte, len(data))
	for i := 0; i < len(data); i += blockSize {
		block.Decrypt(decrypted[i:i+blockSize], data[i:i+blockSize])
	}
	return decrypted, nil
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("invalid padding size")
	}
	unpadding := int(data[length-1])
	if unpadding > length {
		return nil, fmt.Errorf("invalid padding size")
	}
	return data[:(length - unpadding)], nil
}

func processNCM(ncmPath string, cfg *Config, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()
	defer func() { <-sem }() // 释放信号量

	f, err := os.Open(ncmPath)
	if err != nil {
		return
	}
	defer f.Close()

	header := make([]byte, 8)
	f.Read(header)
	if string(header[:4]) != "CTEN" || string(header[4:]) != "FDAM" {
		return
	}

	f.Seek(2, io.SeekCurrent)

	var keyLen uint32
	binary.Read(f, binary.LittleEndian, &keyLen)

	keyData := make([]byte, keyLen)
	f.Read(keyData)

	for i := range keyData {
		keyData[i] ^= 0x64
	}

	decryptedKey, _ := aesECBDecrypt(keyData, coreKey)
	musicKey, _ := pkcs7Unpad(decryptedKey)
	musicKey = musicKey[17:]

	keyBox := make([]byte, 256)
	for i := 0; i < 256; i++ {
		keyBox[i] = byte(i)
	}
	c := byte(0)
	lastByte := byte(0)
	keyOffset := 0
	for i := 0; i < 256; i++ {
		swap := keyBox[i]
		c = (swap + lastByte + musicKey[keyOffset]) & 0xff
		keyOffset = (keyOffset + 1) % len(musicKey)
		keyBox[i] = keyBox[c]
		keyBox[c] = swap
		lastByte = c
	}

	var metaLen uint32
	binary.Read(f, binary.LittleEndian, &metaLen)

	var meta metaData
	if metaLen > 0 {
		metaEncrypted := make([]byte, metaLen)
		f.Read(metaEncrypted)

		for i := range metaEncrypted {
			metaEncrypted[i] ^= 0x63
		}

		if len(metaEncrypted) > 22 {
			metaBase64 := metaEncrypted[22:]
			metaBytes, _ := base64.StdEncoding.DecodeString(string(metaBase64))
			metaJsonBytes, _ := aesECBDecrypt(metaBytes, modifyKey)
			metaJsonStrBytes, _ := pkcs7Unpad(metaJsonBytes)
			if len(metaJsonStrBytes) > 6 {
				jsonStr := string(metaJsonStrBytes[6:])
				json.Unmarshal([]byte(jsonStr), &meta)
			}
		}
	}

	f.Seek(5, io.SeekCurrent)

	var imageSpace uint32
	var imageSize uint32
	binary.Read(f, binary.LittleEndian, &imageSpace)
	binary.Read(f, binary.LittleEndian, &imageSize)

	var coverData []byte
	if imageSize > 0 {
		coverData = make([]byte, imageSize)
		f.Read(coverData)
	}

	if imageSpace > imageSize {
		f.Seek(int64(imageSpace-imageSize), io.SeekCurrent)
	}

	ext := ".mp3"
	if meta.Format == "flac" {
		ext = ".flac"
	}

	fileName := filepath.Base(ncmPath)
	outName := strings.TrimSuffix(fileName, filepath.Ext(fileName)) + ext
	outPath := filepath.Join(cfg.OutputDirectory, outName)

	outFile, err := os.Create(outPath)
	if err != nil {
		return
	}

	chunk := make([]byte, 0x8000)
	for {
		n, err := f.Read(chunk)
		if n > 0 {
			for i := 0; i < n; i++ {
				j := byte((i + 1) & 0xff)
				boxJ := keyBox[j]
				idx := (boxJ + keyBox[(boxJ+j)&0xff]) & 0xff
				chunk[i] ^= keyBox[idx]
			}
			outFile.Write(chunk[:n])
		}
		if err != nil {
			break
		}
	}
	outFile.Close()

	writeMetadata(outPath, ext, meta, coverData)

	if cfg.DeleteOnCompletion {
		f.Close()
		os.Remove(ncmPath)
	}
}

func writeMetadata(path, ext string, meta metaData, cover []byte) {
	artistName := "Unknown"
	if len(meta.Artist) > 0 && len(meta.Artist[0]) > 0 {
		if name, ok := meta.Artist[0][0].(string); ok {
			artistName = name
		}
	}

	if ext == ".mp3" {
		tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
		if err != nil {
			return
		}
		defer tag.Close()

		tag.SetTitle(meta.MusicName)
		tag.SetArtist(artistName)
		tag.SetAlbum(meta.Album)

		if len(cover) > 0 {
			mime := "image/jpeg"
			if bytes.HasPrefix(cover, []byte{0x89, 'P', 'N', 'G'}) {
				mime = "image/png"
			}
			pic := id3v2.PictureFrame{
				Encoding:    id3v2.EncodingUTF8,
				MimeType:    mime,
				PictureType: id3v2.PTFrontCover,
				Description: "Cover",
				Picture:     cover,
			}
			tag.AddAttachedPicture(pic)
		}
		tag.Save()

	} else if ext == ".flac" {
		f, err := flac.ParseFile(path)
		if err != nil {
			return
		}

		vcData := buildVorbisComment(meta.MusicName, artistName, meta.Album)
		vcBlock := &flac.MetaDataBlock{
			Type: flac.VorbisComment,
			Data: vcData,
		}
		f.Meta = append(f.Meta, vcBlock)

		if len(cover) > 0 {
			picData := buildFlacPicture(cover)
			picBlock := &flac.MetaDataBlock{
				Type: flac.Picture,
				Data: picData,
			}
			f.Meta = append(f.Meta, picBlock)
		}

		f.Save(path)
	}
}

func buildVorbisComment(title, artist, album string) []byte {
	buf := new(bytes.Buffer)
	vendor := "ncm_converter"
	binary.Write(buf, binary.LittleEndian, uint32(len(vendor)))
	buf.WriteString(vendor)
	comments := []string{
		"TITLE=" + title,
		"ARTIST=" + artist,
		"ALBUM=" + album,
	}
	binary.Write(buf, binary.LittleEndian, uint32(len(comments)))
	for _, c := range comments {
		binary.Write(buf, binary.LittleEndian, uint32(len(c)))
		buf.WriteString(c)
	}
	return buf.Bytes()
}

func buildFlacPicture(cover []byte) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(3))
	mime := "image/jpeg"
	if bytes.HasPrefix(cover, []byte{0x89, 'P', 'N', 'G'}) {
		mime = "image/png"
	}
	binary.Write(buf, binary.BigEndian, uint32(len(mime)))
	buf.WriteString(mime)
	desc := "Cover"
	binary.Write(buf, binary.BigEndian, uint32(len(desc)))
	buf.WriteString(desc)
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, uint32(0))
	binary.Write(buf, binary.BigEndian, uint32(len(cover)))
	buf.Write(cover)
	return buf.Bytes()
}