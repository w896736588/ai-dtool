#!/usr/bin/env python3
"""
查询指定Git仓库的当前分支及远程跟踪分支

用法: python query_git_branch.py <仓库名称> [--group <分组名称>]

参数:
  仓库名称    必需，要查询的Git仓库名称（支持部分匹配）
  --group     可选，按分组名称过滤以提高精确度

返回:
  打印仓库名称、代码路径、当前分支、远程跟踪分支

依赖:
  api_common (位于 dtool-common/scripts/)

实现:
  - GitConfigList：获取所有仓库配置，按名称+分组匹配目标仓库
  - GitCurrentBranch：通过SSH到远程服务器执行git show-branch获取当前分支信息
"""
import argparse
import sys
import os
import re

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../dtool-common/scripts'))
from api_common import call_api


def strip_ansi(text):
    """移除 ANSI 转义序列和终端控制字符"""
    return re.sub(r'\x1b\[[0-9;]*[a-zA-Z]|\x1b\][^\a]*\a?|\r', '', text)


def find_repo(repo_name, group_name=None):
    """从GitConfigList中查找匹配的仓库"""
    result = call_api("/api/GitConfigList", {})
    if result.get("code") != 0:
        print(f"Error: 查询Git配置失败: {result.get('msg')}", file=sys.stderr)
        sys.exit(1)
    data = result.get("data", {})
    git_list = data.get("git_list", [])
    groups = data.get("git_group_list", [])

    target_group_ids = []
    if group_name:
        for g in groups:
            if group_name.lower() in g.get('name', '').lower():
                target_group_ids.append(g.get('id'))
        if not target_group_ids:
            print(f"Error: 未找到包含 '{group_name}' 的分组", file=sys.stderr)
            sys.exit(1)

    matched = [repo for repo in git_list
               if repo_name.lower() in repo.get('name', '').lower()
               and (not target_group_ids or repo.get('git_group_id') in target_group_ids)]

    if not matched:
        if group_name:
            print(f"Error: 在分组 '{group_name}' 中未找到名称包含 '{repo_name}' 的仓库", file=sys.stderr)
        else:
            print(f"Error: 未找到名称包含 '{repo_name}' 的仓库", file=sys.stderr)
        sys.exit(1)

    return matched[0], groups, matched


def parse_branch_output(raw_output):
    """解析 GitCurrentBranch SSH 命令输出。

    输出格式示例:
      当前分支：
      feature_frog_xxx
      远程分支：
      d9660b3c...\trefs/heads/feature_frog_xxx
    """
    text = strip_ansi(raw_output)
    # 移除终端 prompt 行（以 $ 结尾或包含 ~: 的行）
    lines = []
    for line in text.split('\n'):
        line = line.strip()
        if not line or line.startswith('$ '):
            continue
        lines.append(line)

    current_branch = None
    remote_branch = None
    in_branch_section = False

    for i, line in enumerate(lines):
        if '当前分支' in line:
            in_branch_section = True
            # 下一行是当前分支名
            if i + 1 < len(lines):
                next_line = lines[i + 1]
                if next_line and '远程分支' not in next_line:
                    current_branch = next_line
            continue
        if '远程分支' in line:
            in_branch_section = True
            continue
        if in_branch_section and 'refs/heads/' in line:
            # 格式: hash\trefs/heads/branch_name
            parts = line.split('\t')
            if len(parts) >= 2:
                remote_branch = parts[1].replace('refs/heads/', '').strip()

    # 兜底：直接从全文匹配分支名
    if current_branch is None:
        for line in lines:
            if 'refs/heads/' in line and '\t' in line:
                parts = line.split('\t')
                if len(parts) >= 2:
                    current_branch = parts[1].replace('refs/heads/', '').strip()
                    break

    return current_branch, remote_branch


def get_current_branch(repo_id):
    """通过 SSH 查询远程仓库当前分支"""
    result = call_api("/api/GitCurrentBranch", {"git_id": str(repo_id)})
    if result.get("code") != 0:
        print(f"Error: 查询分支失败: {result.get('msg')}", file=sys.stderr)
        sys.exit(1)

    raw = result.get("data", "")
    if not raw:
        return None, None

    return parse_branch_output(raw)


def main():
    parser = argparse.ArgumentParser(description="查询指定Git仓库的当前分支及远程跟踪分支")
    parser.add_argument("repo_name", help="仓库名称（部分匹配）")
    parser.add_argument("--group", help="分组名称过滤（可选）")
    args = parser.parse_args()

    repo, groups, all_matched = find_repo(args.repo_name, args.group)

    if len(all_matched) > 1:
        group_map = {str(g.get('id')): g.get('name', '未知') for g in groups}
        print(f"提示：找到 {len(all_matched)} 个匹配仓库，已选择最匹配的：")
        for r in all_matched[:5]:
            gid = r.get('git_group_id')
            print(f"  - id={r.get('id')}, name={r.get('name')}, group={group_map.get(str(gid), '未知')}")
        print()

    current_branch, remote_branch = get_current_branch(repo.get("id"))

    print(f"仓库名称: {repo.get('name')}")
    print(f"代码路径: {repo.get('code_path', '未知')}")
    if current_branch:
        print(f"当前分支: {current_branch}")
    if remote_branch:
        print(f"远程跟踪分支: {remote_branch}")
    if not current_branch and not remote_branch:
        print("未找到当前分支信息")


if __name__ == "__main__":
    main()
