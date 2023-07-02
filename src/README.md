# 配置页面

用户安装账户之后，会重定向到这个配置页面，用来配置消息的推送方式。

TODO: 这里应该有一个粗糙的原型图

## 账户切换

因为一个用户可以将 app 安装到自己的账号，以及多个组织的账号，每个账号独立配置，所以需要一个切换账户的 dropdown list。

相关接口:

- GET /api/installations 获取用户安装的账户列表
- GET /api/repos/:installationID 获取这个 installation 允许访问的 repo 列表，用于配置消息推送方式时的 repo 选择

## 消息推送方式配置

配置信息格式：

- allow_repos: 只接受这些 repo 的 star 事件
- mute repos: 不接受这些 repo 的 star 事件
- notify_settings: 消息推送方式列表
  - service: 消息推送服务，目前支持 telegram, discord, bark, webhook
  - 其他字段根据 service 的不同而不同

相关接口:

- GET /api/settings/:account 获取用户的配置信息
- POST /api/settings/:account 更新用户的配置信息
- DELETE /api/settings/:account 删除用户的配置信息
- POST /api/settings/check 检查配置信息是否正确
- POST /api/settings/test 发送测试消息

### Telegram Bot 关联

通过 POST /api/connect/telegram 生成一个 jwt token，指引用户打开 https://t.me/gh_stargazer_bot 并发送这个 token。
bot 会将用户的 telegram username, chat id 与当前的 account 关联起来，之后就可以通过 telegram bot 推送消息给用户。

前端通过轮询 GET /api/connect/telegram 获取关联状态，如果关联成功，会返回用户的 telegram username, chat id。
