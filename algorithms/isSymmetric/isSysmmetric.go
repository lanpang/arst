package main

import "fmt"

type TreeNode struct {
	value int
	left  *TreeNode
	right *TreeNode
}

func main() {
	root := &TreeNode{}
	s := []int{1, 2, 2, 3, 3, 3, 3}
	root.value = s[0]
	initBinaryTree(root, 1, s)
	if isSymmetric(root) {
		fmt.Println("is symmetric")
	} else {
		fmt.Println("not is symmetric")
	}
	fmt.Println("vim-go")
}

func isSymmetric(root *TreeNode) bool {
	if root == nil {
		return true
	}
	if is := treeCompare(root.left, root.right); !is {
		return false
	}
	return true
}

func treeCompare(left, right *TreeNode) bool {
	if left == nil && right == nil {
		return true
	} else if left != nil && right != nil {
		if left.value == right.value {
			if is := treeCompare(left.right, right.left); !is {
				return false
			}
			if is := treeCompare(left.left, right.right); !is {
				return false
			}
			return true
		} else {
			return false
		}
	}
	return false

}

func initBinaryTree(root *TreeNode, index int, s []int) {
	if root == nil {
		return
	}
	if 2*index <= len(s) {
		root.left = &TreeNode{
			value: s[2*index-1],
		}
	}

	if 2*index+1 <= len(s) {
		root.right = &TreeNode{
			value: s[2*index],
		}
	}
	initBinaryTree(root.left, 2*index, s)
	initBinaryTree(root.right, 2*index+1, s)
}
