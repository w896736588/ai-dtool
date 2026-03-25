<template>
  <el-form class="link-config-editor" label-position="top">
    <div class="editor-section">
      <div class="editor-section__header">
        <div class="editor-section__title">链接配置</div>
        <GitActionButton compact size="small" @click="openCreateLinkDialog">新增链接</GitActionButton>
      </div>

      <div v-if="linkItems.length === 0" class="editor-empty">暂无链接，请先新增一条。</div>
      <div v-else class="link-list">
        <div v-for="(item, index) in linkItems" :key="item.uid" class="link-list-item">
          <div class="link-list-item__main">
            <div class="link-list-item__title">
              <span class="link-list-item__index">#{{ index + 1 }}</span>
              <span>{{ item.label || '未命名链接' }}</span>
            </div>
            <div class="link-list-item__meta">{{ item.link || '未配置链接地址' }}</div>
            <div class="link-list-item__desc">
              <span>账号分组：{{ item.account_group_name || '未选择' }}</span>
              <span>Cookie：{{ item.cookie ? '已配置' : '未配置' }}</span>
              <span>请求头：{{ hasHeaders(item.headers) ? '已配置' : '未配置' }}</span>
            </div>
          </div>
          <div class="link-list-item__actions">
            <GitActionButton compact size="small" @click="openEditLinkDialog(index)">编辑</GitActionButton>
            <GitActionButton compact size="small" variant="danger" @click="removeLinkItem(index)">删除</GitActionButton>
          </div>
        </div>
      </div>
    </div>

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

    <el-dialog
      v-model="linkItemDialogVisible"
      title="新增/编辑链接"
      width="760px"
      append-to-body
      class="link-item-dialog"
    >
      <el-form label-position="top" class="link-item-form">
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="展示名称">
              <el-input v-model="linkItemDraft.label" placeholder="例如 生产环境" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="跳转地址">
              <el-input v-model="linkItemDraft.link" placeholder="https://example.com 或 /path" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="浏览器认证用户名">
              <el-input v-model="linkItemDraft.browser_auth_username" placeholder="可选" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="浏览器认证密码">
              <el-input v-model="linkItemDraft.browser_auth_password" placeholder="可选" show-password />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="12">
            <el-form-item label="账号列表">
              <el-select
                v-model="linkItemDraft.account_group_name"
                clearable
                filterable
                placeholder="请选择账号分组"
                style="width: 100%"
              >
                <el-option
                  v-for="group in accountGroupOptions"
                  :key="group.id"
                  :label="group.name"
                  :value="group.name"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Cookie">
              <el-input v-model="linkItemDraft.cookie" placeholder="可选" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="请求头(JSON)">
          <el-input
            v-model="linkItemDraft.headers"
            type="textarea"
            :rows="4"
            placeholder='可选，例如 {"Authorization":"Bearer xxx"}'
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <GitActionButton @click="closeLinkDialog">取消</GitActionButton>
        <GitActionButton @click="saveLinkItem">保存</GitActionButton>
      </template>
    </el-dialog>
  </el-form>
</template>

<script>
import accountSet from '@/utils/base/account_set'
import GitActionButton from '@/components/base/GitActionButton.vue'

const createUid = (prefix) => `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`
const createLinkItem = () => ({
  uid: createUid('link'),
  label: '',
  link: '',
  browser_auth_username: '',
  browser_auth_password: '',
  account_list: '',
  account_group_name: '',
  cookie: '',
  headers: '',
})
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
  try {
    return JSON.parse(text)
  } catch (error) {
    return fallback
  }
}

export default {
  name: 'LinkConfigEditor',
  components: {
    GitActionButton,
  },
  props: {
    modelValue: {
      type: Object,
      default: () => ({}),
    },
  },
  emits: ['update:modelValue'],
  data() {
    return {
      accountGroupOptions: [],
      linkItems: [],
      linkItemDialogVisible: false,
      editingLinkIndex: -1,
      linkItemDraft: createLinkItem(),
      cookieItems: [],
      filterItems: [],
      syncingFromParent: false,
      lastEditorPayloadSignature: '',
    }
  },
  mounted() {
    // 加载账号分组选项 / Load account group options for the single-select field.
    this.loadAccountGroupOptions()
  },
  watch: {
    modelValue: {
      deep: true,
      immediate: true,
      handler(value) {
        this.syncFromModel(value || {})
      },
    },
    linkItems: { deep: true, handler() { this.emitChange() } },
    cookieItems: { deep: true, handler() { this.emitChange() } },
    filterItems: { deep: true, handler() { this.emitChange() } },
  },
  methods: {
    hasHeaders(headersValue) {
      const normalizedHeaders = typeof headersValue === 'string' ? headersValue.trim() : JSON.stringify(headersValue || {})
      return normalizedHeaders !== '' && normalizedHeaders !== '{}'
    },
    // 解析旧配置中的账号组占位符 / Parse persisted placeholder format into group name.
    parseAccountGroupName(accountListValue) {
      const rawValue = String(accountListValue || '').trim()
      const matched = rawValue.match(/^\{group:account:(.+)\}$/)
      return matched ? matched[1] : ''
    },
    // 保存时恢复后端协议 / Convert selected group back to backend placeholder syntax.
    formatAccountListValue(groupName) {
      const normalizedGroupName = String(groupName || '').trim()
      return normalizedGroupName ? `{group:account:${normalizedGroupName}}` : ''
    },
    loadAccountGroupOptions() {
      const _that = this
      accountSet.AccountGroupList(function (response) {
        // 只有接口成功时才覆盖选项 / Replace options only when the API call succeeds.
        if (response && response.ErrCode === 0 && Array.isArray(response.Data)) {
          _that.accountGroupOptions = response.Data
        }
      })
    },
    // 复制链接项草稿，避免弹窗内联动列表 / Clone draft data so dialog edits do not mutate the list before save.
    cloneLinkItem(item = {}) {
      return {
        uid: item.uid || createUid('link'),
        label: item.label || '',
        link: item.link || '',
        browser_auth_username: item.browser_auth_username || '',
        browser_auth_password: item.browser_auth_password || '',
        account_list: item.account_list || '',
        account_group_name: item.account_group_name || this.parseAccountGroupName(item.account_list),
        cookie: item.cookie || '',
        headers: typeof item.headers === 'string' ? item.headers : JSON.stringify(item.headers || {}, null, 2),
      }
    },
    // 生成编辑器受控字段快照 / Build normalized payload for loop-safe sync.
    buildEditorPayload() {
      const links = this.linkItems.map(item => ({
        label: item.label,
        link: item.link,
        browser_auth_username: item.browser_auth_username,
        browser_auth_password: item.browser_auth_password,
        account_list: this.formatAccountListValue(item.account_group_name),
        cookie: item.cookie,
        headers: safeParseJson(item.headers, {}),
      }))
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
        links: JSON.stringify(links),
        linkList: links,
        show_cookies: JSON.stringify(showCookies),
        filter_uris: filterUris,
      }
    },
    createPayloadSignature(payload) {
      return JSON.stringify(payload || {})
    },
    syncFromModel(value) {
      const incomingSignature = this.createPayloadSignature({
        links: value.links || '',
        linkList: Array.isArray(value.linkList) ? value.linkList : safeParseJson(value.links, []),
        show_cookies: value.show_cookies || '',
        filter_uris: value.filter_uris || '',
      })
      // 相同内容不重复同步 / Skip sync when parent payload is effectively unchanged.
      if (!this.syncingFromParent && incomingSignature === this.lastEditorPayloadSignature) {
        return
      }

      this.syncingFromParent = true
      const links = Array.isArray(value.linkList) ? value.linkList : safeParseJson(value.links, [])
      const cookies = safeParseJson(value.show_cookies, [])
      const filters = String(value.filter_uris || '').split('\n').map(item => item.trim()).filter(Boolean)

      this.linkItems = links.map(item => this.cloneLinkItem(item))
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
      this.$nextTick(() => {
        this.syncingFromParent = false
      })
    },
    emitChange() {
      if (this.syncingFromParent) return
      const editorPayload = this.buildEditorPayload()
      const nextSignature = this.createPayloadSignature(editorPayload)
      if (nextSignature === this.lastEditorPayloadSignature) {
        return
      }
      this.lastEditorPayloadSignature = nextSignature
      this.$emit('update:modelValue', {
        ...this.modelValue,
        ...editorPayload,
      })
    },
    // 打开新增弹窗 / Open create dialog for a new link item.
    openCreateLinkDialog() {
      this.editingLinkIndex = -1
      this.linkItemDraft = this.cloneLinkItem()
      this.linkItemDialogVisible = true
    },
    // 打开编辑弹窗 / Open edit dialog for an existing link item.
    openEditLinkDialog(index) {
      this.editingLinkIndex = index
      this.linkItemDraft = this.cloneLinkItem(this.linkItems[index])
      this.linkItemDialogVisible = true
    },
    closeLinkDialog() {
      this.linkItemDialogVisible = false
      this.editingLinkIndex = -1
      this.linkItemDraft = this.cloneLinkItem()
    },
    // 保存链接项，新增或覆盖列表项 / Save dialog draft by insert or replace.
    saveLinkItem() {
      const nextItem = this.cloneLinkItem(this.linkItemDraft)
      if (this.editingLinkIndex >= 0) {
        this.linkItems.splice(this.editingLinkIndex, 1, nextItem)
      } else {
        this.linkItems.push(nextItem)
      }
      this.closeLinkDialog()
    },
    removeLinkItem(index) { this.linkItems.splice(index, 1) },
    addCookieItem() { this.cookieItems.push(createCookieItem()) },
    removeCookieItem(index) { this.cookieItems.splice(index, 1) },
    addFilterItem() { this.filterItems.push(createFilterItem()) },
    removeFilterItem(index) { this.filterItems.splice(index, 1) },
  },
}
</script>

<style scoped>
.editor-section { margin-bottom: 20px; padding: 16px; border: 1px solid #e5eadf; border-radius: 12px; background: #fbfcf8; }
.editor-section__header, .editor-card__header { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.editor-section__header { margin-bottom: 12px; }
.editor-section__title, .editor-card__title { font-weight: 600; color: #405640; }
.editor-card { padding: 14px; border: 1px solid #dde6d8; border-radius: 10px; background: #fff; margin-bottom: 14px; }
.editor-empty { color: #7a8776; font-size: 13px; }
.filter-row { display: flex; align-items: center; gap: 12px; margin-bottom: 10px; }

.link-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.link-list-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 14px;
  border: 1px solid #dde6d8;
  border-radius: 10px;
  background: #fff;
}

.link-list-item__main {
  min-width: 0;
  flex: 1;
}

.link-list-item__title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
  font-weight: 600;
  color: #405640;
}

.link-list-item__index {
  color: #789070;
  font-size: 12px;
}

.link-list-item__meta {
  margin-bottom: 8px;
  color: #556351;
  word-break: break-all;
}

.link-list-item__desc {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  color: #7a8776;
  font-size: 12px;
}

.link-list-item__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.link-item-form {
  width: 100%;
}
</style>
