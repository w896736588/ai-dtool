#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""dtool 浏览器相关接口示例"""

from api_common import call_api


def browser_profile_capture_headers(smart_link_id, label, account="", open_type=0,
                                    reuse_if_open=True, enable_mcp=False):
    """
    打开浏览器配置并在登录后抓取首个接口请求头

    登录完成后刷新页面，抓取首个 xhr/fetch 接口请求的 headers，然后自动关闭浏览器。
    """
    result = call_api("/api/ai/browser/session/capture-headers", {
        "smart_link_id": smart_link_id,
        "label": label,
        "account": account,
        "open_type": open_type,
        "reuse_if_open": reuse_if_open,
        "enable_mcp": enable_mcp,
    })
    if result.get("code") == 0:
        headers = result.get("data", {}).get("headers", {})
        if headers:
            print("headers:")
            for key in sorted(headers.keys()):
                print(f"  {key}: {headers[key]}")
        else:
            print("headers 为空")
    else:
        print(f"抓取失败: {result.get('msg')}")
    return result


if __name__ == "__main__":
    print("=== dtool 浏览器 API 示例 ===\n")
    # browser_profile_capture_headers(12, "登录后首页", account="tester")
