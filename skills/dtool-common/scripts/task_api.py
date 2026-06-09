#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""dtool 任务相关接口示例"""

from api_common import call_api


def append_zcode_session_id(task_id, session_id):
    """向指定任务追加一个 zcode 对话 sessionId（末尾去重）"""
    result = call_api("/api/HomeTaskZcodeSessionIdAppend", {
        "id": task_id,
        "session_id": session_id,
    })
    if result.get("code") == 0:
        print(f"追加成功: task_id={task_id}, session_id={session_id}")
    else:
        print(f"追加失败: {result.get('msg')}")
    return result


if __name__ == "__main__":
    print("=== dtool 任务 API 示例 ===\n")
    # append_zcode_session_id(1, "171ea720-318a-4f68-bec9-b699097d3d80")
