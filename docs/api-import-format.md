# API批量导入JSON格式说明文档

## 概述

本文档描述了用于批量导入API文件夹和接口的JSON格式规范。通过此格式，可以一次性在指定集合下创建完整的文件夹和接口结构。


## JSON格式规范

### 根级别结构

```json
{
  "collection_id": 4,           // 固定为4
  "items": [                    // 必填，要导入的项列表
    // ... 见下文
  ]
}
```

**说明**:
- `collection_id` 必须提供，且必须是已存在的集合ID
- 接口只会向指定集合中添加文件夹和接口，不会创建或修改集合本身

### Item 结构

每个item可以是以下两种类型之一：
- `folder`: 文件夹（只能出现在items根级别）
- `api`: 接口（只能出现在文件夹的children中）

#### Folder 类型

```json
{
  "type": "folder",                    // 必填，固定值"folder"
  "name": "文件夹名称",                // 必填，文件夹名称
  "desc": "文件夹描述",                // 可选，文件夹描述
  "children": [                        // 必填，子项列表（只能是api类型）
    // ... api结构，见下文
  ]
}
```

**说明**:
- folder只能出现在items根级别（集合的直接子项）
- folder的children中只能包含api，不能再嵌套folder
- 如果是POST，那么Content-Type设置为multipart/form-data，并且以此格式生成json
- 生成完的json，帮我压缩一下
## 字段类型说明

### 参数类型 (type field)
- `string`: 字符串类型
- `int`: 整数类型
- `float`: 浮点数类型
- `bool`: 布尔类型
- `file`: 文件类型（用于form-data上传）

### Content-Type 常用值
- `application/json`: JSON格式
- `application/x-www-form-urlencoded`: 表单编码
- `multipart/form-data`: 多部分表单数据
- `text/plain`: 纯文本
- `raw`: 原始数据

## 完整示例

### 示例1：在现有集合中导入文件夹和接口

```json
{
  "collection_id": 1,
  "items": [
    {
      "type": "folder",
      "name": "用户认证",
      "desc": "登录、注册等接口",
      "children": [
        {
          "type": "api",
          "name": "用户登录",
          "method": "POST",
          "url": "$Url$/v1/login",
          "protocol": "https",
          "desc": "用户登录接口",
          "headers": {
            "Content-Type": "application/json"
          },
          "body_json": "{\"username\":\"test\",\"password\":\"123456\"}",
        },
        {
          "type": "api",
          "name": "用户注册",
          "method": "POST",
          "url": "$Url$/v1/register",
          "protocol": "https",
          "desc": "用户注册接口",
          "headers": {
            "Content-Type": "application/json"
          },
          "body_json": "{\"username\":\"test\",\"password\":\"123456\",\"email\":\"test@example.com\"}",
        }
      ]
    },
    {
      "type": "folder",
      "name": "用户信息",
      "desc": "查询、修改用户信息",
      "children": [
        {
          "type": "api",
          "name": "获取用户信息",
          "method": "GET",
          "url": "$Url$/v1/user/profile",
          "protocol": "https",
          "desc": "获取当前用户信息",
          "headers": {
            "Authorization": "Bearer {{token}}"
          },
        }
      ]
    }
  ]
}
```

### 示例2：使用Curl导入到现有集合

```json
{
  "collection_id": 5,
  "items": [
    {
      "type": "folder",
      "name": "订单管理",
      "desc": "订单相关接口",
      "children": [
        {
          "type": "api",
          "name": "创建订单",
          "curlData": "curl -X POST '$Url$/v1/orders' \\n  -H 'Content-Type: application/json' \\n  -H 'Authorization: Bearer {{token}}' \\n  -d '{\"product_id\":123,\"quantity\":1}'",
          "desc": "创建新订单",
        },
        {
          "type": "api",
          "name": "查询订单列表",
          "curlData": "curl -X GET '$Url$/v1/orders?page=1&limit=10' \\n  -H 'Authorization: Bearer {{token}}'",
          "desc": "获取订单列表",
        }
      ]
    }
  ]
}
```

### 示例3：使用查询参数和表单数据

```json
{
  "collection_id": 3,
  "items": [
    {
      "type": "folder",
      "name": "商品管理",
      "desc": "商品相关接口",
      "children": [
        {
          "type": "api",
          "name": "搜索商品",
          "method": "GET",
          "url": "$Url$/v1/products/search",
          "protocol": "https",
          "desc": "搜索商品接口",
          "query_params": [
            {
              "field": "keyword",
              "type": "string",
              "value": "手机",
              "description": "搜索关键词"
            },
            {
              "field": "page",
              "type": "int",
              "value": "1",
              "description": "页码"
            },
            {
              "field": "limit",
              "type": "int",
              "value": "20",
              "description": "每页数量"
            }
          ],
        },
        {
          "type": "api",
          "name": "上传商品图片",
          "method": "POST",
          "url": "$Url$/v1/products/image",
          "protocol": "https",
          "desc": "上传商品图片",
          "content_type": "multipart/form-data",
          "body_form": [
            {
              "field": "image",
              "type": "file",
              "value": "/path/to/image.jpg",
              "description": "商品图片文件"
            },
            {
              "field": "product_id",
              "type": "int",
              "value": "123",
              "description": "商品ID"
            }
          ],
        }
      ]
    }
  ]
}
```
