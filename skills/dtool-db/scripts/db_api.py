#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""dtool database API command line client."""

import argparse
import json
from urllib import error, request


ACTION_PATHS = {
    "tables": "/api/MysqlTables",
    "structure": "/api/MysqlTableStructure",
    "query": "/api/MysqlQuery",
    "exec": "/api/MysqlExec",
}


def call_api(base_url, path, payload, timeout):
    body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
    req = request.Request(
        url=f"{base_url.rstrip('/')}{path}",
        data=body,
        headers={"Content-Type": "application/json; charset=utf-8"},
        method="POST",
    )
    try:
        with request.urlopen(req, timeout=timeout) as resp:
            result = json.loads(resp.read().decode("utf-8"))
    except error.HTTPError as exc:
        response_body = exc.read().decode("utf-8", errors="replace")
        return {"code": -1, "msg": f"HTTP {exc.code}", "data": response_body}
    except Exception as exc:
        return {"code": -1, "msg": str(exc), "data": None}

    if "ErrCode" in result:
        result["code"] = result.get("ErrCode")
    if "ErrMsg" in result:
        result["msg"] = result.get("ErrMsg")
    if "Data" in result:
        result["data"] = result.get("Data")
    return result


def build_parser():
    parser = argparse.ArgumentParser(description="调用 dtool 数据库 API")
    parser.add_argument("action", choices=ACTION_PATHS)
    parser.add_argument("--base-url", default="http://localhost:17170")
    parser.add_argument("--mysql-id", required=True, help="数据库配置 ID（MySQL/PgSQL）")
    parser.add_argument("--table-name")
    parser.add_argument("--sql")
    parser.add_argument("--confirmed-write", action="store_true")
    parser.add_argument("--timeout", type=float, default=60)
    return parser


def main():
    args = build_parser().parse_args()
    payload = {"mysql_id": args.mysql_id}
    if args.action == "structure":
        if not args.table_name:
            raise SystemExit("structure action 需要 --table-name")
        payload["table_name"] = args.table_name
    elif args.action in {"query", "exec"}:
        if not args.sql:
            raise SystemExit(f"{args.action} action 需要 --sql")
        if args.action == "exec" and not args.confirmed_write:
            raise SystemExit("exec action 需要 --confirmed-write，且必须先获得用户明确确认")
        payload["sql"] = args.sql

    result = call_api(args.base_url, ACTION_PATHS[args.action], payload, args.timeout)
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
