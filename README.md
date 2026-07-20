# Walk-Server

基于gbc&mygo框架的毅行后端代码，分成用户端，扫码端，数据大盘三个部分。

项目基于 Go、Gin、GORM Gen 和 [mygo](https://github.com/zjutjh/mygo) 构建，当前代码结构已经按业务域拆分为 `admin`、`user`、`dashboard` 三组接口，并使用 `repo + cache + lock` 的方式处理数据访问、缓存和并发写入。

## 功能概览

### 用户侧

- 微信登录
- 学生 / 教职工 / 校友注册
- 用户信息查询与修改
- 创建队伍、加入队伍、查看队伍信息
- 修改队伍信息、退出队伍、解散队伍

### 管理员侧

- 管理员注册与登录
- 绑定队伍签到码
- 队伍点位打卡
- 查询队伍当前状态
- 查询用户信息 / 扫码查人
- 更改人员状态
- 标记队伍违规
- 确认到达终点
- 重组队伍

### Dashboard

- 总览统计
- 单路线统计
- 点位详情
- 路段人数统计
- 队伍筛选
- 队伍详情
- 队伍失联状态设置
- 权限查询

## 技术栈

- Go
- Gin
- GORM / GORM Gen
- MySQL
- Redis
- mygo

## 目录结构

```text
.
├── api
│   ├── admin          # 管理员接口
│   ├── dashboard      # Dashboard 接口
│   └── user           # 用户接口
├── comm               # 错误码、枚举、锁与公共方法
├── conf               # 配置文件
├── dao
│   ├── cache          # Redis 缓存
│   ├── model          # gorm/gen 生成的 model
│   ├── query          # gorm/gen 生成的 query
│   └── repo           # 数据访问层
├── deploy/sql         # 建表与测试数据
├── middleware         # 登录态、权限等中间件
├── register           # Boot、路由、命令、定时任务注册
└── cmd/gen            # GORM Gen 生成入口
```

## 环境要求

- Go 1.24+
- MySQL 8.x
- Redis

## 配置说明

配置模板见 [`conf/config.example.yaml`](conf/config.example.yaml)。

最少需要确认以下配置：

```yaml
app:
  env: "dev"

http_server:
  addr: ":8888"

db:
  host: "127.0.0.1"
  port: 3306
  username: "jh_user"
  password: "jh_pass"
  database: "jh_db"

redis:
  addrs:
    - "127.0.0.1:6379"
  db: 0
  password: "jh_pass"

lock:
  redis: "redis"

session:
  driver: "memory"
  name: "session"
  secret: "secert"
  redis: ""
```

说明：

- `db` 用于 MySQL 连接
- `redis` 用于缓存与分布式锁
- `lock.redis` 指定锁依赖的 Redis 实例名
- `session.driver` 可以用 `memory` 或 `redis`
- `biz` 中还包含活动开放时间、截止时间、队伍人数上下限等业务配置

## 启动方式

复制配置文件：

```bash
cp conf/config.example.yaml conf/config.yaml
```

启动服务：

```bash
go run .
```

项目入口见 [main.go]。启动后会同时拉起：

- HTTP 服务
- HTTP 服务伴生定时任务

## 编译校验

```bash
GOCACHE=/tmp/gocache go build ./...
```

## 代码生成

项目使用 `gorm.io/gen` 维护 `dao/model` 和 `dao/query`。

生成入口：

- [generate.go]

执行命令：

```bash
go run cmd/gen/generate.go
```

当表结构变更后，建议重新生成。

