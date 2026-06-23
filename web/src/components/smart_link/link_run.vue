<template>
  <el-alert v-if="is_install === 1" :closable="false" show-icon title="正在安装中，看网速大约5-20分钟" type="warning"/>
  <el-alert
      v-if="node_install_tip.show"
      :closable="false"
      show-icon
      type="error"
      style="margin-bottom: 8px;"
  >
    <template #title>未检测到 Node.js，当前无法使用自定义网页</template>
    <div>{{ node_install_tip.install_tip }}</div>
    <el-link :href="node_install_tip.install_url" target="_blank" type="primary" style="margin-top: 4px;">
      前往下载 Node.js
    </el-link>
  </el-alert>
  <div class="link-run-page">
    <div class="link-run-header-card">
      <div class="link-run-header-title">
        <div class="link-run-header-title__main">
          自定义网页工作台
          <span class="link-run-migrate-tip" @click="migrateOldData">
            <template v-if="migrating">迁移中...</template>
            <template v-else>迁移老数据</template>
          </span>
        </div>
        <div class="link-run-header-title__desc">集中管理页面入口、运行方式和流程跳转，顶部操作区独立展示更利于快速切换。</div>
      </div>
      <div class="link-run-toolbar">
        <el-tag size="small" type="info" effect="light">已打开 Page {{ openPageNum }}</el-tag>
        <GitActionButton variant="warning" @click="openAccountSettings">
          <el-icon><User /></el-icon>账号设置
        </GitActionButton>
        <GitActionButton @click="showCreateDialog">
          <el-icon><Plus /></el-icon>创建
        </GitActionButton>
        <GitActionButton @click="install">
          <el-icon><Tools /></el-icon>安装核心
        </GitActionButton>
        <GitActionButton variant="warning" @click="recycle">
          <el-icon><Refresh /></el-icon>释放内存
        </GitActionButton>
        <GitActionButton variant="info" @click="downloadPath">
          <el-icon><Download /></el-icon>下载目录
        </GitActionButton>
        <GitActionButton variant="info" @click="openDataDir">
          <el-icon><FolderOpened /></el-icon>数据存储
        </GitActionButton>
        <GitActionButton variant="info" @click="drawerVisibleMarkdown = true">
          <el-icon><QuestionFilled /></el-icon>帮助文档
        </GitActionButton>
      </div>
    </div>
    <div class="link-run-content">
      <div v-for="group in groupedSmartList" :key="group.groupId" class="link-run-card">
        <div class="link-group-header">
          <span class="link-group-name">{{ group.groupName }}</span>
          <span class="link-group-count">{{ group.items.length }} 个链接</span>
        </div>
        <div class="link-run-links-row">
          <div v-for="link in group.items" :key="link.id" class="link-grid-item">
            <div class="link-grid-item__row link-grid-item__row--top">
              <a class="link-grid-item__label" @click="showEditDialog(link)" :title="link.label">{{ link.label || '未命名' }}</a>
              <div class="link-grid-item__top-right">
                <span v-if="link.runNum" class="link-grid-item__run-num">运行中: {{ link.runNum }}</span>
                <el-icon size="14" class="link-grid-item__edit-icon" @click="showEditDialog(link)"><Edit/></el-icon>
                <el-popconfirm cancel-button-text="取消" confirm-button-text="删除" title="确定删除吗?" @confirm="deleteSmartLinkItem(link)">
                  <template #reference>
                    <el-icon size="14" style="cursor: pointer; color: #999;"><Delete/></el-icon>
                  </template>
                </el-popconfirm>
              </div>
            </div>
            <div class="link-grid-item__row link-grid-item__row--bottom">
              <!-- 账号列表 -->
              <template v-if="link.userList && link.userList.length > 0">
                <el-select v-model="link.chooseUserName" placeholder="选择账号" size="small" class="link-account-select">
                  <el-option v-for="(user, uk) in link.userList" :key="uk" :label="user.user_name" :value="user.user_name"/>
                </el-select>
              </template>

              <!-- 执行操作 -->
              <div class="link-grid-item__exec">
                <GitActionButton v-if="parseInt(link.open_type) === 1 && parseInt(link.open_num) === 0" compact size="small" @click="redirectLink(link)">
                  打开
                </GitActionButton>
                <template v-if="parseInt(link.open_type) === 2 || parseInt(link.open_type) === 3">
<el-select v-if="parseInt(link.open_type) === 2" v-model="link.open_type_new" size="small" style="width: 200px">
                    <el-option v-for="opt in openTypeList" :key="opt.value" :label="opt.label" :value="opt.value"/>
                  </el-select>
                  <el-input v-if="link.open_num > 0" v-model="link.open_num_new" size="small" placeholder="次" style="width: 38px"/>
                  <GitActionButton compact size="small" @click="smartLinkRunItem(link)">执行</GitActionButton>
                </template>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div v-if="groupedSmartList.length === 0" class="link-run-card" style="text-align:center;color:#999;padding:30px;">
        暂无链接，请先"迁移老数据"或点击"创建"新增
      </div>
    </div>
  </div>

  <!-- 新增/编辑弹窗 / Create/Edit dialog -->
  <el-dialog v-model="dialogSmartLink" title="创建/编辑链接" width="90%" class="smart-link-dialog">
    <el-form label-width="auto" class="smart-link-dialog__form">
      <el-form-item label="展示名称(label)">
        <el-input v-model="smartLinkConfig.label" placeholder="例如 生产环境"/>
      </el-form-item>
      <el-form-item label="跳转地址(link)">
        <el-input v-model="smartLinkConfig.link" placeholder="https://example.com"/>
      </el-form-item>
      <el-form-item label="分组">
        <el-select v-model="smartLinkConfig.smart_link_group_id" clearable filterable placeholder="选择分组" style="width: 100%">
          <el-option v-for="g in groupOptions" :key="g.id" :label="g.name" :value="g.id"/>
        </el-select>
      </el-form-item>
      <el-row :gutter="12">
        <el-col :span="12">
          <el-form-item label="浏览器认证用户名">
            <el-input v-model="smartLinkConfig.browser_auth_username" placeholder="可选"/>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="浏览器认证密码">
            <el-input v-model="smartLinkConfig.browser_auth_password" placeholder="可选" show-password/>
          </el-form-item>
        </el-col>
      </el-row>
      <el-row :gutter="12">
        <el-col :span="12">
          <el-form-item label="账号列表">
            <el-select v-model="accountGroupName" clearable filterable placeholder="请选择账号分组" style="width: 100%">
              <el-option v-for="group in accountGroupOptions" :key="group.id" :label="group.name" :value="group.name"/>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="Cookie">
            <el-input v-model="smartLinkConfig.cookie" placeholder="可选"/>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="请求头(JSON)">
        <el-input v-model="smartLinkConfig.headers" type="textarea" :rows="4" placeholder='可选，例如 {"Authorization":"Bearer xxx"}'/>
      </el-form-item>
      <el-form-item label="类型">
        <el-select v-model="smartLinkConfig.open_type" placeholder="选择类型">
          <el-option v-for="opt in openTypeList" :key="opt.value" :label="opt.label" :value="opt.value"/>
        </el-select>
      </el-form-item>
      <el-form-item v-if="parseInt(smartLinkConfig.open_type) !== 1" label="浏览器">
        <el-select v-model="smartLinkConfig.channel" placeholder="选择类型">
          <el-option v-for="opt in channelList" :key="opt.value" :label="opt.label" :value="opt.value"/>
        </el-select>
      </el-form-item>
      <el-form-item label="自动关闭(秒)">
        <el-input v-model="smartLinkConfig.auto_close_second" placeholder="0表示无限"/>
      </el-form-item>
      <el-form-item label="打开次数">
        <el-input v-model="smartLinkConfig.open_num" placeholder="0表示不需要自定义"/>
      </el-form-item>
      <el-form-item label="执行逻辑">
        <el-select v-model="smartLinkConfig.process_id" clearable placeholder="选择执行逻辑">
          <el-option v-for="proc in processList" :key="proc.id" :label="proc.name" :value="proc.id"/>
        </el-select>
      </el-form-item>
      <el-form-item v-if="dialogSmartLink" label="信息提取" class="smart-link-dialog__link-config">
        <LinkConfigEditor v-model="smartLinkConfig" />
      </el-form-item>
      <el-form-item label="排序值">
        <el-input v-model="smartLinkConfig.weight" type="text"/>
      </el-form-item>
    </el-form>
    <template class="dialog-footer">
      <GitActionButton @click="dialogSmartLink = false">取 消</GitActionButton>
      <GitActionButton @click="saveSmartLink">确 定</GitActionButton>
    </template>
  </el-dialog>

  <shellResult ref="shellRef" :btnName="'运行日志'" :isRunning="shellController.isRunning" :shellShowResult="shellController.sshResult" :show-model="shellController.showModel"></shellResult>

  <el-drawer v-model="drawerVisibleMarkdown" direction="rtl" size="90%" title="文档">
    <Markdown v-if="drawerVisibleMarkdown" :markdownType="markdownType"></Markdown>
  </el-drawer>

  <SettingsDialog v-model="accountSettingsVisible" title="账号设置" width="82%" @closed="refreshLinkAfterAccountSettingsClose">
    <AccountSettingPage @changed="handleAccountSettingsChanged" />
  </SettingsDialog>
</template>
<style scoped src="@/css/components/smart_link/link_run.css"></style>
<script>
import smart_link_set from "@/utils/base/smart_link_set"
import ticker_step from "@/utils/base/ticker_step"
import Markdown from "@/components/Markdown.vue";
import Process from '@/utils/base/smart_link_proces'
import shellResult from "@/components/shell/result_button.vue";
import sse from "@/utils/base/sse";
import sseDistribute from "@/utils/base/sse_distribute";
import LinkConfigEditor from "@/components/smart_link/LinkConfigEditor.vue";
import GitActionButton from "@/components/base/GitActionButton.vue";
import SettingsDialog from '@/components/base/SettingsDialog.vue'
import AccountSettingPage from '@/components/set/account.vue'
import accountSet from '@/utils/base/account_set'
import { Plus, Tools, Refresh, Download, QuestionFilled, Delete, User, FolderOpened, Edit } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

export default {
  components: {
    shellResult,
    Markdown,
    Plus,
    Tools,
    Refresh,
    Download,
    QuestionFilled,
    Delete,
    User,
    FolderOpened,
    Edit,
    LinkConfigEditor,
    GitActionButton,
    SettingsDialog,
    AccountSettingPage,
  },
  data() {
    return {
      shellController: {
        sshResult: '',
        sourceSshResult: '',
        isRunning: false,
        showModel: 'button',
        divHeight: 330,
      },
      drawerVisibleMarkdown: false,
      markdownType: 'Link',
      dialogSmartLink: false,
      openTypeList: [
        {label: '通过js直接打开', value: 1},
        {label: '静默打开(内置核心打开)', value: 2},
        {label: '浏览器打开(内置核心打开)', value: 3}
      ],
      channelList: [
        {label: '请选择', value: ''},
        {label: 'chrome(完整浏览器功能)', value: 'chrome'},
        {label: 'chromium(内存占用低)', value: 'chromium'},
      ],
      sse_distribute_id: '',
      processList: [],
      groupOptions: [],
      accountGroupOptions: [],
      accountGroupName: '',
      smartLinkConfig: {
        id: 0, label: '', link: '', smart_link_group_id: 0,
        account_list: '', browser_auth_username: '', browser_auth_password: '',
        cookie: '', headers: '', open_num: 0, open_type: '', channel: '',
        process_id: 0, download_finds: '', auto_close_second: 0, weight: 0,
        show_cookies: '', filter_uris: '', combine_type: 4,
      },
      defaultSmartLinkConfig: {
        id: 0, label: '', link: '', smart_link_group_id: 0,
        account_list: '', browser_auth_username: '', browser_auth_password: '',
        cookie: '', headers: '', open_num: 0, open_type: '', channel: '',
        process_id: 0, download_finds: '', auto_close_second: 0, weight: 0,
        show_cookies: '', filter_uris: '', combine_type: 4,
      },
      name: 'Link',
      smartList: [],
      smartLinkRunList: {},
      tickerKey: 'link',
      versionInfo: {},
      openPageNum: 0,
      is_install: 0,
      node_install_tip: {
        show: false,
        install_url: 'https://nodejs.org/zh-cn/download',
        install_tip: '请先安装 Node.js（建议 LTS 版本），安装完成后刷新当前页面。',
      },
      accountSettingsVisible: false,
      migrating: false,
    }
  },
  computed: {
    // groupedSmartList 将 smartList 按 smart_link_group_id 分组展示
    groupedSmartList() {
      const groups = {}
      // 先按分组收集
      for (let i = 0; i < this.smartList.length; i++) {
        const item = this.smartList[i]
        const gid = Number(item.smart_link_group_id || 0)
        if (!groups[gid]) {
          const groupInfo = this.groupOptions.find(g => Number(g.id) === gid)
          groups[gid] = {
            groupId: gid,
            groupName: groupInfo ? groupInfo.name : '未分组',
            items: [],
          }
        }
        groups[gid].items.push(item)
      }
      // 转为数组：有名字的分组在前，未分组(smart_link_group_id=0)在后
      const result = []
      for (const gid of Object.keys(groups)) {
        if (Number(gid) !== 0) result.push(groups[gid])
      }
      // 未分组(smart_link_group_id=0)的链接追加在最后
      if (groups[0]) {
        result.push(groups[0])
      }
      return result
    },
  },
  mounted: function () {
    this.sse_distribute_id = sseDistribute.GetSseDistributeId('link')
    this.sseCreate()
    this.init()
  },
  activated() {
    this.init()
    this.refreshRuntimeConfigState()
  },
  methods: {
    sseCreate: function () {
      let _that = this
      sseDistribute.RegisterReceive(_that.sse_distribute_id, function (msg) {
        if (msg === sse.SseEventClean) {
          _that.shellController.sshResult = ''
          _that.shellController.sourceSshResult = ''
        } else {
          _that.shellController.sourceSshResult += msg
          _that.shellController.sshResult = _that.shellController.sourceSshResult
        }
      })
    },
    init: function () {
      this.GetProcessList()
      this.GetConfigList()
      this.loadAccountGroupOptions()
      this.tickerRunList()
      this.SmartLinkChromeVersion()
    },
    loadAccountGroupOptions() {
      accountSet.AccountGroupList((response) => {
        if (response && response.ErrCode === 0 && Array.isArray(response.Data)) {
          this.accountGroupOptions = response.Data
        }
      })
    },
    // migrateOldData 点击触发老数据迁移
    migrateOldData: function () {
      if (this.migrating) return
      this.migrating = true
      smart_link_set.SmartLinkMigrateOldData((response) => {
        this.migrating = false
        if (response && response.ErrCode === 0) {
          const data = response.Data || {}
          const parts = [`${data.group_count || 0} 个分组`]
          if (data.process_fixed_count > 0) parts.push(`${data.process_fixed_count} 条执行逻辑已修复`)
          if (data.group_fixed_count > 0) parts.push(`${data.group_fixed_count} 条分组已修复`)
          parts.push(`共 ${data.total_links || 0} 条链接`)
          parts.push(`新增 ${data.migrated_count || 0} 条`)
          if (data.skipped_count > 0) parts.push(`跳过 ${data.skipped_count} 条`)
          if (data.failed_count > 0) parts.push(`失败 ${data.failed_count} 条`)
          ElMessage.success(`迁移完成：${parts.join('，')}`)
          this.GetConfigList()
        } else {
          ElMessage.error((response && response.ErrMsg) || '迁移失败')
        }
      })
    },
    openAccountSettings: function () {
      this.accountSettingsVisible = true
    },
    handleAccountSettingsChanged: function () {
      this.GetConfigList()
    },
    refreshLinkAfterAccountSettingsClose: function () {
      this.GetConfigList()
    },
    applyNodeInstallTip: function (response) {
      let data = response && response.Data ? response.Data : {}
      let needInstall = data.need_install_node === 1
      this.node_install_tip.show = needInstall
      if (needInstall) {
        this.node_install_tip.install_url = data.install_url || 'https://nodejs.org/zh-cn/download'
        this.node_install_tip.install_tip = data.install_tip || '请先安装 Node.js（建议 LTS 版本），安装完成后刷新当前页面。'
      }
      return needInstall
    },
    SmartLinkChromeVersion: function () {
      smart_link_set.SmartLinkChromeVersion(this.sse_distribute_id, (response) => {
        if (response.ErrCode === 0) {
          this.versionInfo = response.Data.version
          this.is_install = response.Data.is_install
          this.applyNodeInstallTip(response)
        } else {
          if (!this.applyNodeInstallTip(response)) {
            ElMessage.error('获取版本失败')
          }
        }
      })
    },
    // smartLinkRunItem 执行某个链接 / Execute a single link
    smartLinkRunItem: function (item) {
      let _that = this
      if (!item) return
      let chooseUser = {}
      if (item.userList && item.userList.length > 0 && item.chooseUserName) {
        for (let i in item.userList) {
          if (item.userList[i].user_name === item.chooseUserName) {
            chooseUser = item.userList[i]
            break
          }
        }
      }
      let runParams = {
        id: item.id,
        label: item.label,
        user_name: chooseUser.user_name || '',
        password: chooseUser.password || '',
        open_num: item.open_num_new || item.open_num || 0,
        open_type: item.open_type_new || item.open_type,
        sse_distribute_id: _that.sse_distribute_id,
      }
      smart_link_set.SmartLinkRun(runParams, (response) => {
        if (response.ErrCode !== 0) {
          if (!_that.applyNodeInstallTip(response)) {
            ElMessage.error(response.ErrMsg || '执行失败')
          }
          return
        }
        ticker_step.Active(_that.tickerKey)
      })
    },
    tickerRunList: function () {
      ticker_step.Register(this.tickerKey, 5, () => { this.runList() })
    },
    runList: function () {
      smart_link_set.SmartLinkRunList(this.sse_distribute_id, (response) => {
        if (response.ErrCode !== 0) {
          this.applyNodeInstallTip(response)
          return
        }
        let runList = response.Data
        this.openPageNum = 0
        this.smartLinkRunList = {}
        for (let i in runList) {
          if (this.smartLinkRunList[runList[i].name]) {
            this.smartLinkRunList[runList[i].name] += runList[i].page_num
          } else {
            this.smartLinkRunList[runList[i].name] = runList[i].page_num
          }
          this.openPageNum += runList[i].page_num
        }
        // 为每个链接分配运行数（兼容包含用户名的 LinkIdLabel）
        for (let i in this.smartList) {
          let item = this.smartList[i]
          let runNamePrefix = "link_id_" + item.id + "_label_" + item.label
          item.runNum = 0
          for (let runName in this.smartLinkRunList) {
            if (runName.startsWith(runNamePrefix)) {
              item.runNum += this.smartLinkRunList[runName]
            }
          }
        }
      })
    },
    // formatAccountList 将账号组名格式化为后端协议格式 / Convert account group name to backend format
    formatAccountList: function (groupName) {
      const name = String(groupName || '').trim()
      return name ? `{group:account:${name}}` : ''
    },
    // parseAccountGroupName 从后端协议格式解析账号组名 / Parse account group name from backend format
    parseAccountGroupName: function (accountListValue) {
      const raw = String(accountListValue || '').trim()
      const matched = raw.match(/^\{group:account:(.+)\}$/)
      return matched ? matched[1] : ''
    },
    saveSmartLink: function () {
      let _that = this
      // 构建 account_list 字段
      _that.smartLinkConfig.account_list = _that.formatAccountList(_that.accountGroupName)
      _that.smartLinkConfig.combine_type = 4
      smart_link_set.SmartLinkItemAdd(_that.smartLinkConfig, function (response) {
        if (response.ErrCode === 0) {
          _that.dialogSmartLink = false
          _that.GetConfigList()
        } else {
          ElMessage.error('保存失败：' + (response.ErrMsg || ''))
        }
        ticker_step.Active(_that.tickerKey)
      })
    },
    showEditDialog: function (item) {
      this.smartLinkConfig = JSON.parse(JSON.stringify(item))
      this.accountGroupName = this.parseAccountGroupName(item.account_list || '')
      this.dialogSmartLink = true
    },
    deleteSmartLinkItem: function (item) {
      smart_link_set.SmartLinkItemDelete({ id: item.id }, (response) => {
        if (response.ErrCode === 0) {
          this.GetConfigList()
        } else {
          ElMessage.error('删除失败')
        }
      })
    },
    showCreateDialog: function () {
      this.smartLinkConfig = JSON.parse(JSON.stringify(this.defaultSmartLinkConfig))
      this.accountGroupName = ''
      this.dialogSmartLink = true
    },
    GetConfigList: function () {
      smart_link_set.SmartLinkItemList((response) => {
        if (response.ErrCode === 0) {
          // 在赋值给 this.smartList 前初始化 chooseUserName，确保 Vue 2 能追踪属性变化
          const list = response.Data.smart_link_list || []
          for (let item of list) {
            item.open_num_new = item.open_num || 0
            item.open_type_new = item.open_type || 2
            item.runNum = 0
            if (Array.isArray(item.userList) && item.userList.length > 0 && !item.chooseUserName) {
              item.chooseUserName = item.userList[0].user_name
            }
          }
          // 排序：按 weight 升序
          list.sort((a, b) => (Number(a.weight) || 0) - (Number(b.weight) || 0))
          this.smartList = list
          this.groupOptions = response.Data.group_list || []
        } else {
          ElMessage.error('获取列表失败')
        }
      })
    },
    GetProcessList: function () {
      Process.SmartProcessList((response) => {
        if (response.ErrCode === 0) {
          this.processList = (response.Data && response.Data.list) ? response.Data.list : []
        }
      })
    },

    redirectLink: function (item) {
      window.open(item.link, '_blank')
      ticker_step.Active(this.tickerKey)
    },
    changeToProcess: function () { this.$emit('changeModelToEditProcess') },
    changeToFlow: function () { this.$emit('changeModelToFlow') },
    downloadPath: function () {
      smart_link_set.SmartLinkDownloadPath(this.sse_distribute_id, (response) => {
        if (response.ErrCode !== 0) {
          if (!this.applyNodeInstallTip(response)) { ElMessage.error('失败') }
        }
      })
    },
    openDataDir: function () {
      smart_link_set.SmartLinkOpenDataDir((response) => {
        if (response.ErrCode !== 0) { ElMessage.error(response.ErrMsg || '打开失败') }
      })
    },
    install: function () {
      smart_link_set.SmartLinkChromeUpdate(this.sse_distribute_id, (response) => {
        if (response.ErrCode === 0) {
          this.GetConfigList(); this.runList()
        } else {
          if (!this.applyNodeInstallTip(response)) { ElMessage.error('失败') }
        }
      })
    },
    recycle: function () {
      smart_link_set.SmartLinkRecycle(this.sse_distribute_id, (response) => {
        if (response.ErrCode === 0) {
          this.GetConfigList(); this.runList()
        } else {
          if (!this.applyNodeInstallTip(response)) { ElMessage.error('失败') }
        }
      })
    },
    refreshRuntimeConfigState: function () {},
  },
}
</script>

<style scoped>
.link-run-page { min-height: calc(100vh - 110px); color: #4a4a4a; }
.link-run-header-card { background: #fff; border: 1px solid #e8e8e0; border-radius: 12px; padding: 12px 16px; margin-bottom: 10px; }
.link-run-header-title { margin-bottom: 8px; }
.link-run-header-title__main { color: #4a4a4a; font-size: 18px; font-weight: 600; }
.link-run-migrate-tip { font-size: 12px; font-weight: 400; color: #409EFF; cursor: pointer; margin-left: 12px; text-decoration: underline; }
.link-run-header-title__desc { margin-top: 6px; color: #74806f; font-size: 13px; line-height: 1.6; }
.link-run-content { padding: 0 2px 2px; }
.link-run-toolbar { display: flex; align-items: center; flex-wrap: wrap; gap: 8px; }
/* 分组卡片 */
.link-run-card { padding: 10px 10px 8px; margin-bottom: 10px; background: #fff; border: 1px solid #e6e8de; border-radius: 10px; }
/* 分组头 */
.link-group-header { display: flex; align-items: baseline; gap: 10px; margin-bottom: 10px; }
.link-group-name { font-size: 15px; font-weight: bold; cursor: pointer; text-decoration: underline; color: #333; }
.link-group-count { font-size: 11px; color: #999; }
/* 链接行 - 弹性布局，根据内容自动调整宽度 */
.link-run-links-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 0;
}
/* 网格链接项 - 宽度由内容决定 */
.link-grid-item {
  padding: 8px 10px;
  border: 1px solid #e8ece0;
  border-radius: 6px;
  background: #fafbf7;
  display: inline-flex;
  flex-direction: column;
  gap: 6px;
  flex: 0 0 auto;
  min-width: 200px;
}
.link-grid-item__row { display: flex; align-items: center; gap: 6px; flex-wrap: nowrap; }
.link-grid-item__row--top { justify-content: space-between; }
.link-grid-item__label { font-size: 13px; font-weight: 600; cursor: pointer; text-decoration: underline; color: #333; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 180px; }
.link-grid-item__top-right { display: flex; align-items: center; gap: 6px; margin-left: auto; flex-shrink: 0; }
.link-grid-item__edit-icon { cursor: pointer; color: #999; flex-shrink: 0; }
.link-grid-item__edit-icon:hover { color: #409EFF; }
.link-grid-item__run-num { font-size: 10px; color: green; white-space: nowrap; }
.link-account-select { width: 140px; }
.link-grid-item__exec { display: flex; align-items: center; gap: 4px; flex-wrap: nowrap; }
.smart-link-dialog :deep(.el-dialog__body) { padding-top: 18px; }
.smart-link-dialog__form { width: 100%; }
.smart-link-dialog__link-config { width: 100%; }
.smart-link-dialog__link-config :deep(.el-form-item__content) { width: 100%; display: block; }
</style>
