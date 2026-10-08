# 信息接口

需要前缀`/v1`。注明管理员接口的不需要前缀。

信息分为三种：

1. 系统通知：内容被评论，评论被评论，内容被点赞，评论被点赞，有人关注您，你关注的人发布了内容，内容和评论违禁，内容和评论被解除违禁，都会收到信息通知。
2. 管理员通知：管理员会广播信息给所有人。
3. 私聊：用户与用户间进行聊天。

所有的信息都聚合在一起，有不同的信息类型：

```
	// who comment your content or comment message
	MessageTypeCommentForContent = 0 // 内容被评论
	MessageTypeCommentForComment = 1 // 评论被评论

	// who good your content or comment message
	MessageTypeGoodContent = 2 // 内容被点赞
	MessageTypeGoodComment = 3 // 评论被点赞

	// comment or content be ban by system message
	MessageTypeContentBan = 4 // 内容被违禁
	MessageTypeCommentBan = 5 // 评论被违禁

	// comment or content be ban by recover message
	MessageTypeContentRecover = 6 // 内容被解除违禁
	MessageTypeCommentRecover = 7 // 评论被解除违禁

	// who follow you message
	MessageTypeFollow = 8 // 有人关注你

	// who you follow publish content
	MessageTypeContentPublish = 9 // 你关注的人发布了内容

	// who send a message to you
	MessageTypePrivate = 10 // 私信

	// global send a message to you
	MessageTypeGlobal = 11 // 管理员通知
```

所有的信息，接收时的接口都是同一个。