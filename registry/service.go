package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

// ServicesURL 是注册中心服务的完整URL
// 所有服务注册和注销请求都发送到此URL
const ServicesURL = "http://localhost" + ServicePort + "/services"

// ServicePort 是注册中心服务监听的端口
// 微服务架构中，注册中心通常在固定端口提供服务
const ServicePort = ":3000"

// registry 结构体是整个服务注册中心的核心
// 它存储和管理所有已注册的微服务信息，并处理服务依赖关系
type registry struct {
	// registrations 存储所有已注册服务的信息
	// 这是注册中心的核心数据结构，包含系统中所有活跃服务
	registrations []Registration

	// mu 是读写互斥锁，保证对注册表的并发访问安全
	// 因为多个服务可能同时注册或注销
	mu *sync.RWMutex
}

// add 方法向注册表中添加新的服务
// 每当有新服务启动并注册时调用此方法
// 此方法还负责处理服务的依赖关系，实现服务发现功能
// 参数:
// - reg: 要添加的服务注册信息
// 返回:
// - error: 添加过程中的错误
func (r *registry) add(reg Registration) error {
	// 加锁，确保并发安全，防止多个服务同时修改注册表
	r.mu.Lock()

	// 添加新服务到注册表
	r.registrations = append(r.registrations, reg)

	// 操作完成后释放锁
	r.mu.Unlock()

	// 执行依赖推送机制
	// 当服务注册并声明依赖时，查找并通知它依赖服务的信息
	err := r.sendRequireServices(reg)
	// log服务通知需要log服务的服务
	r.notify(patch{
		Added: []patchEntry{
			{
				Name: reg.ServiceName,
				URL:  reg.ServiceURL,
			},
		},
	})
	return err
}

// log服务通知需要log服务的服务
func (r registry) notify(fullPatch patch) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, reg := range r.registrations {
		//使用协程并发处理每个服务  并发的发出通知
		go func(reg Registration) {
			for _, reqService := range reg.RequireServices {
				//创建一个patch对象，用于存储依赖更新信息
				p := patch{Added: []patchEntry{}, Removed: []patchEntry{}}
				sendUpdate := false
				//遍历fullPatch中的新增服务
				for _, added := range fullPatch.Added {
					//如果新增服务是当前服务所需的依赖
					if added.Name == reqService {
						//将新增服务添加到p中
						p.Added = append(p.Added, added)
						//设置发送更新标志
						sendUpdate = true
					}
				}
				for _, removed := range fullPatch.Removed {
					if removed.Name == reqService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}
				//如果需要发送更新
				if sendUpdate {
					//发送更新请求
					err := r.sendPatch(p, reg.ServiceUpdateURL)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}(reg)
	}
}

// sendRequireServices 实现服务依赖发现和通知
// 业务流程:
// 1. 检查新注册服务声明的依赖
// 2. 在注册表中查找匹配的依赖服务
// 3. 将找到的依赖服务信息发送给新服务
// 参数:
// - reg: 新注册的服务信息，包含其依赖需求
// 返回:
// - error: 处理过程中的错误
func (r registry) sendRequireServices(reg Registration) error {
	// 使用读锁访问注册表，允许并发读取
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 创建patch对象，用于存储依赖更新信息
	var p patch

	// 双重循环:
	// 外层循环遍历所有已注册服务
	// 内层循环遍历新服务声明的依赖
	// 目的是找到所有匹配的依赖服务
	for _, serviceReg := range r.registrations {
		for _, reqService := range reg.RequireServices {
			// 当找到匹配的依赖服务时
			if serviceReg.ServiceName == reqService {
				// 创建patchEntry并添加到patch中
				p.Added = append(p.Added, patchEntry{
					Name: serviceReg.ServiceName,
					URL:  serviceReg.ServiceURL,
				})
			}
		}
	}

	// 发送依赖更新通知
	// 将找到的依赖服务信息发送到新服务的更新端点
	err := r.sendPatch(p, reg.ServiceUpdateURL)
	if err != nil {
		return err
	}
	return nil
}

// sendPatch 将依赖更新信息发送到指定服务
// 通过HTTP POST请求将patch对象发送到服务的更新端点
// 参数:
// - p: 包含依赖更新信息的patch对象
// - url: 接收更新的服务端点URL
// 返回:
// - error: 发送过程中的错误
func (r registry) sendPatch(p patch, url string) error {
	// 将patch对象序列化为JSON
	d, err := json.Marshal(p)
	if err != nil {
		return err
	}

	// 发送HTTP POST请求
	// Content-Type为application/json
	_, err = http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}
	return nil
}

// remove 方法从注册表中移除服务
// 当服务关闭或需要注销时调用此方法
// 参数:
// - url: 要移除的服务URL
// 返回:
// - error: 移除过程中的错误或服务未找到错误
func (r *registry) remove(url string) error {
	// 查找匹配URL的服务
	for i := range reg.registrations {
		if reg.registrations[i].ServiceURL == url {
			r.notify(patch{
				Removed: []patchEntry{
					{
						Name: r.registrations[i].ServiceName,
						URL:  r.registrations[i].ServiceURL,
					},
				},
			})
			// 加锁确保并发安全
			r.mu.Lock()
			// 通过切片操作移除该服务
			reg.registrations = append(reg.registrations[:i], r.registrations[:i+1]...)
			r.mu.Unlock()
			return nil
		}
	}
	// 未找到匹配服务时返回错误
	return fmt.Errorf("service at url %s not found", url)
}

// 初始化全局注册表实例
// 这是注册中心的单例对象，存储所有服务信息
var reg = registry{
	registrations: make([]Registration, 0),
	mu:            new(sync.RWMutex),
}

// RegistryService 实现了http.Handler接口
// 处理所有服务注册和注销的HTTP请求
// 这是注册中心的HTTP入口点
type RegistryService struct{}

// ServeHTTP 实现http.Handler接口，处理HTTP请求
// 业务流程:
// 1. 接收服务注册(POST)或注销(DELETE)请求
// 2. 解析请求内容
// 3. 更新注册表
// 4. 处理依赖关系
// 参数:
// - w: HTTP响应写入器
// - r: HTTP请求对象
func (s RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 记录收到的请求
	log.Println("Request received")

	// 根据HTTP方法处理不同类型的请求
	switch r.Method {
	case http.MethodPost: // 处理服务注册请求
		// 解析Registration对象
		dec := json.NewDecoder(r.Body)
		var r Registration
		err := dec.Decode(&r)
		if err != nil {
			// 解析失败，返回400错误
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// 记录服务注册信息
		log.Printf("adding service: %v with URL: %v", r.ServiceName, r.ServiceURL)

		// 添加服务到注册表
		// 这会触发依赖处理过程
		err = reg.add(r)
		if err != nil {
			// 添加失败，返回400错误
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	case http.MethodDelete: // 处理服务注销请求
		// 读取请求体中的服务URL
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 提取服务URL并记录
		url := string(payload)
		log.Printf("Removing service at URL: %s", url)

		// 从注册表中移除服务
		err = reg.remove(url)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	default: // 不支持的HTTP方法
		// 返回405 Method Not Allowed错误
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
