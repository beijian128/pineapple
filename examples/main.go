
package main

import (
	"fmt"
	"time"

	"github.com/beijian128/pineapple/internal/discovery"
	"github.com/beijian128/pineapple/internal/utils"
)

func main() {
	fmt.Println("=== Pineapple - etcd 服务发现测试 ===\n")

	// 1. 加载配置
	configPath := "../config/config.yaml"
	fmt.Printf("1. 加载配置文件: %s\n", configPath)
	if err := utils.LoadConfig(configPath); err != nil {
		fmt.Printf("   ❌ 加载配置失败: %v\n", err)
		fmt.Println("\n提示: 请确保 etcd 服务已启动 (默认端口: 2379)")
		return
	}
	fmt.Println("   ✅ 配置加载成功\n")

	// 2. 初始化日志
	fmt.Println("2. 初始化日志系统")
	if err := utils.InitLogger(&utils.AppConfig.Log); err != nil {
		fmt.Printf("   ❌ 初始化日志失败: %v\n", err)
		return
	}
	defer utils.SyncLogger()
	fmt.Println("   ✅ 日志系统初始化成功\n")

	// 3. 初始化 etcd 客户端
	fmt.Println("3. 连接 etcd 服务")
	if err := discovery.InitDiscovery(&utils.AppConfig.Etcd); err != nil {
		fmt.Printf("   ❌ 初始化 etcd 失败: %v\n", err)
		fmt.Println("\n提示: 请检查 etcd 是否已启动在 127.0.0.1:2379")
		return
	}
	defer discovery.CloseDiscovery()
	fmt.Println("   ✅ etcd 连接成功\n")

	// 4. 测试服务注册
	fmt.Println("4. 测试服务注册")
	service1 := &discovery.ServiceInfo{
		Name:    "user-service",
		Addr:    "127.0.0.1",
		Port:    8080,
		Version: "1.0.0",
	}
	if err := discovery.GlobalDiscovery.Register(service1); err != nil {
		fmt.Printf("   ❌ 服务注册失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 服务注册成功: %s:%d\n", service1.Addr, service1.Port)
	defer func() {
		_ = discovery.GlobalDiscovery.Unregister(service1)
	}()

	service2 := &discovery.ServiceInfo{
		Name:    "user-service",
		Addr:    "127.0.0.1",
		Port:    8081,
		Version: "1.0.0",
	}
	if err := discovery.GlobalDiscovery.Register(service2); err != nil {
		fmt.Printf("   ❌ 服务注册失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 服务注册成功: %s:%d\n", service2.Addr, service2.Port)
	defer func() {
		_ = discovery.GlobalDiscovery.Unregister(service2)
	}()
	fmt.Println()

	// 5. 测试服务发现
	fmt.Println("5. 测试服务发现")
	time.Sleep(100 * time.Millisecond)
	services, err := discovery.GlobalDiscovery.Discover("user-service")
	if err != nil {
		fmt.Printf("   ❌ 服务发现失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 发现 %d 个服务:\n", len(services))
	for _, svc := range services {
		fmt.Printf("      - %s:%d (版本: %s)\n", svc.Addr, svc.Port, svc.Version)
	}
	fmt.Println()

	// 6. 等待，演示租约保持
	fmt.Println("6. 演示租约保持 (等待 5 秒，按 Ctrl+C 退出)")
	fmt.Println("   在此期间服务会自动保持租约有效...")
	for i := 5; i > 0; i-- {
		fmt.Printf("   剩余时间: %d 秒\n", i)
		time.Sleep(1 * time.Second)
	}
	fmt.Println()

	// 7. 再次测试服务发现
	fmt.Println("7. 再次测试服务发现")
	services, err = discovery.GlobalDiscovery.Discover("user-service")
	if err != nil {
		fmt.Printf("   ❌ 服务发现失败: %v\n", err)
		return
	}
	fmt.Printf("   ✅ 发现 %d 个服务\n", len(services))
	fmt.Println()

	fmt.Println("=== 测试完成 ===")
	fmt.Println("\n提示: 你可以查看 etcd 中的数据或日志文件了解更多细节")
}
