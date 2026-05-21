---
name: dtool-notify
description: Use when the task involves sending DingTalk (钉钉) group chat notifications — deploy success/failure alerts, automation task completion notices, CI/CD pipeline status messages, or any text notification via DingTalk custom robot webhook.
---

# dtool 钉钉通知技能

- 提供向钉钉群发送文本消息的能力，支持加签验证模式。
- 通过钉钉自定义机器人 Webhook 发送，一行 Python 调用即可。

## 强制约束

1. 调用前，必须向用户确认以下信息：
   - **Webhook 地址**：钉钉机器人的完整 Webhook URL，如 `https://oapi.dingtalk.com/robot/send?access_token=xxx`
   - **加签密钥**（`secret`）：机器人安全设置中"加签"方式下的密钥（如机器人未启用加签则不需要）
   - **消息内容**（`content`）：要发送的文本内容
2. 消息内容不支持 Markdown 渲染，仅纯文本（如需 Markdown 格式请使用 msgtype=markdown 的其他接口）。
3. 单条消息最大 20480 字符，超过会被截断。

## 调用方式

### 命令行使用

```bash
# 无加签
python skills/dtool-notify/scripts/send_dingtalk.py \
  --webhook 'https://oapi.dingtalk.com/robot/send?access_token=xxx' \
  --content '✅ 构建完成'

# 有加签
python skills/dtool-notify/scripts/send_dingtalk.py \
  --webhook 'https://oapi.dingtalk.com/robot/send?access_token=xxx' \
  --secret 'SECxxxx' \
  --content '部署成功，分支: master'
```

### Python 调用

```python
import sys
sys.path.insert(0, r'C:\work\self\cache_manager_api\skills\dtool-notify\scripts')
from send_dingtalk import send_dingtalk

# 无加签
send_dingtalk(
    webhook_url='https://oapi.dingtalk.com/robot/send?access_token=xxx',
    content='任务执行完成',
)

# 有加签
send_dingtalk(
    webhook_url='https://oapi.dingtalk.com/robot/send?access_token=xxx',
    content='任务执行完成',
    secret='SECxxxx',
)
```

## 参数说明

| 参数 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `webhook_url` / `--webhook` | string | 是 | 钉钉机器人 Webhook 完整地址 |
| `content` / `--content` | string | 是 | 消息文本内容，纯文本格式 |
| `secret` / `--secret` | string | 否 | 加签密钥，机器人安全设置中的 SEC 开头的字符串 |

## 加签说明

当钉钉机器人的安全设置选择"加签"方式时，需要在请求 URL 上追加 `timestamp` 和 `sign` 参数。签名计算方式：

1. 获取当前毫秒级时间戳 `timestamp`
2. 计算 HMAC-SHA256(`timestamp\n{secret}`, secret)
3. Base64 编码签名结果
4. 拼接到 Webhook URL：`原始URL&timestamp={timestamp}&sign={sign}`

脚本已内置此逻辑，只需传入 `secret` 参数即可。

## 返回结果

成功时返回 `{"errcode": 0, "errmsg": "ok"}`，失败抛出 `RuntimeError` 并包含 errcode 和 errmsg。

## 推荐场景

- 部署完成通知
- CI/CD 流水线结果推送
- 定时任务执行告警
- 自动化测试结果通知
