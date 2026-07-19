// Package recorder_runtime 把 dtool 录制 runtime（自包含 JS bundle）以 Go embed 方式
// 提供给后端注入流程使用，绕开 iframe / 同源 proxy / ws_token fetch：
//
//   - DtoolRecorderRuntimeJS  返回 standalone.js 的源文本，供 page.Evaluate(...) 注入。
//   - 注入后被测 page 顶部会出现 toolbar，所有 click / input / scroll 动作都会 push 到
//     window.__dtoolRecordBuffer；点「结束并下载」会触发 JSON 下载 + 复制到剪贴板，
//     并把 window.__dtoolRecordResult 暴露给 Playwright / 前端读取。
package recorder_runtime

import (
	_ "embed"
)

//go:embed standalone.js
var recorderRuntimeJS string

// RecorderRuntimeJS 返回 standalone.js 的源文本（带末尾换行）。
func RecorderRuntimeJS() string { return recorderRuntimeJS }