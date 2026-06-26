<template>
  <div class="wf-template-manager">
    <!-- 左侧模板列表 -->
    <div class="wf-template-manager__left">
      <div class="wf-template-manager__section-title">
        <span>模板列表</span>
        <div class="wf-template-manager__section-actions">
          <el-button size="small" @click="triggerImportTemplate">
            <el-icon><Upload /></el-icon> 导入
          </el-button>
          <el-button type="primary" size="small" @click="addTemplate">
            <el-icon><Plus /></el-icon> 新建
          </el-button>
        </div>
      </div>
      <input
        ref="importFileInput"
        type="file"
        accept=".json"
        style="display: none;"
        @change="handleImportFile"
      />
      <div class="wf-template-manager__template-list">
        <div
          v-for="tpl in templates"
          :key="tpl.id"
          :class="['wf-template-manager__template-item', {
            'wf-template-manager__template-item--active': selectedTemplateId === tpl.id,
            'wf-template-manager__template-item--default': tpl.is_default === 1
          }]"
          @click="selectTemplate(tpl)"
        >
          <div class="wf-template-manager__template-name">
            {{ tpl.name }}
            <el-tag v-if="tpl.is_default === 1" size="small" type="success" effect="plain">默认</el-tag>
          </div>
          <div class="wf-template-manager__template-desc">{{ tpl.description || '无描述' }}</div>
          <div class="wf-template-manager__template-actions">
            <el-button v-if="tpl.is_default !== 1" text size="small" type="warning" @click.stop="setTemplateDefault(tpl)">设置默认</el-button>
            <el-button text size="small" type="primary" @click.stop="openTemplateEditDialog(tpl)">编辑</el-button>
            <el-button text size="small" type="success" @click.stop="exportTemplate(tpl)">导出</el-button>
            <el-button v-if="tpl.is_default !== 1" text size="small" type="danger" @click.stop="deleteTemplateConfirm(tpl)">删除</el-button>
          </div>
        </div>
        <div v-if="templates.length === 0" class="wf-template-manager__empty">
          暂无模板，请点击"新建"创建
        </div>
      </div>
    </div>

    <!-- 右侧步骤编辑区域 -->
    <div class="wf-template-manager__right" v-if="selectedTemplate">
      <div class="wf-template-manager__section-title">
        <span>步骤列表（拖拽排序）</span>
        <div class="wf-template-manager__section-actions">
          <el-button size="small" @click="openPlaceholderDialog">
            <el-icon><Document /></el-icon> 内置占位符
          </el-button>
          <el-button type="success" size="small" @click="addStep">
            <el-icon><Plus /></el-icon> 添加步骤
          </el-button>
        </div>
      </div>

      <div class="wf-template-manager__step-list" ref="stepListRef">
        <draggable
          v-model="editingSteps"
          item-key="id"
          handle=".wf-template-manager__step-drag-handle"
          ghost-class="wf-template-manager__step-ghost"
          @end="onStepDragEnd"
        >
          <template #item="{ element, index }">
            <div :class="['wf-template-manager__step-item', {
              'wf-template-manager__step-item--fixed': element.is_fixed === 1,
            }]">
              <div class="wf-template-manager__step-header">
                <!-- 拖拽手柄 -->
                <span class="wf-template-manager__step-drag-handle" v-if="element.is_fixed !== 1">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                    <circle cx="9" cy="5" r="1.5"/><circle cx="15" cy="5" r="1.5"/>
                    <circle cx="9" cy="12" r="1.5"/><circle cx="15" cy="12" r="1.5"/>
                    <circle cx="9" cy="19" r="1.5"/><circle cx="15" cy="19" r="1.5"/>
                  </svg>
                </span>
                <!-- 固定步骤锁图标 -->
                <span class="wf-template-manager__step-lock" v-if="element.is_fixed === 1" title="固定步骤，不可删除">
                  <el-icon><Lock /></el-icon>
                </span>
                <span class="wf-template-manager__step-index">{{ index + 1 }}.</span>
                <el-input
                  v-model="element.name"
                  placeholder="步骤名称"
                  size="small"
                  class="wf-template-manager__step-name"
                  :disabled="element.is_fixed === 1 && element.step_key === 'task-config'"
                  @blur="saveStepSilent(element)"
                />
                <div class="wf-template-manager__step-actions">
                  <el-button
                    v-if="showStepDocumentButton(element)"
                    text
                    size="small"
                    type="info"
                    @click="openDocumentDialog(element)"
                  >
                    文档
                    <el-badge v-if="(element.step_documents_list || []).length > 0" :value="element.step_documents_list.length" class="wf-template-manager__doc-badge" />
                  </el-button>
                  <el-button
                    text
                    size="small"
                    type="warning"
                    @click="openRemarkDialog(element)"
                  >
                    备注
                    <el-icon v-if="element.remark" style="margin-left: 2px;"><Document /></el-icon>
                  </el-button>
                  <el-button
                    text
                    size="small"
                    type="primary"
                    :disabled="element.step_key === 'task-config' || element.step_key === 'requirement-fetch'"
                    :title="element.step_key === 'requirement-fetch' ? '抓取需求为固定步骤，提示词由系统维护' : ''"
                    @click="openPromptDialog(element)"
                  >
                    提示词
                  </el-button>
                  <el-button
                    text
                    size="small"
                    type="danger"
                    :disabled="element.is_fixed === 1"
                    :title="element.is_fixed === 1 ? '固定步骤，不可删除' : ''"
                    @click.stop="deleteStepConfirm(element)"
                  >删除</el-button>
                </div>
              </div>
            </div>
          </template>
        </draggable>

        <div v-if="editingSteps.length === 0" class="wf-template-manager__empty">
          暂无步骤，请点击"添加步骤"
        </div>
      </div>
    </div>

    <!-- 未选择模板时的占位 -->
    <div class="wf-template-manager__right" v-else>
      <div class="wf-template-manager__empty wf-template-manager__empty--full">
        请从左侧选择一个模板，或新建一个模板
      </div>
    </div>

    <!-- 模板信息编辑弹窗 -->
    <el-dialog
      v-model="templateEditDialogVisible"
      title="编辑模板信息"
      width="500px"
      :close-on-click-modal="true"
      class="wf-template-edit-dialog"
    >
      <div v-if="editingTemplate">
        <el-form label-width="80px">
          <el-form-item label="模板名称">
            <el-input
              v-model="dialogTemplateName"
              placeholder="模板名称（必填）"
              maxlength="50"
              show-word-limit
            />
          </el-form-item>
          <el-form-item label="模板备注">
            <el-input
              v-model="dialogTemplateDesc"
              placeholder="模板备注（选填）"
              maxlength="200"
              show-word-limit
              type="textarea"
              :rows="4"
            />
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="templateEditDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="savingTemplate" @click="saveTemplateFromDialog">保存</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 内置占位符弹窗 -->
    <el-dialog
      v-model="placeholderDialogVisible"
      title="内置占位符"
      width="720px"
      :close-on-click-modal="true"
      class="wf-placeholder-dialog"
    >
      <div class="wf-placeholder-list">
        <div
          v-for="ph in displayPlaceholders"
          :key="ph.value"
          class="wf-placeholder-item"
          @click="copyPlaceholder(ph)"
        >
          <div class="wf-placeholder-item__content">
            <div class="wf-placeholder-item__row">
              <span class="wf-placeholder-item__label">{{ ph.label }}</span>
              <code class="wf-placeholder-item__value">{{ ph.value }}</code>
            </div>
            <div v-if="ph.tip" class="wf-placeholder-item__tip">{{ ph.tip }}</div>
          </div>
          <el-icon class="wf-placeholder-item__copy"><CopyDocument /></el-icon>
        </div>
      </div>
    </el-dialog>

    <!-- 提示词编辑弹窗 -->
    <el-dialog
      v-model="promptDialogVisible"
      :title="`编辑提示词 - ${editingStep ? editingStep.name : ''}`"
      width="80%"
      :close-on-click-modal="true"
      class="wf-prompt-dialog"
    >
      <div v-if="editingStep && isRequirementFetchStep(editingStep)" class="wf-prompt-dialog__tip">
        抓取需求为固定步骤，提示词由系统维护，不可手动编辑。
      </div>
      <!-- 内置占位符快速复制区域 -->
      <div v-if="promptDialogPlaceholders.length > 0" class="wf-prompt-dialog__placeholders">
        <span class="wf-prompt-dialog__placeholders-label">可用占位符：</span>
        <code
          v-for="ph in promptDialogPlaceholders"
          :key="ph.value"
          :class="['wf-prompt-dialog__placeholder-tag', `wf-prompt-dialog__placeholder-tag--${ph.group}`]"
          @click="copyPlaceholder(ph)"
        >{{ ph.value }}</code>
      </div>
      <UnifiedMdEditor
        v-if="editingStep"
        v-model="dialogPromptContent"
        :toolbars="promptEditorToolbars"
        class="wf-prompt-dialog__editor"
        :disabled="editingStep && isRequirementFetchStep(editingStep)"
      />
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="promptDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            :disabled="!editingStep || isRequirementFetchStep(editingStep)"
            @click="saveDialogPrompt"
          >
            保存提示词
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 步骤文档配置弹窗 -->
    <el-dialog
      v-model="documentDialogVisible"
      :title="`配置步骤文档 - ${editingStep ? editingStep.name : ''}`"
      width="720px"
      :close-on-click-modal="true"
      class="wf-document-dialog"
    >
      <div v-if="editingStep" class="wf-document-dialog__content">
        <div class="wf-document-dialog__tip">
          为当前步骤预生成若干知识片段文档，创建任务时会自动创建这些片段，并生成对应的占位符供提示词使用。
        </div>
        <div
          v-for="(doc, index) in editingStepDocuments"
          :key="index"
          class="wf-document-item"
        >
          <div class="wf-document-item__header">
            <span class="wf-document-item__index">文档 {{ index + 1 }}</span>
            <div class="wf-document-item__header-right">
              <div class="wf-document-item__api-doc-switch">
                <span class="wf-document-item__api-doc-label">接口文档</span>
                <el-switch v-model="doc.is_api_doc" size="small" />
              </div>
              <el-button text size="small" type="danger" @click="removeStepDocument(index)">删除</el-button>
            </div>
          </div>
          <div class="wf-document-item__body">
            <el-input
              v-model="doc.name"
              placeholder="文档名称（必填，如：接口文档）"
              size="small"
              class="wf-document-item__input"
              maxlength="50"
              show-word-limit
            />
            <div class="wf-document-item__input wf-document-item__placeholder">
              <span class="wf-document-item__placeholder-label">占位符</span>
              <el-input
                v-model="doc.placeholder"
                placeholder="请手动填写占位符，如：{接口文档地址}"
                size="small"
                class="wf-document-item__placeholder-input"
              />
            </div>
            <el-input
              v-model="doc.title"
              placeholder="片段标题（选填，支持 {任务名称} 等占位符）"
              size="small"
              class="wf-document-item__input"
            />
            <el-input
              v-model="doc.content"
              type="textarea"
              :rows="4"
              placeholder="默认内容（选填，支持占位符，创建任务时会写入片段）"
              size="small"
              class="wf-document-item__input"
            />
            <div class="wf-document-item__preview">
              <code>{{ doc.placeholder || '未设置占位符' }}</code>
              <span v-if="doc.placeholder">创建任务后可替换为文档分享链接</span>
              <el-tag v-if="doc.is_api_doc" type="success" size="small" effect="plain">接口文档</el-tag>
            </div>
          </div>
        </div>
        <div class="wf-document-dialog__actions">
          <el-button type="primary" size="small" @click="addStepDocument">
            <el-icon><Plus /></el-icon> 添加文档
          </el-button>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="documentDialogVisible = false">取消</el-button>
          <el-button type="primary" :disabled="!editingStep" @click="saveStepDocuments">
            保存文档配置
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 步骤备注编辑弹窗 -->
    <el-dialog
      v-model="remarkDialogVisible"
      :title="`编辑备注 - ${editingStep ? editingStep.name : ''}`"
      width="500px"
      :close-on-click-modal="true"
      class="wf-remark-dialog"
    >
      <div v-if="editingStep">
        <el-input
          v-model="dialogRemarkContent"
          type="textarea"
          :rows="6"
          placeholder="请输入步骤备注（选填）"
          maxlength="500"
          show-word-limit
        />
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="remarkDialogVisible = false">取消</el-button>
          <el-button type="primary" :disabled="!editingStep" @click="saveDialogRemark">
            保存备注
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { Plus, Lock, CopyDocument, Document, Upload } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import draggable from 'vuedraggable'
import UnifiedMdEditor from '@/components/base/UnifiedMdEditor.vue'
import workflowTemplateApi from '@/utils/base/workflow_template'

let stepIdCounter = 0

// 提示词编辑器工具栏配置
const PROMPT_EDITOR_TOOLBARS = [
  'bold', 'underline', 'italic', '-', 'strikeThrough', 'title', 'sub', 'sup', 'quote', 'unorderedList', 'orderedList', '-', 'codeRow', 'code', 'link', 'image', 'table',
]

// 抓取需求步骤默认文档占位符：指向创建任务时自动保存的需求知识片段
const REQUIREMENT_FETCH_DEFAULT_DOC_PLACEHOLDER = '{需求文档地址}'

// 内置占位符列表（不包含 skills 和步骤文档占位符，这两类会动态生成）
const PROMPT_PLACEHOLDERS = [
  { label: '任务名称', value: '{任务名称}', tip: '替换为当前任务的名称' },
  { label: '步骤ID', value: '{步骤ID}', tip: '替换为当前步骤的 step_key，用于更新步骤执行状态' },
  { label: '需求文档地址', value: '{需求文档地址}', tip: '替换为需求知识片段的分享链接' },
  { label: '需求文档地址ID', value: '{需求文档地址ID}', tip: '替换为需求知识片段的 file_id' },
  { label: '接口开发API地址', value: '{接口开发API地址}', tip: '替换为当前服务的 API 基地址（scheme://host）' },
  { label: '接口开发API的token', value: '{接口开发API的token}', tip: '替换为请求的 Authorization token' },
  { label: '开发项目配置', value: '{开发项目配置}', tip: '替换为开发项目配置的 Markdown 列表' },
  { label: '自定义网页', value: '{自定义网页}', tip: '替换为智能链接（smart_link）的名称和 ID' },
  { label: '账号', value: '{账号}', tip: '替换为智能链接的账号（smart_link_account）' },
  { label: '工作流程ID', value: '{工作流程ID}', tip: '替换为当前工作流程的 ID' },
  { label: '任务ID', value: '{任务ID}', tip: '替换为当前任务的 ID' },
  { label: '开发环境', value: '{开发环境}', tip: '替换为开发环境配置（已递归解析内部占位符）' },
]

export default {
  name: 'WorkflowTemplateManager',
  components: {
    Plus,
    Lock,
    CopyDocument,
    Document,
    Upload,
    draggable,
    UnifiedMdEditor,
  },
  emits: ['templates-loaded'],
  computed: {
    displayPlaceholders() {
      const docs = []
      ;(this.editingSteps || []).forEach(step => {
        (step.step_documents_list || []).forEach(doc => {
          if (doc.placeholder) {
            docs.push({
              label: `${step.name} - ${doc.name}`,
              value: doc.placeholder,
              tip: '创建任务后替换为对应知识片段的分享链接',
              group: 'document',
            })
            // 生成文档ID占位符
            const inner = doc.placeholder.replace(/^\{|\}$/g, '')
            const idPlaceholder = `{${inner}ID}`
            docs.push({
              label: `${step.name} - ${doc.name}（ID）`,
              value: idPlaceholder,
              tip: '创建任务后替换为对应知识片段的 file_id',
              group: 'document',
            })
          }
        })
      })
      // 动态生成 skills 占位符
      const skillPlaceholders = (this.skillList || []).map(name => ({
        label: `${name}地址`,
        value: `{${name}地址}`,
        tip: `替换为 skills/${name} 目录的本地路径`,
        group: 'skill',
      }))
      return this.promptPlaceholders.map(p => ({ ...p, group: 'builtin' })).concat(docs).concat(skillPlaceholders)
    },
    // 当前编辑步骤的所有可用占位符（提示词弹窗顶部展示）
    promptDialogPlaceholders() {
      return this.displayPlaceholders
    },
  },
  data() {
    return {
      templates: [],
      selectedTemplateId: 0,
      selectedTemplate: null,
      editTemplateName: '',
      editTemplateDesc: '',
      editingSteps: [],
      savingTemplate: false,
      loading: false,
      placeholderDialogVisible: false,
      promptDialogVisible: false,
      editingStep: null,
      dialogPromptContent: '',
      documentDialogVisible: false,
      editingStepDocuments: [],
      remarkDialogVisible: false,
      dialogRemarkContent: '',
      promptEditorToolbars: PROMPT_EDITOR_TOOLBARS,
      promptPlaceholders: PROMPT_PLACEHOLDERS,
      skillList: [],
      // 模板信息编辑弹窗
      templateEditDialogVisible: false,
      editingTemplate: null,
      dialogTemplateName: '',
      dialogTemplateDesc: '',
      // 导入
      importingTemplate: false,
    }
  },
  mounted() {
    this.loadTemplates()
    this.loadSkillList()
  },
  methods: {
    // ===== 内置占位符弹窗 =====
    openPlaceholderDialog() {
      this.placeholderDialogVisible = true
    },
    closePlaceholderDialog() {
      this.placeholderDialogVisible = false
    },
    copyPlaceholder(placeholder) {
      const text = placeholder.value
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
          ElMessage.success(`已复制：${text}`)
        }).catch(() => {
          this.fallbackCopy(text)
        })
      } else {
        this.fallbackCopy(text)
      }
    },
    fallbackCopy(text) {
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
      ElMessage.success(`已复制：${text}`)
    },

    // ===== Skills 列表加载 =====
    loadSkillList() {
      workflowTemplateApi.WorkflowSkillList((response) => {
        if (response && response.ErrCode === 0 && response.Data && response.Data.list) {
          this.skillList = response.Data.list.map(item => item.name)
        }
      })
    },

    // ===== 模板数据加载 =====
    loadTemplates() {
      this.loading = true
      workflowTemplateApi.WorkflowTemplateList((response) => {
        this.loading = false
        if (response && response.ErrCode === 0 && response.Data && response.Data.list) {
          this.templates = response.Data.list
          // 设置步骤的 _key 用于展开/折叠跟踪
          this.templates.forEach(t => {
            if (t.steps) {
              t.steps.forEach(s => {
                s._key = 'step_' + s.id
              })
            }
          })
          this.$emit('templates-loaded', this.templates)
          // 如果有模板且未选中，自动选中第一个
          if (this.templates.length > 0 && this.selectedTemplateId === 0) {
            this.selectTemplate(this.templates[0])
          }
        }
      })
    },

    // ===== 模板选择 =====
    selectTemplate(tpl) {
      this.selectedTemplateId = tpl.id
      this.selectedTemplate = tpl
      this.editTemplateName = tpl.name
      this.editTemplateDesc = tpl.description || ''
      this.editingSteps = (tpl.steps || []).map(s => {
        const step = {
          ...s,
          _key: 'step_' + (s.id || ++stepIdCounter),
        }
        step.step_documents_list = this.parseStepDocuments(s.step_documents)
        return step
      })
    },

    // ===== 模板 CRUD =====
    addTemplate() {
      const newTpl = {
        id: 0,
        name: '',
        description: '',
        is_default: 0,
        steps: [],
      }
      this.openTemplateEditDialog(newTpl)
    },

    saveTemplate() {
      if (!this.editTemplateName || !this.editTemplateName.trim()) {
        ElMessage.warning('请填写模板名称')
        return
      }
      this.savingTemplate = true
      workflowTemplateApi.WorkflowTemplateSave({
        id: this.selectedTemplateId || 0,
        name: this.editTemplateName.trim(),
        description: this.editTemplateDesc.trim(),
      }, (response) => {
        this.savingTemplate = false
        if (response && response.ErrCode === 0) {
          const saved = response.Data.template
          if (saved && saved.steps) {
            saved.steps.forEach(s => { s._key = 'step_' + s.id })
          }
          // 更新列表中的模板
          const idx = this.templates.findIndex(t => t.id === this.selectedTemplateId)
          if (idx >= 0) {
            this.templates[idx] = saved
          } else if (this.selectedTemplateId === 0) {
            // 新增模板，替换占位项
            const placeholderIdx = this.templates.findIndex(t => t.id === 0 && t.name === '新模板')
            if (placeholderIdx >= 0) {
              this.templates[placeholderIdx] = saved
            }
          }
          this.selectedTemplateId = saved.id
          this.selectedTemplate = saved
          this.editingSteps = (saved.steps || []).map(s => {
            const step = {
              ...s,
              _key: 'step_' + s.id,
            }
            step.step_documents_list = this.parseStepDocuments(s.step_documents)
            return step
          })
          ElMessage.success('模板已保存')
          this.$emit('templates-loaded', this.templates)
        } else {
          ElMessage.error(response.ErrMsg || '保存失败')
        }
      })
    },

    deleteTemplateConfirm(tpl) {
      ElMessageBox.confirm(`确定删除模板"${tpl.name}"吗？此操作不可恢复。`, '确认删除', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        this.deleteTemplate(tpl)
      }).catch(() => {})
    },

    deleteTemplate(tpl) {
      workflowTemplateApi.WorkflowTemplateDelete(tpl.id, (response) => {
        if (response && response.ErrCode === 0) {
          this.templates = this.templates.filter(t => t.id !== tpl.id)
          if (this.selectedTemplateId === tpl.id) {
            this.selectedTemplateId = 0
            this.selectedTemplate = null
            this.editingSteps = []
          }
          ElMessage.success('模板已删除')
          this.$emit('templates-loaded', this.templates)
        } else {
          ElMessage.error(response.ErrMsg || '删除失败')
        }
      })
    },

    setTemplateDefault(tpl) {
      ElMessageBox.confirm(`确定将"${tpl.name}"设为默认模板吗？`, '设置默认模板', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        workflowTemplateApi.WorkflowTemplateSetDefault(tpl.id, (response) => {
          if (response && response.ErrCode === 0) {
            // 更新所有模板的 is_default 状态
            this.templates.forEach(t => {
              t.is_default = t.id === tpl.id ? 1 : 0
            })
            ElMessage.success(`已将"${tpl.name}"设为默认模板`)
            this.$emit('templates-loaded', this.templates)
          } else {
            ElMessage.error(response.ErrMsg || '设置失败')
          }
        })
      }).catch(() => {})
    },

    // ===== 步骤 CRUD =====
    addStep() {
      if (!this.selectedTemplateId) {
        ElMessage.warning('请先保存模板后再添加步骤')
        return
      }
      const newStep = {
        id: 0,
        template_id: this.selectedTemplateId,
        name: '新步骤',
        step_key: '',
        prompt_content: '',
        sort_order: this.editingSteps.length,
        is_fixed: 0,
        _key: 'step_new_' + (++stepIdCounter),
      }
      this.editingSteps.push(newStep)
    },

    openPromptDialog(step) {
      this.editingStep = step
      this.dialogPromptContent = step.prompt_content || ''
      this.promptDialogVisible = true
    },
    closePromptDialog() {
      this.editingStep = null
      this.dialogPromptContent = ''
      this.promptDialogVisible = false
    },
    isRequirementFetchStep(step) {
      return step && step.step_key === 'requirement-fetch'
    },
    saveDialogPrompt() {
      if (!this.editingStep) return
      this.editingStep.prompt_content = this.dialogPromptContent
      this.saveStepPrompt(this.editingStep, () => {
        this.promptDialogVisible = false
      })
    },

    // ===== 步骤文档配置 =====
    showStepDocumentButton(step) {
      // 仅任务配置固定步骤不显示文档配置按钮；抓取需求允许配置默认需求文档
      if (!step) return false
      return step.step_key !== 'task-config'
    },
    createRequirementFetchDefaultDocument() {
      return {
        id: this.generateStepDocumentId(),
        name: '需求文档',
        placeholder: REQUIREMENT_FETCH_DEFAULT_DOC_PLACEHOLDER,
        title: '',
        content: '',
        is_api_doc: false,
      }
    },
    parseStepDocuments(raw) {
      if (!raw) return []
      try {
        const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
        if (!Array.isArray(parsed)) return []
        return parsed.map(doc => {
          const id = String(doc.id || '').trim() || this.generateStepDocumentId()
          return {
            id,
            name: String(doc.name || '').trim(),
            placeholder: String(doc.placeholder || '').trim(),
            title: String(doc.title || '').trim(),
            content: String(doc.content || '').trim(),
            is_api_doc: !!doc.is_api_doc,
          }
        }).filter(doc => doc.name)
      } catch (e) {
        return []
      }
    },
    stringifyStepDocuments(docs) {
      if (!Array.isArray(docs) || docs.length === 0) return ''
      const list = docs.map(doc => {
        const id = String(doc.id || '').trim() || this.generateStepDocumentId()
        return {
          id,
          name: String(doc.name || '').trim(),
          placeholder: String(doc.placeholder || '').trim(),
          title: String(doc.title || '').trim(),
          content: String(doc.content || '').trim(),
          is_api_doc: !!doc.is_api_doc,
        }
      }).filter(doc => doc.name)
      return list.length > 0 ? JSON.stringify(list) : ''
    },
    generateStepDocumentId() {
      return 'doc_' + Date.now().toString(36) + '_' + Math.random().toString(36).substr(2, 4)
    },
    generateStepDocumentPlaceholder(id) {
      return `{文档_${id}_地址}`
    },
    openDocumentDialog(step) {
      this.editingStep = step
      this.editingStepDocuments = (step.step_documents_list || []).map(doc => ({ ...doc }))
      if (this.editingStepDocuments.length === 0) {
        if (this.isRequirementFetchStep(step)) {
          this.editingStepDocuments.push(this.createRequirementFetchDefaultDocument())
        } else {
          this.editingStepDocuments.push(this.createEmptyStepDocument())
        }
      }
      this.documentDialogVisible = true
    },
    closeDocumentDialog() {
      this.editingStep = null
      this.editingStepDocuments = []
      this.documentDialogVisible = false
    },
    createEmptyStepDocument() {
      const id = this.generateStepDocumentId()
      return {
        id,
        name: '',
        placeholder: '',
        title: '',
        content: '',
        is_api_doc: false,
      }
    },
    addStepDocument() {
      this.editingStepDocuments.push(this.createEmptyStepDocument())
    },
    removeStepDocument(index) {
      this.editingStepDocuments.splice(index, 1)
    },
    validateStepDocuments() {
      for (let i = 0; i < this.editingStepDocuments.length; i++) {
        const doc = this.editingStepDocuments[i]
        const name = String(doc.name || '').trim()
        if (!name) {
          ElMessage.warning(`文档 ${i + 1} 的名称不能为空`)
          return false
        }
      }
      // 校验占位符在同一模板内是否重复
      const currentPlaceholders = new Map()
      for (let i = 0; i < this.editingStepDocuments.length; i++) {
        const ph = String(this.editingStepDocuments[i].placeholder || '').trim()
        if (!ph) continue
        if (currentPlaceholders.has(ph)) {
          ElMessage.warning(`文档占位符 ${ph} 在当前步骤内重复（文档 ${currentPlaceholders.get(ph)} 与 文档 ${i + 1}）`)
          return false
        }
        currentPlaceholders.set(ph, i + 1)
      }
      // 跨步骤校验
      for (const step of this.editingSteps) {
        if (step === this.editingStep) continue
        const stepDocs = step.step_documents_list || []
        for (const doc of stepDocs) {
          const otherPh = String(doc.placeholder || '').trim()
          if (!otherPh) continue
          if (currentPlaceholders.has(otherPh)) {
            ElMessage.warning(`文档占位符 ${otherPh} 与步骤"${step.name}"中的文档"${doc.name}"重复`)
            return false
          }
        }
      }
      return true
    },
    saveStepDocuments() {
      if (!this.editingStep) return
      if (!this.validateStepDocuments()) return
      this.editingStepDocuments.forEach(doc => {
        if (!String(doc.id || '').trim()) {
          doc.id = this.generateStepDocumentId()
        }
      })
      this.editingStepDocuments = this.editingStepDocuments.filter(doc => String(doc.name || '').trim())
      this.editingStep.step_documents = this.stringifyStepDocuments(this.editingStepDocuments)
      this.editingStep.step_documents_list = this.parseStepDocuments(this.editingStep.step_documents)
      this.saveStepSilentWithDocuments(this.editingStep, () => {
        this.documentDialogVisible = false
      })
    },
    saveStepSilentWithDocuments(step, callback) {
      if (!step.name || !step.name.trim()) return
      workflowTemplateApi.WorkflowTemplateStepSave({
        id: step.id || 0,
        template_id: this.selectedTemplateId,
        name: step.name.trim(),
        step_key: step.step_key || '',
        prompt_content: step.prompt_content || '',
        step_documents: step.step_documents || '',
        remark: step.remark || '',
        sort_order: step.sort_order || 0,
      }, (response) => {
        if (response && response.ErrCode === 0 && response.Data && response.Data.step) {
          const saved = response.Data.step
          saved._key = 'step_' + saved.id
          saved.step_documents_list = this.parseStepDocuments(saved.step_documents)
          Object.assign(step, saved)
          ElMessage.success('文档配置已保存')
          if (callback) callback()
        } else {
          ElMessage.error(response.ErrMsg || '保存失败')
        }
      })
    },

    saveStepSilent(step) {
      if (!step.name || !step.name.trim()) return
      // 仅名称变更时静默保存
      workflowTemplateApi.WorkflowTemplateStepSave({
        id: step.id || 0,
        template_id: this.selectedTemplateId,
        name: step.name.trim(),
        step_key: step.step_key || '',
        prompt_content: step.prompt_content || '',
        step_documents: step.step_documents || '',
        remark: step.remark || '',
        sort_order: step.sort_order || 0,
      }, (response) => {
        if (response && response.ErrCode === 0 && response.Data && response.Data.step) {
          const saved = response.Data.step
          saved._key = 'step_' + saved.id
          saved.step_documents_list = this.parseStepDocuments(saved.step_documents)
          Object.assign(step, saved)
        }
      })
    },

    saveStepPrompt(step, callback) {
      workflowTemplateApi.WorkflowTemplateStepSave({
        id: step.id || 0,
        template_id: this.selectedTemplateId,
        name: step.name.trim(),
        step_key: step.step_key || '',
        prompt_content: step.prompt_content || '',
        step_documents: step.step_documents || '',
        remark: step.remark || '',
        sort_order: step.sort_order || 0,
      }, (response) => {
        if (response && response.ErrCode === 0 && response.Data && response.Data.step) {
          const saved = response.Data.step
          saved._key = 'step_' + saved.id
          Object.assign(step, saved)
          ElMessage.success('提示词已保存')
          if (callback) callback()
        } else {
          ElMessage.error(response.ErrMsg || '保存失败')
        }
      })
    },

    deleteStepConfirm(step) {
      if (!step.id) {
        // 未保存的步骤直接从列表中移除
        this.editingSteps = this.editingSteps.filter(s => s._key !== step._key)
        return
      }
      ElMessageBox.confirm(`确定删除步骤"${step.name}"吗？`, '确认删除', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        this.deleteStep(step)
      }).catch(() => {})
    },

    deleteStep(step) {
      workflowTemplateApi.WorkflowTemplateStepDelete(step.id, (response) => {
        if (response && response.ErrCode === 0) {
          this.editingSteps = this.editingSteps.filter(s => s.id !== step.id)
          // 重新排序
          this.saveStepSort()
          ElMessage.success('步骤已删除')
        } else {
          ElMessage.error(response.ErrMsg || '删除失败')
        }
      })
    },

    // ===== 拖拽排序 =====
    onStepDragEnd() {
      // 拖拽结束后，更新每个步骤的 sort_order
      const stepIds = this.editingSteps.map(s => s.id).filter(id => id > 0)
      if (stepIds.length > 0) {
        this.saveStepSort(stepIds)
      }
    },

    saveStepSort(stepIds) {
      if (!stepIds) {
        stepIds = this.editingSteps.map(s => s.id).filter(id => id > 0)
      }
      if (stepIds.length === 0) return
      workflowTemplateApi.WorkflowTemplateStepSort(this.selectedTemplateId, stepIds, () => {})
    },

    // ===== 步骤备注编辑 =====
    openRemarkDialog(step) {
      this.editingStep = step
      this.dialogRemarkContent = step.remark || ''
      this.remarkDialogVisible = true
    },
    saveDialogRemark() {
      if (!this.editingStep) return
      this.editingStep.remark = this.dialogRemarkContent
      this.saveStepSilent(this.editingStep)
      this.remarkDialogVisible = false
      ElMessage.success('备注已保存')
    },

    // ===== 模板信息编辑弹窗 =====
    openTemplateEditDialog(tpl) {
      this.editingTemplate = tpl
      this.dialogTemplateName = tpl.name || ''
      this.dialogTemplateDesc = tpl.description || ''
      this.templateEditDialogVisible = true
    },
    saveTemplateFromDialog() {
      if (!this.dialogTemplateName || !this.dialogTemplateName.trim()) {
        ElMessage.warning('请填写模板名称')
        return
      }
      if (!this.editingTemplate) return
      this.savingTemplate = true
      const templateId = this.editingTemplate.id || 0
      workflowTemplateApi.WorkflowTemplateSave({
        id: templateId,
        name: this.dialogTemplateName.trim(),
        description: this.dialogTemplateDesc.trim(),
      }, (response) => {
        this.savingTemplate = false
        if (response && response.ErrCode === 0) {
          const saved = response.Data.template
          if (saved && saved.steps) {
            saved.steps.forEach(s => { s._key = 'step_' + s.id })
          }
          // 更新列表中的模板
          const idx = this.templates.findIndex(t => t.id === templateId)
          if (idx >= 0) {
            this.templates[idx] = saved
          } else {
            // 新增模板，添加到列表
            this.templates.push(saved)
          }
          this.selectedTemplateId = saved.id
          this.selectedTemplate = saved
          this.editTemplateName = saved.name
          this.editTemplateDesc = saved.description || ''
          this.editingSteps = (saved.steps || []).map(s => {
            const step = {
              ...s,
              _key: 'step_' + s.id,
            }
            step.step_documents_list = this.parseStepDocuments(s.step_documents)
            return step
          })
          this.templateEditDialogVisible = false
          ElMessage.success('模板信息已保存')
          this.$emit('templates-loaded', this.templates)
        } else {
          ElMessage.error(response.ErrMsg || '保存失败')
        }
      })
    },

    // ===== 导出模板 =====
    exportTemplate(tpl) {
      const exportData = {
        name: tpl.name || '',
        description: tpl.description || '',
        steps: (tpl.steps || []).map(s => ({
          name: s.name || '',
          step_key: s.step_key || '',
          prompt_content: s.prompt_content || '',
          step_documents: s.step_documents || '',
          remark: s.remark || '',
          is_fixed: s.is_fixed || 0,
          sort_order: s.sort_order || 0,
        })),
      }
      const jsonStr = JSON.stringify(exportData, null, 2)
      const blob = new Blob([jsonStr], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `workflow_template_${(tpl.name || 'template').replace(/[^\w\u4e00-\u9fa5]/g, '_')}.json`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
      ElMessage.success('模板已导出')
    },

    // ===== 导入模板 =====
    triggerImportTemplate() {
      if (this.$refs.importFileInput) {
        this.$refs.importFileInput.value = ''
        this.$refs.importFileInput.click()
      }
    },
    handleImportFile(event) {
      const file = event.target.files && event.target.files[0]
      if (!file) return
      if (this.importingTemplate) return
      const reader = new FileReader()
      reader.onload = (e) => {
        try {
          const data = JSON.parse(e.target.result)
          if (!data.name || typeof data.name !== 'string' || !data.name.trim()) {
            ElMessage.error('导入失败：模板名称不能为空')
            return
          }
          if (!Array.isArray(data.steps)) {
            ElMessage.error('导入失败：缺少步骤列表')
            return
          }
          this.doImportTemplate(data)
        } catch (err) {
          ElMessage.error('导入失败：JSON 格式无效')
        }
      }
      reader.readAsText(file)
    },
    doImportTemplate(data) {
      this.importingTemplate = true
      const importData = {
        name: data.name.trim(),
        description: (data.description || '').trim(),
        steps: (data.steps || []).map(s => ({
          name: (s.name || '').trim(),
          step_key: s.step_key || '',
          prompt_content: s.prompt_content || '',
          step_documents: s.step_documents || '',
          remark: s.remark || '',
          is_fixed: s.is_fixed || 0,
          sort_order: s.sort_order || 0,
        })).filter(s => s.name),
      }
      workflowTemplateApi.WorkflowTemplateImport(importData, (response) => {
        this.importingTemplate = false
        if (response && response.ErrCode === 0) {
          const saved = response.Data.template
          if (saved && saved.steps) {
            saved.steps.forEach(s => { s._key = 'step_' + s.id })
          }
          this.templates.push(saved)
          this.selectTemplate(saved)
          ElMessage.success('模板导入成功')
          this.$emit('templates-loaded', this.templates)
        } else {
          ElMessage.error(response.ErrMsg || '导入失败')
        }
      })
    },
  },
}
</script>

<style scoped>
.wf-template-manager {
  display: flex;
  gap: 16px;
  height: 100%;
  min-height: 0;
}

/* 左侧模板列表 */
.wf-template-manager__left {
  width: 260px;
  min-width: 260px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  background: var(--el-bg-color);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.wf-template-manager__section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.wf-template-manager__template-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.wf-template-manager__template-item {
  padding: 10px;
  border-radius: 6px;
  cursor: pointer;
  margin-bottom: 6px;
  transition: all 0.2s;
  border: 1px solid transparent;
  position: relative;
}

.wf-template-manager__template-item:hover {
  background: var(--el-fill-color-light);
}

.wf-template-manager__template-item--active {
  background: rgba(103, 194, 58, 0.08);
  border-color: #67c23a;
}

.wf-template-manager__template-item--default {
  /* 默认模板特殊样式 */
}

.wf-template-manager__template-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  display: flex;
  align-items: center;
  gap: 6px;
}

.wf-template-manager__template-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wf-template-manager__template-actions {
  margin-top: 6px;
  text-align: right;
  display: flex;
  gap: 0;
  justify-content: flex-end;
}

.wf-template-manager__template-actions .el-button {
  padding-left: 5px;
  padding-right: 5px;
}

/* 右侧步骤编辑区域 */
.wf-template-manager__right {
  flex: 1;
  border: 1px solid var(--el-border-color-light);
  border-radius: 8px;
  background: var(--el-bg-color);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.wf-template-manager__step-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.wf-template-manager__step-item {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  margin-bottom: 6px;
  background: var(--el-bg-color);
  transition: all 0.2s;
}

.wf-template-manager__step-item:hover {
  border-color: var(--el-border-color);
}

.wf-template-manager__step-item--fixed {
  background: var(--el-fill-color-light);
  opacity: 0.85;
}

.wf-template-manager__step-ghost {
  opacity: 0.4;
  background: rgba(103, 194, 58, 0.1);
}

.wf-template-manager__step-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 10px;
}

.wf-template-manager__step-drag-handle {
  cursor: grab;
  color: var(--el-text-color-placeholder);
  flex-shrink: 0;
}

.wf-template-manager__step-drag-handle:active {
  cursor: grabbing;
}

.wf-template-manager__step-lock {
  color: var(--el-text-color-placeholder);
  flex-shrink: 0;
}

.wf-template-manager__step-index {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
  min-width: 24px;
  text-align: right;
}

.wf-template-manager__step-name {
  flex: 1;
}

.wf-template-manager__step-actions {
  flex-shrink: 0;
  display: flex;
  gap: 4px;
}

.wf-template-manager__step-editor {
  padding: 0 10px 10px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.wf-template-manager__step-md-editor {
  margin-top: 8px;
  height: 300px;
}

.wf-template-manager__step-editor-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
}

/* 空状态 */
.wf-template-manager__empty {
  padding: 40px;
  text-align: center;
  color: var(--el-text-color-placeholder);
  font-size: 14px;
}

.wf-template-manager__empty--full {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
}

/* 标题右侧按钮组 */
.wf-template-manager__section-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 内置占位符弹窗 */
.wf-placeholder-dialog :deep(.el-dialog__body) {
  padding-top: 8px;
}

.wf-placeholder-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 60vh;
  overflow-y: auto;
  padding-right: 4px;
}

.wf-placeholder-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.wf-placeholder-item:hover {
  background: #f8faf4;
  border-color: #b8d4b0;
}

.wf-placeholder-item__content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.wf-placeholder-item__row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.wf-placeholder-item__label {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
  min-width: 120px;
  flex-shrink: 0;
}

.wf-placeholder-item__value {
  font-size: 13px;
  color: #3d6b3d;
  background: #eef4ec;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif;
}

.wf-placeholder-item__tip {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.wf-placeholder-item__copy {
  color: var(--el-text-color-placeholder);
  font-size: 14px;
  margin-top: 2px;
  flex-shrink: 0;
  transition: color 0.2s ease;
}

.wf-placeholder-item:hover .wf-placeholder-item__copy {
  color: #5a8a5a;
}

.wf-prompt-dialog__tip {
  margin-bottom: 12px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

/* 提示词弹窗占位符区域 */
.wf-prompt-dialog__placeholders {
  margin-bottom: 12px;
  padding: 10px 12px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
  line-height: 1.8;
}

.wf-prompt-dialog__placeholders-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-secondary);
  margin-right: 4px;
}

.wf-prompt-dialog__editor {
  height: calc(100vh - 360px);
  min-height: 320px;
}

.wf-prompt-dialog__placeholder-tag {
  display: inline-block;
  margin-right: 6px;
  margin-top: 4px;
  color: #3d6b3d;
  background: #eef4ec;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif;
  font-size: 12px;
  cursor: pointer;
  transition: background 0.2s ease;
}

.wf-prompt-dialog__placeholder-tag:hover {
  background: #dfebd9;
}

.wf-prompt-dialog__placeholder-tag--document {
  color: #8b5e3c;
  background: #fdf6ec;
}

.wf-prompt-dialog__placeholder-tag--document:hover {
  background: #f5e6d0;
}

.wf-prompt-dialog__placeholder-tag--skill {
  color: #5b6abf;
  background: #eef0ff;
}

.wf-prompt-dialog__placeholder-tag--skill:hover {
  background: #dde1fc;
}

/* 步骤文档角标 */
.wf-template-manager__doc-badge {
  margin-left: 4px;
}

.wf-template-manager__doc-badge :deep(.el-badge__content) {
  height: 14px;
  line-height: 14px;
  padding: 0 4px;
  font-size: 10px;
}

/* 步骤文档配置弹窗 */
.wf-document-dialog :deep(.el-dialog__body) {
  padding-top: 8px;
  max-height: 60vh;
  overflow-y: auto;
}

.wf-document-dialog__tip {
  margin-bottom: 12px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
  line-height: 1.6;
}

.wf-document-item {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  margin-bottom: 12px;
  overflow: hidden;
}

.wf-document-item__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--el-fill-color-light);
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.wf-document-item__header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.wf-document-item__api-doc-switch {
  display: flex;
  align-items: center;
  gap: 6px;
}

.wf-document-item__api-doc-label {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.wf-document-item__index {
  font-size: 13px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.wf-document-item__body {
  padding: 12px;
}

.wf-document-item__input {
  margin-bottom: 10px;
}

.wf-document-item__placeholder {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 10px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-size: 13px;
}

.wf-document-item__placeholder-label {
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}

.wf-document-item__placeholder code {
  color: #3d6b3d;
  background: #eef4ec;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif;
}

.wf-document-item__placeholder-input {
  flex: 1;
}

.wf-document-item__preview {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  padding: 8px 10px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.wf-document-item__preview code {
  color: #3d6b3d;
  background: #eef4ec;
  padding: 2px 8px;
  border-radius: 4px;
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif;
}

.wf-document-dialog__actions {
  display: flex;
  justify-content: center;
  margin-top: 4px;
}

/* 步骤备注弹窗 */
.wf-remark-dialog :deep(.el-dialog__body) {
  padding-top: 12px;
}

/* 模板信息编辑弹窗 */
.wf-template-edit-dialog :deep(.el-dialog__body) {
  padding-top: 12px;
}
</style>
