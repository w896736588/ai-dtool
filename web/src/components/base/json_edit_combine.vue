<template>
  <div class="json-editor-combine">
    <el-radio-group v-model="showType" style="margin-left: 10px;">
      <el-radio @change="ChangeShowType(value)" v-for="(value, key) in showTypeList" :key="key" size="large" :label="value.value">
        {{ value.label }}
      </el-radio>
    </el-radio-group>
    <div v-if="showType === 'tree'">
      <Json
          :value="rawData"
          :showSource="false"
          style="width: 100%;"
          mode="tree"
          @change="change"
      />
    </div>
    <div v-if="showType === 'source'">
      <el-input
          type="textarea"
          :rows="20"
          v-model="rawData"
          @input="change"
          placeholder="请输入任意数据"
      />
    </div>

    <div v-if="showType === 'form'">
      <json_list_edit :data="rawData" @update:sourceData="change" ></json_list_edit>
    </div>

  </div>
</template>

<script>
import Json from './json'
import t from "@/utils/base/type"
import json_list_edit from "@/components/base/json_list_edit.vue";

export default {
  name: 'ObjJson',
  components: {
    json_list_edit,
    Json,
  },
  props: {
    // 原始数据
    value: {
      type: [String],
      required: true
    },
    defaultShowType : {
      type: [String],
      required: false,
      default : "auto"
    },
  },
  emits: ['change'],
  data() {
    return {
      showTypeList : [
        {value : "form" , label : "表单"},
        {value : "tree" , label : "树"},
        {value : "source" , label : "源数据"},
      ],
      showType : "",
      rawData : null,
    }
  },
  methods: {
    change : function (newData){
      let _that = this
      if(t.IsObject(newData)){
        newData = newData.config
      }
      console.log('变动了' , newData)
      _that.rawData = newData
      if(t.IsObjectOrArray(newData)){
        _that.$emit('change' , JSON.stringify(newData))
      }else{
        _that.$emit('change' , newData)
      }
    },
    ChangeShowType : function (){
      let _that= this
      console.log('切换到' , _that.showType)
      if (_that.showType === "source"){
          if(t.IsObjectOrArray(_that.rawData)){
            _that.rawData = JSON.stringify(_that.rawData)
          }
      }
    },
  },
  mounted() {
    let _that = this
    if(_that.defaultShowType === "auto"){
      let explainData = null
      try{
        explainData = JSON.parse(_that.value)
        _that.showType = "tree"
        _that.rawData = explainData
      }catch (e){
        //报错了 那肯定是一个字符串
        _that.rawData = _that.value
        _that.showType = "source"
      }
    }else{
      _that.rawData = _that.value
      _that.showType = _that.defaultShowType
    }
  }
}
</script>

<style scoped src="@/css/components/base/json_edit_combine.css"></style>