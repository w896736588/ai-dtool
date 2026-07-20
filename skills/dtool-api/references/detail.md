# dtool-api 详细说明

## 必要约束

- 与用户交互时使用简体中文
- 调用 dtool 前，必须先确认 `base_url`
- 所有请求与响应默认按 UTF-8 处理
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 在未确认集合、文件夹、目标接口前，不直接写入或覆盖数据
- 批量覆盖、删除、移动前，先明确影响范围
- 读取接口说明时，优先看 `references/dtool-api-endpoints.md`
- 需要补充具体调用方式、字段结构或自动化流程时，再去看 `scripts/` 下脚本

## 文件索引

- 接口说明：`references/dtool-api-endpoints.md`
- 按 URI 同步接口：`scripts/sync_api_by_uri.py`
