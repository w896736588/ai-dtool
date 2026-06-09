#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""dtool Docker 相关接口示例"""

from api_common import call_api


def docker_service_restart(docker_id, service):
    """
    重启指定 Docker Compose 中的某个服务

    只需传入 docker_id 和服务名，ssh_id 从配置中自动解析。
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


def docker_service_logs(docker_id, command):
    """
    查询 Docker Compose 服务日志

    command 必须以 "docker compose logs" 开头，且禁止使用 follow 模式。
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


if __name__ == "__main__":
    print("=== dtool Docker API 示例 ===\n")
    # docker_service_restart(1, "nginx")
    # docker_service_logs(1, "docker compose logs nginx")
    # docker_service_logs(1, "docker compose logs --tail 100 nginx")
