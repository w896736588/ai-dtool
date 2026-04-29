---
name: optimize-async-task-entry-display
overview: 优化主页面左下角异步任务入口的展示形式，将当前 4 个文字 badge 改为紧凑的 "任务 1/1/1/1" 格式，数字分别代表运行中/准备中/待处理/失败，各数字用不同颜色区分，鼠标悬停时显示具体任务类型和数量的 tooltip。
todos:
  - id: modify-template
    content: 修改模板：将 4 个 badge 替换为紧凑的"任务 数字/数字/数字/数字"结构
    status: completed
  - id: modify-tooltip
    content: 修改 getAsyncTaskCounterDescription 方法，返回带数量的提示文案
    status: completed
    dependencies:
      - modify-template
  - id: modify-css
    content: 调整 CSS：新增 digit 和 slash 样式类，替换原 badge 圆角背景为纯色数字
    status: completed
    dependencies:
      - modify-template
---

## 产品概述

优化主页面左下角侧边栏底部"异步任务"状态显示区域，将当前 4 个独立 badge 改为紧凑的单行格式。

## 核心功能

- 将当前"运行中 0 / 准备中 0 / 待处理 0 / 失败 0"四个独立 badge 改为 "任务 1/1/1/1" 紧凑格式
- 四个数字分别代表：运行中、准备中、待处理、失败的数量
- 每个数字使用对应颜色区分（绿色/蓝色/橙色/红色），与现有配色保持一致
- "/" 分隔符使用中性色
- 鼠标悬停在单个数字上时，通过 tooltip 显示具体任务类型和数量（如"运行中: 2"）
- 整体按钮的点击行为、运行中动画、SSE 实时更新等现有功能保持不变

## 技术栈

- Vue.js (Options API) + Element Plus
- 纯 CSS (scoped)

## 实现方案

修改仅涉及 `Home.vue` 一个文件的三处区域：模板、方法、样式。

### 模板改造（第 85~121 行）

将当前 4 个独立 `.async-task-entry__badge` 改为内联紧凑结构：

```
任务 [1]/[1]/[1]/[1]
```

每个数字用独立 `<span>` 包裹，带 `:title` 绑定和对应颜色 CSS 类；`/` 用普通 `<span>` 包裹。

### 方法调整（第 1589~1604 行）

`getAsyncTaskCounterDescription` 方法改为返回带数量的描述，如 "运行中: 2"，格式为 "{类型}: {数量}"。

### 样式调整（第 2460~2511 行）

- `.async-task-entry__summary` 改为 `display: inline-flex`，取消 `flex-wrap: wrap`，使数字在同一行内显示
- 新增 `.async-task-entry__digit` 样式，替代原 `.async-task-entry__badge`，去掉圆角 pill 背景，改为纯色文字加粗
- 新增 `.async-task-entry__slash` 样式，中性灰色分隔符
- 保留 `.async-task-entry__badge--*` 颜色变量供新 digit 类复用

## 实现注意事项

- tooltip 继续使用原生 `title` 属性，不引入额外组件，与现有方案一致
- 保留运行中时的 spinner 和 pulse/sheen 动画不受影响
- 异步任务弹窗内的 el-tag 统计栏（第 510~514 行）不改动，保持弹窗信息详细展示
- `getAsyncTaskEntryClassName()` 和 `getAsyncTaskEntryState()` 整体入口背景色逻辑不变

## 目录结构

```
web/src/components/Home.vue  [MODIFY]
├── 模板第 92~117 行：将 4 个 badge span 改为紧凑的 "数字/数字/数字/数字" 结构
├── 方法第 1589~1604 行：getAsyncTaskCounterDescription 返回带数量的提示文案
└── 样式第 2460~2511 行：将 badge 圆角背景改为纯色文字 + 分隔符样式
```

## SubAgent

- **code-explorer**: 用于在实现阶段精确定位 Home.vue 中需要修改的代码行和上下文，确保模板、方法、样式三处修改准确无误