<template>
  <div>
    <iframe v-once :src="src" width="100%;" height="800px;" ></iframe>
  </div>

</template>

<script>
export default {
  data () {
    return {
      name: "Ip",
      src : "https://tool.lu/ip",
    }
  },
}
</script>

<style scoped>

</style>
