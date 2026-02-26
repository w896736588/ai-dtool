<template>
  <div class="kv-table">
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
              <el-button size="small" type="primary">选择文件</el-button>
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
            <el-input-number
                v-model="item.value"
                @blur="handleDataChange"
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
          <el-button link type="danger" @click="removeItem(idx)">删除</el-button>
        </td>
      </tr>
      </tbody>
    </table>

    <div class="footer" style="margin: 10px;">
      <el-button type="primary" link @click="addItem">+ 添加参数</el-button>
<!--      <el-button link @click="handleBulkEdit">批量编辑</el-button>-->
    </div>
  </div>
</template>

<script>
import { Delete, Plus, Edit } from '@element-plus/icons-vue'
import Base from '@/utils/base.js'

export default {
  name: 'KeyValueEditor',
  components: {
    Delete,
    Plus,
    Edit
  },
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
        item.value = 0
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

  }
}
</script>

<style scoped>
.kv-table {
  width: 100%;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #fff;
}

.kv-table-inner {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
}

.kv-table-inner th,
.kv-table-inner td {
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.kv-table-inner th {
  background: #f5f7fa;
  font-weight: 600;
  font-size: 14px;
  color: #606266;
}

.col-key { width: 25%; }
.col-value { width: 25%; }
.col-desc { width: 25%; }
.col-actions { width: 10%; text-align: center; }

.kv-table-inner input,
.kv-table-inner .el-autocomplete,
.kv-table-inner .el-input,
.kv-table-inner .el-select {
  width: 100%;
}

.file-name {
  margin-left: 10px;
  color: #606266;
  font-size: 12px;
}
</style>



