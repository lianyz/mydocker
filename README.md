# My Docker


## 基本用法

```shell
make       # 编译生成 ./bin/mydocker
make net   # 创建网络 testbridge, 网关ip为192.168.10.1
make d1    # 创建容器1 ip地址为 192.168.10.2
make d2    # 创建容器2 ip地址为 192.168.10.3
```

## 两个容器互相访问

在容器d1中访问d2
```shell
ping 192.168.10.3
```

在容器d2中访问d1
```shell
ping 192.168.10.2
```

## 从容器内访问外部网络

在容器d1或d2中访问外部网络
```shell
ping 114.114.114.114
```

## 从外部网络访问容器

在容器d1或d2中开启netcat
```shell
nc -lp 81   # d1
nc -lp 82   # d2
```

在主机上使用网桥地址访问
```shell
telnet 192.168.10.1 81
```

在主机上使用本机地址访问（假设本机地址为 192.168.34.2）
```shell
telnet 192.168.34.2 81
```

## 常见问题

### 容器之间无法互相访问或容器内无法访问外部网络

#### 原因1 容器的路由地址设置错误，应该为192.168.10.1

在容器d1中通过一下命令查看容器内的路由表
```shell
route -n
```

结果应为
```shell
Kernel IP routing table
Destination     Gateway         Genmask         Flags Metric Ref    Use Iface
0.0.0.0         192.168.10.1    0.0.0.0         UG    0      0        0 cif-20813
192.168.10.0    0.0.0.0         255.255.255.0   U     0      0        0 cif-20813
```

#### 原因2 主机iptables filter表中的forward规则检查不通过

当容器1向容器2发送ping包经过bridge转发时，这个包会路过iptables中filter表的FORWORD规则链，
该链默认的policy是DROP,如果没有显式规则去允许bridge转发的包的话，该包会被丢弃. 参考[why linux bridge doesn't work](https://superuser.com/questions/1211852/why-linux-bridge-doesnt-work).

有如下三种解决方案，本项目中采用的是第一种解决方案

1. 显示增加规则允许转发
```shell
iptables -t filter -A FORWARD -i testbridge -j ACCEPT
iptables -t filter -A FORWARD -o testbridge -j ACCEPT
```

2. 修改filter表的FORWORD的默认policy为ACCEPT
```shell
iptables -t filter -P FORWARD ACCEPT
```

3. 禁用iptables对bridge包的检查
```shell
sysctl net.bridge.bridge-nf-call-iptables=0
```

### 无法通过本机地址访问容器内的端口

原因：一般情况下，通过设置iptables的nat表中的PREROUTING规则链，将外部流量的目标地址修改为容器内部的IP地址，
但如果使用的是本机地址,是不会经过PREROUTING规则链的，因此需要设置OUTPUT规则链，如下：

```shell
iptables -t nat -A PREROUTING -p tcp -m tcp --dport 81 -j DNAT --to-destination 192.168.10.2:81
iptables -t nat -A OUTPUT -p tcp -m tcp --dport 81 -j DNAT --to-destination 192.168.10.2:81
```

### 什么是local类型的地址
local类型地址指的是本机的网卡所具有的地址，可以通过以下命令查看

```shell
ip route show table local type local
local 10.0.2.15 dev enp0s3 proto kernel scope host src 10.0.2.15 
local 127.0.0.0/8 dev lo proto kernel scope host src 127.0.0.1 
local 127.0.0.1 dev lo proto kernel scope host src 127.0.0.1 
local 172.17.0.1 dev docker0 proto kernel scope host src 172.17.0.1 
local 192.168.10.1 dev testbridge proto kernel scope host src 192.168.10.1 
local 192.168.16.128 dev vxlan.calico proto kernel scope host src 192.168.16.128 
local 192.168.34.2 dev enp0s8 proto kernel scope host src 192.168.34.2
```

### 如何通过进程号进入到容器的网络空间
在主机上进入容器的网络命名空间
16300为容器的进程ID

```shell
nsenter -t 16300 -n ifconfig
nsenter -t 16300 -n ping 114.114.114.114
```


参考链接：
1. [Container Network Is Simple](https://iximiuz.com/en/posts/container-networking-is-simple/)
2. [从0搭建Linux虚拟网络](https://zhuanlan.zhihu.com/p/199298498)
3. [bridge模式拆解实例分析](https://zhuanlan.zhihu.com/p/206512720)