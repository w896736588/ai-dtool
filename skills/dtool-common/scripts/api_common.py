#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
dtool 通用 API 基础封装

使用前请先向用户确认以下信息，替换下方占位值：
  - base_url: dtool 服务地址（如 http://192.168.1.100:17170）
  - token: 认证令牌
  - mysql_id: 数据库配置 ID（支持 MySQL 和 Pgsql）
"""

import json
from urllib import error, request

BASE_URL = "http://localhost:17170"  # TODO: 替换为用户提供的地址
TOKEN = ""  # TODO: 替换为用户提供的 Token
MYSQL_ID = ""  # TODO: 替换为用户提供的数据库配置 ID（支持 MySQL 和 Pgsql）


def call_api(path, payload):
    """通用 API 调用函数"""
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = request.Request(
        url=f"{BASE_URL}{path}",
        data=body,
        headers={"Content-Type": "application/json; charset=utf-8", "Token": TOKEN},
        method="POST",
    )
    try:
        with request.urlopen(req, timeout=60) as resp:
            result = json.loads(resp.read().decode("utf-8"))
            return normalize_response(result)
    except error.HTTPError as exc:
        body_text = exc.read().decode("utf-8", errors="replace")
        return {"code": -1, "msg": f"HTTP {exc.code}", "data": body_text}
    except Exception as exc:
        return {"code": -1, "msg": str(exc), "data": None}


def normalize_response(result):
    """将后端返回的 ErrCode/ErrMsg/Data 统一映射为 code/msg/data"""
    if "ErrCode" in result:
        result["code"] = result.get("ErrCode")
    if "ErrMsg" in result:
        result["msg"] = result.get("ErrMsg")
    if "Data" in result:
        result["data"] = result.get("Data")
    return result
