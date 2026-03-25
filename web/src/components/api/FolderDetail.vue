<template>
  <div class="folder-detail">
    <div class="folder-header">
      <h2 class="folder-title">{{ folder.name }}</h2>
      <div class="folder-actions">
<!--        <pl-button type="primary" @click="handleEdit">编辑文件夹</pl-button>-->
        <pl-button @click="createApi">新建接口</pl-button>
<!--        <pl-button @click="handleCreateSubfolder">新建子文件夹</pl-button>-->
      </div>
    </div>

    <el-divider/>

    <div class="folder-content">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="基本信息" name="basic">
          <folder-basic-info
              :folder="folder"
              @update="handleUpdateFolder"
              @delete="handleDelete"
          />
        </el-tab-pane>

        <!--        <el-tab-pane label="接口列表" name="apis">-->
        <!--          <api-list-->
        <!--              :folder="folder"-->
        <!--              @edit="handleEditApi"-->
        <!--              @execute="handleExecuteApi"-->
        <!--          />-->
        <!--        </el-tab-pane>-->

        <el-tab-pane label="接口文档" name="api_document">
          <api-document
              :apis="documentApis"
              :folder-name="folder.name"
              :folder-id="folder.id"
          />
        </el-tab-pane>

        <!--        <el-tab-pane label="执行历史" name="history">-->
        <!--          <execution-history :folder-id="folder.id" />-->
        <!--        </el-tab-pane>-->
<!--        <el-tab-pane label="环境变量" name="env_init">-->
<!--          <el-alert title="文件夹初始设置header，优先级低于接口本身的设置" type="info" :closable="false" style="margin: 5px;"/>-->
<!--          <execution-history :folder-id="folder.id"/>-->
<!--        </el-tab-pane>-->
      </el-tabs>
    </div>

    <!-- 编辑文件夹对话框 -->
    <el-dialog
        v-model="editDialogVisible"
        :title="folder.id ? '编辑文件夹' : '新建文件夹'"
        width="500px"
    >
      <folder-edit-form
          :folder="editForm"
          @cancel="editDialogVisible = false"
          @submit="handleSaveFolder"
      />
    </el-dialog>
  </div>
</template>

<script>
import FolderBasicInfo from './FolderBasicInfo.vue'
import FolderEditForm from './FolderEditForm.vue'
import ApiDocument from "@/components/api/ApiDocument.vue";
import Api from "@/utils/base/api";

export default {
  name: 'FolderDetail',
  components: {
    FolderBasicInfo,
    FolderEditForm,
    ApiDocument
  },
  props: {
    folder: {
      type: Object,
      required: true
    },
    handleCreateApi: {  // 接收父组件传递的方法
      type: Function,
      required: true
    },
    activeTabName: {
      type: String,
      default: 'basic'
    }
  },
  data() {
    return {
      activeTab: this.activeTabName || 'basic',
      editDialogVisible: false,
      editForm: {},
      documentApis: []
    }
  },
  watch: {
    folder: {
      handler(newVal) {
        this.editForm = {...newVal}
        if (this.activeTab === 'api_document') {
          this.loadDocumentApis()
          return
        }
        this.documentApis = Array.isArray(newVal.children) ? [...newVal.children] : []
      },
      immediate: true
    },
    activeTabName: {
      handler(newVal) {
        this.activeTab = newVal || 'basic'
      },
      immediate: true
    }
  },
  methods: {
    handleTabChange(tabName) {
      if (tabName === 'api_document') {
        this.loadDocumentApis()
      }
      this.$emit('tab-change', tabName)
    },
    // 中文注释：文档页需要完整接口详情，不能复用树节点里的基础字段列表。
    loadDocumentApis() {
      const _that = this
      if (!_that.folder || !_that.folder.id) {
        _that.documentApis = []
        return
      }
      Api.FolderDetail({
        dir_id: _that.folder.id,
      }, function (res) {
        if (res.ErrCode !== 0) {
          _that.$message.error(res.ErrMsg || '加载接口文档失败')
          return
        }
        const folderDetail = res.Data && res.Data.dir ? res.Data.dir : {}
        _that.documentApis = Array.isArray(folderDetail.children) ? folderDetail.children : []
      })
    },
    handleEdit() {
      this.editForm = {...this.folder}
      this.editDialogVisible = true
    },

    createApi() {
      this.handleCreateApi({
        folderId: this.folder.id,
        name: '新接口',
        // 其他参数...
      })
    },

    handleCreateSubfolder() {
      this.$emit('createSubfolder', this.folder.id)
    },

    handleSaveFolder(updatedFolder) {
      this.$emit('update', updatedFolder)
      this.editDialogVisible = false
      this.$message.success('保存成功')
    },

    handleUpdateFolder(updatedFolder) {
      const _that = this
      const updateData = {
        id: updatedFolder.id,
        name: updatedFolder.name,
        desc: updatedFolder.desc || '',
        collection_id: updatedFolder.collection_id,
        headers: updatedFolder.headers || {}
      }
      Api.CreateDir(updateData, function (res) {
        if (res.ErrCode === 0) {
          _that.$message.success('更新成功')
          // Emit the updated folder with all fields to the parent
          _that.$emit('update', {
            ...updatedFolder,
            ...res.Data,
            headers: res.Data && res.Data.headers ? res.Data.headers : JSON.stringify(updateData.headers || {})
          })
        } else {
          _that.$message.error(res.ErrMsg || '更新失败')
        }
      })
    },

    handleEditApi(api) {
      this.$emit('editApi', api)
    },

    handleExecuteApi(api) {
      this.$emit('executeApi', api)
    },
    handleDelete: function (folder) {
      this.$emit('delete', folder)
    },

    openDocumentPage() {
      const url = `#/ApiDocument/${this.folder.id}`
      window.open(url, '_blank')
    },
  }
}
</script>

<style scoped>
.folder-detail {
  padding: 12px;
  border: 1px solid #e8eee5;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 6px 18px rgba(80, 110, 80, 0.08);
}

.folder-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 10px 12px;
  border: 1px solid #e6ece0;
  border-radius: 10px;
  background: #f7f9f5;
}

.folder-title {
  margin: 0;
  color: #303133;
}

.folder-actions {
  display: flex;
  gap: 12px;
}

.folder-content {
  margin-top: 20px;
}
</style>

