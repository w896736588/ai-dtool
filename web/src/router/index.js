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
          meta: { keepAlive: true },
        },
        {
          path: '/Redis',
          name: 'Redis',
          components: {
            home: () => import('../components/Redis'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/Supervisor',
          name: 'Supervisor',
          components: {
            home: () => import('../components/Supervisor'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/Git',
          name: 'Git',
          components: {
            home: () => import('../components/Git'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/CommonActions',
          name: 'CommonActions',
          components: {
            // 独立主菜单页 / Standalone main menu page.
            home: () => import('../components/tools/CommonActions.vue'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/Variable',
          name: 'Variable',
          components: {
            home: () => import('../components/Variable'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/Link',
          name: 'Link',
          components: {
            home: () => import('../components/Link'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/Set',
          name: 'Set',
          components: {
            home: () => import('../components/Set'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/Tools',
          name: 'Tools',
          components: {
            home: () => import('../components/Tools'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/Docker',
          name: 'Docker',
          components: {
            home: () => import('../components/Docker'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/Markdown',
          name: 'Markdown',
          components: {
            home: () => import('../components/Markdown'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/MemoryFragment',
          name: 'MemoryFragment',
          components: {
            home: () => import('../components/MemoryFragment'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/shellout',
          name: 'shellout',
          components: {
            home: () => import('../components/ShellOut'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/Api',
          name: 'api',
          components: {
            home: () => import('../components/Api'),
          },
          meta: { keepAlive: true },
        },
        {
          path: '/HomeTask',
          name: 'HomeTask',
          components: {
            home: () => import('../components/HomeTask'),
          },
          meta: { keepAlive: true },
        },
      ],
    },
    {
      path: '/ApiDocument/:folderId',
      name: 'api-document',
      component: () => import('../components/ApiDocumentPage'),
      meta: { keepAlive: false },
    },
    {
      path: '/TaskWorkflow/:taskId',
      name: 'task-workflow',
      component: () => import('../components/TaskWorkflow.vue'),
      meta: { keepAlive: false },
    },
    {
      path: '/MemoryFragmentShare',
      name: 'memory-fragment-share',
      component: () => import('../components/memory/MemoryFragmentShare.vue'),
      meta: { keepAlive: false },
    },
    {
      path: '/fullpage',
      name: 'fullpage',
      mode: 'hash', // 使用 hash 模式 / Use hash mode for local routes.
      component: fullPageShellOut,
      meta: {
        fullScreen: true, // 全屏页标记 / Full screen page marker.
      },
    },
  ],
})

// 全局导航守卫：根据当前路由名称动态设置页面标题
router.afterEach((to) => {
  const title = to.name || to.path
  if (title) {
    document.title = title
  }
})

export default router
