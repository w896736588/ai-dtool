# dtool-db 详细说明

## 必要约束

- 调用前，先向用户确认所需参数：`base_url`、`mysql_id`
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 数据库查询优先使用只读方式；涉及写入时必须确认影响范围

## 文件索引

- 数据库接口：`scripts/db_api.py`

## 命令行调用

```bash
python scripts/db_api.py tables --mysql-id 1 --base-url http://localhost:17170
python scripts/db_api.py structure --mysql-id 1 --table-name users
python scripts/db_api.py query --mysql-id 1 --sql "SELECT * FROM users LIMIT 10"
python scripts/db_api.py exec --mysql-id 1 --sql "UPDATE users SET enabled = 1 WHERE id = 1" --confirmed-write
```

`exec` 仅支持后端允许的 `INSERT` / `UPDATE`，且必须在用户明确确认影响范围后传入 `--confirmed-write`。
