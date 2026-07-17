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
        <GitActionButton variant="info" @click="showGroupManage">
          <el-icon><Management /></el-icon>分组管理
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
          <span class="link-group-name" @click="clickGroupName(group)">
            <span class="link-group-id">#{{ group.groupId }}</span>{{ group.groupName }}
          </span>
          <span class="link-group-count">{{ group.items.length }} 个链接</span>
          <el-button size="small" type="primary" link :loading="copyingGroups[group.groupId]" :disabled="copyingGroups[group.groupId]" @click="copyGroup(group)">复制</el-button>
          <el-popconfirm
            :title="group.items.length > 0 ? `分组「${group.groupName}」下有 ${group.items.length} 个链接，删除后这些链接将变为未分组，确定删除该分组？` : `确定删除分组「${group.groupName}」？`"
            @confirm="deleteGroup({ id: group.groupId, name: group.groupName })"
          >
            <template #reference>
              <el-button size="small" type="danger" link :disabled="copyingGroups[group.groupId]">删除</el-button>
            </template>
          </el-popconfirm>
        </div>
        <div v-if="group.items.length > 0" class="link-run-links-row">
          <div v-for="link in group.items" :key="link.id" class="link-grid-item">
            <div class="link-grid-item__row link-grid-item__row--top">
              <a class="link-grid-item__label" @click="showEditDialog(link)" :title="link.label">
                <span class="link-grid-item__id">#{{ link.id }}</span><span class="link-grid-item__name">{{ link.label || '未命名' }}</span>
              </a>
              <div class="link-grid-item__top-right">
                <span v-if="link.runNum" class="link-grid-item__run-num">运行中: {{ link.runNum }}</span>
                <el-dropdown trigger="click" size="small">
                  <el-icon size="14" class="link-grid-item__more-icon"><MoreFilled/></el-icon>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item @click="showEditDialog(link)">编辑</el-dropdown-item>
                      <el-dropdown-item @click="copyLink(link)">复制</el-dropdown-item>
                      <el-dropdown-item @click="confirmDeleteLink(link)">删除</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </div>
            <div class="link-grid-item__row link-grid-item__row--bottom" :class="{ 'link-grid-item__row--no-account': !link.userList || link.userList.length === 0 }">
              <!-- 账号列表 -->
              <template v-if="link.userList && link.userList.length > 0">
                <el-select v-model="link.chooseUserName" placeholder="选择账号" size="small" class="link-account-select">
                  <el-option v-for="(user, uk) in link.userList" :key="uk" :label="user.user_name" :value="user.user_name"/>
                </el-select>
              </template>

              <!-- 执行操作 -->
              <div class="link-grid-item__exec">
                <GitActionButton v-if="parseInt(link.open_type) === 1 && parseInt(link.open_num) === 0" :compact="link.userList && link.userList.length > 0" :size="link.userList && link.userList.length > 0 ? 'small' : 'default'" @click="redirectLink(link)">
                  打开
                </GitActionButton>
                <template v-if="parseInt(link.open_type) === 2 || parseInt(link.open_type) === 3">
<el-select v-if="parseInt(link.open_type) === 2" v-model="link.open_type_new" size="small" style="width: 200px">
                    <el-option v-for="opt in openTypeList" :key="opt.value" :label="opt.label" :value="opt.value"/>
                  </el-select>
                  <el-input v-if="link.open_num > 0" v-model="link.open_num_new" size="small" placeholder="次" style="width: 38px"/>
                  <GitActionButton :compact="link.userList && link.userList.length > 0" :size="link.userList && link.userList.length > 0 ? 'small' : 'default'" :loading="!!runningItems[link.id]" @click="smartLinkRunItem(link)">执行</GitActionButton>
                </template>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="link-group-empty">该分组暂无链接，点击上方「创建」可添加</div>
      </div>
      <div v-if="groupedSmartList.length === 0" class="link-run-card" style="text-align:center;color:#999;padding:30px;">
        暂无分组，请先"迁移老数据"或在「分组管理」中创建分组
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

  <!-- 分组管理弹窗 / Group management dialog -->
  <el-dialog v-model="dialogGroupManage" title="分组管理" width="600px">
    <div style="margin-bottom: 10px;">
      <el-button type="primary" size="small" @click="addGroup">添加分组</el-button>
    </div>
    <el-table :data="allGroups" size="small">
      <el-table-column prop="id" label="#ID" width="60" />
      <el-table-column prop="name" label="组名" min-width="200" />
      <el-table-column label="操作" width="120">
        <template #default="scope">
          <el-button type="primary" link size="small" @click="editGroup(scope.row)">编辑</el-button>
          <el-popconfirm title="确定删除该分组吗?" @confirm="deleteGroup(scope.row)">
            <template #reference>
              <el-button type="danger" link size="small">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <!-- 分组添加/编辑子弹窗 -->
    <el-dialog v-model="dialogEditGroup" :title="editGroupConfig.id ? '编辑分组' : '添加分组'" width="400px" append-to-body>
      <el-form label-width="60px">
        <el-form-item label="组名">
          <el-input v-model="editGroupConfig.name" autocomplete="off" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogEditGroup = false">取消</el-button>
        <el-button type="primary" @click="saveGroup">保存</el-button>
      </template>
    </el-dialog>
  </el-dialog>

  <!-- 编辑分组名弹窗（点击分组名触发）/ Edit group name dialog -->
  <el-dialog v-model="dialogEditGroupName" title="编辑分组" width="400px">
    <el-form label-width="60px">
      <el-form-item label="组名">
        <el-input v-model="editingGroupName" autocomplete="off" @keyup.enter="saveGroupName" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogEditGroupName = false">取消</el-button>
      <el-button type="primary" @click="saveGroupName">保存</el-button>
    </template>
  </el-dialog>
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
import { Plus, Tools, Refresh, Download, QuestionFilled, MoreFilled, User, FolderOpened, Management } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export default {
  components: {
    shellResult,
    Markdown,
    Plus,
    Tools,
    Refresh,
    Download,
    QuestionFilled,
    MoreFilled,
    User,
    FolderOpened,
    Management,
    LinkConfigEditor,
    GitActionButton,
    SettingsDialog,
    AccountSettingPage,
  },
  emits: ['changeModelToEditProcess', 'changeModelToFlow'],
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
      runningItems: {},
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
      copyingGroups: {},
      // 分组管理 / Group management
      dialogGroupManage: false,
      allGroups: [],
      dialogEditGroup: false,
      editGroupConfig: {},
      // 编辑分组名（点击分组名触发）/ Edit group name
      dialogEditGroupName: false,
      editingGroupName: '',
      editingGroupId: 0,
    }
  },
  computed: {
    // groupedSmartList 将 smartList 按 smart_link_group_id 分组展示
    // 注意：以完整的 groupOptions 为基础，确保没有链接的空分组也会展示
    groupedSmartList() {
      const groups = {}
      // 先建立所有分组（含没有链接的空分组）
      for (const g of (this.groupOptions || [])) {
        const gid = Number(g.id || 0)
        groups[gid] = {
          groupId: gid,
          groupName: g.name || '未命名分组',
          items: [],
        }
      }
      // 再把链接归入对应分组（分组不存在时兜底为「未分组」）
      for (let i = 0; i < this.smartList.length; i++) {
        const item = this.smartList[i]
        const gid = Number(item.smart_link_group_id || 0)
        if (!groups[gid]) {
          groups[gid] = {
            groupId: gid,
            groupName: '未分组',
            items: [],
          }
        }
        groups[gid].items.push(item)
      }
      // 保持 groupOptions 的顺序（空分组也排在其应有位置），末尾追加「未分组」
      const result = []
      for (const g of (this.groupOptions || [])) {
        const gid = Number(g.id || 0)
        if (groups[gid]) {
          result.push(groups[gid])
          delete groups[gid]
        }
      }
      // 剩余未匹配的分组（如「未分组」id=0）追加到最后
      for (const gid of Object.keys(groups)) {
        result.push(groups[gid])
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
        } else if (String(msg).indexOf('[SMART_LINK_RUN_DONE]') !== -1) {
          // 忽略执行完成标记，不展示到日志框
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
      _that.runningItems[item.id] = true
      smart_link_set.SmartLinkRun(runParams, (response) => {
        _that.runningItems[item.id] = false
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
    confirmDeleteLink: function (item) {
      ElMessageBox.confirm('确定删除该链接吗？', '提示', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        this.deleteSmartLinkItem(item)
      }).catch(() => {})
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
    // copyLink 以该链接配置为模板新建一条相同配置的新链接 / Duplicate a link as a new one
    copyLink: function (item) {
      const _that = this
      const newItem = {
        label: (item.label || '未命名') + ' - 副本',
        link: item.link || '',
        smart_link_group_id: item.smart_link_group_id || 0,
        account_list: item.account_list || '',
        browser_auth_username: item.browser_auth_username || '',
        browser_auth_password: item.browser_auth_password || '',
        cookie: item.cookie || '',
        headers: item.headers || '',
        open_num: item.open_num || 0,
        open_type: item.open_type || '',
        channel: item.channel || '',
        process_id: item.process_id || 0,
        download_finds: item.download_finds || '',
        auto_close_second: item.auto_close_second || 0,
        weight: item.weight || 0,
        show_cookies: item.show_cookies || '',
        filter_uris: item.filter_uris || '',
        combine_type: item.combine_type || 4,
      }
      smart_link_set.SmartLinkItemAdd(newItem, function (response) {
        if (response.ErrCode === 0) {
          _that.GetConfigList()
          ElMessage.success('已复制并新建链接')
        } else {
          ElMessage.error('复制失败：' + (response.ErrMsg || ''))
        }
      })
    },
    // copyGroup 复制整个分组及其下的所有链接到新分组 / Copy group with all links to a new group
    copyGroup: function (group) {
      const _that = this
      const newName = group.groupName + ' - 副本'
      // 标记 loading 防止重复点击
      _that.copyingGroups[group.groupId] = true
      // Step 1: 创建新分组 / Create new group
      smart_link_set.SetSmartLinkGroupAdd({ name: newName }, function (addGroupResponse) {
        if (addGroupResponse.ErrCode !== 0) {
          ElMessage.error('创建分组副本失败')
          _that.copyingGroups[group.groupId] = false
          return
        }
        // Step 2: 刷新分组列表获取新分组 ID / Refresh group list to get new group ID
        smart_link_set.SetSmartLinkGroupList(function (listResponse) {
          if (listResponse.ErrCode !== 0) {
            ElMessage.error('获取分组列表失败')
            _that.copyingGroups[group.groupId] = false
            return
          }
          const groups = listResponse.Data || []
          const newGroup = groups.find(function (g) { return g.name === newName })
          if (!newGroup) {
            ElMessage.error('未找到新建的分组')
            _that.copyingGroups[group.groupId] = false
            return
          }
          const newGroupId = newGroup.id
          const items = group.items || []
          if (items.length === 0) {
            // 没有链接直接刷新列表
            _that.GetConfigList()
            _that.copyingGroups[group.groupId] = false
            ElMessage.success('分组已复制（无链接）')
            return
          }
          // Step 3: 逐个复制链接 / Copy each link
          var copied = 0
          var failed = 0
          var total = items.length
          function checkDone() {
            copied++
            if (copied + failed === total) {
              _that.GetConfigList()
              _that.copyingGroups[group.groupId] = false
              if (failed > 0) {
                ElMessage.warning('复制完成：成功 ' + (total - failed) + ' 个，失败 ' + failed + ' 个')
              } else {
                ElMessage.success('成功复制 ' + total + ' 个链接到分组"' + newName + '"')
              }
            }
          }
          for (var i = 0; i < items.length; i++) {
            var src = items[i]
            var newItem = {
              label: src.label || '',
              link: src.link || '',
              smart_link_group_id: newGroupId,
              account_list: src.account_list || '',
              browser_auth_username: src.browser_auth_username || '',
              browser_auth_password: src.browser_auth_password || '',
              cookie: src.cookie || '',
              headers: src.headers || '',
              open_num: src.open_num || 0,
              open_type: src.open_type || '',
              channel: src.channel || '',
              process_id: src.process_id || 0,
              download_finds: src.download_finds || '',
              auto_close_second: src.auto_close_second || 0,
              weight: src.weight || 0,
              show_cookies: src.show_cookies || '',
              filter_uris: src.filter_uris || '',
              combine_type: src.combine_type || 4,
            }
            smart_link_set.SmartLinkItemAdd(newItem, function (addResponse) {
              if (addResponse.ErrCode !== 0) {
                failed++
              }
              checkDone()
            })
          }
        })
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
    // showGroupManage 打开分组管理弹窗 / Open group management dialog
    showGroupManage: function () {
      this.dialogGroupManage = true
      this.loadGroupList()
    },
    // loadGroupList 加载所有分组 / Load all groups
    loadGroupList: function () {
      const _that = this
      smart_link_set.SetSmartLinkGroupList(function (response) {
        if (response.ErrCode === 0) {
          _that.allGroups = response.Data || []
        } else {
          ElMessage.error('获取分组列表失败')
        }
      })
    },
    // addGroup 打开添加分组弹窗 / Open add group dialog
    addGroup: function () {
      this.editGroupConfig = { name: '' }
      this.dialogEditGroup = true
    },
    // editGroup 打开编辑分组弹窗 / Open edit group dialog
    editGroup: function (row) {
      this.editGroupConfig = { id: row.id, name: row.name }
      this.dialogEditGroup = true
    },
    // saveGroup 保存分组（添加或编辑）/ Save group (add or edit)
    saveGroup: function () {
      const _that = this
      var config = {
        name: this.editGroupConfig.name || '',
      }
      if (this.editGroupConfig.id) {
        config.id = this.editGroupConfig.id
      }
      smart_link_set.SetSmartLinkGroupAdd(config, function (response) {
        if (response.ErrCode === 0) {
          _that.dialogEditGroup = false
          _that.loadGroupList()
          _that.GetConfigList()
        } else {
          ElMessage.error('保存分组失败')
        }
      })
    },
    // deleteGroup 删除分组 / Delete group
    deleteGroup: function (row) {
      const _that = this
      smart_link_set.SetSmartLinkGroupDelete(row, function (response) {
        if (response.ErrCode === 0) {
          _that.loadGroupList()
          _that.GetConfigList()
        } else {
          ElMessage.error('删除失败')
        }
      })
    },
    // clickGroupName 点击分组名，打开编辑弹窗 / Click group name to edit
    clickGroupName: function (group) {
      this.editingGroupId = group.groupId
      this.editingGroupName = group.groupName
      this.dialogEditGroupName = true
    },
    // saveGroupName 保存分组名修改 / Save group name change
    saveGroupName: function () {
      const _that = this
      if (!this.editingGroupName || !this.editingGroupName.trim()) {
        ElMessage.warning('组名不能为空')
        return
      }
      smart_link_set.SetSmartLinkGroupAdd(
        { id: this.editingGroupId, name: this.editingGroupName.trim() },
        function (response) {
          if (response.ErrCode === 0) {
            _that.dialogEditGroupName = false
            _that.GetConfigList()
            ElMessage.success('分组名已更新')
          } else {
            ElMessage.error('修改失败')
          }
        }
      )
    },
    refreshRuntimeConfigState: function () {},
  },
}
</script>

<style scoped>
.link-run-page {
  --link-primary: #2f9e6b;
  --link-primary-hover: #268a5c;
  --link-primary-soft: #e8f5ee;
  --link-primary-soft-2: #f1f9f4;
  --link-accent: #5ace99;
  --link-border: #e4ece7;
  --link-bg: #ffffff;
  --link-bg-soft: #f6faf7;
  --link-text: #243029;
  --link-text-2: #6b7d72;
  --link-shadow: 0 1px 2px rgba(31, 41, 55, 0.04), 0 2px 8px rgba(31, 41, 55, 0.05);
  --link-shadow-hover: 0 6px 18px rgba(47, 158, 107, 0.16);
  --link-radius: 12px;
  min-height: calc(100vh - 110px);
  color: var(--link-text);
}

/* 顶部标题卡片 */
.link-run-header-card {
  background: var(--link-bg);
  border: 1px solid var(--link-border);
  border-radius: var(--link-radius);
  padding: 16px 20px;
  margin-bottom: 14px;
  box-shadow: var(--link-shadow);
}
.link-run-header-title { margin-bottom: 14px; }
.link-run-header-title__main {
  display: flex;
  align-items: center;
  color: var(--link-text);
  font-size: 19px;
  font-weight: 700;
  letter-spacing: .3px;
}
.link-run-header-title__main::before {
  content: "";
  width: 4px;
  height: 18px;
  border-radius: 3px;
  background: var(--link-primary);
  margin-right: 10px;
}
.link-run-header-title__desc {
  margin-top: 8px;
  color: var(--link-text-2);
  font-size: 13px;
  line-height: 1.6;
  max-width: 780px;
}
.link-run-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
}

/* 主按钮统一为绿色（仅作用于本页，避免影响运行逻辑/流程视图） */
.link-run-page :deep(.git-action-button--primary),
.link-run-page :deep(.pl-button--primary) {
  --git-button-text-color: #ffffff;
  --git-button-border-color: var(--link-primary);
  --git-button-background-color: var(--link-primary);
  --git-button-hover-text-color: #ffffff;
  --git-button-hover-border-color: var(--link-primary-hover);
  --git-button-hover-background-color: var(--link-primary-hover);
  --pl-button-text-color: #ffffff;
  --pl-button-border-color: var(--link-primary);
  --pl-button-background-color: var(--link-primary);
  --pl-button-hover-text-color: #ffffff;
  --pl-button-hover-border-color: var(--link-primary-hover);
  --pl-button-hover-background-color: var(--link-primary-hover);
  color: #ffffff !important;
}

/* 分组卡片 */
.link-run-card {
  padding: 14px 16px 16px;
  margin-bottom: 14px;
  background: var(--link-bg);
  border: 1px solid var(--link-border);
  border-radius: var(--link-radius);
  box-shadow: var(--link-shadow);
}
/* 分组头 */
.link-group-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}
.link-group-name {
  font-size: 16px;
  font-weight: 700;
  cursor: pointer;
  color: var(--link-text);
  display: inline-flex;
  align-items: center;
  gap: 8px;
  transition: color .15s ease;
}
.link-group-name::before {
  content: "";
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--link-primary);
  flex-shrink: 0;
}
.link-group-name:hover { color: var(--link-primary); }
.link-group-id {
  flex-shrink: 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 11px;
  font-weight: 700;
  color: var(--link-primary);
  background: var(--link-primary-soft);
  border: 1px solid var(--link-border);
  padding: 1px 7px;
  border-radius: 6px;
}
.link-group-count {
  font-size: 12px;
  color: var(--link-text-2);
  background: var(--link-bg-soft);
  border: 1px solid var(--link-border);
  padding: 2px 10px;
  border-radius: 999px;
}

/* 链接网格 */
.link-run-links-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin: 0;
}
/* 链接卡片 */
.link-grid-item {
  position: relative;
  flex: 1 1 230px;
  max-width: 340px;
  min-width: 0;
  padding: 12px 14px;
  border: 1px solid var(--link-border);
  border-left: 3px solid var(--link-accent);
  border-radius: 10px;
  background: var(--link-bg);
  display: flex;
  flex-direction: column;
  gap: 10px;
  box-shadow: var(--link-shadow);
  transition: transform .15s ease, box-shadow .15s ease, border-color .15s ease;
}
.link-grid-item:hover {
  transform: translateY(-2px);
  box-shadow: var(--link-shadow-hover);
  border-color: var(--link-primary);
}
.link-grid-item__row { display: flex; align-items: center; gap: 8px; flex-wrap: nowrap; }
.link-grid-item__row--top { justify-content: space-between; }
.link-grid-item__row--bottom { gap: 8px; }
.link-grid-item__row--no-account .link-grid-item__exec {
  width: 100%;
  justify-content: center;
}
.link-grid-item__label {
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  color: var(--link-text);
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  text-decoration: none;
}
.link-grid-item__id {
  flex-shrink: 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 11px;
  font-weight: 700;
  color: var(--link-primary);
  background: var(--link-primary-soft);
  border: 1px solid var(--link-primary-soft);
  padding: 0 6px;
  border-radius: 5px;
  line-height: 18px;
}
.link-grid-item__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}
.link-grid-item__label::before {
  content: "";
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  background-color: var(--link-primary);
  -webkit-mask: url("data:image/svg+xml,%3Csvg%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%2024%2024'%20fill='none'%20stroke='%23000'%20stroke-width='2'%20stroke-linecap='round'%20stroke-linejoin='round'%3E%3Cpath%20d='M10%2013a5%205%200%200%200%207.07%207.07l3-3a5%205%200%200%200-7.07-7.07l-1.72%201.71'/%3E%3Cpath%20d='M14%2011a5%205%200%200%200-7.07-7.07l-3%203a5%205%200%200%200%207.07%207.07l1.71-1.71'/%3E%3C/svg%3E") center / contain no-repeat;
  mask: url("data:image/svg+xml,%3Csvg%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%2024%2024'%20fill='none'%20stroke='%23000'%20stroke-width='2'%20stroke-linecap='round'%20stroke-linejoin='round'%3E%3Cpath%20d='M10%2013a5%205%200%200%200%207.07%207.07l3-3a5%205%200%200%200-7.07-7.07l-1.72%201.71'/%3E%3Cpath%20d='M14%2011a5%205%200%200%200-7.07-7.07l-3%203a5%205%200%200%200%207.07%207.07l1.71-1.71'/%3E%3C/svg%3E") center / contain no-repeat;
}
.link-grid-item__label:hover { color: var(--link-primary); }
.link-grid-item__top-right { display: flex; align-items: center; gap: 6px; margin-left: auto; flex-shrink: 0; }
.link-grid-item__more-icon { cursor: pointer; color: var(--link-text-2); flex-shrink: 0; transition: color .15s ease; }
.link-grid-item__more-icon:hover { color: var(--link-primary); }
.link-grid-item__run-num {
  font-size: 11px;
  font-weight: 600;
  color: var(--link-primary);
  white-space: nowrap;
  background: var(--link-primary-soft);
  border: 1px solid var(--link-primary-soft);
  padding: 1px 8px;
  border-radius: 999px;
}
.link-account-select { flex: 1 1 auto; min-width: 0; }
/* 账号下拉框：默认无边框无背景，悬浮/聚焦才显示绿色边框 */
.link-account-select :deep(.el-select__wrapper) {
  box-shadow: 0 0 0 1px transparent;
  background-color: transparent;
  transition: box-shadow .15s ease, background-color .15s ease;
}
.link-account-select:hover :deep(.el-select__wrapper),
.link-account-select :deep(.el-select__wrapper.is-focused) {
  box-shadow: 0 0 0 1px var(--link-primary);
  background-color: var(--link-bg);
}
/* 兼容旧版 el-input 外层（部分场景兜底） */
.link-account-select :deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px transparent !important;
  background-color: transparent !important;
}
.link-account-select:hover :deep(.el-input__wrapper),
.link-account-select :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--link-primary) !important;
  background-color: var(--link-bg) !important;
}
.link-group-empty {
  font-size: 13px;
  color: var(--link-text-2);
  background: var(--link-bg-soft);
  border: 1px dashed var(--link-border);
  border-radius: 8px;
  padding: 18px 14px;
  text-align: center;
}
.link-grid-item__exec { display: flex; align-items: center; gap: 6px; flex-wrap: nowrap; flex: 0 0 auto; }

.smart-link-dialog :deep(.el-dialog__body) { padding-top: 18px; }
.smart-link-dialog__form { width: 100%; }
.smart-link-dialog__link-config { width: 100%; }
.smart-link-dialog__link-config :deep(.el-form-item__content) { width: 100%; display: block; }
</style>
