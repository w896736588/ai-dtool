<template>
  <div class="tools-container">
    <el-tabs tab-position="left" class="demo-tabs">
      <el-tab-pane label="时间转换">
        <TimeTransfer></TimeTransfer>
      </el-tab-pane>
      <el-tab-pane label="二维码">
        <QrCode></QrCode>
      </el-tab-pane>
<!--      <el-tab-pane label="Json全解析">-->
<!--        <TimeTransfer></TimeTransfer>-->
<!--      </el-tab-pane>-->
      <el-tab-pane label="字符串解析">
        <Decode></Decode>
      </el-tab-pane>
      <el-tab-pane label="Markdown">
        <markdown></markdown>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import QrCode from '@/components/tools/QrCode'
import TimeTransfer from '@/components/tools/timeTransfer'
import Decode from "@/components/tools/Decode.vue";
import Markdown from "@/components/Markdown.vue";
export default {
  name: 'tools',
  data() {
    return {
    }
  },
  components: {
    Markdown,
    Decode,
    QrCode,
    TimeTransfer,
  },
  methods: {

  },
}
</script>

<style scoped>
.tools-container {
  height: 100%;
  overflow-y: auto;
}
</style>
