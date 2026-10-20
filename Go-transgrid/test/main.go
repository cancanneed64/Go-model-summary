package main

import (
	"fmt"
	"geminitrans"
)

func main() {
	//q := "B3a7C4A ^$%$5A  6"
	q := "A1"
	// B3 值为E1 带行列

	st := `
	1 VQ R0 C7 PM QM 1H T2 F4 YN Q4 09 TW 2 4C NK E1 20 XC 1X 6N 7R TW RV 07 0T 3 Q9 E1 5V 7N 42 FX JX VX JE QC 7T 54 4 MW PV D7 X7 KV 6V 29 F8 YP MX VR M7 5 P2 2R 2M 25 29 29 HQ N2 NK KM QE 29 6 5K V8 HF P0 XR FT KW 72 KM WK 29 YC 7 PD 8H HE 81 H8`
	
/*
	st := `
	A
	B
	C
	D
	E
	F
	G
	H I J
	K
	L
	1 VQ R0 C7 PM QM 1H T2 F4 YN Q4 09 TW 2 4C NK E1 20 XC 1X 6N 7R TW RV 07 0T 3 Q9 E1 5V 7N 42 FX JX VX JE QC 7T 54 4 MW PV D7 X7 KV 6V 29 F8 YP MX VR M7 5 P2 2R 2M 25 29 29 HQ N2 NK KM QE 29 6 5K V8 HF P0 XR FT KW 72 KM WK 29 YC 7 PD 8H HE 81 H8`
	
	st := `A B C D E F G H I J K L
		1 6 N K J V 9 9 Q N 9 V 8
		2 Q V F H Y T Y D D 4 W 5
		3 M K 6 8 5 4 0 W 8 C 8 9
		4 2 X E H T 3 7 M 2 H D D
		5 P R C Q 7 4 2 9 J R 8 W
		6 C 5 Y Y W 8 C T D 5 1 K
		7 D K J R 0 9 N 5 M F 7 Y`
		*/
	//fmt.Println(q, st)
	// c_if int, r_if int, cc int, rr int
	//fmt.Println(trans.SearchData(st, q, "1", "1", "0", "0"))
	fmt.Println(geminitrans.SearchData(st, q, "0", "1", "0", "0"))
}

//不带列带行
/*
	//不带行不带列
	st := `
	[[VQ R0 C7 PM QM 1H T2 F4 YN Q4 09 TW] [4C NK E1 20 XC 1X 6N 7R TW RV 07 0T] [Q9 E1 5V 7N 42 FX JX VX JE QC 7T 54] [MW PV D7 X7 KV 6V 29 F8 YP MX VR M7] [P2 2R 2M 25 29 29 HQ N2 NK KM QE 29] [5K V8 HF P0 XR FT KW 72 KM WK 29 YC] [PD 8H HE 81 H8 YN VX EV XF K9 D8]]
	`
*/
//validID := regexp.MustCompile(`\w`)
//trans.SearchData(st, q, 0, 1)

//[[VQ R0 C7 PM QM 1H T2 F4 YN Q4 09 TW] [4C NK E1 20 XC 1X 6N 7R TW RV 07 0T] [Q9 E1 5V 7N 42 FX JX VX JE QC 7T 54] [MW PV D7 X7 KV 6V 29 F8 YP MX VR M7] [P2 2R 2M 25 29 29 HQ N2 NK KM QE 29] [5K V8 HF P0 XR FT KW 72 KM WK 29 YC] [PD 8H HE 81 H8 YN VX EV XF K9 D8]]
