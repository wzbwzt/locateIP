# ip 定位-终端工具

## 支持csv文件导入
ip,count

## 支持指定ip定位

## 支持ip的访问count，来排序

## 倒叙排列,可指定top数量

> - 基于高德地图需要申请高德的开发者key  
> - 排序可以不指定key



```bash
locateIP

Usage:
  locateIP [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  locate      locate
  ls          ls

Flags:
      --ak string     第三方key
  -f, --file string   解析的csv文件路径
  -h, --help          help for locateIP
  -p, --plat string   第三方平台,枚举值eg.gaode/baidu 目前只支持gaode (default "gaode")

Use "locateIP [command] --help" for more information about a command.
```

- 指定ip

```bash
./locateIP --ak ******* locate  -c 10  --ip=192.178.92.239,101.20.65.59
```

- ls 排序

```bash
./locateIP ls -f ./ip.csv  -c 10
top: 10
{"ip":"110.191.215.169","count":165796}
{"ip":"183.162.192.46","count":151582}
{"ip":"120.133.133.87","count":126312}
{"ip":"183.210.164.58","count":82688}
{"ip":"117.189.162.255","count":58132}
{"ip":"119.145.115.118","count":50641}
{"ip":"117.183.139.108","count":47257}
{"ip":"117.36.103.186","count":39594}
{"ip":"183.248.4.120","count":38888}
{"ip":"117.39.60.222","count":37267}
```
- 指定文件
```bash
./locateIP locate  -f ./ip.csv  -c 10 --ak ******* -d

Top: 10
[1/1] Writing moshable file... 100% [==================] ( 6 B/s) 执行完毕

```
