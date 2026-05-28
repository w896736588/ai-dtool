import { createRouter, createWebHashHistory } from 'vue-router'
import Home from '@/components/Home'
import fullPageShellOut from '@/components/shellout/ShellOut.vue'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      mode: 'hash', // 使用 hash 模式 / Use hash mode for local routes.
      component: Home,
      redirect: '/Dashboard',
      children: [
        {
          path: '/Dashboard',
          name: 'Dashboard',
          components: {
            home: () => import('../components/Dashboard'),
          },
          meta: { keepAlive: true, title: '命令行' },
        },
        {
          path: '/Redis',
          name: 'Redis',
          components: {
            home: () => import('../components/Redis'),
          },
          meta: { keepAlive: true, title: 'Redis' },
        },
        {
          path: '/Supervisor',
          name: 'Supervisor',
          components: {
            home: () => import('../components/Supervisor'),
          },
          meta: { keepAlive: true, title: 'Supervisor' },
        },
        {
          path: '/Git',
          name: 'Git',
          components: {
            home: () => import('../components/Git'),
          },
          meta: { keepAlive: true, title: '分支管理' },
        },
        {
          path: '/CommonActions',
          name: 'CommonActions',
          components: {
            home: () => import('../components/tools/CommonActions.vue'),
          },
          meta: { keepAlive: true, title: '常用操作' },
        },
        {
          path: '/Variable',
          name: 'Variable',
          components: {
            home: () => import('../components/Variable'),
          },
          meta: { keepAlive: true, title: '自定义脚本' },
        },
        {
          path: '/Link',
          name: 'Link',
          components: {
            home: () => import('../components/Link'),
          },
          meta: { keepAlive: true, title: '自定义网页' },
        },
        {
          path: '/Set',
          name: 'Set',
          components: {
            home: () => import('../components/Set'),
          },
          meta: { keepAlive: true, title: '设置' },
        },
        {
          path: '/Tools',
          name: 'Tools',
          components: {
            home: () => import('../components/Tools'),
          },
          meta: { keepAlive: true, title: '工具' },
        },
        {
          path: '/Docker',
          name: 'Docker',
          components: {
            home: () => import('../components/Docker'),
          },
          meta: { keepAlive: true, title: '容器管理' },
        },
        {
          path: '/Markdown',
          name: 'Markdown',
          components: {
            home: () => import('../components/Markdown'),
          },
          meta: { keepAlive: true, title: 'Markdown' },
        },
        {
          path: '/MemoryFragment',
          name: 'MemoryFragment',
          components: {
            home: () => import('../components/MemoryFragment'),
          },
          meta: { keepAlive: true, title: '知识片段' },
        },
        {
          path: '/shellout',
          name: 'shellout',
          components: {
            home: () => import('../components/ShellOut'),
          },
          meta: { keepAlive: true, title: '日志监控' },
        },
        {
          path: '/Api',
          name: 'api',
          components: {
            home: () => import('../components/Api'),
          },
          meta: { keepAlive: true, title: '接口管理' },
        },
        {
          path: '/HomeTask',
          name: 'HomeTask',
          components: {
            home: () => import('../components/HomeTask'),
          },
          meta: { keepAlive: true, title: '工作流程' },
        },
        {
          path: '/Mcp',
          name: 'Mcp',
          components: {
            home: () => import('../components/mcp/McpList'),
          },
          meta: { keepAlive: true, title: 'MCP' },
        },
        {
          path: '/Mcp/:mcpType',
          name: 'McpBinding',
          components: {
            home: () => import('../components/mcp/McpBinding'),
          },
          meta: { keepAlive: false, title: 'MCP 绑定' },
        },
        {
          path: '/AgentCli',
          name: 'AgentCli',
          components: {
            home: () => import('../components/agent_cli/AgentCliList'),
          },
          meta: { keepAlive: true, title: 'Agent Cli' },
        },
      ],
    },
    {
      path: '/ApiDocument/:folderId',
      name: 'api-document',
      component: () => import('../components/ApiDocumentPage'),
      meta: { keepAlive: false, title: '接口文档' },
    },
    {
      path: '/TaskWorkflow/:taskId',
      name: 'task-workflow',
      component: () => import('../components/TaskWorkflow.vue'),
      meta: { keepAlive: false, title: '工作流程' },
    },
    {
      path: '/HomeTaskSetting',
      name: 'home-task-setting',
      component: () => import('../components/HomeTaskSettingPage.vue'),
      meta: { keepAlive: false, title: '任务设置' },
    },
    {
      path: '/MemoryFragmentShare',
      name: 'memory-fragment-share',
      component: () => import('../components/memory/MemoryFragmentShare.vue'),
      meta: { keepAlive: false, title: '知识片段分享' },
    },
    {
      path: '/ChatReply/:chatId',
      name: 'chat-reply',
      component: () => import('../components/ChatReplyPage.vue'),
      meta: { keepAlive: false, title: '对话回复' },
    },
    {
      path: '/fullpage',
      name: 'fullpage',
      mode: 'hash', // 使用 hash 模式 / Use hash mode for local routes.
      component: fullPageShellOut,
      meta: {
        fullScreen: true,
        title: '日志监控',
      },
    },
  ],
})

// 全局导航守卫：根据当前路由的 meta.title 动态设置页面标题
router.afterEach((to) => {
  const title = to.meta?.title || to.name || to.path
  if (title) {
    document.title = title
  }
})

export default router
