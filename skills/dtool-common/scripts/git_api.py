#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""dtool Git 相关接口示例"""

from api_common import call_api


def git_upload_file(git_id, local_file_paths):
    """
    上传一个或多个本地文件到远程项目目录

    通过 git_id 获取 SSH 远程连接配置和 tbl_git 的 code_path（远程代码目录），
    将本地文件传输到 code_path/relative_file_path（已存在则覆盖）。
    """
    result = call_api("/api/GitUploadFile", {
        "git_id": git_id,
        "local_file_paths": local_file_paths,
    })
    if result.get("code") == 0:
        file_list = result.get("data", {}).get("list", [])
        for item in file_list:
            print(f"上传成功: {item.get('remote_path')}")
            print(f"  文件名: {item.get('file_name')}")
            print(f"  大小: {item.get('file_size')} 字节")
        print(f"共上传 {len(file_list)} 个文件")
    else:
        print(f"上传失败: {result.get('msg')}")
    return result


def git_current_branch_by_id(git_id):
    """通过 git_id 查询当前分支和远程跟踪分支"""
    result = call_api("/api/GitCurrentBranch", {
        "git_id": git_id,
    })
    if result.get("code") == 0:
        print(result.get("data", ""))
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


def git_pull(git_id):
    """通过 git_id 拉取当前分支最新代码"""
    result = call_api("/api/GitPull", {
        "git_id": git_id,
    })
    if result.get("code") == 0:
        output = result.get("data", "")
        if output:
            print(output)
        else:
            print("拉取完成")
    else:
        print(f"拉取失败: {result.get('msg')}")
    return result


def git_change_branch_by_id(git_id, branch_name):
    """通过 git_id 切换到指定分支"""
    result = call_api("/api/GitChangeBranchById", {
        "git_id": git_id,
        "branch_name": branch_name,
    })
    if result.get("code") == 0:
        output = result.get("data", "")
        if output:
            print(output)
        else:
            print(f"已切换到分支 {branch_name}")
    else:
        print(f"切换分支失败: {result.get('msg')}")
    return result


if __name__ == "__main__":
    print("=== dtool Git API 示例 ===\n")
    # git_upload_file(1, [
    #     {"full_file_path": "/home/user/src/config/config.yaml", "relative_file_path": "src/config/config.yaml"},
    #     {"full_file_path": "/home/user/src/data/data.json", "relative_file_path": "src/data/data.json"},
    # ])
    # git_current_branch_by_id("1")
    # git_pull("1")
    # git_change_branch_by_id("1", "master")
