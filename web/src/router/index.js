import {createRouter, createWebHashHistory} from 'vue-router'
import Home from '@/components/Home'
import fullPageShellOut from '@/components/shellout/ShellOut.vue'
export default createRouter({
    history: createWebHashHistory(),
    routes: [
        {
            path: '/',
            name: 'home',
            mode: 'hash', //使用hash模式 本地路由
            component: Home,
            redirect: '/Dashboard',
            children: [
                {
                    path: '/Dashboard',
                    name: 'Dashboard',
                    components: {
                        home: () => import('../components/Dashboard'),
                    },
                    meta: {keepAlive: true},
                },
                {
                    path: '/Redis',
                    name: 'Redis',
                    components: {
                        home: () => import('../components/Redis'), //指定其中的home  router view
                    },
                    meta: {keepAlive: true},
                },
                {
                    path: '/Supervisor',
                    name: 'Supervisor',
                    components: {
                        home: () => import('../components/Supervisor'), //指定其中的home  router view
                    },
                    meta: {keepAlive: true},
                },
                {
                    path: '/Git',
                    name: 'Git',
                    components: {
                        home: () => import('../components/Git'), //指定其中的home  router view
                    },
                    meta: {keepAlive: true},
                },
                {
                    path: '/Variable',
                    name: 'Variable',
                    components: {
                        home: () => import('../components/Variable'), //指定其中的home  router view
                    },
                    meta: {keepAlive: true},
                },
                {
                    path: '/Link',
                    name: 'Link',
                    components: {
                        home: () => import('../components/Link'), //指定其中的home  router view
                    },
                    meta: {keepAlive: true},
                },
                {
                    path: '/Set',
                    name: 'Set',
                    components: {
                        home: () => import('../components/Set'), //指定其中的home  router view
                    },
                    meta: {keepAlive: true},
                },
                {
                    path: '/Tools',
                    name: 'Tools',
                    components: {
                        home: () => import('../components/Tools'), //指定其中的home  router view
                    },
                    meta: {keepAlive: true},
                },
                {
                    path: '/Docker',
                    name: 'Docker',
                    components: {
                        home: () => import('../components/Docker'), //指定其中的home  router view
                    },
                    meta: {keepAlive: true},
                },
                {
                    path: '/Markdown',
                    name: 'Markdown',
                    components: {
                        home: () => import('../components/Markdown'), //指定其中的home  router view
                    },
                    meta: {keepAlive: true},
                },
                {
                    path: '/shellout',
                    name: 'shellout',
                    components: {
                        home: () => import('../components/ShellOut'), //指定其中的home  router view
                    },
                    meta: {keepAlive: true},
                },
                {
                    path: '/Api',
                    name: 'api',
                    components: {
                        home: () => import('../components/Api'), //指定其中的home  router view
                    },
                    meta: {keepAlive: true},
                },
                // {
                //   path : '/Code',
                //   components: {
                //     home: () => import('../components/Code'), //指定其中的home  router view
                //   },
                //   meta: { keepAlive: true },
                // }
            ],
        },
        {
            path: '/ApiDocument/:folderId',
            name: 'api-document',
            component: () => import('../components/ApiDocumentPage'), //独立接口文档页面
            meta: {keepAlive: false},
        },
        {
            path: '/fullpage',
            name: 'fullpage',
            mode: 'hash', //使用hash模式 本地路由
            component: fullPageShellOut,
            meta: {
                fullScreen: true  // 可选的元信息，用于标识全屏页面
            }
        },
    ],
})
