# BFBP
百度云分享码提取+爆破
> 起因`穷`，网上有的爆破程序都老古董了，用不了，技术有限，只有自己乱整一个能跑的

效率很拉很拉，不过能跑出来

> 注意：脚本要用到云函数做代理，绕频控，然而云函数的网关也容易崩溃，所以，爆破全凭运气。


### 使用
+ 运行160w.go 生成16组密码
+ SCF_COM.go 、 getCookie.go 都部署到云函数【这个就自个去网上找教程了，一大堆】
+ 在 bfbp-main.go 填自己的云函数代理地址
+ 运行bfbp-main
```eg:
go run .\bfbp-main.go -h
go run .\bfbp-main.go -burl https://pan.baidu.com/share/init?surl=b_7kYyEwxZjt_M6VEmMC5A -w 1
```
