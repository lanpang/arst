package main

import "fmt"

var M = make(map[string]int)

func main() {
	fmt.Println("vim-go")
	s := "abgdefghilm"
	maxLen := lengthOfLongestSubstring(s)
	fmt.Printf("max substring length %d\n", maxLen)
}

func lengthOfLongestSubstring(s string) int {
	a := []byte(s)
	var s1 []byte
	var maxLen int
	i := 0
	j := 1
	var index int
	var is bool
	M[string(a[0])] = index
	for {
		if j == len(a) {
			if j-i > maxLen {
				maxLen = j - i
				s1 = a[i:j]
				fmt.Printf("len=%d,s1=%s\n", maxLen, string(s1))
			}
			break
		}
		k := string(a[j])
		if index, is = allUnique(k); is {
			M[k] = j
			j++
		} else {
			delete(M, k)
			if j-i > maxLen {
				maxLen = j - i
				s1 = a[i:j]
			}
			i = index + 1
		}
		fmt.Printf("len=%d,s1=%s\n", maxLen, string(s1))
	}
	return maxLen
}

func allUnique(k string) (int, bool) {
	if index, ok := M[k]; !ok {
		return -1, true
	} else {
		return index, false
	}
}
