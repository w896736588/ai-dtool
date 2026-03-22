<template>
  <el-form :model="localItem" label-width="180px" class="process-item-editor">
    <el-form-item label="名称">
      <el-input v-model="localItem.name" />
    </el-form-item>
    <el-form-item label="类型">
      <el-select v-model="localItem.type" placeholder="请选择类型" style="width: 100%">
        <el-option v-for="option in processTypeOptions" :key="option.value" :label="option.label" :value="option.value" />
      </el-select>
    </el-form-item>
    <el-form-item label="前端执行提示">
      <el-input v-model="localItem.tip" />
    </el-form-item>

    <template v-if="showField('locator')">
      <el-form-item :label="fieldLabel('locator')">
        <el-input v-model="formMeta.locator" :placeholder="fieldPlaceholder('locator')" />
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
        <el-input v-model="formMeta.value" :placeholder="fieldPlaceholder('value')" type="textarea" :rows="textareaRows('value')" />
      </el-form-item>
    </template>
    <template v-if="showField('out_key')">
      <el-form-item :label="fieldLabel('out_key')">
        <el-input v-model="formMeta.out_key" :placeholder="fieldPlaceholder('out_key')" />
      </el-form-item>
    </template>
    <template v-if="showField('check_key')">
      <el-form-item :label="fieldLabel('check_key')">
        <el-input v-model="formMeta.check_key" :placeholder="fieldPlaceholder('check_key')" />
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
    <template v-if="showField('expected_result')">
      <el-form-item :label="fieldLabel('expected_result')">
        <el-select v-model="formMeta.expected_result" style="width: 100%">
          <el-option label="true" value="true" />
          <el-option label="false" value="false" />
        </el-select>
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
        <div class="response-url-editor">
          <div v-for="(item, index) in formMeta.register_response_urls" :key="item.uid" class="response-url-row">
            <el-input v-model="item.url" placeholder="等待地址" />
            <el-input-number v-model="item.wait_second" :min="1" />
            <el-button link type="danger" @click="removeRegisterResponseUrl(index)">删除</el-button>
          </div>
          <el-button size="small" plain @click="addRegisterResponseUrl">新增等待地址</el-button>
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
      <el-input v-model="localItem.domain_limit" />
    </el-form-item>
    <el-form-item label="输出追加到替换列表" v-if="allowAppendToReplace">
      <el-select v-model="localItem.append_to_replace" style="width: 100%">
        <el-option label="添加" value="1" />
        <el-option label="不添加" value="0" />
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
    <el-form-item label="下一个节点ID">
      <el-input v-model="localItem.next_ids" placeholder="多个用逗号分隔，例如 2,3,4" />
    </el-form-item>
  </el-form>
</template>

<script>
const createDefaultItem = () => ({ id: 0, name: '', smart_link_process_id: 0, type: '', locator: '', wait_mills: 3000, tip: '', value: '', out_key: '', check_key: '', weight: 0, domain_limit: '', append_to_replace: '0', is_async: '0', is_error_continue: '0', next_ids: '', x: 0, y: 0 })
const PROCESS_TYPE_FIELDS = {
  text_content: ['locator', 'out_key'],
  redirect_uri: ['value', 'register_response_urls'],
  wait_url: ['response_url', 'wait_second'],
  wait: [],
  bool_result: ['locator', 'out_key', 'expected_result'],
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
  { label: '结束本次打开的网页 close', value: 'close' },
  { label: '存在时等待 no_exist_wait', value: 'no_exist_wait' },
  { label: '提取 canvas 图片 canvas_image', value: 'canvas_image' },
  { label: '输入账号密码 login_username_password', value: 'login_username_password' },
  { label: '删除元素 delete_element', value: 'delete_element' },
]
function safeParseJson(text, fallback) { if (!text) return fallback; try { return JSON.parse(text) } catch (error) { return fallback } }
function createRegisterUrl() { return { uid: `response-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`, url: '', wait_second: 10 } }

export default {
  name: 'ProcessItemEditor',
  props: { modelValue: { type: Object, default: () => createDefaultItem() } },
  emits: ['update:modelValue'],
  data() {
    return {
      localItem: createDefaultItem(),
      formMeta: { locator: '', secondary_locator: '', tertiary_locator: '', value: '', out_key: '', check_key: '', wait_second: 10, wait_count: 3, response_url: '', expected_result: 'true', delete_mode: 'class', register_response_urls: [] },
      syncingFromParent: false,
      processTypeOptions: PROCESS_TYPE_OPTIONS,
    }
  },
  computed: {
    currentFields() { return PROCESS_TYPE_FIELDS[this.localItem.type] || [] },
    allowAppendToReplace() { return this.localItem.type !== 'click' && this.localItem.type !== 'delete_element' },
  },
  watch: {
    modelValue: { deep: true, immediate: true, handler(value) { this.syncFromModel(value || createDefaultItem()) } },
    localItem: { deep: true, handler() { this.emitChange() } },
    formMeta: { deep: true, handler() { this.emitChange() } },
    'localItem.type'(nextType, prevType) { if (!this.syncingFromParent && nextType !== prevType) this.resetMetaForType(nextType) },
  },
  methods: {
    syncFromModel(value) {
      this.syncingFromParent = true
      this.localItem = { ...createDefaultItem(), ...JSON.parse(JSON.stringify(value)), next_ids: value.next_ids || '', append_to_replace: String(value.append_to_replace ?? '0'), is_async: String(value.is_async ?? '0'), is_error_continue: String(value.is_error_continue ?? '0') }
      this.formMeta = this.deserializeMeta(this.localItem)
      this.syncingFromParent = false
    },
    resetMetaForType(type) { this.formMeta = this.deserializeMeta({ ...this.localItem, type, locator: '', value: '', out_key: '', check_key: '' }) },
    deserializeMeta(item) {
      const meta = { locator: item.locator || '', secondary_locator: '', tertiary_locator: '', value: item.value || '', out_key: item.out_key || '', check_key: item.check_key || '', wait_second: 10, wait_count: 3, response_url: '', expected_result: 'true', delete_mode: item.value || 'class', register_response_urls: [] }
      if (item.type === 'wait_url') {
        const parsed = safeParseJson(item.value, {})
        meta.response_url = parsed.ResponseUrl || ''
        meta.wait_second = Number(parsed.WaitSecond || 10)
      } else if (item.type === 'redirect_uri') {
        const parsed = safeParseJson(item.value, null)
        if (parsed && typeof parsed === 'object' && parsed.Url) {
          meta.value = parsed.Url || ''
          meta.register_response_urls = Array.isArray(parsed.RegisterResponseUrl) ? parsed.RegisterResponseUrl.map(v => ({ uid: createRegisterUrl().uid, url: v.Url || '', wait_second: Number(v.WaitSecond || 10) })) : []
        }
      } else if (item.type === 'bool_result') {
        meta.expected_result = item.check_key === 'false' ? 'false' : 'true'
      } else if (item.type === 'no_exist_wait') {
        const [waitSecond, waitCount] = String(item.value || '').split('|')
        meta.wait_second = Number(waitSecond || 10)
        meta.wait_count = Number(waitCount || 3)
      } else if (item.type === 'login_username_password') {
        const parts = String(item.locator || '').split('||')
        meta.locator = parts[0] || ''
        meta.secondary_locator = parts[1] || ''
        meta.tertiary_locator = parts[2] || ''
      }
      return meta
    },
    serializeItem() {
      const item = { ...this.localItem }
      if (item.type === 'wait_url') {
        item.locator = ''
        item.value = JSON.stringify({ ResponseUrl: this.formMeta.response_url, WaitSecond: Number(this.formMeta.wait_second || 10) })
        item.out_key = ''
        item.check_key = ''
      } else if (item.type === 'redirect_uri') {
        item.locator = ''
        item.value = this.formMeta.register_response_urls.length > 0
          ? JSON.stringify({ Url: this.formMeta.value, RegisterResponseUrl: this.formMeta.register_response_urls.filter(v => v.url).map(v => ({ Url: v.url, WaitSecond: Number(v.wait_second || 10) })) })
          : this.formMeta.value
        item.out_key = ''
        item.check_key = ''
      } else if (item.type === 'bool_result') {
        item.locator = this.formMeta.locator
        item.out_key = this.formMeta.out_key
        item.check_key = this.formMeta.expected_result
        item.value = ''
      } else if (item.type === 'no_exist_wait') {
        item.locator = this.formMeta.locator
        item.out_key = this.formMeta.out_key
        item.value = `${Number(this.formMeta.wait_second || 10)}|${Number(this.formMeta.wait_count || 3)}`
        item.check_key = ''
      } else if (item.type === 'login_username_password') {
        item.locator = [this.formMeta.locator, this.formMeta.secondary_locator, this.formMeta.tertiary_locator].filter(Boolean).join('||')
        item.value = ''
        item.out_key = ''
        item.check_key = ''
      } else if (item.type === 'delete_element') {
        item.locator = this.formMeta.locator
        item.value = this.formMeta.delete_mode
        item.out_key = ''
        item.check_key = ''
      } else {
        item.locator = this.formMeta.locator
        item.value = this.formMeta.value
        item.out_key = this.formMeta.out_key
        item.check_key = this.formMeta.check_key
      }
      return item
    },
    emitChange() { if (!this.syncingFromParent) this.$emit('update:modelValue', this.serializeItem()) },
    showField(name) { return this.currentFields.includes(name) },
    fieldLabel(name) {
      const labels = { locator: '主元素定位', secondary_locator: '密码框定位', tertiary_locator: '提交按钮定位', value: '值', out_key: '输出键', check_key: '判断键', wait_second: '等待秒数', wait_count: '轮询次数', response_url: '等待地址', expected_result: '期望判断结果', delete_mode: '删除类型', register_response_urls: '跳转后等待地址' }
      if (this.localItem.type === 'login_username_password' && name === 'locator') return '用户名框定位'
      return labels[name] || name
    },
    fieldPlaceholder(name) {
      if (name === 'locator') return '支持多个 locator，用 || 分隔'
      if (this.localItem.type === 'redirect_uri' && name === 'value') return '例如 /home 或 https://example.com/home'
      if (this.localItem.type === 'input' && name === 'value') return '支持 {user_name} / {password} / {rand}'
      if (name === 'response_url') return '例如 {scheme}://{domain}/api/login'
      return ''
    },
    textareaRows(name) { return name === 'value' && this.localItem.type === 'redirect_uri' ? 2 : 1 },
    addRegisterResponseUrl() { this.formMeta.register_response_urls.push(createRegisterUrl()) },
    removeRegisterResponseUrl(index) { this.formMeta.register_response_urls.splice(index, 1) },
  },
}
</script>

<style scoped>
.response-url-row { display: grid; grid-template-columns: 1fr 120px 60px; gap: 10px; align-items: center; margin-bottom: 10px; }
</style>
