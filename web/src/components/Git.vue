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
        <pl-button class="page-settings-btn" type="warning" plain @click="openGitSettings">
          <el-icon><Setting /></el-icon>设置
        </pl-button>
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
          <pl-button v-loading="btnLoading.pull" type="primary" plain @click="GitPullBranchOrigin">
            <el-icon><Download /></el-icon>拉取
          </pl-button>
          <pl-button v-loading="btnLoading.status" type="primary" plain @click="GitQueryStatus">
            <el-icon><View /></el-icon>状态
          </pl-button>
          <pl-button v-loading="btnLoading.query" type="primary" plain @click="queryCurrentBranch">
            <el-icon><InfoFilled /></el-icon>当前分支
          </pl-button>
          <pl-button v-loading="btnLoading.queryLog" type="primary" plain @click="queryCommitLog">
            <el-icon><Document /></el-icon>日志
          </pl-button>
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
          <pl-button v-loading="btnLoading.change" type="warning" plain @click="GitChangeBranch">
            <el-icon><Switch /></el-icon>{{ showChangeBranch ? '确认切换' : '切换分支' }}
          </pl-button>
        </div>

        <div class="more-actions-group">
          <pl-button type="primary" plain @click="drawerVisibleMarkdown = true">
            <el-icon><QuestionFilled /></el-icon>帮助
          </pl-button>
          <el-dropdown @command="handleDropdownCommand">
            <pl-button type="info" plain>
              更多操作<el-icon class="el-icon--right"><ArrowDown /></el-icon>
            </pl-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="changeBranchRemote">关联远程分支切换</el-dropdown-item>
                <el-dropdown-item v-loading="btnLoading.groupBranch" command="groupBranches">查看当前组全部分支</el-dropdown-item>
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

    <SettingsDialog
      v-model="gitSettingsVisible"
      title="Git 设置"
      width="82%"
      @closed="refreshGitAfterSettingsClose"
    >
      <GitSettingPage @changed="handleGitSettingsChanged" />
    </SettingsDialog>
  </div>
</template>

<script>
import { Download, View, InfoFilled, Document, Switch, QuestionFilled, ArrowDown, Setting } from '@element-plus/icons-vue';
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
import SettingsDialog from '@/components/base/SettingsDialog.vue'
import GitSettingPage from '@/components/set/git.vue'

export default {
  props: {},
  components: {
    Markdown,
    SettingsDialog,
    GitSettingPage,
    shellResult,
    Download,
    View,
    InfoFilled,
    Document,
    Switch,
    QuestionFilled,
    ArrowDown,
    Setting,
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
      gitSettingsVisible: false,
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
        queryLog: false,
        groupBranch: false,
      },
      BranchName: '', //分支名
      BranchNameRemote: '',
      gitGroupConfigList: [],
      gitConfigList: [],
      selectGitConfig: {},
      chooseGroupId: 0,
      chooseGitId: 0,
      sseId: '',
      sse_distribute_id: '',
      sseThrottleStringFunc: null,
    }
  },
  mounted: function () {
    let _that = this
    _that.prepareActionSse('init')
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
  beforeUnmount() {
    if (this.sse_distribute_id) {
      sseDistribute.UnRegisterReceive(this.sse_distribute_id)
    }
  },
  methods: {
    prepareActionSse: function (action) {
      let _that = this
      if (_that.sse_distribute_id) {
        sseDistribute.UnRegisterReceive(_that.sse_distribute_id)
      }
      _that.sse_distribute_id = sseDistribute.GetSseDistributeId(`git`)
      if (!_that.sseThrottleStringFunc) {
        _that.sseThrottleStringFunc = new Throttle_string(50, text => {
          _that.shellController.sshResult += text
          const maxLen = 50000
          if (_that.shellController.sshResult.length > maxLen) {
            _that.shellController.sshResult = _that.shellController.sshResult.slice(-maxLen)
          }
          let result = format.formatResult(
              _that.shellController.sshResult, ['copy', 'color', 'replace'])
          result = format.formatResult(result, ['length'])
          _that.shellController.sshResult = result
        })
      }
      sseDistribute.RegisterReceive(_that.sse_distribute_id, function (msg) {
        // 过滤内部解析与尾包噪声，避免在 Git 界面展示实现细节日志
        const cleanedMsg = _that.sanitizeGitSseOutput(msg)
        if (!cleanedMsg) {
          return
        }
        _that.sseThrottleStringFunc.update(cleanedMsg)
      })
      return _that.sse_distribute_id
    },
    // sanitizeGitSseOutput 清理后端内部标记与命令拼接噪声
    sanitizeGitSseOutput: function (msg) {
      let text = msg || ''
      text = text.replace(/__DT_(LOCAL|REMOTE)_BRANCH_(BEGIN|END)__/g, '')
      text = text.replace(/\n{3,}/g, '\n\n')
      return text.trim() === '' ? '' : text
    },
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
        case 'groupBranches':
          this.GitQueryGroupBranches();
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
    // openGitSettings 打开 Git 设置弹窗，在当前业务页内完成配置维护。
    // Open the Git settings modal so configuration changes happen inside the Git page.
    openGitSettings() {
      this.gitSettingsVisible = true
    },
    // handleGitSettingsChanged 配置保存成功后立即刷新 Git 页面列表与当前选中仓库。
    // Refresh the Git page immediately after settings change so the new config becomes usable.
    handleGitSettingsChanged() {
      this.GetGitConfigList()
    },
    // refreshGitAfterSettingsClose 作为兜底，在弹窗关闭时再次同步页面状态。
    // Refresh once more when the modal closes as a safety net for nested setting changes.
    refreshGitAfterSettingsClose() {
      this.GetGitConfigList()
    },
    GitSaveCredentials(){
      let _that = this
      _that.prepareActionSse('save_credentials')
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
      _that.prepareActionSse('set_safe')
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
    GitQueryGroupBranches() {
      let _that = this
      if (!_that.chooseGroupId || parseInt(_that.chooseGroupId) === 0) {
        _that.$helperNotify.error('请先选择Git分组')
        return
      }
      _that.btnLoading.groupBranch = true
      _that.prepareActionSse('query_group_branches')
      git.GitGroupBranchList({
        git_group_id: _that.chooseGroupId,
        sse_distribute_id: _that.sse_distribute_id,
      }, function (response) {
        if (response.ErrCode !== 0) {
          _that.$helperNotify.error(response.ErrMsg || '查询失败')
        } else {
          _that.$helperNotify.success('查询完成')
        }
        setTimeout(function () {
          _that.btnLoading.groupBranch = false
        }, 500)
      })
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
      _that.prepareActionSse('change_branch_remote')
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
      _that.prepareActionSse('query_current_branch')
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
      _that.prepareActionSse('query_commit_log')
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
      _that.prepareActionSse('pull_branch_origin')
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
      _that.prepareActionSse('query_status')
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
      _that.prepareActionSse('change_branch_remote')
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
      _that.prepareActionSse('change_branch')
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

<style scoped>
.git-page-container {
  padding: 0;
  width: 100%;
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  color: #4a4a4a;
}

.git-header-card {
  background: #ffffff;
  border: 1px solid #e8e8e0;
  border-radius: 12px;
  padding: 16px 18px;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #4a4a4a;
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 12px;
}

.page-settings-btn {
  margin-left: auto;
}

.header-icon {
  width: 20px;
  height: 20px;
  color: #5a8a5a;
}

.project-select-row {
  background: #fafaf7;
  border: 1px solid #edede6;
  border-radius: 10px;
  padding: 10px 12px;
  margin-bottom: 12px;
}

.git-tabs :deep(.el-tabs__header) {
  margin-bottom: 8px;
}

.git-tabs :deep(.el-tabs__item) {
  font-size: 14px;
  color: #707060;
}

.git-tabs :deep(.el-tabs__item.is-active) {
  color: #4f804f;
  font-weight: 600;
}

.git-tabs :deep(.el-tabs__active-bar) {
  background-color: #7cb87c;
}

.git-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: #ecece4;
}

.git-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
}

.git-list :deep(.el-radio) {
  margin-right: 0;
}

.git-list :deep(.el-radio__label) {
  color: #4a4a4a;
}

.git-list :deep(.el-radio__input.is-checked .el-radio__inner) {
  border-color: #6fa56f;
  background: #6fa56f;
}

.git-list :deep(.el-radio__input.is-checked + .el-radio__label) {
  color: #4f804f;
}

.control-row {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.action-buttons,
.branch-input-group,
.more-actions-group {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}

.action-buttons .el-button,
.branch-input-group .el-button,
.more-actions-group .el-button {
  border-radius: 8px;
  border: 1px solid #d8ded2;
  background: #f6f8f3;
  color: #4f804f;
}

.action-buttons .el-button:hover,
.branch-input-group .el-button:hover,
.more-actions-group .el-button:hover {
  background: #eef4ea;
  border-color: #bfd1bf;
  color: #3f6f3f;
}

.branch-input,
.remote-input {
  width: 180px;
}

.branch-input :deep(.el-input__wrapper),
.remote-input :deep(.el-input__wrapper) {
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 0 0 1px #dde3d8 inset;
}

.branch-input :deep(.el-input__wrapper.is-focus),
.remote-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #93b793 inset;
}

.output-card {
  flex: 1;
  min-height: 0;
  height: 100%;
  background: #fff;
  border: 1px solid #e8e8e0;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.output-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid #e8e8e0;
  background: #f7f7f2;
  color: #4a4a4a;
  font-size: 14px;
  font-weight: 600;
  flex-shrink: 0;
}

.output-icon {
  width: 18px;
  height: 18px;
  color: #5a8a5a;
}

.output-content {
  flex: 1;
  overflow: hidden;
  background: #fbfbf8;
}

@media (max-width: 1200px) {
  .control-row {
    align-items: stretch;
  }
}

@media (max-width: 768px) {
  .git-header-card {
    padding: 12px;
  }

  .header-title {
    font-size: 16px;
  }

  .branch-input,
  .remote-input {
    width: 100%;
  }
}
</style>

