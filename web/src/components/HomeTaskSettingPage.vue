<template>
  <div class="home-task-setting-page">
    <header class="home-task-setting-header">
      <div class="home-task-setting-header__main">
        <div class="home-task-setting-header__eyebrow">任务清单</div>
        <h1 class="home-task-setting-header__title">设置</h1>
      </div>
      <div class="home-task-setting-header__actions">
        <el-tooltip content="返回首页" placement="bottom">
          <el-button class="task-workflow-home-btn" @click="goHome">
            <el-icon :size="18"><HomeFilled /></el-icon>
          </el-button>
        </el-tooltip>
        <GitActionButton compact variant="info" @click="goBackToTaskList">
          返回任务清单
        </GitActionButton>
      </div>
    </header>
    <div class="home-task-setting-body">
      <HomeTaskReportSetting />
    </div>
  </div>
</template>

<script>
import { HomeFilled } from '@element-plus/icons-vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import HomeTaskReportSetting from '@/components/set/home_task_report.vue'

export default {
  name: 'HomeTaskSettingPage',
  components: {
    HomeFilled,
    GitActionButton,
    HomeTaskReportSetting,
  },
  methods: {
    goHome() {
      const routeInfo = this.$router.resolve({ path: '/Dashboard' })
      window.open(routeInfo.href, '_blank')
    },
    goBackToTaskList() {
      const routeInfo = this.$router.resolve({ path: '/HomeTask' })
      window.open(routeInfo.href, '_blank')
    },
  },
}
</script>

<style scoped>
.home-task-setting-page {
  min-height: 100vh;
  background: #f7f8f6;
  padding: 0;
}

.home-task-setting-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 28px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  position: sticky;
  top: 0;
  z-index: 10;
}

.home-task-setting-header__eyebrow {
  font-size: 12px;
  color: #909399;
  margin-bottom: 2px;
}

.home-task-setting-header__title {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.home-task-setting-header__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.home-task-setting-body {
  padding: 24px 28px;
}
</style>
