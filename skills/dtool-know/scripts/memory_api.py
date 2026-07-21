#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""dtool 知识片段相关接口示例"""

import os, sys
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../dtool-common/scripts'))

from api_common import call_api


def memory_fragment_create(folder_name, title, content):
    """
    创建新的知识片段

    folder_name 为文件夹标识名称（如 "fragments"），传入空字符串则自动归属默认文件夹。
    title 为知识片段标题。
    content 为知识片段内容（Markdown 格式）。
    """
    result = call_api("/api/MemoryFragmentCreate", {
        "folder_name": folder_name,
        "title": title,
        "content": content,
    })

    if result.get("code") == 0:
        data = result.get("data", {})
        print(f"创建成功: id={data.get('id')}, title={data.get('title')}")
    else:
        print(f"创建失败: {result.get('msg')}")
    return result


def memory_fragment_details_by_id(id):
    """
    通过片段ID获取知识片段详情

    id 为知识片段的唯一标识。
    返回片段的详细信息，包括标题、内容、创建时间等。
    """
    result = call_api("/api/MemoryFragmentInfo", {
        "id": id,
    })

    if result.get("code") == 0:
        data = result.get("data", {})
        print(f"获取成功: id={data.get('id')}, title={data.get('title')}")
        print(f"创建时间: {data.get('create_time')}")
        print(f"更新时间: {data.get('update_time')}")
    else:
        print(f"获取失败: {result.get('msg')}")
    return result


def memory_fragment_update_by_id(id, content, workflow_id):
    """
    通过片段ID更新知识片段内容（不会修改标题）

    id 为知识片段的唯一标识。
    workflow_id （工作流ID）为必传参数，后端会校验片段是否属于该工作流。
    """
    result = call_api("/api/MemoryFragmentSaveById", {
        "workflow_id": int(workflow_id),
        "id": id,
        "content": content,
    })

    if result.get("code") == 0:
        data = result.get("data", {})
        print(f"更新成功: id={data.get('id')}, title={data.get('title')}")
    else:
        print(f"更新失败: {result.get('msg')}")
    return result


if __name__ == "__main__":
    print("=== dtool 知识片段 API 示例 ===\n")

    # 获取知识片段详情
    # memory_fragment_details_by_id(
    #     "uuid-string",
    # )

    # memory_fragment_create(
    #     "fragments",
    #     "示例知识片段",
    #     "## 标题\n\n这是正文内容...",
    # )
    # memory_fragment_update_by_id(
    #     "uuid-string",
    #     "## 更新后的内容\n\n新的正文...",
    #     1,
    # )
