import type { ExtensionAPI } from '@earendil-works/pi-coding-agent';
import { spawn } from 'child_process';
import { existsSync } from 'fs';
import { join, resolve } from 'path';

function pythonCommand(): string {
  return process.env.PYTHON || (process.platform === 'win32' ? 'python' : 'python3');
}

function scriptPath(): string {
  const roots = [process.env.WORKSPACE, process.cwd()].filter(Boolean) as string[];
  for (const root of roots) {
    const candidate = join(resolve(root), 'skills', 'dtool-api', 'scripts', 'sync_api_by_uri.py');
    if (existsSync(candidate)) return candidate;
  }
  throw new Error('找不到 skills/dtool-api/scripts/sync_api_by_uri.py；请设置 WORKSPACE 为 dtool 工作区');
}

async function executeApiTool(params: any, signal?: AbortSignal, onUpdate?: (update: any) => void): Promise<string> {
  const collectionName = String(params.collection_name || '').trim();
  const folderName = String(params.folder_name || '').trim();
  if (!collectionName) throw new Error('collection_name 不能为空');
  if (!folderName) throw new Error('folder_name 不能为空');
  if (!Array.isArray(params.apis) || params.apis.length === 0) throw new Error('apis 必须是非空数组');

  const timeoutMs = Number(params.timeout_ms || 120_000);
  if (!Number.isFinite(timeoutMs) || timeoutMs <= 0) throw new Error('timeout_ms 必须是正数');

  const args = [
    scriptPath(),
    '--base-url', String(params.base_url || 'http://localhost:17170'),
    '--input', '-',
  ];
  if (params.create_folder === true) args.push('--create-folder');

  const input = JSON.stringify({
    collection_name: collectionName,
    folder_name: folderName,
    apis: params.apis,
  });

  onUpdate?.({
    content: [{ type: 'text', text: `接口同步已启动：${collectionName} / ${folderName}` }],
    details: { status: 'running', collection_name: collectionName, folder_name: folderName },
  });

  return await new Promise<string>((resolvePromise, rejectPromise) => {
    const child = spawn(pythonCommand(), args, {
      env: { ...process.env, PYTHONIOENCODING: 'utf-8' },
      windowsHide: true,
      stdio: ['pipe', 'pipe', 'pipe'],
    });
    const stdout: Buffer[] = [];
    const stderr: Buffer[] = [];
    let outputBytes = 0;
    let settled = false;

    const finish = (error?: Error, output?: string) => {
      if (settled) return;
      settled = true;
      clearTimeout(timer);
      signal?.removeEventListener('abort', abort);
      if (error) rejectPromise(error);
      else resolvePromise(output || JSON.stringify({ created: 0, updated: 0 }));
    };
    const abort = () => {
      child.kill();
      finish(new Error('接口同步已取消'));
    };
    const timer = setTimeout(() => {
      child.kill();
      finish(new Error(`接口同步超时（${timeoutMs}ms）`));
    }, timeoutMs);

    signal?.addEventListener('abort', abort, { once: true });
    child.stdout.on('data', (chunk: Buffer) => {
      outputBytes += chunk.length;
      if (outputBytes > 4 * 1024 * 1024) {
        child.kill();
        finish(new Error('接口同步输出超过 4MB 限制'));
        return;
      }
      stdout.push(chunk);
    });
    child.stderr.on('data', (chunk: Buffer) => stderr.push(chunk));
    child.stdin.on('error', (error: Error) => finish(error));
    child.on('error', (error: Error) => finish(error));
    child.on('close', (code: number | null) => {
      const output = Buffer.concat(stdout).toString('utf-8').trim();
      if (code !== 0) {
        finish(new Error(Buffer.concat(stderr).toString('utf-8').trim() || `sync_api_by_uri.py 退出码 ${code}`));
        return;
      }
      finish(undefined, output);
    });
    child.stdin.end(input, 'utf-8');
    if (signal?.aborted) abort();
  });
}

export default function (pi: ExtensionAPI) {
  pi.registerTool({
    name: 'dtool_api',
    description: '按 URI 将接口定义同步到 dtool 指定集合与文件夹；命中更新，未命中创建。写入前必须明确目标集合和文件夹。',
    parameters: {
      type: 'object',
      properties: {
        collection_name: { type: 'string', description: '目标集合名称' },
        folder_name: { type: 'string', description: '目标文件夹名称' },
        apis: {
          type: 'array',
          description: '接口定义列表；每项需包含 name、method、uri/url、content_type、take_result 等接口字段',
          items: { type: 'object' },
        },
        create_folder: { type: 'boolean', description: '文件夹不存在时是否创建，默认 false' },
        base_url: { type: 'string', description: 'dtool 服务地址，默认 http://localhost:17170' },
        timeout_ms: { type: 'number', description: '执行超时毫秒数，默认 120000' },
      },
      required: ['collection_name', 'folder_name', 'apis'],
    },
    async execute(_toolCallId: string, params: any, signal?: AbortSignal, onUpdate?: (update: any) => void) {
      try {
        const text = await executeApiTool(params, signal, onUpdate);
        return { content: [{ type: 'text', text }], details: {} };
      } catch (error: any) {
        const message = String(error?.message || error);
        return {
          content: [{ type: 'text', text: JSON.stringify({ ok: false, error: { code: 'TOOL_EXECUTION_FAILED', message } }, null, 2) }],
          details: { error: message },
        };
      }
    },
  });
}
