<template>
  <div>
    <iframe v-once :src="src" width="100%;" height="800px;" ></iframe>
  </div>

</template>

<script>
export default {
  data () {
    return {
      name: "WxDoc",
      src : "https://developers.weixin.qq.com/doc/offiaccount/Getting_Started/Overview.html",
    }
  },
}
</script>

<style scoped>

</style>
