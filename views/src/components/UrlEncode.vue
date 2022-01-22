<template>
  <div>
    <el-row>
      <br/>
      <el-col :span="11">
        <el-card class="box-card">
          <div slot="header" class="clearfix">
            <span>历史记录</span>
          </div>
          <el-tag type="info" closable @close="removeCacheHistory(value)" style="margin-left: 5px;margin-top:5px;" v-for="(value,key) in history" :key="key">
            <span style="font-size:13px;color:blue;word-wrap:break-word;cursor:default;"  @click="addFromHistory(value)"  >{{ value }}</span>
          </el-tag>
        </el-card>
      </el-col>
      <el-col :span="1">
        &nbsp;
      </el-col>
      <el-col :span="12">
        <el-card class="box-card">
          <div class="grid-content bg-purple">
            <el-input type="textarea" :rows="10" v-model="textValue" placeholder="输入需要转化的内容"></el-input>
          </div>
          <br/>
          <el-button type="primary" @click="urlEncode">urlencode编码</el-button>
          <el-button type="primary" @click="urlDecode">urlencode解码</el-button>
          <el-button type="primary" @click="">md5</el-button>
          <br/> <br/>
          <div class="grid-content bg-purple">
            <el-input type="textarea" :rows="10" v-model="textValueTrans" placeholder="输入需要转化的内容"></el-input>
          </div>
          <br/>
          <el-button type="primary">复制</el-button>
        </el-card>
      </el-col>

    </el-row>

<!--    <iframe v-once :src="src" width="100%;" height="630px;" ></iframe>-->
  </div>

</template>

<script>
export default {
  data () {
    return {
      textValue : "",
      name: "UrlEncode",
      history : [],
      textValueTrans : "",
      src : "http://www.jsons.cn/urlencode/",
    }
  },
  mounted : function (){
    this.initHistory();
  },
  methods : {
    initHistory : function (){
      let history = localStorage.getItem('encodeHistoryList')
      this.history = JSON.parse(history);
      console.log(this.history);
    },
    urlEncode : function (){
      this.textValueTrans = encodeURIComponent(this.textValue);
      this.setCacheHistory();
    },
    urlDecode : function (){
      this.textValueTrans = decodeURIComponent(this.textValue);
      this.setCacheHistory();
    },
    setCacheHistory : function (){
      let boolFind = false
      for(var i in this.history){
        if(this.history[i] === this.textValue){
          boolFind = true
          break
        }
      }
      if(!boolFind){
        //加入到历史中
        this.history.push(this.textValue);
        this.saveHistoryCache();
      }

    },
    addFromHistory : function (textValue){
      this.textValue = textValue;
      this.urlEncode();
    },
    removeCacheHistory : function (removeValue){
      console.log(removeValue);
      let tempHistoryList = [];
      for(var i in this.history){
        if(this.history[i] !== removeValue){
          tempHistoryList.push(this.history[i]);
        }
      }
      this.history = tempHistoryList;
      this.saveHistoryCache();
    },
    saveHistoryCache : function (){
      localStorage.setItem('encodeHistoryList' , JSON.stringify(this.history))
    }
  }
}
</script>

<style scoped>

</style>
