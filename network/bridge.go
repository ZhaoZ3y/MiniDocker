package network

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
	"time"
)

type BridgeNetworkDriver struct {
}

// Name 返回网络驱动的名称
func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

// Create 创建一个新的桥接网络
func (d *BridgeNetworkDriver) Create(subnet string, name string) (*NetWork, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	n := &NetWork{
		Name:    name,
		IpRange: ipRange,
		Driver:  d.Name(),
	}
	// 初始化桥接
	err := d.initBridge(n)
	if err != nil {
		logrus.Errorf("初始化桥接（bridge）时出错: %v", err)
	}

	return n, err
}

// Delete 删除一个网络
func (d *BridgeNetworkDriver) Delete(network NetWork) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	// 删除桥接接口
	return netlink.LinkDel(br)
}

// Connect 连接网络和端点设备
func (d *BridgeNetworkDriver) Connect(network *NetWork, endpoint *Endpoint) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	// 创建一个新的接口属性
	la := netlink.NewLinkAttrs()
	la.Name = endpoint.ID[:5]
	la.MasterIndex = br.Attrs().Index

	// 设置端点设备
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}

	// 添加端点设备
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("添加端点设备时出错：%v", err)
	}

	// 启用端点设备
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("添加端点设备时出错：%v", err)
	}
	return nil
}

// Disconnect 断开网络和端点设备的连接
func (d *BridgeNetworkDriver) Disconnect(network NetWork, endpoint *Endpoint) error {
	// 断开连接的逻辑可根据需求实现
	return nil
}

// initBridge 初始化桥接网络
func (d *BridgeNetworkDriver) initBridge(n *NetWork) error {
	// 尝试获取桥接接口，如果已经存在则跳过
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("添加桥接时出错：%s，错误：%v", bridgeName, err)
	}

	// 设置桥接的IP
	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP

	// 分配桥接的IP地址
	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return fmt.Errorf("在桥接 [ %s ] 上分配地址时出错：%s，错误：%v", gatewayIP, bridgeName, err)
	}

	// 启用桥接接口
	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("设置桥接为启用时出错：%s，错误：%v", bridgeName, err)
	}

	// 设置iptables规则
	if err := setupIPTables(bridgeName, n.IpRange); err != nil {
		return fmt.Errorf("为 %s 设置 iptables 时出错：%v", bridgeName, err)
	}

	return nil
}

// deleteBridge 删除桥接
func (d *BridgeNetworkDriver) deleteBridge(n *NetWork) error {
	bridgeName := n.Name

	// 获取桥接接口
	l, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("获取名为 %s 的链接失败：%v", bridgeName, err)
	}

	// 删除桥接接口
	if err := netlink.LinkDel(l); err != nil {
		return fmt.Errorf("删除桥接接口 %s 时失败：%v", bridgeName, err)
	}

	return nil
}

// setInterfaceIP 设置接口的IP地址
func setInterfaceIP(name string, rawIP string) error {
	retries := 2
	var iface netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		logrus.Debugf("检索新的 bridge netlink 链接 [%s] 时出错……正在重试", name)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("放弃从 netlink 获取新的 bridge 链接，运行 [ip link] 来排查错误：%v", err)
	}
	// 解析IP地址
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}
	// 创建地址并添加到接口
	addr := &netlink.Addr{
		IPNet:     ipNet,
		Peer:      ipNet,
		Label:     "",
		Flags:     0,
		Scope:     0,
		Broadcast: nil,
	}
	return netlink.AddrAdd(iface, addr)
}

// setInterfaceUP 启用接口
func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("获取名为 [%s] 的链接时出错：%v", iface.Attrs().Name, err)
	}

	// 启用接口
	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("为 %s 启用接口时出错：%v", interfaceName, err)
	}
	return nil
}

// createBridgeInterface 创建桥接接口
func createBridgeInterface(bridgeName string) error {
	_, err := net.InterfaceByName(bridgeName)
	// 如果接口已经存在则返回
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}

	// 创建桥接接口
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	br := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("创建桥接 %s 时失败：%v", bridgeName, err)
	}
	return nil
}

// setupIPTables 设置iptables规则
func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	// 显示当前规则（调试用）
	showIptablesRules()

	// 清除可能存在的旧规则
	cleanCmd := fmt.Sprintf("iptables -t nat -F POSTROUTING")
	if _, err := execCommand(cleanCmd); err != nil {
		logrus.Warnf("清除旧规则失败: %v", err)
	}

	// 设置NAT规则 - 使用execCommand替代直接调用
	natCmd := fmt.Sprintf("iptables -t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE",
		subnet.String(), bridgeName)
	if _, err := execCommand(natCmd); err != nil {
		return fmt.Errorf("设置NAT规则失败: %v", err)
	}
	logrus.Infof("NAT规则已添加: %s", natCmd)

	// 添加FORWARD规则允许转发
	forwardCmd1 := fmt.Sprintf("iptables -A FORWARD -i %s -j ACCEPT", bridgeName)
	if _, err := execCommand(forwardCmd1); err != nil {
		logrus.Warnf("设置FORWARD规则1失败: %v", err)
	}

	// 添加反向FORWARD规则，允许相关连接返回
	forwardCmd2 := fmt.Sprintf("iptables -A FORWARD -o %s -m state --state RELATED,ESTABLISHED -j ACCEPT", bridgeName)
	if _, err := execCommand(forwardCmd2); err != nil {
		logrus.Warnf("设置FORWARD规则2失败: %v", err)
	}

	// 再次显示当前规则（确认规则已添加）
	showIptablesRules()

	return nil
}
