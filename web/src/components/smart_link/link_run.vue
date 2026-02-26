<template>
  <el-alert v-if="is_install === 1" :closable="false" show-icon title="正在安装中，看网速大约5-20分钟" type="warning"/>
  <div>
    <div style="margin-bottom: 10px;display: flex;justify-content: center; /* 水平居中 */">
      <el-text class="mx-1">已打开Page({{ openPageNum }})</el-text>&nbsp;
      <el-button type="primary" @click="showCreateDialog">创建</el-button>
      <el-button type="primary" @click="install">安装核心</el-button>
      <el-button type="primary" @click="recycle">释放内存</el-button>
      <el-button type="primary" @click="downloadPath">下载目录</el-button>
      <el-button type="primary" @click="drawerVisibleMarkdown = true">帮助文档</el-button>
<!--      <el-button type="primary" @click="showDialogRunLog">运行日志({{shellController.sshResult.length}})</el-button>&nbsp;-->
      &nbsp;<el-link type="primary" @click="changeToProcess">切换到编辑执行逻辑</el-link>&nbsp;
      <el-link type="primary" @click="changeToFlow">切换到流程图</el-link>&nbsp;
      <!--      <el-link type="primary" @click="showMarkdown">使用说明</el-link>-->
    </div>
    <div v-for="(smartValue, smartLinkIndex) in smartList" :key="smartLinkIndex" class="box-card" style="min-height: 70px;padding:5px;">
      <a style="display: inline-block;text-decoration: underline;cursor:pointer;font-size:17px;font-weight: bold;" @click="showEditDialog(smartValue)">
        {{ smartValue.id + " " + smartValue.name }}
      </a>
      <el-tooltip content="编辑" placement="top">
        <el-icon size="small" style="margin-left:20px;" @click="showEditDialog(smartValue)">
          <Setting/>
        </el-icon>
      </el-tooltip>
      <el-tooltip content="展示账号密码" placement="top">
        <el-icon size="small" style="margin: 10px;" @click="showUserPasswordList(smartValue)">
          <Notebook/>
        </el-icon>
      </el-tooltip>
      <el-tooltip content="删除" placement="top">
        <el-popconfirm
            cancel-button-text="取消"
            confirm-button-text="删除"
            icon-color="#626AEF"
            title="确定删除吗?"
            @confirm="deleteSmartLink(smartValue)"
        >
          <template #reference>
            <el-icon size="small">
              <Delete/>
            </el-icon>
          </template>
        </el-popconfirm>
      </el-tooltip>
      <el-row :gutter="20" style="margin-top: 15px;margin-bottom:10px;">
        <el-col v-for="(linkValue, linkIndex) in smartValue.linkList" :key="linkIndex" :span="4">
          <div class="grid-content bg-purple">
            <!--            选择后内置核心打开-->
            <template v-if="(linkValue.userList && linkValue.userList.length > 0) || parseInt(smartValue.open_num) > 0">
              <!--              供选择的环境列表-->
              <el-radio v-model="smartValue.chooseSmartLinkIndex" :label="linkValue.label"
                        @change="changeChooseLink(smartLinkIndex , linkIndex)">
                {{ linkValue.label }}

                <span v-if="linkValue.runNum" style="font-size: 12px;color:green;">({{ linkValue.runNum }})</span>
              </el-radio>
            </template>

            <!--            直接打开-->
            <template v-if="!linkValue.userList && parseInt(smartValue.open_type) === 1 && parseInt(smartValue.open_num) === 0">
              <el-link style="padding: 10px;" type="primary" @click="redirectLink(linkValue)">
                {{ linkValue.label }}
              </el-link>
            </template>

            <!--            内置核心打开-->
            <template v-if="(!linkValue.userList || linkValue.userList.length === 0) && (parseInt(smartValue.open_type) === 2 || parseInt(smartValue.open_type) === 3) && parseInt(smartValue.open_num) === 0">
              <el-link style="padding: 10px;" type="primary" @click="smartLinkRun(smartLinkIndex,linkIndex)">
                {{ linkValue.label }}
                <span v-if="linkValue.runNum" style="font-size: 12px;color:green;">
                  ({{ linkValue.runNum }})
                </span>
              </el-link>
            </template>

          </div>
        </el-col>
      </el-row>
      <!--      账号列表-->
      <el-form v-if="smartValue.linkList[smartValue.chooseLinkIndex] &&
        (smartValue.linkList[smartValue.chooseLinkIndex].userList || smartValue.open_num > 0 )" :inline="true" class="demo-form-inline"
               label-width="auto" style="margin:auto;">
        <el-form-item v-if="smartValue.linkList[smartValue.chooseLinkIndex].userList && smartValue.linkList[smartValue.chooseLinkIndex].userList.length > 0" label="账号列表">
          <el-select v-model="smartValue.linkList[smartValue.chooseLinkIndex].chooseUserName" placeholder="选择账号">
            <template v-for="(user,userkey) in smartValue.linkList[smartValue.chooseLinkIndex].userList" :key="userkey">
              <el-option :label="user.user_name" :value="user.user_name"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item v-if="smartValue.open_type === 2" label="打开方式">
          <el-select v-model="smartValue.open_type_new" placeholder="选择类型">
            <template v-for="(value,key) in openTypeList" :key="key">
              <el-option :label="value.label" :value="value.value"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item v-if="smartValue.open_num > 0" label="打开数">
          <el-input v-model="smartValue.open_num_new" placeholder="Please input" style="width: 240px"/>
        </el-form-item>
        <el-form-item>
          <el-button v-if="smartValue.linkList[smartValue.chooseLinkIndex].chooseUserName || smartValue.open_num > 0" type="primary" @click="smartLinkRun(smartLinkIndex,null)">
            执行
          </el-button>
        </el-form-item>
      </el-form>
      <el-divider/>
    </div>
  </div>
  <!--新增弹窗-->
  <el-dialog v-model="dialogSmartLink" title="创建/编辑链接" width="90%">
    <el-form label-width="auto" style="max-width: 90%">
      <el-form-item label="名称">
        <el-input v-model="smartLinkConfig.name"/>
      </el-form-item>
      <el-form-item label="类型">
        <el-select v-model="smartLinkConfig.open_type" placeholder="选择类型">
          <template v-for="(value,key) in openTypeList" :key="key">
            <el-option :label="value.label" :value="value.value"/>
          </template>
        </el-select>
      </el-form-item>
      <el-form-item v-if="parseInt(smartLinkConfig.open_type) !== 1" label="Session类型">
        <el-alert :closable="false" show-icon title="如果选择自动合并，将会自动挑选已打开的浏览器运行，没有可用的浏览器时会重新开启新浏览器;如果选择优先上次，那么优先使用上次的数据目录" type="info"/>
        <el-select v-model="smartLinkConfig.combine_type" placeholder="选择类型">
          <template v-for="(value,key) in CombineList" :key="key">
            <el-option :label="value.label" :value="value.value"/>
          </template>
        </el-select>
      </el-form-item>
      <el-form-item v-if="parseInt(smartLinkConfig.combine_type) !== 3" label="浏览器">
        <el-alert :closable="false" show-icon title="如果选择chrome，那么支持播放视频（其他区别还没发现）但是耗费更多内存，会卡一些" type="info"/>
        <el-select v-model="smartLinkConfig.channel" placeholder="选择类型">
          <template v-for="(value,key) in channelList" :key="key">
            <el-option :label="value.label" :value="value.value"/>
          </template>
        </el-select>
      </el-form-item>
      <el-form-item label="自动关闭">
        <el-alert :closable="false" show-icon title="多少秒之内没有操作，页面自动关闭，0表示无限" type="info"/>
        <el-input v-model="smartLinkConfig.auto_close_second" type="text"></el-input>
      </el-form-item>
      <el-form-item label="打开次数">
        <el-alert :closable="false" show-icon title="如果只需要打开一个，那么设置为0，否则设置为1，运行时会出现输入框自定义" type="info"/>
        <el-input v-model="smartLinkConfig.open_num" type="text"></el-input>
      </el-form-item>
      <!--      <el-form-item label="下载匹配项">-->
      <!--        <el-alert title="哪些请求路由会被定义为下载，英文逗号分割" type="info" show-icon :closable="false"/>-->
      <!--        <el-input v-model="smartLinkConfig.download_finds" type="textarea" :rows="5"/>-->
      <!--      </el-form-item>-->
      <el-form-item label="信息提取">
        <el-input v-model="smartLinkConfig.show_cookies" :rows="5" type="textarea"/>
      </el-form-item>
      <el-form-item label="请求拦截（半匹配）">
        <el-input v-model="smartLinkConfig.filter_uris" :rows="5" type="textarea"/>
      </el-form-item>
      <el-form-item label="执行逻辑">
        <el-alert :closable="false" show-icon title="打开链接后执行的流程，切换到编辑执行逻辑页面，可查看执行逻辑" type="info"/>
        <el-select v-model="smartLinkConfig.process_id" placeholder="选择执行逻辑">
          <template v-for="(value,key) in processList" :key="key">
            <el-option :label="value.name" :value="value.id"/>
          </template>
        </el-select>
      </el-form-item>
      <el-form-item v-if="dialogSmartLink" label="链接配置">
        <el-alert :closable="false" :title="linksTip" show-icon type="info"/>
        <JsonEditCombine
            v-if="dialogSmartLink"
            :value="smartLinkConfig.links"
            default-show-type="form"
            mode="tree"
            style="width: 100%;"
            @change="linksChange"
        />
      </el-form-item>
      <el-form-item label="排序值">
        <el-input v-model="smartLinkConfig.weight" type="text"/>
      </el-form-item>
    </el-form>
    <template class="dialog-footer">
      <el-button @click="dialogSmartLink = false">取 消</el-button>
      <el-button type="primary" @click="saveSmartLink">确 定</el-button>
    </template>
  </el-dialog>

  <shellResult ref="shellRef" :btnName="'运行日志'" :isRunning="shellController.isRunning" :shellShowResult="shellController.sshResult" :show-model="shellController.showModel"></shellResult>
  <el-dialog v-model="dialogShowUserPass" title="账号密码列表" width="90%">
    <el-table
        :data="showUserPassList"
        border
        highlight-current-row
        stripe
        style="width: 100%"
    >
      <el-table-column
          label="环境"
          prop="label"
      ></el-table-column>
      <el-table-column
          label="用户名"
          prop="username"
      ></el-table-column>
      <el-table-column
          label="密码"
          prop="password"
      ></el-table-column>
    </el-table>
  </el-dialog>

  <el-drawer
      v-model="drawerVisibleMarkdown"
      direction="rtl"
      size="90%"
      title="文档"
  >
    <Markdown v-if="drawerVisibleMarkdown" :markdownType="markdownType"></Markdown>
  </el-drawer>
</template>
<style>

.demo-form-inline .el-input {
  --el-input-width: 220px;
}


.demo-form-inline {
  display: flex;
  justify-content: center; /* 水平居中 */
}

.el-alert {
  margin: 3px;
}

.demo-form-inline .el-select {
  --el-select-width: 220px;
}
</style>
<script>
import smart_link_set from "@/utils/base/smart_link_set"
import base from "@/utils/base";
import ticker_step from "@/utils/base/ticker_step"
import JsonEditCombine from "@/components/base/json_edit_combine.vue";
import t from "@/utils/base/type"
import Markdown from "@/components/Markdown.vue";
import Init from "@/utils/base/set_init";
import Process from '@/utils/base/smart_link_proces'
import shellResult from "@/components/shell/result_button.vue";
import sse from "@/utils/base/sse";
import sseDistribute from "@/utils/base/sse_distribute";
import shell from "@/utils/base/shell"

export default {
  props: {
    shellShowResult: {
      type: String
    },
  },
  components: {
    shellResult,
    Markdown,
    JsonEditCombine,
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
      dialogShowMarkdown: false,
      dialogShowUserPass: false,
      dialogSsePushLog: false,
      showUserPassList: [],
      dialogSmartLink: false,
      linksTip: '[{"link":"","label":"","userList":[{"user_name":"","password":""}],"browser_auth_username":"","browser_auth_password":""}]',
      // processTip : '[{"type":"","not_exist_Locator":"","Locator":"","uri":"","tip":"","value":""}]',
      openTypeList: [
        {label: '通过js直接打开', value: 1},
        {label: '静默打开(内置核心打开)', value: 2},
        {label: '浏览器打开(内置核心打开)', value: 3}
      ],
      CombineList: [
        {label: '固定目录', value: '4'},
        {label: '优先上次', value: '2'},
        {label: '自动合并', value: '1'},
        {label: '每次重新打开', value: '3'},
      ],
      channelList: [
        {label: '请选择', value: ''},
        {label: 'chrome(完整浏览器功能)', value: 'chrome'},
        {label: 'chromium(内存占用低)', value: 'chromium'},
      ],
      sse_distribute_id : '',
      processList: [],
      smartLinkConfig: {
        id: 0,
        name: '',
        links: '',
        open_num: 0,
        open_type: '',
        status: '',
        combine_type: 0,
        weight: 0,
        download_finds: '',
        channel: '',
        show_cookies: '',
        process_id: 0,
      },
      defaultSmartLinkConfig: {
        id: 0,
        name: '',
        links: '',
        open_num: 0,
        open_type: '',
        status: '',
        combine_type: 0,
        weight: 0,
        download_finds: '',
        channel: '',
        show_cookies: '',
        process_id: 0,
      },
      name: 'Link',
      smartList: [],
      smartValue: {},
      smartLinkProcessList: [],
      smartLinkRunList: {},
      tickerKey: 'link',
      versionInfo: {},
      //当前选中的配置
      chooseSmartLinkIndex: 0, //选中的第几个配置
      //已打开数量
      openPageNum: 0,
      //是否在安装中
      is_install: 0,
    }
  },
  mounted: function () {
    this.sse_distribute_id = sseDistribute.GetSseDistributeId('link')
    this.sseCreate()
    this.init()

  },
  activated() {
    if (Init.GetIsInit('smart_link') === true) {
      let _that = this
      _that.init()
      Init.DelInit('smart_link')
    }
  },
  methods: {
    sseCreate: function () {
      let _that = this
      sseDistribute.RegisterReceive(_that.sse_distribute_id , function (msg,msgType,sseDistributeId){
        if (msg === sse.SseEventClean) {
          _that.shellController.sshResult = ''
          _that.shellController.sourceSshResult = '';
        } else if (msg === sse.SseEventLogin) {
          _that.dialogLoginUserName = true
        } else if (msg.startsWith(sse.SseEventProcess)) { //准备替换
          _that.replaceRegex = msg.replace(sse.SseEventProcess, '')
        } else {
          _that.shellController.sourceSshResult += msg
          _that.shellController.sshResult = _that.shellController.sourceSshResult
        }
      })
    },
    init: function () {
      let _that = this
      _that.GetProcessList()
      _that.GetConfigList()
      let _height = base.GetDivHeight()
      _that.windowChange()
      _that.tickerRunList()
      _that.SmartLinkChromeVersion()
      setTimeout(function () {
        let _height = base.GetDivHeight2()
        _that.shellController.divHeight = parseInt(_height) - 60
        _that.windowChange()
      }, 1000)
    },
    processChange: function (newData) {
      let _that = this
      _that.smartLinkConfig.processNew = newData
    },
    linksChange: function (newData) {
      let _that = this
      _that.smartLinkConfig.linksNew = newData

    },
    SmartLinkChromeVersion: function () {
      let _that = this
      smart_link_set.SmartLinkChromeVersion(_that.sse_distribute_id , function (response) {
        if (response.ErrCode === 0) {
          _that.versionInfo = response.Data.version
          _that.is_install = response.Data.is_install
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
    changeChooseLink: function (smartLinkIndex, linkIndex) {
      let _that = this
      _that.chooseSmartLinkIndex = smartLinkIndex
      _that.smartList[smartLinkIndex].chooseLinkIndex = linkIndex
      ticker_step.Active(_that.tickerKey)
    },
    smartLinkRun: function (smartLinkIndex, linkIndex) {
      let _that = this
      if (smartLinkIndex !== null && smartLinkIndex !== undefined && linkIndex === null) {
        _that.chooseSmartLinkIndex = smartLinkIndex
        smartLinkIndex = _that.chooseSmartLinkIndex
        linkIndex = _that.smartList[_that.chooseSmartLinkIndex].chooseLinkIndex
      }

      let chooseSmartLink = _that.smartList[smartLinkIndex]
      let chooseLink = chooseSmartLink.linkList[linkIndex]

      let chooseUser = {}
      for (let i in chooseLink.userList) {
        if (chooseLink.userList[i].user_name === chooseLink.chooseUserName) {
          chooseUser = chooseLink.userList[i]
        }
      }
      let runParams = {
        id: chooseSmartLink.id,
        label: chooseLink.label,
        user_name: chooseUser.user_name,
        password: chooseUser.password,
        open_num: chooseSmartLink.open_num_new,
        open_type : chooseSmartLink.open_type_new,
        sse_distribute_id : _that.sse_distribute_id,
      }
      smart_link_set.SmartLinkRun(runParams, function (response) {
        ticker_step.Active(_that.tickerKey)
      });
    },
    tickerRunList: function () {
      let _that = this
      ticker_step.Register(_that.tickerKey, 5, function () {
        _that.runList()
      })
    },
    processMsg: function (msg) {
      let _that = this
      //处理进度
      const regReceivingObjects = new RegExp(_that.replaceRegex);

      let regList = [regReceivingObjects]
      let boolFind = false
      for (let i in regList) {
        let reg = regList[i]
        let matchList = msg.match(reg); //收到的消息是否匹配到
        if (t.IsArray(matchList) && matchList.length > 0) {
          let strMatchList = _that.shellController.sourceSshResult.match(reg)
          if (t.IsArray(strMatchList) && strMatchList.length > 0) {
            _that.shellController.sourceSshResult = _that.shellController.sourceSshResult.replace(reg, matchList[0])
            boolFind = true
          }
        }
      }
      if (!boolFind) {
        _that.shellController.sourceSshResult += msg
      }
    },
    runList: function () {
      let _that = this
      smart_link_set.SmartLinkRunList(_that.sse_distribute_id , function (response) {
        let runList = response.Data
        _that.openPageNum = 0
        _that.smartLinkRunList = {};
        for (let i in runList) {
          if (_that.smartLinkRunList[runList[i].name]) {
            _that.smartLinkRunList[runList[i].name] += runList[i].page_num
          } else {
            _that.smartLinkRunList[runList[i].name] = runList[i].page_num
          }
          _that.openPageNum += runList[i].page_num
        }
        //赋值
        for (let i in _that.smartList) {
          for (let j in _that.smartList[i].linkList) {
            _that.smartList[i].linkList[j].runNum = 0
            for (let runName in _that.smartLinkRunList) {
              if (runName === "link_id_"+_that.smartList[i].id + "_label_" + _that.smartList[i].linkList[j].label) {
                _that.smartList[i].linkList[j].runNum = _that.smartLinkRunList[runName]
              }
            }
          }
        }
      });
    },
    saveSmartLink: function () {
      let _that = this
      console.log(_that.smartLinkConfig)
      if (_that.smartLinkConfig.linksNew && _that.smartLinkConfig.linksNew !== '') {
        _that.smartLinkConfig.links = _that.smartLinkConfig.linksNew
        _that.smartLinkConfig.linkList = JSON.parse(_that.smartLinkConfig.links)
      } else {
        _that.smartLinkConfig.linkList = JSON.parse(_that.smartLinkConfig.links)
      }
      smart_link_set.SmartLinkAdd(_that.smartLinkConfig, function (response) {
        if (response.ErrCode === 0) {
          _that.dialogSmartLink = false
        } else {
          _that.$helperNotify.error('失败')
        }
        // 更新页面上的
        for (let i in _that.smartList) {
          if (_that.smartList[i].id === _that.smartLinkConfig.id) {
            _that.smartList[i].linkList = _that.smartLinkConfig.linkList
          }
        }
        ticker_step.Active(_that.tickerKey)
      })
    },
    showEditDialog: function (smartLink) {
      let _that = this
      if (smartLink !== undefined) {
        _that.smartLinkConfig = smartLink
      }
      _that.dialogSmartLink = true
      _that.smartLinkConfig.linksNew = ''
      _that.smartLinkConfig.processNew = ''
    },
    showUserPasswordList: function (smartLink) {
      let linkList = smartLink.linkList
      let showList = []
      if (linkList && t.IsArray(linkList)) {
        for (let i in linkList) {
          //浏览器自带验证
          if (linkList[i].browser_auth_username) {
            showList.push({
              label: linkList[i].label,
              username: linkList[i].browser_auth_username,
              password: linkList[i].browser_auth_password
            })
          }
          //自己设置的账号密码
          let userList = linkList[i].userList
          if (userList && t.IsArray(userList)) {
            for (let j in userList) {
              showList.push({
                label: linkList[i].label,
                username: userList[j].user_name,
                password: userList[j].password
              })
            }
          }
        }
      }
      this.showUserPassList = showList
      this.dialogShowUserPass = true
    },
    deleteSmartLink: function (smartLink) {
      let _that = this
      smart_link_set.SmartLinkDelete(smartLink, function (response) {
        if (response.ErrCode === 0) {
          _that.GetConfigList()
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
    showMarkdown: function () {
      let _that = this
      _that.dialogShowMarkdown = true
    },
    changeToProcess: function () {
      let _that = this
      _that.$emit('changeModelToEditProcess')
    },
    changeToFlow : function () {
      let _that = this
      _that.$emit('changeModelToFlow')
    },
    showDialogRunLog : function (){
      let _that = this
      _that.dialogSsePushLog = true
      setTimeout(function (){
        // shell.ShellDivToBottom(true)
      },500)
    },
    downloadPath: function () {
      let _that = this
      smart_link_set.SmartLinkDownloadPath(_that.sse_distribute_id , function (response) {
        if (response.ErrCode === 0) {
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
    install: function () {
      let _that = this
      smart_link_set.SmartLinkChromeUpdate(_that.sse_distribute_id , function (response) {
        if (response.ErrCode === 0) {
          _that.GetConfigList()
          _that.runList()
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
    recycle: function () {
      let _that = this
      smart_link_set.SmartLinkRecycle(_that.sse_distribute_id , function (response) {
        if (response.ErrCode === 0) {
          _that.GetConfigList()
          _that.runList()
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
    showCreateDialog: function () {
      let _that = this
      _that.smartLinkConfig = _that.defaultSmartLinkConfig
      _that.dialogSmartLink = true
    },
    GetConfigList: function () {
      let _that = this
      smart_link_set.SmartLinkList(function (response) {
        if (response.ErrCode === 0) {
          _that.smartList = response.Data.smart_link_list
          _that.formatSmartList()
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
    GetProcessList: function () {
      let _that = this
      Process.SmartProcessList(function (response) {
        if (response.ErrCode === 0) {
          _that.processList = response.Data.list
        } else {
          _that.$helperNotify.error('获取执行列表失败')
        }
      })
    },
    windowChange: function () {
      let _that = this
      window.addEventListener('resize', function () {
        let _height = base.GetDivHeight2()
        _that.shellController.divHeight = parseInt(_height) - 60
      });
    },
    formatSmartList: function () {
      let _that = this
      _that.chooseSmartLinkIndex = '0' //默认第一个
      for (let i in _that.smartList) {
        _that.smartList[i].linkList = JSON.parse(_that.smartList[i].links)
        _that.smartList[i].open_num_new = _that.smartList[i].open_num
        _that.smartList[i].open_type_new = _that.smartList[i].open_type
        //排序
        // _that.smartList[i].linkList = arr.SortByKey(this.smartList[i].linkList, 'value', 'asc')
        //默认选中第一个的第一个
        if (parseInt(i) === 0) {
          _that.smartList[i].chooseSmartLinkIndex = _that.smartList[i].linkList[0].value //默认选中第一个的第一个
          _that.smartList[i].chooseLinkIndex = 0 //默认选中第一个的第一个的账号列表
        }
        //如果有userList 那么默认第一个
        for (let j in _that.smartList[i].linkList) {
          if (_that.smartList[i].linkList[j].userList && _that.smartList[i].linkList[j].userList.length > 0 && !_that.smartList[i].linkList[j].chooseUserName) {
            _that.smartList[i].linkList[j].chooseUserName = _that.smartList[i].linkList[j].userList[0].user_name
          }
        }
      }
    },
    redirectLink: function (linkValue) {
      let _that = this
      window.open(linkValue.link, '_blank')
      ticker_step.Active(_that.tickerKey)
    }
  }
}
</script>

<style scoped></style>
