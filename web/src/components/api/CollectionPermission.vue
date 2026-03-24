<template>
  <div class="collection-permission">
    <div class="permission-header">
      <pl-button type="primary" @click="handleAddPermission">添加成员</pl-button>
      <pl-button @click="handlePermissionTemplate">权限模板</pl-button>
    </div>

    <el-table :data="permissionList" style="width: 100%" v-loading="loading">
      <el-table-column prop="userName" label="成员" width="180">
        <template #default="{ row }">
          <div class="user-info">
            <el-avatar :size="32" :src="row.avatar" />
            <div class="user-details">
              <div class="user-name">{{ row.userName }}</div>
              <div class="user-email">{{ row.email }}</div>
            </div>
          </div>
        </template>
      </el-table-column>

      <el-table-column prop="role" label="角色" width="150">
        <template #default="{ row }">
          <el-select
              v-model="row.role"
              placeholder="选择角色"
              @change="handleRoleChange(row)"
          >
            <el-option
                v-for="role in roleOptions"
                :key="role.value"
                :label="role.label"
                :value="role.value"
            />
          </el-select>
        </template>
      </el-table-column>

      <el-table-column prop="permissions" label="权限详情">
        <template #default="{ row }">
          <el-tag
              v-for="perm in getPermissionTags(row.role)"
              :key="perm"
              type="info"
              size="small"
              style="margin-right: 8px;"
          >
            {{ perm }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column label="操作" width="120" align="center">
        <template #default="{ row }">
          <pl-button
              type="danger"
              link
              @click="handleRemovePermission(row)"
          >
            移除
          </pl-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 添加成员对话框 -->
    <el-dialog
        v-model="addMemberDialogVisible"
        title="添加成员"
        width="500px"
    >
      <el-form :model="memberForm" label-width="80px">
        <el-form-item label="选择成员">
          <el-select
              v-model="memberForm.userId"
              filterable
              remote
              reserve-keyword
              placeholder="请输入用户名或邮箱搜索"
              :remote-method="searchUsers"
              :loading="searchLoading"
              style="width: 100%"
          >
            <el-option
                v-for="user in userOptions"
                :key="user.id"
                :label="`${user.name} (${user.email})`"
                :value="user.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="分配角色">
          <el-select v-model="memberForm.role" placeholder="选择角色" style="width: 100%">
            <el-option
                v-for="role in roleOptions"
                :key="role.value"
                :label="role.label"
                :value="role.value"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <pl-button @click="addMemberDialogVisible = false">取消</pl-button>
        <pl-button type="primary" @click="handleConfirmAddMember">确认添加</pl-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
export default {
  name: 'CollectionPermission',
  props: {
    collection: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      loading: false,
      permissionList: [],
      addMemberDialogVisible: false,
      searchLoading: false,
      memberForm: {
        userId: '',
        role: 'viewer'
      },
      userOptions: [],
      roleOptions: [
        {
          value: 'owner',
          label: '所有者',
          permissions: ['全部权限']
        },
        {
          value: 'editor',
          label: '编辑者',
          permissions: ['查看', '编辑', '执行', '导出']
        },
        {
          value: 'viewer',
          label: '查看者',
          permissions: ['查看', '执行']
        }
      ]
    }
  },
  watch: {
    collection: {
      handler(newVal) {
        this.loadPermissionData(newVal)
      },
      immediate: true
    }
  },
  methods: {
    loadPermissionData(collection) {
      this.loading = true
      // 模拟加载权限数据
      setTimeout(() => {
        this.permissionList = [
          {
            id: 1,
            userId: 1,
            userName: '管理员',
            email: 'admin@example.com',
            avatar: '',
            role: 'owner'
          },
          {
            id: 2,
            userId: 2,
            userName: '开发人员A',
            email: 'dev1@example.com',
            avatar: '',
            role: 'editor'
          },
          {
            id: 3,
            userId: 3,
            userName: '测试人员B',
            email: 'tester@example.com',
            avatar: '',
            role: 'viewer'
          }
        ]
        this.loading = false
      }, 500)
    },

    getPermissionTags(role) {
      const roleObj = this.roleOptions.find(r => r.value === role)
      return roleObj ? roleObj.permissions : []
    },

    handleAddPermission() {
      this.addMemberDialogVisible = true
      this.memberForm = {
        userId: '',
        role: 'viewer'
      }
    },

    searchUsers(query) {
      if (query) {
        this.searchLoading = true
        // 模拟搜索用户
        setTimeout(() => {
          this.userOptions = [
            {
              id: 4,
              name: '新用户A',
              email: 'newuser@example.com'
            },
            {
              id: 5,
              name: '新用户B',
              email: 'newuser2@example.com'
            }
          ].filter(user =>
              user.name.includes(query) || user.email.includes(query)
          )
          this.searchLoading = false
        }, 500)
      }
    },

    handleConfirmAddMember() {
      if (!this.memberForm.userId) {
        this.$message.error('请选择成员')
        return
      }

      const user = this.userOptions.find(u => u.id === this.memberForm.userId)
      if (user) {
        this.permissionList.push({
          id: Date.now(),
          userId: user.id,
          userName: user.name,
          email: user.email,
          avatar: '',
          role: this.memberForm.role
        })
        this.addMemberDialogVisible = false
        this.$message.success('添加成员成功')
      }
    },

    handleRoleChange(row) {
      this.$message.success(`已更新 ${row.userName} 的角色为 ${this.getRoleLabel(row.role)}`)
    },

    getRoleLabel(roleValue) {
      const role = this.roleOptions.find(r => r.value === roleValue)
      return role ? role.label : roleValue
    },

    handleRemovePermission(row) {
      this.$confirm(`确定要移除成员 "${row.userName}" 吗？`, '确认移除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        const index = this.permissionList.findIndex(item => item.id === row.id)
        if (index !== -1) {
          this.permissionList.splice(index, 1)
          this.$message.success('移除成员成功')
        }
      })
    },

    handlePermissionTemplate() {
      this.$message.info('权限模板功能开发中')
    }
  }
}
</script>

<style scoped>
.collection-permission {
  padding: 12px;
  border: 1px solid #e8eee5;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 6px 18px rgba(80, 110, 80, 0.08);
}

.permission-header {
  margin-bottom: 20px;
  padding: 10px 12px;
  border: 1px solid #e6ece0;
  border-radius: 10px;
  background: #f7f9f5;
}

.user-info {
  display: flex;
  align-items: center;
}

.user-details {
  margin-left: 12px;
  line-height: 1.4;
}

.user-name {
  font-weight: 500;
  color: #4f5f4b;
}

.user-email {
  font-size: 12px;
  color: #909399;
}
</style>

