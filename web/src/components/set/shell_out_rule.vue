<template>
  <div class="rule-set-page">
    <div class="rule-set-header">
      <div>
        <h3 class="rule-set-title">终端规则</h3>
        <p class="rule-set-desc">先定义一套规则，再让终端输出任务选择它。过滤日志用于隐藏噪音，错误告警用于抓取异常信息。</p>
      </div>
      <div class="rule-set-header__actions">
        <pl-button type="primary" @click="openRuleSetDialog()">新增规则集</pl-button>
      </div>
    </div>

    <div class="rule-set-tips">
      <div class="rule-set-tips__title">怎么理解这几个概念</div>
      <div class="rule-set-tips__item">过滤日志：命中后这条日志不再显示，适合心跳、定时打印、无意义轮询。</div>
      <div class="rule-set-tips__item">错误告警：命中后进入告警列表，适合报错、异常、超时、中断等重要日志。</div>
      <div class="rule-set-tips__item">命中后停止后续匹配：这条日志命中当前规则后，不再继续套用下面的规则。</div>
    </div>

    <el-table :data="ruleSetList" border style="width: 100%">
      <el-table-column prop="id" label="#ID" width="80" />
      <el-table-column prop="name" label="名称" min-width="180" />
      <el-table-column prop="description" label="说明" min-width="220" show-overflow-tooltip />
      <el-table-column label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.is_enabled ? 'success' : 'info'" effect="light">
            {{ scope.row.is_enabled ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="rule_item_count" label="规则数" width="100" />
      <el-table-column label="操作" width="220">
        <template #default="scope">
          <pl-button type="primary" link @click="openRuleSetDialog(scope.row)">编辑</pl-button>
          <pl-button type="danger" link @click="removeRuleSet(scope.row)">删除</pl-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="ruleSetDialogVisible" :title="ruleSetForm.id ? '编辑规则集' : '新增规则集'" width="80%" destroy-on-close>
      <el-form label-width="100px" class="rule-set-form">
        <el-form-item label="名称">
          <el-input v-model="ruleSetForm.name" placeholder="例如：PHP 服务日志规则" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="ruleSetForm.description" type="textarea" :rows="2" placeholder="说明这套规则适用于什么场景" />
        </el-form-item>
        <el-form-item label="启用">
          <el-switch v-model="ruleSetForm.is_enabled" />
        </el-form-item>
        <el-form-item>
          <pl-button type="primary" @click="saveRuleSet">保存</pl-button>
          <pl-button @click="ruleSetDialogVisible = false" style="margin-left: 10px;">取消</pl-button>
        </el-form-item>
      </el-form>

      <div class="rule-item-toolbar">
        <div class="rule-item-toolbar__title">规则列表</div>
        <pl-button @click="openRuleItemDialog()">新增规则</pl-button>
      </div>

      <div v-if="ruleSetForm.rule_items.length === 0" class="rule-empty-state">
        当前还没有规则，可以先新增一条“过滤日志”或“错误告警”。
      </div>

      <el-table v-else :data="ruleSetForm.rule_items" border class="rule-item-table" max-height="400">
        <el-table-column label="#" width="70">
          <template #default="scope">
            {{ scope.$index + 1 }}
          </template>
        </el-table-column>
        <el-table-column prop="name" label="规则名称" min-width="180" show-overflow-tooltip />
        <el-table-column label="规则用途" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.rule_type === 'alert' ? 'danger' : 'info'" effect="light">
              {{ getRuleTypeLabel(scope.row.rule_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="匹配方式" width="120">
          <template #default="scope">
            {{ getMatchTypeLabel(scope.row.match_type) }}
          </template>
        </el-table-column>
        <el-table-column prop="pattern" label="要匹配的内容" min-width="220" show-overflow-tooltip />
        <el-table-column label="状态" width="90">
          <template #default="scope">
            <el-switch
              v-model="scope.row.is_enabled"
              @change="toggleRuleItemEnabled(scope.row, scope.$index)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="240">
          <template #default="scope">
            <pl-button type="primary" link @click="openRuleItemDialog(scope.row, scope.$index)">编辑</pl-button>
            <pl-button type="success" link @click="copyRuleItem(scope.row)">复制新增</pl-button>
            <pl-button type="danger" link @click="removeRuleItem(scope.$index)">删除</pl-button>
          </template>
        </el-table-column>
      </el-table>


    </el-dialog>

    <el-dialog
      v-model="ruleItemDialogVisible"
      :title="ruleItemFormIndex >= 0 ? '编辑规则' : '新增规则'"
      width="680px"
      destroy-on-close
    >
      <div class="rule-item-grid">
        <div class="rule-item-field">
          <div class="rule-item-field__label">规则名称</div>
          <el-input v-model="ruleItemForm.name" placeholder="例如：过滤心跳日志" />
        </div>

        <div class="rule-item-field">
          <div class="rule-item-field__label">规则用途</div>
          <el-select v-model="ruleItemForm.rule_type" style="width: 100%">
            <el-option label="过滤日志" value="drop" />
            <el-option label="错误告警" value="alert" />
          </el-select>
          <div class="rule-item-field__tip">{{ getRuleTypeHelp(ruleItemForm.rule_type) }}</div>
        </div>

        <div class="rule-item-field">
          <div class="rule-item-field__label">匹配方式</div>
          <el-select v-model="ruleItemForm.match_type" style="width: 100%">
            <el-option label="包含文字" value="contains" />
            <el-option label="正则匹配" value="regex" />
          </el-select>
          <div class="rule-item-field__tip">{{ getMatchTypeHelp(ruleItemForm.match_type) }}</div>
        </div>

        <div class="rule-item-field">
          <div class="rule-item-field__label">优先级</div>
          <el-input-number v-model="ruleItemForm.priority" :min="0" :step="1" style="width: 100%" />
          <div class="rule-item-field__tip">数字越小越先执行，建议从 0 开始排。</div>
        </div>

        <div class="rule-item-field rule-item-field--full">
          <div class="rule-item-field__label">要匹配的内容</div>
          <el-input v-model="ruleItemForm.pattern" :placeholder="getPatternPlaceholder(ruleItemForm.match_type)" />
          <div class="rule-item-field__tip">只要日志满足这里的条件，就算命中这条规则。</div>
        </div>

        <div class="rule-item-field rule-item-field--full">
          <div class="rule-item-field__label">命中但忽略的内容</div>
          <el-input v-model="ruleItemForm.exclude_pattern" placeholder="可不填，例如：忽略测试环境的同类日志" />
          <div class="rule-item-field__tip">如果日志同时命中了这里的内容，就不会触发当前规则。</div>
        </div>

        <div class="rule-item-field">
          <div class="rule-item-field__label">启用规则</div>
          <el-switch v-model="ruleItemForm.is_enabled" />
        </div>

        <div class="rule-item-field rule-item-field--switch">
          <div class="rule-item-field__label">命中后停止后续匹配</div>
          <el-switch v-model="ruleItemForm.stop_on_match" />
          <div class="rule-item-field__tip">适合优先级较高的规则，避免同一条日志被后续规则重复处理。</div>
        </div>

        <template v-if="ruleItemForm.rule_type === 'alert'">
          <div class="rule-item-field">
            <div class="rule-item-field__label">告警级别</div>
            <el-select v-model="ruleItemForm.alert_level" style="width: 100%">
              <el-option label="一般" value="info" />
              <el-option label="警告" value="warning" />
              <el-option label="严重" value="error" />
            </el-select>
          </div>

          <div class="rule-item-field">
            <div class="rule-item-field__label">告警分类</div>
            <el-input v-model="ruleItemForm.alert_category" placeholder="可不填，例如：数据库 / 超时 / panic" />
          </div>
        </template>
      </div>

      <template #footer>
        <pl-button @click="ruleItemDialogVisible = false">取消</pl-button>
        <pl-button type="primary" @click="confirmRuleItem">确定</pl-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import shellOutRule from '@/utils/base/shell_out_rule'

function createRuleItem() {
  return {
    id: 0,
    name: '',
    rule_type: 'drop',
    match_type: 'contains',
    pattern: '',
    exclude_pattern: '',
    priority: 0,
    is_enabled: true,
    stop_on_match: true,
    alert_level: 'warning',
    alert_category: '',
  }
}

function createRuleSetForm() {
  return {
    id: 0,
    name: '',
    description: '',
    is_enabled: true,
    match_mode: 'line',
    rule_items: [],
  }
}

function parseRuleConfig(configJSON) {
  const defaultConfig = {
    alert_level: 'warning',
    alert_category: '',
  }
  if (!configJSON || String(configJSON).trim() === '') {
    return defaultConfig
  }
  try {
    const parsed = JSON.parse(configJSON)
    return {
      alert_level: parsed.level || 'warning',
      alert_category: parsed.category || '',
    }
  } catch (error) {
    return defaultConfig
  }
}

function buildRuleConfig(item) {
  if (item.rule_type !== 'alert') {
    return '{}'
  }
  return JSON.stringify({
    level: item.alert_level || 'warning',
    category: item.alert_category || '',
  })
}

export default {
  name: 'ShellOutRuleSet',
  data() {
    return {
      ruleSetList: [],
      ruleSetDialogVisible: false,
      ruleItemDialogVisible: false,
      ruleSetForm: createRuleSetForm(),
      ruleItemForm: createRuleItem(),
      ruleItemFormIndex: -1,
    }
  },
  mounted() {
    this.loadRuleSetList()
  },
  methods: {
    // loadRuleSetList 刷新规则集列表，供设置页展示与选择器复用。 // Refresh the rule-set list so settings and shell-out dialogs stay in sync.
    loadRuleSetList() {
      shellOutRule.ShellOutRuleSetList({}, (response) => {
        if (response.ErrCode !== 0) {
          this.$helperNotify.error(response.ErrMsg || '规则集加载失败')
          return
        }
        this.ruleSetList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    getRuleTypeLabel(ruleType) {
      return ruleType === 'alert' ? '错误告警' : '过滤日志'
    },
    getRuleTypeHelp(ruleType) {
      if (ruleType === 'alert') {
        return '命中后会进入告警列表，适合错误、超时、中断等异常日志。'
      }
      return '命中后这条日志不会显示在输出区，适合过滤心跳、轮询、定时打印。'
    },
    getMatchTypeLabel(matchType) {
      return matchType === 'regex' ? '正则匹配' : '包含文字'
    },
    getMatchTypeHelp(matchType) {
      if (matchType === 'regex') {
        return '适合复杂匹配，例如端口、编号、错误模式提取。'
      }
      return '只要日志里包含这段文字，就会命中。'
    },
    getPatternPlaceholder(matchType) {
      if (matchType === 'regex') {
        return '例如：timeout|connection refused'
      }
      return '例如：heartbeat ok'
    },
    normalizeRuleItem(item) {
      const config = parseRuleConfig(item.config_json || '{}')
      return {
        id: item.id || 0,
        name: item.name || '',
        rule_type: item.rule_type || 'drop',
        match_type: item.match_type || 'contains',
        pattern: item.pattern || '',
        exclude_pattern: item.exclude_pattern || '',
        priority: Number(item.priority || 0),
        is_enabled: Number(item.is_enabled) === 1 || item.is_enabled === true,
        stop_on_match: Number(item.stop_on_match) === 1 || item.stop_on_match === true,
        alert_level: config.alert_level || 'warning',
        alert_category: config.alert_category || '',
      }
    },
    openRuleSetDialog(row) {
      if (!row || !row.id) {
        this.ruleSetForm = createRuleSetForm()
        this.ruleSetDialogVisible = true
        return
      }
      shellOutRule.ShellOutRuleSetInfo({ id: row.id }, (response) => {
        if (response.ErrCode !== 0) {
          this.$helperNotify.error(response.ErrMsg || '规则集详情加载失败')
          return
        }
        const info = response.Data?.rule_set || {}
        const items = Array.isArray(response.Data?.rule_items) ? response.Data.rule_items : []
        this.ruleSetForm = {
          id: info.id || 0,
          name: info.name || '',
          description: info.description || '',
          is_enabled: Number(info.is_enabled) === 1,
          match_mode: info.match_mode || 'line',
          rule_items: items.map((item) => this.normalizeRuleItem(item)),
        }
        this.ruleSetDialogVisible = true
      })
    },
    openRuleItemDialog(row, index = -1) {
      this.ruleItemForm = row ? {...row} : createRuleItem()
      this.ruleItemFormIndex = index
      this.ruleItemDialogVisible = true
    },
    confirmRuleItem() {
      if (!String(this.ruleItemForm.name || '').trim()) {
        this.$helperNotify.error('规则名称不能为空')
        return
      }
      if (!String(this.ruleItemForm.pattern || '').trim()) {
        this.$helperNotify.error('要匹配的内容不能为空')
        return
      }
      const nextRuleItem = {...this.ruleItemForm}
      if (this.ruleItemFormIndex >= 0) {
        this.ruleSetForm.rule_items.splice(this.ruleItemFormIndex, 1, nextRuleItem)
      } else {
        this.ruleSetForm.rule_items.push(nextRuleItem)
      }
      this.ruleItemDialogVisible = false
      // 编辑规则确定后直接保存
      this.saveRuleSetWithItems()
    },
    removeRuleItem(index) {
      this.ruleSetForm.rule_items.splice(index, 1)
    },
    copyRuleItem(row) {
      const copiedItem = { ...row, id: 0, name: row.name + '_副本' }
      this.ruleItemForm = copiedItem
      this.ruleItemFormIndex = -1
      this.ruleItemDialogVisible = true
    },
    validateRuleItems() {
      for (let index = 0; index < this.ruleSetForm.rule_items.length; index += 1) {
        const item = this.ruleSetForm.rule_items[index]
        if (!String(item.name || '').trim()) {
          this.$helperNotify.error(`第 ${index + 1} 条规则缺少规则名称`)
          return false
        }
        if (!String(item.pattern || '').trim()) {
          this.$helperNotify.error(`第 ${index + 1} 条规则缺少要匹配的内容`)
          return false
        }
      }
      return true
    },
    // saveRuleSet 只保存规则集基础信息（名称、说明、启用状态），不保存规则列表 // Save only the rule set basic info (name, description, is_enabled), not the rule items
    saveRuleSet() {
      if (!String(this.ruleSetForm.name || '').trim()) {
        this.$helperNotify.error('规则集名称不能为空')
        return
      }
      const payload = {
        id: this.ruleSetForm.id,
        name: this.ruleSetForm.name,
        description: this.ruleSetForm.description,
        is_enabled: this.ruleSetForm.is_enabled ? 1 : 0,
        match_mode: this.ruleSetForm.match_mode,
        rule_items: [],
      }
      shellOutRule.ShellOutRuleSetSave(payload, (response) => {
        if (response.ErrCode !== 0) {
          this.$helperNotify.error(response.ErrMsg || '规则集保存失败')
          return
        }
        this.ruleSetDialogVisible = false
        this.loadRuleSetList()
        this.$helperNotify.success('规则集已保存')
      })
    },
    // saveRuleSetWithItems 保存规则集及所有规则项 // Save rule set with all rule items
    saveRuleSetWithItems() {
      if (!String(this.ruleSetForm.name || '').trim()) {
        this.$helperNotify.error('规则集名称不能为空')
        return
      }
      if (!this.validateRuleItems()) {
        return
      }
      const payload = {
        id: this.ruleSetForm.id,
        name: this.ruleSetForm.name,
        description: this.ruleSetForm.description,
        is_enabled: this.ruleSetForm.is_enabled ? 1 : 0,
        match_mode: this.ruleSetForm.match_mode,
        rule_items: this.ruleSetForm.rule_items.map((item) => ({
          id: item.id || 0,
          name: item.name,
          rule_type: item.rule_type,
          match_type: item.match_type,
          pattern: item.pattern,
          exclude_pattern: item.exclude_pattern,
          priority: Number(item.priority || 0),
          is_enabled: item.is_enabled ? 1 : 0,
          stop_on_match: item.stop_on_match ? 1 : 0,
          config_json: buildRuleConfig(item),
        })),
      }
      shellOutRule.ShellOutRuleSetSave(payload, (response) => {
        if (response.ErrCode !== 0) {
          this.$helperNotify.error(response.ErrMsg || '规则保存失败')
          return
        }
        this.loadRuleSetList()
      })
    },
    // toggleRuleItemEnabled 切换规则启用状态并即时保存 // Toggle rule item enabled status and save immediately
    toggleRuleItemEnabled() {
      this.saveRuleSetWithItems()
    },
    removeRuleSet(row) {
      this.$confirm(`确定删除规则集"${row.name}"吗？`, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        shellOutRule.ShellOutRuleSetDelete({ id: row.id }, (response) => {
          if (response.ErrCode !== 0) {
            this.$helperNotify.error(response.ErrMsg || '规则集删除失败')
            return
          }
          this.loadRuleSetList()
          this.$helperNotify.success('规则集已删除')
        })
      }).catch(() => {})
    },
    importLegacyRules() {
      this.$confirm('这会把旧分组里的过滤正则、错误捕获正则和排除条件导入到新规则中心，并自动绑定到对应分组下的终端输出任务。是否继续？', '导入旧规则', {
        confirmButtonText: '开始导入',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        shellOutRule.ShellOutRuleImportLegacy({}, (response) => {
          if (response.ErrCode !== 0) {
            this.$helperNotify.error(response.ErrMsg || '旧规则导入失败')
            return
          }
          const data = response.Data || {}
          this.loadRuleSetList()
          this.$alert(
            `共扫描 ${data.group_count || 0} 个分组，导入 ${data.imported_rule_set_count || 0} 个规则集、${data.imported_rule_item_count || 0} 条规则，绑定 ${data.bound_shell_out_count || 0} 个终端输出任务。`,
            '导入完成',
            {
              confirmButtonText: '知道了',
            }
          )
        })
      }).catch(() => {})
    },
  },
}
</script>

<style scoped>
.rule-set-page {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.rule-set-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.rule-set-header__actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.rule-set-title {
  margin: 0;
  color: #324132;
}

.rule-set-desc {
  margin: 6px 0 0;
  color: #6f7f6f;
  line-height: 1.6;
}

.rule-set-tips {
  padding: 14px 16px;
  border-radius: 12px;
  border: 1px solid #d9e6d8;
  background: linear-gradient(135deg, #f8fcf7 0%, #edf7ec 100%);
}

.rule-set-tips__title {
  font-weight: 600;
  color: #365137;
  margin-bottom: 8px;
}

.rule-set-tips__item {
  color: #566d57;
  line-height: 1.7;
}

.rule-item-toolbar {
  margin: 14px 0 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.rule-item-toolbar__title {
  font-size: 15px;
  font-weight: 600;
  color: #3d4d3d;
}

.rule-set-form {
  margin-bottom: 8px;
}

.rule-empty-state {
  padding: 18px;
  border: 1px dashed #c8d7c7;
  border-radius: 12px;
  text-align: center;
  color: #687b68;
  background: #fafcf9;
}

.rule-item-table {
  margin-top: 10px;
}

.rule-item-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px 16px;
}

.rule-item-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.rule-item-field--full {
  grid-column: 1 / -1;
}

.rule-item-field--switch {
  justify-content: flex-start;
}

.rule-item-field__label {
  font-size: 13px;
  font-weight: 600;
  color: #405240;
}

.rule-item-field__tip {
  font-size: 12px;
  line-height: 1.6;
  color: #7a8b79;
}

@media (max-width: 900px) {
  .rule-set-header {
    flex-direction: column;
  }

  .rule-item-grid {
    grid-template-columns: 1fr;
  }
}
</style>
