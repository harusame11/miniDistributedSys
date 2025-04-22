package log

import (
	"My_mimiDistributed/registry"
	"bytes"
	"fmt"
	stlog "log"
	"net/http"
)

// SetClientLogger 设置客户端日志记录器
// 允许其他服务(如GradingService)通过HTTP发送日志到中央日志服务
// 业务流程:
// 1. 配置本地日志前缀，包含服务名称，便于日志来源识别
// 2. 设置输出到clientLogger，将日志通过HTTP发送到日志服务
// 参数:
// - serviceURL: 日志服务的URL，通过服务发现获取，例如:http://localhost:4000
// - clientService: 客户端服务的名称，用于标识日志来源
func SetClientLogger(serviceURL string, clientService registry.ServiceName) {
	// 设置日志前缀，包含服务名称，格式为:[ServiceName] -
	stlog.SetPrefix(fmt.Sprintf("[%v] - ", clientService))
	// 清除默认标志(时间日期等)，因为日志服务会添加这些信息
	stlog.SetFlags(0)
	// 将输出重定向到clientLogger，它会将日志发送到远程服务
	stlog.SetOutput(&clientLogger{url: serviceURL})
}

// clientLogger 实现io.Writer接口，用于客户端日志记录
// 它是标准日志库和远程日志服务之间的桥梁
// 当服务调用log.Print等函数时，日志内容会通过此结构发送到中央日志服务
type clientLogger struct {
	// 日志服务的URL，如http://localhost:4000
	url string
}

// Write 实现io.Writer接口，发送日志到远程日志服务
// 当客户端调用log.Print等函数时，最终会调用此方法
// 业务流程:
// 1. 将日志数据包装为HTTP请求
// 2. 发送POST请求到日志服务
// 3. 验证响应并返回结果
// 参数:
// - data: 要记录的日志数据
// 返回:
// - int: 写入的字节数
// - error: 写入过程中的错误
func (cl clientLogger) Write(data []byte) (int, error) {
	// 创建请求体缓冲区
	b := bytes.NewBuffer(data)
	// 发送POST请求到日志服务的/log端点
	res, err := http.Post(cl.url+"/log", "text/plain", b)
	if err != nil {
		// 网络错误或日志服务不可用时返回错误
		return 0, err
	}

	// 检查响应状态码，确保日志成功记录
	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to send log message, status: %v", res.StatusCode)
	}

	// 返回写入的数据长度和nil错误表示成功
	return len(data), nil
}
