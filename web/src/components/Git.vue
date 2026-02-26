<template>
  <div id="mainCard" ref="mainCard" class="box-card" style="text-align: center;">

    <el-tabs v-model="chooseGroupId" :tab-position="tabPosition" class="demo-tabs" style="min-height: 150px" @tab-change="changeGitGroup">
      <el-tab-pane v-for="(groupInfo, k) in gitGroupConfigList" :key="k" :label="groupInfo.name" :name="groupInfo.id">
        <el-row :gutter="20">
          <template v-for="(value, key) in gitConfigList">
            <el-col v-if="value.git_group_id === groupInfo.id" :span="3">
              <div>
                <el-radio v-model="chooseGitId" :label="value.id" size="large" @change="ChangeGit(value)">{{
                    value.name
                  }}
                </el-radio>
              </div>
            </el-col>
          </template>
        </el-row>
      </el-tab-pane>
    </el-tabs>
    <el-button v-loading="btnLoading.pull" type="primary" @click="GitPullBranchOrigin">拉取</el-button>
    <el-button v-loading="btnLoading.status" type="primary" @click="GitQueryStatus">状态</el-button>
    <el-input v-if="showChangeBranch" ref="inputBranchName" v-model="BranchName" placeholder="请输入分支名" style="width: 300px; margin-right: 10px;margin-left:10px;"></el-input>
    <el-button v-loading="btnLoading.change" type="primary" @click="GitChangeBranch">切换分支</el-button>
    <!--    <el-input ref="inputBranchNameRemote" v-if="showChangeBranchRemote" style="width: 300px; margin-right: 10px;margin-left:10px;" v-model="BranchNameRemote" placeholder="请输入分支名"></el-input>-->
    <!--    <el-button type="primary" v-loading="btnLoading.changeRemote" @click="GitChangeBranchRemote">关联远程分支切换（ depth=1）</el-button>-->
    <el-button v-loading="btnLoading.query" type="primary" @click="queryCurrentBranch">查看当前分支</el-button>
    <el-button v-loading="btnLoading.queryLog" type="primary" @click="queryCommitLog">查看日志</el-button>
    <el-button type="primary" @click="drawerVisibleMarkdown = true">帮助文档</el-button>&nbsp;
    <!--    <el-button type="primary" @click="drawerVisibleMarkdown = true">查看git config</el-button>-->
    <!--    <el-button type="primary" v-loading="btnLoading.exec" @click="GitExec">增加保存账号密码配置</el-button>-->

    <el-dropdown @command="handleDropdownCommand">
      <el-button type="primary">
        更多操作<i class="el-icon-arrow-down el-icon--right"></i>
      </el-button>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="changeBranchRemote">关联远程分支切换（depth=1）</el-dropdown-item>
          <el-dropdown-item command="viewGitConfig">查看git config</el-dropdown-item>
          <el-dropdown-item command="saveCredentials">增加保存账号密码配置</el-dropdown-item>
          <el-dropdown-item command="setSafe">设置项目目录安全</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- 将原有的按钮移除，并保留输入框逻辑 -->
    <el-input
        v-if="showChangeBranchRemote"
        ref="inputBranchNameRemote"
        v-model="BranchNameRemote"
        placeholder="请输入分支名"
        style="width: 300px; margin-right: 10px;margin-left:10px;"
    ></el-input>
  </div>
  <p></p>
  <shellResult ref="shellRef" :divHeight="shellController.divHeight" :isRunning="shellController.isRunning" :shellShowResult="shellController.sshResult" :show-model="shellController.showModel"></shellResult>
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
.text {
  font-size: 14px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.item {
  margin-bottom: 10px;
}

</style>
<script>
import git from '../utils/base/git.js'
import shell from "@/utils/base/shell"
import shellResult from "@/components/shell/result_div.vue";
import format from "@/utils/base/format";
import arr from "@/utils/base/array";
import sse from "@/utils/base/sse"
import t from "@/utils/base/type";
import Init from '@/utils/base/set_init'
import base from "@/utils/base";
import Markdown from "@/components/Markdown.vue";
import sseDistribute from "@/utils/base/sse_distribute";
import {Throttle_string} from "@/utils/base/throttle_string";

export default {
  props: {},
  components: {
    Markdown,
    shellResult,
  },
  data() {
    return {
      //shell
      shellController: {
        sshResult: '',
        sourceSshResult: '',
        isRunning: false,
        showModel: 'div',
        divHeight: 250,
      },
      drawerVisibleMarkdown: false,
      name: 'Git',
      //输入框
      showChangeBranch: false,
      showChangeBranchRemote: false,
      tabPosition: 'top',
      markdownType: 'git',
      //按钮状态
      btnLoading: {
        exec: false,
        pull: false,
        change: false,
        changeRemote: false,
        status: false,
        query: false,
      },
      BranchName: '', //分支名
      BranchNameRemote: '',
      gitGroupConfigList: [],
      gitConfigList: [],
      selectGitConfig: {},
      chooseGroupId: 0,
      chooseGitId: 0,
      sseId: '',
    }
  },
  mounted: function () {
    let _that = this
    _that.sse_distribute_id = sseDistribute.GetSseDistributeId('git')
    _that.GetGitConfigList()
    _that.windowChange()
    shell.calculateShellDivHeight(_that)
    _that.test()
  },
  activated: function () {
    let _that = this
    setTimeout(function () {
      shell.calculateShellDivHeight(_that)
    }, 500)
    if (Init.GetIsInit('git') === true) {
      let _that = this
      _that.GetGitConfigList()
      _that.windowChange()
      _that.test()
      Init.DelInit('git')
    }
  },
  methods: {
    test: function () {
    },
    handleDropdownCommand(command) {
      switch (command) {
        case 'changeBranchRemote':
          this.handleChangeBranchRemote();
          break;
        case 'viewGitConfig':
          this.drawerVisibleMarkdown = true;
          this.markdownType = 'git-config'; // 假设有这个类型
          break;
        case 'saveCredentials':
          this.GitSaveCredentials();
          break;
        case 'setSafe':
          this.GitSetSafe();
          break;
        default:
          break;
      }
    },
    GitSaveCredentials(){
      let _that = this
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitSaveCredentials(_that.selectGitConfig, function (response) {
            if (response.ErrCode === 0) {
              _that.$helperNotify.success('成功')
            } else {
              _that.$helperNotify.error('失败')
            }
          }
      )
    },
    GitSetSafe() {
      let _that = this
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.SetSafe(_that.selectGitConfig, function (response) {
            if (response.ErrCode === 0) {
              _that.$helperNotify.success('成功')
            } else {
              _that.$helperNotify.error('失败')
            }
          }
      )
    },
    handleChangeBranchRemote() {
      let _that = this;
      if (!this.showChangeBranchRemote) {
        this.showChangeBranchRemote = true;
        setTimeout(async function () {
          _that.$refs.inputBranchNameRemote?.focus()
        }, 500)
        return
      }
      _that.btnLoading.changeRemote = true
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitChangeBranchRemote(_that.selectGitConfig, _that.BranchNameRemote, function (response) {
            _that.showChangeBranchRemote = false
            setTimeout(function () {
              _that.btnLoading.changeRemote = false
            }, 500)
          }
      )
    },
    tryReconnectionSocket: function () {
      let _that = this
      if (!_that.selectGitConfig || !_that.selectGitConfig.ssh_id) {
        return
      }
      let throttleStringFunc = new Throttle_string(50, text => {
        _that.shellController.sshResult += text
        // 限制长度：最多保留最后 50000 个字符
        const maxLen = 50000;
        if (_that.shellController.sshResult.length > maxLen) {
          _that.shellController.sshResult = _that.shellController.sshResult.slice(-maxLen);
        }
        // 注意这里读取的是“当前最新”结果，避免闭包旧值
        let result = format.formatResult(
            _that.shellController.sshResult, ['copy', 'color', 'replace']);
        result = format.formatResult(result, ['length']);
        _that.shellController.sshResult = result;   // 一次性赋值，减少 watcher 抖动
      });
      sseDistribute.RegisterReceive(_that.sse_distribute_id, function (msg, msgType, sseDistributeId) {
        throttleStringFunc.update(msg)
      })
    },
    chooseDefault: function () {
      let _that = this
      //初始化默认选中的分组和git
      _that.chooseGroupId = git.GitLocalGetLastGroupId()
      _that.chooseGitId = git.GitLocalGetLastGitId()
      for (let i in _that.gitConfigList) {
        if (parseInt(_that.gitConfigList[i].id) === parseInt(_that.chooseGitId)) {
          _that.selectGitConfig = _that.gitConfigList[i]
        }
      }
      if (_that.selectGitConfig && _that.selectGitConfig.id) {
        _that.ChangeGit(_that.selectGitConfig)
      }
    },
    windowChange: function () {
      let _that = this
      window.addEventListener('resize', function () {
        shell.calculateShellDivHeight(_that)
      });
    },
    //切换选择git
    ChangeGit: function (selectGitConfig) {
      let _that = this
      _that.shellController.sshResult = '';
      _that.selectGitConfig = selectGitConfig
      _that.chooseGitId = selectGitConfig.id
      _that.queryCurrentBranch()
      shell.calculateShellDivHeight(_that)
      git.GitLocalSetLastGitId(_that.selectGitConfig.id)
    },
    //获取git和git group列表
    GetGitConfigList: function () {
      let _that = this
      git.GitConfigList({sse_distribute_id: _that.sse_distribute_id}, function (response) {
        if (response.ErrCode === 0) {
          _that.gitConfigList = response.Data.git_list
          arr.SortByKey(_that.gitConfigList, 'name', 'asc')
          _that.gitGroupConfigList = response.Data.git_group_list
          _that.chooseDefault()
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
    //切换分组
    changeGitGroup: function () {
      let _that = this
      git.GitLocalSetLastGroupId(_that.chooseGroupId)
      if (_that.gitConfigList.length === 0) {
        return
      }
      _that.ChangeGit(_that.gitConfigList[0])
    },
    //查询分支
    queryCurrentBranch: function () {
      let _that = this
      _that.showChangeBranch = false
      _that.showChangeBranchRemote = false
      _that.btnLoading.query = true
      _that.tryReconnectionSocket()
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitCurrentBranch(_that.selectGitConfig, function (response) {
        setTimeout(function () {
          _that.btnLoading.query = false
        }, 500)
      })
    },
    //查询commit log
    queryCommitLog: function () {
      let _that = this
      _that.btnLoading.queryLog = true
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitCommitLog(_that.selectGitConfig, function (response) {
        setTimeout(function () {
          _that.btnLoading.queryLog = false
        }, 500)
      })
    },
    //拉取代码
    GitPullBranchOrigin: function () {
      let _that = this
      _that.btnLoading.pull = true
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitPullBranchOrigin(_that.selectGitConfig, function (response) {
        setTimeout(function () {
          _that.btnLoading.pull = false
        }, 500)
      })
    },
    //查询状态
    GitQueryStatus: function () {
      let _that = this
      _that.btnLoading.status = true
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitQueryStatus(_that.selectGitConfig, function (response) {
        setTimeout(function () {
          _that.btnLoading.status = false
        }, 500)
      })
    },
    GitChangeBranchRemote: function () {
      let _that = this
      if (!this.showChangeBranchRemote) {
        this.showChangeBranchRemote = true
        setTimeout(async function () {
          _that.$refs.inputBranchNameRemote?.focus()
        }, 500)
        return
      }
      _that.btnLoading.changeRemote = true
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitChangeBranchRemote(_that.selectGitConfig, _that.BranchNameRemote, function (response) {
            _that.showChangeBranchRemote = false
            setTimeout(function () {
              _that.btnLoading.changeRemote = false
            }, 500)
          }
      )
    },
    //切换分支
    GitChangeBranch: function () {
      let _that = this
      if (!this.showChangeBranch) {
        this.showChangeBranch = true
        setTimeout(async function () {
          _that.$refs.inputBranchName?.focus()
        }, 500)
        return
      }
      _that.btnLoading.change = true
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitChangeBranch(_that.selectGitConfig, _that.BranchName, function (response) {
            _that.showChangeBranch = false
            setTimeout(function () {
              _that.btnLoading.change = false
            }, 500)
          }
      )
    },
  },
}
</script>

<style scoped></style>
