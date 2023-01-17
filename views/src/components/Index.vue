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
      <el-menu-item index="Redis">
        Redis
      </el-menu-item>
      <el-menu-item index="Consumer">
        消费者
      </el-menu-item>
      <el-menu-item index="Git">
        Git/Log
      </el-menu-item>
      <el-menu-item index="WechatKefu">
        微信客服
      </el-menu-item>
      <el-menu-item index="Vip">
        版本变更
      </el-menu-item>
      <el-menu-item index="Ssh">
        服务器配置
      </el-menu-item>

      <el-submenu index="Doc">
        <template slot="title">开发文档</template>
        <el-menu-item index="Doc-1">
          <a style="color:white;" target="_blank" href="https://developers.weixin.qq.com/doc/offiaccount/Getting_Started/Overview.html">微信开发文档</a>
        </el-menu-item>
        <el-menu-item index="Doc-2">
          <a style="color:white;" target="_blank" href="https://work.weixin.qq.com/kf/">微信客服文档</a>
        </el-menu-item>
        <el-menu-item index="Doc-4">
          <a style="color:white;" target="_blank" href="https://open.work.weixin.qq.com/api/doc/90000/90135/90664">企业微信文档</a>
        </el-menu-item>
        <el-menu-item index="Doc-3">
          <a style="color:white;" target="_blank" href="https://element.eleme.cn/#/zh-CN/component/installation">ElementUI</a>
        </el-menu-item>
      </el-submenu>

      <el-submenu index="DocTool">
        <template slot="title">工具地址</template>
        <el-menu-item index="DocTool-1">
          <a style="color:white;" target="_blank" href="https://www.json.cn/">Json格式化</a>
        </el-menu-item>
        <el-menu-item index="DocTool-2">
          <a style="color:white;" target="_blank" href="https://tool.lu/timestamp/">时间戳</a>
        </el-menu-item>
        <el-menu-item index="DocTool-4">
          <a style="color:white;" target="_blank" href="https://ip.tool.chinaz.com/github.com">Ip查询</a>
        </el-menu-item>
        <el-menu-item index="DocTool-3">
          <a style="color:white;" target="_blank" href="https://fanyi.baidu.com/">翻译</a>
        </el-menu-item>
        <el-menu-item index="DocTool-3">
          <a style="color:white;" target="_blank" href="https://cli.im/">二维码</a>
        </el-menu-item>
      </el-submenu>

    </el-menu>
<!--    内容-->
    <Redis v-show="menuName === 'Redis'"></Redis>
    <Consumer v-show="menuName === 'Consumer'"></Consumer>
    <WechatKefu v-show="menuName === 'WechatKefu'"></WechatKefu>
    <Ssh v-show="menuName === 'Ssh'"></Ssh>
    <Git v-show="menuName === 'Git'"></Git>
    <Vip v-show="menuName === 'Vip'"></Vip>
  </div>

</template>

<script>
import Redis from "./Cache"
import Consumer from "./Consumer"
import WechatKefu from "./WechatKefu"
import Ssh from "./Ssh"
import Git from "./Git"
import Vip from "./Vip"
export default {
  data () {
    return {
      name: "Redis",
      menuName : "Redis",
    }
  },
  methods: {
    handleSelect(key, keyPath) {
      this.menuName = keyPath[0];
    }
  },
  components : {
    Redis,
    Consumer,
    WechatKefu,
    Ssh,
    Git,
    Vip,
  },
}
</script>

<style scoped>

</style>
