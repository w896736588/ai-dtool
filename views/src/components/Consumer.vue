<template>
  <el-card>
     <el-select v-model="chooseEvnTitle">
      <el-option
        v-for="(value,key) in env.envList"
        :key="value.UniKey"
        :label="value.Title"
        :value="value.Title">
      </el-option>
    </el-select>
    <el-select v-model="execType">
      <el-option
        v-for="(value,key) in opTypeList"
        :key="value.ExecType"
        :label="value.Name"
        :value="value.ExecType">
      </el-option>
    </el-select>
    <el-input v-if="execType === 'change_branch'" style="width:400px;margin-right:20px;" v-model="paramTwo" placeholder="请输入分支名"></el-input>

    <el-button type="primary" @click="exec">执 行</el-button>

    <el-input style="margin-top: 20px;" type="textarea" v-model="execResult" rows="20"></el-input>
  </el-card>
  <!--    <el-card>-->
  <!--      <el-input type="textarea" v-model="execResult" rows="20"></el-input>-->
  <!--    </el-card>-->
</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";

export default {
  data() {
    return {
      name: "Consumer",
      tabPosition: 'left',
      ws: null,
      heartInternal: null,
      wsUrl: 'ws://localhost:7778/redisWebSocket',
      //接口地址
      apiHost: 'http://localhost:7070',
      //ssh config
      sshConfig: {
        username: "frog",
        name: "sshXkf",
        password: "frog987^%$321_220",
        host: "121.40.109.241",
        port: "22",
      },
      //选中的环境
      chooseEvnTitle: "common3",
      //按环境
      env: {
        envList: [
          {
            Name: "common3",
            Title : "common3",
            ParamOne: "yii_customer_service", //代码地址
          },
          {
            Name: "common3",
            Title : "common31",
            ParamOne: "yii_customer_service", //代码地址
          }
        ],
      },
      //操作类型
      paramTwo : "",
      execResult: "",//操作结果
      execType: "query_current_branch",
      opTypeList: [
        {
          "Name": "查询当前分支",
          "ExecType": "query_current_branch",
        },
        {
          "Name": "更新当前分支到最新代码",
          "ExecType": "pull_branch_origin",
        },
        {
          "Name": "切换分支并更新到最新",
          "ExecType": "change_branch",
        },
        {
          "Name": "切换微信客服到当前环境",
          "ExecType": "change_wechat_kefu",
        },
      ],
    }
  },
  mounted: function () {

  },
  methods: {
    //执行
    exec: function () {
      let _that = this
      //找到环境配置
      let env_config = {};
      for (let i in this.env.envList) {
        if (this.env.envList[i].Title === this.chooseEvnTitle) {
          env_config = this.env.envList[i]
          break
        }
      }
      if (env_config === {}) {
        _that.error("不存在的配置");
        return
      }
      env_config.SshConfig = _that.sshConfig
      console.log(env_config)
      console.log(this.chooseEvnTitle)
      //根据类型判断
      let params = {
        ExecType: this.execType,
        EnvName: env_config.Name,
        SshConfig: env_config.SshConfig,
        ParamOne: env_config.ParamOne,
      }
      switch (this.execType) {
        case "query_current_branch":
          Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
            _that.success('查询成功');
            _that.execResult = response.Data
          });
          break;
        case "change_branch":
          params.paramTwo = _that.paramTwo
          Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
            _that.success('查询成功');
            _that.execResult = response.Data
          });
          break;
        case "pull_branch_origin":
          Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
            _that.success('查询成功');
            _that.execResult = response.Data
          });
          break;
      }
    },
    success: function (msg) {
      Message.success(msg);
      //this.$notify({title: '提示', message: msg, type: 'success'});
    },
    warning: function (msg) {
      Message.warning(msg);
      //this.$notify({title: '提示', message: msg, type: 'warning'});
    },
    info: function (msg) {
      Message.info(msg);
      //this.$notify({title: '提示', message: msg});
    },
    error: function (msg) {
      Message.error(msg);
      //this.$notify({title: '提示', message: msg, type: 'error'});
    },
  },
}
</script>

<style scoped>

</style>
