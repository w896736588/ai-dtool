<template>
  <div class="folder-edit-form">
    <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        label-position="left"
    >
      <el-form-item label="文件夹名称" prop="name">
        <el-input
            v-model="form.name"
            placeholder="请输入文件夹名称"
            maxlength="50"
            show-word-limit
        />
      </el-form-item>

      <el-form-item label="父文件夹">
        <el-select
            v-model="form.parentId"
            placeholder="选择父文件夹"
            clearable
            style="width: 100%"
        >
          <el-option
              v-for="folder in parentFolderOptions"
              :key="folder.id"
              :label="folder.name"
              :value="folder.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="描述信息">
        <el-input
            v-model="form.description"
            type="textarea"
            :rows="4"
            placeholder="请输入文件夹描述"
            maxlength="200"
            show-word-limit
        />
      </el-form-item>

      <el-form-item label="访问权限">
        <el-radio-group v-model="form.permission">
          <el-radio value="private">私有（仅自己可见）</el-radio>
          <el-radio value="public">公开（团队可见）</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item label="标签">
        <el-select
            v-model="form.tags"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="选择或输入标签"
            style="width: 100%"
        >
          <el-option
              v-for="tag in presetTags"
              :key="tag"
              :label="tag"
              :value="tag"
          />
        </el-select>
      </el-form-item>

      <el-form-item>
        <pl-button type="primary" @click="handleSubmit">保存</pl-button>
        <pl-button @click="handleCancel">取消</pl-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
export default {
  name: 'FolderEditForm',
  props: {
    folder: {
      type: Object,
      default: () => ({})
    }
  },
  data() {
    return {
      form: {
        name: '',
        parentId: '',
        description: '',
        permission: 'private',
        tags: []
      },
      rules: {
        name: [
          { required: true, message: '请输入文件夹名称', trigger: 'blur' },
          { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
        ]
      },
      parentFolderOptions: [
        { id: 1, name: '用户管理系统' },
        { id: 2, name: '订单系统' },
        { id: 3, name: '支付接口' }
      ],
      presetTags: ['核心接口', '测试用例', '开发中', '已上线', '待优化']
    }
  },
  watch: {
    folder: {
      handler(newVal) {
        this.loadFormData(newVal)
      },
      immediate: true
    }
  },
  methods: {
    loadFormData(folder) {
      if (folder && folder.id) {
        this.form = {
          name: folder.name || '',
          parentId: folder.parentId || '',
          description: folder.description || '',
          permission: folder.permission || 'private',
          tags: folder.tags || []
        }
      } else {
        this.form = {
          name: '',
          parentId: '',
          description: '',
          permission: 'private',
          tags: []
        }
      }
    },

    handleSubmit() {
      this.$refs.formRef.validate((valid) => {
        if (valid) {
          this.$emit('submit', {
            ...this.folder,
            ...this.form,
            updateTime: new Date().toISOString()
          })
        } else {
          this.$message.error('请检查表单数据')
          return false
        }
      })
    },

    handleCancel() {
      this.$emit('cancel')
    }
  }
}
</script>

<style scoped>
.folder-edit-form {
  padding: 12px;
  border: 1px solid #e8eee5;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 6px 18px rgba(80, 110, 80, 0.08);
}

.folder-edit-form :deep(.el-form-item:last-child .el-form-item__content) {
  gap: 10px;
}

.folder-edit-form :deep(.el-input__wrapper),
.folder-edit-form :deep(.el-select .el-input__wrapper),
.folder-edit-form :deep(.el-textarea__inner) {
  border-radius: 8px;
}

.folder-edit-form :deep(.el-radio-group) {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}
</style>

