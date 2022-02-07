# httpclient
基于walk 实现网络工具

### 支持功能

|                |
| :------------- |
| HTTP客户端     |
| HTTP服务端     |
| TCP客户端      |
| TCP服务端      |
| 二维码生成工具 |
| 二维码识别工具 |

### HTTP客户端配置

说明：http客户端类的postman功能更轻量级，更简单，能够随机动态生成数据

```
url = http://localhost:8080
 #提交方式【 POST GET 】
method = POST
 #数据格式【application/x-www-form-urlencoded / application/json / text/xml / text/html / multipart/form-data】
headers.Content-Type =  application/json
 #暂停时间（秒）
sleep = 0 
 #随机数字生成（流水号）
number = int:999999
 #随机字符串
strvalue = str:32
 #19位长度日期时间
date1 = date:19
 #14位长度日期时间
date1 = date:14
 #8位长度日期
date2 = date:8
 #6位长度时间
date3 = date:6
 #从指定位置请求报文中的指定位置取值
up_value = req:0:txn_no
 #从响应报文中查找值，格式：res:(开头):0(从第0个交易结果中取值):<a>||</a>(以||分划取查找开始位置||结束位置)
result_val = res:0:<a>||</a>
 #报文结束符号
=END=
```

#### 固定配置-基于Content-Type动态创建报文体格式，json,表单格式

```
url = http://localhost:8080
 #提交方式【 POST GET 】
method = POST
 #数据格式【application/x-www-form-urlencoded / application/json / text/xml / text/html / multipart/form-data】
headers.Content-Type =  application/json
 #暂停时间（秒）
sleep = 0 
 #报文结束符号（可同时配置多个请求报文，以=END=分划）
=END=
```

#### 可选配置

```
#随机数字生成（流水号）
number = int:999999
 #随机字符串
strvalue = str:32
 #19位长度日期时间（2022-01-01 15:11:11）
date1 = date:19
 #14位长度日期时间（20220101151111）
date1 = date:14
 #8位长度日期（20220101）
date2 = date:8
 #6位长度时间（（151111））
date3 = date:6
 #从指定位置请求报文中的指定位置取值（从上一个报文中请求中获取txn_no的值）
up_value = req:0:txn_no
```

#### 原始上送

```
url = http://localhost:8080
 #提交方式【 POST GET 】
method = POST
 #数据格式【application/x-www-form-urlencoded / application/json / text/xml / text/html / multipart/form-data】
headers.Content-Type =  application/json
data = {"xxx":"xxx"}
```

### 压测

压测会按配置线程数并发数循环发送请求。

### 定义加密算法

可修改代码自定义验签规则，实现加密报文请求。如:微信，支付宝支付签名请求等。
