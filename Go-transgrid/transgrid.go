package trans

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// 读取CSV文件并获取指定列和行的数据
func ReadCSV(filename string) [][]string {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	//fmt.Println(rows)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(rows)
	return rows
}


func SearchData(st string, q string, c_if string, r_if string, cc string, rr string) string {

	//只留下字符数字并将列表结果转为字符串
	var goal []string
	var (
		new_c  int //是否带列序列
		new_r  int //是否带行序列
		new_cc int //列
		new_rr int //行
	)

	st, goal, new_c, new_r, new_cc, new_rr = PreDeal(st, q, c_if, r_if, cc, rr)

	//处理表格
	data := Dealdata(st, new_c, new_r, new_cc, new_rr)
	//fmt.Println(data)
	ans := []string{}

	for _, idx := range goal {
		//列
		m := string(idx[0])
		ci := int([]rune(m)[0]) - int('A')

		//行
		n := int(idx[1]) - int('0')
		ri := n - 1

		//fmt.Println(ri, ci, string(data[ri][ci]))

		ans = append(ans, string(data[ri][ci]))
	}

	revise_ans := strings.Join(ans, "")
	//fmt.Println(ans)
	return revise_ans

}

func PreDeal(st string, q string, c_if string, r_if string, cc string, rr string) (string, []string, int, int, int, int) {
	//表格预处理
	//只留下字符数字并将列表结果转为字符串
	re := regexp.MustCompile(`\w`)
	re_int := regexp.MustCompile(`\d`)
	//fmt.Println(st, q)
	//表格格式塌缩变成字符串，-1代表所有匹配数量
	q = strings.Join(re.FindAllString(q, -1), "")
	st = strings.Join(re.FindAllString(st, -1), "")
	//fmt.Println(st, q)
	c_if = strings.Join(re_int.FindAllString(c_if, -1), "")
	new_c, _ := strconv.Atoi(c_if)

	r_if = strings.Join(re_int.FindAllString(r_if, -1), "")
	new_r, _ := strconv.Atoi(r_if)
	cc = strings.Join(re_int.FindAllString(cc, -1), "")
	new_cc, _ := strconv.Atoi(cc)
	c_if = strings.Join(re_int.FindAllString(rr, -1), "")
	new_rr, _ := strconv.Atoi(rr)

	//fmt.Println(new_c, new_r, new_cc, new_rr)
	//fmt.Printf("%v\n", new_c)
	//查询转大写
	q = strings.ToUpper(q)
	//按2个字符方式取表格 L7
	goal := splitStringByInterval(q, 2)
	//goal [L7]
	//fmt.Println(goal,q)
	//fmt.Println(st, goal, new_c, new_r, new_cc, new_rr)
	return st, goal, new_c, new_r, new_cc, new_rr
}

func splitStringByInterval(str string, interval int) []string {
	var data []string

	for i := 0; i < len(str); i += interval {
		end := i + interval
		if end > len(str) {
			end = len(str)
		}
		data = append(data, str[i:end])
		
	}

	return data
}
// 去除空格定义表格，进行取样
func Dealdata(data string, c_if int, r_if int, cc int, rr int) [][]string {

	//c_if，r_if判断是否带序列，cc，rr为列和行
	//fmt.Println(c_if, r_if, cc, rr)
	//判断行列是否默认为空，空则按12和7计算
	if cc == 0 {
		cc = 12
	}
	if rr == 0 {
		rr = 7
	}
	//待定
	/*
		if cc >= 20 || rr >= 20 {
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
				}
			}()
			l := make([][]string, 1)
			return l
		}
	*/

	if c_if == 1 {
		//跳过前缀行
		data = data[cc:]
		//fmt.Println(len(data), data)
	}
	//1 VQ R0 C7 PM QM 1H T2 F4 YN Q4 09 TW...
	//通过note判断1个字符还是2个字符
	var note int
	//是否存在数据不足的情况，补0
	var ex_note int = 0

	if len(data) > (cc+1)*rr {
		note = 2
	} else {
		note = 1
	}
	//fmt.Println(string(data[1:13]))
	//data := strings.Fields(data)
	//fmt.Println(data[0])
	//fmt.Println("PASS")
	var sl, remain_str string
	l := make([][]string, rr)
	for i := 0; i < rr; i++ {
		//跳过行首序列
		//先存到字符串sl中，再录入列表l,得到的sl已经去掉了表头
		if r_if == 1 {
			if ex_note == 1 {

				remain_str = strings.Repeat("0", cc*note)
				sl = remain_str
				
				goto label
			}
			//第一行直接取
			if i == 0 {
				//fmt.Println(data)
				sl = data[1 : cc*note+1]
				//fmt.Println(sl)
			} else {
				//第二行开始data往后移
				//当最后一行遍历时直接取余下的数
				//
				fmt.Println(len(data))
				//fmt.Println(cc*note, len(data))
				//fmt.Println(data[cc*note+1:])
				data = data[cc*note+1:]

				if len(data) > cc*note {

					sl = data[1 : cc*note+1]
					fmt.Println(sl)
				} else {

					sl = data[1:]
					//remain_str = "00"
					remain_str = strings.Repeat("0", cc*note-len(data)+1)
					sl += remain_str
					ex_note = 1
					fmt.Println(sl)
				}
			}
			//当表格大于输入行列，余下数字不取进表格中，照常索引
			//当表格小于输入行列，余下空余用空替代，照常索引
			//fmt.Println(sl)
			//if len(sl)//cc > rr
			//sl  [PD 8H HE 81 H8 YN VX EV XF K9 D8]
			//n=6,6+2<23,n=n+2
		label:
			for n := 0; n+note < len(sl)+1; n = n + note {
				//fmt.Println(n, len(sl))
				l[i] = append(l[i], sl[n:n+note])
				//fmt.Println(l[i])
			}
			//fmt.Println(len(data), cc*note)
			//fmt.Println(l)

		} else {
			if ex_note == 1 {

				remain_str = strings.Repeat("0", cc*note)
				sl = remain_str
				goto end
			}
			//第一行直接取
			if i == 0 {
				//fmt.Println(data)
				sl = data[:cc*note]
				//fmt.Println(sl)
			} else {
				//第二行开始data往后移
				//当最后一行遍历时直接取余下的数

				fmt.Println(len(data))

				data = data[cc*note+1:]

				if len(data) > cc*note {

					sl = data[:cc*note]
					fmt.Println(sl)
				} else {

					sl = data
					//remain_str = "00"
					remain_str = strings.Repeat("0", cc*note-len(data)+1)
					sl += remain_str
					ex_note = 1
					fmt.Println(sl)
				}
			}

		end:
		
			for n := 0; n+note < len(sl)+1; n = n + note {
				//fmt.Println(n, len(sl))
				l[i] = append(l[i], sl[n:n+note])
				//fmt.Println(l[i])
			}
			/*
				if len(data) > cc*note {
					sl = data[:cc*note]

				} else if len(data) <= cc*note {
					sl = data
				}
				//fmt.Println(sl)
				for n := 0; n+note < len(sl)+1; n = n + note {
					l[i] = append(l[i], sl[n:n+note])
				}
				//fmt.Println(l)
				if len(data) > cc*note {
					data = data[cc*note:]
				}
			*/
		}
	}
	//fmt.Printf("%#v\n", l)
	fmt.Println(l)
	return l
}

/*

func splitStringByInterval(str string, interval int) []string {
    // 1. 先转为 rune 切片（按"字符"算，而不是按"字节"算）
    runes := []rune(str) 
    var data []string

    // 2. 这里的 len(runes) 就是字符个数了
    for i := 0; i < len(runes); i += interval {
        end := i + interval
        if end > len(runes) {
            end = len(runes)
        }
        // 3. 截取 rune 切片，再转回 string
        data = append(data, string(runes[i:end]))
    }

    return data
}
	*/