<template>
  <el-card>
      <div v-for="(parentTypeValue,k) in parentTypeList">
        <h3 style="display: inline-block;">
          {{ parentTypeValue.Title }}
        </h3>
        <el-row :gutter="20">
          <el-col :span="2" v-for="(value,key) in codeEnvList" style="margin:5px;"
                  v-if="value.ParentType === parentTypeValue.Name">
            <div>
              <el-radio @change="queryCurrentBranch(value)" size="medium " v-model="chooseCodeUnikey" :label="value.Unikey">{{value.Name}}
              </el-radio>
            </div>
          </el-col>
        </el-row>

      </div>
      <br/>
      <el-button type="primary" :loading="btnLoading.pull" @click="GitPullBranchOrigin">拉取最新代码
      </el-button>
      <el-button type="primary" :loading="btnLoading.status" @click="GitQueryStatus">查看分支变更</el-button>
      <el-input v-if="showChangeBranch" style="width:300px;margin-right:20px;" v-model="BranchName" placeholder="请输入分支名"></el-input>
      <el-button type="primary" :loading="btnLoading.change" @click="GitChangeBranch">切换分支
      </el-button>
      <el-button type="primary" :loading="btnLoading.query" @click="queryCurrentBranch">查看分支
      </el-button>


<!--    <el-input style="margin-top: 20px;" id="resultTextarea" type="textarea" v-model="execResult" rows="25"></el-input>-->
  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";
import git from "../utils/api/git.js"
import mod from "../utils/api/module.js"
import base from "../utils/api/base.js"

export default {
  data() {
    return {
      name: "Git",
      //接口地址
      apiHost: '',
      //输入框
      showChangeBranch: false,
      //选中的环境
      chooseCodeUnikey: "xkf-common3",
      //代码环境
      codeEnvList: [],
      //按钮状态
      btnLoading: {
        exec: false,
        pull: false,
        change: false,
        status: false,
        query : false,
      },
      parentTypeList: [
        {Title: "小客服（php）", Name: "xkf"},
        {Title: "企微（php）", Name: "wk"},
        {Title: "视频号小店（golang）", Name: "weixin_shop_golang"},
        // {Title: "预发布", Name: "prodTest"},
      ],
      BranchName: "",  //分支名
      execResult: "",//操作结果
    }
  },
  mounted: function () {
    let that = this
    this.apiHost = base.GetApiHost()
    this.codeEnvList = mod.GetCodeConfigList()
    setTimeout(function () {
      that.queryCurrentBranch()
    },1500);

  },
  methods: {
    queryCurrentBranch : function (){
      let that = this
      //显示隐藏终端
      let codeConfig = git.GitGetCodeConfigByUnikey(this.chooseCodeUnikey)
      this.$parent.$parent.showTerminal(base.GetUnikey() , codeConfig.SshName)
      that.btnLoading.query = true;
      git.GitCurrentBranch(this.chooseCodeUnikey , function (response){
        that.execResult = response.Data;
        setTimeout(function () {
          that.textareaScroll()
          that.btnLoading.query = false;
        }, 500)
      })
    },
    GitPullBranchOrigin : function () {
      let that = this
      that.btnLoading.pull = true;
      git.GitPullBranchOrigin(this.chooseCodeUnikey , function (response){
        that.execResult = response.Data;
        setTimeout(function () {
          that.textareaScroll()
          that.btnLoading.pull = false;
        }, 500)
      })
    },
    GitQueryStatus : function () {
      let that = this
      that.btnLoading.status = true;
      git.GitQueryStatus(this.chooseCodeUnikey , function (response){
        that.execResult = response.Data;
        setTimeout(function () {
          that.textareaScroll()
          that.btnLoading.status = false;
        }, 500)
      })
    },
    GitChangeBranch : function () {
      if(!this.showChangeBranch){
        this.showChangeBranch = true;
        return;
      }
      let that = this
      that.btnLoading.change = true;
      git.GitChangeBranch(this.chooseCodeUnikey ,this.BranchName, function (response){
        that.execResult = response.Data;
        that.showChangeBranch = false;
        setTimeout(function () {
          that.textareaScroll()
          that.btnLoading.change = false;
        }, 500)
      })
    },
    //textarea滚动到最后
    textareaScroll: function () {
      //document.getElementById("resultTextarea").scrollTop = document.getElementById("resultTextarea").scrollHeight;
    },
  },
}
</script>

<style scoped>

</style>
