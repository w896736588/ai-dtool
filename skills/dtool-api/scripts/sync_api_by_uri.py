#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
按 URI 在 dtool 接口开发模块中执行"导入或更新"。

输入 JSON 示例：
{
  "collection_name": "默认集合",
  "folder_name": "用户中心",
  "apis": [
    {
      "name": "用户登录",
      "method": "POST",
      "uri": "$Url$/v1/login",
      "protocol": "https",
      "desc": "登录接口",
      "headers": {"Content-Type": "application/json"},
      "query_params": [
        {"field": "version", "type": "string", "value": "v1", "description": "接口版本，固定传 v1，表示第一版协议"}
      ],
      "content_type": "application/json",
      "body_form": [],
      "body_json": "{\"username\":\"demo\",\"password\":\"123456\"}",
      "body_raw": "",
      "take_result": [
        {"key": "code", "type": "number", "desc": "状态码，0表示成功"},
        {"key": "data.token", "type": "string", "desc": "认证令牌"}
      ]
    }
  ]
}

注意：
- type 字段只接受: string / integer / float / boolean / file (禁止使用 int；bool 和 boolean 均可，推荐 boolean)
- 请求参数如果是常量、固定值、枚举值或布尔开关，必须在 description 中列出每个值和含义
- content_type 必须根据后端控制器实际代码判断，不得默认 application/json
- take_result 必须填写，描述接口返回字段含义
- base-url 和 Token 必须由用户提供，所有请求都会携带 Header: Token
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple
from urllib import request, error


def normalize_uri(uri: str) -> str:
    """规范化 URI，降低字符串差异导致的误判。"""
    value = (uri or "").strip()
    while len(value) > 1 and value.endswith("/"):
        value = value[:-1]
    return value.lower()


def post_json(base_url: str, token: str, path: str, payload: Dict[str, Any]) -> Dict[str, Any]:
    """发送 JSON POST 请求并返回响应 JSON。"""
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = request.Request(
        url=f"{base_url}{path}",
        data=body,
        headers={"Content-Type": "application/json; charset=utf-8", "Token": token},
        method="POST",
    )
    try:
        with request.urlopen(req, timeout=30) as resp:
            data = resp.read().decode("utf-8")
            return json.loads(data)
    except error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"HTTP {exc.code} {path} 失败: {detail}") from exc


def get_collections(base_url: str, token: str) -> List[Dict[str, Any]]:
    """获取集合树。"""
    response = post_json(base_url, token, "/api/Collections", {})
    data = response.get("data") or {}
    return data.get("list") or []


def find_collection(collections: List[Dict[str, Any]], collection_name: str) -> Optional[Dict[str, Any]]:
    """按集合名查找集合。"""
    for item in collections:
        if str(item.get("name", "")).strip() == collection_name.strip():
            return item
    return None


def find_folder(collection: Dict[str, Any], folder_name: str) -> Optional[Dict[str, Any]]:
    """在集合下按名称查找文件夹。"""
    for folder in collection.get("children") or []:
        if str(folder.get("name", "")).strip() == folder_name.strip():
            return folder
    return None


def create_folder(base_url: str, token: str, collection_id: int, folder_name: str) -> Dict[str, Any]:
    """新建文件夹。"""
    response = post_json(base_url, token, "/api/CreateDir", {"collection_id": collection_id, "name": folder_name})
    data = response.get("data") or {}
    if not data.get("id"):
        raise RuntimeError(f"创建文件夹失败，返回: {response}")
    return data


def get_folder_apis(base_url: str, token: str, folder_id: int) -> List[Dict[str, Any]]:
    """获取文件夹下接口列表。"""
    response = post_json(base_url, token, "/api/FolderDetail", {"dir_id": folder_id})
    data = response.get("data") or {}
    folder = data.get("dir") or {}
    return folder.get("children") or []


def build_create_api_payload(
    api_item: Dict[str, Any],
    collection_id: int,
    folder_id: int,
    existed_api_id: Optional[int],
) -> Dict[str, Any]:
    """构建 CreateApi 参数。"""
    payload: Dict[str, Any] = {
        "collection_id": collection_id,
        "folder_id": folder_id,
        "name": api_item.get("name", "未命名接口"),
        "method": api_item.get("method", "GET"),
        "url": api_item.get("uri") or api_item.get("url") or "",
        "protocol": api_item.get("protocol", "https"),
        "desc": api_item.get("desc", ""),
        "headers": api_item.get("headers", {}),
        "query_params": api_item.get("query_params", []),
        "content_type": api_item.get("content_type", ""),
        "body_form": api_item.get("body_form", []),
        "body_json": api_item.get("body_json", ""),
        "body_raw": api_item.get("body_raw", ""),
        "take_result": api_item.get("take_result", []),
    }
    if api_item.get("env_id"):
        payload["env_id"] = api_item["env_id"]
    if existed_api_id:
        payload["id"] = existed_api_id
    return payload


def sync_apis(base_url: str, token: str, collection_name: str, folder_name: str, apis: List[Dict[str, Any]], create_folder_if_missing: bool) -> Tuple[int, int]:
    """执行同步：按 URI 命中则更新，否则创建。"""
    collections = get_collections(base_url, token)
    collection = find_collection(collections, collection_name)
    if not collection:
        raise RuntimeError(f"集合不存在: {collection_name}")

    collection_id = int(collection.get("id") or 0)
    if collection_id <= 0:
        raise RuntimeError("集合 ID 无效")

    folder = find_folder(collection, folder_name)
    if not folder:
        if not create_folder_if_missing:
            raise RuntimeError(f"文件夹不存在: {folder_name}，可使用 --create-folder 自动创建")
        folder = create_folder(base_url, token, collection_id, folder_name)

    folder_id = int(folder.get("id") or 0)
    if folder_id <= 0:
        raise RuntimeError("文件夹 ID 无效")

    existed_apis = get_folder_apis(base_url, token, folder_id)
    uri_index: Dict[Tuple[str, str], Dict[str, Any]] = {}
    for api in existed_apis:
        uri_key = normalize_uri(str(api.get("url") or ""))
        method_key = str(api.get("method") or "GET").upper().strip()
        if uri_key:
            uri_index[(uri_key, method_key)] = api

    created = 0
    updated = 0
    for api_item in apis:
        raw_uri = str(api_item.get("uri") or api_item.get("url") or "").strip()
        if not raw_uri:
            raise RuntimeError(f"接口缺少 uri/url: {api_item}")

        method = str(api_item.get("method") or "GET").upper().strip()
        existed = uri_index.get((normalize_uri(raw_uri), method))
        existed_api_id = int(existed.get("id") or 0) if existed else None

        payload = build_create_api_payload(api_item, collection_id, folder_id, existed_api_id)
        post_json(base_url, token, "/api/CreateApi", payload)

        if existed_api_id:
            updated += 1
        else:
            created += 1

    return created, updated


def main() -> int:
    """脚本入口。"""
    parser = argparse.ArgumentParser(description="按 URI 同步 dtool 接口（命中更新，未命中创建）")
    parser.add_argument("--base-url", required=True, help="用户提供的 dtool 服务请求地址")
    parser.add_argument("--token", required=True, help="用户提供的 Header Token 值")
    parser.add_argument("--input", required=True, help="输入 JSON 文件路径")
    parser.add_argument("--create-folder", action="store_true", help="若文件夹不存在则自动创建")
    args = parser.parse_args()

    payload = json.loads(Path(args.input).read_text(encoding="utf-8"))
    collection_name = str(payload.get("collection_name") or "").strip()
    folder_name = str(payload.get("folder_name") or "").strip()
    apis = payload.get("apis") or []

    if not collection_name:
        raise RuntimeError("input 缺少 collection_name")
    if not folder_name:
        raise RuntimeError("input 缺少 folder_name")
    if not isinstance(apis, list) or not apis:
        raise RuntimeError("input 缺少 apis 或 apis 为空")
    token = args.token.strip()
    if not token:
        raise RuntimeError("缺少 Header Token 值")

    created, updated = sync_apis(args.base_url.rstrip("/"), token, collection_name, folder_name, apis, args.create_folder)
    print(json.dumps({"created": created, "updated": updated}, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        raise SystemExit(1)
