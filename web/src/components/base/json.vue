<template>
  <div class="json-editor-container">
    <div class="editor-switch" v-if="showSource">
      <el-switch
          v-model="showRawData"
          active-text="显示源数据"
          inactive-text="格式化显示"
          @change="handleSwitchChange"
      />
    </div>

    <div v-if="showRawData" class="raw-editor">
      <el-input
          type="textarea"
          :rows="20"
          v-model="rawData"
          @input="handleRawDataBlur"
          placeholder="请输入JSON数据"
      />
    </div>

    <JsonEditorVue
        v-else
        :model-value="internalData"
        :mode="mode"
        :main-menu-bar="false"
        :status-bar="false"
        @change="handleChange"
        @json-change="handleChange"
        @json-save="handleChange"
        @has-error="handleError"
    />
  </div>
</template>

<script>
import JsonEditorVue from 'json-editor-vue'

export default {
  name: 'JsonEditor',
  components: {
    JsonEditorVue
  },
  props: {
    // 原始数据
    value: {
      type: [Object, Array, String, Number, Boolean],
      required: true
    },
    // 编辑器模式：tree(树形) | text(文本)
    mode: {
      type: String,
      default: 'tree',
      validator: (value) => ['tree', 'text'].includes(value)
    },
    showSource: {
      type: [Boolean],
      default: true
    },
  },
  emits: ['change'],
  data() {
    return {
      showRawData: false,
      rawData: '',
      internalData: this.normalizeData(this.value),
    }
  },
  watch: {
    value: {
      handler(newVal) {
        this.internalData = this.normalizeData(newVal)
        this.rawData = this.stringifyData(newVal)
      },
      deep: true
    },
  },
  methods: {
    normalizeData(data) {
      try {
        if (typeof data === 'string') {
          return JSON.parse(data)
        }
        return JSON.parse(JSON.stringify(data))
      } catch (e) {
        return {}
      }
    },
    stringifyData(data) {
      try {
        if (typeof data === 'string') {
          return data
        }
        return JSON.stringify(data, null, 2)
      } catch (e) {
        return ''
      }
    },
    handleChange(newData) {
      let _that = this
      _that.internalData = newData.json
      this.$emit('change', _that.internalData)
    },
    handleError(e) {
      console.log('报错了', e)
    },
    handleSwitchChange(val) {
      if (val) {
        // 切换到原始数据模式时，更新rawData
        this.rawData = this.stringifyData(this.internalData)
      } else {
        // 切换回格式化模式时，尝试解析rawData
        try {
          const parsed = JSON.parse(this.rawData)
          this.internalData = parsed
          this.$emit('change', parsed)
        } catch (e) {
          // 解析失败保持原样
          console.warn('JSON解析失败，保持原数据', e)
        }
      }
    },
    handleRawDataBlur() {
      this.$emit('change', this.rawData)
    },
  },

  mounted() {
    // 初始化rawData
    this.rawData = this.stringifyData(this.value)
  }
}
</script>

<style scoped src="@/css/components/base/json.css"></style>