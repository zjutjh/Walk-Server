# Walk2021-server

> 本文档中的所有路径都用 / 开头，这个根目录指代项目所在的目录

### 功能说明
当前版本为 V1.0.1 

缺失的功能有
1. 随机队伍

### 数据返回说明
一定要使用 /utility/response.go 下的函数来返回数据

### 如何使用配置配置文件
配置文件默认在 /config 目录下的 config.yaml，日后会添加上动态生成配置文件，和读取不同位置的配置文件的功能

配置文件样例为 /config/config.example.yaml 文件

配置文件 /config/config.yaml 不可以上传到 Github 上，否则重要开发信息泄漏，后果自负 

### 项目文件说明
```text
./
├── LICENSE
├── README.md
├── config     (配置文件目录)
│        ├── config.example.yaml
│        └── config.yaml
├── controller (控制器 -> 每个路由对应的回调函数)
│        ├── basic.go
│        ├── ······
│        ├── team.go
│        └── user.go
├── go.mod (go 项目文件)
├── go.sum (go 依赖版本控制文件)
├── main.go
├── middleware (中间件 -> 在请求传入到控制函数前对请求数据做一些处理)
│        ├── auth.go
│        └── validity.go
├── model      (数据库模型 -> 用来描述数据库表的结构体)
│        ├── person.go
│        ├── team.go
│        └── team_count.go
├── utility    (工具函数 -> 一些常用的工具函数 比如说获取当前是毅行报名第几天的函数)
│        ├── crypto.go
│        ├── date.go
│        ├── initial
│        │       ├── init_config.go
│        │       ├── init_db.go
│        │       └── init_router.go
│        ├── jwt.go
│        ├── response.go
│        ├── serve.go
│        └── wechat.go
└── walk-server
```

### 如何启动
#### 开启 go module 和换源
[https://goproxy.cn](https://goproxy.cn) 

请按照这个网站的说明开启 go module 特性, 并切换 go proxy

#### 创建数据库
手动在 mysql 中创建一个数据库，名字任意

#### 修改配置文件
在 /config 文件夹下复制 config.example.yaml 文件并将复制出来的文件改名为
config.yaml, 修改对应的选项，注意数据库部分，要根据你的本机数据库进行调整

#### 调整 mysql 并调大 Linux 内核支持的最大文件句柄数（服务上线时对 Linux 的调整)

首先调整 mysql 的线程数和最大连接数

根据 **2021** 年毅行报名的经验:

> 最大连接数可以调整为 900
> 
> 线程池数量可以调整为 64

然后请根据这个网站调整 Linux 服务器内核支持打开的最大文件数

[http://woshub.com/too-many-open-files-error-linux/](http://woshub.com/too-many-open-files-error-linux/)

#### 编译项目
```bash
go build
```

#### 后台运行
```bash
nohup ./程序名 &
```

> Go 编译会自动安装依赖