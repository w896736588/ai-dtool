import type { ExtensionAPI } from '@earendil-works/pi-coding-agent';
import { spawn } from 'child_process';
import { existsSync } from 'fs';
import { join, resolve } from 'path';

const LOCAL_ACTIONS = new Set(['changes', 'diff']);
const REMOTE_ACTIONS = new Set([
  'config_list', 'current_branch', 'pull', 'change_branch', 'status', 'commit_log',
  'change_remote_branch', 'remote_branches', 'group_branches', 'quick_create_branch',
  'set_safe', 'save_credentials', 'upload_file',
  'settings_list', 'config_add', 'config_delete', 'group_list', 'group_add', 'group_delete', 'quick_list',
]);

function pythonCommand(): string {
  return process.env.PYTHON || (process.platform === 'win32' ? 'python' : 'python3');
}

function scriptPath(): string {
  const roots = [process.env.WORKSPACE, process.cwd()].filter(Boolean) as string[];
  for (const root of roots) {
    const candidate = join(resolve(root), 'skills', 'dtool-git', 'scripts', 'git_tool.py');
    if (existsSync(candidate)) return candidate;
  }
  throw new Error('找不到 skills/dtool-git/scripts/git_tool.py；请设置 WORKSPACE 为 dtool 工作区');
}

function pushArg(args: string[], name: string, value: unknown): void {
  if (value === undefined || value === null || value === '') return;
  args.push(name, String(value));
}

async function executeGitTool(params: any, signal?: AbortSignal, onUpdate?: (update: any) => void): Promise<string> {
  const action = String(params.action || '');
  const args: string[] = [scriptPath()];

  if (LOCAL_ACTIONS.has(action)) {
    args.push(action);
    pushArg(args, '--local-dir', params.local_dir || process.cwd());
    pushArg(args, '--compare-mode', params.compare_mode || 'branch');
    pushArg(args, '--compare-branch', params.compare_branch);
    pushArg(args, '--compare-strategy', params.compare_strategy || 'merge_base');
    pushArg(args, '--scope', params.scope || 'all');
    pushArg(args, '--file-path', params.file_path);
    for (const exclude of params.extra_excludes || []) pushArg(args, '--exclude', exclude);
    pushArg(args, '--max-files', params.max_files);
    if (params.include_workspace) args.push('--include-workspace');
    if (params.include_untracked === false) args.push('--no-include-untracked');
    if (action === 'diff') {
      pushArg(args, '--context-lines', params.context_lines);
      pushArg(args, '--max-diff-bytes', params.max_diff_bytes);
      if (params.include_content) args.push('--include-content');
    }
  } else if (REMOTE_ACTIONS.has(action)) {
    args.push('remote', action);
    pushArg(args, '--base-url', params.base_url || 'http://localhost:17170');
    pushArg(args, '--git-id', params.git_id);
    pushArg(args, '--git-group-id', params.git_group_id);
    pushArg(args, '--branch-name', params.branch_name);
    pushArg(args, '--base-branch', params.base_branch);
    pushArg(args, '--branch-type', params.branch_type);
    pushArg(args, '--business-en', params.business_en);
    if (params.local_file_paths) {
      pushArg(args, '--local-file-paths', JSON.stringify(params.local_file_paths));
    }
    if (params.payload) pushArg(args, '--payload-json', JSON.stringify(params.payload));
    if (params.discard_local_changes) args.push('--discard-local-changes');
  } else {
    throw new Error(`未知 action: ${action}`);
  }

  const timeoutMs = Number(params.timeout_ms || (LOCAL_ACTIONS.has(action) ? 120_000 : 600_000));
  const maxOutputBytes = 64 * 1024 * 1024;
  onUpdate?.({
    content: [{ type: 'text', text: `Git 工具已启动：${action}` }],
    details: { status: 'running', action },
  });

  return await new Promise<string>((resolvePromise, rejectPromise) => {
    const child = spawn(pythonCommand(), args, {
      env: { ...process.env, DTOOL_TOKEN: String(params.token || '') },
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
      else resolvePromise(output || JSON.stringify({ ok: true }));
    };
    const abort = () => {
      child.kill();
      finish(new Error('Git 工具调用已取消'));
    };
    const timer = setTimeout(() => {
      child.kill();
      finish(new Error(`Git 工具执行超时（${timeoutMs}ms）`));
    }, timeoutMs);

    signal?.addEventListener('abort', abort, { once: true });
    child.stdout.on('data', (chunk: Buffer) => {
      outputBytes += chunk.length;
      if (outputBytes > maxOutputBytes) {
        child.kill();
        finish(new Error('Git 工具输出超过 64MB 限制'));
        return;
      }
      stdout.push(chunk);
    });
    child.stderr.on('data', (chunk: Buffer) => stderr.push(chunk));
    child.on('error', (error: Error) => finish(error));
    child.on('close', (code: number | null) => {
      const output = Buffer.concat(stdout).toString('utf-8').trim();
      if (code !== 0) {
        if (output) {
          finish(undefined, output);
          return;
        }
        finish(new Error(Buffer.concat(stderr).toString('utf-8').trim() || `git_tool.py 退出码 ${code}`));
        return;
      }
      finish(undefined, output);
    });
    if (signal?.aborted) abort();
  });
}

export default function (pi: ExtensionAPI) {
  pi.registerTool({
    name: 'dtool_git',
    description: '统一 Git 工具。本地使用 changes/diff，将当前 HEAD 与明确指定的 compare_branch 对比；远程 action 覆盖 dtool Git 页面操作。',
    parameters: {
      type: 'object',
      properties: {
        action: {
          type: 'string',
          enum: [...LOCAL_ACTIONS, ...REMOTE_ACTIONS],
          description: '本地：changes、diff；远程：Git 页面操作及 settings/config/group 管理 action',
        },
        local_dir: { type: 'string', description: '本地 Git 仓库目录' },
        compare_mode: { type: 'string', enum: ['branch', 'workspace'], description: 'branch 比较分支提交；workspace 比较 HEAD 与工作区' },
        compare_branch: { type: 'string', description: 'branch 模式必填，由用户明确指定的对比分支，例如 master 或 origin/master' },
        compare_strategy: { type: 'string', enum: ['merge_base', 'tree'], description: '默认 merge_base；tree 直接比较两个提交快照' },
        include_workspace: { type: 'boolean', description: 'branch 模式是否把暂存区和未暂存修改合入最终差异' },
        include_untracked: { type: 'boolean', description: '工作区比较是否包含未跟踪文件，默认 true' },
        scope: { type: 'string', enum: ['all', 'frontend', 'backend'], description: 'frontend=web/**；backend=排除 web/**' },
        file_path: { type: 'string', description: 'diff 指定单文件；不传则返回范围内完整 diff' },
        extra_excludes: { type: 'array', items: { type: 'string' }, description: '额外排除的仓库相对路径' },
        max_files: { type: 'number', description: '返回的最大文件条数，默认 500' },
        context_lines: { type: 'number', description: 'diff 上下文行数，默认 3' },
        max_diff_bytes: { type: 'number', description: 'diff 最大 UTF-8 字节数，默认 200000' },
        timeout_ms: { type: 'number', description: '工具超时毫秒数；本地默认 120000，远程默认 600000' },
        include_content: { type: 'boolean', description: '单文件 diff 是否附带新旧内容；AI 默认不需要' },
        base_url: { type: 'string', description: 'dtool 服务地址' },
        token: { type: 'string', description: 'dtool Token' },
        git_id: { type: 'string', description: '远程仓库配置 ID' },
        git_group_id: { type: 'string', description: 'group_branches 的 Git 分组 ID' },
        branch_name: { type: 'string', description: '切换的目标分支' },
        discard_local_changes: { type: 'boolean', description: '破坏性选项；明确授权后才可设 true' },
        base_branch: { type: 'string', description: 'quick_create_branch 的起始分支' },
        branch_type: { type: 'string', enum: ['feature', 'hotfix'] },
        business_en: { type: 'string' },
        local_file_paths: {
          type: 'array',
          items: {
            type: 'object',
            properties: {
              full_file_path: { type: 'string' },
              relative_file_path: { type: 'string' },
            },
            required: ['full_file_path', 'relative_file_path'],
          },
        },
        payload: { type: 'object', description: 'config/group/settings 管理 action 的原始请求对象' },
      },
      required: ['action'],
    },
    async execute(_toolCallId: string, params: any, signal?: AbortSignal, onUpdate?: (update: any) => void) {
      try {
        const text = await executeGitTool(params, signal, onUpdate);
        return { content: [{ type: 'text', text }], details: {} };
      } catch (error: any) {
        const message = String(error?.message || error);
        return { content: [{ type: 'text', text: JSON.stringify({ ok: false, error: { code: 'TOOL_EXECUTION_FAILED', message } }, null, 2) }], details: { error: message } };
      }
    },
  });
}
