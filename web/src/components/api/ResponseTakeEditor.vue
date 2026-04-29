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

<style scoped src="@/css/components/api/ResponseTakeEditor.css"></style>




