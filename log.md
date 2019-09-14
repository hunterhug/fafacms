## Log Update

### 201909

1. 支持传入flag `-time_zone`参数设置时区，默认为东八区北京时间。
2. 修复某些空指针异常，补充节点内容相关API文档，前端优化。
3. 自动化部署[脚本](install/README.md)数据库升级：`mysql:5.7.27`，`redis:5.0.5`，管理工具集成：`phpmyadmin:edge-4.9`。
4. 强制要求节点和内容SEO必须存在，方便前端开发。
5. 修复内容API若干缺陷，细化内容筛选条件。
6. 增加删除历史内容API，细化历史内容筛选条件。

### 201908

1. 可选择不将内容记录进历史。
2. 改为单点登录，任何用户登录会挤掉其他端的用户。
3. 用户访问临时令牌加长到7天有效。
4. 数据库切为utf8mb4完善文本保存。
5. 自动创建数据库，不必再手动创建。
6. 启动时打印版本号
7. 完善邮件发送逻辑，验证码有效期为5分钟，激活码重置不需要之前的激活码。
8. 完善用户注册，忘记密码，用户，用户组操作等前端API文档。
9. 完成用户注册，登录，忘记密码，注销前端UI。

### 201904-06

1. 支持本地文件模式，也支持阿里云对象存储文件上下载
2. 注册用户，用户忘记密码发邮件功能
3. 取消了cookie功能，让大前端自己实现，只实现自定义的用户session redis功能，改为token API模式，需授权API均检测该token，token和用户信息均保存在redis，缓存击穿时再从mysql加载
4. 实现自动创建数据库，数据库表，填充管理员账户。
5. 自动将admin URL置于变量内存中，避免管理员API权限过滤时，频繁查找数据库
6. 文章节点，文章等有菜单排序的功能，均支持强大拖拽排序
7. 实现自动docker hub打包镜像发布，提供从数据库到后端的一键部署脚本。

### 201904

1. 调研。设计产品。
2. 资源准备：阿里云对象存储测试，邮件测试，测试服务器准备。
3. 代码框架预设，抽离整理，实现基本功能。