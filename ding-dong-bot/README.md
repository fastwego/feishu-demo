# 如何在飞书平台上 5 分钟内打造一个叮咚机器人

## 在飞书注册一个机器人应用

- 配置名称、图标等基本信息,获取应用的 appid/appsecret
![](img/step-1-new-bot.png)

- 启动机器人功能
![](img/step-2-bot-enable.png)

- 配置机器人 回调 url ，获取 EncryptKey 和 Token

注意：此时需要校验 url 有效，请看代码部分

![](img/step-3-bot-event.png)

- 发布应用
![](img/step-4-bot-release.png)

## 安装 fastwego/feishu 开发 sdk

`go get -u github.com/fastwego/feishu`

## 开发机器人

### 配置

- 将飞书应用的配置更新到 `.env` 文件

- 编写代码：[main.go](main.go)

## 编译 & 部署 到服务器

`GOOS=linux go build`

`chmod +x ./ding-dong-bot && ./ding-dong-bot`

## 测试发送消息

应用发布后，可以在工作台看到入口

![](img/apps.jpg)


给机器人发送消息，机器人就会回复 `dong`

![](img/ding-dong.jpg)

## 结语

恭喜你！5分钟内就完成了一款飞书机器人开发

完整演示代码：[https://github.com/fastwego/feishu-demo](https://github.com/fastwego/feishu-demo)