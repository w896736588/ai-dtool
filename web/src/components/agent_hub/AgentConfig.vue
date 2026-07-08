<template>
  <div class="agent-config">
    <div class="config-header">
      <el-button text @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ agentName }} 配置</h2>
    </div>

    <div class="config-body">
      <el-tabs v-model="activeTab" tab-position="left" class="config-tabs">
        <!-- Agent 配置 -->
        <el-tab-pane label="Agent 配置" name="basic">
          <el-form :model="configForm" label-width="130px" class="config-form">
            <el-form-item label="Agent 名称">
              <el-input v-model="configForm.name" placeholder="给 Agent 起个名字" />
            </el-form-item>
            <el-form-item label="Agent 类型">
              <el-tag>{{ typeLabel(configForm.type) }}</el-tag>
            </el-form-item>
            <el-divider content-position="left">LLM 配置</el-divider>
            <el-form-item label="Provider">
              <el-select v-model="selectedProviderId" style="width:100%" placeholder="请选择 Provider" @change="onProviderChange">
                <el-option v-for="p in providerList" :key="p.id" :label="p.name" :value="p.id" />
              </el-select>
              <div class="field-hint">从全局配置中选择 LLM 服务提供商</div>
            </el-form-item>
            <el-form-item label="默认模型">
              <el-select v-model="selectedModelId" style="width:100%" placeholder="请选择模型" :disabled="!selectedProviderId">
                <el-option v-for="m in currentProviderModels" :key="m.id" :label="m.name + ' (' + m.model + ')'" :value="m.id" />
              </el-select>
              <div class="field-hint">启动 Agent 时使用的默认模型</div>
            </el-form-item>
            <el-divider content-position="left">高级选项</el-divider>
            <el-form-item label="会话存储目录">
              <el-input v-model="piConfig.session_dir" placeholder="留空使用默认目录" />
              <div class="field-hint">Pi 会话 JSONL 文件的存储路径，留空则默认 logs/pi_agent_sessions</div>
            </el-form-item>
            <el-form-item label="额外启动参数">
              <el-input v-model="piConfig.extra_args" placeholder="例如：--no-session" />
              <div class="field-hint">空格分隔的额外命令行参数，如 --no-session 禁用会话持久化</div>
            </el-form-item>
            <el-form-item label="安装状态">
              <el-tag :type="installed ? 'success' : 'danger'">{{ installed ? '已安装' : '未安装' }}</el-tag>
              <span v-if="!installed" style="margin-left: 8px; color: #909399; font-size: 12px;">{{ installHint }}</span>
              <el-button size="small" @click="checkInstall" style="margin-left:12px">重新检测</el-button>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" @click="saveConfig">保存配置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Skills -->
        <el-tab-pane label="Skills" name="skills">
          <div class="skills-toolbar">
            <el-button type="primary" size="small" @click="openSkillAdd('skill')">添加 Skill</el-button>
            <span class="skills-hint">Skills 是可复用的按需能力模块，让 Agent 加载特定场景的专业知识或工具</span>
          </div>
          <el-table :data="skillList" class="config-table" empty-text="暂无 Skills">
            <el-table-column prop="name" label="名称" min-width="140" />
            <el-table-column label="命令" width="140">
              <template #default="{ row }">{{ skillCmd(row) }}</template>
            </el-table-column>
            <el-table-column label="描述" min-width="200">
              <template #default="{ row }">
                <span style="color:#909399;font-size:13px">{{ skillDesc(row) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-switch :model-value="row.enabled === 1" @change="toggleSkill(row)" size="small" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button text size="small" @click="openSkillEdit(row)">编辑</el-button>
                <el-button text size="small" type="danger" @click="deleteSkill(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 扩展 -->
        <el-tab-pane label="扩展" name="tools">
          <div class="skills-toolbar">
            <el-button type="primary" size="small" @click="openSkillAdd('tool')">添加 Tool</el-button>
            <el-divider direction="vertical" />
            <el-button size="small" @click="openBuiltinDialog">内置工具</el-button>
            <span class="skills-hint">所有扩展保存在 ~/.pi/agent/extensions/ 目录下</span>
          </div>
          <el-table :data="mergedToolList" class="config-table" empty-text="暂无 Tools">
            <el-table-column prop="name" label="名称" min-width="130" />
            <el-table-column label="来源" width="90">
              <template #default="{ row }">
                <el-tag v-if="row._source === 'fs'" size="small" type="success">文件系统</el-tag>
                <el-tag v-else size="small" type="info">内置</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="工具名" width="140">
              <template #default="{ row }">{{ row._source === 'fs' ? '-' : toolName(row) }}</template>
            </el-table-column>
            <el-table-column label="描述" min-width="160">
              <template #default="{ row }">
                <span style="color:#909399;font-size:13px">{{ row._source === 'fs' ? '来自文件系统的扩展' : skillDesc(row) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="脚本" width="80">
              <template #default="{ row }">
                <el-tag v-if="row._source === 'fs'" size="small" type="success">已安装</el-tag>
                <el-tag v-else size="small" :type="hasScript(row) ? 'success' : 'info'">
                  {{ hasScript(row) ? '已编写' : '未编写' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-switch v-if="row._source !== 'fs'" :model-value="row.enabled === 1" @change="toggleSkill(row)" size="small" />
                <el-tag v-else size="small" type="success">已安装</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="140">
              <template #default="{ row }">
                <template v-if="row._source === 'fs'">
                  <el-button text size="small" type="danger" @click="removeInstalledTool(row)">移除</el-button>
                </template>
                <template v-else>
                  <el-button text size="small" @click="openSkillEdit(row)">编辑</el-button>
                  <el-button text size="small" type="danger" @click="deleteSkill(row)">删除</el-button>
                </template>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 工作空间 -->
        <el-tab-pane label="工作空间" name="workspaces">
          <div style="margin-bottom:12px">
            <el-button type="primary" size="small" @click="showWorkspaceDialog = true">添加工作空间</el-button>
          </div>
          <el-table :data="workspaces" class="config-table" empty-text="暂无工作空间，请先添加">
            <el-table-column prop="name" label="名称" width="180" />
            <el-table-column prop="path" label="路径" min-width="300" />
            <el-table-column label="操作" width="80">
              <template #default="{ row }">
                <el-button text size="small" type="danger" @click="deleteWorkspace(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

        <!-- 模型配置 -->
        <el-tab-pane label="模型配置" name="models">
          <div class="model-config-tab">
            <div class="model-config-header">
              <el-button type="primary" size="small" @click="showAddProvider">新增 Provider</el-button>
              <span class="skills-hint">管理所有 LLM 服务商及其模型</span>
            </div>
            <div class="provider-list">
              <div v-for="p in fullProviders" :key="p.id" class="provider-card" :class="{ 'provider-card--expanded': isProviderExpanded(p.id) }">
                <div class="provider-card__header" @click="toggleExpand(p.id)">
                  <div class="provider-card__info">
                    <span class="provider-card__name">{{ p.name }}</span>
                    <el-tag size="small" effect="plain">{{ p.provider_type }}</el-tag>
                    <span class="provider-card__url">{{ p.base_url }}</span>
                  </div>
                  <div class="provider-card__actions" @click.stop>
                    <el-button text size="small" @click="showEditProvider(p)">编辑</el-button>
                    <el-button text size="small" @click="showAddModel(p)">+ 添加模型</el-button>
                    <el-button text size="small" type="danger" @click="deleteProvider(p)">删除</el-button>
                    <el-icon class="provider-card__arrow" :class="{ 'provider-card__arrow--open': isProviderExpanded(p.id) }">
                      <ArrowRight />
                    </el-icon>
                  </div>
                </div>
                <div v-if="isProviderExpanded(p.id)" class="provider-card__models">
                  <el-table :data="getProviderModels(p.id)" class="config-table" empty-text="暂无模型">
                    <el-table-column prop="name" label="展示名" min-width="130" />
                    <el-table-column prop="model" label="模型标识" min-width="180" />
                    <el-table-column label="类型" width="70">
                      <template #default="{ row }">
                        <el-tag size="small" effect="light">{{ row.model_type === 'embedding' ? '嵌入' : 'LLM' }}</el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column label="上下文窗口" width="110">
                      <template #default="{ row }">
                        {{ fmtCtx(row.context_size) }}
                      </template>
                    </el-table-column>
                    <el-table-column label="操作" width="200">
                      <template #default="{ row }">
                        <el-button text size="small" @click="showEditModel(row)">编辑</el-button>
                        <el-button text size="small" type="danger" @click="deleteModel(row)">删除</el-button>
                        <el-button text size="small" type="warning" :loading="testingModelId === row.id" @click="testModel(row)">测试</el-button>
                      </template>
                    </el-table-column>
                  </el-table>
                </div>
              </div>
              <div v-if="fullProviders.length === 0" class="empty-hint">暂无 Provider，请点击上方按钮添加</div>
            </div>
          </div>
        </el-tab-pane>

        <!-- 推荐扩展 -->
        <el-tab-pane label="推荐扩展" name="envtools">
          <div class="envtools-toolbar">
            <span class="skills-hint">推荐的 Pi 扩展，安装后可增强 Agent 能力。已安装的可以直接在此移除。</span>
            <el-button size="small" @click="loadEnvTools">刷新状态</el-button>
          </div>
          <el-table :data="envTools" class="config-table" empty-text="暂无可用环境工具">
            <el-table-column label="工具" min-width="220">
              <template #default="{ row }">
                <div style="display:flex;align-items:center;gap:8px">
                  <span style="font-size:18px">{{ row.icon }}</span>
                  <div>
                    <div style="font-weight:500">{{ row.name }}</div>
                    <div style="font-size:12px;color:#909399">{{ row.description }}</div>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="140">
              <template #default="{ row }">
                <template v-if="row.key === 'headroom' && headroomStatus.running">
                  <el-tag type="success" size="small">运行中</el-tag>
                  <span v-if="headroomStatus.pid" style="font-size:11px;color:#909399;margin-left:4px">PID:{{ headroomStatus.pid }}</span>
                </template>
                <template v-else>
                  <el-tag :type="envToolStatusType(row)" size="small">{{ envToolStatusLabel(row) }}</el-tag>
                </template>
              </template>
            </el-table-column>
            <el-table-column label="版本" width="100">
              <template #default="{ row }">
                <span style="font-size:12px;color:#909399">{{ row.version || '-' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" min-width="260">
              <template #default="{ row }">
                <!-- Headroom 专属操作 -->
                <template v-if="row.key === 'headroom'">
                  <el-button v-if="!row.installed" text size="small" type="primary" @click="showEnvToolInstall(row)">
                    查看安装指引
                  </el-button>
                  <template v-if="row.installed">
                    <el-button text size="small" type="primary" @click="showHeadroomConfig(row)">
                      配置
                    </el-button>
                    <el-button v-if="!headroomStatus.running" text size="small" type="success" @click="headroomProcess('start')">
                      启动
                    </el-button>
                    <el-button v-if="headroomStatus.running" text size="small" type="warning" @click="headroomProcess('stop')">
                      停止
                    </el-button>
                    <el-button v-if="headroomStatus.running" text size="small" @click="headroomProcess('restart')">
                      重启
                    </el-button>
                  </template>
                  <el-button text size="small" @click="openEnvToolHomepage(row)">
                    主页 ↗
                  </el-button>
                </template>
                <!-- 其他工具通用操作 -->
                <template v-else>
                  <el-button v-if="!row.installed" text size="small" type="primary" @click="showEnvToolInstall(row)">
                    查看安装指引
                  </el-button>
                  <el-button v-if="row.installed && !row.extension_installed" text size="small" type="warning" @click="envToolAction(row, 'activate')">
                    安装扩展
                  </el-button>
                  <el-button v-if="row.extension_installed" text size="small" type="danger" @click="envToolRemove(row)">
                    移除
                  </el-button>
                  <el-button text size="small" @click="openEnvToolHomepage(row)">
                    主页 ↗
                  </el-button>
                </template>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>

      </el-tabs>
    </div>

    <!-- Skill/Tool 编辑对话框 -->
    <el-dialog v-model="showSkillDialog" :title="dialogTitle" width="620px" :close-on-click-modal="true">
      <el-form :model="skillForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="skillForm.name" placeholder="唯一标识名称" />
        </el-form-item>
        <el-form-item label="类型">
          <el-tag size="small" :type="skillForm.skill_type === 'tool' ? 'warning' : ''">
            {{ skillForm.skill_type === 'tool' ? 'Tool（自定义工具）' : 'Skill（按需技能）' }}
          </el-tag>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="skillForm.description" type="textarea" :rows="2" placeholder="描述此功能" />
        </el-form-item>

        <!-- Skill 专属字段 -->
        <template v-if="skillForm.skill_type === 'skill'">
          <el-form-item label="命令名">
            <el-input v-model="skillForm.command" placeholder="斜杠命令名，如 my-review" />
            <div class="field-hint">Agent 通过 /命令名 触发此 Skill</div>
          </el-form-item>
          <el-form-item label="提示词内容">
            <el-input v-model="skillForm.prompt" type="textarea" :rows="6" placeholder="Skill 的详细提示词/指令内容" />
            <div class="field-hint">当 Agent 加载此 Skill 时执行的系统提示词</div>
          </el-form-item>
        </template>

        <!-- Tool 专属字段 -->
        <template v-if="skillForm.skill_type === 'tool'">
          <el-form-item label="工具名">
            <el-input v-model="skillForm.tool_name" placeholder="工具函数名，如 search_code" />
            <div class="field-hint">Agent 调用的函数名</div>
          </el-form-item>
          <el-form-item label="工具描述">
            <el-input v-model="skillForm.tool_description" type="textarea" :rows="2" placeholder="描述工具的功能，会发送给 LLM" />
          </el-form-item>
          <el-form-item label="参数定义">
            <div class="param-list">
              <div v-for="(p, idx) in skillForm.parameters" :key="idx" class="param-row">
                <el-input v-model="p.name" placeholder="参数名" size="small" style="width:120px" />
                <el-select v-model="p.type" size="small" style="width:90px">
                  <el-option label="string" value="string" />
                  <el-option label="number" value="number" />
                  <el-option label="boolean" value="boolean" />
                </el-select>
                <el-input v-model="p.description" placeholder="参数描述" size="small" style="flex:1" />
                <el-checkbox v-model="p.required" size="small" style="margin-left:4px">必填</el-checkbox>
                <el-button text size="small" type="danger" @click="skillForm.parameters.splice(idx, 1)">×</el-button>
              </div>
              <el-button size="small" @click="skillForm.parameters.push({ name:'', type:'string', description:'', required:false })">
                + 添加参数
              </el-button>
            </div>
          </el-form-item>
          <el-form-item label="脚本代码">
            <div class="script-editor-wrapper">
              <div class="script-header">
                <span class="script-path">~/.pi/agent/extensions/{{ skillForm.name || 'tool' }}.ts</span>
                <span class="script-lang">TypeScript</span>
              </div>
              <el-input
                v-model="skillForm.script_content"
                type="textarea"
                :rows="14"
                placeholder="在此编写此工具的 TypeScript 实现代码。&#10;&#10;示例：&#10;export default function (pi: ExtensionAPI) {&#10;  pi.registerTool({&#10;    name: 'my_tool',&#10;    description: '工具描述',&#10;    parameters: Type.Object({}),&#10;    async execute(toolCallId, params, signal, onUpdate, ctx) {&#10;      return {&#10;        content: [{ type: 'text', text: 'Hello!' }],&#10;        details: {},&#10;      };&#10;    },&#10;  });&#10;}"
                style="font-family: 'Cascadia Code', 'Fira Code', monospace; font-size: 13px; line-height: 1.5;"
              />
            </div>
            <div class="field-hint">
              工具的实际执行代码。使用 pi.registerTool() 注册，需包含完整的 TypeScript 实现。
              <a href="https://pi-doc.com/docs/latest/extensions.html" target="_blank" style="color:#409eff">查看文档</a>
            </div>
          </el-form-item>
        </template>

        <el-form-item label="启用">
          <el-switch v-model="skillForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSkillDialog = false">取消</el-button>
        <el-button type="primary" @click="saveSkill" :disabled="!skillForm.name.trim()">保存</el-button>
      </template>
    </el-dialog>

    <!-- 工作空间对话框 -->
    <el-dialog v-model="showWorkspaceDialog" title="添加工作空间" width="480px" :close-on-click-modal="true">
      <el-form label-width="60px">
        <el-form-item label="名称">
          <el-input v-model="workspaceForm.name" placeholder="例如：my-project" />
        </el-form-item>
        <el-form-item label="路径">
          <el-input v-model="workspaceForm.path" placeholder="例如：C:/work/my-project" />
          <div class="field-hint">本地项目的绝对路径</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showWorkspaceDialog = false">取消</el-button>
        <el-button type="primary" @click="saveWorkspace">保存</el-button>
      </template>
    </el-dialog>

    <!-- Provider 编辑对话框 -->
    <el-dialog v-model="showProviderDlg" :title="editingProviderId ? '编辑 Provider' : '新增 Provider'" width="480px" :close-on-click-modal="true">
      <el-form label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="providerForm.name" placeholder="例如：OpenAI" />
        </el-form-item>
        <el-form-item label="请求格式">
          <el-select v-model="providerForm.request_format" style="width:100%">
            <el-option label="OpenAI Chat Completions" value="openai" />
            <el-option label="OpenAI Responses" value="openai-responses" />
            <el-option label="Anthropic Messages" value="anthropic" />
            <el-option label="DeepSeek (OpenAI兼容)" value="deepseek" />
            <el-option label="Google Generative AI" value="google" />
          </el-select>
          <div class="field-hint">选择 API 的请求格式，决定 Pi 调用时的 endpoint 路径</div>
        </el-form-item>
        <el-form-item label="基础域名">
          <el-input v-model="providerForm.base_url" placeholder="例如：https://api.openai.com" />
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="providerForm.api_key" type="password" show-password placeholder="API 认证密钥" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showProviderDlg = false">取消</el-button>
        <el-button type="primary" @click="saveProvider">保存</el-button>
      </template>
    </el-dialog>

    <!-- Model 编辑对话框 -->
    <el-dialog v-model="showModelDlg" :title="editingModelId ? '编辑模型' : '新增模型'" width="480px" :close-on-click-modal="true">
      <el-form label-width="100px">
        <el-form-item label="所属 Provider">
          <el-select v-model="modelForm.provider_id" style="width:100%" :disabled="editingModelId > 0">
            <el-option v-for="p in fullProviders" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="展示名">
          <el-input v-model="modelForm.name" placeholder="例如：GPT-4o" />
        </el-form-item>
        <el-form-item label="模型标识">
          <el-input v-model="modelForm.model" placeholder="例如：gpt-4o" />
        </el-form-item>
        <el-form-item label="上下文窗口">
          <el-input-number v-model="modelForm.context_size" :min="1000" :max="4194304" :step="1000" style="width:100%" />
          <div class="field-hint">以 token 为单位的最大上下文窗口大小</div>
        </el-form-item>
        <el-form-item label="模型类型">
          <el-select v-model="modelForm.model_type" style="width:100%">
            <el-option label="LLM（大语言模型）" value="llm" />
            <el-option label="嵌入模型" value="embedding" />
          </el-select>
        </el-form-item>
        <el-form-item label="URI">
          <el-input v-model="modelForm.uri" :placeholder="defaultUriForProvider(modelForm.provider_id)" />
          <div class="field-hint">留空则根据 Provider 类型自动推断</div>
        </el-form-item>
        <el-form-item label="完整地址">
          <div class="url-preview">{{ buildUrl(getProviderBaseUrl(modelForm.provider_id), editingModelId ? modelForm.uri : (modelForm.uri || defaultUriForProvider(modelForm.provider_id))) || '-' }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showModelDlg = false">取消</el-button>
        <el-button type="primary" @click="saveModel">保存</el-button>
      </template>
    </el-dialog>

    <!-- 内置工具对话框 -->
    <el-dialog v-model="showBuiltinDialog" title="Dtool 内置 Tools" width="640px" :close-on-click-modal="true">
      <el-table :data="builtinTools" max-height="400" empty-text="暂无内置工具，请在 internal/app/dtool/data/ 目录下添加">
        <el-table-column prop="name" label="名称" width="120" />
        <el-table-column label="来源" width="120">
          <template #default="{ row }">
            <el-tag size="small" type="info">data/{{ row.dir_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="180">
          <template #default="{ row }">
            <span style="color:#909399;font-size:13px">{{ row.description }}</span>
          </template>
        </el-table-column>
        <el-table-column label="参数" width="70">
          <template #default="{ row }">
            <span style="color:#909399;font-size:12px">{{ (row.parameters || []).length }} 个</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button text size="small" type="primary" @click="installBuiltinTool(row)">安装</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="field-hint" style="margin-top: 12px;">
        内置工具存放在 <code>internal/app/dtool/data/</code> 目录下，每个子目录对应一个工具。
        <a href="https://pi-doc.com/docs/latest/extensions.html" target="_blank" style="color:#409eff">Pi Extensions 文档</a>
      </div>
      <template #footer>
        <el-button @click="showBuiltinDialog = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- Headroom 配置对话框 -->
    <el-dialog v-model="showHeadroomConfigDialog" title="Headroom 代理配置" width="640px" :close-on-click-modal="true">
      <el-form :model="headroomConfig" label-width="140px">
        <el-form-item label="代理端口">
          <el-input-number v-model="headroomConfig.port" :min="1" :max="65535" style="width:160px" />
          <div class="field-hint">代理服务监听端口，Agent 将请求发给 localhost:此端口</div>
        </el-form-item>
        <el-divider content-position="left">大模型服务商上游地址（留空使用默认值）</el-divider>
        <el-form-item label="Anthropic">
          <el-input v-model="headroomConfig.anthropic_api_url" placeholder="默认: https://api.anthropic.com" />
        </el-form-item>
        <el-form-item label="OpenAI">
          <el-input v-model="headroomConfig.openai_api_url" placeholder="默认: https://api.openai.com" />
        </el-form-item>
        <el-form-item label="Gemini">
          <el-input v-model="headroomConfig.gemini_api_url" placeholder="默认: https://generativelanguage.googleapis.com" />
        </el-form-item>
        <el-form-item label="Cloud Code">
          <el-input v-model="headroomConfig.cloudcode_api_url" placeholder="默认: https://cloudcode-pa.googleapis.com" />
        </el-form-item>
        <el-form-item label="Vertex AI">
          <el-input v-model="headroomConfig.vertex_api_url" placeholder="默认: https://us-central1-aiplatform.googleapis.com" />
        </el-form-item>
      </el-form>
      <div class="field-hint" style="margin-top:8px">
        启动代理后，设置环境变量 <code>ANTHROPIC_BASE_URL=http://localhost:{端口}</code> 或 <code>OPENAI_BASE_URL=http://localhost:{端口}/v1</code> 即可让 Agent 经 Headroom 压缩上下文。
      </div>
      <template #footer>
        <el-button @click="showHeadroomConfigDialog = false">取消</el-button>
        <el-button type="primary" @click="saveHeadroomConfig">保存配置</el-button>
      </template>
    </el-dialog>

    <!-- 环境工具安装指引对话框 -->
    <el-dialog v-model="showEnvToolDialog" :title="'安装 ' + (envToolDialogData.name || '')" width="560px" :close-on-click-modal="true">
      <div style="display:flex;align-items:center;gap:8px;margin-bottom:16px">
        <span style="font-size:22px">{{ envToolDialogData.icon }}</span>
        <span style="font-weight:500;font-size:16px">{{ envToolDialogData.name }}</span>
      </div>
      <p style="color:#606266;margin-bottom:20px">{{ envToolDialogData.description }}</p>
      <div style="background:#f5f7fa;border-radius:6px;padding:16px;margin-bottom:12px">
        <div style="font-size:12px;color:#909399;margin-bottom:8px">安装命令（在终端执行）：</div>
        <el-input
          :model-value="envToolDialogData.install_cmd_hint"
          readonly
          type="textarea"
          :rows="1"
          style="font-family:'Cascadia Code',monospace;font-size:13px"
        />
      </div>
      <div v-if="envToolDialogData.activate_cmd_hint" style="background:#fdf6ec;border-radius:6px;padding:16px">
        <div style="font-size:12px;color:#909399;margin-bottom:8px">安装后在终端执行激活命令：</div>
        <el-input
          :model-value="envToolDialogData.activate_cmd_hint"
          readonly
          type="textarea"
          :rows="1"
          style="font-family:'Cascadia Code',monospace;font-size:13px"
        />
      </div>
      <div v-if="envToolDialogData.homepage" style="margin-top:12px">
        <a :href="envToolDialogData.homepage" target="_blank" style="color:#409eff;font-size:13px">查看项目主页 →</a>
      </div>
      <template #footer>
        <el-button @click="showEnvToolDialog = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import Base from '@/utils/base.js'
import aiSet from '@/utils/base/ai_set'
import { ArrowLeft, ArrowRight, Plus } from '@element-plus/icons-vue'

export default {
  name: 'AgentConfig',
  components: { ArrowLeft, ArrowRight, Plus },
  data() {
    return {
      activeTab: 'basic',
      agentId: 0,
      agentName: '',
      installed: false,
      installHint: '',

      configForm: { name: '', type: '' },
      piConfig: { session_dir: '', extra_args: '' },

      // Agent 配置用的 Provider/Model 列表
      providerList: [],
      allModels: [],
      selectedProviderId: null,
      selectedModelId: null,

      // 模型配置 Tab 用的完整数据
      fullProviders: [],
      expandedProviderIds: {},  // { pid: true } 跟踪展开状态

      // Provider 对话框
      showProviderDlg: false,
      editingProviderId: 0,
      providerForm: { name: '', request_format: 'openai', base_url: '', api_key: '' },

      // Model 对话框
      showModelDlg: false,
      editingModelId: 0,
      modelForm: { provider_id: 0, name: '', model_type: 'llm', model: '', uri: '', context_size: 128000 },
      testingModelId: 0,  // 正在测试的模型 ID

      skills: [],
      showSkillDialog: false,
      editingSkillId: null,
      skillForm: this.emptySkillForm(),

      workspaces: [],
      showWorkspaceDialog: false,
      workspaceForm: { name: '', path: '' },

      builtinTools: [],
      showBuiltinDialog: false,

      envTools: [],
      showEnvToolDialog: false,
      envToolDialogData: {},

      // Headroom 配置
      showHeadroomConfigDialog: false,
      headroomConfig: {
        port: 8787,
        anthropic_api_url: '',
        openai_api_url: '',
        gemini_api_url: '',
        cloudcode_api_url: '',
        vertex_api_url: ''
      },
      headroomStatus: {
        installed: false,
        version: '',
        running: false,
        pid: 0,
        started_at: 0
      },

      installedTools: []
    }
  },
  computed: {
    skillList() {
      return this.skills.filter(s => s.skill_type === 'skill')
    },
    toolList() {
      return this.skills.filter(s => s.skill_type === 'tool')
    },
    mergedToolList() {
      // DB 中的 tools
      const dbTools = this.skills.filter(s => s.skill_type === 'tool').map(t => ({ ...t, _source: 'db' }))
      // 文件系统中有但 DB 中没有的（按 name 去重）
      const dbNames = new Set(dbTools.map(t => t.name))
      const fsOnly = this.installedTools
        .filter(t => !dbNames.has(t.name))
        .map(t => ({ name: t.name, file_path: t.file_path, _source: 'fs' }))
      return [...dbTools, ...fsOnly]
    },
    dialogTitle() {
      if (this.editingSkillId) return '编辑 ' + (this.skillForm.skill_type === 'tool' ? 'Tool' : 'Skill')
      return '添加 ' + (this.skillForm.skill_type === 'tool' ? 'Tool' : 'Skill')
    },
    currentProviderModels() {
      if (!this.selectedProviderId) return []
      return this.allModels.filter(m => parseInt(m.provider_id) === parseInt(this.selectedProviderId))
    }

  },
  mounted() {
    this.agentId = parseInt(this.$route.query.agent_id) || 0
    if (!this.agentId) {
      this.$router.push('/AgentHub')
      return
    }
    this.loadData()
  },
  methods: {
    emptySkillForm() {
      return {
        name: '', skill_type: 'skill', description: '', enabled: true,
        command: '', prompt: '',
        tool_name: '', tool_description: '', parameters: [],
        script_content: ''
      }
    },
    loadData() {
      this.loadProviderModels()
      Base.BasePost('/api/AgentV2List', {}, (res) => {
        const agents = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
        const agent = agents.find(a => a.id === this.agentId)
        if (agent) {
          this.agentName = agent.name
          this.installed = agent.installed
          this.installHint = agent.install_hint
          this.configForm = { name: agent.name, type: agent.type }
          if (agent.config) {
            try {
              const cfg = JSON.parse(agent.config)
              if (cfg.session_dir) this.piConfig.session_dir = cfg.session_dir
              if (cfg.extra_args) this.piConfig.extra_args = cfg.extra_args
              if (cfg.provider_id) this.selectedProviderId = cfg.provider_id
              if (cfg.model_id) this.selectedModelId = cfg.model_id
            } catch (e) {}
          }
        }
      })
      this.loadSkills()
      this.loadWorkspaces()
      this.loadEnvTools()
    },
    saveConfig() {
      const configObj = {
        provider_id: this.selectedProviderId || 0,
        model_id: this.selectedModelId || 0,
        session_dir: this.piConfig.session_dir,
        extra_args: this.piConfig.extra_args
      }
      const config = JSON.stringify(configObj)
      Base.BasePost('/api/AgentV2Save', {
        id: this.agentId,
        name: this.configForm.name,
        type: this.configForm.type,
        config: config
      }, () => {
        this.$message.success('配置已保存')
        this.loadData()
      })
    },
    checkInstall() {
      Base.BasePost('/api/AgentV2CheckInstall', { type: this.configForm.type }, (res) => {
        this.installed = (res.ErrCode === 0 && res.Data) ? res.Data.installed : false
        this.installHint = (res.ErrCode === 0 && res.Data) ? (res.Data.install_hint || '') : ''
      })
    },

    loadProviderModels() {
      Base.BasePost('/api/AgentV2ProviderModels', {}, (res) => {
        if (res.ErrCode === 0 && res.Data && res.Data.providers) {
          const providers = res.Data.providers
          this.providerList = providers.map(p => ({ id: p.id, name: p.name, provider_type: p.provider_type }))
          // 模型配置 Tab 用的完整数据
          this.fullProviders = providers.map(p => ({ ...p }))
          // 默认展开所有 Provider
          for (const p of providers) {
            this.expandedProviderIds[p.id] = true
          }
          this.allModels = []
          for (const p of providers) {
            for (const m of (p.models || [])) {
              m.provider_id = p.id
              this.allModels.push(m)
            }
          }
        }
      })
    },
    onProviderChange() {
      this.selectedModelId = null
    },

    // ========== 模型配置 Tab：Provider 管理 ==========
    getProviderModels(pid) {
      return this.allModels.filter(m => parseInt(m.provider_id) === parseInt(pid))
    },
    isProviderExpanded(pid) {
      return this.expandedProviderIds[pid] === true
    },
    toggleExpand(pid) {
      if (this.expandedProviderIds[pid]) {
        this.expandedProviderIds = { ...this.expandedProviderIds, [pid]: false }
      } else {
        this.expandedProviderIds = { ...this.expandedProviderIds, [pid]: true }
      }
    },
    showAddProvider() {
      this.editingProviderId = 0
      this.providerForm = { name: '', request_format: 'openai', base_url: '', api_key: '' }
      this.showProviderDlg = true
    },
    showEditProvider(p) {
      this.editingProviderId = p.id
      this.providerForm = { id: p.id, name: p.name, request_format: p.provider_type, base_url: p.base_url, api_key: p.api_key }
      this.showProviderDlg = true
    },
    saveProvider() {
      if (!this.providerForm.name || !this.providerForm.base_url) {
        this.$message.warning('请填写名称和基础域名')
        return
      }
      aiSet.AiProviderAdd({
        id: this.editingProviderId || undefined,
        name: this.providerForm.name,
        request_format: this.providerForm.request_format,
        base_url: this.providerForm.base_url,
        api_key: this.providerForm.api_key
      }, (res) => {
        if (res.ErrCode === 0) {
          this.showProviderDlg = false
          this.loadProviderModels()
          this.$message.success('保存成功')
        }
      })
    },
    deleteProvider(p) {
      this.$confirm(`确定删除 Provider「${p.name}」？关联的模型也会被删除。`, '提示', { type: 'warning' }).then(() => {
        aiSet.AiProviderDelete({ id: p.id }, (res) => {
          if (res.ErrCode === 0) {
            this.loadProviderModels()
            this.$message.success('已删除')
          }
        })
      }).catch(() => {})
    },

    // ========== 模型配置 Tab：Model 管理 ==========
    defaultUriForProvider(pid) {
      const p = this.fullProviders.find(p => parseInt(p.id) === parseInt(pid))
      if (!p) return '/v1/chat/completions'
      const fmt = p.provider_type || p.request_format || 'openai'
      switch (fmt) {
        case 'anthropic': return '/v1/messages'
        case 'openai-responses': return '/v1/responses'
        case 'google': return '/v1beta/models'
        default: return '/v1/chat/completions'
      }
    },
    showAddModel(p) {
      this.expandedProviderIds[p.id] = true
      this.editingModelId = 0
      this.modelForm = { provider_id: p.id, name: '', model_type: 'llm', model: '', uri: '', context_size: 128000 }
      this.showModelDlg = true
    },
    showEditModel(m) {
      this.editingModelId = m.id
      this.modelForm = {
        id: m.id, provider_id: m.provider_id, name: m.name,
        model_type: m.model_type || 'llm', model: m.model,
        uri: m.uri || '', context_size: m.context_size || 128000
      }
      this.showModelDlg = true
    },
    saveModel() {
      if (!this.modelForm.model) {
        this.$message.warning('请填写模型标识')
        return
      }
      // 新建时 URI 为空才给默认值，编辑时原样保存
      const uri = this.editingModelId ? this.modelForm.uri : (this.modelForm.uri || this.defaultUriForProvider(this.modelForm.provider_id))
      aiSet.AiModelAdd({
        id: this.editingModelId || undefined,
        provider_id: this.modelForm.provider_id,
        name: this.modelForm.name || this.modelForm.model,
        model_type: this.modelForm.model_type,
        model: this.modelForm.model,
        uri: uri,
        context_size: this.modelForm.context_size || 128000
      }, (res) => {
        if (res.ErrCode === 0) {
          this.showModelDlg = false
          this.loadProviderModels()
          this.$message.success('保存成功')
        }
      })
    },
    deleteModel(m) {
      this.$confirm(`确定删除模型「${m.name || m.model}」？`, '提示', { type: 'warning' }).then(() => {
        aiSet.AiModelDelete({ id: m.id }, (res) => {
          if (res.ErrCode === 0) {
            this.loadProviderModels()
            this.$message.success('已删除')
          }
        })
      }).catch(() => {})
    },
    testModel(m) {
      const provider = this.fullProviders.find(p => p.id === m.provider_id)
      const providerName = provider ? provider.name : 'Unknown'
      this.testingModelId = m.id
      Base.BasePost('/api/AgentV2ModelTest', {
        provider_id: m.provider_id,
        model_id: m.id
      }, (res) => {
        this.testingModelId = 0
        if (res.ErrCode === 0) {
          const resp = res.Data.response || ''
          this.$notify({
            title: `测试通过 — ${providerName} / ${m.name || m.model}`,
            message: resp,
            type: 'success',
            duration: 5000
          })
        } else {
          this.$notify({
            title: `测试失败 — ${providerName} / ${m.name || m.model}`,
            message: res.ErrMsg || '未知错误',
            type: 'error',
            duration: 8000
          })
        }
      })
    },
    buildUrl(base, uri) {
      const cleanBase = (base || '').replace(/\/+$/, '')
      if (!cleanBase) return uri || ''
      if (!uri) return cleanBase
      const cleanUri = uri.startsWith('/') ? uri : '/' + uri
      return cleanBase + cleanUri
    },
    getProviderBaseUrl(pid) {
      const p = this.fullProviders.find(p => parseInt(p.id) === parseInt(pid))
      return p ? p.base_url : ''
    },
    fmtCtx(size) {
      if (!size) return '—'
      if (size >= 1000000) return (size / 1000000).toFixed(1) + 'M'
      if (size >= 1000) return (size / 1000).toFixed(0) + 'K'
      return String(size)
    },

    // Skills & Tools
    loadSkills() {
      Base.BasePost('/api/AgentV2SkillList', { agent_id: this.agentId }, (res) => {
        this.skills = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
      })
      this.loadInstalledTools()
    },
    skillDesc(row) {
      try {
        const cfg = JSON.parse(row.config || '{}')
        return cfg.description || '-'
      } catch (e) { return '-' }
    },
    skillCmd(row) {
      try {
        const cfg = JSON.parse(row.config || '{}')
        return cfg.command ? '/' + cfg.command : '-'
      } catch (e) { return '-' }
    },
    toolName(row) {
      try {
        const cfg = JSON.parse(row.config || '{}')
        return cfg.tool_name || '-'
      } catch (e) { return '-' }
    },
    hasScript(row) {
      try {
        const cfg = JSON.parse(row.config || '{}')
        return !!(cfg.script_content && cfg.script_content.trim())
      } catch (e) { return false }
    },
    openSkillAdd(type) {
      this.editingSkillId = null
      this.skillForm = this.emptySkillForm()
      this.skillForm.skill_type = type
      this.showSkillDialog = true
    },
    openSkillEdit(row) {
      this.editingSkillId = row.id
      try {
        const cfg = JSON.parse(row.config || '{}')
        this.skillForm = {
          name: row.name,
          skill_type: row.skill_type,
          description: cfg.description || '',
          enabled: row.enabled === 1,
          command: cfg.command || '',
          prompt: cfg.prompt || '',
          tool_name: cfg.tool_name || '',
          tool_description: cfg.tool_description || '',
          parameters: cfg.parameters || [],
          script_content: cfg.script_content || ''
        }
      } catch (e) {
        this.skillForm = {
          name: row.name,
          skill_type: row.skill_type,
          description: '',
          enabled: row.enabled === 1,
          command: '', prompt: '',
          tool_name: '', tool_description: '', parameters: [],
          script_content: ''
        }
      }
      this.showSkillDialog = true
    },
    saveSkill() {
      const s = this.skillForm
      let configObj = {}
      if (s.skill_type === 'skill') {
        configObj = {
          description: s.description,
          command: s.command,
          prompt: s.prompt
        }
      } else {
        configObj = {
          description: s.description || s.tool_description,
          tool_name: s.tool_name,
          tool_description: s.tool_description,
          parameters: s.parameters.filter(p => p.name.trim()).map(p => ({
            name: p.name, type: p.type, description: p.description, required: !!p.required
          })),
          script_content: s.script_content
        }
      }

      Base.BasePost('/api/AgentV2SkillSave', {
        id: this.editingSkillId || undefined,
        agent_id: this.agentId,
        name: s.name,
        skill_type: s.skill_type,
        config: JSON.stringify(configObj),
        enabled: s.enabled ? 1 : 0
      }, () => {
        this.showSkillDialog = false
        this.loadSkills()
      })
    },
    toggleSkill(row) {
      Base.BasePost('/api/AgentV2SkillSave', {
        id: row.id, agent_id: this.agentId,
        name: row.name, skill_type: row.skill_type,
        config: row.config,
        enabled: row.enabled === 1 ? 0 : 1
      }, () => { this.loadSkills() })
    },
    deleteSkill(row) {
      this.$confirm('确定删除此 ' + (row.skill_type === 'tool' ? 'Tool' : 'Skill') + '？', '提示', { type: 'warning' }).then(() => {
        Base.BasePost('/api/AgentV2SkillDelete', { id: row.id }, () => { this.loadSkills() })
      }).catch(() => {})
    },

    // 内置工具
    openBuiltinDialog() {
      this.showBuiltinDialog = true
      this.loadBuiltinTools()
    },
    loadBuiltinTools() {
      Base.BasePost('/api/AgentV2BuiltinToolList', {}, (res) => {
        this.builtinTools = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
      })
    },
    installBuiltinTool(tool) {
      this.$confirm(`安装内置工具「${tool.name}」？安装后可在 Tools 列表中查看和编辑。`, '确认安装', { type: 'info' }).then(() => {
        const configObj = {
          description: tool.description,
          tool_name: tool.tool_name,
          tool_description: tool.tool_description,
          parameters: tool.parameters || [],
          script_content: tool.script_content || ''
        }
        Base.BasePost('/api/AgentV2SkillSave', {
          agent_id: this.agentId,
          name: tool.name,
          skill_type: 'tool',
          config: JSON.stringify(configObj),
          enabled: 1
        }, () => {
          this.$message.success('工具已安装')
          this.showBuiltinDialog = false
          this.loadSkills()
        })
      }).catch(() => {})
    },

    // 环境工具
    loadEnvTools() {
      Base.BasePost('/api/AgentV2EnvToolList', {}, (res) => {
        this.envTools = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
      })
      this.loadHeadroomStatus()
    },
    // Headroom 代理
    loadHeadroomStatus() {
      Base.BasePost('/api/AgentV2HeadroomStatus', { agent_id: this.agentId }, (res) => {
        if (res.ErrCode === 0 && res.Data) {
          this.headroomStatus = {
            installed: res.Data.installed || false,
            version: res.Data.version || '',
            running: res.Data.running || false,
            pid: res.Data.pid || 0,
            started_at: res.Data.started_at || 0
          }
          if (res.Data.port) {
            this.headroomConfig = {
              port: res.Data.port || 8787,
              anthropic_api_url: res.Data.anthropic_api_url || '',
              openai_api_url: res.Data.openai_api_url || '',
              gemini_api_url: res.Data.gemini_api_url || '',
              cloudcode_api_url: res.Data.cloudcode_api_url || '',
              vertex_api_url: res.Data.vertex_api_url || ''
            }
          }
        }
      })
    },
    showHeadroomConfig(row) {
      this.showHeadroomConfigDialog = true
    },
    saveHeadroomConfig() {
      Base.BasePost('/api/AgentV2HeadroomConfigSave', {
        agent_id: this.agentId,
        config: this.headroomConfig
      }, () => {
        this.$message.success('Headroom 配置已保存')
        this.showHeadroomConfigDialog = false
        this.loadHeadroomStatus()
      })
    },
    headroomProcess(action) {
      const labels = { start: '启动', stop: '停止', restart: '重启' }
      Base.BasePost('/api/AgentV2HeadroomProcess', {
        agent_id: this.agentId,
        action: action
      }, (res) => {
        if (res.ErrCode === 0) {
          this.$message.success(`Headroom ${labels[action] || action}成功`)
          this.loadHeadroomStatus()
        }
      })
    },
    showEnvToolInstall(tool) {
      this.envToolDialogData = tool
      this.showEnvToolDialog = true
    },
    openEnvToolHomepage(tool) {
      if (tool.homepage) window.open(tool.homepage, '_blank')
    },
    envToolStatusType(row) {
      if (row.key === 'headroom') {
        if (this.headroomStatus.running) return 'success'
        return row.installed ? 'warning' : 'info'
      }
      if (row.extension_installed) return 'success'
      if (row.installed) return 'warning'
      return 'info'
    },
    envToolStatusLabel(row) {
      if (row.key === 'headroom') {
        if (this.headroomStatus.running) return '代理运行中'
        return row.installed ? '已安装' : '未安装'
      }
      if (row.extension_installed) return '已安装'
      if (row.installed) return '已安装(未激活)'
      return '未安装'
    },
    envToolRemove(row) {
      this.$confirm(`确定从 Pi 扩展目录中移除「${row.name}」？将删除对应的 .ts 文件。`, '确认移除', { type: 'warning' }).then(() => {
        Base.BasePost('/api/AgentV2EnvToolAction', {
          agent_id: this.agentId,
          key: row.key,
          action: 'remove'
        }, () => {
          this.$message.success('已移除')
          this.loadEnvTools()
        })
      }).catch(() => {})
    },
    loadInstalledTools() {
      Base.BasePost('/api/AgentV2InstalledToolList', {}, (res) => {
        this.installedTools = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
      })
    },
    removeInstalledTool(row) {
      this.$confirm(`确定删除扩展「${row.name}.ts」？此操作不可撤销。`, '确认删除', { type: 'warning' }).then(() => {
        Base.BasePost('/api/AgentV2InstalledToolRemove', { name: row.name }, () => {
          this.$message.success('已删除')
          this.loadInstalledTools()
        })
      }).catch(() => {})
    },
    envToolAction(row, action) {
      const actionLabels = { activate: '激活', deactivate: '停用' }
      const label = actionLabels[action] || action
      this.$confirm(`确定${label}「${row.name}」？操作后可能需要重启 Pi Agent 会话才能生效。`, '确认操作', { type: 'info' }).then(() => {
        Base.BasePost('/api/AgentV2EnvToolAction', {
          agent_id: this.agentId,
          key: row.key,
          action: action
        }, (res) => {
          if (res.ErrCode === 0 && res.Data) {
            const cmd = res.Data.command || ''
            const msg = res.Data.message || ''
            this.$alert(
              `<div style="margin-bottom:8px">${msg}</div>
               <div style="background:#f5f7fa;border-radius:4px;padding:10px;font-family:'Cascadia Code',monospace;font-size:13px;word-break:break-all">${cmd}</div>`,
              `${label}命令`,
              { dangerouslyUseHTMLString: true, confirmButtonText: '复制并关闭', beforeClose: (action, instance, done) => {
                if (action === 'confirm' && cmd) {
                  navigator.clipboard.writeText(cmd).then(() => {
                    this.$message.success('命令已复制到剪贴板')
                  })
                }
                done()
              }}
            )
            this.loadEnvTools()
          }
        })
      }).catch(() => {})
    },

    // 工作空间
    loadWorkspaces() {
      Base.BasePost('/api/AgentV2WorkspaceList', { agent_id: this.agentId }, (res) => {
        this.workspaces = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
      })
    },
    saveWorkspace() {
      if (!this.workspaceForm.name || !this.workspaceForm.path) {
        this.$message.warning('请填写名称和路径')
        return
      }
      Base.BasePost('/api/AgentV2WorkspaceSave', {
        agent_id: this.agentId,
        name: this.workspaceForm.name,
        path: this.workspaceForm.path
      }, () => {
        this.showWorkspaceDialog = false
        this.workspaceForm = { name: '', path: '' }
        this.loadWorkspaces()
      })
    },
    deleteWorkspace(row) {
      this.$confirm('确定删除此工作空间？', '提示', { type: 'warning' }).then(() => {
        Base.BasePost('/api/AgentV2WorkspaceDelete', { id: row.id }, () => { this.loadWorkspaces() })
      }).catch(() => {})
    },

    typeLabel(type) {
      const map = { pi: 'Pi', codex: 'Codex CLI', 'claude-code': 'Claude Code' }
      return map[type] || type
    }
  }
}
</script>

<style scoped>
.agent-config {
  padding: 24px;
  height: calc(100vh - 60px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.config-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  flex-shrink: 0;
}
.config-header h2 { margin: 0; font-size: 20px; }

.config-body {
  flex: 1;
  overflow: hidden;
}

.config-tabs {
  height: 100%;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}
.config-tabs :deep(.el-tabs__header) {
  width: 140px;
  border-right: 1px solid #e4e7ed;
  background: #fafafa;
  border-radius: 8px 0 0 8px;
}
.config-tabs :deep(.el-tabs__nav-wrap) {
  padding-top: 12px;
}
.config-tabs :deep(.el-tabs__item) {
  height: 44px;
  line-height: 44px;
  text-align: left;
  padding-left: 24px !important;
  padding-right: 24px !important;
  font-size: 14px;
  color: #606266;
}
.config-tabs :deep(.el-tabs__item.is-active) {
  color: #409eff;
  background: #ecf5ff;
}
.config-tabs :deep(.el-tabs__content) {
  padding: 24px 32px;
  overflow-y: auto;
  height: 100%;
}

.config-form {
  max-width: 680px;
}

.config-table {
  margin-top: 16px;
}

.skills-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
}
.skills-hint {
  font-size: 12px;
  color: #909399;
}

.envtools-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.field-hint {
  font-size: 11px;
  color: #c0c4cc;
  margin-top: 4px;
  line-height: 1.4;
}

.param-list { width: 100%; }
.param-row {
  display: flex;
  gap: 6px;
  align-items: center;
  margin-bottom: 6px;
}

.script-editor-wrapper {
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
}
.script-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #dcdfe6;
}
.script-path {
  font-size: 12px;
  color: #909399;
  font-family: 'Cascadia Code', 'Fira Code', monospace;
}
.script-lang {
  font-size: 11px;
  color: #c0c4cc;
}
.script-editor-wrapper :deep(.el-textarea__inner) {
  border: none;
  border-radius: 0;
  resize: vertical;
  min-height: 280px;
}
.script-editor-wrapper :deep(.el-textarea__inner:focus) {
  box-shadow: none;
}

/* ===== 模型配置 Tab ===== */
.model-config-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.provider-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.provider-card {
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  overflow: hidden;
  transition: border-color .2s;
}
.provider-card:hover { border-color: #c0c4cc; }
.provider-card--expanded { border-color: #409eff; }

.provider-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  user-select: none;
}
.provider-card__header:hover { background: #fafafa; }

.provider-card__info {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.provider-card__name {
  font-weight: 600;
  font-size: 14px;
}
.provider-card__url {
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 260px;
}

.provider-card__actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}
.provider-card__arrow {
  margin-left: 6px;
  font-size: 14px;
  color: #c0c4cc;
  transition: transform .2s;
}
.provider-card__arrow--open { transform: rotate(90deg); }

.provider-card__models {
  border-top: 1px solid #ebeef5;
  padding: 12px 16px 16px;
}

.url-preview {
  font-size: 12px;
  color: #909399;
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  background: #f5f7fa;
  padding: 8px 12px;
  border-radius: 4px;
  word-break: break-all;
}

.empty-hint { padding: 16px; text-align: center; color: #c0c4cc; font-size: 13px; }


</style>
