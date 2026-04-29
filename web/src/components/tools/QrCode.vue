<template>
  <div ref="mainCard" style="text-align: center;">
    <el-row :gutter="20">
      <el-col :span="12" v-for="(value, key) in qrCodeList" :key="key">
        <el-form label-width="auto" style="max-width: 600px">
          <el-form-item label="">
            <el-input v-model="value.qrContent" type="textarea" class="inputQrContent" rows="10" placeholder="输入地址"/>
          </el-form-item>
          <el-form-item label="">
            <div style="display: flex; margin: auto">
              <QrcodeVue v-if="value.qrContent !== ''" :value="value.qrContent" foreground-color="red" :size="200" level="H" ></QrcodeVue>
            </div>
          </el-form-item>
        </el-form>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import QrcodeVue from "qrcode.vue";
export default {
  props : {
    shellShowResult : {
      type : String
    },
  },
  components: {
    QrcodeVue,
  },
  data() {
    return {
      qrCodeList : [
        {
          qrContent : '',
        },
        {
          qrContent : '',
        },
        {
          qrContent : '',
        },
        {
          qrContent : '',
        },
        {
          qrContent : '',
        },
        {
          qrContent : '',
        }
      ],
    }
  },
}
</script>

<style scoped src="@/css/components/tools/QrCode.css"></style>
