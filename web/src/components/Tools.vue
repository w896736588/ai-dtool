<template>
  <div class="tools-container">
    <el-tabs tab-position="left" class="demo-tabs">
      <el-tab-pane label="时间转换">
        <TimeTransfer />
      </el-tab-pane>
      <el-tab-pane label="二维码">
        <QrCode />
      </el-tab-pane>
      <el-tab-pane label="字符串解析">
        <Decode />
      </el-tab-pane>
      <el-tab-pane label="Markdown">
        <Markdown />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import QrCode from '@/components/tools/QrCode'
import TimeTransfer from '@/components/tools/timeTransfer'
import Decode from '@/components/tools/Decode.vue'
import Markdown from '@/components/Markdown.vue'

export default {
  name: 'tools',
  components: {
    Markdown,
    Decode,
    QrCode,
    TimeTransfer,
  },
}
</script>

<style scoped src="@/css/components/Tools.css"></style>
