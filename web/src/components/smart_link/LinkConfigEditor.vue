<template>
  <el-form class="link-config-editor" label-position="top">
    <div class="editor-section">
      <div class="editor-section__header">
        <div class="editor-section__title">信息提取</div>
        <GitActionButton compact size="small" @click="addCookieItem">新增规则</GitActionButton>
      </div>
      <div v-if="cookieItems.length === 0" class="editor-empty">暂无信息提取规则。</div>
      <div v-for="(item, index) in cookieItems" :key="item.uid" class="editor-card">
        <div class="editor-card__header">
          <div class="editor-card__title">规则 {{ index + 1 }}</div>
          <GitActionButton compact size="small" variant="danger" @click="removeCookieItem(index)">删除</GitActionButton>
        </div>
        <el-row :gutter="12">
          <el-col :span="8">
            <el-form-item label="查找类型">
              <el-select v-model="item.find_type" style="width: 100%">
                <el-option label="cookie" value="cookie" />
                <el-option label="any" value="any" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="显示名称">
              <el-input v-model="item.label" />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="查找键">
              <el-input v-model="item.find_key" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="正则查找键">
              <el-input v-model="item.regex_find_key" placeholder="可选" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="格式化列表">
              <el-input v-model="item.format_list_text" placeholder="多个值用逗号分隔，例如 url_decode" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="域名列表">
          <el-input v-model="item.domain_list_text" placeholder="多个域名用逗号分隔" />
        </el-form-item>
      </div>
    </div>

    <div class="editor-section">
      <div class="editor-section__header">
        <div class="editor-section__title">请求拦截（半匹配）</div>
        <GitActionButton compact size="small" @click="addFilterItem">新增规则</GitActionButton>
      </div>
      <div v-if="filterItems.length === 0" class="editor-empty">暂无请求拦截规则。</div>
      <div v-for="(item, index) in filterItems" :key="item.uid" class="filter-row">
        <el-input v-model="item.value" :placeholder="`规则 ${index + 1}`" />
        <GitActionButton compact size="small" variant="danger" @click="removeFilterItem(index)">删除</GitActionButton>
      </div>
    </div>
  </el-form>
</template>

<script>
import GitActionButton from '@/components/base/GitActionButton.vue'

const createUid = (prefix) => `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`
const createCookieItem = () => ({
  uid: createUid('cookie'),
  find_type: 'cookie',
  format_list_text: '',
  find_key: '',
  regex_find_key: '',
  label: '',
  domain_list_text: '',
})
const createFilterItem = () => ({ uid: createUid('filter'), value: '' })

function safeParseJson(text, fallback) {
  if (!text) return fallback
  try { return JSON.parse(text) } catch (error) { return fallback }
}

export default {
  name: 'LinkConfigEditor',
  components: { GitActionButton },
  props: {
    modelValue: { type: Object, default: () => ({}) },
  },
  emits: ['update:modelValue'],
  data() {
    return {
      cookieItems: [],
      filterItems: [],
      syncingFromParent: false,
      lastEditorPayloadSignature: '',
    }
  },
  watch: {
    modelValue: {
      deep: true, immediate: true,
      handler(value) { this.syncFromModel(value || {}) },
    },
    cookieItems: { deep: true, handler() { this.emitChange() } },
    filterItems: { deep: true, handler() { this.emitChange() } },
  },
  methods: {
    buildEditorPayload() {
      const showCookies = this.cookieItems.map(item => ({
        find_type: item.find_type,
        format_list: item.format_list_text.split(',').map(v => v.trim()).filter(Boolean),
        find_key: item.find_key,
        regex_find_key: item.regex_find_key,
        label: item.label,
        Domain_list: item.domain_list_text.split(',').map(v => v.trim()).filter(Boolean),
      }))
      const filterUris = this.filterItems.map(item => item.value.trim()).filter(Boolean).join('\n')
      return {
        show_cookies: JSON.stringify(showCookies),
        filter_uris: filterUris,
      }
    },
    createPayloadSignature(payload) {
      return JSON.stringify(payload || {})
    },
    syncFromModel(value) {
      const incomingSignature = this.createPayloadSignature({
        show_cookies: value.show_cookies || '',
        filter_uris: value.filter_uris || '',
      })
      if (!this.syncingFromParent && incomingSignature === this.lastEditorPayloadSignature) return

      this.syncingFromParent = true
      const cookies = safeParseJson(value.show_cookies, [])
      const filters = String(value.filter_uris || '').split('\n').map(item => item.trim()).filter(Boolean)

      this.cookieItems = Array.isArray(cookies) ? cookies.map(item => ({
        uid: createUid('cookie'),
        find_type: item.find_type || 'cookie',
        format_list_text: Array.isArray(item.format_list) ? item.format_list.join(', ') : '',
        find_key: item.find_key || '',
        regex_find_key: item.regex_find_key || '',
        label: item.label || '',
        domain_list_text: Array.isArray(item.Domain_list) ? item.Domain_list.join(', ') : '',
      })) : []
      this.filterItems = filters.map(item => ({ uid: createUid('filter'), value: item }))

      this.lastEditorPayloadSignature = this.createPayloadSignature(this.buildEditorPayload())
      this.$nextTick(() => { this.syncingFromParent = false })
    },
    emitChange() {
      if (this.syncingFromParent) return
      const editorPayload = this.buildEditorPayload()
      const nextSignature = this.createPayloadSignature(editorPayload)
      if (nextSignature === this.lastEditorPayloadSignature) return
      this.lastEditorPayloadSignature = nextSignature
      this.$emit('update:modelValue', { ...this.modelValue, ...editorPayload })
    },
    addCookieItem() { this.cookieItems.push(createCookieItem()) },
    removeCookieItem(index) { this.cookieItems.splice(index, 1) },
    addFilterItem() { this.filterItems.push(createFilterItem()) },
    removeFilterItem(index) { this.filterItems.splice(index, 1) },
  },
}
</script>

<style scoped src="@/css/components/smart_link/LinkConfigEditor.css"></style>
