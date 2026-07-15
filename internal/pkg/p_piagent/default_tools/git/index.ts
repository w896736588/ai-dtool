import type { ExtensionAPI } from '@earendil-works/pi-coding-agent';
import { spawnSync } from 'child_process';
import { readFileSync, existsSync, statSync } from 'fs';
import { extname, basename, join, resolve } from 'path';

// ===================== 通用 git 封装 =====================

type GitResult = { stdout: any; code: number; stderr: string };

function git(args: string[], cwd: string, encoding: 'utf-8' | 'buffer' = 'utf-8'): GitResult {
  const r = spawnSync('git', args, { cwd, encoding, maxBuffer: 64 * 1024 * 1024 });
  return {
    stdout: (r.stdout as any) ?? (encoding === 'buffer' ? Buffer.alloc(0) : ''),
    code: r.status ?? 0,
    stderr: (r.stderr as any)?.toString() ?? '',
  };
}

function gitOrThrow(args: string[], cwd: string, msg?: string): string {
  const r = git(args, cwd);
  if (r.code !== 0) throw new Error(msg ?? `git ${args.join(' ')} 失败: ${r.stderr.trim()}`);
  return typeof r.stdout === 'string' ? r.stdout : r.stdout.toString();
}

function nonEmptyLines(s: string): string[] {
  return (s || '').split('\n').map((x) => x.trim()).filter(Boolean);
}

// ===================== 远程 dtool API =====================

async function callApi(baseUrl: string, token: string, path: string, payload: any): Promise<any> {
  const url = baseUrl.replace(/\/+$/, '') + path;
  const resp = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json; charset=utf-8', Token: token },
    body: JSON.stringify(payload),
  });
  const data = await resp.json();
  if ('ErrCode' in data) data.code = data.ErrCode;
  if ('ErrMsg' in data) data.msg = data.ErrMsg;
  if ('Data' in data) data.data = data.Data;
  return data;
}

function formatRemote(r: any): string {
  if (!r || r.code !== 0) return `操作失败: ${r?.msg || JSON.stringify(r)}`;
  const d = r.data;
  if (Array.isArray(d?.list)) {
    const lines = d.list
      .map((it: any) => `上传成功: ${it.remote_path}\n  文件名: ${it.file_name}\n  大小: ${it.file_size} 字节`)
      .join('\n');
    return `${lines}\n共上传 ${d.list.length} 个文件`;
  }
  if (typeof d === 'string') return d || '操作成功';
  return JSON.stringify(d ?? r, null, 2);
}

// ===================== 类型判定 =====================

const BINARY_EXT = new Set([
  '.exe', '.dll', '.so', '.dylib', '.bin', '.dat', '.zip', '.tar', '.gz', '.7z', '.rar',
  '.pdf', '.doc', '.docx', '.xls', '.xlsx', '.ppt', '.pptx', '.ttf', '.otf', '.woff',
  '.woff2', '.eot', '.mp3', '.mp4', '.avi', '.mov', '.mkv', '.webm', '.wav', '.flac',
  '.ogg', '.o', '.a', '.lib', '.class', '.jar', '.war', '.pyc', '.pyo', '.wasm', '.ico',
  '.cur', '.db', '.sqlite', '.sqlite3', '.node', '.lock', '.sum', '.whl', '.tgz', '.iso',
  '.dmg', '.pkg', '.deb', '.rpm', '.apk', '.ipa', '.msi', '.patch',
]);

const IMAGE_EXT = new Set(['.png', '.jpg', '.jpeg', '.gif', '.webp', '.bmp', '.tiff', '.tif']);

function getExt(p: string): string {
  const name = basename(p).toLowerCase();
  for (const c of ['.tar.gz', '.tar.bz2', '.tar.xz']) if (name.endsWith(c)) return c;
  return extname(name);
}

function isExcluded(p: string): boolean {
  return /\/dist\/|\/dist$/.test(p.replace(/\\/g, '/'));
}

// ===================== action: branch_diff =====================

const EXCLUDE = ['--', '.', ':(exclude)**/dist/**'];

function parseNumstat(out: string): Record<string, [number, number]> {
  const m: Record<string, [number, number]> = {};
  for (const line of (out || '').split('\n')) {
    if (!line.trim()) continue;
    const parts = line.split('\t');
    if (parts.length < 3) continue;
    const addStr = parts[0], delStr = parts[1];
    const filepath = parts.slice(2).join('\t');
    let a = 0, d = 0;
    if (addStr === '-' || delStr === '-') { a = 1; d = 1; }
    else { a = parseInt(addStr) || 0; d = parseInt(delStr) || 0; }
    m[filepath] = [a, d];
  }
  return m;
}

function actionBranchDiff(baseBranch: string, cwd: string): string {
  gitOrThrow(['rev-parse', '--show-toplevel'], cwd, '当前目录不是 git 仓库');
  gitOrThrow(['rev-parse', '--verify', baseBranch], cwd, `基分支 '${baseBranch}' 不存在`);
  const mergeBase = gitOrThrow(['merge-base', baseBranch, 'HEAD'], cwd, `无法计算 '${baseBranch}' 与当前分支的 merge-base`).trim();

  const committed = new Set(nonEmptyLines(git(['diff', '--name-only', mergeBase, 'HEAD', ...EXCLUDE], cwd).stdout));
  const staged = new Set(nonEmptyLines(git(['diff', '--name-only', '--cached', ...EXCLUDE], cwd).stdout));
  const modified = new Set(nonEmptyLines(git(['diff', '--name-only', ...EXCLUDE], cwd).stdout));
  const untracked = new Set(nonEmptyLines(git(['ls-files', '--others', '--exclude-standard', '.'], cwd).stdout));

  const cStat = parseNumstat(git(['diff', '--numstat', mergeBase, 'HEAD', ...EXCLUDE], cwd).stdout);
  const sStat = parseNumstat(git(['diff', '--numstat', '--cached', ...EXCLUDE], cwd).stdout);
  const mStat = parseNumstat(git(['diff', '--numstat', ...EXCLUDE], cwd).stdout);
  const uStat: Record<string, [number, number]> = {};
  for (const f of untracked) {
    try { uStat[f] = [nonEmptyLines(readFileSync(join(cwd, f), 'utf-8')).length, 0]; }
    catch { uStat[f] = [0, 0]; }
  }

  const all = new Set([...committed, ...staged, ...modified, ...untracked]);
  const out: string[] = [];
  for (const f of [...all].sort()) {
    const st: string[] = [];
    let add = 0, del = 0;
    if (committed.has(f)) { st.push('Committed'); const [a, d] = cStat[f] || [0, 0]; add += a; del += d; }
    if (staged.has(f)) { st.push('Staged'); const [a, d] = sStat[f] || [0, 0]; add += a; del += d; }
    if (modified.has(f)) { st.push('Modified'); const [a, d] = mStat[f] || [0, 0]; add += a; del += d; }
    if (untracked.has(f)) { st.push('Untracked'); const [a, d] = uStat[f] || [0, 0]; add += a; del += d; }
    out.push(`${f}\t[${st.join(',')}]\t${add}\t${del}`);
  }
  return out.join('\n') || '当前分支没有改动文件';
}

// ===================== action: file_diff =====================

function gitShowBuf(refPath: string, cwd: string): Buffer | null {
  const r = git(['show', refPath], cwd, 'buffer');
  if (r.code !== 0) return null;
  return Buffer.isBuffer(r.stdout) ? r.stdout : Buffer.from(r.stdout);
}

function actionFileDiff(baseBranch: string, filePath: string, cwd: string): string {
  const workspaceMode = baseBranch === '_workspace_';
  gitOrThrow(['rev-parse', '--show-toplevel'], cwd, '当前目录不是 git 仓库');
  if (isExcluded(filePath)) throw new Error('文件位于 dist 目录下，已按规则过滤');
  const norm = filePath.replace(/\\/g, '/');
  const ext = getExt(filePath);

  let oldRef: string;
  if (workspaceMode) {
    gitOrThrow(['rev-parse', '--verify', 'HEAD'], cwd, '当前仓库没有任何提交（HEAD 不存在）');
    oldRef = 'HEAD';
  } else {
    gitOrThrow(['rev-parse', '--verify', baseBranch], cwd, `基分支 '${baseBranch}' 不存在`);
    oldRef = gitOrThrow(['merge-base', baseBranch, 'HEAD'], cwd, `无法计算 '${baseBranch}' 与当前分支的 merge-base`).trim();
  }

  if (BINARY_EXT.has(ext)) {
    const oldR = git(['cat-file', '-s', `${oldRef}:${norm}`], cwd);
    const oldSize = oldR.code === 0 ? parseInt(oldR.stdout.trim()) || 0 : 0;
    let newSize = 0;
    try { newSize = statSync(resolve(cwd, filePath)).size; } catch { /* ignore */ }
    return JSON.stringify({ is_binary: true, file_type: ext, old_size: oldSize, new_size: newSize }, null, 2);
  }

  if (IMAGE_EXT.has(ext)) {
    const oldBuf = gitShowBuf(`${oldRef}:${norm}`, cwd);
    const oldB64 = oldBuf ? oldBuf.toString('base64') : '';
    let newB64 = '';
    try { newB64 = readFileSync(resolve(cwd, filePath)).toString('base64'); } catch { /* ignore */ }
    return JSON.stringify({ is_image: true, image_type: ext.replace(/^\./, ''), old_image: oldB64, new_image: newB64 }, null, 2);
  }

  const parts: string[] = [];
  if (workspaceMode) {
    const cached = git(['diff', '--cached', '--', norm], cwd).stdout;
    if (cached.trim()) parts.push(cached);
    const wt = git(['diff', '--', norm], cwd).stdout;
    if (wt.trim()) parts.push(wt);
  } else {
    const committed = git(['diff', oldRef, 'HEAD', '--', norm], cwd).stdout;
    if (committed.trim()) parts.push(committed);
    const cached = git(['diff', '--cached', '--', norm], cwd).stdout;
    if (cached.trim()) parts.push(cached);
    const wt = git(['diff', '--', norm], cwd).stdout;
    if (wt.trim()) parts.push(wt);
  }
  const oldContent = git(['show', `${oldRef}:${norm}`], cwd).stdout;
  let newContent = '';
  try { newContent = readFileSync(resolve(cwd, filePath), 'utf-8'); } catch { /* ignore */ }
  return JSON.stringify({ diff: parts.join('\n'), old_content: oldContent, new_content: newContent }, null, 2);
}

// ===================== action: file_changes =====================

function categorizeStatus(code: string): string {
  const c = code.trim();
  if (c === '??') return 'untracked';
  const idx = c[0] ?? '';
  const wt = c[1] ?? '';
  if (idx === 'A' || wt === 'A') return 'added';
  if (idx === 'D' || wt === 'D') return 'deleted';
  if (idx === 'R' || wt === 'R') return 'renamed';
  if (idx === 'C' || wt === 'C') return 'copied';
  if (idx === 'M' || wt === 'M') return 'modified';
  return 'other';
}

function extractPath(line: string): string {
  const rest = line.slice(3).trim();
  if (rest.includes(' -> ')) return rest.split(' -> ').pop()!.trim();
  return rest;
}

function getGitDiff(cwd: string, mergeBase: string, file?: string): string {
  const pathArgs = file ? ['--', file] : ['--', '.', ':(exclude)**/dist/**'];
  let r = git(['diff', mergeBase, 'HEAD', ...pathArgs], cwd);
  if (r.code === 0 && r.stdout.trim()) return r.stdout;
  r = git(['diff', '--cached', ...pathArgs], cwd);
  if (r.code === 0 && r.stdout.trim()) return r.stdout;
  r = git(['diff', ...pathArgs], cwd);
  return r.stdout;
}

function actionFileChanges(localDir: string, parentBranch: string, withDiffs: boolean, targetFile: string): string {
  const raw = git(['status', '--short'], localDir).stdout;
  const summary: Record<string, number> = { added: 0, modified: 0, deleted: 0, renamed: 0, untracked: 0, other: 0, total: 0 };
  const files: any[] = [];
  for (const line of raw.split('\n')) {
    if (!line.trim()) continue;
    const code = line.slice(0, 2);
    const cat = categorizeStatus(code);
    summary[cat] = (summary[cat] || 0) + 1;
    summary.total++;
    files.push({ path: extractPath(line), type: cat, status_code: code });
  }
  const result: any = { local_dir: localDir, summary, files };

  if ((withDiffs || targetFile) && parentBranch) {
    try {
      const mb = git(['merge-base', parentBranch, 'HEAD'], localDir).stdout.trim();
      if (!mb) {
        result.diff_error = `无法计算 '${parentBranch}' 与 HEAD 的 merge-base`;
      } else if (targetFile) {
        result.diff = getGitDiff(localDir, mb, targetFile);
      } else {
        const names = new Set<string>();
        for (const cmd of [['diff', '--name-only', mb, 'HEAD'], ['diff', '--name-only', '--cached'], ['diff', '--name-only']]) {
          nonEmptyLines(git([...cmd, '--', '.', ':(exclude)**/dist/**'], localDir).stdout).forEach((f) => names.add(f));
        }
        const diffs: Record<string, string> = {};
        for (const f of [...names].sort()) diffs[f] = getGitDiff(localDir, mb, f);
        result.diffs = diffs;
      }
    } catch (e: any) {
      result.diff_error = String(e?.message || e);
    }
  }
  return JSON.stringify(result, null, 2);
}

// ===================== action: frontend / backend diff =====================

const FRONTEND_SPECS = [
  '*.js', '*.jsx', '*.ts', '*.tsx', '*.vue', '*.css', '*.scss', '*.sass', '*.less',
  '*.styl', '*.html', '*.htm', '*.json', '*.mjs', '*.cjs', '*.svg', '*.png', '*.jpg',
  '*.jpeg', '*.gif', '*.webp', '*.ico', '*.map', 'package.json', 'package-lock.json',
  'pnpm-lock.yaml', 'yarn.lock', 'npm-shrinkwrap.json', 'tsconfig.json', 'jsconfig.json',
  'vite.config.*', 'webpack.config.*', 'vue.config.*', 'postcss.config.*', 'tailwind.config.*',
  'babel.config.*', '.browserslistrc', '.npmrc', '.nvmrc', ':(exclude)**/dist/**',
];

const BACKEND_SPECS = [
  '*.php', '*.go', '*.py', '*.rb', '*.java', '*.kt', '*.kts', '*.scala', '*.groovy',
  '*.cs', '*.fs', '*.rs', '*.c', '*.cc', '*.cpp', '*.cxx', '*.h', '*.hh', '*.hpp', '*.hxx',
  '*.m', '*.mm', '*.swift', '*.sh', '*.bash', '*.zsh', '*.fish', '*.ps1', '*.bat', '*.cmd',
  '*.pl', '*.pm', '*.t', '*.lua', '*.sql', '*.prisma', '*.proto', '*.thrift', '*.graphql',
  '*.gql', '*.ini', '*.conf', '*.config', '*.cfg', '*.cnf', '*.env', '*.env.*',
  '*.properties', '*.toml', '*.yaml', '*.yml', '*.xml', '*.xsd', '*.xsl', '*.wsdl', '*.json',
  '*.json5', '*.ndjson', '*.txt', '*.logrotate', '*.service', '*.socket', '*.timer', '*.mount',
  '*.target', 'Dockerfile', 'Dockerfile.*', 'docker-compose.yml', 'docker-compose.yaml',
  'compose.yml', 'compose.yaml', 'Makefile', 'GNUmakefile', 'makefile', 'Taskfile.yml',
  'Taskfile.yaml', '.env', '.env.*', '.gitignore', '.gitattributes', '.editorconfig',
  '.dockerignore', '.sqlfluff', '.golangci.yml', '.golangci.yaml', '.flake8', '.pylintrc',
  '.ruff.toml', 'pyproject.toml', 'poetry.lock', 'Pipfile', 'Pipfile.lock',
  'requirements.txt', 'requirements-dev.txt', 'go.mod', 'go.sum', 'Cargo.toml', 'Cargo.lock',
  'Gemfile', 'Gemfile.lock', 'composer.json', 'composer.lock', 'pom.xml', 'build.gradle',
  'build.gradle.kts', 'settings.gradle', 'settings.gradle.kts', '*.md', ':(exclude)**/dist/**',
];

function combinedDiff(baseBranch: string, specs: string[], extraExcludes: string[], cwd: string): string {
  gitOrThrow(['rev-parse', '--show-toplevel'], cwd, '当前目录不是 git 仓库');
  gitOrThrow(['rev-parse', '--verify', baseBranch], cwd, `基分支 '${baseBranch}' 不存在`);
  const mb = gitOrThrow(['merge-base', baseBranch, 'HEAD'], cwd).trim();

  const pathspecs = [...specs];
  for (const p of extraExcludes) {
    const n = p.replace(/\\/g, '/').replace(/^\/+|\/+$/g, '');
    if (n) { pathspecs.push(`:(exclude)${n}/**`); pathspecs.push(`:(exclude)${n}`); }
  }

  const files = new Set<string>();
  for (const cmd of [['diff', '--name-only', mb, 'HEAD'], ['diff', '--name-only', '--cached'], ['diff', '--name-only']]) {
    nonEmptyLines(git([...cmd, '--', ...pathspecs], cwd).stdout).forEach((f) => files.add(f));
  }
  if (files.size === 0) return '当前分支没有匹配范围内的改动';

  const parts: string[] = [];
  for (const f of [...files].sort()) {
    let r = git(['diff', mb, 'HEAD', '--', f], cwd);
    if (r.code === 0 && r.stdout.trim()) { parts.push(r.stdout); continue; }
    r = git(['diff', '--cached', '--', f], cwd);
    if (r.code === 0 && r.stdout.trim()) { parts.push(r.stdout); continue; }
    r = git(['diff', '--', f], cwd);
    if (r.code === 0 && r.stdout.trim()) parts.push(r.stdout);
  }
  return parts.join('\n');
}

// ===================== 工具注册 =====================

export default function (pi: ExtensionAPI) {
  pi.registerTool({
    name: 'dtool_git',
    description: 'dtool Git 工具：远程（上传文件/查当前分支/拉取/切分支）与本地（分支改动列表/单文件 diff/变更汇总/前端后端范围 diff）。',
    parameters: {
      type: 'object',
      properties: {
        action: { type: 'string', description: '操作类型：upload_file | current_branch | pull | change_branch | branch_diff | file_diff | file_changes | frontend_diff | backend_diff' },
        git_id: { type: 'string', description: '远程操作所需的 git 仓库 ID' },
        base_url: { type: 'string', description: 'dtool 服务地址，默认 http://localhost:17170' },
        token: { type: 'string', description: 'dtool 认证 Token' },
        branch_name: { type: 'string', description: 'change_branch 的目标分支名' },
        base_branch: { type: 'string', description: 'branch_diff/file_diff/frontend_diff/backend_diff 的对比基分支；file_diff 可用 _workspace_ 表示对比 HEAD 与工作区' },
        file_path: { type: 'string', description: 'file_diff 的目标文件路径' },
        local_dir: { type: 'string', description: '本地 git 仓库目录，默认当前工作目录' },
        parent_branch: { type: 'string', description: 'file_changes 的对比基分支（可选）' },
        with_diffs: { type: 'boolean', description: 'file_changes 是否返回所有文件 diff' },
        target_file: { type: 'string', description: 'file_changes 单文件 diff 路径' },
        extra_excludes: { type: 'array', items: { type: 'string' }, description: 'frontend_diff/backend_diff 额外排除的路径列表（可选）' },
        local_file_paths: {
          type: 'array',
          description: 'upload_file 要上传的文件列表',
          items: {
            type: 'object',
            properties: {
              full_file_path: { type: 'string', description: '本地文件完整路径' },
              relative_file_path: { type: 'string', description: '相对远程代码目录的路径' },
            },
            required: ['full_file_path', 'relative_file_path'],
          },
        },
      },
      // 只有 action 是真正必填；其余字段在 execute 中都有默认值/守卫（如 base_url、token、local_dir），
      // 或仅特定 action 使用（如 branch_name、file_path、parent_branch、local_file_paths 等），不必全部强制。
      required: ['action'],
    },
    async execute(toolCallId: string, params: any, signal: any, onUpdate: any, ctx: any) {
      try {
        const action = params.action;
        const cwd = params.local_dir && existsSync(params.local_dir) ? params.local_dir : process.cwd();
        const base = params.base_url || 'http://localhost:17170';
        const token = params.token || '';

        let text = '';
        switch (action) {
          case 'upload_file': {
            const r = await callApi(base, token, '/api/GitUploadFile', { git_id: params.git_id, local_file_paths: params.local_file_paths || [] });
            text = formatRemote(r);
            break;
          }
          case 'current_branch': {
            const r = await callApi(base, token, '/api/GitCurrentBranch', { git_id: params.git_id });
            text = formatRemote(r);
            break;
          }
          case 'pull': {
            const r = await callApi(base, token, '/api/GitPull', { git_id: params.git_id });
            text = formatRemote(r);
            break;
          }
          case 'change_branch': {
            const r = await callApi(base, token, '/api/GitChangeBranchById', { git_id: params.git_id, branch_name: params.branch_name });
            text = formatRemote(r);
            break;
          }
          case 'branch_diff':
            text = actionBranchDiff(params.base_branch, cwd);
            break;
          case 'file_diff':
            text = actionFileDiff(params.base_branch, params.file_path, cwd);
            break;
          case 'file_changes':
            text = actionFileChanges(cwd, params.parent_branch || '', !!params.with_diffs, params.target_file || '');
            break;
          case 'frontend_diff':
            text = combinedDiff(params.base_branch, FRONTEND_SPECS, params.extra_excludes || [], cwd);
            break;
          case 'backend_diff':
            text = combinedDiff(params.base_branch, BACKEND_SPECS, params.extra_excludes || [], cwd);
            break;
          default:
            text = `未知 action: ${action}`;
        }
        return { content: [{ type: 'text', text }], details: {} };
      } catch (e: any) {
        return { content: [{ type: 'text', text: `错误: ${e?.message || e}` }], details: { error: String(e?.message || e) } };
      }
    },
  });
}
