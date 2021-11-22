<template>
  <div>
    <iframe v-once :src="src" width="100%;" height="800px;" ></iframe>
  </div>

</template>

<script>
export default {
  data () {
    return {
      name: "Unix",
      src : "https://tool.lu/timestamp/",
    }
  },
}
</script>

<style scoped>

</style>
