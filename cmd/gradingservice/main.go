package main

import (
	"My_mimiDistributed/grades.go"
	"My_mimiDistributed/log"
	"My_mimiDistributed/registry"
	"My_mimiDistributed/service"
	"context"
	"fmt"
	stlog "log"
)

func main() {

	// 设置服务主机名和端口
	host, port := "localhost", "6000"
	// 构造服务完整地址，用于注册到注册中心
	serviceAddress := fmt.Sprintf("http://%s:%s", host, port)

	// 创建服务注册信息对象
	r := registry.Registration{
		ServiceName:      registry.GradingService, // 服务名称
		ServiceURL:       serviceAddress,
		RequireServices:  []registry.ServiceName{registry.LogService},
		ServiceUpdateURL: serviceAddress + "/services",
	}

	// 调用service包的Start函数启动服务
	// 参数依次为：上下文、注册信息、主机名、端口、HTTP处理函数
	ctx, err := service.Start(
		context.Background(),
		r,
		host,
		port,
		grades.RegisterHandlers, // 注册HTTP路由处理函数
	)

	// 如果启动过程中出现错误，记录并退出
	if err != nil {
		stlog.Fatalln(err)
	}
	if logProvider, err := registry.GetProvider(registry.LogService); err == nil {
		fmt.Printf("log service found at %s\n", logProvider)
		log.SetClientLogger(logProvider, r.ServiceName)
	}

	// 阻塞等待上下文被取消（服务关闭信号）
	<-ctx.Done()

	// 输出服务关闭消息
	fmt.Println("shutting down log service")
}
