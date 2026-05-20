#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
dtool 通用工具 API 调用示例
包含：Git文件上传、数据库表查询（MySQL/Pgsql）、表结构查询、SQL查询、Docker服务重启、Docker日志查询

使用前请先向用户确认以下信息，替换下方占位值：
  - base_url: dtool 服务地址（如 http://192.168.1.100:17170）
  - token: 认证令牌
  - mysql_id: 数据库配置 ID（支持 MySQL 和 Pgsql）
"""

import json
from urllib import request, error

# ============================================================
# 以下三个变量必须向用户确认后填入
# ============================================================
BASE_URL = "http://localhost:17170"  # TODO: 替换为用户提供的地址
TOKEN = ""                           # TODO: 替换为用户提供的 Token
MYSQL_ID = ""                        # TODO: 替换为用户提供的数据库配置 ID（支持 MySQL 和 Pgsql）


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
            return _normalize_response(result)
    except error.HTTPError as exc:
        body_text = exc.read().decode("utf-8", errors="replace")
        return {"code": -1, "msg": f"HTTP {exc.code}", "data": body_text}
    except Exception as exc:
        return {"code": -1, "msg": str(exc), "data": None}


def _normalize_response(result):
    """将后端返回的 ErrCode/ErrMsg/Data 统一映射为 code/msg/data"""
    if "ErrCode" in result:
        result["code"] = result.get("ErrCode")
    if "ErrMsg" in result:
        result["msg"] = result.get("ErrMsg")
    if "Data" in result:
        result["data"] = result.get("Data")
    return result


# ============================================================
# 1. 上传文件到远程项目
# 通过 git_id 获取 SSH 远程连接配置和 tbl_git 的 code_path，将本地文件传输到远程服务器
# ============================================================
def git_upload_file(git_id, local_file_paths):
    """
    上传一个或多个本地文件到远程项目目录

    通过 git_id 获取 SSH 远程连接配置和 tbl_git 的 code_path（远程代码目录），
    将本地文件传输到 code_path/relative_file_path（已存在则覆盖）。

    参数:
        git_id: Git 配置 ID（关联 tbl_git 表，用于获取远程连接信息和项目路径）
        local_file_paths: 文件路径数组，每个元素为字典:
            {"full_file_path": "本机绝对文件路径", "relative_file_path": "项目目录下的相对文件路径"}
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



# ============================================================
# 2. 查询数据库所有表（MySQL/Pgsql）
# ============================================================
def mysql_tables():
    """查询数据库配置对应的所有表（支持 MySQL 和 Pgsql）"""
    result = call_api("/api/MysqlTables", {
        "mysql_id": MYSQL_ID,
    })
    if result.get("code") == 0:
        table_list = result.get("data", {}).get("list", [])
        print(f"共 {len(table_list)} 张表:")
        for t in table_list:
            name = t.get("table_name", "")
            comment = t.get("table_comment", "")
            print(f"  {name}" + (f"  -- {comment}" if comment else ""))
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


# ============================================================
# 3. 查询数据库表结构（MySQL/Pgsql）
# ============================================================
def mysql_table_structure(table_name):
    """
    查询数据库表结构（支持 MySQL 和 Pgsql）

    参数:
        table_name: 表名
    """
    result = call_api("/api/MysqlTableStructure", {
        "mysql_id": MYSQL_ID,
        "table_name": table_name,
    })
    if result.get("code") == 0:
        fields = result.get("data", {}).get("list", [])
        print(f"\n表 {table_name} 结构 ({len(fields)} 个字段):")
        print(f"  {'字段':<20} {'类型':<20} {'允许空':<6} {'键':<6} {'默认值':<10} {'备注'}")
        print(f"  {'-'*20} {'-'*20} {'-'*6} {'-'*6} {'-'*10} {'-'*20}")
        for f in fields:
            field = f.get("Field", "")
            ftype = f.get("Type", "")
            null = f.get("Null", "")
            key = f.get("Key", "")
            default = str(f.get("Default", ""))
            comment = f.get("Comment", "")
            print(f"  {field:<20} {ftype:<20} {null:<6} {key:<6} {default:<10} {comment}")
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


# ============================================================
# 4. 执行数据库查询（仅 SELECT，支持 MySQL/Pgsql）
# ============================================================
def mysql_query(sql):
    """
    执行数据库 SELECT 查询（支持 MySQL 和 Pgsql）

    参数:
        sql: SELECT 语句

    示例:
        mysql_query("SELECT * FROM users LIMIT 10")
        mysql_query("SELECT COUNT(*) AS total FROM orders")
    """
    result = call_api("/api/MysqlQuery", {
        "mysql_id": MYSQL_ID,
        "sql": sql,
    })
    if result.get("code") == 0:
        rows = result.get("data", {}).get("list", [])
        if not rows:
            print("查询结果为空")
        else:
            # 打印表格式结果
            columns = list(rows[0].keys())
            # 计算每列宽度
            widths = {}
            for col in columns:
                widths[col] = max(
                    len(str(col)),
                    max(len(str(row.get(col, ""))) for row in rows)
                )
            # 表头
            header = " | ".join(str(col).ljust(widths[col]) for col in columns)
            print(f"\n{header}")
            print("-" * len(header))
            # 数据行
            for row in rows:
                line = " | ".join(str(row.get(col, "")).ljust(widths[col]) for col in columns)
                print(line)
            print(f"\n共 {len(rows)} 行")
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


# ============================================================
# 4.1 执行数据库写入（INSERT/UPDATE，支持 MySQL/Pgsql）
# ============================================================
def mysql_exec(sql):
    """
    执行数据库写入操作（支持 MySQL 和 Pgsql）

    允许 INSERT、UPDATE 语句，禁止 DROP、TRUNCATE、ALTER 等危险操作。

    参数:
        sql: INSERT 或 UPDATE 语句

    示例:
        mysql_exec("INSERT INTO users (name, age) VALUES ('test', 20)")
        mysql_exec("UPDATE users SET age = 21 WHERE name = 'test'")
    """
    result = call_api("/api/MysqlExec", {
        "mysql_id": MYSQL_ID,
        "sql": sql,
    })
    if result.get("code") == 0:
        print(f"执行成功: {result.get('data', '')}")
    else:
        print(f"执行失败: {result.get('msg')}")
    return result


# ============================================================
# 5. 重启 Docker Compose 服务
# ============================================================
def docker_service_restart(docker_id, service):
    """
    重启指定 Docker Compose 中的某个服务

    只需传入 docker_id（对应 dtool 中 Docker Compose 配置的 ID）和服务名，
    ssh_id 从配置中自动解析，无需手动指定。

    参数:
        docker_id: Docker Compose 配置 ID（整数）
        service: 要重启的服务名（如 "nginx"、"php-fpm"）

    示例:
        docker_service_restart(1, "nginx")
        docker_service_restart(3, "php-fpm")
    """
    result = call_api("/api/DockerServiceRestart", {
        "docker_id": docker_id,
        "service": service,
    })
    if result.get("code") == 0:
        print(f"服务 {service} 重启成功")
    else:
        print(f"重启失败: {result.get('msg')}")
    return result


# ============================================================
# 6. 查询 Docker Compose 服务日志
# ============================================================
def docker_service_logs(docker_id, command):
    """
    查询 Docker Compose 服务日志

    通过 docker_id 自动解析 SSH 连接，在 compose yml 目录下执行用户提供的 logs 命令。
    command 必须以 "docker compose logs" 开头，否则接口会拒绝执行。

    参数:
        docker_id: Docker Compose 配置 ID（整数）
        command: 日志查询命令，必须以 "docker compose logs" 开头

    示例:
        docker_service_logs(1, "docker compose logs nginx")
        docker_service_logs(1, "docker compose logs --tail 100 nginx")
        docker_service_logs(3, "docker compose logs --since 30m nginx php-fpm")
    """
    if not command.startswith("docker compose logs"):
        print("command 必须以 'docker compose logs' 开头")
        return {"code": -1, "msg": "command 必须以 'docker compose logs' 开头", "data": None}
    if " -f" in command or " --follow" in command:
        print("禁止使用 -f / --follow 参数，会导致持续输出")
        return {"code": -1, "msg": "禁止使用 -f / --follow 参数", "data": None}
    result = call_api("/api/DockerServiceLogs", {
        "docker_id": docker_id,
        "command": command,
    })
    if result.get("code") == 0:
        logs_content = result.get("data", {}).get("logs", "")
        if logs_content:
            print(logs_content)
        else:
            print("日志为空")
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


# ============================================================
# 7. 通过 git_id 查询当前分支
# ============================================================
def git_current_branch_by_id(git_id):
    """
    通过 git_id 查询当前分支和远程跟踪分支

    参数:
        git_id: Git 配置 ID（关联 tbl_git 表）
    """
    result = call_api("/api/GitCurrentBranch", {
        "git_id": git_id,
    })
    if result.get("code") == 0:
        print(result.get("data", ""))
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


# ============================================================
# 8. 拉取当前分支最新代码
# ============================================================
def git_pull(git_id):
    """
    通过 git_id 拉取当前分支最新代码

    参数:
        git_id: Git 配置 ID（关联 tbl_git 表）
    """
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


# ============================================================
# 9. 通过 git_id 切换分支
# ============================================================
def git_change_branch_by_id(git_id, branch_name):
    """
    通过 git_id 切换到指定分支(切换可能持续时间比较长，请等待返回，最长等待30分钟)

    参数:
        git_id: Git 配置 ID（关联 tbl_git 表）
        branch_name: 要切换到的分支名（如 master、dev、feature_xxx）
    """
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


# ============================================================
# 10. 网页截图
# ============================================================
def screenshot(url, full_page=False, width=1920, height=1080, timeout=30, selector="", save_path=""):
    """
    对指定网页进行截图，返回 base64 编码的 PNG 图片

    参数:
        url: 目标网页地址 (必填)
        full_page: 是否截取完整页面 (默认 False，仅截取可视区域)
        width: 视口宽度 (默认 1920)
        height: 视口高度 (默认 1080)
        timeout: 导航超时秒数 (默认 30)
        selector: CSS 选择器，截取指定元素 (可选)
        save_path: 保存为本地文件的路径 (可选，不填则不保存)
    """
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
            import base64
            with open(save_path, "wb") as f:
                f.write(base64.b64decode(image_base64))
            print(f"截图已保存到: {save_path}")
        else:
            print(f"截图成功 (url={data.get('url')}, "
                  f"full_page={data.get('full_page')}, "
                  f"尺寸={data.get('width')}x{data.get('height')})")
    else:
        print(f"截图失败: {result.get('msg')}")
    return result


# ============================================================
# 11. 打开浏览器配置并在登录后抓取首个接口请求头
# ============================================================
def browser_profile_capture_headers(smart_link_id, label, account="", open_type=0,
                                    reuse_if_open=True, enable_mcp=False):
    """
    使用与 browser_profile_open 一致的参数，在登录完成后刷新页面，
    抓取首个 xhr/fetch 接口请求的 headers，然后自动关闭浏览器。

    参数:
        smart_link_id: 自定义网页配置 ID
        label: 要打开的链接标签名
        account: 账号名（可选）
        open_type: 打开类型（可选）
        reuse_if_open: 如果已打开是否复用（可选）
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


# ============================================================
# 12. 更新知识片段（按文件路径）
# 知识片段以 Markdown 文件存储在 fragments/ 目录下，按
# {年份}/{月份}/{uuid}.md 的目录结构组织。通过传入相对于
# fragments/ 的路径定位片段并更新其内容。
# ============================================================
def memory_fragment_update_by_path(relative_path, content):
    """
    通过相对路径更新知识片段内容（不会修改标题）

    知识片段以 Markdown 文件存储，此接口通过传入相对于知识片段文件夹
    （fragments/）的路径定位片段，然后更新其内容。
    注意：不传 title 参数，禁止修改片段标题。

    参数:
        relative_path: 相对于知识片段文件夹（fragments/）的路径（必填）
                       如 "2026/05/a59db79a-3e4d-4f37-a02d-1bf87cc0c590.md"
        content: 新的 Markdown 正文内容（必填），支持以下占位符：
                 - {需求文档纯文本文件相对地址}: 后端自动替换为 relative_path 的值

    示例:
        memory_fragment_update_by_path(
            "2026/05/a59db79a-3e4d-4f37-a02d-1bf87cc0c590.md",
            "## 更新后的内容\\n\\n新的正文..."
        )
    """
    # 从文件路径中提取片段 ID（文件名去掉 .md 后缀）
    filename = relative_path.replace("\\", "/").split("/")[-1]
    fragment_id = filename.rsplit(".", 1)[0] if "." in filename else filename

    payload = {
        "id": fragment_id,
        "content": content,
    }

    result = call_api("/api/MemoryFragmentSave", payload)
    if result.get("code") == 0:
        data = result.get("data", {})
        print(f"更新成功: id={data.get('id')}, title={data.get('title')}")
    else:
        print(f"更新失败: {result.get('msg')}")
    return result




# ============================================================
# 12. 向任务追加 zcode 对话 sessionId
# ============================================================
def append_zcode_session_id(task_id, session_id):
    """
    向指定任务追加一个 zcode 对话 sessionId（末尾去重）

    参数:
        task_id: 首页任务 ID（整数）
        session_id: zcode 对话的 sessionId（字符串）

    示例:
        append_zcode_session_id(1, "171ea720-318a-4f68-bec9-b699097d3d80")
    """
    result = call_api("/api/HomeTaskZcodeSessionIdAppend", {
        "id": task_id,
        "session_id": session_id,
    })
    if result.get("code") == 0:
        print(f"追加成功: task_id={task_id}, session_id={session_id}")
    else:
        print(f"追加失败: {result.get('msg')}")
    return result


# ============================================================
# 使用示例
# ============================================================
if __name__ == "__main__":
    # 检查配置
    if not TOKEN:
        print("请先设置 TOKEN（向用户确认后填入）")
        exit(1)

    print("=== dtool 通用工具 API 示例 ===\n")
    # 示例1: 上传文件到远程项目（需提供 git_id）
    # git_upload_file(1, [
    #     {"full_file_path": "/home/user/src/config/config.yaml", "relative_file_path": "src/config/config.yaml"},
    #     {"full_file_path": "/home/user/src/data/data.json", "relative_file_path": "src/data/data.json"},
    # ])

    # 示例2: 查询数据库所有表（需设置 MYSQL_ID）
    # mysql_tables()

    # 示例3: 查询表结构（需设置 MYSQL_ID）
    # mysql_table_structure("users")

    # 示例4: 执行 SELECT 查询（需设置 MYSQL_ID）
    # mysql_query("SELECT * FROM users LIMIT 10")

    # 示例5: 重启 Docker Compose 服务（需提供 docker_id 和服务名）
    # docker_service_restart(1, "nginx")

    # 示例6: 查询 Docker Compose 服务日志（需提供 docker_id 和 command）
    # docker_service_logs(1, "docker compose logs nginx")
    # docker_service_logs(1, "docker compose logs --tail 100 nginx")

    # 示例7: 查询当前分支（需提供 git_id）
    # git_current_branch_by_id("1")

    # 示例8: 拉取当前分支最新代码（需提供 git_id）
    # git_pull("1")

    # 示例9: 切换分支（需提供 git_id 和目标分支名）
    # git_change_branch_by_id("1", "master")
    # git_change_branch_by_id("1", "dev")

    # 示例10: 网页截图
    # screenshot("https://www.baidu.com")
    # screenshot("https://www.baidu.com", full_page=True, save_path="page.png")
    # screenshot("https://www.baidu.com", selector="#main-content", save_path="element.png")

    # 示例11: 登录后刷新页面，抓取首个接口请求头
    # browser_profile_capture_headers(12, "登录后首页", account="tester")

    # 示例12: 按路径更新知识片段
    # memory_fragment_update_by_path(
    #     "2026/05/uuid.md",
    #     "## 更新后的内容\n\n新的正文...",
    # )
