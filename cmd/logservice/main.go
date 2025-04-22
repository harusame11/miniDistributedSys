package main

import (
	"My_mimiDistributed/log"
	"My_mimiDistributed/registry"
	"My_mimiDistributed/service"
	"context"
	"fmt"
	stlog "log"
)

// main函数是日志服务的入口点
// 日志服务负责接收其他服务发送的日志信息并将其写入文件
func main() {
	// 初始化日志系统，指定日志文件路径
	log.Run("./distributed.log")

	// 设置服务主机名和端口
	host, port := "localhost", "4000"
	// 构造服务完整地址，用于注册到注册中心
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	// 创建服务注册信息对象
	r := registry.Registration{
		ServiceName:      registry.LogService, // 服务名称
		ServiceURL:       serviceAddress,      // 服务地址
		RequireServices:  make([]registry.ServiceName, 0),
		ServiceUpdateURL: serviceAddress + "/services",
	}

	// 调用service包的Start函数启动服务
	// 参数依次为：上下文、注册信息、主机名、端口、HTTP处理函数
	ctx, err := service.Start(
		context.Background(),
		r,
		host,
		port,
		log.RegisterHandlers, // 注册HTTP路由处理函数
	)

	// 如果启动过程中出现错误，记录并退出
	if err != nil {
		stlog.Fatalln(err)
	}

	// 阻塞等待上下文被取消（服务关闭信号）
	<-ctx.Done()

	// 输出服务关闭消息
	fmt.Println("shutting down log service")
}
