# 白荆回廊释放技能后自动暂停

必须用管理员运行，原理很简单就是识图，检测到放完技能就鼠标按一下暂停。

## 源码构建

```bash
go build -ldflags "-s -w"
```