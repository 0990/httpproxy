## httpproxy
http代理服务器，支持http,https,支持自定义域名路由

## 使用
 根据平台在此选择下载文件，解压后直接执行二进制文件即可（linux平台需要加执行权限)

### 配置说明
 解压后同目录下的config.json是配置文件，各项配置字段说明如下  
 ```
  bind_addr 监听端口 
  hosts 哪些目标域名走本代理（*代表所有）
  next_proxy_addr 非hosts的域名走的代理服务器地址,可为空
  verbose 详细日志开关
  stats_addr 查看本代理代理数据地址，可为空
  pprof_addr pprof调试地址，可为空
```
### 自定义路由说明
在特定环境下使用，比如本机设定的http代理服务器A在公网,，这时需要访问的局域网内http服务（比如192.168.0.199:8080）也会被转到A,访问不了，这时可以在本机或局域网内再架设一台http代理服B，hosts配置中，配上192.168.0.199，next_proxy_addr配上A的地址,
本机的http代理地址改成B,这样192.168.0.199会由B正常访问到，而除此之外的hosts会由B将流量全转给A，A再代理出去
## TODO
* 请求头keepalive处理

## Thanks
[elazarl/goproxy](https://github.com/elazarl/goproxy)  

