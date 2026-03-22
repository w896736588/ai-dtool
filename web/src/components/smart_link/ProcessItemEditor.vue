<template>
  <el-form :model="localItem" label-width="180px" class="process-item-editor">
    <el-form-item label="名称">
      <el-input v-model="localItem.name" />
    </el-form-item>

    <el-form-item label="类型">
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
      <el-form-item :label="fieldLabel('locator')">
        <div class="list-editor">
          <template v-if="useLocatorExpressionEditor">
            <div class="locator-expression-toolbar">
              <el-select v-model="formMeta.locator_joiner" class="locator-expression-toolbar__select">
                <el-option label="单个定位" value="single" />
                <el-option label="且条件 &&" value="and" />
                <el-option label="或条件 ||" value="or" />
                <el-option label="原始表达式" value="raw" />
              </el-select>
              <div class="locator-expression-toolbar__tip">
                支持 `!selector` 表示不存在才成功，`selector|first` 表示只取首个。
              </div>
            </div>

            <template v-if="formMeta.locator_joiner === 'raw'">
              <el-input
                v-model="formMeta.locator_raw"
                type="textarea"
                :rows="3"
                placeholder="例如 .username&&!.btn.login_as_reg_btn 或 .dialog|first"
              />
            </template>

            <template v-else>
              <div v-for="(item, index) in formMeta.locator_list" :key="item.uid" class="locator-expression-row">
                <el-input v-model="item.value" :placeholder="fieldPlaceholder('locator')" />
                <el-select v-model="item.exist_mode" class="locator-expression-row__mode">
                  <el-option label="存在" value="exist" />
                  <el-option label="不存在" value="not_exist" />
                </el-select>
                <el-select v-model="item.match_mode" class="locator-expression-row__mode">
                  <el-option label="全部匹配" value="all" />
                  <el-option label="仅首个" value="first" />
                </el-select>
                <el-button class="action-btn action-btn--danger" size="small" native-type="button" @click="removeLocatorRow(index)">删除</el-button>
              </div>
              <el-button class="action-btn action-btn--primary" size="small" native-type="button" @click="addLocatorRow">新增定位</el-button>
            </template>
          </template>

          <template v-else>
            <div v-for="(item, index) in formMeta.locator_list" :key="item.uid" class="list-editor__row">
              <el-input v-model="item.value" :placeholder="fieldPlaceholder('locator')" />
              <el-button class="action-btn action-btn--danger" size="small" native-type="button" @click="removeLocatorRow(index)">删除</el-button>
            </div>
            <el-button class="action-btn action-btn--primary" size="small" native-type="button" @click="addLocatorRow">新增定位</el-button>
          </template>
        </div>
      </el-form-item>
    </template>

    <template v-if="showField('secondary_locator')">
      <el-form-item :label="fieldLabel('secondary_locator')">
        <el-input v-model="formMeta.secondary_locator" :placeholder="fieldPlaceholder('secondary_locator')" />
      </el-form-item>
    </template>

    <template v-if="showField('tertiary_locator')">
      <el-form-item :label="fieldLabel('tertiary_locator')">
        <el-input v-model="formMeta.tertiary_locator" :placeholder="fieldPlaceholder('tertiary_locator')" />
      </el-form-item>
    </template>

    <template v-if="showField('value')">
      <el-form-item :label="fieldLabel('value')">
        <el-input
          v-model="formMeta.value"
          :placeholder="fieldPlaceholder('value')"
          type="textarea"
          :rows="textareaRows('value')"
        />
      </el-form-item>
    </template>

    <template v-if="showField('out_key')">
      <el-form-item :label="fieldLabel('out_key')">
        <el-input v-model="formMeta.out_key" placeholder="例如 login_state" />
      </el-form-item>
    </template>

    <template v-if="showField('check_key')">
      <el-form-item :label="fieldLabel('check_key')">
        <el-input v-model="formMeta.check_key" placeholder="例如 login_state" />
      </el-form-item>
    </template>

    <template v-if="showField('wait_second')">
      <el-form-item :label="fieldLabel('wait_second')">
        <el-input-number v-model="formMeta.wait_second" :min="1" />
      </el-form-item>
    </template>

    <template v-if="showField('wait_count')">
      <el-form-item :label="fieldLabel('wait_count')">
        <el-input-number v-model="formMeta.wait_count" :min="1" />
      </el-form-item>
    </template>

    <template v-if="showField('response_url')">
      <el-form-item :label="fieldLabel('response_url')">
        <el-input v-model="formMeta.response_url" :placeholder="fieldPlaceholder('response_url')" />
      </el-form-item>
    </template>

    <template v-if="showField('delete_mode')">
      <el-form-item :label="fieldLabel('delete_mode')">
        <el-select v-model="formMeta.delete_mode" style="width: 100%">
          <el-option label="按 class 删除" value="class" />
        </el-select>
      </el-form-item>
    </template>

    <template v-if="showField('register_response_urls')">
      <el-form-item :label="fieldLabel('register_response_urls')">
        <div class="list-editor">
          <div v-for="(item, index) in formMeta.register_response_urls" :key="item.uid" class="response-url-row">
            <el-input v-model="item.url" placeholder="等待地址" />
            <el-input-number v-model="item.wait_second" :min="1" />
            <el-button class="action-btn action-btn--danger" size="small" native-type="button" @click="removeRegisterResponseUrl(index)">删除</el-button>
          </div>
          <el-button class="action-btn action-btn--primary" size="small" native-type="button" @click="addRegisterResponseUrl">新增等待地址</el-button>
        </div>
      </el-form-item>
    </template>

    <template v-if="showField('bool_result_rules')">
      <el-form-item label="主元素定位">
        <div class="list-editor">
          <div v-for="(item, index) in formMeta.bool_result_rules" :key="item.uid" class="bool-result-row">
            <el-input v-model="item.locator" placeholder="例如 .user-info.ant-dropdown-trigger" />
            <el-select v-model="item.return" style="width: 140px">
              <el-option label="存在返回 true" :value="true" />
              <el-option label="存在返回 false" :value="false" />
            </el-select>
            <el-button class="action-btn action-btn--danger" size="small" native-type="button" @click="removeBoolResultRule(index)">删除</el-button>
          </div>
          <el-button class="action-btn action-btn--primary" size="small" native-type="button" @click="addBoolResultRule">新增定位</el-button>
        </div>
      </el-form-item>
    </template>

    <el-form-item label="权重">
      <el-input-number v-model="localItem.weight" :min="0" />
    </el-form-item>

    <el-form-item label="等待时长(ms)">
      <el-input-number v-model="localItem.wait_mills" :min="0" />
    </el-form-item>

    <el-form-item label="域名限制">
      <el-input v-model="localItem.domain_limit" placeholder="可选，例如 example.com" />
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

    <el-form-item label="下一节点">
      <el-select
        v-model="formMeta.next_id_list"
        multiple
        clearable
        filterable
        collapse-tags
        collapse-tags-tooltip
        placeholder="请选择后续节点"
        style="width: 100%"
      >
        <el-option
          v-for="option in nextNodeOptions"
          :key="option.id"
          :label="`#${option.id} ${option.name || option.type || '未命名节点'}`"
          :value="String(option.id)"
        />
      </el-select>
    </el-form-item>
  </el-form>
</template>

<script>
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

const PROCESS_TYPE_FIELDS = {
  text_content: ['locator', 'out_key'],
  redirect_uri: ['value', 'register_response_urls'],
  wait_url: ['response_url', 'wait_second'],
  wait: [],
  bool_result: ['bool_result_rules', 'out_key'],
  bool_exist: ['locator', 'out_key'],
  click: ['locator'],
  input: ['locator', 'value', 'out_key'],
  close: [],
  no_exist_wait: ['locator', 'wait_second', 'wait_count', 'out_key'],
  canvas_image: ['locator', 'out_key'],
  login_username_password: ['locator', 'secondary_locator', 'tertiary_locator'],
  delete_element: ['locator', 'delete_mode'],
}

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

export default {
  name: 'ProcessItemEditor',
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
        wait_second: 10,
        wait_count: 3,
        response_url: '',
        delete_mode: 'class',
        register_response_urls: [],
        bool_result_rules: [],
        next_id_list: [],
      },
      syncingFromParent: false,
      lastSerializedSignature: '',
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
    nextNodeOptions() {
      return (this.processItemOptions || []).filter(item => String(item.id) !== String(this.localItem.id || ''))
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
    },
    deserializeMeta(item) {
      const meta = {
        locator_list: [],
        locator_joiner: 'single',
        locator_raw: '',
        secondary_locator: '',
        tertiary_locator: '',
        value: item.value || '',
        out_key: item.out_key || '',
        check_key: item.check_key || '',
        wait_second: 10,
        wait_count: 3,
        response_url: '',
        delete_mode: item.value || 'class',
        register_response_urls: [],
        bool_result_rules: [],
        next_id_list: String(item.next_ids || '').split(',').map(v => v.trim()).filter(Boolean),
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
        const parsed = safeParseJson(item.value, {})
        meta.response_url = parsed.ResponseUrl || ''
        meta.wait_second = Number(parsed.WaitSecond || 10)
      } else if (item.type === 'redirect_uri') {
        const parsed = safeParseJson(item.value, null)
        if (parsed && typeof parsed === 'object' && parsed.Url) {
          meta.value = parsed.Url || ''
          meta.register_response_urls = Array.isArray(parsed.RegisterResponseUrl)
            ? parsed.RegisterResponseUrl.map(v => ({
              uid: createRegisterUrl().uid,
              url: v.Url || '',
              wait_second: Number(v.WaitSecond || 10),
            }))
            : []
        }
      } else if (item.type === 'no_exist_wait') {
        const [waitSecond, waitCount] = String(item.value || '').split('|')
        meta.wait_second = Number(waitSecond || 10)
        meta.wait_count = Number(waitCount || 3)
        Object.assign(meta, this.decodeLocatorExpression(item.locator))
      } else if (item.type === 'login_username_password') {
        const parts = String(item.locator || '').split('||')
        meta.locator_list = []
        meta.secondary_locator = parts[1] || ''
        meta.tertiary_locator = parts[2] || ''
        if (parts[0]) {
          meta.locator_list = [{ uid: createLocatorRow().uid, value: parts[0] }]
        }
      } else if (item.type === 'delete_element') {
        meta.locator_list = this.decodeLocatorList(item.locator, '|')
      } else if (this.showTypeField(item.type, 'locator')) {
        Object.assign(meta, this.decodeLocatorExpression(item.locator))
      }

      return meta
    },
    // 判断当前类型是否支持 locator 表达式编辑 / Decide whether this type should use the structured locator expression editor.
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
    // 解析后端 locator 表达式 / Parse backend locator expression into structured rows.
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
      // 同时存在 && 和 || 时前端无法稳定还原优先级，切到原始模式兜底 / Fall back to raw mode when mixed operators cannot be losslessly reconstructed.
      if (hasAnd && hasOr) {
        return {
          locator_joiner: 'raw',
          locator_raw: normalizedLocator,
          locator_list: [createLocatorRow()],
        }
      }

      const separator = hasAnd ? '&&' : (hasOr ? '||' : '')
      const segments = separator ? normalizedLocator.split(separator) : [normalizedLocator]
      const locatorList = segments
        .map(segment => this.parseLocatorSegment(segment))
        .filter(Boolean)

      return {
        locator_joiner: hasAnd ? 'and' : (hasOr ? 'or' : 'single'),
        locator_raw: normalizedLocator,
        locator_list: locatorList.length > 0 ? locatorList : [createLocatorRow()],
      }
    },
    // 解析单条 locator 规则 / Parse a single locator segment including negation and first-only flags.
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
    // 序列化 locator 规则行 / Serialize structured locator rows back to backend expression format.
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

      if (this.formMeta.locator_joiner === 'and') {
        return locatorList.join('&&')
      }
      if (this.formMeta.locator_joiner === 'or') {
        return locatorList.join('||')
      }
      return locatorList[0]
    },
    serializeItem() {
      const item = {
        ...this.localItem,
        next_ids: this.formMeta.next_id_list.join(','),
      }

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
        item.check_key = ''
      } else if (item.type === 'wait_url') {
        item.locator = ''
        item.value = JSON.stringify({
          ResponseUrl: this.formMeta.response_url,
          WaitSecond: Number(this.formMeta.wait_second || 10),
        })
        item.out_key = ''
        item.check_key = ''
      } else if (item.type === 'redirect_uri') {
        item.locator = ''
        item.value = this.formMeta.register_response_urls.length > 0
          ? JSON.stringify({
            Url: this.formMeta.value,
            RegisterResponseUrl: this.formMeta.register_response_urls
              .filter(v => v.url)
              .map(v => ({
                Url: v.url,
                WaitSecond: Number(v.wait_second || 10),
              })),
          })
          : this.formMeta.value
        item.out_key = ''
        item.check_key = ''
      } else if (item.type === 'no_exist_wait') {
        item.locator = this.serializeLocatorExpression()
        item.value = `${Number(this.formMeta.wait_second || 10)}|${Number(this.formMeta.wait_count || 3)}`
        item.out_key = this.formMeta.out_key
        item.check_key = ''
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
        item.check_key = this.formMeta.check_key
      }

      return item
    },
    emitChange() {
      if (this.syncingFromParent) return
      const serializedItem = this.serializeItem()
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
    fieldLabel(fieldName) {
      const labels = {
        locator: '主元素定位',
        secondary_locator: '密码框定位',
        tertiary_locator: '提交按钮定位',
        value: '值',
        out_key: '输出键',
        check_key: '判断键',
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
      if (fieldName === 'locator') return '例如 .username、!.login-btn、.dialog|first'
      if (fieldName === 'secondary_locator') return '例如 #password'
      if (fieldName === 'tertiary_locator') return '例如 .submit-btn'
      if (this.localItem.type === 'redirect_uri' && fieldName === 'value') return '例如 /home 或 https://example.com/home'
      if (this.localItem.type === 'input' && fieldName === 'value') return '支持 {user_name} / {password} / {rand}'
      if (fieldName === 'response_url') return '例如 {scheme}://{domain}/api/login'
      return ''
    },
    textareaRows(fieldName) {
      return fieldName === 'value' && this.localItem.type === 'redirect_uri' ? 2 : 1
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
  },
}
</script>

<style scoped>
/* 统一局部操作按钮风格，避免弹窗内出现白底黑字按钮 / Keep local action buttons aligned with page visual style. */
.action-btn {
  border-radius: 999px;
  font-weight: 600;
  transition: all 0.2s ease;
}

.action-btn--primary {
  color: #2f6b45;
  border-color: #b8d7c1;
  background: linear-gradient(180deg, #f4fbf6 0%, #e8f5ec 100%);
}

.action-btn--primary:hover {
  color: #255639;
  border-color: #97c2a5;
  background: linear-gradient(180deg, #edf8f0 0%, #dceee2 100%);
}

.action-btn--danger {
  color: #b44343;
  border-color: #efc4c4;
  background: #fff4f4;
}

.action-btn--danger:hover {
  color: #963737;
  border-color: #e5a7a7;
  background: #feeaea;
}

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
  width: 180px;
}

.locator-expression-toolbar__tip {
  color: #6b7b68;
  font-size: 12px;
}

.locator-expression-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 110px 110px 60px;
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
  grid-template-columns: 1fr 60px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.response-url-row {
  grid-template-columns: 1fr 120px 60px;
}

.bool-result-row {
  display: grid;
  grid-template-columns: 1fr 140px 60px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}
</style>
