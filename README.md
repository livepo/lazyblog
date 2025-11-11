## lazyblog

初始化数据库： `go run cmd/main.go --initdb`

运行: `go run cmd/main.go`

## 文章发表

重名文件覆盖式发布

```sh
curl -v -X POST \
  -H "X-Admin-Token: YOUR_ADMIN_TOKEN_HERE" \
  -F "file=@/path/to/local/file.md" \
  http://localhost:8080/admin/publish
```
## 效果
见 [阿Q的博客](https://docset.vip)

## TODO
- search
- toc ?
