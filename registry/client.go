package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"sync"
)

// RegisterService 向注册中心注册微服务
// 业务流程:
// 1. 解析更新URL并设置更新处理器
// 2. 将注册信息序列化为JSON
// 3. 发送POST请求到注册中心
// 4. 验证注册成功
// 参数:
// - r: 包含服务名称、URL和依赖信息的注册对象
// 返回:
// - error: 注册过程中的错误
func RegisterService(r Registration) error {
	// 解析ServiceUpdateURL，提取路径部分
	// 此URL将用于接收依赖服务更新通知
	serviceUpdateURL, err := url.Parse(r.ServiceUpdateURL)
	if err != nil {
		return err
	}

	// 注册HTTP处理器来接收依赖更新通知
	// 所有发送到ServiceUpdateURL的请求都会由serviceUpdateHandler处理
	http.Handle(serviceUpdateURL.Path, &serviceUpdateHandler{})

	// 创建一个字节缓冲区，用于存储JSON编码后的注册信息
	buf := new(bytes.Buffer)

	// 创建JSON编码器，将输出写入缓冲区
	enc := json.NewEncoder(buf)

	// 将注册信息编码为JSON
	err = enc.Encode(r)
	if err != nil {
		return err
	}

	// 发送HTTP POST请求到注册中心的/services端点
	// 携带JSON格式的注册信息作为请求体
	res, err := http.Post(ServicesURL, "application/json", buf)
	if err != nil {
		return err
	}

	// 检查响应状态码，确保注册成功
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register service. Registry service"+
			" responded with code %v", res.StatusCode)
	}
	return nil
}

// ShutdownService 向注册中心发送服务注销请求
// 服务关闭时调用此函数，从注册中心移除服务信息
// 参数:
// - url: 要注销的服务URL
// 返回:
// - error: 注销过程中的错误
func ShutdownService(url string) error {
	// 创建DELETE请求，携带服务URL作为请求体
	req, err := http.NewRequest(http.MethodDelete, ServicesURL,
		bytes.NewBuffer([]byte(url)))
	if err != nil {
		return err
	}

	// 发送请求到注册中心
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// 检查响应状态码，确保注销成功
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to deregister service. Registry "+
			"service responded with code %v", res.StatusCode)
	}
	return nil
}

// providers 结构管理服务依赖和服务发现
// 它存储每个服务类型的可用实例列表，并提供负载均衡功能
// 例如: LogService相对于GradingService就是一个provider
type providers struct {
	// services是服务类型到服务URL列表的映射
	// 例如: {"LogService": ["http://localhost:4000", "http://localhost:4001"]}
	services map[ServiceName][]string

	// mutex保护并发访问
	mutex *sync.RWMutex
}

// Update 处理依赖服务的更新通知
// 当接收到注册中心发送的patch对象时调用此方法
// 它会更新本地缓存的服务提供者列表
// 参数:
// - pat: 包含新增和移除服务的patch对象
func (p *providers) Update(pat patch) {
	// 加锁确保并发安全
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 处理新增的服务
	for _, patchEntry := range pat.Added {
		// 如果是首次添加此类型的服务，初始化切片
		if _, ok := p.services[patchEntry.Name]; !ok {
			p.services[patchEntry.Name] = make([]string, 0)
		}
		// 将服务URL添加到对应服务类型的列表中
		p.services[patchEntry.Name] = append(p.services[patchEntry.Name],
			patchEntry.URL)
	}

	// 处理移除的服务
	for _, patchEntry := range pat.Removed {
		if providersURLs, ok := p.services[patchEntry.Name]; ok {
			// 遍历URL列表，找到并移除匹配的URL
			for i := range providersURLs {
				if providersURLs[i] == patchEntry.URL {
					p.services[patchEntry.Name] = append(providersURLs[:i],
						providersURLs[i+1:]...)
					break
				}
			}
		}
	}
}

// get 根据服务名称获取一个可用的服务URL
// 如果有多个实例，会随机选择一个，实现简单的负载均衡
// 参数:
// - name: 服务名称
// 返回:
// - string: 服务URL
// - error: 查找过程中的错误
func (p providers) get(name ServiceName) (string, error) {
	// 获取指定服务类型的所有URL
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	providers, ok := p.services[name]
	if !ok {
		return "", fmt.Errorf("no providers available for service %v", name)
	}

	// 随机选择一个URL，实现简单的负载均衡
	idx := int(rand.Float32() * float32(len(providers)))
	return providers[idx], nil
}

// GetProvider 是get方法的公共包装器
// 允许外部代码获取服务URL而无需直接访问providers实例
// 参数:
// - name: 服务名称
// 返回:
// - string: 服务URL
// - error: 查找过程中的错误
func GetProvider(name ServiceName) (string, error) {
	return prov.get(name)
}

// 全局providers实例，存储本地缓存的服务信息
var prov = providers{
	services: make(map[ServiceName][]string),
	mutex:    new(sync.RWMutex),
}

// serviceUpdateHandler 处理来自注册中心的服务更新通知
// 当依赖服务发生变化时，注册中心会向此处理器发送更新
type serviceUpdateHandler struct{}

// ServeHTTP 实现http.Handler接口，处理依赖服务的更新通知
// 业务流程:
// 1. 验证请求方法
// 2. 解析patch对象
// 3. 更新本地服务提供者缓存
// 参数:
// - w: HTTP响应写入器
// - r: HTTP请求对象
func (suh serviceUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 只接受POST请求
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 解析请求体中的patch对象
	dec := json.NewDecoder(r.Body)
	var p patch
	err := dec.Decode(&p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("Update received %v\n", p)

	// 更新本地服务提供者缓存
	// 这会更新services映射，添加新的服务URL或移除不可用的服务
	prov.Update(p)
}
