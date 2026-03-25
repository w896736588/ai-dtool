# 小工具常用操作面板设计

## 目标

在小工具模块中新增一个可扩展的“常用操作”tab，作为后续系统级辅助操作的统一承载面板。

首个落地能力为“按端口查询占用进程并确认结束”：

- 用户输入端口后，先查询当前占用该端口的进程信息
- 前端展示结构化结果，包括 PID、进程名、协议、监听地址
- 用户显式确认后，再结束目标进程
- 支持 Windows、Linux、macOS

## 现状

- [Tools.vue](C:/work/frog/dev_tool_master/web/src/components/Tools.vue) 当前只负责渲染小工具左侧 tab，现有内容为时间转换、二维码、字符串解析、Markdown
- 小工具页面尚无可扩展的“动作面板”容器
- 路由与抽屉入口已经存在，不需要改页面导航结构
- 后端目前没有“查询端口占用”或“按 PID 结束进程”的接口
- 仓库中虽有 shell / command 相关基础，但没有现成的跨平台端口占用解析逻辑

## 设计

### 前端结构

- 在 [Tools.vue](C:/work/frog/dev_tool_master/web/src/components/Tools.vue) 中新增 `常用操作` tab
- 新建 [CommonActions.vue](C:/work/frog/dev_tool_master/web/src/components/tools/CommonActions.vue) 作为动作面板入口组件
- `CommonActions.vue` 内部使用“操作卡片”组织内容，首张卡片为“端口进程管理”
- 当前不做复杂配置驱动，只在组件层保证结构可扩展：
  - 每个动作卡片独立维护自己的输入、加载态、结果区和错误提示
  - 后续新增动作时，直接新增并列卡片或拆分子组件

### 首个动作交互

端口进程管理卡片流程：

1. 输入端口
2. 点击“查询占用进程”
3. 展示查询结果列表
4. 用户点击“结束进程”
5. 弹出二次确认
6. 调用结束接口
7. 成功后刷新当前端口占用结果

交互约束：

- 端口必须是 `1-65535`
- 允许一个端口对应多个结果，前端按列表展示，逐条确认结束
- 查询不到结果时给出明确提示，不进入确认流程
- 结束成功后提示成功并刷新
- 结束失败时保留原结果，便于重试或排查

### 后端接口

新增两个基础接口，避免把“查询”和“结束”混成单一副作用接口：

1. `POST /api/ToolPortProcessList`
   - 输入：`port`
   - 输出：占用该端口的进程列表

2. `POST /api/ToolPortProcessKill`
   - 输入：`pid`
   - 输出：是否成功、错误信息

这样设计的原因：

- 更符合“先查再确认”的用户流程
- 接口语义清晰，便于前端组合
- 后续若增加“只查询不结束”“批量展示多个端口状态”等能力，可直接复用查询接口

### 数据结构

查询接口返回统一结构：

```json
{
  "port": 8080,
  "items": [
    {
      "pid": 12345,
      "command": "node",
      "protocol": "tcp",
      "address": "0.0.0.0:8080"
    }
  ]
}
```

字段要求：

- `pid`：整数
- `command`：尽量返回进程名；若系统命令拿不到，则返回原始名称或空串
- `protocol`：统一归一化为小写，如 `tcp`
- `address`：监听地址与端口

### 跨平台实现

查询端口占用：

- Windows：
  - 使用 `netstat -ano -p tcp`
  - 筛选 `LISTENING` 且本地地址命中指定端口的记录
  - 通过 `tasklist /FI "PID eq <pid>" /FO CSV /NH` 查询进程名

- Linux / macOS：
  - 优先使用 `lsof -nP -iTCP:<port> -sTCP:LISTEN`
  - 解析结果得到进程名、PID、协议、监听地址

结束进程：

- Windows：`taskkill /PID <pid> /F`
- Linux / macOS：`kill -9 <pid>`

实现策略：

- 将“命令构造”和“输出解析”拆为独立纯函数，便于测试
- 控制器只做参数校验、调用服务、响应封装
- 若运行环境缺少 `lsof`，Unix 系统返回明确错误，不静默降级

### 代码组织

建议新增一个聚焦职责的后端文件，例如：

- [tool_process.go](C:/work/frog/dev_tool_master/internal/app/dtool/controller/tool_process.go)：控制器入口
- [tool_process_test.go](C:/work/frog/dev_tool_master/internal/app/dtool/controller/tool_process_test.go)：命令解析与参数校验测试

如果项目内更适合将解析逻辑放入 `common` 或 `pkg`，也可提取成小型 helper，但不应把整块逻辑塞进已有大文件中。

### 错误处理

需要覆盖的错误场景：

- 端口为空或非法
- 未找到占用进程
- 查询命令执行失败
- 结束命令执行失败
- 权限不足
- Unix 环境缺少 `lsof`

响应要求：

- 前端看到的是可理解的错误文案，而不是原始堆栈
- 后端日志保留原始命令失败信息，便于排查

## 风险点

- `lsof` 在部分精简 Linux 环境中可能不存在，需要给出明确报错
- 某些场景下一个端口可能存在 IPv4 / IPv6 多条监听记录，前端需要允许结果列表展示而非假定单条记录
- `taskkill /F` 和 `kill -9` 都是强制结束，若目标是关键系统进程，会有误操作风险，因此必须保留用户确认步骤

## 验证

- 后端测试覆盖：
  - 端口参数校验
  - Windows `netstat` 输出解析
  - Unix `lsof` 输出解析
  - 查询结果为空场景
- 前端手工验证覆盖：
  - 常用操作 tab 正常显示
  - 输入非法端口时阻止提交
  - 查询成功后正确展示列表
  - 二次确认后才能结束进程
  - 结束成功后刷新结果
  - Windows、Linux、macOS 至少分别验证一套真实命令路径或在对应系统完成回归
