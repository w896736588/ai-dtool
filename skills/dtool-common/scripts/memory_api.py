#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""dtool 知识片段相关接口示例"""

from api_common import call_api


def memory_fragment_update_by_path(relative_path, content):
    """
    通过相对路径更新知识片段内容（不会修改标题）

    传入的是相对于 fragments/ 的路径。
    """
    filename = relative_path.replace("\\", "/").split("/")[-1]
    fragment_id = filename.rsplit(".", 1)[0] if "." in filename else filename

    result = call_api("/api/MemoryFragmentSave", {
        "id": fragment_id,
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
    # memory_fragment_update_by_path(
    #     "2026/05/uuid.md",
    #     "## 更新后的内容\n\n新的正文...",
    # )
