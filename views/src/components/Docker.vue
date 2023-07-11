<template>
  <div>
    <el-card>
      <el-card>
        <h3 style="display: inline-block;">
          Docker管理
        </h3>
        <el-row :gutter="20">
          <el-col :span="2" v-for="(valueDocker,key) in dockerList" style="margin:5px;">
            <div>
              <el-radio @change="chooseDocketFunc(valueDocker)"  v-model="chooseDocketId" size="medium " :label="valueDocker.Name">
                {{ valueDocker.Name }}
              </el-radio>
            </div>
          </el-col>
        </el-row>

        <br/>
        <el-button type="primary" disabled>检查</el-button>
        <el-button type="primary" :loading="loadingStatus['restart_docker']" @click="restartDocker(chooseDocker)">重启
        </el-button>
        <el-button type="primary" :loading="loadingStatus['show_compose']" @click="showCompose(chooseDocker)">查看compose
        </el-button>
      </el-card>
      <el-input style="margin-top: 20px;" id="resultTextarea" type="textarea" v-model="execResult" rows="25"></el-input>
    </el-card>
  </div>
</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";

let dockerList = require("../config/dockerList.json")
export default {
  data() {
    return {
      name: "Vip",
      //接口地址
      apiHost: 'http://localhost:7070',
      //ssh config
      xkfDevDbConfig: {},
      sshConfig: {},
      prodTestSshConfig: {},
      //选中的vip版本
      chooseVipLevel: -1,
      chooseSystemType: -1,
      //过期时间
      expiredTime: '',
      chooseVipUserName: '',
      dockerList: dockerList,
      //总的操作类型
      ExecType: "",
      execResult: "",//操作结果
      redisConfigList: [],
      loadingStatus: {},
      chooseDocker: {"Name": "common3", "Id": "xkf_common3" , "SshName" : "wk"},
      chooseDocketId : 'common3',
    }
  },
  mounted: function () {
    let that = this
    this.apiHost = this.$helperConfig.getApiHost()
    this.sshConfig = this.$helperConfig.getXkfDevSshConfig()
    this.xkfDevDbConfig = this.$helperConfig.getXkfDevSshConfig()
    this.wkSshConfig = this.$helperConfig.getWkDevSshConfig()
    this.loadingStatus = this.$helperLoad.getExecTypeStatus()
    this.queryDockerPs()
    setInterval(function (){
      that.queryDockerPs()
    } , 60000);

  },
  methods: {
    chooseDocketFunc : function (value){
      this.chooseDocker = value
    },
    queryDockerPs : function (){
      let currentDateTime = this.$helperCommon.getCurrentDateTime()
      let _that = this
      //根据类型判断
      let params = {
        SshConfig: this.sshConfig,
        WkSshConfig : this.wkSshConfig,
        ExecType: 'docker_ps',
      }
      _that.setLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.cancelLoading(params)
        _that.execResult = response.Data
        //分析结果
        let notifyList = []
        let dockerList = response.Data.split('\n')
        let rightDockerList = []
        for(let i in dockerList){
          let temp = dockerList[i]
          for(let j = 0;j < 20;j++){
            temp = temp.replace('  ',' ')
          }
          rightDockerList.push(temp)
        }
        //开始分割
        for(let i in rightDockerList){
          let temp = rightDockerList[i].split(' ')
          if(temp.length <= 0){
            continue;
          }
          if(temp[0] === "CONTAINER"){
            continue
          }
          let cpu = parseFloat(temp[2])
          let memory = parseFloat(temp[6])
          console.log(cpu , memory , temp)
          if(cpu > 95 || memory > 95){
            notifyList.push( { name: currentDateTime + ' ' + temp[1] + ' ：cpu：' + temp[2] + '，内存：' + temp[6], type: 'danger' })
          }else if(cpu > 90 || memory > 90){
            notifyList.push( { name: currentDateTime + ' ' + temp[1] + ' ：cpu：' + temp[2] + '，内存：' + temp[6], type: 'warning' })
          }
        }
        _that.$parent.showNotify(notifyList)
      });
    },
    getSshConfig: function (value) {
      //根据类型判断
      let chooseSshConfig = this.sshConfig
      if (value.SshName === 'wk') {
        chooseSshConfig = this.wkSshConfig
      }
      return chooseSshConfig
    },
    //重启docker
    restartDocker: function (dockerValue) {
      let _that = this
      //根据类型判断
      let params = {
        SshConfig: this.getSshConfig(dockerValue),
        DockerCodeName: dockerValue.Name,
        ExecType: 'restart_docker',
      }
      _that.setLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.cancelLoading(params)
        _that.execResult = response.Data
      });
    },
    //查看compose配置内容
    showCompose: function (dockerValue) {
      let _that = this
      //根据类型判断
      let params = {
        SshConfig: this.getSshConfig(dockerValue),
        DockerCodeName: dockerValue.Name,
        ExecType: 'show_compose',
      }
      _that.setLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.cancelLoading(params)
        _that.execResult = response.Data
      });
    },
    redirectLink: function (linkValue) {
      this.execResult = linkValue.link
      window.open(this.execResult, '_blank');
    },
    setLoading: function (params) {
      this.loadingStatus[params.ExecType] = true
      let that = this
      setTimeout(function () {
        that.loadingStatus[params.ExecType] = false
      }, 25000)
    },
    cancelLoading: function (params) {
      let that = this
      setTimeout(function (){
        that.loadingStatus[params.ExecType] = false
      } , 1000)
    },
  },
}
</script>

<style scoped>

</style>
