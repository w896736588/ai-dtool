<template>
  <div class="layout-container">
    <!-- 左侧菜单 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <span class="logo">🛠️</span>
        <span class="title">DevTools</span>
      </div>
      
      <el-menu
        :default-active="menuName"
        active-text-color="#3a7a3a"
        background-color="#f5f5f0"
        text-color="#5a5a5a"
        router
        class="sidebar-menu"
        @select="handleSelect"
      >
        <el-menu-item index="/Dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>首页</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('redis')" index="/Redis">
          <el-icon><Coin /></el-icon>
          <span>Redis</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('supervisor')" index="/Supervisor">
          <el-icon><Setting /></el-icon>
          <span>Supervisor</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('git')" index="/Git">
          <el-icon><Folder /></el-icon>
          <span>Git</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('login')" index="/Link">
          <el-icon><Link /></el-icon>
          <span>自定义网页</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('variable')" index="/Variable">
          <el-icon><Document /></el-icon>
          <span>自定义脚本</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('memory_fragment')" index="/MemoryFragment">
          <el-icon><Memo /></el-icon>
          <span>记忆片段</span>
        </el-menu-item>
        <!-- <el-menu-item v-if="checkModuleOpen('info_crawl')" index="/InfoCrawl">
          <el-icon><Connection /></el-icon>
          <span>信息抓取</span>
        </el-menu-item> -->
        <el-menu-item v-if="checkModuleOpen('docker')" index="/Docker">
          <el-icon><Box /></el-icon>
          <span>Docker</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('api')" index="/Api">
          <el-icon><Connection /></el-icon>
          <span>接口开发</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('shellout')" index="/shellout">
          <el-icon><Monitor /></el-icon>
          <span>终端输出</span>
        </el-menu-item>
        <el-menu-item index="/Set">
          <el-icon><Tools /></el-icon>
          <span>配置</span>
        </el-menu-item>
      </el-menu>

      <!-- 底部工具栏 -->
      <div class="sidebar-footer">
        <el-tag v-if="ip" size="small" type="info" @click="copyIp()" style="cursor: pointer; margin-bottom: 8px;">
          {{ ip }}
        </el-tag>
        <div class="footer-buttons">
          <el-tag size="small" style="cursor: pointer;" @click="OpenNewBlank()">
            新页卡
          </el-tag>
          <el-tag size="small" style="cursor: pointer;" @click="drawerVisibleTools = true">
            小工具
          </el-tag>
          <el-tag size="small" style="cursor: pointer;" @click="openSshConnectionsDialog">
            当前SSH连接数 {{ sshConnectionCount }}
          </el-tag>
        </div>
        <el-button v-if="loginInfo.dialog" size="small" @click="loginInfo.dialog = true">登录</el-button>
      </div>
    </aside>

    <!-- 主内容区域 -->
    <main class="main-content">
      <router-view v-slot="{ Component, route }" name="home">
        <keep-alive>
          <component :is="Component" ref="currentRef"/>
        </keep-alive>
      </router-view>
    </main>
  </div>

  <el-drawer
    v-model="drawerVisibleTools"
    direction="rtl"
    size="90%"
    title="小工具"
  >
    <tools></tools>
  </el-drawer>

  <el-dialog v-model="loginInfo.dialog" title="登录" width="500">
    <el-form>
      <el-form-item :label-width="80" label="username">
        <el-input v-model="loginInfo.username" autocomplete="off"/>
      </el-form-item>
      <el-form-item :label-width="80" label="password">
        <el-input v-model="loginInfo.password" autocomplete="off" show-password/>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="loginInfo.dialog = false">取消</el-button>
        <el-button type="primary" @click="login">保存</el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog
    v-model="sshConnectionsDialogVisible"
    title="当前SSH连接列表"
    width="82%"
    top="6vh"
    class="ssh-connections-dialog"
  >
    <div class="ssh-dialog-toolbar">
      <el-tag type="success" effect="light">连接数 {{ sshConnectionCount }}</el-tag>
      <el-button size="small" type="primary" plain @click="refreshSshConnections(true)">刷新列表</el-button>
    </div>
    <el-table
      v-loading="sshConnectionsLoading"
      :data="sshConnections"
      stripe
      border
      style="width: 100%"
      class="ssh-connections-table"
      max-height="62vh"
      empty-text="暂无活跃SSH连接"
    >
      <el-table-column prop="ssh_name" label="SSH" width="180" />
      <el-table-column prop="shell_client_id" label="客户端ID" width="220" />
      <el-table-column prop="current_command" label="当前命令" min-width="320" show-overflow-tooltip />
      <el-table-column prop="status" label="状态" width="120">
        <template #default="scope">
          <el-tag
            size="small"
            effect="light"
            :type="scope.row.status === 'busy' ? 'warning' : 'success'"
          >
            {{ scope.row.status || '-' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="connect_time" label="连接开始时间" width="180" />
      <el-table-column prop="connect_seconds" label="连接时长(秒)" width="120" />
      <el-table-column prop="type" label="类型" width="120" />
    </el-table>
  </el-dialog>
</template>

<script>
import base from '../utils/base'
import mod from '../utils/module'
import socket from '../utils/base/socket'
import store from "@/utils/base/store";
import format from "@/utils/base/format"
import Clipboard from "clipboard";
import copy from "@/utils/base/copy"
import module from "@/utils/module"
import baseApi from '@/utils/base/base_api'
import sshSet from '@/utils/base/ssh_set'
import Tools from "@/components/Tools.vue";
import Markdown from '@/components/Markdown.vue'
import { 
  HomeFilled,
  Coin,
  Setting,
  Folder,
  Link,
  Document,
  Memo,
  Box,
  Connection,
  Monitor,
  Tools as ToolsIcon
} from "@element-plus/icons-vue";

export default {
  data() {
    return {
      drawerVisibleTools: false,
      drawerVisibleMarkdown: false,
      loginInfo: {
        dialog: false,
        username: 'default',
        password: '111',
      },
      menuKeyStore: 'lastMenuName.v2',
      menuName: '/Dashboard',
      minHeightMap: {},
      showShellMap: ['/Git', '/Consumer', '/WechatKefu'],
      tags: [],
      showTextarea: true,
      shellShowResult: "",
      sshMapList: [],
      xtermMapList: [],
      runCommand: '',
      openModuleList: [],
      term: '',
      lastShellInfo: {
        sshId: "",
        business: "",
        lastShellInfo: "",
      },
      ip: '',
      sshConnectionCount: 0,
      sshConnections: [],
      sshConnectionsDialogVisible: false,
      sshConnectionsLoading: false,
      sshConnectionTimer: null,
    }
  },
  created() {
    window.handleCopy = copy.handleCopy;
  },
  mounted: function () {
    let _that = this
    _that.openModuleList = module.GetOpenModuleList()
    base.BaseLogin(_that.loginInfo.username, _that.loginInfo.password, function (response) {
      if (response.ErrCode === 0) {
        store.setStore('token', response.Data.token)
      } else {
        _that.$helperNotify.error('登录失败')
      }
    })
    this.forceIp(false)
    this.refreshSshConnections(false)
    this.sshConnectionTimer = setInterval(() => {
      this.refreshSshConnections(false)
    }, 5000)
    this.menuName = this.$helperStore.getStore(this.menuKeyStore)
    if (this.$route.path !== this.menuName && this.menuName != null) {
      this.$router.push(this.menuName)
    }
    window.addEventListener('resize', function () {});
  },
  provide() {
    return {
      showTerminal: this.showTerminal,
      resizeTerminal: this.resizeTerminal,
    };
  },
  methods: {
    OpenNewBlank: function () {
      window.open(window.location.href, '_blank');
    },
    copyIp: function () {
      let index = copy.SetCopyContent(this.ip)
      copy.handleCopy(index)
    },
    forceIp: function (forceIp) {
      let _that = this
      baseApi.Ip({}, function (ip) {
        _that.ip = ip
      }, forceIp)
    },
    login: function () {
      let _that = this
      base.BaseLogin(_that.loginInfo.username, _that.loginInfo.password, function (response) {
        if (response.ErrCode === 0) {
          store.setStore('token', response.Data.token)
          window.location.reload()
        } else {
          _that.$helperNotify.error('登录失败')
        }
      })
    },
    checkModuleOpen: function (moduleName) {
      return this.openModuleList.includes(moduleName)
    },
    resetConn: function () {
      store.removeStore('Unikey')
    },
    showNotify: function (notifyList) {
      this.tags = notifyList
    },
    showTerminal(uniqueKey) {
      this.lastShellInfo.uniqueKey = uniqueKey
      this.shellSetShowResult(uniqueKey)
      this.shellDrawerScrollTop(2000)
    },
    resizeTerminal: function () {},
    shellSetShowResult: function (uniqueKey) {
      for (let i in this.sshMapList) {
        if (this.sshMapList[i].uniqueKey === uniqueKey) {
          this.shellShowResult = this.sshMapList[i].shellResult
        }
      }
    },
    shellDrawerScrollTop: function (milliseconds) {
      setTimeout(function () {
        let obj = document.getElementById('showShellResult')
        if (obj) {
          obj.scrollTop = obj.scrollHeight + 200
        }
      }, milliseconds)
    },
    openSshConnectionsDialog() {
      this.sshConnectionsDialogVisible = true
      this.refreshSshConnections(true)
    },
    refreshSshConnections(showLoading) {
      if (showLoading) {
        this.sshConnectionsLoading = true
      }
      const _that = this
      sshSet.SshList(function (sshResponse) {
        const sshNameMap = {}
        if (sshResponse && sshResponse.ErrCode === 0 && Array.isArray(sshResponse.Data)) {
          sshResponse.Data.forEach(item => {
            sshNameMap[String(item.id)] = item.name || `#${item.id}`
          })
        }
        sshSet.GetConnections(function (connResponse) {
          if (_that.sshConnectionsLoading) {
            _that.sshConnectionsLoading = false
          }
          if (!(connResponse && connResponse.ErrCode === 0)) {
            _that.sshConnectionCount = 0
            _that.sshConnections = []
            return
          }
          const list = Array.isArray(connResponse.Data?.connections) ? connResponse.Data.connections : []
          const normalized = list.map(conn => {
            const shellClientId = String(conn.shell_client_id || '')
            const sshId = shellClientId.split('#')[0]
            return {
              ...conn,
              ssh_name: sshNameMap[sshId] || `#${sshId || '-'}`
            }
          })
          _that.sshConnectionCount = normalized.length
          _that.sshConnections = normalized
        })
      })
    },
    handleSelect(key, keyPath) {
      let _that = this
      if (keyPath[0].indexOf('Doc-') >= 0) {
        return
      }
      if (keyPath[0].indexOf('Ignore-') >= 0) {
        return;
      }
      this.menuName = keyPath[0]
      this.$helperStore.setStore(_that.menuKeyStore, this.menuName)
    },
  },
  beforeUnmount() {
    if (this.sshConnectionTimer) {
      clearInterval(this.sshConnectionTimer)
      this.sshConnectionTimer = null
    }
  },
  components: {
    HomeFilled,
    Coin,
    Setting,
    Folder,
    Link,
    Document,
    Memo,
    Box,
    Connection,
    Monitor,
    ToolsIcon,
    Markdown,
    Tools,
    Clipboard,
  },
}
</script>

<style scoped>
.layout-container {
  display: flex;
  height: 100vh;
  width: 100%;
  background-color: #f8f8f5;
}

.sidebar {
  width: 140px;
  background-color: #f5f5f0;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  border-right: 1px solid #e8e8e0;
}

.sidebar-header {
  height: 50px;
  display: flex;
  align-items: center;
  padding: 0 12px;
  border-bottom: 1px solid #e8e8e0;
}

.logo {
  font-size: 20px;
  margin-right: 6px;
}

.title {
  color: #4a4a4a;
  font-size: 16px;
  font-weight: 600;
}

.sidebar-menu {
  flex: 1;
  border-right: none;
  overflow-y: auto;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 140px;
}

.sidebar-footer {
  padding: 10px;
  border-top: 1px solid #e8e8e0;
  display: flex;
  flex-direction: column;
  align-items: stretch;
}

.footer-buttons {
  display: flex;
  flex-direction: column;
  width: 100%;
  gap: 6px;
  margin-bottom: 8px;
}

.footer-buttons .el-tag {
  width: 100%;
  box-sizing: border-box;
  justify-content: center;
  white-space: normal;
  text-align: center;
  line-height: 1.2;
  padding: 4px 6px;
}

.main-content {
  flex: 1;
  overflow: auto;
  background-color: #fafaf7;
  height: 100%;
  padding: 20px;
  box-sizing: border-box;
}

/* 覆盖 Element Plus 菜单样式 */
.sidebar-menu {
  padding: 6px 0;
}

.sidebar-menu .el-menu-item {
  height: 40px;
  line-height: 40px;
  margin: 2px 6px;
  border-radius: 6px;
  padding-left: 12px !important;
}

.sidebar-menu .el-menu-item:hover {
  background-color: #e8f5e8 !important;
  border-radius: 6px;
}

.sidebar-menu .el-menu-item.is-active {
  background-color: #dcedc8 !important;
  border-radius: 6px;
  color: #3a7a3a !important;
}

.sidebar-menu .el-menu-item .el-icon {
  margin-right: 8px;
  font-size: 16px;
}

.sidebar-menu .el-menu-item span {
  font-size: 13px;
  font-weight: 500;
}

.ssh-dialog-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 12px;
  padding: 10px 12px;
  border: 1px solid #e8eee4;
  border-radius: 10px;
  background: linear-gradient(135deg, #f6fbf4 0%, #ffffff 100%);
}

.ssh-connections-table :deep(.el-table__header-wrapper th) {
  background: #f5faf4;
  color: #44584a;
  font-weight: 600;
}

.ssh-connections-table :deep(.el-table__row td) {
  padding-top: 9px;
  padding-bottom: 9px;
}

@media (max-width: 900px) {
  .ssh-dialog-toolbar {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
