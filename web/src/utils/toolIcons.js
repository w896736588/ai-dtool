// 工具名 -> 自绘 SVG 图标 + 柔色圆底 的映射，供对话气泡与配置页内置工具共用。
// 找不到匹配时返回 { found: false }，调用方应回退为仅展示工具名。

// 自绘 24x24 描边图标（仅内部 path/形状，颜色由 currentColor 控制）
const ICONS = {
  terminal: '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M7 9l3 3-3 3"/><path d="M13 15h4"/>',
  document: '<path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z"/><path d="M14 3v5h5"/><path d="M9 13h6"/><path d="M9 17h6"/>',
  pencil:   '<path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z"/>',
  search:   '<circle cx="11" cy="11" r="7"/><path d="M21 21l-4.3-4.3"/>',
  folder:   '<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>',
  checklist:'<path d="M9 6h11"/><path d="M9 12h11"/><path d="M9 18h11"/><path d="M4.5 6.5l1 1 1.5-1.5"/><path d="M4.5 12.5l1 1 1.5-1.5"/><path d="M4.5 18.5l1 1 1.5-1.5"/>',
  globe:    '<circle cx="12" cy="12" r="9"/><path d="M3 12h18"/><path d="M12 3c3 3 3 15 0 18M12 3c-3 3-3 15 0 18"/>',
  bulb:     '<path d="M9 18h6"/><path d="M10 22h4"/><path d="M12 2a7 7 0 0 0-4 12c1 1 1 2 1 3h6c0-1 0-2 1-3a7 7 0 0 0-4-12z"/>',
  plug:     '<path d="M9 2v6"/><path d="M15 2v6"/><path d="M7 8h10v3a5 5 0 0 1-10 0z"/><path d="M12 16v6"/>',
  database: '<ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v14c0 1.7 3.6 3 8 3s8-1.3 8-3V5"/><path d="M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3"/>',
  branch:   '<circle cx="6" cy="5" r="2"/><circle cx="6" cy="19" r="2"/><circle cx="18" cy="9" r="2"/><path d="M6 7v10"/><path d="M18 11c0 4-8 2-8 6"/>',
  workflow: '<rect x="3" y="9" width="6" height="6" rx="1.5"/><rect x="15" y="3" width="6" height="6" rx="1.5"/><rect x="15" y="15" width="6" height="6" rx="1.5"/><path d="M9 12h3"/><path d="M15 6h-3v12h3"/>',
  bell:     '<path d="M18 8a6 6 0 1 0-12 0c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.7 21a2 2 0 0 1-3.4 0"/>',
  browser:  '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M3 9h18"/><circle cx="6.5" cy="6.5" r=".6"/><circle cx="9" cy="6.5" r=".6"/>',
  file:     '<path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z"/><path d="M14 3v5h5"/>'
}

const META = {
  // 通用代码工具
  bash:      { svg: ICONS.terminal, color: '#9254de', bg: '#f9f0ff' },
  shell:     { svg: ICONS.terminal, color: '#9254de', bg: '#f9f0ff' },
  terminal:  { svg: ICONS.terminal, color: '#9254de', bg: '#f9f0ff' },
  command:   { svg: ICONS.terminal, color: '#9254de', bg: '#f9f0ff' },
  sh:        { svg: ICONS.terminal, color: '#9254de', bg: '#f9f0ff' },
  read:      { svg: ICONS.document, color: '#409eff', bg: '#eaf2ff' },
  read_file: { svg: ICONS.document, color: '#409eff', bg: '#eaf2ff' },
  edit:      { svg: ICONS.pencil,  color: '#13c2c2', bg: '#e6fffb' },
  write:     { svg: ICONS.pencil,  color: '#13c2c2', bg: '#e6fffb' },
  write_file:{ svg: ICONS.pencil,  color: '#13c2c2', bg: '#e6fffb' },
  grep:      { svg: ICONS.search,  color: '#595959', bg: '#f5f5f5' },
  search:    { svg: ICONS.search,  color: '#595959', bg: '#f5f5f5' },
  glob:      { svg: ICONS.folder,  color: '#1677ff', bg: '#e6f4ff' },
  ls:        { svg: ICONS.folder,  color: '#1677ff', bg: '#e6f4ff' },
  files:     { svg: ICONS.folder,  color: '#1677ff', bg: '#e6f4ff' },
  task:      { svg: ICONS.checklist, color: '#2f54eb', bg: '#f0f5ff' },
  tasks:     { svg: ICONS.checklist, color: '#2f54eb', bg: '#f0f5ff' },
  todo:      { svg: ICONS.checklist, color: '#2f54eb', bg: '#f0f5ff' },
  web:       { svg: ICONS.globe,  color: '#eb2f96', bg: '#fff0f6' },
  fetch:     { svg: ICONS.globe,  color: '#eb2f96', bg: '#fff0f6' },
  browser:   { svg: ICONS.globe,  color: '#eb2f96', bg: '#fff0f6' },
  think:     { svg: ICONS.bulb,   color: '#722ed1', bg: '#f9f0ff' },
  thinking:  { svg: ICONS.bulb,   color: '#722ed1', bg: '#f9f0ff' },
  // p_piagent 内置工具（dtool_*）
  dtool_api:        { svg: ICONS.plug,     color: '#13c2c2', bg: '#e6fffb' },
  dtool_db:         { svg: ICONS.database, color: '#d48806', bg: '#fff7e6' },
  dtool_git:        { svg: ICONS.branch,   color: '#fa541c', bg: '#fff2e8' },
  dtool_read_file:  { svg: ICONS.document, color: '#409eff', bg: '#eaf2ff' },
  dtool_workflow:   { svg: ICONS.workflow, color: '#722ed1', bg: '#f9f0ff' },
  dtool_notify:     { svg: ICONS.bell,     color: '#eb2f96', bg: '#fff0f6' },
  dtool_playwright: { svg: ICONS.browser,  color: '#52c41a', bg: '#f6ffed' },
  dtool_common:     { svg: ICONS.file,     color: '#1677ff', bg: '#e6f4ff' }
}

export function getToolMeta(name) {
  const n = (name || '').toLowerCase()
  let m = META[n]
  if (!m) {
    // 关键词兜底归类（仅当精确匹配失败时）
    if (n.includes('git')) m = META.dtool_git
    else if (n.includes('db') || n.includes('sql') || n.includes('database')) m = META.dtool_db
    else if (n.includes('api')) m = META.dtool_api
    else if (n.includes('workflow')) m = META.dtool_workflow
    else if (n.includes('notify') || n.includes('ding') || n.includes('mail')) m = META.dtool_notify
    else if (n.includes('playwright')) m = META.dtool_playwright
    else if (n.includes('read')) m = META.read
    else if (n.includes('edit') || n.includes('write')) m = META.edit
    else if (n.includes('bash') || n.includes('shell') || n.includes('terminal')) m = META.bash
    else if (n.includes('search') || n.includes('grep')) m = META.grep
    else if (n.includes('task') || n.includes('todo')) m = META.task
    else if (n.includes('web') || n.includes('fetch') || n.includes('browser')) m = META.web
  }
  // 找不到任何匹配 -> 保底：调用方仅展示工具名，不渲染图标
  return m ? { ...m, found: true } : { found: false }
}
