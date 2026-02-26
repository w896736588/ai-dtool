<template>
  <div id="mainCard" ref="mainCard" class="box-card" style="height: 100%;">
    <div style="margin-bottom: 10px;">
      <el-button type="primary" @click="createVariableDirectory">
        创建脚本
      </el-button>
      <el-button type="primary" @click="drawerVisibleMarkdown = true">帮助文档</el-button>&nbsp;
    </div>
    <el-tabs v-model="chooseVariableId" class="demo-tabs" tab-position="left" @tabChange="changeVariableTab">
      <template v-for="(variableVal, key) in variableList" :key="key" class="scrollbar-demo-item">
        <el-tab-pane :label="variableVal.name" :name="variableVal.id">
          <el-row>
            <el-col :span="12">
              <div class="grid-content bg-purple">
                <el-tabs v-model="chooseVariable.activeCmdTab" class="demo-tabs1" :style="{ height: shellController.divHeight - 10 + 'px' }" >
                  <el-tab-pane label="执行" name="run">
                    <el-alert :closable="false" show-icon :title="variableVal.desc" type="info" v-if="variableVal.desc !== ''"/>
                    <el-alert :closable="false" show-icon title="每个选项和输入框都只支持选择一次，暂不支持多次选择，如果选错了点击重置" type="info"/>
                    <el-form label-width="auto">
                      <template v-for="(value,key1) in run_form_list" :key="key1">
                        <el-form-item v-if="value.CmdType === '3'" :label="value.Input.Label">
                          <el-input v-model="value.Input.Value" :disabled="value.disabled" style="width:80%;margin-right: 5px;"/>
                          <el-button v-if="!value.disabled" type="primary" @click="cmdSet(value.Id , value.Input.Value)">
                            确认
                          </el-button>
                        </el-form-item>
                        <el-form-item v-if="value.CmdType === '17'" :label="value.Input.Label">
                          <el-input v-model="value.Input.Value" :disabled="value.disabled" :rows="5" style="width:80%;margin-right: 5px;" type="textarea"/>
                          <el-button v-if="!value.disabled" type="primary" @click="cmdSet(value.Id , value.Input.Value)">
                            确认
                          </el-button>
                        </el-form-item>
                        <el-form-item v-if="(value.CmdType === '9' || value.CmdType === '12' || value.CmdType === '14')" :label="value.Select.Label">
                          <el-radio-group v-model="value.Select.Value">
                            <template v-for="(optionValue,optionKey) in value.Select.OptionList" :key="optionKey">
                              <el-radio :disabled="value.disabled" :value="optionValue.Value" @change="cmdSet(value.Id , value.Select.Value)">
                                {{ optionValue.Label }}
                              </el-radio>
                            </template>
                          </el-radio-group>
                        </el-form-item>
                        <!--                        </template>-->
                      </template>
                      <div class="button-container">
                        <el-button v-loading="chooseVariable.isRunning" :disabled="is_run === 0 || is_finish === 1" type="primary" @click="cmdRun">
                          执 行
                        </el-button>
                        <el-button type="primary" @click="refreshRun">
                          重 置
                        </el-button>
                      </div>
                    </el-form>

                  </el-tab-pane>
                  <el-tab-pane label="编辑" name="edit">
                    <el-alert :closable="false" show-icon title="编辑后刷新页面，暂时没错自动刷新" type="info"/>
                    <el-form label-width="auto">
                      <el-form-item label="名称">
                        <el-input v-model="chooseVariable.name"/>
                      </el-form-item>
                      <el-form-item label="描述">
                        <el-input type="textarea" v-model="chooseVariable.desc" rows="3"/>
                      </el-form-item>
                      <el-form-item>
                        <el-button size="small" type="primary" @click="VariableAdd">保存</el-button>
                        <el-button size="small" type="primary" @click="showAddVariableCmd(default_variable_cmd)">
                          创建命令
                        </el-button>

                        <el-popconfirm
                            cancel-button-text="取消"
                            confirm-button-text="删除"
                            icon-color="#626AEF"
                            title="确定删除吗?"
                            @confirm="VariableDelete"
                        >
                          <template #reference>
                            <el-button size="small" type="danger">删除</el-button>
                          </template>
                        </el-popconfirm>
                      </el-form-item>
                    </el-form>
                    <div class="demo-collapse" style="height:100%;font-size:14px;">
                      <div class="flex flex-wrap gap-4">
                        <el-card v-for="(variable_cmd,key) in chooseVariable.variable_cmd_list" :key="key" shadow="hover" style="">
                          <el-tag type="primary">{{ typeNameMap['type' + variable_cmd.type] }}</el-tag>&nbsp;
                          <el-tag type="warning">{{runTypeName[variable_cmd.run_type]}}</el-tag>
                          {{ variable_cmd.name }}
                          <span v-if="variable_cmd.type === '10' && variable_cmd.options !== ''">账号密码：{{
                              variable_cmd.options
                            }}</span>
                          <span v-if="variable_cmd.result_key !== ''" style="cursor: copy;" @click="copyVariable(variable_cmd.result_key)"> {{
                              variable_cmd.result_key
                            }}</span>
                          <a v-if="variable_cmd.type === '10'" :href="variable_cmd.remark" style="margin-left: 10px;" target="_blank">{{
                              variable_cmd.remark
                            }}</a>
                          <div style="display: inline-block;float: right;" class="weight-input">
                            <el-input link type="text" v-model="variable_cmd.weight" @blur="VariableCmdAdd(variable_cmd)" style="display: inline;width:20px;"></el-input>
                            <el-button link type="primary" @click="showAddVariableCmd(variable_cmd)">编辑
                            </el-button>
                            <el-popconfirm
                                cancel-button-text="取消"
                                confirm-button-text="删除"
                                icon-color="#626AEF"
                                title="确定删除吗?"
                                @confirm="VariableCmdDel(variable_cmd)"
                            >
                              <template #reference>
                                <el-button link type="danger">删除</el-button>
                              </template>
                            </el-popconfirm>
                          </div>

                        </el-card>
                      </div>
                    </div>
                  </el-tab-pane>
                </el-tabs>
                <p></p>
                <p></p>
              </div>
            </el-col>
            <el-col :span="12">
              <shellResult ref="shellRef" :divHeight="shellController.divHeight" :isRunning="shellController.isRunning" :shellShowResult="shellController.sshResult" :show-model="shellController.showModel"></shellResult>
            </el-col>
          </el-row>
        </el-tab-pane>
      </template>
    </el-tabs>

    <!--新增命令弹窗-->
    <el-dialog v-model="dialogVariableCmd" :append-to-body="true" title="命令" width="1000">
      <el-form label-width="auto" style="max-width: 1000px">
        <el-form-item label="名称">
          <el-input v-model="variable_cmd.name"/>
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="variable_cmd.type" placeholder="选择类型" @change="changeVariableCmdType(variable_cmd)">
            <template v-for="(value,key) in typeList" :key="key">
              <el-option :label="value.label" :value="value.value"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item label="执行类型">
          <el-select v-model="variable_cmd.run_type" placeholder="选择执行类型">
            <template v-for="(value,key) in runTypeList" :key="key">
              <el-option :label="value.label" :value="value.value"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item label="排序">
          <el-input v-model="variable_cmd.weight" type="text"/>
        </el-form-item>
        <el-form-item v-if="variable_cmd.type === `1`" label="sql">
          <el-alert :closable="false" show-icon title="定义数据库 [id=xx] 换行" type="info"/>
          <el-input v-model="variable_cmd.sql" class="text-bash" rows="10" type="textarea"></el-input>
        </el-form-item>
        <el-form-item v-if="variable_cmd.type === `9`" label="选项值">
          <JsonEditCombine
              v-if="dialogVariableCmd"
              :value="variable_cmd.options"
              mode="tree"
              style="width: 100%;"
              @change="optionsChange"
          />

        </el-form-item>
        <el-form-item v-if="variable_cmd.type === `2`" label="命令">
          <el-alert :closable="false" show-icon title="定义ssh [id=xx] 换行" type="info"/>
          <el-input v-model="variable_cmd.cmd" rows="5" style="overflow-x:scroll;" type="textarea"/>
        </el-form-item>
        <el-form-item v-if="variable_cmd.type !== `10`" label="输出的key">
          <el-input v-model="variable_cmd.result_key" type="text"/>
        </el-form-item>

        <el-form-item v-if="variable_cmd.type === `15`" label="选择大模型">
          <el-select v-model="variable_cmd.smart_link_id" placeholder="选择链接" @change="changeVariableCmdSmartLink">
            <template v-for="(value,key) in smartLinkList" :key="key">
              <el-option :label="value.name" :value="value.id"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item v-if="variable_cmd.type === `15`" label="注册的链接">
          <el-alert :closable="false" show-icon title="定义链接 [id=xx] 换行" type="info"/>
          <JsonEditCombine
              v-if="dialogVariableCmd"
              :value="variable_cmd.options"
              mode="tree"
              style="width: 100%;"
              @change="optionsChange"
          />
        </el-form-item>
        <el-form-item v-if="variable_cmd.type === `15`" label="选择label">
          <el-select v-model="variable_cmd.smart_link_label" placeholder="选择">
            <template v-for="(value,key) in smartLinkLabelList" :key="key">
              <el-option :label="value.label" :value="value.label"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item v-if="variable_cmd && variable_cmd.type === '8'" label="脚本">
          <el-alert :closable="false" show-icon title="定义ssh [id=xx] 换行" type="info"/>
          <el-input v-model="variable_cmd.bash" class="text-bash" rows="10" type="textarea"></el-input>
        </el-form-item>
        <el-form-item v-if="variable_cmd && variable_cmd.type === '18'" label="linux命令">
          <el-alert :closable="false" show-icon title="定义ssh [id=xx] 多个命令换行" type="info"/>
          <el-input v-model="variable_cmd.bash" class="text-bash" rows="10" type="textarea"></el-input>
        </el-form-item>
        <el-form-item v-if="variable_cmd && variable_cmd.type === '19'" label="Bat">
          <el-input v-model="variable_cmd.bash" class="text-bash" rows="10" type="textarea"></el-input>
        </el-form-item>
        <el-form-item v-if="variable_cmd && variable_cmd.type === '20'" label="上传定义">
          <el-input v-model="variable_cmd.bash" class="text-bash" rows="10" type="textarea"></el-input>
        </el-form-item>
        <el-form-item v-if="variable_cmd && variable_cmd.type === '5'" label="Curl">
          <JsonEditCombine
              v-if="dialogVariableCmd"
              :value="variable_cmd.options"
              mode="tree"
              style="width: 100%;"
              @change="optionsChange"
          />
        </el-form-item>
        <el-form-item v-if="variable_cmd && variable_cmd.type === '16'" label="内容收集">
          <el-input v-model="variable_cmd.options" rows="10" type="textarea"></el-input>
        </el-form-item>
        <el-form-item v-if="variable_cmd && variable_cmd.type === '11'" label="删除的key">
          <el-input v-model="variable_cmd.bash" class="text-bash" rows="4" type="textarea"></el-input>
        </el-form-item>
        <el-form-item v-if="variable_cmd && variable_cmd.type === '10'" label="跳转地址">
          <el-input v-model="variable_cmd.remark" class="text-bash" rows="3" type="textarea"></el-input>
        </el-form-item>
        <el-form-item v-if="variable_cmd.type === `10`" label="账号密码">
          <el-input v-model="variable_cmd.options" class="text-bash" rows="3" type="textarea"></el-input>
        </el-form-item>
        <el-form-item label="运行判断">
          <el-alert :closable="false" show-icon title=" = != in not in,四种选项左右必须加空格" type="info"/>
          <el-input v-model="variable_cmd.checks" rows="3" type="textarea"></el-input>
        </el-form-item>
        <el-form-item label="默认值">
          <el-input v-model="variable_cmd.default" type="text"></el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVariableCmd = false">取 消</el-button>
        <el-button type="primary" @click="VariableCmdAdd">确 定</el-button>
      </template>
    </el-dialog>

    <!--新增脚本弹窗-->
    <el-dialog v-model="dialogVariable" :append-to-body="true" title="创建合集" width="600">
      <el-form label-width="auto" style="max-width: 600px">
        <el-form-item label="名称">
          <el-input v-model="variable_add.name"/>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVariable = false">取 消</el-button>
        <el-button type="primary" @click="VariableCreate">确 定</el-button>
      </template>
    </el-dialog>

    <!--新增脚本弹窗-->
    <el-dialog v-model="dialogLoginUserName" :append-to-body="true" title="输入账号密码" width="600">
      <el-form label-width="auto" style="max-width: 600px">
        <el-form-item label="账号">
          <el-input v-model="LoginUsername"/>
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="LoginPassword"/>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogLoginUserName = false">取 消</el-button>
        <el-button type="primary" @click="VariableSetLogin">确 定</el-button>
      </template>
    </el-dialog>
  </div>
  <el-drawer
      v-model="drawerVisibleMarkdown"
      title="文档"
      direction="rtl"
      size="90%"
  >
    <Markdown v-if="drawerVisibleMarkdown" :markdownType="markdownType"></Markdown>
  </el-drawer>
</template>
<style scoped>
.button-container {
  justify-content: center; /* 将按钮水平居中 */
  gap: 8px; /* 可选：调整按钮之间的间距 */
  margin-top: 10px;
}

</style>
<script>
import base from '../utils/base'
import shell from "@/utils/base/shell"
import shellResult from "@/components/shell/result_markdown.vue";
import variable from "@/utils/base/variable_set"
import redisHashList from "@/components/redis/tableHash.vue";
import notify from "@/utils/base/notify"
import copy from "@/utils/base/copy"
import choose from "@/utils/base/choose"
import "codemirror/mode/javascript/javascript.js"
import Codemirror from "codemirror-editor-vue3"
import 'codemirror/lib/codemirror.css';
import smartLink from "@/utils/base/smart_link_set"
import sse from "@/utils/base/sse";
import a from "@/utils/base/array"
import t from "@/utils/base/type";
import JsonEditCombine from "@/components/base/json_edit_combine.vue"
import Markdown from "@/components/Markdown.vue";
import sseDistribute from "@/utils/base/sse_distribute";

export default {
  props: {},
  components: {
    Markdown,
    shellResult,
    JsonEditCombine,
  },
  data() {
    return {
      resizeHandler: null,
      sseId : '',
      drawerVisibleMarkdown : false,
      markdownType: 'Variable',
      shellController: {
        sshResult: '',
        sourceSshResult :'',
        isRunning: false,
        showModel: 'markdown',
        divHeight: 330,
      },
      cmOptionList: {
        sql: {
          mode: 'text/x-sql'
        },
        bash: {
          mode: "shell"
        },
      },
      typeNameMap: {
        type1: 'Mysql执行',
        type3: '输入框',
        type5: 'Curl',
        type8: 'Bash脚本',
        type9: '单项选择',
        type11: 'Redis操作',
        type15: '自动化链接',
        type16: '内容收集',
        type17: '输入框(textarea)',
        type18: 'linux命令',
        type19: 'Bat',
        type20: '上传文件',
      },
      typeList: [
        {"label": "Mysql执行", "value": "1"},
        {"label": "输入框", "value": "3"},
        {"label": "Curl", "value": "5"},
        {"label": "Bash脚本", "value": "8"},
        {"label": "单项选择", "value": "9"},
        {"label": "Redis操作", "value": "11"},
        {"label": "自动化链接", "value": "15"},
        {"label": "内容收集", "value": "16"},
        {"label": "输入框(textarea)", "value": "17"},
        {"label": "linux命令", "value": "18"},
        {"label": "Bat", "value": "19"},
        {"label": "上传文件", "value": "20"},
      ],
      runTypeList: [
        {"label": "输出表单", "value": "form"},
        {"label": "中间层", "value": "middle"},
        {"label": "最终执行", "value": "run"},
      ],
      runTypeName : {
        form: '输出表单',
        middle: '中间层',
        run: '最终执行',
      },
      activeNames: [
        1
      ],
      dialogLoginUserName : false,
      LoginUsername : '',
      LoginPassword: '',
      chooseVariableId: '-1',
      chooseVariable: {},
      scrollHeight: 600,
      filterValue: '',
      variable: {
        name: '',
        desc : '',
      },
      variable_add: {
        name: '',
        type: '1',
      },
      variable_cmd: {
        id: '',
        name: '',
        type: '',
        variable_id: '',
        sql: '',
        remark: '',
        bash: '',
        input_key: '',
        weight: '',
        result_key: '',
        options: '',
        checks: '',
        default: '',
        smart_link_id: '',
        smart_link_label: '',
        run_type : 0,
      },
      default_form: {
        name: '',
      },
      default_variable_cmd: {
        id: '',
        name: '',
        type: '',
        variable_id: '',
        sql: '',
        remark: '',
        bash: '',
        input_key: '',
        weight: '',
        result_key: '',
        options: '',
        checks: '',
        default: '',
        smart_link_id: '',
        smart_link_label: '',
        run_type : 0,
      },
      dialogVariable: false,
      dialogVariableCmd: false,
      name: 'Variable',
      tabPosition: 'top',
      //按钮状态
      btnLoading: {
        exec: false,
        pull: false,
        change: false,
        status: false,
        query: false,
      },
      storeKey: 'variable:id',
      variableList: [],
      //执行参数
      dialogRun: false,
      chooseVariableIndex: 0,
      smartLinkList: [],
      smartLinkLabelList: [],

      is_run: 0,
      is_finish: 0,
      run_cmd_id: 0,
      replace_list: {},
      run_form_list: [],

      replaceRegex : '上传进度:\\s+\\d+%\\s+\\(\\d+\\/\\d+\\s+bytes\\)',
    }
  },
  mounted: function () {
    let _that = this
    _that.GetConfigList(true)
    smartLink.SmartLinkList(function (response) {
      _that.smartLinkList = response.Data.smart_link_list
      _that.changeVariableCmdSmartLink()
    })
    _that.bindWindowResize()
    _that.$nextTick(function () {
      _that.updateLayoutHeight()
      window.requestAnimationFrame(function () {
        _that.updateLayoutHeight()
      })
      setTimeout(function () {
        _that.updateLayoutHeight()
      }, 120)
    })
    _that.sse_distribute_id = sseDistribute.GetSseDistributeId('variable')
    _that.sseCreate()
  },
  activated: function () {
    let _that = this
    _that.$nextTick(function () {
      _that.updateLayoutHeight()
      window.requestAnimationFrame(function () {
        _that.updateLayoutHeight()
        _that.scrollOutputToBottom()
      })
    })
  },
  beforeUnmount: function () {
    let _that = this
    _that.unbindWindowResize()
    if (_that.sse_distribute_id) {
      sseDistribute.UnRegisterReceive(_that.sse_distribute_id)
    }
  },
  methods: {
    updateLayoutHeight: function () {
      let _height = parseInt(base.GetDivHeight2())
      if (!Number.isFinite(_height)) {
        _height = window.innerHeight - 120
      }
      let nextHeight = _height - 60
      this.shellController.divHeight = Math.max(260, nextHeight)
    },
    bindWindowResize: function () {
      let _that = this
      if (_that.resizeHandler) {
        return
      }
      _that.resizeHandler = function () {
        _that.updateLayoutHeight()
      }
      window.addEventListener('resize', _that.resizeHandler)
    },
    unbindWindowResize: function () {
      if (!this.resizeHandler) {
        return
      }
      window.removeEventListener('resize', this.resizeHandler)
      this.resizeHandler = null
    },
    scrollOutputToBottom: function () {
      this.$nextTick(function () {
        const outputContainer = document.getElementById('showShellResult')
        if (!outputContainer) {
          return
        }
        outputContainer.scrollTop = outputContainer.scrollHeight
      })
    },
    optionsChange: function (newData) {
      let _that = this
      _that.windowChange()
      _that.variable_cmd.options_new = newData
    },
    bashChange: function (newData) {
      let _that = this
      _that.variable_cmd.bash_new = newData
    },
    changeVariableTab: function (changeToRun) {
      let _that = this
      _that.windowChange()
      _that.cleanRun()
      for (let j in this.variableList) {
        if (_that.variableList[j].id === this.chooseVariableId) {
          _that.chooseVariable = _that.variableList[j]
        }
      }
      this.showVariable(changeToRun)
    },
    cleanRun : function (){
      let _that = this
      _that.windowChange()
      _that.is_run = 0
      _that.is_finish = 0
      _that.run_cmd_id = 0
      _that.replace_list = {}
      _that.run_form_list = []
      _that.shellController.sshResult = ''
      _that.shellController.sourceSshResult = ''
      sseDistribute.UnRegisterReceive(_that.sse_distribute_id)
      _that.sse_distribute_id = sseDistribute.GetSseDistributeId('variable')
      _that.sseCreate()
    },
    refreshRun: function () {
      let _that = this
      _that.cleanRun()
      _that.cmdRun() //重新运行
    },
    cmdRun: function () { //准备运行
      let _that = this
      _that.dialogRun = true
      _that.chooseVariable.isRunning = true
      variable.VariableRun(
          _that.sse_distribute_id,
          _that.chooseVariableId,
          parseInt(_that.run_cmd_id),
          _that.is_run,
          JSON.stringify(_that.replace_list),
          function (response) {
            _that.chooseVariable.isRunning = false
            if (response.ErrCode === 0) {
              if(parseInt(response.Data.VariableId) !== parseInt(_that.chooseVariableId)){
                return
              }
              if (response.Data.RunStatus === 0) { //不处理最终执行
                _that.run_form_list.push(response.Data.Form)
                _that.run_cmd_id = response.Data.Form.Id
                _that.replace_list = response.Data.ReplaceList
                //如果是单选 并且只有一个选项 那么直接运行
                if(response.Data.Form.CmdType === '9' && response.Data.Form.Select.OptionList.length === 1){
                  setTimeout(function () {
                    _that.cmdSet(_that.run_cmd_id , response.Data.Form.Select.OptionList[0].Value)
                    for  (let i = 0; i < _that.run_form_list.length; i++) {
                      if(_that.run_form_list[i].Id === _that.run_cmd_id){
                        _that.run_form_list[i].Select.Value = response.Data.Form.Select.OptionList[0].Value
                      }
                    }
                  },400)
                }
              } else if (response.Data.RunStatus === 1) {//可以最终执行
                _that.is_run = 1
                _that.run_cmd_id = response.Data.Form.Id
                _that.replace_list = response.Data.ReplaceList
              } else if (response.Data.RunStatus === 2) {//结束
                _that.is_finish = 1
              }
            }
          })

    },
    cmdSet: function (cmd_id, edit_value) { //准备运行
      let _that = this
      _that.dialogRun = true
      _that.chooseVariable.isRunning = true
      variable.VariableSet(
          _that.chooseVariableId,
          parseInt(cmd_id),
          JSON.stringify(_that.replace_list),
          edit_value,
          function (response) {
            _that.chooseVariable.isRunning = true
            if (response.ErrCode === 0) {
              //禁用
              for (let j in _that.run_form_list) {
                if(parseInt(_that.run_form_list[j].Id) === parseInt(cmd_id)){
                  _that.run_form_list[j].disabled = true
                }
              }
              if (response.Data.RunStatus === 2) {
                _that.is_finish = 1
              }else {
                _that.replace_list = response.Data.ReplaceList
                _that.run_cmd_id = response.Data.Form.Id
                //继续下一步
                _that.cmdRun()
              }
            }
          })

    },
    sseCreate: function () {
      let _that = this
      sseDistribute.RegisterReceive(_that.sse_distribute_id , function (msg,msgType,sseDistributeId){
        if(msg === sse.SseEventClean){
          _that.shellController.sshResult = ''
          _that.shellController.sourceSshResult = '';
        }else if(msg === sse.SseEventLogin){
          _that.dialogLoginUserName = true
        }else if(msg.startsWith(sse.SseEventProcess)){ //准备替换
          _that.replaceRegex = msg.replace(sse.SseEventProcess, '')
        }else{
          _that.processMsg(msg)
          _that.shellController.sshResult = _that.shellController.sourceSshResult
        }
      })
    },
    processMsg : function (msg){
      let _that = this
      //处理进度
      const regReceivingObjects = new RegExp(_that.replaceRegex);

      let regList = [regReceivingObjects]
      let boolFind = false
      for (let i in regList){
        let reg = regList[i]
        let matchList = msg.match(reg); //收到的消息是否匹配到
        if(t.IsArray(matchList) && matchList.length > 0){
          let strMatchList = _that.shellController.sourceSshResult.match(reg)
          if(t.IsArray(strMatchList) && strMatchList.length > 0){
            _that.shellController.sourceSshResult = _that.shellController.sourceSshResult.replace(reg, matchList[0])
            boolFind = true
          }
        }
      }
      if(!boolFind){
        _that.shellController.sourceSshResult += msg
      }
    },
    //创建合集
    createVariableDirectory: function () {
      let _that = this
      _that.dialogVariable = true
      _that.variable_add.name = ''
      _that.variable_add.type = '1'
    },
    showAddVariableCmd: function (editValue) {
      let _that = this
      _that.dialogVariableCmd = true
      _that.variable_cmd = editValue
      _that.changeVariableCmdType(_that.variable_cmd)
      _that.changeVariableCmdSmartLink()
    },
    VariableCmdAdd: function (saveForm) {
      let _that = this
      if (saveForm && saveForm.id) {
        _that.variable_cmd = saveForm
      }
      _that.variable_cmd.variable_id = _that.chooseVariableId
      if (_that.variable_cmd.options_new) { //编辑过
        if (t.IsObjectOrArray(_that.variable_cmd.options_new)) {
          _that.variable_cmd.options = JSON.stringify(_that.variable_cmd.options_new)
        } else {
          _that.variable_cmd.options = _that.variable_cmd.options_new
        }
      }
      variable.VariableCmdAdd(_that.variable_cmd, function (response) {
        if (response.ErrCode === 0) {
          _that.dialogVariableCmd = false
          // _that.changeVariableTab()
          _that.GetConfigList(true, false)
          notify.success('编辑成功')
        }
        // _that.showVariable(_that.variable, true)
      })
    },
    VariableCmdDel: function (saveForm) {
      let _that = this
      if (saveForm && saveForm.id) {
        _that.variable_cmd = saveForm
      }
      _that.variable_cmd.variable_id = _that.chooseVariableId
      variable.VariableCmdDel(_that.variable_cmd, function (response) {
        _that.dialogVariableCmd = false
        // _that.changeVariableTab()
        //_that.showVariable(_that.variable, true)
        _that.GetConfigList(true, false)
        notify.success('编辑成功')
      })
    },
    changeVariableCmdType: function (variable_cmd) {
      let _that = this
      if (variable_cmd.type === '15') { //重新加载链接
        smartLink.SmartLinkList(function (response) {
          _that.smartLinkList = response.Data.smart_link_list
        })
      }
    },
    changeVariableCmdSmartLink: function () {
      let _that = this
      for (let i in _that.smartLinkList) {
        if (_that.smartLinkList[i].id === _that.variable_cmd.smart_link_id) {
          _that.smartLinkLabelList = JSON.parse(_that.smartLinkList[i].links)
        }
      }
    },
    copyVariable: function (key) {
      let index = copy.SetCopyContent(key)
      copy.handleCopy(index)
    },
    chooseDefault: function (changeToRun) {
      let _that = this
      let chooseData = choose.ChooseDefault(_that.storeKey, _that.variableList, _that.default_form)
      _that.chooseVariableId = chooseData.id
      _that.variable = chooseData.config
      _that.chooseVariable = chooseData.config
      //存储选中
      if (_that.chooseVariableId !== '') {
        _that.showVariable(changeToRun)
      }
    },
    //编辑
    VariableAdd: function () {
      let _that = this
      variable.VariableAdd(_that.chooseVariable, function (response) {
        if (response.ErrCode === 0) {
          notify.success('操作成功')
          _that.GetConfigList(false)
          _that.changeVariableTab(false)
        }
      })
    },
    VariableSetLogin : function (){
      let _that = this
      variable.VariableSetLogin(_that.LoginUsername , _that.LoginPassword, function (response) {
        if (response.ErrCode === 0) {
          notify.success('操作成功')
          _that.dialogLoginUserName = false
        }
      })
    },
    //新增
    VariableCreate: function () {
      let _that = this
      variable.VariableAdd(_that.variable_add, function (response) {
        if (response.ErrCode === 0) {
          notify.success('操作成功')
          _that.GetConfigList(false)
          _that.changeVariableTab()
          _that.dialogVariable = false
        }
      })
    },
    VariableDelete: function () {
      let _that = this
      variable.VariableDelete(_that.chooseVariable, function (response) {
        if (response.ErrCode === 0) {
          _that.variable = {name: ''}
          _that.GetConfigList(true)
        }
      })
    },
    windowChange: function () {
      this.updateLayoutHeight()
    },
    showVariable: function (changeToRun) {
      let _that = this
      if (changeToRun !== false) {
        _that.chooseVariable.activeCmdTab = 'run'
      }
      choose.ChooseId(_that.storeKey, _that.chooseVariableId)
      if(changeToRun !== false){
        _that.cmdRun()
      }
      _that.refreshCmRef()
    },
    refreshCmRef: function () {
      let _that = this
      if (_that.$refs && _that.$refs.cmRef) {
        for (let i in _that.$refs.cmRef) {
          if (_that.$refs.cmRef[i] && "refresh" in _that.$refs.cmRef[i]) {
            _that.$refs.cmRef[i].refresh()
          }
        }
      }
    },
    GetConfigList: function (chooseDefault, changeToRun) {
      let _that = this
      variable.VariableList(function (response) {
        if (response.ErrCode === 0) {
          _that.variableList = response.Data.variable_list
          if (chooseDefault === true) {
            _that.chooseDefault(changeToRun)
          }
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
  },
}
</script>

<style scoped>
.weight-input :deep(.el-input__inner) {
  width: 20px;
}
.demo-tabs1 {
  height: 100%;      /* 让动态 height 生效 */
  overflow-y: auto;  /* 竖向滚动 */
}

/* 防止 Element-Plus 内部 padding 撑爆 */
.demo-tabs1 :deep(.el-tabs__content) {
  height: 100%;
}
.demo-tabs1 :deep(.el-tab-pane) {
  height: 100%;
  overflow-y: auto; /* 再给 pane 也加一道保险 */
}


</style>
