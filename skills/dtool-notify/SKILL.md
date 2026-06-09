---
name: dtool-notify
description: Use when the task involves sending DingTalk group notifications through a custom robot webhook.
---

# dtool-notify

## 这个 skill 可以做什么

- 发送钉钉群文本通知
- 支持普通 Webhook 模式
- 支持带签名密钥的安全模式
- 适用于部署通知、自动化任务通知、CI/CD 结果通知

## 必要约束

- 调用前，先向用户确认 `webhook`、消息内容，以及是否需要 `secret`
- 该 skill 默认发送纯文本消息，不负责 Markdown 富文本
- 单条消息不宜过长，超长内容应先压缩
- 需要具体调用参数或脚本入口时，再去看 `scripts/send_dingtalk.py`

## 细节位置

- 钉钉通知脚本：`scripts/send_dingtalk.py`
