
# Pineapple - etcd 服务发现模块

基于 Go 语言开发的轻量级 etcd 服务发现库。

## 技术栈

- **编程语言**: Go 1.21+
- **服务发现**: etcd
- **日志**: zap + lumberjack
- **配置管理**: viper

## 项目结构

```
pineapple/
├── internal/               # 内部代码
│   ├── discovery/         # 服务发现 (核心模块)
│   └── utils/             # 工具库 (配置、日志)
├── config/                # 配置文件
├── go.mod
├── go.sum
└── README.md
```

## 快速开始

### 前置要求

- Go 1.21+
- etcd

### 安装

```bash
# 克隆项目
git clone https://github.com/beijian128/pineapple.git
cd pineapple

# 安装依赖
go mod tidy
```

## 测试指南

### 前置要求

在测试之前，请确保以下服务已启动：

1. **etcd 服务** (默认端口: 2379)
   - 下载: https://github.com/etcd-io/etcd/releases
   - 启动命令: `etcd`
   - 验证: `curl http://localhost:2379/health`

### 快速测试

#### 1. 使用示例程序测试

项目提供了一个完整的测试程序，你可以直接运行：

```bash
cd pineapple/examples
go run discovery_test.go
```

这个测试程序会：
- 加载配置文件
- 初始化日志系统
- 连接 etcd
- 注册两个服务
- 发现服务
- 演示租约保持

#### 2. 手动测试

你也可以创建一个简单的程序进行测试：

```go
package main

import (
	"fmt"
	"time"

	"github.com/beijian128/pineapple/internal/discovery"
	"github.com/beijian128/pineapple/internal/utils"
)

func main() {
	// 1. 加载配置
	if err := utils.LoadConfig("../config/config.yaml"); err != nil {
		panic(err)
	}

	// 2. 初始化日志
	if err := utils.InitLogger(&utils.AppConfig.Log); err != nil {
		panic(err)
	}
	defer utils.SyncLogger()

	// 3. 连接 etcd
	if err := discovery.InitDiscovery(&utils.AppConfig.Etcd); err != nil {
		panic(err)
	}
	defer discovery.CloseDiscovery()

	// 4. 注册服务
	service := &discovery.ServiceInfo{
		Name:    "game-server",
		Addr:    "127.0.0.1",
		Port:    9000,
		Version: "1.0.0",
	}
	if err := discovery.GlobalDiscovery.Register(service); err != nil {
		panic(err)
	}
	fmt.Println("服务注册成功！")

	// 5. 等待一段时间，演示租约保持
	time.Sleep(30 * time.Second)

	// 6. 程序退出时会自动注销服务
}
```

### 使用示例

#### 1. 配置文件 `config.yaml`

```yaml
etcd:
  endpoints: ["localhost:2379"]
  dial_timeout: "5s"
  lease_ttl: 10

log:
  level: "info"
  filename: "logs/pineapple.log"
  max_size: 100
  max_backups: 3
  max_age: 28
```

#### 2. 初始化并使用服务发现

```go
package main

import (
	"github.com/beijian128/pineapple/internal/discovery"
	"github.com/beijian128/pineapple/internal/utils"
)

func main() {
	// 加载配置
	if err := utils.LoadConfig("./config/config.yaml"); err != nil {
		panic(err)
	}

	// 初始化日志
	if err := utils.InitLogger(&utils.AppConfig.Log); err != nil {
		panic(err)
	}
	defer utils.SyncLogger()

	// 初始化 etcd 服务发现
	if err := discovery.InitDiscovery(&utils.AppConfig.Etcd); err != nil {
		panic(err)
	}
	defer discovery.CloseDiscovery()

	// 创建服务信息
	service := &discovery.ServiceInfo{
		Name:    "my-service",
		Addr:    "127.0.0.1",
		Port:    8080,
		Version: "1.0.0",
	}

	// 注册服务
	if err := discovery.GlobalDiscovery.Register(service); err != nil {
		panic(err)
	}
	defer discovery.GlobalDiscovery.Unregister(service)

	// 发现服务
	services, err := discovery.GlobalDiscovery.Discover("my-service")
	if err != nil {
		panic(err)
	}
}
```

## 核心模块

### 1. 服务发现 (discovery/)

- etcd 客户端封装
- 服务注册与发现
- 租约保持与自动续期
- 服务健康检查

### 2. 工具库 (utils/)

- 配置管理 (Viper)
- 日志封装 (Zap + Lumberjack)

## 开发计划

- [x] 基础 etcd 服务发现功能
- [ ] 服务健康检查增强
- [ ] 服务状态通知
- [ ] 负载均衡
- [ ] 完善文档和示例

## License

MIT
