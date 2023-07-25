<template>
  <div>
    <el-menu
      :default-active="menuName"
      class="el-menu-demo"
      mode="horizontal"
      @select="handleSelect"
      background-color="#545c64"
      text-color="#fff"
      active-text-color="#ffd04b">
      <el-menu-item index="CacheIndex">
        Redis
      </el-menu-item>
      <el-menu-item index="Consumer">
        消费者
      </el-menu-item>
      <el-menu-item index="Git">
        Git
      </el-menu-item>
      <el-menu-item index="WechatKefu">
        微信客服
      </el-menu-item>
      <el-menu-item index="Vip">
        版本变更
      </el-menu-item>
      <el-menu-item index="Link">
        登录/链接
      </el-menu-item>
      <el-menu-item index="Docker">
        Docker
      </el-menu-item>
      <el-menu-item index="Model">
        Model/建表Sql
      </el-menu-item>
<!--      <el-menu-item index="Tools">-->
<!--        小工具-->
<!--      </el-menu-item>-->
      <el-menu-item index="Ssh">
        服务器配置
      </el-menu-item>

      <el-submenu index="Doc">
        <template slot="title">开发文档</template>
        <el-menu-item index="Doc-1">
          <a style="color:white;" target="_blank" href="https://developers.weixin.qq.com/doc/offiaccount/Getting_Started/Overview.html">微信开发文档</a>
        </el-menu-item>
        <el-menu-item index="Doc-2">
          <a style="color:white;" target="_blank" href="https://kf.weixin.qq.com/api/doc/path/93304">微信客服文档</a>
        </el-menu-item>
        <el-menu-item index="Doc-4">
          <a style="color:white;" target="_blank" href="https://open.work.weixin.qq.com/api/doc/90000/90135/90664">企业微信文档</a>
        </el-menu-item>
        <el-menu-item index="Doc-3">
          <a style="color:white;" target="_blank" href="https://element.eleme.cn/#/zh-CN/component/installation">ElementUI</a>
        </el-menu-item>
      </el-submenu>

<!--      <el-submenu index="DocTool">-->
<!--        <template slot="title">工具地址</template>-->
<!--        <el-menu-item index="DocTool-1">-->
<!--          <a style="color:white;" target="_blank" href="https://www.json.cn/">Json格式化</a>-->
<!--        </el-menu-item>-->
<!--        <el-menu-item index="DocTool-2">-->
<!--          <a style="color:white;" target="_blank" href="https://tool.lu/timestamp/">时间戳</a>-->
<!--        </el-menu-item>-->
<!--        <el-menu-item index="DocTool-4">-->
<!--          <a style="color:white;" target="_blank" href="https://ip.tool.chinaz.com/github.com">Ip查询</a>-->
<!--        </el-menu-item>-->
<!--        <el-menu-item index="DocTool-3">-->
<!--          <a style="color:white;" target="_blank" href="https://fanyi.baidu.com/">翻译</a>-->
<!--        </el-menu-item>-->
<!--        <el-menu-item index="DocTool-3">-->
<!--          <a style="color:white;" target="_blank" href="https://cli.im/">二维码</a>-->
<!--        </el-menu-item>-->
<!--      </el-submenu>-->

    </el-menu>
<!--    内容-->
    <div style="height:60px;">
      <el-tag style="margin:10px;"
              v-for="tag in tags"
              :key="tag.name"
              closable
              :type="tag.type">
        {{tag.name}}
      </el-tag>
    </div>

    <Consumer v-show="menuName === 'Consumer'"></Consumer>
    <CacheIndex v-show="menuName === 'CacheIndex'"></CacheIndex>
    <WechatKefu v-show="menuName === 'WechatKefu'"></WechatKefu>
    <Ssh v-show="menuName === 'Ssh'"></Ssh>
    <Git v-show="menuName === 'Git'"></Git>
    <Vip v-show="menuName === 'Vip'"></Vip>
    <Model v-show="menuName === 'Model'"></Model>
    <Link v-show="menuName === 'Link'"></Link>
    <Docker :showNotify="showNotify" v-show="menuName === 'Docker'"></Docker>
<!--    <Tools v-show="menuName === 'Tools'"></Tools>-->
  </div>

</template>

<script>
import CacheIndex from "./CacheIndex"
import Consumer from "./Consumer"
import WechatKefu from "./WechatKefu"
import Ssh from "./Ssh"
import Git from "./Git"
import Vip from "./Vip"
import Model from "./Model"
import Link from "./Link"
import Docker from "./Docker"
import Tools from "./Tools"
export default {
  data () {
    return {
      name: "CacheIndex",
      menuName : "CacheIndex",
      tags : [
      ],
    }
  },
  mounted : function (){
    this.menuName = this.$helperStore.getStore('lastMenuName')
    if(!this.$helperConfig.getXkfDevSshConfig() || !this.$helperConfig.getWkDevSshConfig() || !this.$helperConfig.getXkfDevDbConfig()){
      this.menuName = 'Ssh';
    }
  },
  methods: {
    showNotify : function (notifyList){
      this.tags = notifyList
    },
    handleSelect(key, keyPath) {
      this.menuName = keyPath[0];
      this.$helperStore.setStore('lastMenuName' , this.menuName)
    }
  },
  components : {
    CacheIndex,
    Consumer,
    WechatKefu,
    Ssh,
    Git,
    Vip,
    Model,
    Link,
    Docker,
    Tools,
  },
}
</script>

<style scoped>

</style>
