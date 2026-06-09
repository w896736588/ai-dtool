---
name: dtool-api
description: Use when working with the dtool API module to query collections, folders, and APIs, create or update APIs, import APIs, run APIs, manage environments, or adjust API structure.
---

# dtool-api

## 这个 skill 可以做什么

- 查询接口开发模块中的集合、文件夹、接口列表和接口详情
- 在目标文件夹中创建接口或按 URI 更新已有接口
- 批量导入整个文件夹的接口定义
- 查询、创建、更新集合环境和环境变量
- 运行接口、查看运行结果、提取返回字段、生成调用代码
- 删除集合、删除文件夹、删除接口、移动接口
- 在补接口文档前，先基于分支 diff 定位接口变更

## 必要约束

- 与用户交互时使用简体中文
- 调用 dtool 前，必须先向用户确认 `base_url` 和 `Token`
- 所有请求与响应默认按 UTF-8 处理
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 在未确认集合、文件夹、目标接口前，不直接写入或覆盖数据
- 批量覆盖、删除、移动前，先明确影响范围
- 读取接口说明时，优先看 `references/dtool-api-endpoints.md`
- 需要补充具体调用方式、字段结构或自动化流程时，再去看 `scripts/` 下脚本

## 细节位置

- 接口说明：`references/dtool-api-endpoints.md`
- 按 URI 创建或更新接口：`scripts/sync_api_by_uri.py`
