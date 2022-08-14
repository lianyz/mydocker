# mydocker 自己动手实现容器

[toc]

## 编译运行环境
1. Ubuntu 20.04
2. CentOS Linux 8.5

## 基本用法

```shell
make init  # 完成初始化的准备工作

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
该链默认的policy是DROP,如果没有显式规则去允许bridge转发的包的话，该包会被丢弃. 参考链接：[why linux bridge doesn't work](https://superuser.com/questions/1211852/why-linux-bridge-doesnt-work).

有如下三种解决方案，本项目中采用的是第一种解决方案

1. 显示增加规则允许转发
```shell
# External traffic
iptables -t filter -A FORWARD -i testbridge -j ACCEPT
# Local traffic (since it doesn't pass the PREROUTING chain)
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
可以参考以下链接：[WTF addrtype in iptables manpage](https://www.linuxquestions.org/questions/linux-networking-3/wtf-addrtype-in-iptables-manpage-746659/), [Docker's NAT table output chain rule](https://stackoverflow.com/questions/26963362/dockers-nat-table-output-chain-rule).
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
在主机上进入容器的网络命名空间，容器进程的ID为16300

```shell
nsenter -t 16300 -n ifconfig
nsenter -t 16300 -n ping 114.114.114.114
```

### syscall.Exec语句后的log日志为什么没有输出
exec会执行参数指定的命令，但是并不会创建新的进程，只在当前进程空间内执行，即替换当前进程的执行内容，他们
重用同一个进程号PID，所以syscall.Exec只能是main函数的最后一条指令，它后面的代码不会被执行到。

在bash中执行exec ls, exec是用被执行的命令行替换掉当前的shell进程，且exec命令后的其他命令将不再执行。
例如在当前shell中执行exec ls，表示执行ls这条命令来替换当前的shell，即为执行完后会退出当前shell。

### Linux shell脚本中，$@和$#分别是什么意思
$@ 表示所有脚本参数的内容
$# 表示所有脚本参数的个数

例如有文件名为test.sh的脚本如下：

```shell
######################
!/bin/sh
echo "argc: $#"
echo "argv: $@"
######################
```

执行脚本：
```shell
./test.sh first_arg second_arg
```

执行结果如下：
```shell
./test.sh: line 2: !/bin/sh: No such file or directory
argc: 3
argv: first_arg second arg
```

### 导出alpine镜像
docker run -ti alpine sh
docker ps
docker export -o nginx.tar [container-id]

### 使用alpine镜像启动nginx

```shell
# 启动镜像
mydocker network create --driver bridge --subnet 192.168.10.1/24 br0
mydocker run -ti --name a --net br0 alpine sh

# 修改主机名配置文件
echo "alpinelinux" > /etc/hostname 
#使用新设置的主机名立刻生效
hostname -F /etc/hostname

# 配置DNS Server
echo "nameserver 8.8.8.8" > /etc/resolv.conf

# 更新apk的源为国内的源
cat > /etc/apk/repositories <<EOF
http://mirrors.ustc.edu.cn/alpine/v3.10/main
http://mirrors.ustc.edu.cn/alpine/v3.10/community
EOF

# 安装软件包
apk update
apk add nginx
apk add openrc


# 执行该命令时，当前目录不能为/dev，否则会报错误，导致nginx启动失败
openrc

 # You are attempting to run an openrc service on a
 # system which openrc did not boot.
 # You may be inside a chroot or you may have used
 # another initialization system to boot this system.
 # In this situation, you will get unpredictable results!
 # If you really want to do this, issue the following command:
touch /run/openrc/softlevel



# Tell openrc loopback and net are already there, since docker handles the networking
echo 'rc_provide="loopback net"' >> /etc/rc.conf

apk add curl

# 启动nginx服务
/etc/init.d/nginx start 

# 查看nginx服务状态
/etc/init.d/nginx status

# 测试试Nginx服务是否正常，返回nginx的404页面错误，表明服务已正常
curl 192.168.10.2

# 让nginx在前台运行
echo "daemon off;" >> /etc/nginx/nginx.conf

# 在主机进程中导出镜像至 ./nginx-alpine.tar
mydocker commit a nginx-alpine

# 在主机进程中运行新的镜像
d run -d --name bird -net br0 -p 8888:80 nginx-alpine nginx

# 在主机进程中查看nginx进程运行情况
ps -ef | grep nginx

# 进入容器内部
mydocker exec bird sh

# 查看内部IP，发现是192.168.10.3
curl 192.168.10.3

# 在主机进程测试（主机IP为192.168.34.2）
curl 192.168.34.2:8888
```



### 制作myflask镜像

```shell
d run -ti -net br0 --name myflask myalpine sh

# 进入容器后运行
apk add python3
cd /root
wget wget https://bootstrap.pypa.io/get-pip.py
python3 get-pip.py

pip install flask
pip install redis
apk add vim
touch /root/app.py

# 在主机进程中运行
cd /root
mydocker commit myflask myflask
```


### 运行redis和flask
```shell
mydocker network create --driver bridge --subnet 192.168.20.1/24 br0
mydocker run -d --name myredis -net br0 myredis /usr/bin/redis-server
mydocker logs myredis

# 显示结果为
1:C 13 Aug 2022 13:51:46.048 # oO0OoO0OoO0Oo Redis is starting oO0OoO0OoO0Oo
1:C 13 Aug 2022 13:51:46.048 # Redis version=5.0.11, bits=64, commit=23c8f9b2, modified=0, pid=1, just started
1:C 13 Aug 2022 13:51:46.048 # Warning: no config file specified, using the default config. In order to specify a config file use /usr/bin/redis-server /path/to/redis.conf
1:M 13 Aug 2022 13:51:46.049 * Increased maximum number of open files to 10032 (it was originally set to 1024).
1:M 13 Aug 2022 13:51:46.050 * Running mode=standalone, port=6379.
1:M 13 Aug 2022 13:51:46.050 # Server initialized
1:M 13 Aug 2022 13:51:46.051 * Ready to accept connections


mydocker exec myredis sh
ifconfig

# 显示的IP地址应为192.168.20.2

mydocker run -ti -net br0 --name myflask -p 5000:5000 myflask python /root/app.py
```




### 在alpine上安装并启动nginx后，执行cat /dev/null 命令，报错WARNING: ca-certificates.crt does not contain exactly one certificate or CRL: skipping


```shell
# mount

overlay on / type overlay (rw,relatime,lowerdir=/root/alpine,upperdir=/root/writeLayer/a,workdir=/root/work)
proc on /proc type proc (rw,nosuid,nodev,noexec,relatime)
tmpfs on /dev type tmpfs (rw,nosuid,mode=755,inode64)
```

在容器创建时，完成 mount tmpfs后，通过os.Create创建/dev/null文件   [container/init.go 103 line]


### 启动自己定义的容器

```shell
d run -ti --name bird2 -net br0 -p 8888:80 nginx-alpine nginx
```


参考链接：
1. [Container Network Is Simple](https://iximiuz.com/en/posts/container-networking-is-simple/)
2. [从0搭建Linux虚拟网络](https://zhuanlan.zhihu.com/p/199298498)
3. [bridge模式拆解实例分析](https://zhuanlan.zhihu.com/p/206512720)