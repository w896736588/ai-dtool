<template>
  <div class="kv-table">
    <table class="kv-table-inner">
      <thead>
      <tr>
        <th class="col-value">提取表达式</th>
        <th class="col-env">提取到环境参数</th>
        <th class="col-desc">描述</th>
        <th class="col-actions">操作</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="(item, idx) in localData" :key="item.id">
        <td>
          <el-input
              v-model="item.value"
              placeholder="data.ta"
              @blur="handleDataChange"
          />
        </td>
        <td>
          <el-select
              v-model="item.item_key"
              placeholder="选择环境参数"
              @change="handleDataChange"
          >
            <el-option
                v-for="envItem in envItems"
                :key="envItem.id"
                :label="envItem.key"
                :value="envItem.key"
            />
          </el-select>
        </td>
        <td>
          <el-input v-model="item.description" placeholder="描述" @blur="handleDataChange" />
        </td>
        <td class="col-actions">
          <pl-button plain size="small" type="danger" class="delete-rule-btn" @click="removeItem(idx)">删除</pl-button>
        </td>
      </tr>
      </tbody>
    </table>

    <div class="footer" style="margin: 5px;">
      <pl-button type="primary" plain size="small" class="add-rule-btn" @click="addItem">+ 添加提取规则</pl-button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'JsonExtractEditor',
  props: {
    modelValue: {
      type: Array,
      default: () => []
    },
    envItems : {
      type : Array,
      default: () => []
    }
  },
  data() {
    return {
      localData: [],
      nextId: 1,
    }
  },
  watch: {
    modelValue: {
      handler(newVal) {
        this.updateLocalData(newVal)
      },
      immediate: true,
      deep: true
    }
  },
  mounted() {

  },
  methods: {
    updateLocalData(sourceData) {
      if (!sourceData || sourceData.length === 0) {
        this.localData = [{ value: '', item_key: '', description: '' }]
        return
      }

      this.localData = sourceData.map(item => ({
        value: item.value !== undefined ? String(item.value) : '',
        item_key: item.item_key || '',
        description: item.description || ''
      }))

      if (this.localData.length === 0) {
        this.localData.push({ value: '', item_key: '', description: '' })
      }
    },

    emitUpdate() {
      this.$emit('update',  this.localData)
    },

    handleDataChange() {
      this.emitUpdate()
    },

    addItem() {
      this.localData.push({ value: '', item_key: '', description: '' })
      this.emitUpdate()
    },

    removeItem(index) {
      this.localData.splice(index, 1)
      if (this.localData.length === 0) {
        this.addItem()
      }
      this.emitUpdate()
    }
  }
}
</script>

<style scoped>
.kv-table {
  width: 100%;
  border: 1px solid #e6ece0;
  border-radius: 10px;
  background: #fff;
  overflow: hidden;
}

.kv-table-inner {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
}

.kv-table-inner th,
.kv-table-inner td {
  padding: 8px 12px;
  border-bottom: 1px solid #eef3ec;
}

.kv-table-inner th {
  background: #f7f9f5;
  font-weight: 600;
  font-size: 14px;
  color: #4e594a;
}

.col-value { width: 35%; }
.col-env { width: 25%; }
.col-desc { width: 25%; }
.col-actions { width: 15%; text-align: center; }

.kv-table-inner .el-input,
.kv-table-inner .el-select {
  width: 100%;
}

.add-rule-btn {
  border-radius: 8px;
}

.delete-rule-btn {
  border-radius: 8px;
}
</style>




