<template>
  <el-form :model="localItem" label-width="180px" class="process-item-editor">
    <el-form-item label="名称" :error="fieldError('name')">
      <el-input v-model="localItem.name" />
    </el-form-item>

    <el-form-item label="类型" :error="fieldError('type')">
      <el-select v-model="localItem.type" placeholder="请选择类型" style="width: 100%">
        <el-option
          v-for="option in processTypeOptions"
          :key="option.value"
          :label="option.label"
          :value="option.value"
        />
      </el-select>
    </el-form-item>

    <el-form-item label="前端执行提示">
      <el-input v-model="localItem.tip" placeholder="可选，展示给执行中的用户提示" />
    </el-form-item>

    <template v-if="showField('locator')">
      <el-form-item :label="fieldLabel('locator')" :error="fieldError('locator')">
        <div class="list-editor">
          <template v-if="useLocatorExpressionEditor">
            <div class="locator-expression-toolbar">
              <el-select v-model="formMeta.locator_joiner" class="locator-expression-toolbar__select">
                <el-option
                  v-for="option in locatorJoinerOptions"
                  :key="option.value"
                  :label="option.label"
                  :value="option.value"
                />
                <el-option label="高级表达式" value="raw" />
              </el-select>
              <div class="locator-expression-toolbar__tip">
                先选匹配方式，再逐条填写定位条件。“要求找到”表示页面上必须出现该元素；“要求找不到”表示该元素不能出现；最后一项用来决定遇到多个同类元素时，是检查全部，还是只按第一个来判断。
              </div>
            </div>

            <div v-if="showTextContentLocatorSummary" class="locator-purpose-card">
              <div class="locator-purpose-card__title">提取规则说明</div>
              <div class="locator-purpose-card__text">
                这个类型最终会按整条定位表达式去查找元素并读取文本内容，并没有单独的“提取目标字段”。如果只是想稳定提取一个元素的文本，建议优先使用“只找一个元素”并只保留 1 条定位。
              </div>
            </div>

            <template v-if="useRawLocatorTextarea">
              <el-input
                v-model="formMeta.locator_raw"
                type="textarea"
                :rows="3"
                placeholder="例如 .username&&!.btn.login_as_reg_btn 或 .dialog|first"
              />
            </template>

            <template v-else>
              <div
                v-for="(item, index) in formMeta.locator_list"
                :key="item.uid"
                class="locator-expression-row"
              >
                <el-input v-model="item.value" :placeholder="fieldPlaceholder('locator')" />
                <el-select
                  v-if="showLocatorExistMode"
                  v-model="item.exist_mode"
                  class="locator-expression-row__mode"
                >
                  <el-option label="要求找到" value="exist" />
                  <el-option label="要求找不到" value="not_exist" />
                </el-select>
                <el-select v-model="item.match_mode" class="locator-expression-row__mode">
                  <el-option label="多个时按默认方式处理" value="all" />
                  <el-option label="多个时只取第一个" value="first" />
                </el-select>
                <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeLocatorRow(index)">
                  删除
                </GitActionButton>
              </div>
              <GitActionButton
                v-if="formMeta.locator_joiner !== 'single'"
                compact
                size="small"
                native-type="button"
                @click="addLocatorRow"
              >
                新增定位条件
              </GitActionButton>
            </template>
          </template>

          <template v-else>
            <div v-for="(item, index) in formMeta.locator_list" :key="item.uid" class="list-editor__row">
              <el-input v-model="item.value" :placeholder="fieldPlaceholder('locator')" />
              <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeLocatorRow(index)">
                删除
              </GitActionButton>
            </div>
            <GitActionButton compact size="small" native-type="button" @click="addLocatorRow">
              新增定位条件
            </GitActionButton>
          </template>

          <div class="field-guide">{{ fieldGuide('locator') }}</div>
          <div v-if="locatorBehaviorSummary" class="locator-behavior-summary">
            <div class="locator-behavior-summary__title">当前查找方式</div>
            <div class="locator-behavior-summary__text">{{ locatorBehaviorSummary }}</div>
          </div>
        </div>
      </el-form-item>
    </template>

    <template v-if="showField('secondary_locator')">
      <el-form-item :label="fieldLabel('secondary_locator')" :error="fieldError('secondary_locator')">
        <el-input v-model="formMeta.secondary_locator" :placeholder="fieldPlaceholder('secondary_locator')" />
        <div class="field-guide">{{ fieldGuide('secondary_locator') }}</div>
      </el-form-item>
    </template>

    <template v-if="showField('tertiary_locator')">
      <el-form-item :label="fieldLabel('tertiary_locator')" :error="fieldError('tertiary_locator')">
        <el-input v-model="formMeta.tertiary_locator" :placeholder="fieldPlaceholder('tertiary_locator')" />
        <div class="field-guide">{{ fieldGuide('tertiary_locator') }}</div>
      </el-form-item>
    </template>

    <template v-if="localItem.type === 'redirect_uri'">
      <el-form-item label="跳转地址" :error="fieldError('value')">
        <el-input v-model="formMeta.value" placeholder="例如 /login 或 https://example.com/login" />
        <div class="field-guide">{{ fieldGuide('value') }}</div>
      </el-form-item>

      <el-form-item label="跳转后等待地址" :error="fieldError('register_response_urls')">
        <div class="list-editor">
          <div v-for="(item, index) in formMeta.register_response_urls" :key="item.uid" class="response-url-row">
            <el-input v-model="item.url" placeholder="等待地址，例如 /home 或 https://example.com/home" />
            <el-input-number v-model="item.wait_second" :min="1" :controls="false" class="plain-number-input" />
            <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeRegisterResponseUrl(index)">
              删除
            </GitActionButton>
          </div>
          <GitActionButton compact size="small" native-type="button" @click="addRegisterResponseUrl">
            新增等待地址
          </GitActionButton>
          <div class="field-guide">{{ fieldGuide('register_response_urls') }}</div>
        </div>
      </el-form-item>
    </template>

    <template v-else-if="showField('value')">
      <el-form-item :label="fieldLabel('value')" :error="fieldError('value')">
        <el-input
          v-model="formMeta.value"
          :placeholder="fieldPlaceholder('value')"
          type="textarea"
          :rows="textareaRows('value')"
        />
        <div class="field-guide">{{ fieldGuide('value') }}</div>
      </el-form-item>
    </template>

    <template v-if="localItem.type === 'wait_url'">
      <el-form-item label="等待地址" :error="fieldError('response_url')">
        <el-input v-model="formMeta.response_url" :placeholder="fieldPlaceholder('response_url')" />
        <div class="field-guide">{{ fieldGuide('response_url') }}</div>
      </el-form-item>

      <el-form-item label="等待秒数" :error="fieldError('wait_second')">
        <el-input-number v-model="formMeta.wait_second" :min="1" :controls="false" class="plain-number-input" />
      </el-form-item>
    </template>

    <template v-else>
      <template v-if="showField('wait_second')">
        <el-form-item :label="fieldLabel('wait_second')" :error="fieldError('wait_second')">
          <el-input-number v-model="formMeta.wait_second" :min="1" :controls="false" class="plain-number-input" />
        </el-form-item>
      </template>

      <template v-if="showField('response_url')">
        <el-form-item :label="fieldLabel('response_url')" :error="fieldError('response_url')">
          <el-input v-model="formMeta.response_url" :placeholder="fieldPlaceholder('response_url')" />
          <div class="field-guide">{{ fieldGuide('response_url') }}</div>
        </el-form-item>
      </template>

      <template v-if="localItem.type !== 'redirect_uri' && showField('register_response_urls')">
        <el-form-item :label="fieldLabel('register_response_urls')" :error="fieldError('register_response_urls')">
          <div class="list-editor">
            <div v-for="(item, index) in formMeta.register_response_urls" :key="item.uid" class="response-url-row">
              <el-input v-model="item.url" placeholder="等待地址" />
              <el-input-number v-model="item.wait_second" :min="1" :controls="false" class="plain-number-input" />
              <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeRegisterResponseUrl(index)">
                删除
              </GitActionButton>
            </div>
            <GitActionButton compact size="small" native-type="button" @click="addRegisterResponseUrl">
              新增等待地址
            </GitActionButton>
            <div class="field-guide">{{ fieldGuide('register_response_urls') }}</div>
          </div>
        </el-form-item>
      </template>
    </template>

    <template v-if="showField('out_key')">
      <el-form-item :label="fieldLabel('out_key')" :error="fieldError('out_key')">
        <el-input v-model="formMeta.out_key" placeholder="例如 {login_state}" />
        <div class="field-guide">{{ fieldGuide('out_key') }}</div>
      </el-form-item>
    </template>

    <template v-if="showField('check_key')">
      <el-form-item :label="fieldLabel('check_key')" :error="fieldError('check_key')">
        <div class="list-editor">
          <el-select v-model="formMeta.check_mode" class="check-mode-select">
            <el-option label="可以执行" value="none" />
            <el-option label="按前面结果判断" value="bool" />
            <el-option label="按内容比较判断" value="compare" />
          </el-select>

          <template v-if="formMeta.check_mode === 'bool'">
            <div v-for="(item, index) in formMeta.check_rule_list" :key="item.uid" class="check-rule-row">
              <el-select v-model="item.key" filterable placeholder="请选择前面节点的输出">
                <el-option
                  v-for="option in checkKeyOptions"
                  :key="`${option.value}-${option.label}`"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <el-select v-model="item.expect" class="check-rule-row__mode">
                <el-option label="必须为真" value="true" />
                <el-option label="必须为假" value="false" />
              </el-select>
              <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeCheckRule(index)">
                删除
              </GitActionButton>
            </div>
            <GitActionButton
              compact
              size="small"
              native-type="button"
              :disabled="checkKeyOptions.length === 0 || usedCheckKeyCount >= checkKeyOptions.length"
              @click="addCheckRule"
            >
              新增判断条件
            </GitActionButton>
          </template>

          <template v-if="formMeta.check_mode === 'compare'">
            <div class="compare-rule-row">
              <el-select v-model="formMeta.compare_rule.left" filterable placeholder="请选择左侧输出">
                <el-option
                  v-for="option in checkKeyOptions"
                  :key="`left-${option.value}-${option.label}`"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <el-select v-model="formMeta.compare_rule.operator" class="compare-rule-row__operator">
                <el-option label="等于" value="==" />
                <el-option label="不等于" value="!=" />
              </el-select>
              <div class="compare-rule-right">
                <el-input
                  v-model="formMeta.compare_rule.right"
                  placeholder="请输入固定字符串，或点击下方快捷填入注入值"
                />
                <div class="compare-rule-right__actions">
                  <GitActionButton
                    v-for="option in compareRightOptions"
                    :key="`right-${option.value}`"
                    compact
                    size="small"
                    native-type="button"
                    @click="applyCompareRightQuickPick(option.value)"
                  >
                    {{ option.label }}
                  </GitActionButton>
                </div>
              </div>
            </div>
          </template>
        </div>
        <div class="field-guide">{{ fieldGuide('check_key') }}</div>
      </el-form-item>
    </template>

    <template v-if="showField('wait_count')">
      <el-form-item :label="fieldLabel('wait_count')" :error="fieldError('wait_count')">
        <el-input-number v-model="formMeta.wait_count" :min="1" :controls="false" class="plain-number-input" />
      </el-form-item>
    </template>

    <template v-if="showField('delete_mode')">
      <el-form-item :label="fieldLabel('delete_mode')">
        <el-select v-model="formMeta.delete_mode" style="width: 100%">
          <el-option label="按 class 删除" value="class" />
        </el-select>
      </el-form-item>
    </template>

    <template v-if="showField('bool_result_rules')">
      <el-form-item label="主元素定位规则" :error="fieldError('bool_result_rules')">
        <div class="list-editor">
          <div v-for="(item, index) in formMeta.bool_result_rules" :key="item.uid" class="bool-result-row">
            <el-input v-model="item.locator" placeholder="例如 .user-info.ant-dropdown-trigger" />
            <el-select v-model="item.return" style="width: 140px">
              <el-option label="存在返回 true" :value="true" />
              <el-option label="存在返回 false" :value="false" />
            </el-select>
            <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeBoolResultRule(index)">
              删除
            </GitActionButton>
          </div>
          <GitActionButton compact size="small" native-type="button" @click="addBoolResultRule">
            新增定位
          </GitActionButton>
          <div class="field-guide">{{ fieldGuide('bool_result_rules') }}</div>
        </div>
      </el-form-item>
    </template>

    <el-form-item label="权重">
      <el-input-number v-model="localItem.weight" :min="0" />
    </el-form-item>

    <el-form-item label="等待时长(ms)" :error="fieldError('wait_mills')">
      <el-input-number v-model="localItem.wait_mills" :min="0" />
    </el-form-item>

    <el-form-item label="域名限制" :error="fieldError('domain_limit')">
      <el-input v-model="localItem.domain_limit" placeholder="可选，例如 example.com" />
      <div class="field-guide">{{ fieldGuide('domain_limit') }}</div>
    </el-form-item>

    <el-form-item v-if="allowAppendToReplace" label="输出追加到替换列表">
      <el-select v-model="localItem.append_to_replace" style="width: 100%">
        <el-option label="追加" value="1" />
        <el-option label="不追加" value="0" />
      </el-select>
    </el-form-item>

    <el-form-item label="执行方式">
      <el-select v-model="localItem.is_async" style="width: 100%">
        <el-option label="同步" value="0" />
        <el-option label="异步" value="1" />
      </el-select>
    </el-form-item>

    <el-form-item label="出错后是否继续">
      <el-select v-model="localItem.is_error_continue" style="width: 100%">
        <el-option label="中断" value="0" />
        <el-option label="继续" value="1" />
      </el-select>
    </el-form-item>
  </el-form>
</template>

<script>
import GitActionButton from '@/components/base/GitActionButton.vue'

const {
  PROCESS_ITEM_FIELD_GUIDES,
  PROCESS_TYPE_FIELDS,
  parseCheckConfig,
  parseRedirectUriValue,
  parseWaitUrlValue,
  serializeCheckConfig,
  serializeRedirectUriValue,
  serializeWaitUrlValue,
  validateProcessItemForm,
} = require('../../utils/smart_link_process_validation.cjs')

const createDefaultItem = () => ({
  id: 0,
  name: '',
  smart_link_process_id: 0,
  type: '',
  locator: '',
  wait_mills: 3000,
  tip: '',
  value: '',
  out_key: '',
  check_key: '',
  weight: 0,
  domain_limit: '',
  append_to_replace: '0',
  is_async: '0',
  is_error_continue: '0',
  next_ids: '',
  x: 0,
  y: 0,
})

const PROCESS_TYPE_OPTIONS = [
  { label: '提取元素内容 text_content', value: 'text_content' },
  { label: '跳转 redirect_uri', value: 'redirect_uri' },
  { label: '等待接口完成 wait_url', value: 'wait_url' },
  { label: '等待毫秒 wait', value: 'wait' },
  { label: '判断输出 bool_result', value: 'bool_result' },
  { label: '判断存在 bool_exist', value: 'bool_exist' },
  { label: '点击元素 click', value: 'click' },
  { label: '输入信息 input', value: 'input' },
  { label: '关闭页面 close', value: 'close' },
  { label: '不存在时等待 no_exist_wait', value: 'no_exist_wait' },
  { label: '提取 canvas 图片 canvas_image', value: 'canvas_image' },
  { label: '输入账号密码 login_username_password', value: 'login_username_password' },
  { label: '删除元素 delete_element', value: 'delete_element' },
]

function safeParseJson(text, fallback) {
  if (!text) return fallback
  try {
    return JSON.parse(text)
  } catch (error) {
    return fallback
  }
}

function createRegisterUrl() {
  return {
    uid: `response-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    url: '',
    wait_second: 10,
  }
}

function createLocatorRow() {
  return {
    uid: `locator-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    value: '',
    exist_mode: 'exist',
    match_mode: 'all',
  }
}

function createBoolResultRule() {
  return {
    uid: `bool-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    locator: '',
    return: true,
  }
}

function createCheckRule() {
  return {
    uid: `check-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    key: '',
    expect: 'true',
  }
}

function createCompareRule() {
  return {
    left: '',
    operator: '==',
    right: '',
  }
}

// normalizeTokenLabel 用于统一下拉标签展示，兼容已带大括号的旧 out_key。
// normalizeTokenLabel normalizes option labels and preserves legacy wrapped out_key values.
function normalizeTokenLabel(value) {
  const normalizedValue = String(value || '').trim()
  if (!normalizedValue) return ''
  return normalizedValue.startsWith('{') && normalizedValue.endsWith('}')
    ? normalizedValue
    : `{${normalizedValue}}`
}

function withRegisterUid(list) {
  return (Array.isArray(list) ? list : []).map((item) => ({
    uid: item.uid || createRegisterUrl().uid,
    url: item.url || '',
    wait_second: Number(item.wait_second || 10),
  }))
}

function withCheckRuleUid(list) {
  return (Array.isArray(list) ? list : []).map((item) => ({
    uid: item.uid || createCheckRule().uid,
    key: item.key || '',
    expect: item.expect || 'true',
  }))
}

export default {
  name: 'ProcessItemEditor',
  components: {
    GitActionButton,
  },
  props: {
    modelValue: {
      type: Object,
      default: () => createDefaultItem(),
    },
    processItemOptions: {
      type: Array,
      default: () => [],
    },
  },
  emits: ['update:modelValue'],
  data() {
    return {
      localItem: createDefaultItem(),
      formMeta: {
        locator_list: [],
        locator_joiner: 'single',
        locator_raw: '',
        secondary_locator: '',
        tertiary_locator: '',
        value: '',
        out_key: '',
        check_key: '',
        check_mode: 'none',
        check_rule_list: [],
        compare_rule: createCompareRule(),
        wait_second: 10,
        wait_count: 3,
        response_url: '',
        delete_mode: 'class',
        register_response_urls: [],
        bool_result_rules: [],
      },
      syncingFromParent: false,
      lastSerializedSignature: '',
      fieldErrors: {},
      processTypeOptions: PROCESS_TYPE_OPTIONS,
    }
  },
  computed: {
    currentFields() {
      return PROCESS_TYPE_FIELDS[this.localItem.type] || []
    },
    allowAppendToReplace() {
      return this.localItem.type !== 'click' && this.localItem.type !== 'delete_element'
    },
    useLocatorExpressionEditor() {
      return this.showField('locator') && this.supportLocatorExpression(this.localItem.type)
    },
    locatorJoinerOptions() {
      const optionList = [{ label: '只找一个元素', value: 'single' }]
      if (this.localItem.type !== 'text_content') {
        optionList.push(
          { label: '多个条件都满足', value: 'and' },
          { label: '多个条件满足其一', value: 'or' }
        )
      } else if (this.formMeta.locator_joiner === 'and' || this.formMeta.locator_joiner === 'or') {
        optionList.push({
          label: this.formMeta.locator_joiner === 'and' ? '多个条件都满足（旧配置）' : '多个条件满足其一（旧配置）',
          value: this.formMeta.locator_joiner,
        })
      }
      return optionList
    },
    useRawLocatorTextarea() {
      return this.formMeta.locator_joiner === 'raw'
        || (this.localItem.type === 'text_content'
          && (this.formMeta.locator_joiner === 'and' || this.formMeta.locator_joiner === 'or'))
    },
    showLocatorExistMode() {
      return !(this.localItem.type === 'text_content' && this.formMeta.locator_joiner === 'single')
    },
    checkKeyOptions() {
      const processList = Array.isArray(this.processItemOptions) ? this.processItemOptions : []
      const currentIndex = processList.findIndex(item => String(item.id) === String(this.localItem.id))
      const availableList = currentIndex >= 0 ? processList.slice(0, currentIndex) : processList
      return availableList
        .filter(item => String(item.out_key || '').trim() !== '')
        .map(item => ({
          value: String(item.out_key || '').trim(),
          label: `${item.name || item.type || '未命名节点'}.${normalizeTokenLabel(item.out_key)}`,
        }))
    },
    compareRightOptions() {
      return [
        { value: '{user_name}', label: '当前账号用户名 {user_name}' },
        { value: '{password}', label: '当前账号密码 {password}' },
      ]
    },
    showTextContentLocatorSummary() {
      return this.localItem.type === 'text_content' && this.formMeta.locator_joiner !== 'raw'
    },
    locatorBehaviorSummary() {
      if (!this.showField('locator')) return ''
      if (this.useRawLocatorTextarea) {
        const rawText = String(this.formMeta.locator_raw || '').trim()
        return rawText
          ? `当前使用高级表达式查找：${rawText}。系统会按这条原始表达式直接处理。`
          : ''
      }

      const locatorList = (Array.isArray(this.formMeta.locator_list) ? this.formMeta.locator_list : [])
        .filter(item => String(item && item.value || '').trim() !== '')

      if (locatorList.length === 0) return ''

      const conditionText = locatorList
        .map((item) => {
          const locatorValue = String(item.value || '').trim()
          const existText = item.exist_mode === 'not_exist' ? '不能出现' : '必须出现'
          const matchText = item.match_mode === 'first' ? '多个同类元素时只按第一个处理' : '多个同类元素时按默认方式处理'
          return this.localItem.type === 'text_content' && this.formMeta.locator_joiner === 'single'
            ? `读取“${locatorValue}”的文本，${matchText}`
            : `“${locatorValue}”${existText}，并且${matchText}`
        })

      if (this.formMeta.locator_joiner === 'single') {
        return `当前只按 1 条定位查找：${conditionText[0]}。`
      }
      if (this.formMeta.locator_joiner === 'and') {
        return `当前要求以下条件全部同时成立后才算找到：${conditionText.join('；')}。`
      }
      if (this.formMeta.locator_joiner === 'or') {
        return `当前只要下面任意 1 条条件成立就算找到：${conditionText.join('；')}。`
      }
      return ''
    },
    usedCheckKeyCount() {
      return (Array.isArray(this.formMeta.check_rule_list) ? this.formMeta.check_rule_list : [])
        .map(item => String(item && item.key || '').trim())
        .filter(Boolean)
        .length
    },
  },
  watch: {
    modelValue: {
      deep: true,
      immediate: true,
      handler(value) {
        this.syncFromModel(value || createDefaultItem())
      },
    },
    localItem: {
      deep: true,
      handler() {
        this.emitChange()
      },
    },
    formMeta: {
      deep: true,
      handler() {
        this.emitChange()
      },
    },
    'localItem.type'(nextType, prevType) {
      if (!this.syncingFromParent && nextType !== prevType) {
        this.resetMetaForType(nextType)
      }
    },
    'formMeta.locator_joiner'(nextValue) {
      if (this.syncingFromParent) return
      if (nextValue === 'single') {
        const firstRow = Array.isArray(this.formMeta.locator_list) && this.formMeta.locator_list.length > 0
          ? this.formMeta.locator_list[0]
          : createLocatorRow()
        if (this.localItem.type === 'text_content') {
          firstRow.exist_mode = 'exist'
        }
        this.formMeta.locator_list = [firstRow]
      }
      if ((nextValue === 'and' || nextValue === 'or') && this.formMeta.locator_list.length === 0) {
        this.formMeta.locator_list = [createLocatorRow()]
      }
    },
  },
  methods: {
    createSignature(payload) {
      return JSON.stringify(payload || {})
    },
    syncFromModel(value) {
      const normalizedValue = {
        ...createDefaultItem(),
        ...JSON.parse(JSON.stringify(value || {})),
        next_ids: (value && value.next_ids) || '',
        append_to_replace: String((value && value.append_to_replace) ?? '0'),
        is_async: String((value && value.is_async) ?? '0'),
        is_error_continue: String((value && value.is_error_continue) ?? '0'),
      }
      const incomingSignature = this.createSignature(normalizedValue)
      if (!this.syncingFromParent && incomingSignature === this.lastSerializedSignature) {
        return
      }
      this.syncingFromParent = true
      this.localItem = normalizedValue
      this.formMeta = this.deserializeMeta(this.localItem)
      this.fieldErrors = {}
      this.lastSerializedSignature = this.createSignature(this.serializeItem())
      this.$nextTick(() => {
        this.syncingFromParent = false
      })
    },
    resetMetaForType(type) {
      this.formMeta = this.deserializeMeta({
        ...this.localItem,
        type,
        locator: '',
        value: '',
        out_key: '',
        check_key: '',
      })
      this.fieldErrors = {}
    },
    deserializeMeta(item) {
      // deserializeMeta 负责把后端存储结构转换成前端表单状态。
      // deserializeMeta converts backend payloads into editable form state.
      const meta = {
        locator_list: [],
        locator_joiner: 'single',
        locator_raw: '',
        secondary_locator: '',
        tertiary_locator: '',
        value: item.value || '',
        out_key: item.out_key || '',
        check_key: item.check_key || '',
        check_mode: 'none',
        check_rule_list: [],
        compare_rule: createCompareRule(),
        wait_second: 10,
        wait_count: 3,
        response_url: '',
        delete_mode: item.value || 'class',
        register_response_urls: [],
        bool_result_rules: [],
      }

      if (item.type === 'bool_result') {
        const rules = safeParseJson(item.locator, [])
        meta.bool_result_rules = Array.isArray(rules)
          ? rules.map(rule => ({
            uid: createBoolResultRule().uid,
            locator: rule.locator || '',
            return: rule.return !== false,
          }))
          : []
        if (meta.bool_result_rules.length === 0) {
          meta.bool_result_rules = [createBoolResultRule()]
        }
      } else if (item.type === 'wait_url') {
        Object.assign(meta, parseWaitUrlValue(item.value))
      } else if (item.type === 'redirect_uri') {
        const redirectMeta = parseRedirectUriValue(item.value)
        meta.value = redirectMeta.value
        meta.register_response_urls = withRegisterUid(redirectMeta.register_response_urls)
      } else if (item.type === 'no_exist_wait') {
        const [waitSecond, waitCount] = String(item.value || '').split('|')
        meta.wait_second = Number(waitSecond || 10)
        meta.wait_count = Number(waitCount || 3)
        Object.assign(meta, this.decodeLocatorExpression(item.locator))
      } else if (item.type === 'login_username_password') {
        const parts = String(item.locator || '').split('||')
        meta.locator_list = parts[0] ? [{ uid: createLocatorRow().uid, value: parts[0] }] : [createLocatorRow()]
        meta.secondary_locator = parts[1] || ''
        meta.tertiary_locator = parts[2] || ''
      } else if (item.type === 'delete_element') {
        meta.locator_list = this.decodeLocatorList(item.locator, '|')
      } else if (this.showTypeField(item.type, 'locator')) {
        Object.assign(meta, this.decodeLocatorExpression(item.locator))
      }

      if (item.type === 'text_content' && (meta.locator_joiner === 'and' || meta.locator_joiner === 'or')) {
        meta.locator_raw = item.locator || ''
      }
      if (item.type === 'text_content' && meta.locator_joiner === 'single' && meta.locator_list[0]) {
        meta.locator_list[0].exist_mode = 'exist'
      }

      if (this.showTypeField(item.type, 'check_key')) {
        const checkConfig = parseCheckConfig(item.check_key || '')
        meta.check_mode = checkConfig.mode
        meta.check_rule_list = withCheckRuleUid(checkConfig.bool_rules)
        meta.compare_rule = {
          ...createCompareRule(),
          ...(checkConfig.compare_rule || {}),
        }
      }

      return meta
    },
    supportLocatorExpression(type) {
      return this.showTypeField(type, 'locator')
        && type !== 'delete_element'
        && type !== 'login_username_password'
    },
    decodeLocatorList(rawLocator, separator) {
      const list = String(rawLocator || '')
        .split(separator)
        .map(v => v.trim())
        .filter(Boolean)
        .map(v => ({ uid: createLocatorRow().uid, value: v }))
      return list.length > 0 ? list : [createLocatorRow()]
    },
    decodeLocatorExpression(rawLocator) {
      const normalizedLocator = String(rawLocator || '').trim()
      if (!normalizedLocator) {
        return {
          locator_joiner: 'single',
          locator_raw: '',
          locator_list: [createLocatorRow()],
        }
      }

      const hasAnd = normalizedLocator.includes('&&')
      const hasOr = normalizedLocator.includes('||')
      if (hasAnd && hasOr) {
        return {
          locator_joiner: 'raw',
          locator_raw: normalizedLocator,
          locator_list: [createLocatorRow()],
        }
      }

      const separator = hasAnd ? '&&' : (hasOr ? '||' : '')
      const segments = separator ? normalizedLocator.split(separator) : [normalizedLocator]
      const locatorList = segments.map(segment => this.parseLocatorSegment(segment)).filter(Boolean)

      return {
        locator_joiner: hasAnd ? 'and' : (hasOr ? 'or' : 'single'),
        locator_raw: normalizedLocator,
        locator_list: locatorList.length > 0 ? locatorList : [createLocatorRow()],
      }
    },
    parseLocatorSegment(segment) {
      const normalizedSegment = String(segment || '').trim()
      if (!normalizedSegment) return null

      const partList = normalizedSegment.split('|').map(item => item.trim()).filter(Boolean)
      let locatorValue = partList[0] || ''
      const existMode = locatorValue.startsWith('!') ? 'not_exist' : 'exist'
      if (existMode === 'not_exist') {
        locatorValue = locatorValue.slice(1)
      }

      return {
        uid: createLocatorRow().uid,
        value: locatorValue,
        exist_mode: existMode,
        match_mode: partList.includes('first') ? 'first' : 'all',
      }
    },
    serializeLocatorExpression() {
      if (this.formMeta.locator_joiner === 'raw') {
        return String(this.formMeta.locator_raw || '').trim()
      }

      const locatorList = this.formMeta.locator_list
        .map(item => {
          const locatorValue = String(item.value || '').trim()
          if (!locatorValue) return ''
          const prefix = item.exist_mode === 'not_exist' ? '!' : ''
          const suffix = item.match_mode === 'first' ? '|first' : ''
          return `${prefix}${locatorValue}${suffix}`
        })
        .filter(Boolean)

      if (locatorList.length === 0) {
        return ''
      }
      if (this.formMeta.locator_joiner === 'and') return locatorList.join('&&')
      if (this.formMeta.locator_joiner === 'or') return locatorList.join('||')
      return locatorList[0]
    },
    serializeItem() {
      // serializeItem 负责把前端表单重新编码成后端需要的流程项结构。
      // serializeItem serializes editable form state back into backend process payloads.
      const item = {
        ...this.localItem,
        next_ids: '',
      }
      const checkKeyExpression = serializeCheckConfig({
        mode: this.formMeta.check_mode,
        bool_rules: this.formMeta.check_rule_list,
        compare_rule: this.formMeta.compare_rule,
      })

      if (item.type === 'bool_result') {
        item.locator = JSON.stringify(
          this.formMeta.bool_result_rules
            .filter(rule => String(rule.locator || '').trim() !== '')
            .map(rule => ({
              locator: String(rule.locator || '').trim(),
              return: rule.return !== false,
            }))
        )
        item.value = ''
        item.out_key = this.formMeta.out_key
        item.check_key = checkKeyExpression
      } else if (item.type === 'wait_url') {
        item.locator = ''
        item.value = serializeWaitUrlValue(this.formMeta)
        item.out_key = ''
        item.check_key = checkKeyExpression
      } else if (item.type === 'redirect_uri') {
        item.locator = ''
        item.value = serializeRedirectUriValue(this.formMeta)
        item.out_key = ''
        item.check_key = checkKeyExpression
      } else if (item.type === 'no_exist_wait') {
        item.locator = this.serializeLocatorExpression()
        item.value = `${Number(this.formMeta.wait_second || 10)}|${Number(this.formMeta.wait_count || 3)}`
        item.out_key = this.formMeta.out_key
        item.check_key = checkKeyExpression
      } else if (item.type === 'login_username_password') {
        const userLocator = this.formMeta.locator_list[0] ? this.formMeta.locator_list[0].value : ''
        item.locator = [userLocator, this.formMeta.secondary_locator, this.formMeta.tertiary_locator].filter(Boolean).join('||')
        item.value = ''
        item.out_key = ''
        item.check_key = ''
      } else if (item.type === 'delete_element') {
        item.locator = this.formMeta.locator_list.map(v => v.value.trim()).filter(Boolean).join('|')
        item.value = this.formMeta.delete_mode
        item.out_key = ''
        item.check_key = ''
      } else {
        if (this.showTypeField(item.type, 'locator')) {
          item.locator = this.supportLocatorExpression(item.type)
            ? this.serializeLocatorExpression()
            : this.formMeta.locator_list.map(v => v.value.trim()).filter(Boolean).join('||')
        } else {
          item.locator = ''
        }
        item.value = this.formMeta.value
        item.out_key = this.formMeta.out_key
        item.check_key = checkKeyExpression
      }

      return item
    },
    emitChange() {
      if (this.syncingFromParent) return
      const serializedItem = this.serializeItem()
      if (Object.keys(this.fieldErrors).length > 0) {
        this.runValidation()
      }
      const nextSignature = this.createSignature(serializedItem)
      if (nextSignature === this.lastSerializedSignature) {
        return
      }
      this.lastSerializedSignature = nextSignature
      this.$emit('update:modelValue', serializedItem)
    },
    showTypeField(type, fieldName) {
      return (PROCESS_TYPE_FIELDS[type] || []).includes(fieldName)
    },
    showField(fieldName) {
      return this.currentFields.includes(fieldName)
    },
    fieldGuide(fieldName) {
      if (fieldName === 'locator') {
        if (this.localItem.type === 'text_content') {
          if (this.formMeta.locator_joiner === 'single') {
            return '提取元素内容时，后端会直接读取这条定位命中的元素文本。单条定位时不需要再配置“要求找不到”，最稳妥的用法就是只保留 1 条定位。'
          }
          if (this.formMeta.locator_joiner === 'and') {
            return '这是旧的多条件提取配置。后端会按整条 && 表达式处理，不存在“第一条单独提取、后面只做条件”的独立语义，所以这里改成原始表达式展示更准确。'
          }
          if (this.formMeta.locator_joiner === 'or') {
            return '这是旧的多条件提取配置。后端会按整条 || 表达式处理，最终提取目标不够直观，所以这里改成原始表达式展示更准确。'
          }
        }
        if (this.formMeta.locator_joiner === 'and') {
          return '当前为“多个条件都满足”：下面每一条定位条件都必须符合，节点才会继续执行。如果页面上找到多个同类元素，“多个时按默认方式处理”表示沿用后端默认查找方式；“多个时只取第一个”表示只看第一个匹配元素。'
        }
        if (this.formMeta.locator_joiner === 'or') {
          return '当前为“多个条件满足其一”：下面任意一条定位条件符合即可。如果页面上找到多个同类元素，“多个时按默认方式处理”表示沿用后端默认查找方式；“多个时只取第一个”表示只看第一个匹配元素。'
        }
        if (this.formMeta.locator_joiner === 'raw') {
          return '高级表达式模式适合兼容旧配置或直接粘贴后端语法；普通场景建议优先使用上面的结构化选项。'
        }
        if (this.formMeta.locator_joiner === 'single') {
          return '当前为“只找一个元素”：通常只需要填 1 条定位。“多个时按默认方式处理”表示沿用后端默认查找方式；“多个时只取第一个”表示只取第一个匹配元素。'
        }
      }
      if (fieldName === 'check_key') {
        if (this.formMeta.check_mode === 'compare') {
          return '内容比较会生成类似 {login_user}!={user_name}、{login_user}!={password} 或 {sign_in_btn}==Sign in 的条件。左侧选前面节点的输出，比较类型选“等于/不等于”，右侧既可以直接选择注入值，也可以输入固定字符串。'
        }
        if (this.formMeta.check_mode === 'bool') {
          return '结果判断会生成类似 {need_login}、{need_login}&&{qrcode_dialog} 或 {need_login}&&{need_change_to_password} 的条件。后端这里只支持多个条件用 && 同时成立，不支持或条件；单独一个条件就表示该输出为 true 时执行。'
        }
        return '可以不限制执行，也可以按前面节点的结果判断，或按文本内容比较判断。'
      }
      return PROCESS_ITEM_FIELD_GUIDES[fieldName] || ''
    },
    fieldError(fieldName) {
      return this.fieldErrors[fieldName] || ''
    },
    runValidation() {
      const validationResult = validateProcessItemForm({
        item: this.serializeItem(),
        formMeta: this.formMeta,
      })
      this.fieldErrors = validationResult.fieldErrors || {}
      return validationResult
    },
    validateForSave() {
      return this.runValidation().valid
    },
    fieldLabel(fieldName) {
      const labels = {
        locator: '主元素定位',
        secondary_locator: '密码框定位',
        tertiary_locator: '提交按钮定位',
        value: '值',
        out_key: '输出键',
        check_key: '是否执行判断',
        wait_second: '等待秒数',
        wait_count: '轮询次数',
        response_url: '等待地址',
        delete_mode: '删除类型',
        register_response_urls: '跳转后等待地址',
      }
      if (this.localItem.type === 'login_username_password' && fieldName === 'locator') {
        return '用户名框定位'
      }
      return labels[fieldName] || fieldName
    },
    fieldPlaceholder(fieldName) {
      if (fieldName === 'locator') return '例如 .username 或 .login-btn'
      if (fieldName === 'secondary_locator') return '例如 #password'
      if (fieldName === 'tertiary_locator') return '例如 .submit-btn'
      if (this.localItem.type === 'input' && fieldName === 'value') return '支持 {user_name} / {password} / {rand}'
      if (fieldName === 'response_url') return '例如 {scheme}://{domain}/api/login'
      return ''
    },
    textareaRows(fieldName) {
      return fieldName === 'value' ? 2 : 1
    },
    addRegisterResponseUrl() {
      this.formMeta.register_response_urls.push(createRegisterUrl())
    },
    removeRegisterResponseUrl(index) {
      this.formMeta.register_response_urls.splice(index, 1)
    },
    addLocatorRow() {
      this.formMeta.locator_list.push(createLocatorRow())
    },
    removeLocatorRow(index) {
      this.formMeta.locator_list.splice(index, 1)
    },
    addBoolResultRule() {
      this.formMeta.bool_result_rules.push(createBoolResultRule())
    },
    removeBoolResultRule(index) {
      this.formMeta.bool_result_rules.splice(index, 1)
    },
    addCheckRule() {
      this.formMeta.check_rule_list.push(createCheckRule())
    },
    removeCheckRule(index) {
      this.formMeta.check_rule_list.splice(index, 1)
    },
    // applyCompareRightQuickPick 用于把注入变量快捷写入比较右值输入框。
    // applyCompareRightQuickPick applies injected-variable shortcuts into the compare right input.
    applyCompareRightQuickPick(value) {
      this.formMeta.compare_rule.right = String(value || '')
    },
  },
}
</script>

<style scoped>
.list-editor {
  width: 100%;
}

.locator-expression-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.locator-expression-toolbar__select {
  width: 220px;
}

.locator-expression-toolbar__tip,
.field-guide {
  color: #6b7b68;
  font-size: 12px;
  line-height: 1.5;
}

.field-guide {
  margin-top: 8px;
}

.locator-purpose-card {
  margin-bottom: 10px;
  padding: 10px 12px;
  background: #f6f8f4;
  border: 1px solid #dfe8d7;
  border-radius: 8px;
}

.locator-purpose-card__title {
  color: #48653a;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 4px;
}

.locator-purpose-card__text {
  color: #6b7b68;
  font-size: 12px;
  line-height: 1.6;
}

.locator-behavior-summary {
  margin-top: 10px;
  padding: 10px 12px;
  background: #fffdf5;
  border: 1px solid #efe4b0;
  border-radius: 8px;
}

.locator-behavior-summary__title {
  color: #8a6b13;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 4px;
}

.locator-behavior-summary__text {
  color: #7b714d;
  font-size: 12px;
  line-height: 1.6;
}

.locator-expression-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 130px 130px 60px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.locator-expression-row__mode {
  width: 100%;
}

.list-editor__row,
.response-url-row {
  display: grid;
  grid-template-columns: 1fr 120px 60px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.list-editor__row {
  grid-template-columns: 1fr 60px;
}

.bool-result-row {
  display: grid;
  grid-template-columns: 1fr 140px 60px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.check-rule-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 140px 60px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.check-rule-row__mode {
  width: 100%;
}

.check-mode-select {
  width: 220px;
  margin-bottom: 10px;
}

.compare-rule-row {
  display: grid;
  grid-template-columns: minmax(260px, 1fr) 140px minmax(320px, 1fr);
  gap: 10px;
  align-items: start;
  margin-bottom: 10px;
}

.compare-rule-row__operator {
  width: 100%;
}

.compare-rule-right {
  min-width: 0;
}

.compare-rule-right__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.plain-number-input {
  width: 100%;
}
</style>
