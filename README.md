# My Docker

在主机上进入容器的网络命名空间
16300为容器的进程ID

```shell
nsenter -t 16300 -n ifconfig
nsenter -t 16300 -n ping 114.114.114.114
```

参考链接：[abcd](https://mp.weixin.qq.com/s/YUmAxbQs-KFQkKrOUTusZg).