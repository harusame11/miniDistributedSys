# 迷你分布式系统 (Mini Distributed System)---------(跟敲版本，go语言版本 ： go1.22)

这是一个用Go语言实现的微服务架构演示项目，展示了微服务的基本组件和交互模式。

## 项目概述

本项目实现了一个简单的分布式系统，包含以下核心服务：

- **注册中心 (Registry Service)**: 管理服务注册与发现
- **日志服务 (Log Service)**: 集中式日志记录
- **成绩服务 (Grading Service)**: 学生成绩管理
- **门户服务 (Portal Service)**: 用户界面入口

系统采用微服务架构，各服务通过HTTP协议通信，使用中央注册中心实现服务发现。

## 项目结构

```
.
├── cmd/                    # 各个服务的入口点
│   ├── logservice/         # 日志服务
│   ├── registryservice/    # 注册中心服务
│   ├── gradingservice/     # 成绩服务
│   └── portal/             # 门户服务
├── log/                    # 日志服务的核心实现
│   ├── client.go           # 日志客户端
│   └── server.go           # 日志服务器
├── registry/               # 服务注册与发现
│   ├── client.go           # 注册中心客户端
│   ├── registration.go     # 服务注册数据结构
│   └── service.go          # 注册中心服务实现
├── service/                # 通用服务组件
│   └── server.go           # 服务启动与生命周期管理
└── go.mod                  # Go模块定义
```

## 核心功能

### 1. 服务注册与发现

- 服务启动时向注册中心注册自身信息
- 服务可声明对其他服务的依赖
- 注册中心将依赖服务信息推送给需要的服务
- 服务关闭时自动从注册中心注销

### 2. 集中式日志记录

- 所有服务通过日志服务统一记录日志
- 日志服务提供REST API接收日志消息
- 日志持久化到文件系统

### 3. 成绩管理

- 提供学生成绩的CRUD操作
- 依赖日志服务记录操作日志

### 4. 用户界面

- 门户服务提供Web界面
- 与成绩服务交互，展示学生成绩

## 技术特点

- **松耦合架构**: 服务之间通过REST API交互，降低耦合度
- **服务发现**: 基于HTTP的轻量级服务注册与发现机制
- **依赖注入**: 使用控制反转模式简化服务实现
- **优雅启停**: 支持服务的优雅启动和关闭
- **并发处理**: 利用Go协程实现高效并发

## 启动顺序

由于服务间存在依赖关系，请按以下顺序启动服务：

1. 注册中心服务 (Registry Service)
2. 日志服务 (Log Service)
3. 成绩服务 (Grading Service)
4. 门户服务 (Portal Service)

## 启动方法

```bash
# 启动注册中心
cd cmd/registryservice
go run main.go

# 启动日志服务
cd cmd/logservice
go run main.go

# 启动成绩服务
cd cmd/gradingservice
go run main.go

# 启动门户服务
cd cmd/portal
go run main.go
```

## 访问服务

- 注册中心: http://localhost:3000
- 日志服务: http://localhost:4000
- 成绩服务: http://localhost:5000
- 门户服务: http://localhost:6000

## 开发与扩展

要添加新服务，请遵循以下步骤：

1. 在`registry/registration.go`中添加新的服务类型常量
2. 在`cmd`目录下创建新服务的入口点
3. 实现服务的核心逻辑
4. 使用`service.Start()`启动服务并注册到注册中心

## 项目特点

这个迷你分布式系统虽然简单，但展示了微服务架构的核心概念和最佳实践，适合作为学习微服务设计模式的参考项目。
