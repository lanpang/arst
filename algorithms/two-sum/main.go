package main

import "fmt"

func main() {
	nums := []int{2, 9, 0, 15}
	target := 9
	result := twoSum(nums, target)
	fmt.Println(result)
	fmt.Println("vim-go")
}

func twoSum(s []int, target int) []int {
	m := make(map[int]int)
	result := make([]int, 2)
	for i, v := range s {
		m[v] = i
	}
	for k, v := range m {
		if index, ok := m[target-k]; ok {
			result[0], result[1] = v, index
			return result
		}
	}
	return result
}
