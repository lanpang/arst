### 两数之和

#### 题目
给定一个整数数组 nums 和一个目标值 target，请你在该数组中找出和为目标值的那 两个 整数，并返回他们的数组下标。

你可以假设每种输入只会对应一个答案。但是，你不能重复利用这个数组中同样的元素。

**示例:**

给定 nums = [2, 7, 11, 15], target = 9

因为 nums[0] + nums[1] = 2 + 7 = 9
所以返回 [0, 1]

#### 分析

第一反应是使用map,将数组值存入map,key=array[index],value=index.时间复杂度O(n),空间复杂度O(n).
```
for key,value := range m {
  if index,ok := m[target -key]; ok {
    return value,index
  }
}
```

### 两数相加
#### 题目

给出两个 非空 的链表用来表示两个非负的整数。其中，它们各自的位数是按照 逆序 的方式存储的，并且它们的每个节点只能存储 一位 数字。

如果，我们将这两个数相加起来，则会返回一个新的链表来表示它们的和。

您可以假设除了数字 0 之外，这两个数都不会以 0 开头。

**示例：**
```
输入：(2 -> 4 -> 3) + (5 -> 6 -> 4)
输出：7 -> 0 -> 8
原因：342 + 465 = 807
```
#### 思路分析
需要两个指针，顺序遍历两个链表。两个链表的第n个元素和进位数，三个值一起相加。加起来得到的和 模10,是求和的值在该位的值，如果三个值的和加起来比9大，则通过三个值的和除以10计算进位。
当两条链表中的一个链表遍历完，则计算结束。输出和。
