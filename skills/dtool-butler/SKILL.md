# dtool-butler

管家自用技能模块，通过 http_call 调用 dtool API 完成各类任务，可复用操作步骤存档于 step/ 目录。

## 说明

- 所有任务通过 `http_call` 工具调用 dtool HTTP API 完成，不编写 Python 脚本
- 任务执行前优先检索 `index/step.md` 查找已有步骤文件
- 可复用的操作流程归档为 .md 步骤文件，存放在 `step/` 目录
- dtool HTTP API 基地址 `http://localhost:17170`，Token 参数必填
