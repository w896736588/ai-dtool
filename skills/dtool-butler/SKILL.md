# dtool-butler

管家自用技能模块，提供任务管理、状态查询等工具脚本。

## 说明

- 所有脚本必须使用 `api_common.py` 调用 dtool API（禁止裸写 urllib.request）
- 所有脚本仅依赖 Python 标准库 + api_common.py，无需 pip 安装
- 通过 dtool HTTP API 获取数据，默认地址 `http://localhost:17170`
- Token 参数必填
