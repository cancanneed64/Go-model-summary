package geminitrans

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// ReadCSV 读取CSV文件并返回所有行数据。
// 错误不会导致程序终止，而是通过返回值报告。
func ReadCSV(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("读取CSV数据失败: %w", err)
	}

	return rows, nil
}

// SearchData 搜索并提取数据，整合了预处理和矩阵解析流程。
func SearchData(st string, q string, c_if string, r_if string, cc string, rr string) (string, error) {
	
	var goal []string
	var (
		skipColHeader int // 是否带列序列
		hasRowHeader  int // 是否带行序列
		cols          int // 列
		rows          int // 行
	)

	// 预处理函数应该返回错误
	st, goal, skipColHeader, hasRowHeader, cols, rows, err := PreDeal(st, q, c_if, r_if, cc, rr)
	if err != nil {
		return "", fmt.Errorf("预处理参数失败: %w", err)
	}

	// 处理表格
	data := Dealdata(st, skipColHeader, hasRowHeader, cols, rows)
	
	ans := make([]string, 0, len(goal))

	for _, idx := range goal {
		if len(idx) < 2 {
			// 如果目标索引格式不对 (如只剩一个字符)
			log.Printf("警告: 目标索引 '%s' 格式不正确，跳过。", idx)
			continue
		}

		// 列索引 (m = 'A', 'B', ...)
		colRune := []rune(idx)[0]
		ci := int(colRune) - int('A')

		// 行索引 (n = '1', '2', ...)
		rowRune := []rune(idx)[1]
		n := int(rowRune) - int('0')
		ri := n - 1

		// 安全检查索引是否越界
		if ri >= len(data) || ri < 0 || ci >= len(data[0]) || ci < 0 {
			// 如果索引越界，则使用空字符串代替
			ans = append(ans, "")
			log.Printf("警告: 索引 [%s] 越界 (行:%d, 列:%d)，使用空值代替。", idx, ri, ci)
			continue
		}

		ans = append(ans, data[ri][ci])
	}

	revise_ans := strings.Join(ans, "")
	return revise_ans, nil
}

// PreDeal 预处理输入字符串和参数，返回清理后的数据。
func PreDeal(st string, q string, c_if string, r_if string, cc string, rr string) (string, []string, int, int, int, int, error) {
	// 表格预处理：只留下字符数字
	re := regexp.MustCompile(`\w`)
	re_int := regexp.MustCompile(`\d`)
	
	// 清理表格数据和查询字符串
	q = strings.Join(re.FindAllString(q, -1), "")
	st = strings.Join(re.FindAllString(st, -1), "")
	
	// 清理并转换参数
	var err error
	
	c_if = strings.Join(re_int.FindAllString(c_if, -1), "")
	new_c, err := strconv.Atoi(c_if)
	if err != nil && c_if != "" {
		return "", nil, 0, 0, 0, 0, fmt.Errorf("c_if 参数转换失败: %w", err)
	}

	r_if = strings.Join(re_int.FindAllString(r_if, -1), "")
	new_r, err := strconv.Atoi(r_if)
	if err != nil && r_if != "" {
		return "", nil, 0, 0, 0, 0, fmt.Errorf("r_if 参数转换失败: %w", err)
	}
	
	cc = strings.Join(re_int.FindAllString(cc, -1), "")
	new_cc, err := strconv.Atoi(cc)
	if err != nil && cc != "" {
		return "", nil, 0, 0, 0, 0, fmt.Errorf("cc 参数转换失败: %w", err)
	}

	rr = strings.Join(re_int.FindAllString(rr, -1), "") // 注意这里修正了原代码中错误的变量名 c_if
	new_rr, err := strconv.Atoi(rr)
	if err != nil && rr != "" {
		return "", nil, 0, 0, 0, 0, fmt.Errorf("rr 参数转换失败: %w", err)
	}

	// 查询转大写
	q = strings.ToUpper(q)
	
	// 按2个字符方式取表格 (如 "L7" -> ["L7"])
	goal := splitStringByInterval(q, 2)
	
	return st, goal, new_c, new_r, new_cc, new_rr, nil
}

// splitStringByInterval 将字符串按固定间隔切片
func splitStringByInterval(str string, interval int) []string {
	// 使用 runes 避免多字节字符导致的切片问题 (虽然当前场景只处理 \w，但习惯上更安全)
	runes := []rune(str)
	var data []string

	for i := 0; i < len(runes); i += interval {
		end := i + interval
		if end > len(runes) {
			end = len(runes)
		}
		data = append(data, string(runes[i:end]))
	}

	return data
}

// Dealdata 解析字符串为二维矩阵 (基于游标的安全版本)
func Dealdata(data string, skipColHeader int, hasRowHeader int, cols int, rows int) [][]string {

	// 1. 参数默认值处理
	if cols == 0 {
		cols = 12
	}
	if rows == 0 {
		rows = 7
	}

	// 2. 自动判断单元格位宽 (1位还是2位)
	note := 1
	if len(data) > (cols+1)*rows {
		note = 2
	}

	// 3. 定义游标 cursor，用于记录当前读取到的位置
	cursor := 0

	// 3.1 处理列头跳过
	if skipColHeader == 1 {
		if cursor+cols <= len(data) {
			cursor += cols
		} else {
			cursor = len(data)
		}
	}

	// 初始化结果集
	result := make([][]string, rows)

	// 4. 循环每一行
	for i := 0; i < rows; i++ {
		result[i] = make([]string, 0, cols)

		// 4.1 确定并安全跳过行首/分隔符
		skipCount := 0
		if hasRowHeader == 1 {
			skipCount = 1
		} else if i > 0 {
			skipCount = 1
		}

		if cursor+skipCount <= len(data) {
			cursor += skipCount
		} else {
			cursor = len(data)
		}

		// 4.2 提取当前行的有效数据字符串 (Extract & Pad)
		rowNeedLen := cols * note
		var rowStr string

		// 计算当前剩余数据长度
		remaining := len(data) - cursor
		
		if remaining <= 0 {
			// 数据已读完或游标越界，直接补 0
			rowStr = strings.Repeat("0", rowNeedLen)
		} else if remaining >= rowNeedLen {
			// 数据充足
			rowStr = data[cursor : cursor+rowNeedLen]
			cursor += rowNeedLen
		} else {
			// 数据不足，截取剩下的并补 0
			rowStr = data[cursor:]
			paddingCount := rowNeedLen - remaining
			rowStr += strings.Repeat("0", paddingCount)
			cursor = len(data) 
		}

		// 4.3 将行字符串切分为单元格 (Split)
		for j := 0; j < len(rowStr); j += note {
			// 避免最后一次切片因计算误差导致越界
			if j+note <= len(rowStr) {
				result[i] = append(result[i], rowStr[j:j+note])
			}
		}
	}

	// 移除调试打印
	return result
}