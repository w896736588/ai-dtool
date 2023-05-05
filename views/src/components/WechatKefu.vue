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
              <el-radio size="medium " v-model="chooseWechatKefuAppid" :label="value.app_id">
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
      <el-button type="primary" :loading="btnLoading.wechatKefuChange" @click="ExecType = 'wechat_kefu_change';exec()">切换到当前选择的环境</el-button>
      <el-button type="primary" :loading="btnLoading.wechatKefuStatus" @click="ExecType = 'wechat_kefu_status';exec()">查看所在环境</el-button>
    </el-card>

    <el-input style="margin-top: 20px;" type="textarea" v-model="execResult" rows="25"></el-input>
  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";
export default {
  data() {
    return {
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
      //按钮状态
      btnLoading : {
        wechatKefuStatus : false,
        wechatKefuChange : false,
      },
      //微信客服合集
      chooseWechatKefuName: "",
      chooseWechatKefuAppid: "",
      wechatKefuList: [],
      execResult: "",//操作结果
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
    this.queryEnvWechatKefuList()
  },
  methods: {
    //执行
    exec: function () {
      let _that = this
      //找到环境配置
      let env_config = this.$helperConfig.getCodeEnvConfigByCodeEnvName(this.codeEnvList , this.chooseEvnName)
      if (env_config === {}) {
        _that.$helperNotify.error("不存在的配置");
        return
      }
      env_config.SshConfig = _that.sshConfig
      //根据类型判断
      let params = {
        SshConfig: env_config.SshConfig,
        CodePath: env_config.CodePath,
        ExecType: this.ExecType,
        WechatKefuAppid: this.chooseWechatKefuAppid,
        DockerList: this.dockerList,
        DockerId: this.$helperConfig.getDockerIdByCodeEnvConfig(this.dockerList , env_config),
        DockerCodePath: env_config.DockerCodePath,
        xkfDevDbConfig : this.xkfDevDbConfig,
        WkSshConfig : this.wkSshConfig,
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
      _that.setBtnLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.cancelBtnLoading(params)
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
      this.$helperConfig.get
      console.log(env_config)
      //根据类型判断
      let params = {
        Account : this.$helperConfig.getUserNameByEnvCode(this.userNameList , env_config.NameTitle),
        ExecType: 'query_env_wechatkefu_list',
        xkfDevDbConfig : this.xkfDevDbConfig,
      }
      //按钮加载状态
      _that.setBtnLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.cancelBtnLoading(params)
        _that.wechatKefuList = JSON.parse(response.Data)
      });
    },
    setBtnLoading : function (params){
      if(params.ExecType === 'wechat_kefu_status'){
        this.btnLoading.wechatKefuStatus = true
      }else if(params.ExecType === 'wechat_kefu_change'){
        this.btnLoading.wechatKefuChange = true
      }
      let _this = this
      let _set_params = params
      setTimeout(function (){
        _this.cancelBtnLoading(_set_params)
      } , 15000)
    },
    cancelBtnLoading : function (params){
      if(params.ExecType === 'wechat_kefu_status'){
        this.btnLoading.wechatKefuStatus = false
      }else if(params.ExecType === 'wechat_kefu_change'){
        this.btnLoading.wechatKefuChange = false
      }
    },
  },
}
</script>

<style scoped>

</style>
