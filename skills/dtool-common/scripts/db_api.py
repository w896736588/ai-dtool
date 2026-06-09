#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""dtool 数据库相关接口示例"""

from api_common import MYSQL_ID, TOKEN, call_api


def mysql_tables():
    """查询数据库配置对应的所有表（支持 MySQL 和 Pgsql）"""
    result = call_api("/api/MysqlTables", {
        "mysql_id": MYSQL_ID,
    })
    if result.get("code") == 0:
        table_list = result.get("data", {}).get("list", [])
        print(f"共 {len(table_list)} 张表:")
        for table in table_list:
            name = table.get("table_name", "")
            comment = table.get("table_comment", "")
            print(f"  {name}" + (f"  -- {comment}" if comment else ""))
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


def mysql_table_structure(table_name):
    """查询数据库表结构（支持 MySQL 和 Pgsql）"""
    result = call_api("/api/MysqlTableStructure", {
        "mysql_id": MYSQL_ID,
        "table_name": table_name,
    })
    if result.get("code") == 0:
        fields = result.get("data", {}).get("list", [])
        print(f"\n表 {table_name} 结构 ({len(fields)} 个字段):")
        print(f"  {'字段':<20} {'类型':<20} {'允许空':<6} {'键':<6} {'默认值':<10} {'备注'}")
        print(f"  {'-' * 20} {'-' * 20} {'-' * 6} {'-' * 6} {'-' * 10} {'-' * 20}")
        for field_info in fields:
            field = field_info.get("Field", "")
            field_type = field_info.get("Type", "")
            nullable = field_info.get("Null", "")
            key = field_info.get("Key", "")
            default = str(field_info.get("Default", ""))
            comment = field_info.get("Comment", "")
            print(f"  {field:<20} {field_type:<20} {nullable:<6} {key:<6} {default:<10} {comment}")
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


def mysql_query(sql):
    """执行数据库 SELECT 查询（支持 MySQL 和 Pgsql）"""
    result = call_api("/api/MysqlQuery", {
        "mysql_id": MYSQL_ID,
        "sql": sql,
    })
    if result.get("code") == 0:
        rows = result.get("data", {}).get("list", [])
        if not rows:
            print("查询结果为空")
        else:
            columns = list(rows[0].keys())
            widths = {}
            for column in columns:
                widths[column] = max(
                    len(str(column)),
                    max(len(str(row.get(column, ""))) for row in rows),
                )
            header = " | ".join(str(column).ljust(widths[column]) for column in columns)
            print(f"\n{header}")
            print("-" * len(header))
            for row in rows:
                line = " | ".join(str(row.get(column, "")).ljust(widths[column]) for column in columns)
                print(line)
            print(f"\n共 {len(rows)} 行")
    else:
        print(f"查询失败: {result.get('msg')}")
    return result


def mysql_exec(sql):
    """执行数据库写入操作（支持 MySQL 和 Pgsql）"""
    result = call_api("/api/MysqlExec", {
        "mysql_id": MYSQL_ID,
        "sql": sql,
    })
    if result.get("code") == 0:
        print(f"执行成功: {result.get('data', '')}")
    else:
        print(f"执行失败: {result.get('msg')}")
    return result


if __name__ == "__main__":
    if not TOKEN:
        print("请先设置 TOKEN（向用户确认后填入）")
        raise SystemExit(1)

    print("=== dtool 数据库 API 示例 ===\n")
    # mysql_tables()
    # mysql_table_structure("users")
    # mysql_query("SELECT * FROM users LIMIT 10")
    # mysql_exec("UPDATE users SET age = 21 WHERE name = 'test'")
