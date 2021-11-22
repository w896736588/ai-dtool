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
    </el-menu>
  </div>

</template>

<script>
export default {
  data() {
    return {
      activeIndex2: '1'
    };
  },
  props : {
    menuName:String
  },
  methods: {
    handleSelect(key, keyPath) {
      console.log(key, keyPath);
    }
  }
}
</script>

<style scoped>

</style>
