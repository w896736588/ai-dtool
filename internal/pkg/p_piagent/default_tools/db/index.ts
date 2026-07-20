import type { ExtensionAPI } from '@earendil-works/pi-coding-agent';
import { spawn } from 'child_process';
import { existsSync } from 'fs';
import { join, resolve } from 'path';

const ACTIONS = new Set(['tables', 'structure', 'query', 'exec']);

function pythonCommand(): string {
  return process.env.PYTHON || (process.platform === 'win32' ? 'python' : 'python3');
}

function scriptPath(): string {
  const roots = [process.env.WORKSPACE, process.cwd()].filter(Boolean) as string[];
  for (const root of roots) {
    const candidate = join(resolve(root), 'skills', 'dtool-db', 'scripts', 'db_api.py');
    if (existsSync(candidate)) return candidate;
  }
  throw new Error('找不到 skills/dtool-db/scripts/db_api.py；请设置 WORKSPACE 为 dtool 工作区');
}

function pushArg(args: string[], name: string, value: unknown): void {
  if (value === undefined || value === null || value === '') return;
  args.push(name, String(value));
}

async function executeDbTool(params: any, signal?: AbortSignal, onUpdate?: (update: any) => void): Promise<string> {
  const action = String(params.action || '');
  if (!ACTIONS.has(action)) throw new Error(`未知 action: ${action}`);
  if (!params.mysql_id) throw new Error('mysql_id 不能为空');
  if (action === 'structure' && !params.table_name) throw new Error('structure action 需要 table_name');
  if ((action === 'query' || action === 'exec') && !params.sql) throw new Error(`${action} action 需要 sql`);
  if (action === 'exec' && params.confirmed_write !== true) {
    throw new Error('exec action 仅可在用户明确确认写入及影响范围后调用，并设置 confirmed_write=true');
  }

  const timeoutMs = Number(params.timeout_ms || 60_000);
  if (!Number.isFinite(timeoutMs) || timeoutMs <= 0) throw new Error('timeout_ms 必须是正数');

  const args: string[] = [scriptPath(), action];
  pushArg(args, '--base-url', params.base_url || 'http://localhost:17170');
  pushArg(args, '--mysql-id', params.mysql_id);
  pushArg(args, '--table-name', params.table_name);
  pushArg(args, '--sql', params.sql);
  pushArg(args, '--timeout', timeoutMs / 1000);
  if (params.confirmed_write === true) args.push('--confirmed-write');

  onUpdate?.({
    content: [{ type: 'text', text: `数据库工具已启动：${action}` }],
    details: { status: 'running', action },
  });

  return await new Promise<string>((resolvePromise, rejectPromise) => {
    const child = spawn(pythonCommand(), args, {
      env: {
        ...process.env,
        PYTHONIOENCODING: 'utf-8',
      },
      windowsHide: true,
      stdio: ['ignore', 'pipe', 'pipe'],
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
      else resolvePromise(output || JSON.stringify({ code: 0, data: null }));
    };
    const abort = () => {
      child.kill();
      finish(new Error('数据库工具调用已取消'));
    };
    const timer = setTimeout(() => {
      child.kill();
      finish(new Error(`数据库工具执行超时（${timeoutMs}ms）`));
    }, timeoutMs);

    signal?.addEventListener('abort', abort, { once: true });
    child.stdout.on('data', (chunk: Buffer) => {
      outputBytes += chunk.length;
      if (outputBytes > 4 * 1024 * 1024) {
        child.kill();
        finish(new Error('数据库工具输出超过 4MB 限制，请缩小查询范围'));
        return;
      }
      stdout.push(chunk);
    });
    child.stderr.on('data', (chunk: Buffer) => stderr.push(chunk));
    child.on('error', (error: Error) => finish(error));
    child.on('close', (code: number | null) => {
      const output = Buffer.concat(stdout).toString('utf-8').trim();
      if (code !== 0) {
        finish(new Error(Buffer.concat(stderr).toString('utf-8').trim() || `db_api.py 退出码 ${code}`));
        return;
      }
      finish(undefined, output);
    });
    if (signal?.aborted) abort();
  });
}

export default function (pi: ExtensionAPI) {
  pi.registerTool({
    name: 'dtool_db',
    description: '通过 dtool 数据库配置操作 MySQL/PgSQL：列出表、查看表结构、执行 SELECT；INSERT/UPDATE 必须先获得用户明确确认。',
    parameters: {
      type: 'object',
      properties: {
        action: { type: 'string', enum: [...ACTIONS], description: 'tables | structure | query | exec' },
        mysql_id: { type: 'string', description: 'dtool 数据库配置 ID（支持 MySQL/PgSQL）' },
        table_name: { type: 'string', description: 'structure action 的表名' },
        sql: { type: 'string', description: 'query 的 SELECT，或 exec 的 INSERT/UPDATE' },
        confirmed_write: { type: 'boolean', description: 'exec 必须为 true，表示用户已明确确认写入及影响范围' },
        base_url: { type: 'string', description: 'dtool 服务地址，默认 http://localhost:17170' },
        timeout_ms: { type: 'number', description: '执行超时毫秒数，默认 60000' },
      },
      required: ['action', 'mysql_id'],
    },
    async execute(_toolCallId: string, params: any, signal?: AbortSignal, onUpdate?: (update: any) => void) {
      try {
        const text = await executeDbTool(params, signal, onUpdate);
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
