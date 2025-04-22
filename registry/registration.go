package registry

// Registration 结构体定义了服务注册所需的信息
// 每个微服务在注册时都需要提供这些信息，它是服务注册与发现的核心数据结构
type Registration struct {
	// ServiceName 是服务的唯一标识名称，用于在注册中心标识服务类型
	ServiceName ServiceName

	// ServiceURL 是服务的完整URL地址，包括协议、主机名和端口
	// 其他服务将使用此URL与该服务通信
	ServiceURL string

	// RequireServices 声明此服务依赖的其他服务列表
	// 例如：GradingService依赖LogService进行日志记录
	// 注册中心会根据此字段向服务推送其依赖服务的信息
	RequireServices []ServiceName

	// ServiceUpdateURL 是服务用来接收依赖更新的回调URL
	// 注册中心通过向此URL发送POST请求通知服务其依赖的变化
	// 例如：http://localhost:6000/services
	ServiceUpdateURL string
}

// ServiceName 是服务名称的类型别名
// 使用类型别名可以提供类型安全，并允许定义服务类型常量
// 这比使用普通字符串更安全，可避免拼写错误
type ServiceName string

// 系统中的服务类型常量
// 在代码中使用这些常量而非直接使用字符串，提高类型安全性
const (
	// LogService 是日志服务的名称，提供集中式日志记录功能
	LogService = ServiceName("LogService")

	// GradingService 是成绩服务的名称，提供学生成绩管理功能
	GradingService = ServiceName("GradingService")

	// PortalService 是门户服务的名称，提供学生成绩管理功能
	PortalService = ServiceName("PortalService")
)

// patchEntry 表示单个服务更新条目
// 用于构建服务依赖更新通知
type patchEntry struct {
	// 服务名称
	Name ServiceName
	// 服务URL
	URL string
}

// patch 结构用于服务依赖更新通知
// 注册中心通过发送patch对象通知服务其依赖的变化
type patch struct {
	// Added 包含新增的依赖服务信息
	Added []patchEntry

	// Removed 包含被移除的依赖服务信息
	Removed []patchEntry
}
