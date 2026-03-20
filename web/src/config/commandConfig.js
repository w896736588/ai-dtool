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
  // 历史命令
  {
    command: 'history',
    name: '历史命令',
    icon: '🕘',
    desc: '查看历史命令并回填输入框',
    aliases: ['his', 'h', '历史', '历史记录'],
    module: null,
    needTarget: true,
    dynamicChildren: 'historyList'
  },

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
        action: 'dockerServices',
        needTarget: true,
        dynamicChildren: 'dockerComposeList'
      },
      {
        command: 'status',
        name: '运行状态',
        desc: '查看运行状态',
        action: 'dockerStatus',
        needTarget: true,
        dynamicChildren: 'dockerComposeList'
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
    aliases: ['g', 'gi', '代码管理', '仓库'],
    module: 'git',
    path: '/Git',
    children: [
      {
        command: 'pull',
        name: '拉取',
        desc: '拉取远程代码',
        aliases: ['pl', 'update', 'sync', '拉取', '更新'],
        action: 'gitPull',
        needTarget: true,
        dynamicChildren: 'gitProjectList'
      },
      {
        command: 'status',
        name: '状态',
        desc: '查看状态',
        aliases: ['st', 'stat', 'check', '状态', '状态检查'],
        action: 'gitStatus',
        needTarget: true,
        dynamicChildren: 'gitProjectList'
      },
      {
        command: 'branch',
        name: '当前分支',
        desc: '查看当前分支',
        aliases: [
          'br',
          'current',
          'curr',
          'show',
          'show-branch',
          'showbranch',
          'current-branch',
          'branch',
          '当前',
          '当前分支',
          '分支'
        ],
        action: 'gitBranch',
        needTarget: true,
        dynamicChildren: 'gitProjectList'
      },
      {
        command: 'group-branches',
        name: '组内分支总览',
        desc: '查看某个Git组内所有环境当前分支和远程分支',
        aliases: ['gb', 'group-branch', 'group-branches', '组分支', '分组分支', '组内分支'],
        action: 'gitGroupBranches',
        needTarget: true,
        dynamicChildren: 'gitGroupList'
      },
      {
        command: 'log',
        name: '日志',
        desc: '查看提交日志',
        aliases: ['lg', 'history', 'his', 'commit', '日志', '提交记录'],
        action: 'gitLog',
        needTarget: true,
        dynamicChildren: 'gitProjectList'
      },
      {
        command: 'checkout',
        name: '切换分支',
        desc: '切换到指定分支',
        aliases: ['ch', 'co', 'switch', '切换', '切分支', '切换分支', '换分支'],
        action: 'gitCheckout',
        needTarget: true,
        dynamicChildren: 'gitProjectList',
        needInput: true,
        inputPlaceholder: '请输入要切换的分支名'
      },
      {
        command: 'checkout-remote',
        name: '关联远程分支切换',
        desc: '切换并关联远程分支',
        aliases: ['chr', 'cor', 'remote', 'track', '关联', '关联远程', '远程切换'],
        action: 'gitCheckoutRemote',
        needTarget: true,
        dynamicChildren: 'gitProjectList',
        needInput: true,
        inputPlaceholder: '请输入远程分支名'
      },
      {
        command: 'quick-create-branch',
        name: '快捷创建分支',
        desc: '按仓库/基线分支/分支类型快速建分支并推送',
        aliases: ['qcb', 'quick-branch', 'new-branch', '快捷建分支'],
        action: 'gitQuickCreateBranch',
        needTarget: true,
        dynamicChildren: 'gitProjectList',
        nextDynamicChildren: 'gitRemoteBranchList',
        needInput: true,
        inputPlaceholder: '请输入业务英文（仅英文、数字、下划线）'
      },
      {
        command: 'save-credentials',
        name: '保存账号密码配置',
        desc: '执行 git credential 配置',
        aliases: ['save', 'cred', 'remember', 'credential', '记住密码', '保存凭据'],
        action: 'gitSaveCredentials',
        needTarget: true,
        dynamicChildren: 'gitProjectList'
      },
      {
        command: 'set-safe',
        name: '设置目录安全',
        desc: '将当前目录加入 git safe.directory',
        aliases: ['safe', 'trust', 'safe-dir', '安全目录', '信任目录'],
        action: 'gitSetSafe',
        needTarget: true,
        dynamicChildren: 'gitProjectList'
      },
      {
        command: 'view-config',
        name: '查看 git config 文档',
        desc: '跳转到 Git 页面查看文档',
        aliases: ['cfg', 'config', 'git-config', 'show-config', '配置'],
        action: 'gitViewConfig'
      },
      {
        command: 'help',
        name: 'Git 帮助',
        desc: '跳转到 Git 页面帮助',
        aliases: ['h', '?', 'doc', 'docs', '帮助', '文档'],
        action: 'gitHelp'
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
        dynamicChildren: 'supervisorEnvList',
        nextDynamicChildren: 'supervisorProcessList'
      },
      {
        command: 'stop',
        name: '停止进程',
        desc: '停止指定进程',
        action: 'supervisorStop',
        needTarget: true,
        dynamicChildren: 'supervisorEnvList',
        nextDynamicChildren: 'supervisorProcessList'
      },
      {
        command: 'config',
        name: '查看配置',
        desc: '查看进程配置',
        action: 'supervisorConfig',
        needTarget: true,
        dynamicChildren: 'supervisorEnvList',
        nextDynamicChildren: 'supervisorProcessList'
      },
      {
        command: 'refresh-cache',
        name: '刷新缓存',
        desc: '刷新重启服务列表缓存',
        action: 'supervisorRefreshProcessCache',
        needTarget: true,
        dynamicChildren: 'supervisorEnvList'
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
    action: 'shell',
    needTarget: true,
    dynamicChildren: 'shellOutList'
  },
  {
    command: 'link',
    name: '自定义网页',
    icon: '🔗',
    desc: '自定义网页链接',
    module: 'login',
    path: '/Link',
    action: 'linkRun',
    needTarget: true,
    dynamicChildren: 'linkEnvList'
  }
]

export default commandConfig
