# ipt2socks
将iptables(TPROXY)透明代理流量转换为socks5流量的golang版工具，类似工具有[zfl9/ipt2socks](https://github.com/zfl9/ipt2socks)

## 特性
* 支持转成sock4流量
* 支持转成socks5(tcp&udp)流量

## 使用
```
ipt2socks --listen=0.0.0.0:60080 --proxy=socks5://127.0.0.1:1080
```

### 参数
```
--listen 本地监听地址，格式为x.x.x.x:xx
--proxy sock5代理地址，格式为sock5://x.x.x.x:xx或sock4://x.x.x.x:xx
--udptimeout udp超时时间（单位秒)
--verbose 若指定此选项，则将会打印较为详尽的运行时日志
```



