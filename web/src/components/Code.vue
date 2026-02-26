<template>
  <div class="box-card" ref="mainCard" id="mainCard" >
    <el-row :gutter="24">
      <el-col :span="12">
        <el-form>
          <el-form-item label="模型">
            <el-radio-group v-model="form.model">
              <template v-for="(value, key) in modelList" :key="key">
                <el-radio :value="value.value">{{value.label}}</el-radio>
              </template>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="操作">
            <el-checkbox-group v-model="form.opList">
              <template v-for="(value, key) in opList" :key="key">
                <el-checkbox :value="value.value" name="op">
                  {{value.label}}
                </el-checkbox>
              </template>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item label="分表类型" v-if="state.arr.Exist('model' , form.opList)">
            <el-radio-group v-model="form.modelType">
              <template v-for="(value, key) in modelTypeList" :key="key">
                <el-radio :value="value.value">{{value.label}}</el-radio>
              </template>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="输入分表数" v-if="form.modelType.indexOf('mod') > -1">
            <el-input v-model="form.mod"></el-input>
          </el-form-item>
          <el-form-item label="输入建表sql" v-if="state.arr.Exist('model' , form.opList)">
            <el-input type="textarea" rows="10" v-model="form.sql"></el-input>
          </el-form-item>
          <el-form-item label="缓存类型" v-if="state.arr.Exist('service' , form.opList)">
            <el-radio-group v-model="form.cacheType">
              <template v-for="(value, key) in cacheTypeList" :key="key">
                <el-radio :value="value.value">{{value.label}}</el-radio>
              </template>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="缓存主字段" v-if="form.cacheType !== 'no'">
            <el-input type="text" v-model="form.main_template_field"></el-input>
          </el-form-item>
          <el-form-item label="缓存次字段" v-if="form.cacheType === 'hash_admin_custom'">
            <el-input type="text" v-model="form.child_template_field"></el-input>
          </el-form-item>
          <el-form-item label="action操作" v-if="state.arr.Exist('action' , form.opList)">
            <el-checkbox-group v-model="form.actionList">
              <template v-for="(value, key) in actionTypeList" :key="key">
                <el-checkbox :value="value.value" name="action">
                  {{value.label}}
                </el-checkbox>
              </template>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item label="其他设置">
            <el-checkbox-group v-model="form.otherSetList">
              <template v-for="(value, key) in otherSetList" :key="key">
                <el-checkbox :value="value.value" name="otherSet">
                  {{value.label}}
                </el-checkbox>
              </template>
            </el-checkbox-group>
          </el-form-item>
          <el-form-item label="接口前缀" v-if="state.arr.Exist('memo' , form.otherSetList)">
            <el-input type="text" v-model="form.actionPrefix"></el-input>
          </el-form-item>
          <el-form-item>
            <el-button v-loading="loading.run" type="primary" @click="aiRun">执行</el-button>
            <el-button>Cancel</el-button>
          </el-form-item>
        </el-form>
      </el-col>
      <el-col :span="12">
        <MarkdownRenderer :source="result"></MarkdownRenderer>
      </el-col>
    </el-row>

  </div>
</template>
<script>
import MarkdownRenderer  from '@/components/base/markdown.vue'
import arr from "@/utils/base/array"
import ai from "@/utils/base/ai"
import socket from "@/utils/base/socket";
import sse from "@/utils/base/sse"
import base from "@/utils/base";
import t from "@/utils/base/type"
export default {
  name: 'code',
  setup() {
    const state = ({
      arr: arr // 将 arr 转换为响应式对象
    });

    return { state };
  },
  data() {
    return {
      modelList : [
        {value: 'qwen2.5-coder-32b-instruct' , label : '通义千问2.5-Coder-32B'},
        {value: 'qwen2.5-coder-3b-instruct' , label : '通义千问2.5-Coder-3B'},
      ],
      opList : [
        {value: 'model' , label : 'model生成'},
        {value: 'action' , label : 'action生成'},
        {value: 'service' , label : 'service生成'},
      ],
      cacheTypeList : [
        {value: 'no' , label : '无缓存'},
        {value: 'string_single' , label : 'string单条缓存'},
        {value: 'string_all' , label : 'string多条缓存'},
        {value: 'hash_admin_custom' , label : 'hash自定义缓存'},
      ],
      actionTypeList : [
        {value: 'list' , label : '列表'},
        {value: 'detail' , label : '单个明细'},
        {value: 'create' , label : '创建及编辑'},
        {value: 'delete' , label : '删除'},
      ],
      serviceTypeList : [
        {value: 'list' , label : '列表'},
        {value: 'detail' , label : '单个明细'},
        {value: 'create' , label : '创建及编辑'},
        {value: 'delete' , label : '删除'},
      ],
      modelTypeList : [
        {value: 'no' , label : '不分表'},
        {value: 'year' , label : '按年分表'},
        {value: 'year_month' , label : '按年月分表'},
        {value: 'mod' , label : '按admin_user_id取模'},
        {value: 'year_mod' , label : '按年分表并按admin_user_id取模'},
        {value: 'year_month_mod' , label : '按年月分表并按admin_user_id取模'},
      ],
      otherSetList : [
        {value: 'memo' , label : '生成注释'},
      ],
      form : {
        model : 'qwen2.5-coder-3b-instruct',
        modelType : 'no',
        mod : '10',
        sql : 'CREATE TABLE `task_gd_list_config` (\n' +
            '  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,\n' +
            '  `admin_user_id` int(11) unsigned NOT NULL DEFAULT \'0\',\n' +
            '  `staff_user_id` int(11) unsigned NOT NULL DEFAULT \'0\' COMMENT \'哪个客服的\',\n' +
            '  `list_field_json` text NOT NULL COMMENT \'列表展示的字段 json\',\n' +
            '  `filter_field_json` text NOT NULL COMMENT \'筛选展示的项 json\',\n' +
            '  `create_time` int(11) unsigned NOT NULL DEFAULT \'0\',\n' +
            '  `update_time` int(11) unsigned NOT NULL DEFAULT \'0\',\n' +
            '  PRIMARY KEY (`id`) USING BTREE,\n' +
            '  UNIQUE KEY `idx_staff_user_id` (`staff_user_id`) USING BTREE\n' +
            ') ENGINE=InnoDB AUTO_INCREMENT=1564 DEFAULT CHARSET=utf8 COMMENT=\'小客服工单列表自定义配置（展示字段和筛选项）\';',
        cacheType : 'no',
        main_template_field : 'staff_user_id',
        child_template_field : '',
        opList : ['model' , 'service' , 'action'],
        actionList : [],
        otherSetList : ['memo'],
        actionPrefix : '',
      },
      result : '',
      socketKey : 'ai_code',
      loading : {
        run : false,
      },
    }
  },
  components: {
    MarkdownRenderer,
  },
  mounted : function() {
   this.socketInit()
  },
  methods: {
    socketInit : function (){
      let _that = this
      // let createRet = socket.SocketCreate(_that.socketKey , 500 , function (totalReceiveMsg){
      //   _that.result = totalReceiveMsg
      // })
      sse.SseCreate(_that.socketKey , 500 , function (msg){
        _that.result += msg
      })
    },
    aiRun : function (){
      let _that = this
      _that.result = ''
      _that.loading.run = true
      ai.Ai(_that.form , function (res){
        _that.loading.run = false
      })
    }
  },
}
</script>