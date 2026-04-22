
# Pineapple - 分布式游戏服务器框架

基于 Go 语言开发的高性能分布式游戏服务器框架。

## 技术栈

- **编程语言**: Go 1.21+
- **服务发现**: etcd
- **数据序列化**: Protocol Buffers
- **数据库**: MongoDB
- **缓存**: Redis
- **网络协议**: TCP / KCP / WebSocket / gRPC
- **日志**: zap + lumberjack
- **配置管理**: viper

## 项目结构

```
pineapple/
├── cmd/                    # 可执行程序入口
│   ├── gateway/            # 网关服务
│   ├── login/              # 登录服务
│   └── game/               # 游戏服务
├── internal/               # 内部代码
│   ├── net/               # 网络层
│   ├── discovery/         # 服务发现
│   ├── router/            # 消息路由
│   ├── storage/           # 数据存储
│   ├── cluster/           # 分布式组件
│   └── utils/             # 工具库
├── api/                   # protobuf 定义
│   └── proto/
├── config/                # 配置文件
├── scripts/               # 脚本工具
├── docs/                  # 文档
├── go.mod
├── go.sum
└── README.md
```

## 快速开始

### 前置要求

- Go 1.21+
- etcd
- MongoDB
- Redis

### 运行

```bash
# 克隆项目
git clone https://github.com/beijian128/pineapple.git
cd pineapple

# 安装依赖
go mod tidy

# 运行网关服务
go run cmd/gateway/main.go
```

## 核心模块

### 1. 网络层 (net/)
- TCP 服务器
- WebSocket 支持 (待实现)
- KCP 协议支持 (待实现)
- gRPC 服务支持 (待实现)
- 消息编解码

### 2. 服务发现 (discovery/)
- etcd 客户端封装
- 服务注册与发现
- 健康检查
- 租约保持

### 3. 消息路由 (router/)
- 消息分发路由
- 处理器注册机制
- 中间件支持 (Logger, Recovery)
- Context 上下文传递
- 消息ID定义

### 4. 数据存储 (storage/)
- MongoDB 封装
- Redis 封装

### 5. 工具库 (utils/)
- 配置管理
- 日志封装

## 消息路由使用示例

```go
r := router.NewRouter()

// 使用中间件
r.Use(router.Recovery())
r.Use(router.Logger())

// 注册处理器
r.RegisterFunc(router.MsgIDLoginRequest, func(c *router.Context) {
    // 处理登录请求
    resp := &amp;net.Message{
        MsgID: router.MsgIDLoginResponse,
        Data:  []byte(`{"code":0,"message":"ok"}`),
    }
    if r, ok := c.Get("router").(*router.Router); ok {
        _ = r.Send(c.Conn, resp)
    }
})

// 集成到 TCP 服务器
handler := &amp;GatewayHandler{router: r}
server := net.NewTCPServer(8888, handler)
```

## 配置说明

参见 `config/config.yaml`

## 开发计划

- [x] 项目基础搭建
- [ ] 网络层完善 (WebSocket/KCP/gRPC)
- [x] 消息路由系统
- [ ] 分布式组件
- [ ] 示例游戏服务
- [ ] 压力测试
- [ ] 文档完善

## License

MIT
