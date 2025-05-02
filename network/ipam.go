package network

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath = "/var/run/MiniDocker/network/ipam/subnet.json"

// IPAM （IP地址管理）结构体
type IPAM struct {
	SubnetAllocatorPath string             // 子网分配器路径
	Subnets             *map[string]string // 存储子网分配信息的映射
}

// 默认的IP地址分配器实例
var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

// load 加载已有的IP分配信息
func (ipam *IPAM) load() error {
	// 检查配置文件是否存在
	if _, err := os.Stat(ipam.SubnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在则直接返回
		} else {
			return err // 其他错误则返回
		}
	}
	// 打开配置文件
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}
	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}

	// 解析JSON数据
	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		logrus.Errorf("转储分配信息时出错：%v", err)
		return err
	}
	return nil
}

// dump 将分配信息保存到文件
func (ipam *IPAM) dump() error {
	// 获取配置文件所在的目录
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	// 如果目录不存在，则创建目录
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(ipamConfigFileDir, 0644)
		} else {
			return err
		}
	}
	// 打开文件准备写入
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}

	// 将子网分配信息转为JSON格式
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}

	// 写入文件
	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return err
	}

	return nil
}

// Allocate 分配一个IP地址
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 初始化子网分配信息
	ipam.Subnets = &map[string]string{}

	// 从文件加载分配信息
	err = ipam.load()
	if err != nil {
		logrus.Errorf("转储分配信息时出错：%v", err)
	}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	// 获取子网掩码的大小
	one, size := subnet.Mask.Size()

	// 如果该子网没有分配信息，则初始化分配信息
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}

	// 找到第一个可用的IP地址
	for c := range (*ipam.Subnets)[subnet.String()] {
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			ip = subnet.IP
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			ip[3] += 1
			break
		}
	}

	// 保存更新后的分配信息
	ipam.dump()
	return
}

// Release 释放一个已分配的IP地址
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	// 初始化子网分配信息
	ipam.Subnets = &map[string]string{}

	_, subnet, _ = net.ParseCIDR(subnet.String())

	// 加载已有分配信息
	err := ipam.load()
	if err != nil {
		logrus.Errorf("转储分配信息时出错：%v", err)
	}

	// 计算要释放的IP地址在子网中的位置
	c := 0
	releaseIP := ipaddr.To4()
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}

	// 标记该IP地址为未分配
	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)

	// 保存更新后的分配信息
	ipam.dump()
	return nil
}
