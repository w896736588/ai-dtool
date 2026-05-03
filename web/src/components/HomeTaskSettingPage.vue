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
      <div class="home-task-setting-tabs">
        <div
          v-for="tab in tabs"
          :key="tab.name"
          class="home-task-setting-tab"
          :class="{ 'home-task-setting-tab--active': activeTab === tab.name }"
          @click="activeTab = tab.name"
        >
          {{ tab.label }}
        </div>
      </div>
      <div class="home-task-setting-content">
        <HomeTaskReportSetting :active-tab="activeTab" />
      </div>
    </div>
  </div>
</template>

<script>
import { HomeFilled } from '@element-plus/icons-vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import HomeTaskReportSetting from '@/components/set/home_task_report.vue'

const TABS = [
  { name: 'daily-report', label: '工作日报 AI' },
  { name: 'prompt-template', label: '工作流提示词' },
  { name: 'dev-environment', label: '开发环境提示词' },
  { name: 'tapd', label: 'TAPD 抓取' },
]

export default {
  name: 'HomeTaskSettingPage',
  components: {
    HomeFilled,
    GitActionButton,
    HomeTaskReportSetting,
  },
  data() {
    return {
      activeTab: TABS[0].name,
      tabs: TABS,
    }
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
  /* 页面高度改为自适应，避免内部 Tab 内容被固定视口裁切。 */
  /* Use adaptive page height so nested tab content is not clipped by the viewport. */
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f7f8f6;
}

.home-task-setting-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 28px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  flex-shrink: 0;
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
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fff;
  margin: 12px 16px 16px;
  border-radius: 10px;
  border: 1px solid #e8e8e0;
  overflow: visible;
}

.home-task-setting-tabs {
  display: flex;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 16px;
  flex-shrink: 0;
}

.home-task-setting-tab {
  padding: 12px 18px;
  font-size: 14px;
  color: #606266;
  cursor: pointer;
  border-bottom: 2px solid transparent;
  transition: color 0.2s, border-color 0.2s;
  white-space: nowrap;
}

.home-task-setting-tab:hover {
  color: #3a7a3a;
}

.home-task-setting-tab--active {
  color: #3a7a3a;
  font-weight: 600;
  border-bottom-color: #3a7a3a;
}

.home-task-setting-content {
  flex: 1;
  overflow: visible;
  padding: 16px 20px;
}
</style>
