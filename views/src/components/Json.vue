<template>
  <div>
    <iframe v-once :src="src" width="100%;" height="800px;" ></iframe>
  </div>

</template>

<script>
export default {
  data () {
    return {
      name: "Json",
      src : "https://www.bejson.com/explore/index_new/",
    }
  },
}
</script>

<style scoped>

</style>
