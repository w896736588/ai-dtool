#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""dtool 截图接口示例"""

import base64

from api_common import call_api


def screenshot(url, full_page=False, width=1920, height=1080, timeout=30, selector="", save_path=""):
    """对指定网页进行截图，返回 base64 编码的 PNG 图片"""
    payload = {
        "url": url,
        "full_page": full_page,
        "width": width,
        "height": height,
        "timeout": timeout,
    }
    if selector:
        payload["selector"] = selector

    result = call_api("/api/Screenshot", payload)
    if result.get("code") == 0:
        data = result.get("data", {})
        image_base64 = data.get("image", "")
        if save_path and image_base64:
            with open(save_path, "wb") as file_obj:
                file_obj.write(base64.b64decode(image_base64))
            print(f"截图已保存到: {save_path}")
        else:
            print(
                f"截图成功 (url={data.get('url')}, "
                f"full_page={data.get('full_page')}, "
                f"尺寸={data.get('width')}x{data.get('height')})"
            )
    else:
        print(f"截图失败: {result.get('msg')}")
    return result


if __name__ == "__main__":
    print("=== dtool 截图 API 示例 ===\n")
    # screenshot("https://www.baidu.com")
    # screenshot("https://www.baidu.com", full_page=True, save_path="page.png")
