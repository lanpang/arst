### 题目
给定一个字符串，请你找出其中不含有重复字符的 最长子串 的长度。

**示例 1:**
```
输入: "abcabcbb"
输出: 3 
解释: 因为无重复字符的最长子串是 "abc"，所以其长度为 3。
```

**示例 2:**
```
输入: "bbbbb"
输出: 1
解释: 因为无重复字符的最长子串是 "b"，所以其长度为 1。
```

**示例 3:**

```
输入: "pwwkew"
输出: 3
解释: 因为无重复字符的最长子串是 "wke"，所以其长度为 3。
     请注意，你的答案必须是 子串 的长度，"pwke" 是一个子序列，不是子串。
```

### 题解
看到这个题目，没有任何思路。
看了一下解题思路，使用滑动窗口法，滑动窗口法的方法。找最长非重复子串，一定是连续的。
使用map判定滑动窗口的唯一性。
```
s[i],s[j]  
[i,j)是滑动窗口的内容,uniqueFunc是判定滑动窗口的函数，
if uniqueFunc(){
j++
}else{
  i++
}
```
