# 1.添加两个network namespace
ip netns add ns01
ip netns add ns02
ip netns list

# 2.添加两对网卡
ip link add veth0 type veth peer name br0-veth0
ip link add veth1 type veth peer name br0-veth1

# 3.分别把veth0和veth1添加到 ns01 和 ns02 中
ip link set veth0 netns ns01
ip link set veth1 netns ns02

# 4.创建网桥
ip link add br0 type bridge

# 5.把br0-veth0、br0-veth1桥接到br0并查看
ip link set dev br0-veth0 master br0
ip link set dev br0-veth1 master br0
brctl show

# 6.设置IP地址
ip addr add 192.168.1.1/24 dev br0
ip netns exec ns01 ifconfig veth0 192.168.1.2/24 up
ip netns exec ns02 ifconfig veth1 192.168.1.3/24 up
ip link set br0-veth0 up
ip link set br0-veth1 up
ip link set br0 up

# 7.设置路由
ip netns exec ns01 ip route add default via 192.168.1.1
ip netns exec ns02 ip route add default via 192.168.1.1
ip netns exec ns01 route -n
ip netns exec ns02 route -n

# 8.设置防火墙 SNAT 的 IP 伪装
iptables -t nat -A POSTROUTING -s 192.168.1.0/24 -j MASQUERADE

# 9.启动bridge的forward功能
iptables -A FORWARD -i br0 -j ACCEPT

# 10.启动各namespace中的lo接口
ip netns exec ns01 ip link set lo up
ip netns exec ns02 ip link set lo up
