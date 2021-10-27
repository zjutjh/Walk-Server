# Walk2021-server

### 如何使用配置配置文件
配置文件默认在 ./config 目录下的 config.yaml，日后会添加上动态生成配置文件，和读取不同位置的配置文件的功能

配置文件样例为 ./config/config.example.yaml 文件

配置文件 ./config/config.yaml 不可以上传到 Github 上，否则重要开发信息泄漏，后果自负 

### 项目文件说明

### main.go 
程序的入口，同时用来绑定路由

### controller
这个地方存放每个路由对应的控制器

### utility 
用来存放一些一些工具小函数

比如说获取用户 open ID 的相关函数
