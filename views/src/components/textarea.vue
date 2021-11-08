<template>
  <el-form ref="form" >
    <el-form-item v-if="cache.cacheType === 'string'" style="margin-top:10px;">
      <el-input type="textarea" v-model="textarea" rows="20"></el-input>
    </el-form-item>
  </el-form>
</template>

<script>
export default {
  name: "textarea"
}
</script>

<style scoped>

</style>
