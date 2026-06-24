# 查询仓库当前分支

## 任务描述
根据仓库名称查询该仓库的当前所在分支。

## 前置条件
- 知道目标仓库的名称（如 `common3`、`dtool-butler` 等）

## 执行步骤

**⚠️ 严格使用 http_call 调用 API，禁止编写 Python 脚本（run_script）或创建临时脚本文件（file_write）。**

### 步骤1: 查询 Git 配置列表，获取目标仓库的 ID
使用 http_call 工具调用：
- 路径: `/api/GitConfigList`
- 请求体:
```json
{
  "page": 1,
  "page_size": 100
}
```

#### 响应处理
- 关注字段: `Data.git_list` → 仓库信息数组
- 遍历 `Data.git_list`，找到 `name` 等于目标仓库名称的项
- 若 `git_list` 中仓库数超过 100 条，增大 `page_size` 至 200 后重试一次
- 记录该仓库的 `id`（数字）和 `code_path`（代码路径）
- 若 `git_list` 中未找到，检查 `Data.git_group_list` 确认所有分组和仓库名称，向用户列出所有可用仓库名

### 步骤2: 根据仓库 ID 查询当前分支
使用 http_call 工具调用：
- 路径: `/api/GitCurrentBranch`
- 请求体:
```json
{
  "git_id": {步骤1获取的仓库ID}
}
```

#### 响应处理
- 关注字段: `Data` → 当前分支名称（字符串）
- 若 `ErrCode` 为 0 且 `Data` 不为空，则分支名即为结果
- 若该接口失败，可尝试备选接口：`/api/GitQueryCurrentBranch`，参数相同

## 结果汇总
输出当前仓库所在分支的名称，例如：`feature_xxx`、`master` 等。

回复格式示例：
> 仓库 **{仓库名}** 当前分支为 **{分支名}**