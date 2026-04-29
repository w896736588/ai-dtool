<template>
  <div class="kv-table">
    <div class="kv-table-scroll">
    <table class="kv-table-inner">
      <thead>
      <tr>
        <th class="col-key" style="width: 200px;">键</th>
        <th class="col-value" style="width: 120px;">类型</th>
        <th class="col-value">值</th>
        <th class="col-desc">描述</th>
        <th class="col-actions">操作</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="(item, idx) in list" :key="item.id">
        <td>
          <el-input
              v-model="item.field"
              placeholder="键名"
              @blur="handleDataChange"
          />
        </td>
        <td>
          <el-select
              v-model="item.type"
              placeholder="类型"
              @change="handleTypeChange(idx)"
              @blur="handleDataChange"
          >
            <el-option
                v-for="type in typeOptions"
                :key="type.value"
                :label="type.label"
                :value="type.value"
            />
          </el-select>
        </td>
        <td>
          <div v-if="item.type === 'file'">
            <el-upload
                :ref="`upload-${item.id}`"
                :auto-upload="false"
                :show-file-list="false"
                :on-change="(file) => handleFileChange(idx, file , item)"
                :accept="item.accept || '*'"
                @blur="handleDataChange"
            >
              <pl-button size="small" type="primary">选择文件</pl-button>
            </el-upload>
            <span style="display: inline-block;font-size:13px;">{{item.value}}</span>
            <span style="word-break: break-all;" v-if="item.file" class="file-name">{{ item.file.name }}</span>
          </div>
          <div v-else-if="item.type === 'array(string)'">
            <el-input
                v-model="item.value"
                placeholder="用逗号分隔的值"
                @blur="handleDataChange"
            />
          </div>
          <div v-else-if="item.type === 'integer'">
            <el-input
                :model-value="formatIntegerValue(item.value)"
                inputmode="numeric"
                placeholder="整数值"
                @input="handleIntegerInput(idx, $event)"
                @blur="handleIntegerBlur(idx)"
            />
          </div>
          <div v-else-if="item.type === 'json'">
            <el-input
                v-model="item.value"
                type="textarea"
                :rows="Number(2)"
                placeholder="JSON格式数据"
                @blur="handleDataChange"
            />
          </div>
          <div v-else-if="item.type === 'form'">
            <el-input
                v-model="item.value"
                type="textarea"
                :rows="Number(2)"
                placeholder="表单数据"
                @blur="handleDataChange"
            />
          </div>
          <div v-else-if="item.type === 'raw'">
            <el-input
                v-model="item.value"
                type="textarea"
                :rows="Number(2)"
                placeholder="原始数据"
                @blur="handleDataChange"
            />
          </div>
          <div v-else>
            <el-input v-model="item.value" placeholder="值" @blur="handleDataChange"/>
          </div>
        </td>
        <td>
          <el-input type="textarea" :rows="Number(2)" v-model="item.description" placeholder="描述" @blur="handleDataChange"/>
        </td>
        <td class="col-actions">
          <pl-button plain size="small" type="danger" class="delete-rule-btn" @click="removeItem(idx)">删除</pl-button>
        </td>
      </tr>
      </tbody>
    </table>
    </div>

    <div class="footer" style="margin: 10px;">
      <pl-button type="primary" plain size="small" class="add-rule-btn" @click="addItem">+ 添加参数</pl-button>
<!--      <pl-button link @click="handleBulkEdit">批量编辑</pl-button>-->
    </div>
  </div>
</template>

<script>
import Base from '@/utils/base.js'

export default {
  name: 'KeyValueEditor',
  props: {
    list: {
      type: Array,
      default: () => []
    },
  },
  data() {
    return {
      bulkEditVisible: false,
      bulkEditText: '',
      nextId: 1,
      typeOptions: [
        { value: 'string', label: '字符串' },
        { value: 'float', label: '浮点值' },
        { value: 'integer', label: '整数' },
        { value: 'file', label: '文件' },
        { value: 'boolean', label: '布尔值' },
      ]
    }
  },
  methods: {
    handleDataChange : function (){
      this.$emit('update', this.list)
    },
    addItem() {
      this.list.push({ id: this.nextId++, field: '', type: 'string', value: '', description: '' })
      this.handleDataChange()
    },

    removeItem(index) {
      this.list.splice(index, 1)
      // 如果删除了所有行，添加一个空行
      if (this.list.length === 0) {
        this.addItem()
      }
      this.handleDataChange()
    },

    handleTypeChange(index) {
      // 根据类型重置值
      const item = this.list[index]
      if (item.type === 'integer') {
        item.value = '0'
      } else if (item.type === 'array(string)') {
        item.value = ''
      } else if (item.type === 'json') {
        item.value = '{}'
      } else if (item.type === 'file') {
        item.value = null
        item.file = null
      } else {
        item.value = ''
      }
    },

    handleFileChange(index, file , item) {
      let _that = this
      Base.UploadFile(file.raw, (res) => {
        if(res.ErrCode !== 0){
          _that.$message.error(res.ErrMsg)
          return
        }
        item.value = res.Data.url
        _that.list[index].value = res.Data.url
        _that.handleDataChange()
      })

    },

    handleBulkEdit() {
      this.bulkEditText = this.list
          .filter(item => item.field && item.field.trim())
          .map(item => `${item.field}=${item.value}`)
          .join('\n')
      this.$emit('bulk-edit', this.bulkEditText)
    },
    formatIntegerValue(value) {
      if (value === null || value === undefined) {
        return ''
      }
      return String(value)
    },
    sanitizeIntegerValue(value) {
      const text = String(value ?? '')
      const hasLeadingMinus = text.trimStart().startsWith('-')
      const digits = text.replace(/\D/g, '')
      return hasLeadingMinus ? `-${digits}` : digits
    },
    handleIntegerInput(index, value) {
      this.list[index].value = this.sanitizeIntegerValue(value)
      this.handleDataChange()
    },
    handleIntegerBlur(index) {
      const normalized = this.sanitizeIntegerValue(this.list[index].value)
      this.list[index].value = normalized === '-' ? '' : normalized
      this.handleDataChange()
    },

  }
}
</script>

<style scoped src="@/css/components/api/KeyValueEditor.css"></style>




