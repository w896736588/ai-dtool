import { RecorderTransport } from './transport'

function mount(opts) {
  // proxy.html 已被 dtool 同源 iframe 加载；此页面内没有外层父 page
  // 在 proxy.html 自身的 body 内挂工具条
  const iframe = document.querySelector('iframe[data-dtool-recorder-proxy]') // proxy.html 自身的回填？—— 不可
  // 实际：proxy.html 是独立 HTML，里面直接挂工具条；事件 catch 不到外层 page。
  // 因此改为：proxy.html 内的 JS 通过 window.parent.postMessage 与外层 AddInitScript 协同？
  // 简化：本任务版本，proxy.html 仅作为"iframe 容器"加载 recorder-runtime 代码；
  // 真正的工具条挂在被录 page body（由 AddInitScript 第二段额外 appendChild 完成）。
  // 见 index.js 的处理。
  return null
}

export { mount }