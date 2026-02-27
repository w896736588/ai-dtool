/**
 * 命令自动补全配置文件
 * 支持多级命令补全，类似终端的自动补全体验
 * 
 * 配置结构说明：
 * - command: 命令关键字（用户输入的触发词）
 * - name: 显示名称
 * - icon: 图标
 * - desc: 描述
 * - module: 模块名（用于权限过滤）
 * - path: 跳转路径（可选，如果有则直接跳转）
 * - children: 子命令列表（可选，用于多级补全）
 * - dynamicChildren: 动态获取子命令的函数名（可选，从API获取）
 * - action: 执行动作（可选，用于执行特定操作）
 */

const commandConfig = [
  // Docker 命令
  {
    command: 'docker',
    name: 'Docker',
    icon: '🐳',
    desc: 'Docker容器管理',
    module: 'docker',
    path: '/Docker',
    children: [
      {
        command: 'ps',
        name: '服务列表',
        desc: '查看服务列表',
        action: 'dockerServices'
      },
      {
        command: 'status',
        name: '运行状态',
        desc: '查看运行状态',
        action: 'dockerStatus'
      },
      {
        command: 'up',
        name: '启动',
        desc: '启动容器 (up -d)',
        action: 'dockerUp',
        needTarget: true,
        dynamicChildren: 'dockerComposeList'
      },
      {
        command: 'restart',
        name: '重启',
        desc: '重启容器',
        action: 'dockerRestart',
        needTarget: true,
        dynamicChildren: 'dockerComposeList'
      },
      {
        command: 'stop',
        name: '停止',
        desc: '停止容器',
        action: 'dockerStop',
        needTarget: true,
        dynamicChildren: 'dockerComposeList'
      },
      {
        command: 'config',
        name: '查看配置',
        desc: '查看 compose.yml',
        action: 'dockerConfig',
        needTarget: true,
        dynamicChildren: 'dockerComposeList'
      },
      {
        command: 'env',
        name: '查看env',
        desc: '查看环境变量',
        action: 'dockerEnv',
        needTarget: true,
        dynamicChildren: 'dockerComposeList'
      },
      {
        command: 'quick-restart',
        name: '快速重启',
        desc: '快速重启默认服务',
        needTarget: true,
        dynamicChildren: 'dockerComposeList',
        // 选择项目后，自动加载服务列表
        nextDynamicChildren: 'dockerServiceList',
        action: 'dockerQuickRestart'
      },
      {
        command: 'quick-stop',
        name: '快速停止',
        desc: '快速停止默认服务',
        needTarget: true,
        dynamicChildren: 'dockerComposeList',
        // 选择项目后，自动加载服务列表
        nextDynamicChildren: 'dockerServiceList',
        action: 'dockerQuickStop'
      }
    ]
  },

  // Git 命令
  {
    command: 'git',
    name: 'Git',
    icon: '📚',
    desc: 'Git管理',
    module: 'git',
    path: '/Git',
    children: [
      {
        command: 'pull',
        name: '拉取',
        desc: '拉取远程代码',
        action: 'gitPull',
        needTarget: true,
        dynamicChildren: 'gitProjectList'
      },
      {
        command: 'status',
        name: '状态',
        desc: '查看状态',
        action: 'gitStatus',
        needTarget: true,
        dynamicChildren: 'gitProjectList'
      },
      {
        command: 'branch',
        name: '当前分支',
        desc: '查看当前分支',
        action: 'gitBranch',
        needTarget: true,
        dynamicChildren: 'gitProjectList'
      },
      {
        command: 'log',
        name: '日志',
        desc: '查看提交日志',
        action: 'gitLog',
        needTarget: true,
        dynamicChildren: 'gitProjectList'
      },
      {
        command: 'checkout',
        name: '切换分支',
        desc: '切换到指定分支',
        action: 'gitCheckout',
        needTarget: true,
        dynamicChildren: 'gitProjectList',
        needInput: true,
        inputPlaceholder: '请输入分支名'
      }
    ]
  },

  // Supervisor 命令
  {
    command: 'supervisor',
    name: 'Supervisor',
    icon: '⚙️',
    desc: '进程管理',
    module: 'supervisor',
    path: '/Supervisor',
    children: [
      {
        command: 'status',
        name: '查看状态',
        desc: '查看所有进程状态',
        action: 'supervisorStatus',
        needTarget: true,
        dynamicChildren: 'supervisorEnvList'
      },
      {
        command: 'restart-all',
        name: '重启所有',
        desc: '重启所有进程',
        action: 'supervisorRestartAll',
        needTarget: true,
        dynamicChildren: 'supervisorEnvList'
      },
      {
        command: 'restart',
        name: '重启进程',
        desc: '重启指定进程',
        action: 'supervisorRestart',
        needTarget: true,
        dynamicChildren: 'supervisorProcessList'
      },
      {
        command: 'stop',
        name: '停止进程',
        desc: '停止指定进程',
        action: 'supervisorStop',
        needTarget: true,
        dynamicChildren: 'supervisorProcessList'
      },
      {
        command: 'config',
        name: '查看配置',
        desc: '查看进程配置',
        action: 'supervisorConfig',
        needTarget: true,
        dynamicChildren: 'supervisorProcessList'
      }
    ]
  },

  // 终端输出命令
  {
    command: 'shell',
    name: '终端输出',
    icon: '💻',
    desc: '终端输出查看',
    module: 'shellout',
    path: '/shellout',
    children: [
      {
        command: 'create',
        name: '创建',
        desc: '创建新的终端输出任务',
        action: 'shellCreate'
      },
      {
        command: 'list',
        name: '任务列表',
        desc: '查看所有任务',
        action: 'shellList',
        needTarget: true,
        dynamicChildren: 'shellOutList'
      },
      {
        command: 'run',
        name: '运行任务',
        desc: '运行指定任务',
        action: 'shellRun',
        needTarget: true,
        dynamicChildren: 'shellOutList'
      }
    ]
  },

  // Redis 命令
  {
    command: 'redis',
    name: 'Redis',
    icon: '🗃️',
    desc: 'Redis管理',
    module: 'redis',
    path: '/Redis',
    children: [
      {
        command: 'info',
        name: '信息',
        desc: '查看Redis信息',
        action: 'redisInfo',
        needTarget: true,
        dynamicChildren: 'redisEnvList'
      },
      {
        command: 'keys',
        name: '键列表',
        desc: '查看键列表',
        action: 'redisKeys',
        needTarget: true,
        dynamicChildren: 'redisEnvList'
      }
    ]
  },

  // 其他快捷命令
  {
    command: 'api',
    name: '接口开发',
    icon: '🔌',
    desc: 'API接口开发',
    module: 'api',
    path: '/Api'
  },
  {
    command: 'set',
    name: '配置',
    icon: '🔧',
    desc: '系统配置',
    module: null,
    path: '/Set'
  },
  {
    command: 'link',
    name: '自定义网页',
    icon: '🔗',
    desc: '自定义网页链接',
    module: 'login',
    path: '/Link'
  },
  {
    command: 'variable',
    name: '自定义脚本',
    icon: '📝',
    desc: '自定义脚本管理',
    module: 'variable',
    path: '/Variable'
  }
]

export default commandConfig
