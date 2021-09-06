# TodoList

- [x] **系统后端功能**
    - [x] 内部用户及用户组，资源和授权管理（管理员）
        - [x] 创建用户组
        - [x] 删除用户组
        - [x] 获取组信息
        - [x] 更新组信息
        - [x] 创建用户
        - [x] 获取用户信息
        - [x] 更改用户密码
        - [x] 禁用用户
        - [x] 用户绑定组
        - [x] 资源绑定组
        - [x] 列出组下的用户
        - [x] 列出组下的资源
        - [x] 授权拦截
        - [x] 列出用户
        - [x] 列出用户组
        - [x] 列出资源
        - [x] 用户登录（个人）
        - [x] 用户登出（个人）
        - [x] 获取个人信息（个人）
        - [x] 更新个人信息（个人或管理员）
        - [x] 赋予用户VIP，设置用户状态（管理员）       
    - [x] 文件功能（个人）
        - [x] 上传文件
        - [x] 图片文件裁剪
        - [x] 阿里对象存储保存
        - [x] 列出文件（个人或管理员）
        - [x] 更改文件描述，标签，是否隐藏 （个人或管理员）
    - [x] 内容节点及内容功能（个人）
        - [x] 节点功能
            - [x] 创建节点
            - [x] 更新节点SEO
            - [x] 更新节点名字和描述
            - [x] 更新节点背景图
            - [x] 设置节点隐藏
            - [x] 设置节点的父节点
            - [x] 拖曳排序节点
            - [x] 删除节点
            - [x] 获取节点信息
            - [x] 列出节点（个人或管理员）
        - [x] 内容功能（个人）
            - [x] 创建内容
            - [x] 更新内容SEO
            - [x] 更新内容的标题和正文
            - [x] 更新内容背景图
            - [x] 设置内容隐藏
            - [x] 设置内容置顶
            - [x] 设置内容密码
            - [x] 设置内容是否可评论
            - [x] 更改内容的节点
            - [x] 拖曳排序内容
            - [x] 内容更新历史记录
            - [x] 发布内容
            - [x] 恢复历史版本中的内容
            - [x] 获取内容信息（个人或管理员）
            - [x] 获取历史版本内容（个人或管理员）
            - [x] 列出内容（个人或管理员）
            - [x] 列出历史版本内容（个人或管理员）
            - [x] 设置内容状态，违禁/正常（管理员）
            - [x] 内容丢到回收站
            - [x] 回收站内容恢复
            - [x] 删除回收站内容
            - [x] 删除历史版本内容
    - [x] 内容评论功能（个人）
        - [x] 新增评论
        - [x] 取消匿名
        - [x] 回复评论
        - [x] 列出评论（管理员）
        - [x] 设置评论状态，违禁/正常（管理员）
        - [x] 删除评论
    - [x] 点赞功能（个人）
        - [x] 内容点赞
        - [x] 评论点赞
    - [x] 举报功能（个人）
        - [x] 内容举报
        - [x] 评论举报
    - [x] 关注和被关注功能（个人）
        - [x] 添加关注
        - [x] 解除关注
        - [x] 查看关注你的人
        - [x] 查看你关注的人
        - [x] 查看A用户关注的人
        - [x] 查看关注B用户的人
        - [x] 列出所有关系（管理员）
    - [x] 站内信功能（个人）
        - [x] 关注用户发布内容站内信
        - [x] 评论以及评论回复站内信
        - [x] 关注站和被关注站内信
        - [x] 内容和评论点赞站内信
        - [x] 内容和评论违禁站内信
        - [x] 内容和评论解禁站内信
        - [x] 内容和评论解禁站内信
        - [x] 拉取站内信    
        - [x] 批量站内信已读
        - [x] 批量删除站内信
        - [x] 查看全部站内信（管理员）
        - [x] 创建通知全局站内信（管理员）               
        - [x] 列出通知全局站内信（管理员）
        - [x] 修改通知全局站内信状态（管理员）
- [x] **系统中端功能（个人）**
    - [x] 普通用户功能
        - [x] 用户注册
        - [x] 用户激活
        - [x] 重置激活码
        - [x] 找回密码
        - [x] 邮箱发送激活码和密码重置码
    - [x] 列出所有激活未封禁用户
    - [x] 获取某用户信息
    - [x] 列出某用户全部节点信息
    - [x] 获取某用户某节点以及其子节点信息
    - [x] 统计用户不同时期发布的正常内容数量
    - [x] 列出用户已发布的正常内容
    - [x] 获取文章内容
    - [x] 获取内容下评论
    - [x] 获取内容点赞情况
    - [x] 获取评论点赞情况
- [ ] **系统前端UI功能**
    - [ ] 普通用户后台界面
    - [ ] 管理员后台界面
    - [ ] 前端首页界面
- [x] **API文档**
    - [x] 一期基本文档
    - [x] 二期基本文档
- [ ] **附加功能**
    - [ ] 搜索功能
    - [x] 用户间私信
    - [ ] 内容标签功能
    - [ ] 验证码功能  
    
当用户量突破一定数量时，关闭注册，或者收费注册。作为一个小社区而存在。当并发数和数据量巨大无比时，开启阿里云oss和使用k8s副本部署，tidb分布式mysql可缓解，问题不大。

# 前端网站开发/移动端开发

正在进行中...

# API后端需求整理

`1.0`版本，后端已经实现了基本的权限控制，用户，文件，节点，内容等模块功能，做一个简单的多用户内容系统已经绰绰有余。

升级版`V2`需求整理。

- ~~评论：~~

~~用户对内容评论，对评论评论，仿QQ音乐。~~

- ~~-站内信：~~

~~通知发布文章，被别人评论，内容被别人点赞，文章被点赞，文章被拉黑，对别人发表评论，发表的评论被禁。站内信可设置删除。~~

- ~~关注好友~~

~~被关注，关注的好友发文章会站内信通知。~~

- 验证码

每次获取验证码API，生成码和回答存入redis，有效期60s，某些接口需携带生成码和回答，进行验证，无论通不通过都删除redis。不考虑互斥。

- 社交登录

Github和其他社交登陆。

- 赞助

内容打赏，没有库存和退款概念，不记名，不需要登录。列出赞助名单。站内信，有人给你打赏了。管理员后台查看所有赞助记录。

- ~~系统消息，私聊~~

- 搜索