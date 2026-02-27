<template>
  <div class="git-page-container">
    <!-- 顶部操作区域 -->
    <div class="git-header-card">
      <div class="header-title">
        <svg class="header-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2z" stroke="currentColor" stroke-width="2"/>
          <circle cx="12" cy="12" r="3" fill="currentColor"/>
          <path d="M12 2v4M12 18v4M2 12h4M18 12h4" stroke="currentColor" stroke-width="2"/>
        </svg>
        <span>Git 版本管理</span>
      </div>
      
      <!-- 项目选择 -->
      <div class="project-select-row">
        <el-tabs v-model="chooseGroupId" :tab-position="tabPosition" class="git-tabs" @tab-change="changeGitGroup">
          <el-tab-pane v-for="(groupInfo, k) in gitGroupConfigList" :key="k" :label="groupInfo.name" :name="groupInfo.id">
            <div class="git-list">
              <template v-for="(value, key) in gitConfigList" :key="key">
                <el-radio 
                  v-if="value.git_group_id === groupInfo.id"
                  v-model="chooseGitId" 
                  :label="value.id" 
                  size="large" 
                  @change="ChangeGit(value)"
                >
                  {{ value.name }}
                </el-radio>
              </template>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>

      <!-- 操作按钮 -->
      <div class="control-row">
        <div class="action-buttons">
          <el-button v-loading="btnLoading.pull" type="primary" plain @click="GitPullBranchOrigin">
            <el-icon><Download /></el-icon>拉取
          </el-button>
          <el-button v-loading="btnLoading.status" type="primary" plain @click="GitQueryStatus">
            <el-icon><View /></el-icon>状态
          </el-button>
          <el-button v-loading="btnLoading.query" type="primary" plain @click="queryCurrentBranch">
            <el-icon><InfoFilled /></el-icon>当前分支
          </el-button>
          <el-button v-loading="btnLoading.queryLog" type="primary" plain @click="queryCommitLog">
            <el-icon><Document /></el-icon>日志
          </el-button>
        </div>
        
        <div class="branch-input-group">
          <el-input 
            v-if="showChangeBranch" 
            ref="inputBranchName" 
            v-model="BranchName" 
            placeholder="请输入分支名"
            class="branch-input"
            @keyup.enter="GitChangeBranch"
          ></el-input>
          <el-button v-loading="btnLoading.change" type="warning" plain @click="GitChangeBranch">
            <el-icon><Switch /></el-icon>{{ showChangeBranch ? '确认切换' : '切换分支' }}
          </el-button>
        </div>

        <div class="more-actions-group">
          <el-button type="primary" plain @click="drawerVisibleMarkdown = true">
            <el-icon><QuestionFilled /></el-icon>帮助
          </el-button>
          <el-dropdown @command="handleDropdownCommand">
            <el-button type="info" plain>
              更多操作<el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="changeBranchRemote">关联远程分支切换</el-dropdown-item>
                <el-dropdown-item command="viewGitConfig">查看 git config</el-dropdown-item>
                <el-dropdown-item command="saveCredentials">保存账号密码配置</el-dropdown-item>
                <el-dropdown-item command="setSafe">设置目录安全</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>

        <el-input
          v-if="showChangeBranchRemote"
          ref="inputBranchNameRemote"
          v-model="BranchNameRemote"
          placeholder="请输入远程分支名"
          class="branch-input remote-input"
          @keyup.enter="handleChangeBranchRemote"
        ></el-input>
      </div>
    </div>

    <!-- 输出窗口 -->
    <div class="output-card">
      <div class="output-header">
        <svg class="output-icon" viewBox="0 0 24 24" fill="none">
          <rect x="2" y="3" width="20" height="14" rx="2" stroke="currentColor" stroke-width="2"/>
          <path d="M8 21h8M12 17v4" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg>
        <span>执行输出</span>
      </div>
      <div ref="outputContent" class="output-content">
        <shellResult 
          ref="shellRef" 
          :divHeight="shellController.divHeight" 
          :isRunning="shellController.isRunning" 
          :shellShowResult="shellController.sshResult" 
          :show-model="shellController.showModel"
        ></shellResult>
      </div>
    </div>

    <el-drawer
      v-model="drawerVisibleMarkdown"
      direction="rtl"
      size="90%"
      title="文档"
    >
      <Markdown v-if="drawerVisibleMarkdown" :markdownType="markdownType"></Markdown>
    </el-drawer>
  </div>
</template>

<script>
import { Download, View, InfoFilled, Document, Switch, QuestionFilled, ArrowDown } from '@element-plus/icons-vue';
import git from '../utils/base/git.js'
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
    Download,
    View,
    InfoFilled,
    Document,
    Switch,
    QuestionFilled,
    ArrowDown,
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
    _that.calculateOutputDivHeight()
    _that.test()
  },
  activated: function () {
    let _that = this
    setTimeout(function () {
      _that.calculateOutputDivHeight()
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
    calculateOutputDivHeight: function () {
      let _that = this
      _that.$nextTick(function () {
        const outputContent = _that.$refs.outputContent
        if (!outputContent) {
          return
        }
        const rect = outputContent.getBoundingClientRect()
        const viewportHeight = window.innerHeight || document.documentElement.clientHeight
        const safeBottomSpace = 12
        const nextHeight = Math.max(viewportHeight - rect.top - safeBottomSpace, 220)
        _that.shellController.divHeight = nextHeight
      })
    },
    test: function () {
    },
    handleDropdownCommand(command) {
      switch (command) {
        case 'changeBranchRemote':
          this.handleChangeBranchRemote();
          break;
        case 'viewGitConfig':
          this.drawerVisibleMarkdown = true;
          this.markdownType = 'git-config';
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
        this.calculateOutputDivHeight()
        setTimeout(async function () {
          _that.$refs.inputBranchNameRemote?.focus()
        }, 500)
        return
      }
      _that.btnLoading.changeRemote = true
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitChangeBranchRemote(_that.selectGitConfig, _that.BranchNameRemote, function (response) {
            _that.showChangeBranchRemote = false
            _that.calculateOutputDivHeight()
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
        const maxLen = 50000;
        if (_that.shellController.sshResult.length > maxLen) {
          _that.shellController.sshResult = _that.shellController.sshResult.slice(-maxLen);
        }
        let result = format.formatResult(
            _that.shellController.sshResult, ['copy', 'color', 'replace']);
        result = format.formatResult(result, ['length']);
        _that.shellController.sshResult = result;
      });
      sseDistribute.RegisterReceive(_that.sse_distribute_id, function (msg, msgType, sseDistributeId) {
        throttleStringFunc.update(msg)
      })
    },
    chooseDefault: function () {
      let _that = this
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
        _that.calculateOutputDivHeight()
      });
    },
    ChangeGit: function (selectGitConfig) {
      let _that = this
      _that.shellController.sshResult = '';
      _that.selectGitConfig = selectGitConfig
      _that.chooseGitId = selectGitConfig.id
      _that.queryCurrentBranch()
      _that.calculateOutputDivHeight()
      git.GitLocalSetLastGitId(_that.selectGitConfig.id)
    },
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
    changeGitGroup: function () {
      let _that = this
      git.GitLocalSetLastGroupId(_that.chooseGroupId)
      if (_that.gitConfigList.length === 0) {
        return
      }
      _that.ChangeGit(_that.gitConfigList[0])
    },
    queryCurrentBranch: function () {
      let _that = this
      _that.showChangeBranch = false
      _that.showChangeBranchRemote = false
      _that.calculateOutputDivHeight()
      _that.btnLoading.query = true
      _that.tryReconnectionSocket()
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitCurrentBranch(_that.selectGitConfig, function (response) {
        setTimeout(function () {
          _that.btnLoading.query = false
        }, 500)
      })
    },
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
        this.calculateOutputDivHeight()
        setTimeout(async function () {
          _that.$refs.inputBranchNameRemote?.focus()
        }, 500)
        return
      }
      _that.btnLoading.changeRemote = true
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitChangeBranchRemote(_that.selectGitConfig, _that.BranchNameRemote, function (response) {
            _that.showChangeBranchRemote = false
            _that.calculateOutputDivHeight()
            setTimeout(function () {
              _that.btnLoading.changeRemote = false
            }, 500)
          }
      )
    },
    GitChangeBranch: function () {
      let _that = this
      if (!this.showChangeBranch) {
        this.showChangeBranch = true
        this.calculateOutputDivHeight()
        setTimeout(async function () {
          _that.$refs.inputBranchName?.focus()
        }, 500)
        return
      }
      _that.btnLoading.change = true
      _that.selectGitConfig.sse_distribute_id = _that.sse_distribute_id
      git.GitChangeBranch(_that.selectGitConfig, _that.BranchName, function (response) {
            _that.showChangeBranch = false
            _that.calculateOutputDivHeight()
            setTimeout(function () {
              _that.btnLoading.change = false
            }, 500)
          }
      )
    },
  },
}
</script>

<style>
/* 页面容器 */
.git-page-container {
  padding: 0;
  width: 100%;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

/* 顶部卡片样式 - 蓝色渐变 */
.git-header-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  padding: 20px 24px;
  margin-bottom: 16px;
  box-shadow: 0 4px 20px rgba(102, 126, 234, 0.25);
  flex-shrink: 0;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #fff;
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 16px;
}

.header-icon {
  width: 28px;
  height: 28px;
  color: #fff;
}

/* 项目选择 */
.project-select-row {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 12px;
  padding: 12px 16px;
  margin-bottom: 16px;
}

.git-tabs :deep(.el-tabs__header) {
  margin-bottom: 8px;
}

.git-tabs :deep(.el-tabs__item) {
  font-size: 14px;
  color: #606266;
}

.git-tabs :deep(.el-tabs__item.is-active) {
  color: #667eea;
  font-weight: 600;
}

.git-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: #ebeef5;
}

.git-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.git-list :deep(.el-radio__label) {
  color: #303133;
}

.git-list :deep(.el-radio__input.is-checked .el-radio__inner) {
  border-color: #667eea;
  background: #667eea;
}

.git-list :deep(.el-radio__input.is-checked + .el-radio__label) {
  color: #667eea;
}

/* 控制行 */
.control-row {
  display: flex;
  gap: 16px;
  align-items: center;
  flex-wrap: wrap;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-buttons .el-button {
  background: rgba(255, 255, 255, 0.9);
  border: none;
}

.action-buttons .el-button:hover {
  background: #fff;
}

.action-buttons .el-button--primary {
  color: #667eea;
}

.action-buttons .el-button--primary:hover {
  color: #764ba2;
}

.branch-input-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

.branch-input {
  width: 180px;
}

.branch-input :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.branch-input-group .el-button--warning {
  background: rgba(255, 255, 255, 0.9);
  border: none;
  color: #e6a23c;
}

.branch-input-group .el-button--warning:hover {
  background: #fff;
  color: #667eea;
}

.more-actions-group {
  display: flex;
  gap: 8px;
}

.more-actions-group .el-button {
  background: rgba(255, 255, 255, 0.9);
  border: none;
}

.more-actions-group .el-button:hover {
  background: #fff;
}

.more-actions-group .el-button--primary {
  color: #667eea;
}

.more-actions-group .el-button--info {
  color: #909399;
}

.remote-input {
  width: 180px;
}

.remote-input :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 输出窗口 */
.output-card {
  flex: 1;
  min-height: 0;
  height: 100%;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.output-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  background: linear-gradient(135deg, #434343 0%, #000000 100%);
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;
}

.output-icon {
  width: 22px;
  height: 22px;
  color: #fff;
}

.output-content {
  flex: 1;
  overflow: hidden;
  background: #f8f9fa;
}

/* 响应式 */
@media (max-width: 1200px) {
  .control-row {
    flex-direction: column;
    align-items: stretch;
  }
  
  .action-buttons,
  .branch-input-group,
  .more-actions-group {
    flex-wrap: wrap;
  }
  
  .branch-input {
    width: 100%;
    max-width: 200px;
  }
}

@media (max-width: 768px) {
  .git-header-card {
    padding: 16px;
  }
  
  .header-title {
    font-size: 18px;
  }
  
  .action-buttons {
    justify-content: center;
  }
}
</style>
