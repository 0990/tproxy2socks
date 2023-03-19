# tproxy2socks
将iptables(TPROXY)透明代理流量转换为socks5流量的golang版工具，类似工具有[ipt2socks](https://github.com/zfl9/ipt2socks),[redsocks](https://github.com/darkk/redsocks)

## 特性
* 支持转成sock4,socks4a流量
* 支持转成socks5(tcp&udp)流量

## 使用
```
tproxy2socks --listen=0.0.0.0:60080 --proxy=socks5://127.0.0.1:1080
```

### 参数
```
--listen 本地监听地址，格式为x.x.x.x:xx,默认为0.0.0.0:60080
--proxy sock5代理地址，格式为sock5://x.x.x.x:xx或sock4://x.x.x.x:xx,默认为socks5://127.0.0.1:1080
--udptimeout udp超时时间（单位秒)，默认为60s
--loglevel 日志打印等级，有debug,info,warn,error，默认为error
```



