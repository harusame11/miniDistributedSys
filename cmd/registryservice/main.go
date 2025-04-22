package main

import (
	"My_mimiDistributed/registry"
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// main函数是注册中心服务的入口点
// 注册中心是整个微服务架构的核心组件，负责服务发现和注册
// 业务流程:
// 1. 设置HTTP处理函数，用于服务注册、注销和查询
// 2. 启动HTTP服务器监听请求
// 3. 等待服务终止
func main() {
	// 创建HTTP多路复用器
	// 用于将不同路径的请求路由到相应的处理函数
	// registry.RegistryService实现了ServeHTTP方法，可处理/services路径的请求
	http.Handle("/services", &registry.RegistryService{})

	// 创建上下文用于控制服务生命周期
	// 当服务需要关闭时，可以取消这个上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 定义等待组，用于等待所有goroutine完成
	// 这确保服务在关闭前完成所有必要的清理工作
	var wg sync.WaitGroup

	// 添加一个等待任务
	wg.Add(1)

	// 启动一个goroutine运行HTTP服务器
	// 使用goroutine避免阻塞主流程
	go func() {
		// 启动HTTP服务器在8000端口监听
		// 服务发现和注册的所有API都通过这个端口提供
		log.Println(http.ListenAndServe(":3000", nil))

		// 当服务器关闭时，取消上下文
		// 这会通知所有使用此上下文的goroutine结束工作
		cancel()

		// 标记等待任务完成
		wg.Done()
	}()

	// 启动一个goroutine监听用户输入，实现优雅关闭
	// 这提供了一种通过控制台手动关闭注册中心的方式
	go func() {
		// 等待上下文结束信号
		// 当另一个goroutine调用cancel()时，这里会收到信号
		<-ctx.Done()

		// 打印提示信息，表示服务即将关闭
		fmt.Println("Registry service shutting down")

		// 标记等待任务完成
		wg.Done()
	}()

	// 打印提示信息，表示服务已启动
	fmt.Println(" Registry service started. Press any key to stop")

	// 等待用户输入，提供手动关闭服务的方式
	var s string
	fmt.Scanln(&s)

	// 用户输入后，取消上下文，通知所有goroutine
	// 这是优雅关闭的第一步，确保相关资源能够正确释放
	cancel()

	// 等待所有goroutine完成，确保优雅关闭
	// 这是微服务设计中的最佳实践，避免资源泄露
	wg.Wait()
}
