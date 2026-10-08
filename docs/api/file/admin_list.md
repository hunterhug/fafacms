# /file/admin/list

列出某用户上传的文件。管理员接口。

参考 `/file/list` 接口，请求时如果多一个参数`user_id`。

表示查询该用户下的文件，否则查询所有用户文件。

如：

```
POST /file/admin/list

{
	"url": "/storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg",
	"store_type": -1,
	"is_picture": -1
    "user_id":1
}
```

