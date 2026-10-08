# /node/sort

节点排序，可实现拖曳排序。可以拖曳同一层，也可以将二级节点拖到顶层，也可以将顶层拖到二级。

## 请求

```
POST /node/sort


{
	"xid": 1,
	"yid": 2
}
```

参数说明：

| 字段   |      含义   |类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| xid | X节点ID | int | X节点不能等于Y节点| 是 |
| yid | Y节点ID | int | | 否 |

把X节点拖到Y节点下面，X会成为Y的弟弟，当YID为空时，直接把X节点拉到X节点所在层的最顶层。排序数字`sort_num`越大，排越后，本质是自动维护一个递增的排序队列。

X和Y非空时有以下若干种情况：

```
1. 不能拖曳，X是Y的父亲。
--- *X
    --- *Y

2. 同级。
--- *X
--- Z
--- *Y

变成

--- Z
--- *Y
--- *X

3. 二级拖到顶层。
--- *Y
--- Z
    --- L
    --- *X
    --- K
    
变成

--- *Y
--- *X
--- Z
    --- L
    --- K
    
    
4. 顶层拖到二级。
--- *X
--- Z
    --- L
    --- *Y
    --- K
       
变成
   
--- Z
    --- L
    --- *Y
    --- *X
    --- K
    
4. 不能拖曳，顶层拖到二级,X有儿子。
--- *X
    --- P
    --- O
--- Z
    --- L
    --- *Y
    --- K
```


逻辑：

```
先把x假装删掉，比x大的都-1，依次顶上x的位置，把大于y排序的节点都+1，腾出位置给x （sort_num）
同一级
	y>x  y=5 x=2
	  --- a	0		--- a 0			--- a 0
	  --- b 1 ==>	--- b 1 	==>	--- b 1
	  --- x	2		--- xc 2		--- xc 2
	  --- c	3		--- d 3			--- d 3
	  --- d 4		--- y 4			--- y 4
	  --- y	5		--- e 5			---			==> x=5(设置为原来的Y)
	  --- e	6		---  			--- e 6

	y<x  y=2 x=5
	  --- a	0		--- a 0			--- a 0
	  --- b 1 ==>	--- b 1 	==>	--- b 1
	  --- y	2		--- y 2			--- y 2		==> x=3(设置为原来的Y+1)
	  --- c	3		--- c 3			--- c 4
	  --- d 4		--- d 4			--- d 5
	  --- x	5		--- xe 5		---	xe 6
	  --- e	6		---

不同级
	y=1
	  --- a	0		--- a 0				--- a 0
	  --- b 1 	==>	--- b 1 		==>	--- b 1
	  	--- c 0			--- c 0				--- c 0
	  	--- y 1			--- y 1				--- y 1		==> x=2(设置为原来的Y+1)
	  	--- d 2			--- d 2				--- d 3
	  --- x	2		--- xe 2			---	xe 2
	  --- e	3
```

## 响应

正常：

```
{
  "flag": true,
  "cid": "cfe0d18c74154f21a56169ee9b313d01"
}
```

X节点不存在:

```
{
  "flag": false,
  "cid": "c5b6568cf9444a469a4c08091cfb862e",
  "error": {
    "id": 101001,
    "msg": "content node not found:x node not found"
  }
}
```


Y节点不存在：

```
{
  "flag": false,
  "cid": "433fcdb7f7634365a9e85f65b09a110b",
  "error": {
    "id": 101001,
    "msg": "content node not found:y node not found"
  }
}
```

不能拖曳自己：

```
{
  "flag": false,
  "cid": "772d0aaac6c04656bc75319945343293",
  "error": {
    "id": 100010,
    "msg": "paras input not right:xid=yid not right"
  }
}
```

X是Y的爸爸，不可以给儿子做弟弟：

```
{
  "flag": false,
  "cid": "772d0aaac6c04656bc75319945343293",
  "error": {
    "id": 101003,
    "msg": "content node sort conflict:can not move node to be his child's brother"
  }
}
```

X已经有孩子，不可以给是别人儿子的Y当弟弟，因为这样就变成三层了：

```
{
  "flag": false,
  "cid": "772d0aaac6c04656bc75319945343293",
  "error": {
    "id": 101003,
    "msg": "content node sort conflict:x has child can not move to be other's child's brother"
  }
}
```