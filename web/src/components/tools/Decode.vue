<template>
  <div style="text-align: center;">
    <el-input
        style="margin-top: 10px"
        id="resultTextarea"
        placeholder="输入内容"
        type="textarea"
        v-model="codeStr"
        rows="5"
        @input="exec"
    ></el-input>
    <div style="text-align: center;margin-top:5px;" class="box-card">
      <el-input style="width: 300px; margin-right: 10px;margin-left:10px;" v-model="searchKey" @input="searchList" placeholder="请输入搜索 多个搜索以空格分割"></el-input>
    <span>({{searchLineNum}})</span>
    </div>
    <div style="text-align: center;margin-top:5px;" class="box-card">
      <el-checkbox v-model="decodeList.split" label="换行分割" @change="exec"/>
      <el-checkbox v-model="decodeList.unicode_decode" label="unicode解码" @change="exec"/>
      <el-checkbox v-model="decodeList.url_decode" label="urldecode" @change="exec"/>
      <el-checkbox v-model="decodeList.base64_decode" label="base64解码" @change="exec"/>
      <el-checkbox v-model="decodeList.php_unserializer" label="php序列化" @change="exec"/>
      <el-checkbox v-model="decodeList.json_decode" label="json解码" @change="exec"/>
    </div>

<!--    <el-input v-for="(v,k) in searchLineList"-->
<!--        style="margin-top: 5px;"-->
<!--        type="textarea"-->
<!--        v-model="v.value"-->
<!--        class="pretty-json-p"-->
<!--        ref="textarea"-->
<!--    ></el-input>-->
  </div>
  <template v-for="(v,k) in searchLineList">
    <div style="margin-top: 8px;">
      <pl-button class="copy-btn" link @click="copyJson(v.value)">复制</pl-button>
      <pre class="pretty-json" ref="textarea">{{ v.value }}</pre>
    </div>

  </template>

</template>

<script>
import base from "@/utils/base"
import php from "@/utils/base/php"
import list from "@/utils/base/list"
import t from "@/utils/base/type"
import copy from "@/utils/base/copy";

export default {
  props : {
    source : {
      type : String,
      default: ``
    },
  },
  data() {
    return {
      name: 'Decode',
      codeStr: '',
      searchKey : '',
      chooseTableType: 1,
      tableNum: 10,
      modelResult: '',
      sshConfig: {},
      lineList : [],
      searchLineList : [],
      searchLineNum : 0,
      decodeList : {
        split : true,//换行分割
        //解码
        url_decode : true,//urldecode
        unicode_decode : true,//unicode解码
        base64_decode : true,//base64解码
        php_unserializer : true,//php反序列化
        json_decode : true, //json decode
        //编码
        md5_code : false,//md5编码
        base64_encode : false, //base64编码
      },
    }
  },
  mounted: function () {
    let _that = this
    if(_that.source !== ``){
      _that.codeStr = _that.source
      _that.exec()
    }
  },
  methods: {
    copyJson : function(copyContent){
      let index = copy.SetCopyContent(copyContent)
      copy.handleCopy(index)
    },
    handleInput(index) {
      let _that = this
      setTimeout(function (){
        if(_that.$refs.textarea[index]){
          // const textarea = _that.$refs.textarea[index].$el.querySelector('textarea');
          // textarea.style.height = 'auto'; // 重置高度
          // textarea.style.height = textarea.scrollHeight + 'px'; // 根据内容设置高度
        }
      },100)
    },
    //执行
    exec: function () {
      let _that = this
      _that.lineList = []
      _that.searchLineList = []
      //分割多行
      _that.splitLine()
      //php序列化
      _that.phpUnSerializer()
      //base64接入码
      _that.base64decode()
      //Unicode转义
      _that.unicode()
      //urldecode
      _that.urldecode()
      //json解码
      _that.jsonDecode()
      //触发一次空搜索
      _that.searchList()
    },
    unicode : function (){
      let _that = this
      if(!_that.decodeList.unicode_decode){
        return
      }
      for (let i = 0; i < _that.lineList.length; i++) {
        _that.lineList[i].value = _that.lineList[i].value.replace(/\\u([\dA-Fa-f]{4})/g, function(match, grp) {
          return String.fromCharCode(parseInt(grp, 16));
        });
      }
    },
    urldecode : function (){
      let _that = this
      if(!_that.decodeList.url_decode){
        return
      }
      for (let i = 0; i < _that.lineList.length; i++) {
        try {
          _that.lineList[i].value = decodeURIComponent(_that.lineList[i].value)
        }catch (e) {

        }
      }
    },
    phpUnSerializer : function (){
      let _that = this
      if(!_that.decodeList.php_unserializer){
        return
      }
      for (let i = 0; i < _that.lineList.length; i++) {
        try {
          php.PhpUnserialize2(_that.lineList[i].value , function (response){
            if (response.ErrCode !== 0) {
            } else {
              if(t.IsArray(response.Data)){
                _that.lineList[i].value = response.Data[0]
                  for (let j = 0; j < response.Data.length; j++) {
                    if(j === 0){
                      continue;
                    }
                    _that.lineList.push({
                      value : response.Data[j]
                    })
                  }
              }else{
                //_that.lineList[i].value = response.Data
              }
            }
          })
        }catch (e) {

        }
      }
      console.log('解压完' , _that.lineList)
    },
    base64decode : function (){
      let _that = this
      if(!_that.decodeList.base64_decode){
        return
      }
      for (let i = 0; i < _that.lineList.length; i++) {
        try {
          if(base.IsBase64(_that.lineList[i].value)){
            _that.lineList[i].value = atob(_that.lineList[i].value)
          }
        }catch (e) {

        }
      }
    },
    jsonDecode : function (){
      let _that = this
      if(!_that.decodeList.json_decode){
        return
      }
      for (let i = 0; i < _that.lineList.length; i++) {
        try {
          _that.lineList[i].value = JSON.parse(_that.lineList[i].value)
          console.log('解码后的',_that.lineList[i].value)
        }catch (e) {
          console.log('json解码失败',e)
        }
      }
    },
    splitLine : function (){
      let _that = this
      if(_that.decodeList.split){
        let lineList = _that.codeStr.split(/\r?\n/).filter(line => line.trim() !== '');
        for (let i = 0; i < lineList.length; i++) {
          _that.lineList.push({
            value : lineList[i]
          })
        }
      }else{
        _that.lineList.push({
          value : _that.codeStr
        })
      }
    },
    searchList : function (){
      let _that = this
      console.log(_that.lineList)
      let ret = list.QuickSearch(_that.searchKey , _that.lineList , ['value'])
      console.log(ret)
      _that.searchLineList = ret.list
      _that.searchLineNum = ret.searchNum
      for (let i = 0; i < _that.searchLineList.length; i++){
        _that.handleInput(i)
      }
    },
  },
}
</script>

<style scoped></style>
