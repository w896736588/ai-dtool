"""
发送钉钉群机器人文本消息。

用法:
    from send_dingtalk import send_dingtalk
    send_dingtalk(webhook_url='https://oapi.dingtalk.com/robot/send?access_token=xxx', secret='SEC...', content='消息内容')

钉钉自定义机器人文档参考:
    https://open.dingtalk.com/document/orgapp/custom-robots-send-group-chat-messages
"""

import base64
import hashlib
import hmac
import json
import time
import urllib.error
import urllib.request


def _build_sign(secret):
    """根据 secret 和时间戳生成签名。"""
    timestamp = str(round(time.time() * 1000))
    sign_str = f"{timestamp}\n{secret}"
    mac = hmac.new(
        secret.encode("utf-8"),
        sign_str.encode("utf-8"),
        hashlib.sha256,
    )
    sign = base64.b64encode(mac.digest()).decode("utf-8")
    return timestamp, sign


def send_dingtalk(webhook_url, content, secret=None):
    """向钉钉群发送文本消息。

    Args:
        webhook_url: 钉钉机器人 Webhook 地址，如 https://oapi.dingtalk.com/robot/send?access_token=xxx
        content: 消息文本内容
        secret: 机器人加签密钥（可选，如机器人未设置加签则不需要）
    """
    url = webhook_url.strip()

    if secret:
        timestamp, sign = _build_sign(secret.strip())
        url = f"{url}&timestamp={timestamp}&sign={sign}"

    data = {
        "msgtype": "text",
        "text": {
            "content": content,
        },
    }

    req = urllib.request.Request(
        url,
        data=json.dumps(data, ensure_ascii=False).encode("utf-8"),
        headers={"Content-Type": "application/json; charset=utf-8"},
        method="POST",
    )

    try:
        with urllib.request.urlopen(req, timeout=10) as resp:
            result = json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as e:
        raise RuntimeError(f"HTTP {e.code}: {e.read().decode('utf-8', errors='replace')}")
    except urllib.error.URLError as e:
        raise RuntimeError(f"请求失败: {e.reason}")

    errcode = result.get("errcode")
    if errcode != 0:
        raise RuntimeError(f"发送失败: errcode={errcode}, errmsg={result.get('errmsg')}")

    return result


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="发送钉钉群机器人文本消息")
    parser.add_argument("--webhook", required=True, help="钉钉机器人 Webhook 地址")
    parser.add_argument("--secret", default=None, help="机器人加签密钥（可选）")
    parser.add_argument("--content", required=True, help="消息文本内容")

    args = parser.parse_args()
    try:
        result = send_dingtalk(args.webhook, args.content, args.secret)
        print(f"发送成功: {json.dumps(result, ensure_ascii=False)}")
    except RuntimeError as e:
        print(f"错误: {e}")
        exit(1)
