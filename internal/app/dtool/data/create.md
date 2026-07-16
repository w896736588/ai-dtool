# Dtool 内置工具创建指南

Pi Agent 通过 **extensions（扩展）** 机制支持自定义工具。每个工具是一个 TypeScript 文件，使用 `pi.registerTool()` 注册。

## 目录规范

在 `internal/app/dtool/data/` 下，**每个子目录 = 一个内置工具**：

```
internal/app/dtool/data/
├── create.md                    ← 本文件
└── <tool_dir>/                  ← 工具目录名（作为唯一标识）
    ├── meta.json                ← 工具元数据
    └── index.ts                 ← TypeScript 实现脚本
```

## meta.json 字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | ✅ | 工具显示名称（中文友好） |
| `tool_name` | string | ✅ | Pi `registerTool` 中的 `name`，函数唯一标识 |
| `description` | string | ✅ | 工具简要描述 |
| `tool_description` | string | ✅ | 发送给 LLM 的功能描述，影响 AI 调用准确度 |
| `parameters` | array | 可选 | 工具参数定义 |

### parameters 数组元素

| 字段 | 类型 | 说明 |
|------|------|------|
| `name` | string | 参数名 |
| `type` | string | 参数类型：`string` / `number` / `boolean` |
| `description` | string | 参数描述 |
| `required` | bool | 是否必填 |

## TypeScript 脚本模板

```typescript
import type { ExtensionAPI } from '@earendil-works/pi-coding-agent';
import { Type } from 'typebox';

export default function (pi: ExtensionAPI) {
  pi.registerTool({
    name: '<tool_name>',      // 与 meta.json 中的 tool_name 一致
    description: '<发送给 LLM 的描述>',
    parameters: Type.Object({
      // 参数定义，使用 TypeBox schema
      param1: Type.String({ description: '参数1描述' }),
    }),
    async execute(toolCallId, params, signal, onUpdate, ctx) {
      // === 在这里编写工具执行逻辑 ===
      // 返回值格式：
      return {
        content: [{ type: 'text', text: '返回给 LLM 的文本内容' }],
        details: {},
      };
    },
  });
}
```

## 注意事项

1. **Node.js 内置模块可直接使用**（如 `fs`, `path`, `child_process`）
2. **npm 依赖**需在工具目录下添加 `package.json` 并执行 `npm install`
3. **退出码**: `execute` 抛出错误会被标记为 `isError: true`，不要返回错误字符串
4. **工作目录**: `ctx` 对象中包含当前工作空间信息
5. **输出限制**: 建议返回内容不超过 50KB / 2000 行，避免超出上下文窗口

## 完整示例

以下是一个"读取文件内容"工具的完整实现：

**目录**: `internal/app/dtool/data/read_file/`

**meta.json**:
```json
{
  "name": "读取文件",
  "tool_name": "dtool_read_file",
  "description": "读取指定文件的内容并返回",
  "tool_description": "Read the content of a file at the given path. Returns the file content as text.",
  "parameters": [
    {
      "name": "filePath",
      "type": "string",
      "description": "文件的绝对或相对路径",
      "required": true
    },
    {
      "name": "maxLines",
      "type": "number",
      "description": "最大返回行数，默认 500",
      "required": false
    }
  ]
}
```

**index.ts**:
```typescript
import type { ExtensionAPI } from '@earendil-works/pi-coding-agent';
import { Type } from 'typebox';
import * as fs from 'fs';
import * as path from 'path';

export default function (pi: ExtensionAPI) {
  pi.registerTool({
    name: 'dtool_read_file',
    description: 'Read the content of a file at the given path. Returns the file content as text.',
    parameters: Type.Object({
      filePath: Type.String({ description: '文件路径' }),
      maxLines: Type.Optional(Type.Number({ description: '最大返回行数', default: 500 })),
    }),
    async execute(toolCallId, params, signal, onUpdate, ctx) {
      const filePath = params.filePath;
      const maxLines = params.maxLines || 500;
      
      if (!fs.existsSync(filePath)) {
        throw new Error(`文件不存在: ${filePath}`);
      }
      
      const content = fs.readFileSync(filePath, 'utf-8');
      const lines = content.split('\n');
      const truncated = lines.slice(0, maxLines).join('\n');
      const suffix = lines.length > maxLines ? `\n\n...(共 ${lines.length} 行，已截断至 ${maxLines} 行)` : '';
      
      return {
        content: [{ type: 'text', text: truncated + suffix }],
        details: {
          path: filePath,
          totalLines: lines.length,
          size: content.length,
        },
      };
    },
  });
}
```

## 相关文档

- Pi Extensions 官方文档: https://pi-doc.com/docs/latest/extensions.html
- TypeBox Schema: https://github.com/sinclairzx81/typebox
