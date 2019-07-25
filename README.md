# Gim
Gim 是一款高性能的即时通讯系统，它使用 Go（golang） 开发完成。

## 核心逻辑

![image](https://github.com/wangxiaoqiange/gim/blob/develop/gim.png)

## 功能列表

- 聊天室

    > 一个比较自由的空间，用户可以随意加入、退出聊天室。 典型应用场景如：公开课、网络直播。

- 用户资料托管

    > 服务端存储用户资料如：姓名、性别、年龄、城市、联系方式等。可用于消息的定向推送等。

- 消息推送服务

    > 向选定用户群推送消息，使命必达。

## 如何使用 ?

### 服务端

```bash
shell > docker pull wangxiaoqiang/gim:0.0.1

shell > docker run -d --name gim -v config.json:/code/src -p 8080:8080 -p 8081:8081 -p 8082:8082 wangxiaoqiang/gim:0.0.1
```

### 客户端

```bash

```
