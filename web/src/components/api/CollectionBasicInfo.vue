<template>
  <div class="collection-basic-info">
    <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        label-position="left"
    >
      <el-form-item label="集合ID">
        <el-tag type="info" size="small">{{ collection.id }}</el-tag>
      </el-form-item>

      <el-form-item label="集合名称" prop="name">
        <el-input
            v-model="form.name"
            placeholder="请输入集合名称"
            maxlength="50"
            show-word-limit
        />
      </el-form-item>

      <el-form-item label="描述信息" prop="description">
        <el-input
            v-model="form.description"
            type="textarea"
            :rows="4"
            placeholder="请输入集合描述"
            maxlength="200"
            show-word-limit
        />
      </el-form-item>

      <el-form-item>
        <pl-button type="primary" @click="handleSave">保存更改</pl-button>
<!--        <pl-button @click="handleReset">重置</pl-button>-->
        <pl-button type="danger" @click="handleDelete" v-if="collection.id">删除集合</pl-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import Api from "@/utils/base/api";

export default {
  name: 'CollectionBasicInfo',
  props: {
    collection: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      form: {
        name: '',
        description: '',
        tags: [],
        version: '1.0.0',
        baseUrl: '',
        protocol: 'https'
      },
      rules: {
        name: [
          { required: true, message: '请输入集合名称', trigger: 'blur' },
          { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ],
        description: [
          { max: 200, message: '描述信息不能超过 200 个字符', trigger: 'blur' }
        ]
      },
      presetTags: ['用户管理', '订单系统', '支付接口', '数据统计', '系统管理']
    }
  },
  watch: {
    collection: {
      handler(newVal) {
        this.loadCollectionData(newVal)
      },
      immediate: true,
      deep: true
    }
  },
  methods: {
    loadCollectionData(collection) {
      if (collection && collection.id) {
        this.form = {
          name: collection.name || '',
          description: collection.description || '',
          tags: collection.tags || [],
          version: collection.version || '1.0.0',
          baseUrl: collection.baseUrl || '',
          protocol: collection.protocol || 'https'
        }
      } else {
        this.handleReset()
      }
    },

    handleSave() {
      this.$refs.formRef.validate((valid) => {
        if (valid) {
          this.$emit('update', {
            ...this.collection,
            ...this.form
          })
          this.$message.success('保存成功')
        } else {
          this.$message.error('请检查表单数据')
          return false
        }
      })
    },

    handleReset() {
      if (this.collection && this.collection.id) {
        this.loadCollectionData(this.collection)
      } else {
        this.form = {
          name: '',
          description: '',
          tags: [],
          version: '1.0.0',
          baseUrl: '',
          protocol: 'https'
        }
      }
      this.$refs.formRef.clearValidate()
    },

    handleDelete() {
      let _that = this
      _that.$emit('delete', _that.collection)
    }
  }
}
</script>

<style scoped src="@/css/components/api/CollectionBasicInfo.css"></style>

