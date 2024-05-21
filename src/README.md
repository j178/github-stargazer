# 配置页面

用户安装 App 之后，会重定向到这个配置页面，用来配置消息的推送方式。

TODO: 这里应该有一个粗糙的原型图

## 账户切换

因为一个用户可以将 app 安装到自己的账号，以及多个组织的账号，每个账号独立配置，所以需要一个切换账户的 dropdown list。

相关接口:

- GET /api/installations 获取用户安装的账户列表
- GET /api/repos/:installationID 获取这个 installation 允许访问的 repo 列表，用于配置 allow_repos 和 mute_repos 时的 repo 选择

## 消息推送方式配置

配置信息格式：

- allow_repos: 只接受这些 repo 的 star 事件
- mute_repos: 忽略这些 repo 的 star 事件
- notify_settings: 消息推送方式列表
  - service: 消息推送服务，目前支持 telegram, discord, bark, webhook
  - 其他字段根据 service 的不同而不同

相关接口:

- GET /api/settings/:account 获取用户的配置信息
- POST /api/settings/:account 更新用户的配置信息
- DELETE /api/settings/:account 删除用户的配置信息
- POST /api/settings/check 检查配置信息是否正确
- POST /api/settings/test 发送测试消息

最关键的就是 POST /api/settings/:account，其他接口都是辅助的。示例：

```sh
curl --location 'https://github-stargazer.vercel.app/api/settings/j178' \
--header 'Cookie: session=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJqMTc4IiwiZXhwIjoxNjg4MzA3MzQxfQ.18C6qQ3E1ogMAz11MRy_ntDJ3gB6nn7GK2EmINCf2oc' \
--header 'Content-Type: application/json' \
--data '{
    "notify_settings": [
        {
            "service": "telegram",
            "chat_id": "379650434"
        },
        {
            "service": "discord",
            "webhook_id": "112493303843915455",
            "webhook_token": "2TuyPdT1Pty9BNfs5kqpamj5sH1pcJRSIhETrdZzmBLn7lQcbDrNdwFr0BigVcq47mf"
        },
        {
            "service": "bark",
            "key": "U4EbzbdojJSzyT8YDtDQrV"
        }
    ],
    "allow_repos": null,
    "mute_repos": null
}'
```

### Telegram Bot 关联

通过 POST /api/connect/telegram 生成一个 token，指引用户打开 https://t.me/gh_stargazer_bot?start=<token> 。
bot 会将用户的 telegram username, chat id 与当前的 account 关联起来，之后就可以通过 telegram bot 推送消息给用户。

前端通过轮询 GET /api/connect/telegram/:token 获取关联状态，如果关联成功，会返回用户的 telegram username, chat id。
