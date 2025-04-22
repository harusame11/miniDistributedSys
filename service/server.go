package service

import (
	"My_mimiDistributed/registry"
	"context"
	"fmt"
	"log"
	"net/http"
)

// Start 函数用于启动微服务
// 这是一个通用的服务启动函数，适用于系统中的所有微服务
// 微服务架构设计模式：提取共同的服务启动逻辑，实现代码复用
// 业务流程:
// 1. 注册服务的HTTP处理函数
// 2. 启动HTTP服务器
// 3. 向注册中心注册服务
// 4. 返回可控制服务生命周期的上下文
// 参数:
// - ctx: 上下文，用于控制服务生命周期
// - reg: 服务注册信息，包含服务名称和URL
// - host: 服务主机名
// - port: 服务监听端口
// - registerHandlesFunc: 注册HTTP路由的回调函数
// 返回:
// - context.Context: 可用于服务生命周期管理的上下文
// - error: 启动过程中的错误
func Start(ctx context.Context, reg registry.Registration, host, port string,
	registerHandlesFunc func()) (context.Context, error) {
	// 调用传入的函数注册HTTP路由处理器
	// 这是依赖注入和控制反转的示例，服务框架不需要知道具体的HTTP处理逻辑
	registerHandlesFunc()

	// 启动HTTP服务器，返回包含取消功能的上下文
	// 这一步使服务开始监听指定端口，准备接收请求
	ctx = startService(ctx, reg.ServiceName, host, port)

	// 向注册中心注册当前服务
	// 这样其他服务就能发现并使用此服务
	// 注册过程还会使当前服务获得它所依赖的服务信息
	err := registry.RegisterService(reg)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}

// startService 启动HTTP服务器并设置优雅关闭机制
// 这是内部辅助函数，负责HTTP服务器的实际启动和生命周期管理
// 微服务最佳实践：实现优雅启动和关闭，确保服务状态一致性
// 业务流程:
// 1. 创建可取消的上下文
// 2. 配置并启动HTTP服务器
// 3. 设置用户控制和监控机制
// 4. 设置服务关闭时的自动注销
// 参数:
// - ctx: 父上下文
// - serviceName: 服务名称，用于日志和提示
// - host: 服务主机名
// - port: 服务监听端口
// 返回:
// - context.Context: 带取消功能的派生上下文
func startService(ctx context.Context, serviceName registry.ServiceName, host, port string) context.Context {
	// 创建一个可取消的上下文，派生自传入的上下文
	// 这使得服务可以被外部信号或内部错误优雅地终止
	ctx, cancel := context.WithCancel(ctx)

	// 创建HTTP服务器实例
	// Go标准库提供的http.Server包含丰富的配置选项
	var srv http.Server

	// 设置服务器监听地址
	// 例如端口4000则为:4000
	srv.Addr = ":" + port

	// 启动一个goroutine运行HTTP服务器
	// 使用goroutine避免阻塞主流程
	// 当服务器关闭或出错时，会调用cancel()
	go func() {
		log.Println(srv.ListenAndServe())
		// 当服务器关闭时，向注册中心注销服务
		// 这确保注册中心维护的服务列表是最新的
		err := registry.ShutdownService(fmt.Sprintf("http://%s:%s", host, port))
		if err != nil {
			log.Println(err)
		}
		cancel()
	}()

	// 启动一个goroutine监听用户输入，实现优雅关闭
	// 这提供了一种通过控制台手动关闭服务的方式
	go func() {
		fmt.Printf(" %v start ,press any key to stop service \n", serviceName)
		var s string
		// 阻塞等待用户输入
		fmt.Scanln(&s)

		// 用户输入后，向注册中心注销服务
		err := registry.ShutdownService(fmt.Sprintf("http://%s:%s", host, port))
		if err != nil {
			log.Println(err)
		}

		// 优雅关闭HTTP服务器
		// Shutdown会等待所有活跃连接完成后再关闭
		srv.Shutdown(ctx)

		// 取消上下文，通知所有监听此上下文的goroutine
		cancel()
	}()

	return ctx
}
