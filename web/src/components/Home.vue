<template>
  <el-row :gutter="10" style="background-color: #545c64;margin-left:0px;margin-right:0px;">
    <el-col :span="20">
      <el-menu :default-active="menuName" :ellipsis="false" active-text-color="#ffd04b"
               background-color="#545c64" mode="horizontal" router size="15"
               text-color="#fff" @select="handleSelect">
        <el-menu-item v-if="checkModuleOpen('redis')" index="/Redis">Redis</el-menu-item>
        <el-menu-item v-if="checkModuleOpen('supervisor')" index="/Supervisor">Supervisor</el-menu-item>
        <el-menu-item v-if="checkModuleOpen('git')" index="/Git">Git</el-menu-item>
        <el-menu-item v-if="checkModuleOpen('login')" index="/Link">自定义网页</el-menu-item>
        <el-menu-item v-if="checkModuleOpen('variable')" index="/Variable">自定义脚本</el-menu-item>
        <!--        <el-menu-item v-if="checkModuleOpen('tools')" index="/Tools">小工具</el-menu-item>-->
        <el-menu-item v-if="checkModuleOpen('docker')" index="/Docker">Docker</el-menu-item>
<!--        <el-menu-item v-if="checkModuleOpen('markdown')" index="/Markdown">Markdown</el-menu-item>-->
        <el-menu-item v-if="checkModuleOpen('api')" index="/Api">接口开发</el-menu-item>
        <el-menu-item v-if="checkModuleOpen('shellout')" index="/shellout">
          终端输出
        </el-menu-item>
        <el-menu-item index="/Set">配置</el-menu-item>
      </el-menu>
    </el-col>
    <el-col :span="4" style="display: flex; align-items: center; justify-content: flex-end; padding-right: 20px;">
      <el-tag v-if="ip" style="margin-right: 10px;color:black;" type="info" @click="copyIp()">
        <i class="el-icon-link"></i> {{ ip }}
      </el-tag>
      <el-tag style="margin-right: 10px;color:black;cursor:pointer;" @click="OpenNewBlank()">
        新页卡
      </el-tag>
      <el-tag style="margin-right: 10px;color:black;cursor:pointer;" @click="drawerVisibleTools = true">
        小工具
      </el-tag>
      <el-button v-if="loginInfo.dialog" size="small" @click="loginInfo.dialog = true">登录</el-button>
    </el-col>
  </el-row>

  <el-main id="routerV" style="min-height: calc(100vh - 100px); padding: 20px;">
    <router-view v-slot="{ Component,route }" name="home">
      <keep-alive>
        <component :is="Component" ref="currentRef"/>
      </keep-alive>
    </router-view>
  </el-main>

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
import Tools from "@/components/Tools.vue";
import Markdown from '@/components/Markdown.vue'
import {Burger} from "@element-plus/icons-vue";

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
      menuKeyStore: 'lastMenuName.v1',
      menuName: '/Redis',
      minHeightMap: {
        //内容框最小高度配置
      },
      showShellMap: [
        //哪些菜单展示终端
        '/Git',
        '/Consumer',
        '/WechatKefu'
      ],
      tags: [],
      showTextarea: true,
      shellShowResult: "",
      //存储socket链接和结果
      sshMapList: [],
      xtermMapList: [],
      runCommand: '',
      openModuleList: [],
      //Xterm
      term: '',
      lastShellInfo: {
        sshId: "",
        business: "",
        lastShellInfo: "",

      },
      ip: '',
    }
  },
  created() {
    window.handleCopy = copy.handleCopy;
  },
  mounted: function () {
    let _that = this
    //开启的服务
    _that.openModuleList = module.GetOpenModuleList()
    //登录注册
    base.BaseLogin(_that.loginInfo.username, _that.loginInfo.password, function (response) {
      if (response.ErrCode === 0) {
        store.setStore('token', response.Data.token)
      } else {
        _that.$helperNotify.error('登录失败')
      }
    })
    //外网IP
    this.forceIp(false)

    //处理默认打开的页卡
    this.menuName = this.$helperStore.getStore(this.menuKeyStore)
    // if (!this.$helperConfig.getXkfDevSshConfig() || !this.$helperConfig.getWkDevSshConfig() || !this.$helperConfig.getXkfDevDbConfig()) {
    //   this.menuName = '/Set'
    // }
    if (this.$route.path !== this.menuName && this.menuName != null) {
      this.$router.push(this.menuName)
    }
    //监听页面大小变化
    window.addEventListener('resize', function () {
    });

  },
  provide() {
    return {
      showTerminal: this.showTerminal,
      resizeTerminal: this.resizeTerminal,
    };
  },
  methods: {
    combineMsg: function () {
      let _that = this
      setTimeout(function () {

      })
    },
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
      // window.location.reload();
    },
    showNotify: function (notifyList) {
      this.tags = notifyList
    },
    //供子组件调用
    showTerminal(uniqueKey) {
      this.lastShellInfo.uniqueKey = uniqueKey
      this.shellSetShowResult(uniqueKey)
      this.shellDrawerScrollTop(2000)
    },
    resizeTerminal: function () {

    },
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
  components: {
    Burger,
    Markdown,
    Tools,
    Clipboard,
  },
}
</script>

<style scoped>

.menu-item-with-badge {
  display: flex;
  align-items: center;
  position: relative;
  gap: 4px;
}

.menu-badge {
  position: absolute;
  top: -28px;
  right: -30px;
}

/* 竖排按钮容器（固定在右侧中间） */
.vertical-button {
  position: fixed;
  right: -0;
  top: 50%;
  transform: translateY(-50%);
  z-index: 1000;
  font-size: 10px;

  /* 背景和边框 */
  background: #409eff !important; /* Element Plus 主色 */
  color: white !important; /* 文字颜色 */
  border: none !important;
  border-radius: 4px 0 0 4px !important; /* 左侧圆角 */
  padding: 16px 8px !important;
  box-shadow: -2px 0 8px rgba(0, 0, 0, 0.1); /* 左侧阴影 */

  /* 文字竖排 */
  letter-spacing: 2px;
}

/* 悬停效果 */
.vertical-button:hover {
  background: #337ecc !important; /* 深一点的蓝色 */
  box-shadow: -2px 0 8px rgba(0, 0, 0, 0.2);
}

/* 移除 Element Plus 按钮的默认样式干扰 */
.vertical-button :deep(.el-button__text) {
  display: inline-block;
}
</style>

