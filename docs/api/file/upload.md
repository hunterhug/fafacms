# /file/upload

文件上传。

## 请求

需要前缀，以后不特殊说明都带有前缀 `/v1`。

```
POST /v1/file/upload
```

使用表单 `Content-Type=multipart/form-data` 请求，以后不特殊说明，都使用 `JSON` 形式请求。

表单参数如下：

| 字段   |      含义   | 类型  |   参数 |  必选 |
|----------|--------|------|------|------|
| type | 上传类型 | string | 默认：other，参考以下 |否 |
| describe |    文件描述   |   string |  空 | 否 |
| file | 文件 |    []byte | 二进制文件 | 是 |
| tag | 文件标签（主要对文件分组） | string | 默认：other | 否|

文件类型 `type` 可选 `image`,`flash` 等，参考以下，同时文件要符合指定的类型。

```json
{
"image": {"jpg", "jpeg", "png", "gif"},

"flash": {"swf", "flv"},

"media": {"swf", "flv", "mp3", "wav", "wma", "wmv", "mid", "avi", "mpg", "asf","rm", "rmvb"},

"file": {
    "doc", "docx", "xls", "xlsx", "ppt", "htm", "html", "txt", "zip", "rar", "gz", "bz2", "pdf"},

"other": {
    "jpg", "jpeg", "png", "bmp", "gif", "swf", "flv", "mp3","wav", "wma", "wmv", "mid", "avi", "mpg", "asf", "rm", "rmvb","doc", "docx", "xls", "xlsx", "ppt","htm", "html", "txt", "zip", "rar", "gz", "bz2"}
}
```

## 响应

如果上传的文件为图片，会进行裁剪，裁剪为等宽度裁剪，宽度500px。

文件根据用户，会进行冗余处理，同一个用户上传同一个文件四次，后台只会保留第一次文件。


正常：

```
{
  "flag": true,
  "cid": "87eea9edd6bb408488c6068f9c7645de",
  "data": {
    "file_name": "admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg",
    "really_file_name": "Overlooking_by_Lance_Asper.jpg",
    "size": 1459819,
    "url": "/storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg",
    "url_x": "/storage_x/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg",
    "is_picture": true,
    "addon": "",
    "oss": false
  }
}
```

第二次上传：

```
{
  "flag": true,
  "cid": "1f9f67ba99f5437993279c7b26355fb2",
  "data": {
    "file_name": "admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg",
    "really_file_name": "Overlooking_by_Lance_Asper.jpg",
    "size": 1459819,
    "url": "/storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg",
    "url_x": "/storage_x/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg",
    "is_picture": true,
    "addon": "file the same in server",
    "oss": false
  }
}
```

参数说明：

| 字段   |      含义   | 类型  | 
|----------|--------|------|
| file_name | 文件保存名 | string |
| really_file_name |    上传时文件名   | string |
| size |    文件字节数   | int |
| is_picture |    文件是否为图片   | bool |
| addon |    补充信息如：file the same in server | string |
| oss |    是否保存在对象存储中   | bool |
| url |    文件路径   | string |
| url_x |    图片裁剪后的小尺寸文件路径 (当文件是图片时才会出现) | string |


当 `oss` 为 `false` 时可拼接`域名+url`后打开，如：

```
域名为: http://127.0.0.1:8080
url为： /storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg
```

拼接之后：

```
# 源文件
http://127.0.0.1:8080/storage/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg

# 小尺寸
http://127.0.0.1:8080/storage_x/admin/other/admin_638ecd9624492360284948ba16773aa9d984ba2d04930055a6787b45999fe673.jpg
```

当开启对象存储时，不需要拼接路径。开启OSS后正常返回如下：

```
{
  "flag": true,
  "cid": "e3441d9037ef4f65a37e261e3afe15e2",
  "data": {
    "file_name": "admin_4ecba48def78a35f444e439c11801772019d2058b00583f98a5aee5fa337149e.png",
    "really_file_name": "企业微信截图_52d12f71-3467-4991-87af-b3ff26dadac4.png",
    "size": 466054,
    "url": "xxx.oss-cn-shenzhen.aliyuncs.com/storage/admin/other/admin_4ecba48def78a35f444e439c11801772019d2058b00583f98a5aee5fa337149e.png",
    "url_x": "xxx.oss-cn-shenzhen.aliyuncs.com/storage_x/admin/other/admin_4ecba48def78a35f444e439c11801772019d2058b00583f98a5aee5fa337149e.png",
    "is_picture": true,
    "addon": "",
    "oss": true
  }
}
```


不正常时，如上传非法类型文件：

```
{
  "flag": false,
  "cid": "c3503e6e728840688e12c6b2106f15a4",
  "error": {
    "id": 100101,
    "msg": "upload file type not permit:file suffix: json not permit"
  }
}
```