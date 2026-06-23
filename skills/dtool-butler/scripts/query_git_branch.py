#!/usr/bin/env python3
"""查询指定 Git 分组和项目的当前分支

用法:
    python query_git_branch.py <group_name> <project_name>
示例:
    python query_git_branch.py chatwiki dev4
"""
import sys
import api_common

MAX_ITERATIONS = 100  # 仅用于风格统一，实际无循环

def main():
    if len(sys.argv) != 3:
        print("用法: python query_git_branch.py <group_name> <project_name>")
        sys.exit(1)
    
    group_name = sys.argv[1]
    project_name = sys.argv[2]

    # 1. 获取所有 Git 配置
    config = api_common.call_api("/api/GitConfigList", {})
    if config.get("code") != 0:
        print(f"查询Git配置失败: {config.get('msg')}")
        sys.exit(1)

    git_list = config.get("data", {}).get("git_list", [])
    git_group_list = config.get("data", {}).get("git_group_list", [])

    # 查找分组
    target_group = None
    for g in git_group_list:
        if g.get("name") == group_name:
            target_group = g
            break

    if not target_group:
        print(f"未找到分组: {group_name}")
        sys.exit(1)

    # 查找项目
    target_item = None
    for item in git_list:
        if item.get("git_group_id") == target_group["id"] and item.get("name") == project_name:
            target_item = item
            break

    if not target_item:
        print(f"未在分组 '{group_name}' 下找到项目 '{project_name}'")
        sys.exit(1)

    # 2. 查询当前分支
    branch_result = api_common.call_api("/api/GitCurrentBranch", {"git_id": target_item["id"]})
    if branch_result.get("code") == 0:
        data = branch_result.get("data", "")
        print(f"项目 '{project_name}' (分组 '{group_name}') 的当前分支: {data}")
    else:
        print(f"查询分支失败: {branch_result.get('msg')}")

if __name__ == "__main__":
    main()