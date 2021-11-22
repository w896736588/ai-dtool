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
      <el-menu-item index="RedisCache">
        Redis缓存
      </el-menu-item>
      <el-menu-item index="Json">
        JSON格式化
      </el-menu-item>
      <el-menu-item index="Unix">
        时间戳
      </el-menu-item>
      <el-menu-item index="Ip">
        IP查询
      </el-menu-item>

      <el-submenu index="Doc">
        <template slot="title">开发文档</template>
        <el-menu-item index="Doc-1">
          <a style="color:white;" target="_blank" href="https://developers.weixin.qq.com/doc/offiaccount/Getting_Started/Overview.html">微信开发文档</a>
        </el-menu-item>
        <el-menu-item index="Doc-2">
          <a style="color:white;" target="_blank" href="https://work.weixin.qq.com/kf/">微信客服文档</a>
        </el-menu-item>
      </el-submenu>

    </el-menu>
<!--    内容-->
    <RedisCache v-if="menuName === 'RedisCache'"></RedisCache>
    <Json v-if="menuName === 'Json'"></Json>
    <Unix v-if="menuName === 'Unix'"></Unix>
    <Ip v-if="menuName === 'Ip'"></Ip>
  </div>

</template>

<script>
import RedisCache from "./CacheIndex"
import Json from "./Json"
import Unix from "./Unix"
import Ip from "./Ip"
import WxDoc from "./WxDoc"
export default {
  data () {
    return {
      name: "Index",
      menuName : "RedisCache",
    }
  },
  methods: {
    handleSelect(key, keyPath) {
      this.menuName = keyPath[0];
    }
  },
  components : {
    Json,
    RedisCache,
    Unix,
    Ip,
    WxDoc,
  },
}
</script>

<style scoped>

</style>
