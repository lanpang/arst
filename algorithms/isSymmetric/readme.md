#### 思路分析

#### 1 题目
例如，二叉树 `[1,2,2,3,4,4,3]` 是对称的。
```
    1
   / \
  2   2
 / \ / \
3  4 4  3
```

但是下面这个 `[1,2,2,null,3,null,3]` 则不是镜像对称的:

```
   1
   / \
  2   2
   \   \
   3    3
```

**说明:**

如果你可以运用递归和迭代两种方法解决这个问题，会很加分。


**分析：**

我最初看到这个问题时，没有思路，觉得可以通过栈的方式，使用先序，后序，中序遍历的方式，可以找到某种规律，然后得出结论。
这个思路我没有继续想下去，首先时间复杂度太高，其次也不一定能成功。

树本来就是递归定义的，如果能使用递归解决这个问题时间复杂度可能会少一点。（递归的时间复杂度应该是O(n)）

**递归的思路：**
递归要有一个开始，一个结束条件，一个不变式。
我认为开始条件应给是从root节点开始，如果root->left 和root->right相等，则继续迭代。
结束条件是，到一个空子树，该方向的递归结束。
不变子式，是什么呢？

经过分析，发现 两个子树的镜像对比 永远是rootLeft->rigth 这个子树和 rootRight-left这两个子树对比数据，rootLeft->left 和rootRight->right子树这两个子树进行对比。
