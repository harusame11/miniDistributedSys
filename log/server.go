package log

import (
	"io"
	stlog "log"
	"net/http"
	"os"
)

// 全局日志记录器实例，用于写入日志文件
// 它由Run函数初始化，并由write函数使用
var log *stlog.Logger

// fileLog 是一个自定义字符串类型，实现了io.Writer接口
// 用作日志的目标写入器，将日志写入指定的文件路径
// 在微服务架构中，分离日志记录逻辑是一个良好实践
type fileLog string

// Write 实现io.Writer接口的方法，将数据写入到文件
// 每次写入都会打开文件，写入数据后关闭，确保数据被持久化
// 这种方式虽然有性能开销，但确保了每条日志都被写入磁盘，避免丢失
// 参数:
// - data: 要写入的字节数据
// 返回:
// - int: 写入的字节数
// - error: 写入过程中的错误
func (fl fileLog) Write(data []byte) (int, error) {
	// 打开文件，如果文件不存在则创建，以追加模式写入
	// O_CREATE: 如果文件不存在则创建
	// O_WRONLY: 以只写模式打开
	// O_APPEND: 追加写入，不覆盖已有内容
	// 0600: 设置文件权限为仅所有者可读写
	f, err := os.OpenFile(string(fl), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}
	// 确保文件被关闭，防止资源泄漏
	// defer确保即使发生错误也会执行关闭操作
	defer f.Close()

	// 写入数据并返回结果
	// 返回写入的字节数和可能的错误
	return f.Write(data)
}

// Run 初始化日志系统
// 此函数在日志服务启动时被调用，设置日志记录器
// 业务流程:
// 1. 创建指向指定文件的日志记录器
// 2. 设置日志格式和前缀
// 参数:
// - destination: 日志文件的路径
func Run(destination string) {
	// 创建新的日志记录器，使用fileLog作为输出目标
	// 参数1: io.Writer接口实现，这里是fileLog类型
	// 参数2: 日志前缀，每条日志前都会添加此前缀
	// 参数3: 标准日志标志，包含时间、日期等信息
	log = stlog.New(fileLog(destination), "[go] - ", stlog.LstdFlags)
}

// RegisterHandlers 注册HTTP路由处理函数
// 这是日志服务的核心，设置HTTP接口用于接收日志请求
// 在服务启动时被调用，注册/log路径的处理函数
func RegisterHandlers() {
	// 注册/log路径的HTTP处理函数
	// 这是日志服务对外暴露的唯一接口
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		// 根据HTTP方法类型处理请求
		switch r.Method {
		case http.MethodPost: // 只处理POST请求
			// 读取请求体内容，这是要记录的日志消息
			msg, err := io.ReadAll(r.Body)

			// 错误处理：如果读取出错或内容为空，返回400错误
			// 日志记录需要有实际内容才有意义
			if err != nil || len(msg) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// 将消息写入日志文件
			// 将字节数据转换为字符串传递给write函数
			write(string(msg))

			// 默认返回200 OK状态码

		default: // 对于非POST请求，返回405方法不允许
			// 日志服务只接受POST方法，其他方法被拒绝
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

// write 将消息写入日志文件
// 这是一个内部函数，简化了日志记录过程
// 参数:
// - message: 要记录的日志消息
func write(message string) {
	// 使用全局日志记录器写入消息
	// Printf格式化输出，%v是值的默认格式
	// 添加换行符确保每条日志占一行
	log.Printf("%v\n", message)
}
