# Openai-Balance

> 一个由 go 语言编写的 Openai 余额获取程序，只需要提供 Api Key 即可获取

## Why
Openai四月份的更新，导致credit_grants接口只能使用浏览器token访问，若使用Api Key则会返回:```Your request to GET /dashboard/billing/credit_grants must be made with a session key (that is, it can only be made from the browser). You made it with the following key type: secret.```，大意就是不再支持Api Key来调用这个接口了。

## How
参考网上大佬的方案，暂时只能使用subscription接口获取总额度，然后使用usage接口获取使用量，相减来得到余额。

## Usage
1.下载程序到服务器
```shell
wget https://github.com/LaoYutang/openai-balance/releases/download/release/openai-balance-linux-amd64 -O openai-balance
chmod +x openai-balance
```
2.使用获取余额
```shell
./openai-balance -k $你的Api Key
```
> 另外可以是用 -p 添加网络代理地址 例子：-p http://user@127.0.0.1:8888
