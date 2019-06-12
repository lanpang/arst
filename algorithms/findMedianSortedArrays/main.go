package main

import "fmt"

func main() {
	fmt.Println("vim-go")
	nums1 := []int{3, 4}
	nums2 := []int{-2, -1}
	median := findMedianSortedArrays(nums1, nums2)
	fmt.Printf("%2f\n", median)
}

func findMedianSortedArrays(nums1 []int, nums2 []int) float64 {
	i := 0
	j := 0
	len1 := len(nums1)
	len2 := len(nums2)
	var m int
	m = (len1 + len2) / 2

	for {
		if i < len1 && j < len2 {
			if (i + j) == m {
				if (len1+len2)%2 == 1 {
					if nums1[i] > nums2[j] {
						fmt.Println("aa", j)
						return float64(nums2[j]) / 1.0
					} else {
						fmt.Println("bb")
						return float64(nums1[i]) / 1.0
					}
				} else {
					return float64(nums1[i]+nums2[j]) / 2.0
				}
				break
			}
			if nums1[i] < nums2[j] {
				i++
			} else {
				j++
			}
		} else {
			if i == len1 {
				if (len1+len2)%2 == 1 {
					m = (len1 + len2) / 2
					return float64(nums2[m-i]) / 1.0
				} else {
					m = (len1 + len2) / 2
					return float64(nums2[m-i]+nums2[m-i-1]) / 2.0
				}
			}
			if j == len2 {
				if (len1+len2)%2 == 1 {
					m = (len1 + len2) / 2
					return float64(nums1[m-i]) / 1.0
				} else {
					m = (len1 + len2) / 2
					return float64(nums1[m-i]+nums1[m-i-1]) / 2.0
				}

			}
			break
		}
	}
	return 0.0
}
