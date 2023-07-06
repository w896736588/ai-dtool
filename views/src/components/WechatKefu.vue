<template>
  <el-card>
    <el-card>
<!--      代码环境-->
      <div>
        <h3 style="display: inline-block;">
        当前所选环境的微信客服
        </h3>
        <el-row :gutter="20">
          <el-col :span="5" v-for="(value,key) in wechatKefuList" style="margin:5px;display: inline-block;">
            <div>
              <el-radio size="medium " @change="changeWechatKefu" v-model="chooseWechatKefuAppid" :label="value.app_id">
                {{ value.app_name }}
              </el-radio>
            </div>
          </el-col>
        </el-row>
<!--        <el-input style="width:500px;display: inline-block;margin-top:5px;" v-model="chooseWechatKefuAppid" placeholder="请输入微信客服appid或应用id"></el-input>-->
        <h3 style="display: inline-block;">
          环境
        </h3>
        <el-row :gutter="20" style="margin-top:5px;">
          <el-col v-if="value.ParentType === 'xkf'" :span="2" v-for="(value,key) in codeEnvList" style="margin:5px;">
            <div>
              <el-radio size="medium " v-model="chooseEvnName" @change="queryEnvWechatKefuList" :label="value.Name">{{value.NameTitle}}</el-radio>
            </div>
          </el-col>
        </el-row>
      </div>
      <br/>
      <el-button type="primary" :loading="loadingStatus['wechat_kefu_change']" @click="exec('wechat_kefu_change')">切换到当前选择的环境</el-button>
      <el-button type="primary" :loading="loadingStatus['wechat_kefu_status']" @click="exec('wechat_kefu_status')">查看所在环境</el-button>
      <el-button type="primary" :loading="loadingStatus['WechatKefuChannelQrList']" @click="queryEnvWechatKefuQrList()">二维码访问</el-button>
    </el-card>
    <el-card style="margin-top: 10px;">
      <h3 style="display: inline-block;">
        二维码
      </h3>
      <div v-for="(value,key) in qrList">
        <h4 style="display: inline-block;">
          {{value.channel_name}}
        </h4>

        <el-row :gutter="20" style="margin-top:5px;">
          <el-col v-for="(link,key1) in value.link_list" :span="2"  style="margin:5px;">
            <div>
              <el-button round style="display: inline-block;" @click="showQrCode(link)">{{link.staff_name}}</el-button>
            </div>
          </el-col>
        </el-row>

      </div>
    </el-card>

    <el-input style="margin-top: 20px;" type="textarea" v-model="execResult" rows="25"></el-input>
    <el-dialog :title="showQrCodeTitle" :visible.sync="isShowCard" width="500px" center>
      <div style="display: flex;margin: auto;">
        <div id="qrCode" ref="qrCodeDiv" style="margin: auto;" center></div>
      </div>
      <div style="word-break:break-all;margin: 5px;">{{qrCodeContent}}</div>
    </el-dialog>
  </el-card>

</template>

<script>
import Vue from "vue";
import QRCode from 'qrcodejs2';
export default {
  data() {
    return {
      isShowCard : false,
      qrCodeContent : '',
      showQrCodeTitle : '',
      name: "WechatKefu",
      userNameList : [],
      //接口地址
      apiHost: '',
      //ssh config
      sshConfig: {},
      wkSshConfig : {},
      xkfDevDbConfig : {},
      //选中的环境
      chooseEvnName: "common3-xkf",
      //代码环境
      codeEnvList: [],
      //docker
      dockerList: [],
      //微信客服合集
      chooseWechatKefuName: "",
      chooseWechatKefuAppid: "",
      wechatKefuList: [],
      execResult: "",//操作结果
      ercode : "@/assets/2code.png",

      qrList : [],
      loadingStatus : {},
    }
  },
  mounted: function () {
    this.sshConfig = this.$helperConfig.getXkfDevSshConfig()
    this.apiHost = this.$helperConfig.getApiHost()
    this.wkSshConfig = this.$helperConfig.getWkDevSshConfig()
    this.xkfDevDbConfig = this.$helperConfig.getXkfDevDbConfig()
    this.codeEnvList = this.$helperConfig.getCodeEnvList()
    this.dockerList = this.$helperConfig.getDockerList()
    this.userNameList = this.$helperConfig.getUsernameList()
    this.loadingStatus = this.$helperLoad.getExecTypeStatus()
    this.queryEnvWechatKefuList()
  },
  components: {
    QRCode
  },
  methods: {
    changeWechatKefu : function () { //切换微信客服
      this.queryEnvWechatKefuQrList()
    },
    createQrCode : function (link){
      this.qrCodeContent = link.short_code
      this.showQrCodeTitle = link.staff_name
      this.$nextTick(()=>{
        this.$refs.qrCodeDiv.innerHTML = '';//二维码清除
        new QRCode(this.$refs.qrCodeDiv, {
          text: link.short_code,//二维码链接，参数是否添加看需求
          width: 200,//二维码宽度
          height: 200,//二维码高度
          colorDark: "#333333", //二维码颜色
          colorLight: "#ffffff", //二维码背景色
          correctLevel: QRCode.CorrectLevel.L //容错率，L/M/H
        });
      })
    },
    showQrCode : function (link){
      this.isShowCard = true
      this.createQrCode(link)
    },
    //执行
    exec: function (execType) {
      let _that = this
      this.ExecType = execType
      //找到环境配置
      let env_config = this.$helperConfig.getCodeEnvConfigByCodeEnvName(this.codeEnvList , this.chooseEvnName)
      if (env_config === {}) {
        _that.$helperNotify.error("不存在的配置");
        return
      }

      //根据类型判断
      let params = {
        SshConfig: _that.sshConfig,
        CodePath: env_config.CodePath,
        ExecType: this.ExecType,
        WechatKefuAppid: this.chooseWechatKefuAppid,
        DockerList: this.dockerList,
        DockerId: this.$helperConfig.getDockerIdByCodeEnvConfig(this.dockerList , env_config),
        DockerCodePath: env_config.DockerCodePath,
        xkfDevDbConfig : this.xkfDevDbConfig,
        WkSshConfig : this.wkSshConfig,
        SshName : env_config.SshName,
      }
      if (params.ExecType === 'wechat_kefu_status' && params.WechatKefuAppid === '') {
        _that.$helperNotify.error('请输入应用id或appid')
        return
      } else if (params.ExecType === 'wechat_kefu_change' && (params.WechatKefuAppid === '' || params.CodePath === '')) {
        _that.$helperNotify.error('选择微信客服以及代码环境')
        return
      }
      if (params.ExecType === 'wechat_kefu_change' && params.DockerId === ``) {
        _that.$helperNotify.error('代码环境找不到对应的docker')
        return
      }
      //按钮加载状态
      _that.setLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.cancelLoading(params)
      });
    },
    queryEnvWechatKefuList : function (){
      let _that = this
      _that.chooseWechatKefuAppid = ''
      //找到环境配置
      let env_config = this.$helperConfig.getCodeEnvConfigByCodeEnvName(this.codeEnvList , this.chooseEvnName)
      if (env_config === {}) {
        _that.$helperNotify.error("不存在的配置");
        return
      }
      //根据类型判断
      let params = {
        Account : this.$helperConfig.getUserNameByEnvCode(this.userNameList , env_config.NameTitle),
        ExecType: 'query_env_wechatkefu_list',
        xkfDevDbConfig : this.xkfDevDbConfig,
      }
      //按钮加载状态
      _that.setLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.cancelLoading(params)
        _that.wechatKefuList = JSON.parse(response.Data)
      });
    },
    //二维码列表
    queryEnvWechatKefuQrList : function () {
      let _that = this
      //找到环境配置
      let env_config = this.$helperConfig.getCodeEnvConfigByCodeEnvName(this.codeEnvList, this.chooseEvnName)
      if (env_config === {}) {
        _that.$helperNotify.error("不存在的配置");
        return
      }
      if (_that.chooseWechatKefuAppid === '') {
        _that.$helperNotify.error('请选择微信客服')
        return
      }
      //根据类型判断
      let params = {
        WechatKefuAppid: _that.chooseWechatKefuAppid,
        ExecType: 'WechatKefuChannelQrList',
        xkfDevDbConfig: this.xkfDevDbConfig,
      }
      //按钮加载状态
      _that.setLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.cancelLoading(params)
        _that.qrList = JSON.parse(response.Data)
      })
    },
    setLoading : function (params){
      this.loadingStatus[params.ExecType] = true
      let that = this
      setTimeout(function (){
        that.loadingStatus[params.ExecType] = false
      } , 25000)
    },
    cancelLoading : function (params){
      let that = this
      setTimeout(function (){
        that.loadingStatus[params.ExecType] = false
      } , 1000)
    },
  },
}
</script>

<style >
  #qrCode img {
    margin:auto !important;
  }
</style>
