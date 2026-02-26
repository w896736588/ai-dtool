<template>
  <template>
    <el-row>
      <br />
      <el-col :span="5">
        <el-card class="box-card">
          <template #header>
            <span>历史记录</span>
          </template>
          <el-tag
            type="info"
            closable
            @close="removeCacheHistory(value)"
            style="margin-left: 5px; margin-top: 5px"
            v-for="(value, key) in history"
            :key="key"
          >
            <span
              style="
                font-size: 13px;
                color: blue;
                word-wrap: break-word;
                cursor: default;
              "
              @click="addFromHistory(value)"
              >{{ value }}</span
            >
          </el-tag>
        </el-card>
      </el-col>
      <el-col :span="1"> &nbsp; </el-col>
      <el-col :span="18">
        <el-card class="box-card">
          <div class="grid-content bg-purple">
            <el-input
              type="textarea"
              :rows="10"
              v-model="textValue"
              placeholder="输入需要转化的内容"
            ></el-input>
          </div>
          <br />
          <el-button type="primary" @click="urlEncode">urlencode编码</el-button>
          <el-button type="primary" @click="urlDecode">urlencode解码</el-button>
          <el-button type="primary" @click="md5">md5</el-button>
          <br />
          <br />
          <div class="grid-content bg-purple">
            <el-input
              ref="copyResult"
              type="textarea"
              :rows="10"
              v-model="textValueTrans"
              placeholder="结果"
            ></el-input>
          </div>
          <br />
          <el-button @click="copyResult('copyResult')" type="primary"
            >复制</el-button
          >
        </el-card>
      </el-col>
    </el-row>

    <!--    <iframe v-once :src="src" width="100%;" height="630px;" ></iframe>-->
  </template>
</template>

<script>
import md5 from 'js-md5'
import { ElMessage as Message } from 'element-plus'
export default {
  data() {
    return {
      textValue: '',
      name: 'UrlEncode',
      history: [],
      textValueTrans: '',
      src: 'http://www.jsons.cn/urlencode/',
    }
  },
  components: {
  },
  mounted: function () {
    this.initHistory()
  },
  methods: {
    initHistory: function () {
      let history = localStorage.getItem('encodeHistoryList')
      if (history === '' || history === null) {
        this.history = []
      } else {
        this.history = JSON.parse(history)
      }
    },
    urlEncode: function () {
      this.textValueTrans = encodeURIComponent(this.textValue)
      this.setCacheHistory()
    },
    urlDecode: function () {
      this.textValueTrans = decodeURIComponent(this.textValue)
      this.setCacheHistory()
    },
    md5: function () {
      this.textValueTrans = md5(this.textValue)
      this.setCacheHistory()
    },
    setCacheHistory: function () {
      let boolFind = false
      for (var i in this.history) {
        if (this.history[i] === this.textValue) {
          boolFind = true
          break
        }
      }
      if (!boolFind) {
        //加入到历史中
        this.history.push(this.textValue)
        this.saveHistoryCache()
      }
    },
    addFromHistory: function (textValue) {
      this.textValue = textValue
      this.urlEncode()
    },
    removeCacheHistory: function (removeValue) {
      console.log(removeValue)
      let tempHistoryList = []
      for (var i in this.history) {
        if (this.history[i] !== removeValue) {
          tempHistoryList.push(this.history[i])
        }
      }
      this.history = tempHistoryList
      this.saveHistoryCache()
    },
    saveHistoryCache: function () {
      localStorage.setItem('encodeHistoryList', JSON.stringify(this.history))
    },
    copyResult: function (elemRef) {
      let target
      let succeed = false
      if (this.$refs[elemRef]) {
        target = this.$refs[elemRef]
        // 选择内容
        let currentFocus = document.activeElement
        target.focus()
        target.setSelectionRange(0, target.value.length)
        // 复制内容
        try {
          succeed = document.execCommand('copy')
          alert('内容复制成功')
        } catch (e) {
          succeed = false
        }
        // 恢复焦点
        if (currentFocus && typeof currentFocus.focus === 'function') {
          currentFocus.focus()
        }
      }
      return succeed
    },
    success: function (msg) {
      Message.success(msg)
      //this.$notify({title: '提示', message: msg, type: 'success'});
    },
    warning: function (msg) {
      Message.warning(msg)
      //this.$notify({title: '提示', message: msg, type: 'warning'});
    },
    info: function (msg) {
      Message.info(msg)
      //this.$notify({title: '提示', message: msg});
    },
    error: function (msg) {
      Message.error(msg)
      //this.$notify({title: '提示', message: msg, type: 'error'});
    },
  },
}
</script>

<style scoped></style>
