package network

import (
	"MiniDocker/container"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"
)

// NetWork 代表一个网络
type NetWork struct {
	Name    string     // 网络名称
	IpRange *net.IPNet // 网络地址范围
	Driver  string     // 网络驱动
}

// Endpoint 代表一个网络端点
type Endpoint struct {
	ID          string           `json:"ID"`          // 端点 ID
	Device      netlink.Veth     `json:"dev"`         // 设备
	IPAddress   net.IP           `json:"ip"`          // IP 地址
	MacAddress  net.HardwareAddr `json:"mac"`         // MAC 地址
	PortMapping []string         `json:"portMapping"` // 端口映射
	Network     *NetWork         // 所属网络
}

// Driver 定义了网络驱动的接口
type Driver interface {
	Name() string                                         // 驱动名称
	Create(subnet string, name string) (*NetWork, error)  // 创建网络
	Delete(network NetWork) error                         // 删除网络
	Connect(network *NetWork, endpoint *Endpoint) error   // 连接网络
	Disconnect(network NetWork, endpoint *Endpoint) error // 断开网络
}

var (
	defaultNetworkPath = "/var/run/MiniDocker/network/network/" // 默认的网络存储路径
	drivers            = map[string]Driver{}                    // 存储支持的网络驱动
	networks           = map[string]*NetWork{}                  // 存储已创建的网络
)

// dump 将网络信息序列化并保存到指定的路径
func (nw *NetWork) dump(dumpPath string) error {
	// 检查目录是否存在
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0644) // 创建目录
		} else {
			return err
		}
	}
	newPath := path.Join(dumpPath, nw.Name) // 拼接网络路径
	// 打开文件
	nwFile, err := os.OpenFile(newPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		logrus.Errorf("打开文件错误：%s ", err)
		return err
	}
	defer nwFile.Close()

	// 将网络信息序列化为 JSON
	nwJson, err := json.Marshal(nw)
	if err != nil {
		logrus.Errorf("序列化错误：%s ", err)
		return err
	}
	// 写入文件
	_, err = nwFile.Write(nwJson)
	if err != nil {
		logrus.Errorf("写入文件错误：%s ", err)
		return err
	}
	return nil
}

// load 从文件加载网络配置
func (nw *NetWork) load(dumpPath string) error {
	// 打开配置文件
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		logrus.Errorf("打开配置文件错误：%s ", err)
		return err
	}
	defer nwConfigFile.Close()
	// 读取文件内容
	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		logrus.Errorf("读取配置文件错误：%s ", err)
		return err
	}
	// 反序列化 JSON
	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		logrus.Errorf("反序列化错误：%s ", err)
		return err
	}
	return nil
}

// remove 删除网络配置文件
func (nw *NetWork) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name)) // 删除文件
	}
}

// Init 初始化网络管理，加载已存在的网络
func Init() error {
	var bridgeDriver = BridgeNetworkDriver{} // 创建桥接网络驱动
	drivers[bridgeDriver.Name()] = &bridgeDriver

	// 检查默认网络路径是否存在
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(defaultNetworkPath, 0644) // 创建默认网络目录
		} else {
			return err
		}
	}

	// 遍历默认网络路径，加载网络配置
	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") {
			return nil
		}
		_, nwName := path.Split(nwPath)
		nw := &NetWork{
			Name: nwName,
		}

		if err := nw.load(nwPath); err != nil {
			logrus.Errorf("加载网络错误: %s", err)
		}

		networks[nwName] = nw // 将加载的网络保存到 networks
		return nil
	})

	logrus.Infof("加载的网络: %v", networks)

	return nil
}

// CreateNetWork 创建新的网络，并将网络信息保存到文件中
func CreateNetWork(driver string, subnet string, name string) error {
	_, cidr, _ := net.ParseCIDR(subnet)
	// 通过IPAM分配IP地址
	gatewayIP, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = gatewayIP
	// 创建网络
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	// 将网络信息保存到文件中
	return nw.dump(defaultNetworkPath)
}

// Connect 将容器连接到指定的网络，配置容器的 IP 地址和端口映射
func Connect(networkName string, info *container.Info) error {
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("未找到网络: %s", networkName)
	}

	// 分配容器IP地址
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}

	// 创建网络端点
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", info.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: info.PortMapping,
	}
	// 调用网络驱动挂载和配置网络端点
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	// 配置容器的 IP 地址和路由
	if err = configEndpointIpAddressAndRoute(ep, info); err != nil {
		return err
	}

	// 配置端口映射
	return configPortMapping(ep, info)
}

// ListNetwork 列出所有网络的信息
func ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "名称（NAME）\tIP 范围（IpRange）\t驱动（Driver）\n")
	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver,
		)
	}
	if err := w.Flush(); err != nil {
		logrus.Errorf("刷新错误 %v", err)
		return
	}
}

// DeleteNetwork 删除指定的网络，并释放 IP 地址
func DeleteNetwork(networkName string) error {
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("不存在这样的网络: %s", networkName)
	}

	// 释放IP地址
	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("释放IP地址错误: %s", err)
	}

	// 删除网络
	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("删除网络错误: %s", err)
	}

	return nw.remove(defaultNetworkPath) // 删除网络配置文件
}

// configEndpointIpAddressAndRoute 配置网络端点的 IP 地址和路由信息
func configEndpointIpAddressAndRoute(ep *Endpoint, info *container.Info) error {
	// 获取网络设备
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("获取网络设备 %s 错误: %v", ep.Device.PeerName, err)
	}

	defer enterContainerNetns(&peerLink, info)()

	// 配置网络设备的 MAC 地址
	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress

	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}

	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}

	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")

	// 配置默认路由
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}

	return nil
}

// enterContainerNetns 将容器的网络设备从宿主机的网络命名空间移动到容器的网络命名空间
func enterContainerNetns(enLink *netlink.Link, info *container.Info) func() {
	// 获取容器的网络命名空间
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", info.Pid), os.O_RDONLY, 0)
	if err != nil {
		logrus.Errorf("获取容器网络命名空间错误, %v", err)
	}

	// 获取容器的网络命名空间文件描述符
	nsFD := f.Fd()
	// 锁定当前线程到操作系统
	runtime.LockOSThread()

	// 修改veth peer 另外一端移到容器的namespace中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		logrus.Errorf("设置网络命名空间错误, %v", err)
	}

	// 获取当前的网络命名空间
	origns, err := netns.Get()
	if err != nil {
		logrus.Errorf("获取当前网络命名空间错误, %v", err)
	}

	// 设置当前进程到新的网络命名空间，并在函数执行完成之后再恢复到之前的namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		logrus.Errorf("设置网络命名空间错误, %v", err)
	}
	return func() {
		// 恢复到原来的网络命名空间
		netns.Set(origns)
		// 关闭网络命名空间文件
		origns.Close()
		// 取消锁定当前线程到操作系统
		runtime.UnlockOSThread()
		// 关闭文件
		f.Close()
	}
}

// configPortMapping 配置容器的端口映射
func configPortMapping(ep *Endpoint, info *container.Info) error {
	// 遍历端口映射列表
	for _, pm := range ep.PortMapping {
		// 分割端口映射字符串
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			logrus.Errorf("端口映射格式错误，%v", pm)
			continue
		}
		// 将端口映射添加到 iptables
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		// 执行 iptables 命令进行端口映射
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			logrus.Errorf("iptables 错误，%v", output)
			continue
		}
	}
	return nil
}
