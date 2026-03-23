# 接口详情页 Tab 切换跳过失焦保存 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 避免接口详情页在按 `Tab` 切换到下一个输入框时立即自动保存，导致新焦点被重绘打断。

**Architecture:** 保持 `ApiDetail.vue` 作为接口详情页唯一保存入口，不重做接口保存链路。通过在详情根节点记录 `Tab` 导航状态，并在现有 `handleSave` 分支里对由 `Tab` 引发的 blur 做一次性跳过，兼容主表单和子编辑器通过 `update` 事件上抛的保存触发。

**Tech Stack:** Vue 3、Element Plus

---

### Task 1: 接口详情页保存触发保护

**Files:**
- Modify: `web/src/components/api/ApiDetail.vue`

**Step 1: Write the failing verification target**

定义目标行为：
- 鼠标点击其他区域失焦时，仍然自动保存
- 按 `Tab` 切换到下一个输入框时，本次失焦不触发保存
- 手动点击 `保存` 按钮和 `Ctrl+S` 不受影响

**Step 2: Implement minimal state guard**

在详情页组件增加：
- `isTabNavigating` 状态
- `Tab` 按键按下/释放的标记逻辑
- `handleSave` 中的跳过分支

**Step 3: Verify build**

Run: `npm run prod`
Workdir: `web`
Expected: PASS
