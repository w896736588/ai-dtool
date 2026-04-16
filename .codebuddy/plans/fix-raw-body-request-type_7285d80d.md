---
name: fix-raw-body-request-type
overview: 修复 POST 请求使用 raw/text/plain 类型 body 时报错"不支持的请求类型"的问题，补全前后端数据链路中缺失的 body_raw 支持
todos:
  - id: fix-api-run
    content: 修复 api.go Run() 方法，增加 raw/text/plain POST 请求执行分支
    status: completed
  - id: fix-api-define
    content: 修复 define.go ApiDefine 结构体，增加 BodyRaw 字段
    status: completed
  - id: fix-controller-api
    content: 修复 controller/api.go 两处 MapTakeKeys 增加 body_raw 字段提取
    status: completed
  - id: fix-curl-client
    content: 修复 p_curl/curl.go GetGsHttpClient 增加 raw/text/plain 分支
    status: completed
  - id: fix-frontend-save
    content: 修复 Api.vue 前端保存逻辑，增加 body_raw 字段传递和初始化
    status: completed
---

## 用户需求

在接口开发模块中，POST 请求选择 Raw 或 text/plain 作为请求体类型后，输入字符串并执行时报错"运行失败，不支持的请求类型"。需要修复 raw/text/plain 请求体类型从保存到执行的完整数据链路。

## 产品概述

修复 API 接口开发功能中 Raw 和 text/plain 两种请求体类型的支持，使其能够正常保存 body_raw 数据并通过 HTTP 请求发送原始字符串。

## 核心功能

- 后端 Run() 方法增加 raw/text/plain 的 POST 请求执行分支，使用 BodyRaw 字段发送原始字符串
- 后端控制器字段提取增加 body_raw，确保数据能保存到数据库
- 前端保存逻辑补充 body_raw 字段传递
- Curl 执行器同步增加 raw/text/plain 分支支持

## 技术栈

- 后端：Go (Gin 框架)
- 前端：Vue.js + Element Plus
- HTTP 客户端：gitee.com/Sxiaobai/gs/v2/gshttp

## 实现方案

在现有架构基础上，补全 raw/text/plain 类型在数据链路中的缺失环节。核心思路：对 raw 类型，使用类似 PostJson 的方式发送原始字符串，但 Content-Type 设置为对应的值（text/plain 或 application/octet-stream）。

### 修改点分析

**后端修改（4个文件）：**

1. `internal/app/dtool/api/api.go` Run() 方法：在 POST 分支中增加 `raw` 和 `text/plain` 的处理，使用 `gshttp.PostJson(url).BodyStr(h.CurlStruct.BodyRaw)` 并通过 Header 设置正确的 Content-Type（raw 类型使用 application/octet-stream，text/plain 使用 text/plain）。这是因为 gshttp 库的 PostJson 本质是 POST + 设置 Content-Type + BodyStr，所以需要覆盖 Content-Type。

2. `internal/app/dtool/controller/api.go`：

- L521-522 `MapTakeKeys` 增加 `body_raw` 字段
- L925 批量导入 `MapTakeKeys` 增加 `body_raw` 字段

3. `internal/app/dtool/api/define.go`：`ApiDefine` 结构体增加 `BodyRaw string` 字段

4. `internal/pkg/p_curl/curl.go` GetGsHttpClient()：增加 `raw` 和 `text/plain` 分支

**前端修改（1个文件）：**

5. `web/src/components/Api.vue`：

- L2080-2092 保存数据对象增加 `body_raw: api.body_raw_data`
- L531-546 `createApi` 初始化对象增加 `body_raw: ''`
- L547-560 `copyApi` 初始化对象增加 `body_raw: ''`

## 实现注意事项

- Run() 中对 raw 类型的处理需要覆盖 Content-Type header，因为 `gshttp.PostJson` 会自动设置 `application/json`，需要手动覆盖为 `text/plain` 或 `application/octet-stream`
- `body_raw` 在数据库中作为文本字段存储，不需要 SQL 表结构变更（已确认 NewApi 中已读取 body_raw 字段，说明 tbl_api 表已有该列）
- 对于 raw 类型，前端传递的 body_raw_data 是纯字符串，直接作为请求 body 发送

## 目录结构

```
project-root/
├── internal/
│   ├── app/dtool/api/
│   │   ├── api.go            # [MODIFY] Run() 增加 raw/text/plain POST 分支
│   │   └── define.go         # [MODIFY] ApiDefine 结构体增加 BodyRaw 字段
│   ├── app/dtool/controller/
│   │   └── api.go            # [MODIFY] MapTakeKeys 两处增加 body_raw
│   └── pkg/p_curl/
│       └── curl.go           # [MODIFY] GetGsHttpClient 增加 raw/text/plain 分支
└── web/src/components/
    └── Api.vue               # [MODIFY] 保存数据增加 body_raw 字段
```