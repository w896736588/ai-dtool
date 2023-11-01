<template>
  <div>

    <el-menu
      :default-active="menuName"
      class="el-menu-demo"
      router
      mode="horizontal"
      @select="handleSelect"
      background-color="#545c64"
      text-color="#fff"
      active-text-color="#ffd04b">
      <el-menu-item index="/CacheIndex">
        Redis
      </el-menu-item>
      <el-menu-item index="/Consumer">
        消费者
      </el-menu-item>
      <el-menu-item index="/Git">
        Git
      </el-menu-item>
      <el-menu-item index="/WechatKefu">
        微信客服
      </el-menu-item>
      <el-menu-item index="/Vip">
        版本变更
      </el-menu-item>
      <el-menu-item index="/Link">
        登录/链接
      </el-menu-item>
      <el-menu-item index="/Docker">
        Docker
      </el-menu-item>
      <el-menu-item index="/Model">
        Model/建表Sql
      </el-menu-item>
      <!--      <el-menu-item index="Tools">-->
      <!--        小工具-->
      <!--      </el-menu-item>-->
      <el-menu-item index="/Ssh">
        服务器配置
      </el-menu-item>

      <el-submenu index="">
        <template slot="title">开发文档</template>
        <el-menu-item>
          <a style="color:white;" target="_blank" href="https://developers.weixin.qq.com/doc/offiaccount/Getting_Started/Overview.html">微信开发文档</a>
        </el-menu-item>
        <el-menu-item>
          <a style="color:white;" target="_blank" href="https://kf.weixin.qq.com/api/doc/path/93304">微信客服文档</a>
        </el-menu-item>
        <el-menu-item>
          <a style="color:white;" target="_blank" href="https://open.work.weixin.qq.com/api/doc/90000/90135/90664">企业微信文档</a>
        </el-menu-item>
        <el-menu-item>
          <a style="color:white;" target="_blank" href="https://element.eleme.cn/#/zh-CN/component/installation">ElementUI</a>
        </el-menu-item>
      </el-submenu>
    </el-menu>
    <el-tag style="margin:10px;"
            v-for="tag in tags"
            :key="tag.name"
            closable
            :type="tag.type">
      {{tag.name}}
    </el-tag>
    <el-main style="height:370px;">
      <keep-alive>
        <router-view  name="home"></router-view>
      </keep-alive>
    </el-main>
    <div class="sticky-textarea"  v-if="terShow() && v.boolShow" v-for="(v,k) in shellMapList" >
      <el-input :id="v.unikey + '-' + v.shellName" style="width: 100%;resize: vertical; " type="textarea" v-model="v.shellResult" rows="10"></el-input>
      <div style="background-color: #e4e4e4;padding:5px;width: 100%;">
        <span style="font-weight: 500;font-size:14px;">终端:{{v.shellName}}</span>
        <div style="display: inline-block;margin-left: 50%;">
          <input style="height:20px;width: 400px;border:1px solid #409EFF;padding:3px;" autocomplete="off" placeholder="输入执行命令回车" v-model="runCommand"></input>
          <button @click="toggleTextarea">执行命令</button>
          <button @click="toggleTextarea">展开</button>
          <button @click="toggleTextarea">收起</button>
        </div>

      </div>
    </div>
  </div>

</template>
<style>
.sticky-textarea {
  background-color: white;
  position: fixed;
  bottom: 0px;
  left: 0;
  right: 0;
  height:36%;
  width: 100%;
}
</style>

<script>
import base from "../utils/api/base";
import mod from "../utils/api/module";
import 'xterm/css/xterm.css'
import 'xterm/lib/xterm.js'
import { Terminal } from 'xterm'
// import { FitAddon } from 'xterm-addon-fit'
// import { AttachAddon } from 'xterm-addon-attach'


export default {
  data () {
    return {
      menuName : "/CacheIndex",
      minHeightMap : {//内容框最小高度配置

      },
      showShellMap : [ //哪些菜单展示终端
        '/Git',
        '/Consumer',
      ],
      tags : [],
      showTextarea : true,
      shellResult : [],
      //存储socket链接和结果
      shellMapList : [],
      runCommand : '',
    }
  },
  mounted : function (){
    let _that = this
    //先注册服务
    base.BaseCheckService()
    //建立socket 延时初始化
    setTimeout(function (){
      _that.initSocket()
    } , 1000);

    //处理默认打开的页卡
    this.menuName = this.$helperStore.getStore('lastMenuName')
    if(!this.$helperConfig.getXkfDevSshConfig() || !this.$helperConfig.getWkDevSshConfig() || !this.$helperConfig.getXkfDevDbConfig()){
      this.menuName = '/Ssh';
    }
    if (this.$route.path !== this.menuName) {
      this.$router.push(this.menuName)
    }
  },
  methods: {
    terShow : function (){
      for(let i in this.showShellMap){
        if(this.showShellMap[i] === this.menuName){
          return true;
        }
      }
      return false;
    },
    showNotify : function (notifyList){
      this.tags = notifyList
    },
    //供子组件调用
    showTerminal (unikey , shellName){
      for(let i in this.shellMapList){
        this.shellMapList[i].boolShow = this.shellMapList[i].unikey === unikey && this.shellMapList[i].shellName === shellName;
      }
    },

    handleSelect(key, keyPath) {
      if(keyPath[0].indexOf('Doc-') >= 0){
        return
      }
      this.menuName = keyPath[0];
      this.$helperStore.setStore('lastMenuName' , this.menuName)
    },
    toggleTextarea : function (){
      this.showTextarea = false;
    },
    initSocket : function (){
      let _that = this
      let unikey = base.GetUnikey()
      let shellConfigList = mod.GetShellConfigList()
      for(let i in shellConfigList){
        let shellName = shellConfigList[i].name
        _that.shellMapList.push({
          unikey : unikey,
          shellName : shellName,
          shellResult : '',
          boolShow : false,
        })
        base.SetSocketErrorFunc(unikey , shellName , function (error) {

        })
        base.SetSocketOnOpenFunc(unikey , shellName , function (){
          base.SetSocketHeart(unikey , shellName)
        })
        base.SetSocketMessageFunc(unikey , shellName , function (msg) {
          for(let i in _that.shellMapList){
            if(_that.shellMapList[i].unikey === unikey && _that.shellMapList[i].shellName === shellName){
              _that.shellMapList[i].shellResult += msg
            }
          }
          setTimeout(function (){
            if(document.getElementById(unikey + '-' + shellName)){
              document.getElementById(unikey + '-' + shellName).scrollTop = document.getElementById(unikey + '-' + shellName).scrollHeight + 200;
            }
          } , 500)
        })
      }
    }
  },
  components : {

  },
}
</script>

<style scoped>

</style>
