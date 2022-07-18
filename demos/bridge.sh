# 1. 添加两个ns
[root@k8svip ~]# ip netns add ns01
[root@k8svip ~]# ip netns add ns02
[root@k8svip ~]# ip netns list
ns02
ns01
[root@k8svip ~]#

# 2. 添加两对网卡
[root@k8svip ~]# ip link add veth0 type veth peer name br0-veth0
[root@k8svip ~]# ip link add veth1 type veth peer name br0-veth1
[root@k8svip ~]#

# 3. 分别把veth0和veth1添加到ns01和02 network namespace中
[root@k8svip ~]# ip link set veth0 netns ns01
[root@k8svip ~]# ip link set veth1 netns ns02
[root@k8svip ~]#

# 4. 创建网卡
[root@k8svip ~]# ip link add name br0 type bridge
[root@k8svip ~]#

# 5. 把br0-veth0 br0-veth1桥接到br0并查看
[root@k8svip ~]# ip link set dev br0-veth0 master br0
[root@k8svip ~]# ip link set dev br0-veth1 master br0
[root@k8svip ~]# brctl show
bridge name bridge id STP enabled interfaces
br0 8000.4e9afbd7f2b7 no    br0-veth0 br0-veth1
[root@k8svip ~]#

# 6. 设置IP地址
[root@k8svip ~]# ip addr add 192.168.1.1/24 dev br0
[root@k8svip ~]# ip netns exec ns01 ifconfig veth0 192.168.1.2/24 up
[root@k8svip ~]# ip netns exec ns02 ifconfig veth1 192.168.1.3/24 up
[root@k8svip ~]# ip link set br0-veth0 up
[root@k8svip ~]# ip link set br0-veth1 up
[root@k8svip ~]# ip link set br0 up
[root@k8svip ~]#

# 7. 设置路由
[root@k8svip ~]# ip netns exec ns01 ip route add default via 192.168.1.1
[root@k8svip ~]# ip netns exec ns02 ip route add default via 192.168.1.1
[root@k8svip ~]# ip netns exec ns01 route -n
Kernel IP routing table
Destination Gateway Genmask Flags Metric Ref Use Iface
0.0.0.0    192.168.1.1    0.0.0.0    UG 0    0    0 veth0
192.168.1.0    0.0.0.0    255.255.255.0   U 0      0        0 veth0
[root@k8svip ~]#
[root@k8svip ~]# ip netns exec ns02 route -n
Kernel IP routing table
Destination Gateway Genmask Flags Metric Ref Use Iface
0.0.0.0         192.168.1.1     0.0.0.0         UG 0      0        0 veth1
192.168.1.0     0.0.0.0         255.255.255.0   U 0      0        0 veth1
[root@k8svip ~]#

# 8. 设置防火墙 SNAT 的 IP 伪装
[root@k8svip ~]# iptables -t nat -A POSTROUTING -s 192.168.1.0/24 -j MASQUERADE

# 9. 单独network namespace 内的IP ping不通的话，需要启动lo
[root@k8svip ~]# ip netns exec ns01 ip link set lo up
[root@k8svip ~]# ip netns exec ns02 ip link set lo up

# 最最重要的一步，一般linux会把bridge的fowrding禁用，所以还需要进行
          #- iptable的配置， iptables -A FORWARD -i bridgename -j ACCEPT
          #- 或者给bridge配置ip address，在container（netns）里面配置网关为bridge的ip address
[root@k8svip ~]# iptables -A FORWARD -i br0 -j ACCEPT


参考地址：https://www.modb.pro/db/50733