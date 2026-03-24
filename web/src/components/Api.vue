<template>
  <div class="collection-container" :class="{ resizing: sidebarResizing }">
    <!-- 左侧集合列表 -->
    <div class="left-sidebar" :style="{ width: sidebarWidth + 'px' }">
      <!-- 集合列表区域 -->
      <div class="collection-section">
        <div class="section-header">
          <span class="section-header-title">集合列表</span>
          <div class="section-header-actions">
            <pl-button class="toolbar-btn toolbar-btn-small" size="small" type="primary" plain @click="openCreateCollectionDialog">
              <el-icon><Plus /></el-icon>新建集合
            </pl-button>
            <pl-button class="toolbar-btn toolbar-btn-small" size="small" type="primary" plain @click="drawerVisibleMarkdown = true">
              <el-icon><QuestionFilled /></el-icon>文档
            </pl-button>
            <el-popover placement="bottom-end" :width="360" trigger="click">
              <template #reference>
                <pl-button class="toolbar-btn toolbar-btn-mini" size="small" type="info" plain>
                  <el-icon><Tools /></el-icon>Skills安装
                </pl-button>
              </template>
              <div class="skill-install-popover">
                <div class="skill-install-title">dtool-api-import-update 安装</div>
                <div class="skill-install-text">把下面 ZIP 地址发给编辑器 AI，让它按 ZIP 安装 skills。</div>
                <div class="skill-install-feature">
                  功能说明：安装后可通过 AI 直接在此工具中生成接口（支持集合/文件夹选择、按 URI 导入或更新）。
                </div>
                <el-input :model-value="skillInstallZipUrl" readonly class="skill-install-url" />
                <div class="skill-install-actions">
                  <pl-button size="small" type="primary" plain @click="copyText(skillInstallZipUrl, 'ZIP 地址已复制')">复制 ZIP 地址</pl-button>
                  <pl-button size="small" type="primary" plain @click="copyText(skillInstallPrompt, 'AI 安装提示已复制')">复制安装提示</pl-button>
                  <el-link :href="skillInstallZipUrl" target="_blank" type="primary">打开链接</el-link>
                </div>
                <el-input
                  type="textarea"
                  :rows="2"
                  readonly
                  class="skill-install-url"
                  :model-value="'请先安装这个 skills zip：' + skillInstallZipUrl + '。'"
                />
              </div>
            </el-popover>
          </div>
        </div>
        <div class="collection-list">
          <div class="tree-wrapper">
            <!--            :default-expanded-keys="defaultExpandedKeys" 指定节点默认展开-->
            <el-tree
                ref="collectionTreeRef"
                :data="treeData"
                :default-expand-all="false"
                :expand-on-click-node="false"
                :highlight-current="true"
                :lazy="true"
                :load="loadTreeNode"
                :props="treeProps"
                :draggable="true"
                :allow-drop="allowTreeNodeDrop"
                :allow-drag="allowTreeNodeDrag"
                node-key="uniqueid"
                tabindex="0" @keyup="handleKeyUp"
                @node-click="handleNodeClick"
                @node-expand="handleNodeExpand"
                @node-collapse="handleNodeCollapse"
                @node-drop="handleTreeNodeDrop"
            >
              <template #default="{ node, data }">
                <div class="tree-node" @dblclick.stop="handleNodeDoubleClick(data)">
                  <span class="node-icon">
                    <el-icon v-if="data.type === 'collection'"><Files/></el-icon>
                    <el-icon v-else-if="data.type === 'folder'"><Folder/></el-icon>
                    <!--                    <el-icon v-else><Document /></el-icon>-->
                  </span>
                  <span v-if="data.type === 'folder'" :title="node.label + '(' + getNodeChildCount(data) + ')'" class="node-label" style="font-weight: 500;">{{
                      node.label + '(' + getNodeChildCount(data) + ')'
                    }}</span>
                  <span v-if="data.type === 'collection'" :title="node.label + '(' + getNodeChildCount(data) + ')'" class="node-label" style="font-weight: 800;">{{
                      node.label + '(' + getNodeChildCount(data) + ')'
                    }}</span>
                  <span v-if="data.type === 'api'" :title="node.label" class="node-label">
                    <el-tag v-if="data.method === 'GET'" size="small" type="success">G</el-tag>
                    <el-tag v-if="data.method === 'POST'" size="small" type="primary">P</el-tag>
                    {{ node.label }}
                  </span>
                  <span v-if="data.type === 'collection'" class="node-actions">
                    <pl-button link type="primary" @click.stop="toggleCollection(data)">
                      <el-dropdown>
                      <pl-button link type="primary" @click.stop>
                        <el-icon><More/></el-icon>
                      </pl-button>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item command="copy_api" icon="CopyDocument" @click="createNewDir(data)">创建文件夹</el-dropdown-item>
                          <el-dropdown-item command="json_import" icon="Upload" @click="openJsonImportDialog(data)">通过json导入</el-dropdown-item>
                          <el-dropdown-item command="delete_collection" icon="Delete" @click="handleCollectionDelete(data)" style="color:red;">删除集合</el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                      </el-dropdown>
                    </pl-button>
                  </span>
                  <span v-else-if="data.type === 'api'" class="node-actions">
                    <el-dropdown>
                      <pl-button link type="primary" @click.stop>
                        <el-icon><More/></el-icon>
                      </pl-button>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item command="copy_api" icon="CopyDocument" @click="handleApiAction('copy_api' , data)">复制接口</el-dropdown-item>
                          <el-dropdown-item command="delete_api" icon="Move" @click="handleApiAction('delete_api' , data)">删除接口</el-dropdown-item>
                          <el-dropdown-item command="down_api" icon="Move" @click="handleApiAction('down_api' , data)">下移</el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                  </span>
                  <span v-else-if="data.type === 'folder'" class="node-actions">
                    <el-dropdown>
                      <pl-button link type="primary" @click.stop>
                        <el-icon><More/></el-icon>
                      </pl-button>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item command="copy_api" icon="CopyDocument" @click="handleFolderCreateApi()">创建接口</el-dropdown-item>
                          <el-dropdown-item command="delete_folder" icon="Delete" style="color:red;" @click="handleFolderDelete(data)">删除</el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                  </span>
                </div>
              </template>
            </el-tree>
          </div>
        </div>
      </div>

      <!-- 归档列表区域 -->
      <!--      <div class="archive-section">-->
      <!--        <div class="section-header" @click="toggleArchive">-->
      <!--          <span>归档列表</span>-->
      <!--          <el-icon :class="['collapse-icon', { 'rotate-180': archiveExpanded }]">-->
      <!--            <ArrowDown />-->
      <!--          </el-icon>-->
      <!--        </div>-->
      <!--        <div v-show="archiveExpanded" class="archive-list">-->
      <!--          <div-->
      <!--              v-for="item in archivedItems"-->
      <!--              :key="item.id"-->
      <!--              class="archive-item"-->
      <!--              @click="handleArchiveItemClick(item)"-->
      <!--          >-->
      <!--            <el-icon><Document /></el-icon>-->
      <!--            <span>{{ item.name }}</span>-->
      <!--          </div>-->
      <!--        </div>-->
      <!--      </div>-->
    </div>

    <!-- 左右面板拖拽分隔条 -->
    <div class="panel-resizer" @mousedown="startSidebarResize"></div>

    <!-- 右侧展示面板 -->
    <div class="right-panel">
      <!-- 工作区 Tab 条 -->
      <div v-if="openTabs.length > 0" class="workspace-tabs">
        <el-tabs
            v-model="activeTabKey"
            type="card"
            closable
            @tab-remove="closeWorkspaceTab"
            @tab-change="handleWorkspaceTabChange"
        >
          <el-tab-pane
              v-for="tab in openTabs"
              :key="tab.key"
              :label="tab.title"
              :name="tab.key"
          >
            <template #label>
              <span class="workspace-tab-label">
                <el-icon v-if="tab.type === 'collection'" class="tab-icon"><Files /></el-icon>
                <el-icon v-else-if="tab.type === 'folder'" class="tab-icon"><Folder /></el-icon>
                <el-icon v-else class="tab-icon"><Document /></el-icon>
                <span class="tab-title">{{ tab.title }}</span>
              </span>
            </template>
          </el-tab-pane>
        </el-tabs>
      </div>

      <!-- 顶部信息栏（仅集合 tab 显示） -->
      <div v-if="selectedItem && selectedItem.type === 'collection'" class="panel-header">
        <div class="header-left">
          <div v-if="selectedItem" class="selected-info">
            <el-icon class="info-icon">
              <FolderOpened v-if="selectedItem.type === 'collection'"/>
              <Folder v-else-if="selectedItem.type === 'folder'"/>
              <Document v-else/>
            </el-icon>
            <div class="info-content">
              <div class="info-name">{{ selectedItem.name }}</div>
              <div v-if="selectedItem.desc" class="info-desc">
                {{ selectedItem.desc }}
              </div>
            </div>
          </div>
          <div v-else class="no-selection">
            请选择集合、文件夹或接口
          </div>
        </div>

        <div v-if="selectedItem" class="header-right">
          <pl-button v-if="selectedItem.type === 'collection'" type="primary" @click="createNewDir(selectedItem)">创建文件夹
          </pl-button>
          <pl-button type="primary" @click="executeAll">运行全部</pl-button>
          <pl-button @click="exportCollection">导出</pl-button>
        </div>
      </div>

      <!-- 内容区域 -->
      <div :class="['panel-content', { 'panel-content--flush': selectedItem && selectedItem.type === 'folder' }]">
        <!-- 集合设置 -->
        <div v-if="selectedItem && selectedItem.type === 'collection'" class="collection-settings">
          <el-tabs :model-value="getActiveCollectionInnerTab(selectedItem)" type="card" @tab-change="handleCollectionInnerTabChange">
            <el-tab-pane label="基本信息" name="basic">
              <collection-basic-info
                  :collection="selectedItem"
                  @delete="handleCollectionDelete"
                  @update="handleCollectionUpdate"
              />
            </el-tab-pane>
            <el-tab-pane label="环境变量" name="environment">
              <collection-environment
                  :collection="selectedItem"
                  :environments="environmentList"
              />
            </el-tab-pane>
            <!--            <el-tab-pane label="权限设置" name="permission">-->
            <!--              <collection-permission :collection="selectedItem"/>-->
            <!--            </el-tab-pane>-->
          </el-tabs>
        </div>

        <!-- 文件夹信息 -->
        <div v-else-if="selectedItem && selectedItem.type === 'folder'" class="folder-info">
          <folder-detail
              :folder="selectedItem"
              :active-tab-name="getActiveFolderInnerTab(selectedItem)"
              :handleCreateApi="handleFolderCreateApi"
              @tab-change="handleFolderInnerTabChange"
              @delete="handleFolderDelete"
              @update="handleFolderUpdate"
          />
        </div>

        <!-- 接口设置 -->
        <div v-else-if="selectedItem && selectedItem.type === 'api'" class="api-settings">
          <el-skeleton v-if="apiDetailLoading" :rows="12" animated />
          <api-detail
              v-else
              ref="refApiDetail"
              :environment="currentEnvironment"
              @execute="executeApi"
              @update="handleApiUpdate"
          />
        </div>

        <!-- 空状态 -->
        <div v-else class="empty-state">
          <el-empty description="请从左侧选择集合、文件夹或接口"/>
        </div>
      </div>
    </div>
  </div>

  <el-dialog v-model="dialogShow.createCollection" title="创建集合" width="500" @keydown.enter.prevent="handleDialogEnter('createCollection', $event)">
    <el-form @submit.prevent>
      <el-form-item :label-width="80" label="集合名称">
        <el-input v-model="dialogData.createCollection.name" autocomplete="off"/>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="dialogShow.createCollection = false">取消</pl-button>
        <pl-button type="primary" @click="createNewCollection">保存</pl-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="dialogShow.createDir" title="创建文件夹" width="500" tabindex="0" @keydown.enter.prevent="handleDialogEnter('createDir', $event)">
    <el-form @submit.prevent>
      <el-form-item :label-width="80" label="文件夹名称">
        <el-input v-model="dialogData.createDir.name" autocomplete="off"/>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="dialogShow.createDir = false">取消</pl-button>
        <pl-button type="primary" @click="createNewDir">保存</pl-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="dialogShow.createApi" title="创建接口" width="700" tabindex="0" @keydown.enter.prevent="handleDialogEnter('createApi', $event)">
    <el-tabs v-model="createApiType" tabindex="0">
      <el-tab-pane label="基础接口" name="params">
        <el-form>
          <el-form-item :label-width="80" label="接口名称">
            <el-input v-model="dialogData.createApi.name" autocomplete="off"/>
          </el-form-item>
          <el-form-item :label-width="80" label="接口地址">
            <el-input v-model="dialogData.createApi.url" autocomplete="off" placeholder="以斜杠开头"/>
          </el-form-item>
          <el-form-item :label-width="80" label="请求方式">
            <el-select
                v-model="dialogData.createApi.method"
                class="select"
                placeholder="筛选分组"
            >
              <el-option key="POST" label="POST" value="POST"/>
              <el-option key="GET" label="GET" value="GET"/>
            </el-select>
          </el-form-item>
          <el-form-item :label-width="80" label="请求协议">
            <el-select
                v-model="dialogData.createApi.protocol"
                class="select"
                placeholder="筛选分组"
            >
              <el-option key="http" label="http" value="http"/>
              <el-option key="https" label="https" value="https"/>
            </el-select>
          </el-form-item>
        </el-form>
      </el-tab-pane>
      <el-tab-pane label="从Curl导入" name="curl">
        <span>支持Chrome复制cmd bash,apifox生成shell代码</span><br/>
        <el-input v-model="dialogData.createApi.curlData" :rows="Number(15)" placehold type="textarea"></el-input>
      </el-tab-pane>
    </el-tabs>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="dialogShow.createApi = false">取消</pl-button>
        <pl-button type="primary" @click="handleFolderCreateApi">保存</pl-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="dialogShow.copyApi" title="复制接口" width="500" @keydown.enter.prevent="handleDialogEnter('copyApi', $event)">
    <el-form @submit.prevent>
      <el-form-item :label-width="80" label="接口名称">
        <el-input v-model="dialogData.copyApi.name" autocomplete="off" placeholder="请输入新的接口名称"/>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="dialogShow.copyApi = false">取消</pl-button>
        <pl-button type="primary" @click="copyApi">保存</pl-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="dialogShow.jsonImport" title="通过JSON导入" width="800" @keydown.enter.prevent="handleDialogEnter('jsonImport', $event)">
    <el-form :model="dialogData.jsonImport" label-width="120px" @submit.prevent>
      <el-form-item label="选择集合">
        <el-select v-model="dialogData.jsonImport.collection_id" placeholder="请选择集合" style="width: 100%;">
          <el-option
            v-for="collection in treeData"
            :key="collection.id"
            :label="collection.name"
            :value="collection.id"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="JSON数据">
        <el-input
          v-model="dialogData.jsonImport.json"
          type="textarea"
          :rows="15"
          placeholder='请输入JSON数据，例如：{"collection_id": 4, "items": [...]}'
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="dialogShow.jsonImport = false">取消</pl-button>
        <pl-button type="primary" @click="apiImportJson">导入</pl-button>
      </div>
    </template>
  </el-dialog>

  <el-drawer
      v-model="drawerVisibleMarkdown"
      direction="rtl"
      size="90%"
      title="文档"
  >
    <Markdown v-if="drawerVisibleMarkdown" :markdownType="markdownType"></Markdown>
  </el-drawer>

</template>

<script>
import {FolderOpened, Folder, Document, More, Plus, QuestionFilled, Tools} from '@element-plus/icons-vue'
import CollectionBasicInfo from './api/CollectionBasicInfo'
import CollectionEnvironment from './api/CollectionEnvironment'
import FolderDetail from './api/FolderDetail'
import ApiDetail from './api/ApiDetail'
import Markdown from '@/components/Markdown.vue'
import Api from '@/utils/base/api'
import ArrayUtil from '@/utils/base/array'
import KeyDebounceDetector from "@/utils/base/keyup"
import store from "@/utils/base/store";

export default {
  name: 'CollectionManager',
  components: {
    FolderOpened,
    Folder,
    Document,
    More,
    Plus,
    QuestionFilled,
    Tools,
    CollectionBasicInfo,
    CollectionEnvironment,
    FolderDetail,
    ApiDetail,
    Markdown
  },
  data() {
    return {
      // 树形数据
      treeData: [],
      treeProps: {
        children: 'children',
        label: 'name',
        isLeaf: 'isLeaf'
      },
      createApiType: 'params',
      defaultExpandedKeys: [],

      // 归档相关
      archiveExpanded: false,
      archivedItems: [],

      // 选中项（当前激活 tab 的数据镜像，用于兼容现有详情组件和旧逻辑）
      selectedItem: null,
      apiDetailLoading: false,

      // 工作区 tab 状态
      openTabs: [],
      activeTabKey: '',
      activeCollectionInnerTabMap: {},
      activeFolderInnerTabMap: {},

      // 文档drawer
      drawerVisibleMarkdown: false,
      markdownType: 'api',
      skillInstallZipUrl: 'https://gitee.com/Sxiaobai/skills/blob/master/dtool-api.zip',
      skillInstallPrompt: '请安装这个 skills zip： https://gitee.com/Sxiaobai/skills/blob/master/dtool-api.zip',

      // 环境相关
      currentEnvironment: 'dev',
      environmentList: [
        {label: '开发环境', value: 'dev'},
        {label: '测试环境', value: 'test'},
        {label: '生产环境', value: 'prod'}
      ],

      // 标签页
      collectionActiveTab: 'basic',
      dialogShow: {
        createCollection: false, //创建集合弹窗
        createDir: false, //创建文件夹弹窗
        createApi: false, //创建接口弹窗
        copyApi: false, //复制接口弹窗
        jsonImport: false, //JSON导入弹窗
      },
      dialogData: {
        createCollection: {
          uniqueid: '',
          name: '',
        },
        createDir: {
          uniqueid: '',
          name: '',
          parent_id: '',
          parent_type: '',
          headers: '{}',
        },
        createApi: {
          uniqueid: '',
          folder_id: '',
          collection_id: '',
          name: '',
          method: 'POST',
          url: '',
          protocol: 'http',
          desc: '',
          headers: {},
          query_params: "[]",
          content_type: "",
          body_form: "[]",
          body_json: {},
          curlData: '',
        },
        copyApi: {
          uniqueid: '',
          folder_id: '',
          collection_id: '',
          name: '',
          method: '',
          url: '',
          protocol: '',
          desc: '',
          headers: {},
          query_params: {},
          content_type: "",
          body_form: {},
          body_json: {},
        },
        jsonImport: {
          collection_id: '',
          json: '',
        }
      },
      // 弹窗保存防重入，避免回车和点击导致重复提交
      dialogSubmitting: {
        createCollection: false,
        createDir: false,
        createApi: false,
        copyApi: false,
        jsonImport: false,
      },
      keyup: null,
      // 左侧树面板宽度（支持拖拽与缓存）
      sidebarWidth: 280,
      // 是否正在拖拽调整左侧宽度
      sidebarResizing: false,
      // 拖拽起始坐标
      resizeStartX: 0,
      // 拖拽起始宽度
      resizeStartWidth: 280,
      // 拖拽事件处理器缓存（用于可靠解绑）
      sidebarResizeMoveHandler: null,
      sidebarResizeUpHandler: null,
    }
  },
  mounted() {
    this.sidebarResizeMoveHandler = this.handleSidebarResize.bind(this)
    this.sidebarResizeUpHandler = this.stopSidebarResize.bind(this)
    // 初始化左侧面板宽度缓存
    this.loadSidebarWidthCache()
    this.loadCollectionData()
    this.loadArchivedItems()
  },
  beforeUnmount() {
    // 组件销毁时兜底移除拖拽事件
    this.stopSidebarResize()
  },
  methods: {

    // ==================== 工作区 Tab 方法 ====================

    // 构建稳定的 tab key（collection:1 / folder:2 / api:3）
    buildWorkspaceTabKey(node) {
      if (!node || !node.type || !node.id) {
        return ''
      }
      return `${node.type}:${node.id}`
    },

    // 根据 tab key 获取已打开的 tab
    getWorkspaceTabByKey(tabKey) {
      if (!tabKey) {
        return null
      }
      return this.openTabs.find(tab => tab.key === tabKey) || null
    },

    // 获取当前激活的 tab
    getActiveWorkspaceTab() {
      return this.getWorkspaceTabByKey(this.activeTabKey)
    },
    getActiveCollectionInnerTab(collection) {
      if (!collection || !collection.id) {
        return 'basic'
      }
      return this.activeCollectionInnerTabMap[String(collection.id)] || 'basic'
    },
    handleCollectionInnerTabChange(tabName) {
      const tab = this.getActiveWorkspaceTab()
      if (!tab || tab.type !== 'collection') {
        return
      }
      this.activeCollectionInnerTabMap[String(tab.id)] = tabName || 'basic'
    },
    getActiveFolderInnerTab(folder) {
      if (!folder || !folder.id) {
        return 'basic'
      }
      return this.activeFolderInnerTabMap[String(folder.id)] || 'basic'
    },
    handleFolderInnerTabChange(tabName) {
      const tab = this.getActiveWorkspaceTab()
      if (!tab || tab.type !== 'folder') {
        return
      }
      this.activeFolderInnerTabMap[String(tab.id)] = tabName || 'basic'
    },

    // 创建新的工作区 tab
    createWorkspaceTab(node) {
      if (!node || !node.type) {
        return null
      }
      return {
        key: this.buildWorkspaceTabKey(node),
        type: node.type,
        id: node.id,
        uniqueid: node.uniqueid,
        title: node.name || '未命名',
        data: { ...node },
        loaded: false,
        loading: false
      }
    },

    // 打开或激活工作区 tab
    async openWorkspaceTab(node, options = {}) {
      const { reload = true } = options
      if (!node || !node.type) {
        return
      }
      const tabKey = this.buildWorkspaceTabKey(node)
      let tab = this.getWorkspaceTabByKey(tabKey)
      
      // 如果 tab 不存在，创建并添加
      if (!tab) {
        tab = this.createWorkspaceTab(node)
        if (!tab) {
          return
        }
        this.openTabs.push(tab)
      } else {
        tab.data = { ...tab.data, ...node }
        tab.title = node.name || tab.title
        tab.uniqueid = node.uniqueid || tab.uniqueid
      }
      
      // 激活 tab
      await this.activateWorkspaceTab(tabKey, { reload })
    },

    // 激活指定 tab
    async activateWorkspaceTab(tabKey, options = {}) {
      const { reload = false } = options
      const tab = this.getWorkspaceTabByKey(tabKey)
      if (!tab) {
        return
      }
      
      this.activeTabKey = tabKey
      
      // 同步 selectedItem
      this.syncSelectedItemFromActiveTab()
      
      // 同步左侧树高亮
      await this.highlightWorkspaceTreeNode(tab)
      
      // 如果需要重新加载或 tab 未加载
      if (reload || !tab.loaded) {
        await this.reloadWorkspaceTab(tabKey)
      }
    },

    // 从当前激活 tab 同步 selectedItem
    syncSelectedItemFromActiveTab() {
      const tab = this.getActiveWorkspaceTab()
      if (!tab) {
        this.selectedItem = null
        return
      }
      this.selectedItem = {
        ...tab.data,
        type: tab.type,
        id: tab.id,
        uniqueid: tab.uniqueid,
      }
    },

    // 关闭工作区 tab
    closeWorkspaceTab(tabKey) {
      const index = this.openTabs.findIndex(tab => tab.key === tabKey)
      if (index === -1) {
        return
      }
      
      // 移除 tab
      this.openTabs.splice(index, 1)
      
      // 如果关闭的是当前激活的 tab，切换到相邻 tab
      if (this.activeTabKey === tabKey) {
        if (this.openTabs.length > 0) {
          // 切换到相邻 tab（优先右侧，否则左侧）
          const newIndex = Math.min(index, this.openTabs.length - 1)
          const nextTab = this.openTabs[newIndex]
          this.activateWorkspaceTab(nextTab.key, { reload: false })
        } else {
          // 没有 tab 了，清空状态
          this.activeTabKey = ''
          this.selectedItem = null
        }
      }
    },

    // 关闭指定文件夹下的所有 tab
    closeWorkspaceTabsByFolder(folderId) {
      const tabsToClose = this.openTabs.filter(tab =>
        (tab.type === 'folder' && parseInt(tab.id) === parseInt(folderId)) ||
        (tab.type === 'api' && parseInt(tab.data.folder_id) === parseInt(folderId))
      )
      tabsToClose.forEach(tab => this.closeWorkspaceTab(tab.key))
    },

    // 关闭指定集合下的所有 tab
    closeWorkspaceTabsByCollection(collectionId) {
      const tabsToClose = this.openTabs.filter(tab =>
        (tab.type === 'collection' && parseInt(tab.id) === parseInt(collectionId)) ||
        (tab.type === 'folder' && parseInt(tab.data.collection_id) === parseInt(collectionId)) ||
        (tab.type === 'api' && parseInt(tab.data.collection_id) === parseInt(collectionId))
      )
      tabsToClose.forEach(tab => this.closeWorkspaceTab(tab.key))
    },

    // 处理 tab 切换
    async handleWorkspaceTabChange(tabKey) {
      const tab = this.getWorkspaceTabByKey(tabKey)
      if (!tab) {
        return
      }
      
      // 同步 selectedItem
      this.syncSelectedItemFromActiveTab()
      
      // 同步左侧树高亮
      await this.highlightWorkspaceTreeNode(tab)
    },

    // 高亮左侧树节点
    async ensureNodeVisibleInTree(tab) {
      if (!tab) {
        return null
      }
      if (tab.type === 'collection') {
        return this.findCollectionNode(tab.id)
      }
      if (tab.type === 'folder') {
        const collectionNode = this.findCollectionNode(tab.data.collection_id)
        if (!collectionNode) {
          return null
        }
        await this.ensureCollectionFoldersLoaded(collectionNode)
        return this.findFolderNode(tab.data.collection_id, tab.id)
      }
      if (tab.type === 'api') {
        const collectionNode = this.findCollectionNode(tab.data.collection_id)
        if (!collectionNode) {
          return null
        }
        await this.ensureCollectionFoldersLoaded(collectionNode)
        const folderNode = this.findFolderNode(tab.data.collection_id, tab.data.folder_id)
        if (!folderNode) {
          return null
        }
        await this.ensureFolderApisLoaded(folderNode)
        return this.findApiNode(tab.data.collection_id, tab.data.folder_id, tab.id)
      }
      return null
    },
    async highlightWorkspaceTreeNode(tab) {
      if (!tab || !this.$refs.collectionTreeRef) {
        return
      }
      const visibleNode = await this.ensureNodeVisibleInTree(tab)
      const targetKey = visibleNode && visibleNode.uniqueid ? visibleNode.uniqueid : tab.uniqueid
      if (targetKey) {
        this.$refs.collectionTreeRef.setCurrentKey(targetKey)
      }
    },

    // 重新加载工作区 tab
    async reloadWorkspaceTab(tabKey) {
      const tab = this.getWorkspaceTabByKey(tabKey)
      if (!tab) {
        return
      }
      
      if (tab.type === 'collection') {
        await this.reloadCollectionTab(tab)
      } else if (tab.type === 'folder') {
        await this.reloadFolderTab(tab)
      } else if (tab.type === 'api') {
        await this.reloadApiTab(tab)
      }
    },

    // 重新加载集合 tab
    async reloadCollectionTab(tab) {
      if (!tab || tab.loading) {
        return
      }
      tab.loading = true
      try {
        const data = await this.requestApi('CollectionListBasic', {})
        const collection = (data.list || []).find((item) => parseInt(item.id) === parseInt(tab.id))
        if (collection) {
          const collectionNode = this.findCollectionNode(tab.id)
          const normalizedCollection = this.normalizeCollectionNode(collection)
          if (collectionNode) {
            normalizedCollection.children = Array.isArray(collectionNode.children) ? collectionNode.children : []
            normalizedCollection.loaded = collectionNode.loaded
            normalizedCollection.loading = collectionNode.loading
            Object.assign(collectionNode, normalizedCollection)
            tab.data = { ...collectionNode }
            tab.title = collectionNode.name || tab.title
          } else {
            tab.data = normalizedCollection
            tab.title = normalizedCollection.name || tab.title
          }
        }
        tab.loaded = true
        this.syncSelectedItemFromActiveTab()
      } finally {
        tab.loading = false
      }
    },

    // 重新加载文件夹 tab
    async reloadFolderTab(tab) {
      if (!tab || tab.loading) {
        return
      }
      tab.loading = true
      try {
        const detailData = await this.requestApi('FolderDetail', {
          dir_id: tab.id,
        })
        const folderDetail = detailData.dir || null
        const folderNode = this.findFolderNode(tab.data.collection_id, tab.id)
        if (folderNode) {
          if (folderDetail) {
            this.syncFolderNodeFields(folderNode, folderDetail)
          }
          await this.loadFolderApis(folderNode, true)
          folderNode.child_count = Array.isArray(folderNode.children) ? folderNode.children.length : 0
          folderNode.isLeaf = folderNode.child_count <= 0
          tab.data = { ...folderNode }
          tab.title = folderNode.name || tab.title
        } else if (folderDetail) {
          const normalizedFolder = this.normalizeFolderNode(folderDetail, folderDetail.collection_id)
          normalizedFolder.children = Array.isArray(folderDetail.children)
            ? folderDetail.children.map((api) => this.normalizeApiNode(api, normalizedFolder.id, normalizedFolder.collection_id))
            : []
          normalizedFolder.loaded = true
          normalizedFolder.child_count = normalizedFolder.children.length
          normalizedFolder.isLeaf = normalizedFolder.child_count <= 0
          tab.data = normalizedFolder
          tab.title = normalizedFolder.name || tab.title
        }
        this.syncSelectedItemFromActiveTab()
        tab.loaded = true
      } finally {
        tab.loading = false
      }
    },

    // 重新加载接口 tab
    async reloadApiTab(tab) {
      if (!tab || tab.loading) {
        return
      }
      tab.loading = true
      this.apiDetailLoading = true
      try {
        const data = await this.requestApi('ApisDetailByIds', {
          ids: [tab.id],
        })
        const detail = Array.isArray(data.list) ? data.list[0] : null
        if (detail) {
          tab.data = {
            ...detail,
            type: 'api',
            id: tab.id,
            uniqueid: tab.uniqueid,
          }
          tab.title = detail.name || tab.title
          this.syncSelectedItemFromActiveTab()
          const treeApiNode = this.findApiNode(detail.collection_id, detail.folder_id, detail.id)
          if (treeApiNode) {
            this.syncApiNodeFields(treeApiNode, detail)
          }
        }
        tab.loaded = true
      } catch (error) {
        this.$message.error(error.message || '加载接口详情失败')
      } finally {
        tab.loading = false
        this.apiDetailLoading = false
      }
      await this.$nextTick()
      if (this.$refs.refApiDetail && this.activeTabKey === tab.key && tab.data) {
        this.$refs.refApiDetail.InitApiDetail(tab.data)
      }
    },

    // 更新或插入 tab 数据
    upsertWorkspaceTabData(nodeLike) {
      if (!nodeLike || !nodeLike.type || !nodeLike.id) {
        return
      }
      const tabKey = this.buildWorkspaceTabKey(nodeLike)
        const tab = this.getWorkspaceTabByKey(tabKey)
      
      if (tab) {
        // 更新已有 tab
        tab.data = { ...tab.data, ...nodeLike }
        tab.title = nodeLike.name || tab.title
        tab.uniqueid = nodeLike.uniqueid || tab.uniqueid
        tab.loaded = true
        // 如果是当前激活的 tab，同步 selectedItem
        if (this.activeTabKey === tabKey) {
          this.syncSelectedItemFromActiveTab()
        }
      }
    },

    // ==================== 树节点方法 ====================

    // 读取并应用左侧面板宽度缓存
    loadSidebarWidthCache() {
      let _that = this
      const cacheWidth = parseInt(store.getStore('api_sidebar_width') || '', 10)
      if (!Number.isNaN(cacheWidth)) {
        _that.sidebarWidth = _that.clampSidebarWidth(cacheWidth)
      }
    },

    // 限制左侧面板宽度范围，避免过窄或过宽影响布局
    clampSidebarWidth(width) {
      const minWidth = 240
      const maxWidth = 520
      return Math.min(maxWidth, Math.max(minWidth, width))
    },

    // 开始拖拽调整左右分栏
    startSidebarResize(event) {
      let _that = this
      if (window.innerWidth <= 1200) {
        return
      }
      _that.sidebarResizing = true
      _that.resizeStartX = event.clientX
      _that.resizeStartWidth = _that.sidebarWidth
      document.addEventListener('mousemove', _that.sidebarResizeMoveHandler)
      document.addEventListener('mouseup', _that.sidebarResizeUpHandler)
    },

    // 拖拽过程中实时更新左侧宽度
    handleSidebarResize(event) {
      let _that = this
      if (!_that.sidebarResizing) {
        return
      }
      const deltaX = event.clientX - _that.resizeStartX
      _that.sidebarWidth = _that.clampSidebarWidth(_that.resizeStartWidth + deltaX)
    },

    // 结束拖拽并写入缓存
    stopSidebarResize() {
      let _that = this
      if (_that.sidebarResizing) {
        store.setStore('api_sidebar_width', _that.sidebarWidth)
      }
      _that.sidebarResizing = false
      document.removeEventListener('mousemove', _that.sidebarResizeMoveHandler)
      document.removeEventListener('mouseup', _that.sidebarResizeUpHandler)
    },


    normalizeCollectionNode(collection) {
      return {
        ...collection,
        type: 'collection',
        children: [],
        isLeaf: this.resolveTreeNodeLeafState('collection', collection.child_count),
        loaded: false,
        loading: false,
      }
    },
    resolveTreeNodeLeafState(type, childCount) {
      if (type === 'api') {
        return true
      }
      const count = Number(childCount)
      if (!Number.isNaN(count)) {
        return count <= 0
      }
      return false
    },
    getNodeChildCount(node) {
      if (!node) {
        return 0
      }
      if (typeof node.child_count === 'number') {
        return node.child_count
      }
      if (Array.isArray(node.children)) {
        return node.children.length
      }
      return 0
    },
    normalizeFolderNode(folder, collectionId) {
      return {
        ...folder,
        collection_id: folder.collection_id || collectionId,
        headers: folder.headers || '{}',
        type: 'folder',
        children: [],
        isLeaf: this.resolveTreeNodeLeafState('folder', folder.child_count),
        loaded: false,
        loading: false,
      }
    },
    normalizeApiNode(api, folderId, collectionId) {
      return {
        ...api,
        folder_id: api.folder_id || folderId,
        collection_id: api.collection_id || collectionId,
        type: 'api',
        children: [],
        isLeaf: true,
        loaded: true,
        loading: false,
      }
    },
    requestApi(methodName, params) {
      return new Promise((resolve, reject) => {
        Api[methodName](params || {}, function (res) {
          if (res.ErrCode === 0) {
            resolve(res.Data || {})
            return
          }
          reject(new Error(res.ErrMsg || '请求失败'))
        })
      })
    },
    findCollectionNode(collectionId) {
      return this.treeData.find((collection) => parseInt(collection.id) === parseInt(collectionId)) || null
    },
    findFolderNode(collectionId, folderId) {
      const collection = this.findCollectionNode(collectionId)
      if (!collection || !Array.isArray(collection.children)) {
        return null
      }
      return collection.children.find((folder) => parseInt(folder.id) === parseInt(folderId)) || null
    },
    findApiNode(collectionId, folderId, apiId) {
      const folder = this.findFolderNode(collectionId, folderId)
      if (!folder || !Array.isArray(folder.children)) {
        return null
      }
      return folder.children.find((api) => parseInt(api.id) === parseInt(apiId)) || null
    },
    syncTreeNodeChildren(nodeKey, children) {
      if (!this.$refs.collectionTreeRef || !nodeKey) {
        return
      }
      this.$refs.collectionTreeRef.updateKeyChildren(nodeKey, Array.isArray(children) ? children : [])
    },
    syncCollectionNodeFields(target, source) {
      if (!target || !source) {
        return
      }
      Object.assign(target, {
        name: source.name,
        create_time: source.create_time,
        update_time: source.update_time,
      })
    },
    syncFolderNodeFields(target, source) {
      if (!target || !source) {
        return
      }
      Object.assign(target, {
        name: source.name,
        desc: source.desc,
        headers: source.headers || '{}',
        collection_id: source.collection_id,
        create_time: source.create_time,
        update_time: source.update_time,
      })
    },
    syncApiNodeFields(target, source) {
      if (!target || !source) {
        return
      }
      Object.assign(target, {
        name: source.name,
        method: source.method,
        url: source.url,
        desc: source.desc,
        env_id: source.env_id,
        weight: source.weight,
        update_time: source.update_time,
      })
    },
    async loadCollectionFolders(collectionNode, force = false) {
      if (!collectionNode) {
        return []
      }
      if (collectionNode.loading) {
        return collectionNode.children || []
      }
      if (!force && collectionNode.loaded) {
        return collectionNode.children || []
      }
      collectionNode.loading = true
      try {
        const data = await this.requestApi('CollectionFoldersBasic', {
          collection_id: collectionNode.id,
        })
        collectionNode.children = (data.list || []).map((folder) => this.normalizeFolderNode(folder, collectionNode.id))
        collectionNode.child_count = collectionNode.children.length
        collectionNode.isLeaf = collectionNode.child_count <= 0
        collectionNode.loaded = true
        this.sortListByIdOrder(collectionNode.children, this.getTreeSortCache().folders[String(collectionNode.id)] || [])
        this.syncTreeNodeChildren(collectionNode.uniqueid, collectionNode.children)
        return collectionNode.children
      } finally {
        collectionNode.loading = false
      }
    },
    async loadFolderApis(folderNode, force = false) {
      if (!folderNode) {
        return []
      }
      if (folderNode.loading) {
        return folderNode.children || []
      }
      if (!force && folderNode.loaded) {
        return folderNode.children || []
      }
      folderNode.loading = true
      try {
        const data = await this.requestApi('FolderApisBasic', {
          folder_id: folderNode.id,
        })
        folderNode.children = (data.list || []).map((api) => this.normalizeApiNode(api, folderNode.id, folderNode.collection_id))
        folderNode.child_count = folderNode.children.length
        folderNode.isLeaf = folderNode.child_count <= 0
        folderNode.loaded = true
        this.applyFolderApiSort(folderNode.collection_id, folderNode.id)
        this.syncTreeNodeChildren(folderNode.uniqueid, folderNode.children)
        return folderNode.children
      } finally {
        folderNode.loading = false
      }
    },
    async ensureCollectionFoldersLoaded(collectionNode) {
      if (!collectionNode) {
        return []
      }
      try {
        return await this.loadCollectionFolders(collectionNode, false)
      } catch (error) {
        this.$message.error(error.message || '加载集合文件夹失败')
        return []
      }
    },
    async ensureFolderApisLoaded(folderNode) {
      if (!folderNode) {
        return []
      }
      try {
        return await this.loadFolderApis(folderNode, false)
      } catch (error) {
        this.$message.error(error.message || '加载文件夹接口失败')
        return []
      }
    },
    async refreshCollectionFolders(collectionId) {
      const collectionNode = this.findCollectionNode(collectionId)
      if (!collectionNode) {
        return []
      }
      try {
        const folders = await this.loadCollectionFolders(collectionNode, true)
        this.upsertWorkspaceTabData(collectionNode)
        return folders
      } catch (error) {
        this.$message.error(error.message || '刷新集合文件夹失败')
        return []
      }
    },
    async refreshFolderApis(collectionId, folderId) {
      const folderNode = this.findFolderNode(collectionId, folderId)
      if (!folderNode) {
        return []
      }
      try {
        const apis = await this.loadFolderApis(folderNode, true)
        this.upsertWorkspaceTabData(folderNode)
        return apis
      } catch (error) {
        this.$message.error(error.message || '刷新文件夹接口失败')
        return []
      }
    },
    async loadApiDetail(apiNode) {
      if (!apiNode || !apiNode.id) {
        return null
      }
      let loadedDetail = null
      this.apiDetailLoading = true
      try {
        const data = await this.requestApi('ApisDetailByIds', {
          ids: [apiNode.id],
        })
        const detail = Array.isArray(data.list) ? data.list[0] : null
        if (!detail) {
          throw new Error('未获取到接口详情')
        }
        const treeApiNode = this.findApiNode(apiNode.collection_id, apiNode.folder_id, apiNode.id)
        if (treeApiNode) {
          this.syncApiNodeFields(treeApiNode, detail)
        }
        const activeTab = this.getActiveWorkspaceTab()
        if (activeTab && activeTab.type === 'api' && parseInt(activeTab.id) === parseInt(detail.id)) {
          activeTab.data = {
            ...detail,
            type: 'api',
            id: activeTab.id,
            uniqueid: activeTab.uniqueid,
          }
          activeTab.title = detail.name || activeTab.title
          activeTab.loaded = true
        }
        this.selectedItem = {
          ...detail,
          type: 'api',
          id: apiNode.id,
          uniqueid: apiNode.uniqueid,
        }
        loadedDetail = this.selectedItem
      } catch (error) {
        this.selectedItem = null
        this.$message.error(error.message || '加载接口详情失败')
        return null
      } finally {
        this.apiDetailLoading = false
      }
      await this.$nextTick()
      if (this.$refs.refApiDetail && loadedDetail && this.selectedItem && this.selectedItem.type === 'api') {
        this.$refs.refApiDetail.InitApiDetail(loadedDetail)
      }
      return loadedDetail
    },
    // 加载集合数据
    loadCollectionData() {
      let _that = this
      Api.CollectionListBasic({}, function (res) {
        if (res.ErrCode === 0) {
          _that.treeData = (res.Data.list || []).map((collection) => _that.normalizeCollectionNode(collection))
          // 加载集合树后按本地缓存恢复排序
          _that.applyTreeSortCache()
          _that.initTreeExpansion()
        } else {
          _that.$message.error(res.ErrMsg)
        }
      })
    },

    // 初始化树展开状态
    initTreeExpansion() {
      let _that = this
      // 仅恢复集合/文件夹展开状态，使用稳定的 type:id 作为缓存键
      const expandedStateCache = _that.getExpandedStateCache()
      if (!expandedStateCache.initialized) {
        _that.setExpandedStateCache([])
        return
      }
      _that.$nextTick(() => {
        _that.restoreExpandedNodes(expandedStateCache.expandedKeys)
      })
    },
    async loadTreeNode(node, resolve) {
      if (!node || node.level === 0) {
        resolve(this.treeData)
        return
      }
      const data = node.data
      if (!data || !data.type) {
        resolve([])
        return
      }
      if (data.type === 'collection') {
        resolve(await this.ensureCollectionFoldersLoaded(data))
        return
      }
      if (data.type === 'folder') {
        resolve(await this.ensureFolderApisLoaded(data))
        return
      }
      resolve([])
    },

    // 处理节点展开
    handleNodeExpand(data) {
      let _that = this
      _that.updateExpandedCache(data, true)
    },

    // 处理节点折叠
    handleNodeCollapse(data) {
      let _that = this
      _that.updateExpandedCache(data, false)
    },

    // 获取树展开缓存（仅 collection/folder）
    getExpandedStateCache() {
      const cacheText = store.getStore('collection_tree_expand_state')
      if (!cacheText) {
        return {
          initialized: false,
          expandedKeys: []
        }
      }
      try {
        const cacheData = JSON.parse(cacheText)
        // 兼容旧版本：历史上直接存的是数组
        if (Array.isArray(cacheData)) {
          return {
            initialized: true,
            expandedKeys: cacheData
          }
        }
        if (cacheData && typeof cacheData === 'object') {
          return {
            initialized: cacheData.initialized === true,
            expandedKeys: Array.isArray(cacheData.expandedKeys) ? cacheData.expandedKeys : []
          }
        }
        return {
          initialized: false,
          expandedKeys: []
        }
      } catch (e) {
        return {
          initialized: false,
          expandedKeys: []
        }
      }
    },
    // 写入树展开缓存
    setExpandedStateCache(expandedState) {
      store.setStore('collection_tree_expand_state', JSON.stringify({
        initialized: true,
        expandedKeys: Array.isArray(expandedState) ? expandedState : []
      }))
    },
    // 构建稳定的缓存键（collection:1 / folder:2）
    buildExpandStateKey(data) {
      if (!data || !data.type) {
        return ''
      }
      if (data.type !== 'collection' && data.type !== 'folder') {
        return ''
      }
      return `${data.type}:${data.id}`
    },
    // 从树中同步当前展开状态到缓存
    syncExpandedStateCacheFromTree() {
      let _that = this
      if (!_that.$refs.collectionTreeRef || !_that.$refs.collectionTreeRef.store) {
        return
      }
      const expandedState = []
      _that.$refs.collectionTreeRef.store._getAllNodes().forEach(node => {
        if (!node || !node.expanded || !node.data) {
          return
        }
        const key = _that.buildExpandStateKey(node.data)
        if (key && !expandedState.includes(key)) {
          expandedState.push(key)
        }
      })
      _that.setExpandedStateCache(expandedState)
    },
    // 按缓存恢复展开状态
    async restoreExpandedNodes(expandedState) {
      let _that = this
      if (!_that.$refs.collectionTreeRef) {
        return
      }
      const expandedKeys = Array.isArray(expandedState) ? expandedState : []
      for (const collection of _that.treeData) {
        const collectionKey = _that.buildExpandStateKey(collection)
        if (collectionKey && expandedKeys.includes(collectionKey)) {
          await _that.ensureCollectionFoldersLoaded(collection)
          const collectionNode = _that.$refs.collectionTreeRef.getNode(collection.uniqueid)
          if (collectionNode) {
            collectionNode.expand()
          }
        }
      }
      for (const collection of _that.treeData) {
        if (!Array.isArray(collection.children)) {
          continue
        }
        for (const folder of collection.children) {
          const folderKey = _that.buildExpandStateKey(folder)
          if (folderKey && expandedKeys.includes(folderKey)) {
            await _that.ensureFolderApisLoaded(folder)
            const folderNode = _that.$refs.collectionTreeRef.getNode(folder.uniqueid)
            if (folderNode) {
              folderNode.expand()
            }
          }
        }
      }
    },
    // 更新展开缓存
    updateExpandedCache(data, isExpanded) {
      let _that = this
      const nodeKey = _that.buildExpandStateKey(data)
      if (!nodeKey) {
        return
      }
      let expandedState = _that.getExpandedStateCache()
      expandedState = Array.isArray(expandedState.expandedKeys) ? expandedState.expandedKeys : []
      if (isExpanded) {
        if (!expandedState.includes(nodeKey)) {
          expandedState.push(nodeKey)
        }
      } else {
        expandedState = expandedState.filter(key => key !== nodeKey)
        // 收起集合时，清理其下文件夹展开缓存，确保刷新后状态一致
        if (data.type === 'collection' && Array.isArray(data.children)) {
          data.children.forEach(folder => {
            const folderKey = _that.buildExpandStateKey(folder)
            expandedState = expandedState.filter(key => key !== folderKey)
          })
        }
      }
      _that.setExpandedStateCache(expandedState)
    },

    // 加载归档项
    loadArchivedItems() {
      this.archivedItems = [
        {id: 'arc1', name: '旧版用户接口', type: 'api'},
        {id: 'arc2', name: '测试集合', type: 'collection'}
      ]
    },

    // 切换集合展开/收起
    toggleCollection(collection) {
      collection.collapsed = !collection.collapsed
      if (collection.collapsed) {
        // 收起时移除展开的key
        this.defaultExpandedKeys = this.defaultExpandedKeys.filter(key => key !== collection.id)
      } else {
        // 展开时添加key
        this.defaultExpandedKeys.push(collection.id)
      }
      this.$refs.collectionTreeRef.updateKeyChildren(collection.uniqueid, collection.children)
    },

    // 切换归档列表
    toggleArchive() {
      this.archiveExpanded = !this.archiveExpanded
    },
    // 允许拖拽的节点类型（集合/文件夹/接口）
    allowTreeNodeDrag(draggingNode) {
      const dragData = draggingNode && draggingNode.data ? draggingNode.data : {}
      return dragData.type === 'collection' || dragData.type === 'folder' || dragData.type === 'api'
    },
    // 控制可放置位置：仅支持同类型、同层级排序，不支持 inner
    allowTreeNodeDrop(draggingNode, dropNode, dropType) {
      const dragData = draggingNode && draggingNode.data ? draggingNode.data : {}
      const dropData = dropNode && dropNode.data ? dropNode.data : {}
      if (dropType === 'inner') {
        return false
      }
      if (dragData.type !== dropData.type) {
        return false
      }
      if (dragData.type === 'collection') {
        return true
      }
      if (dragData.type === 'folder') {
        return parseInt(dragData.collection_id) === parseInt(dropData.collection_id)
      }
      if (dragData.type === 'api') {
        return parseInt(dragData.folder_id) === parseInt(dropData.folder_id)
      }
      return false
    },
    // 节点拖拽后同步保存排序缓存
    handleTreeNodeDrop() {
      let _that = this
      _that.syncTreeSortCacheFromTree()
      _that.$message.success('排序已保存')
    },
    // 获取树排序缓存
    getTreeSortCache() {
      const cacheText = store.getStore('collection_tree_sort_state')
      if (!cacheText) {
        return {
          collections: [],
          folders: {},
          apis: {},
        }
      }
      try {
        const cacheData = JSON.parse(cacheText)
        return {
          collections: Array.isArray(cacheData.collections) ? cacheData.collections : [],
          folders: cacheData.folders && typeof cacheData.folders === 'object' ? cacheData.folders : {},
          apis: cacheData.apis && typeof cacheData.apis === 'object' ? cacheData.apis : {},
        }
      } catch (e) {
        return {
          collections: [],
          folders: {},
          apis: {},
        }
      }
    },
    // 写入树排序缓存
    setTreeSortCache(cacheData) {
      store.setStore('collection_tree_sort_state', JSON.stringify(cacheData))
    },
    // 按 id 顺序缓存对列表排序（未命中缓存项会排在后面）
    sortListByIdOrder(list, idOrder) {
      if (!Array.isArray(list) || !Array.isArray(idOrder) || idOrder.length === 0) {
        return
      }
      const orderMap = {}
      idOrder.forEach((id, index) => {
        orderMap[String(id)] = index
      })
      list.sort((a, b) => {
        const aKey = String(a && a.id)
        const bKey = String(b && b.id)
        const aOrder = orderMap[aKey]
        const bOrder = orderMap[bKey]
        const aMiss = aOrder === undefined
        const bMiss = bOrder === undefined
        if (aMiss && bMiss) {
          return parseInt(a.id) - parseInt(b.id)
        }
        if (aMiss) {
          return 1
        }
        if (bMiss) {
          return -1
        }
        return aOrder - bOrder
      })
    },
    // 应用本地缓存排序到集合树
    applyTreeSortCache() {
      let _that = this
      const sortCache = _that.getTreeSortCache()
      _that.sortListByIdOrder(_that.treeData, sortCache.collections)
      _that.treeData.forEach(collection => {
        if (!Array.isArray(collection.children)) {
          return
        }
        _that.sortListByIdOrder(collection.children, sortCache.folders[String(collection.id)] || [])
        collection.children.forEach(folder => {
          if (!Array.isArray(folder.children)) {
            return
          }
          _that.sortListByIdOrder(folder.children, sortCache.apis[String(folder.id)] || [])
        })
      })
    },
    // 将当前树结构同步到本地排序缓存
    syncTreeSortCacheFromTree() {
      let _that = this
      const sortCache = {
        collections: [],
        folders: {},
        apis: {},
      }
      _that.treeData.forEach(collection => {
        sortCache.collections.push(collection.id)
        const folderList = Array.isArray(collection.children) ? collection.children : []
        sortCache.folders[String(collection.id)] = folderList.map(folder => folder.id)
        folderList.forEach(folder => {
          const apiList = Array.isArray(folder.children) ? folder.children : []
          sortCache.apis[String(folder.id)] = apiList.map(api => api.id)
        })
      })
      _that.setTreeSortCache(sortCache)
    },
    // 对单个文件夹接口列表应用排序缓存
    applyFolderApiSort(collectionId, folderId) {
      let _that = this
      const sortCache = _that.getTreeSortCache()
      for (let i in _that.treeData) {
        if (parseInt(collectionId) !== parseInt(_that.treeData[i].id)) {
          continue
        }
        for (let j in _that.treeData[i].children) {
          if (parseInt(folderId) !== parseInt(_that.treeData[i].children[j].id)) {
            continue
          }
          const apiList = _that.treeData[i].children[j].children
          _that.sortListByIdOrder(apiList, sortCache.apis[String(folderId)] || [])
        }
      }
    },
    handleKeyUp: function (event) {
      let _that = this
      _that.initKeyUp()
      _that.keyup.keyUp(event.key)
      event.preventDefault()
    },
    initKeyUp: function () {
      let _that = this
      if (_that.keyup !== null) {
        return
      }
      _that.keyup = new KeyDebounceDetector(function (key1, key2) {
        if ((key1 === 'Control' && key2 === 's') || (key1 === 's' && key2 === 'Control')) {
          return
        } else if ((key1 === 'Control' && key2 === 'Enter') || (key1 === 'Enter' && key2 === 'Control')) {
          if (_that.selectedItem && _that.selectedItem.type === 'api') {
            _that.$nextTick(() => {
              _that.$refs.refApiDetail.handleExecute();
            });
          }

        }
      }, 500)
    },
    // 处理节点点击：单击打开或激活工作区 tab
    async handleNodeClick(data) {
      let _that = this
      if (!data || !data.type) {
        return
      }
      
      // 打开或激活工作区 tab，重复点击时会重新加载数据
      await _that.openWorkspaceTab(data, { reload: true })
    },
    // 处理节点双击：集合和文件夹双击时切换展开/收起
    handleNodeDoubleClick(data) {
      if (!data || (data.type !== 'collection' && data.type !== 'folder')) {
        return
      }
      const treeRef = this.$refs.collectionTreeRef
      if (!treeRef) {
        return
      }
      const node = treeRef.getNode(data.uniqueid)
      if (!node) {
        return
      }
      if (node.expanded) {
        node.collapse()
        // 双击收起时主动同步本地展开缓存，避免刷新后状态残留
        this.updateExpandedCache(data, false)
        return
      }
      if (data.type === 'collection') {
        this.ensureCollectionFoldersLoaded(data)
      } else if (data.type === 'folder') {
        this.ensureFolderApisLoaded(data)
      }
      node.expand()
      // 双击展开时主动同步本地展开缓存，避免依赖树事件遗漏
      this.updateExpandedCache(data, true)
    },
    // 处理归档项点击
    handleArchiveItemClick(item) {
      this.openWorkspaceTab(item, { reload: false })
    },
    // 处理弹窗内回车提交（统一入口，避免重复触发）
    handleDialogEnter(dialogType, event) {
      if (!event || event.isComposing) {
        return
      }
      const targetTag = ((event.target && event.target.tagName) || '').toUpperCase()
      // 文本域内回车默认作为换行，不触发提交
      if (targetTag === 'TEXTAREA') {
        return
      }
      event.preventDefault()
      event.stopPropagation()
      this.submitDialog(dialogType)
    },
    // 根据弹窗类型提交保存逻辑
    submitDialog(dialogType) {
      const submitMap = {
        createCollection: () => this.createNewCollection(),
        createDir: () => this.createNewDir(),
        createApi: () => this.handleFolderCreateApi(),
        copyApi: () => this.copyApi(),
        jsonImport: () => this.apiImportJson(),
      }
      if (submitMap[dialogType]) {
        submitMap[dialogType]()
      }
    },
    // 开始弹窗提交，返回 true 表示已有提交在进行中
    beginDialogSubmit(dialogType) {
      if (this.dialogSubmitting[dialogType]) {
        return true
      }
      this.dialogSubmitting[dialogType] = true
      return false
    },
    // 结束弹窗提交，释放提交锁
    endDialogSubmit(dialogType) {
      if (this.dialogSubmitting[dialogType] !== undefined) {
        this.dialogSubmitting[dialogType] = false
      }
    },
    // 列表去重追加，防止重复节点渲染
    pushUniqueByKey(list, item, key) {
      if (!Array.isArray(list) || !item) {
        return
      }
      const uniqueKey = key || 'uniqueid'
      const exists = list.some((node) => String(node && node[uniqueKey]) === String(item[uniqueKey]))
      if (!exists) {
        list.push(item)
      }
    },
    openCreateCollectionDialog() {
      this.dialogData.createCollection = { uniqueid: '', name: '' }
      this.dialogShow.createCollection = true
    },
    copyText(text, successMsg) {
      let _that = this
      const val = (text || '').trim()
      if (!val) {
        _that.$message.error('无可复制内容')
        return
      }
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(val).then(() => {
          _that.$message.success(successMsg)
        }).catch(() => {
          _that.$message.error('复制失败，请手动复制')
        })
        return
      }
      const input = document.createElement('textarea')
      input.value = val
      document.body.appendChild(input)
      input.select()
      try {
        document.execCommand('copy')
        _that.$message.success(successMsg)
      } catch (e) {
        _that.$message.error('复制失败，请手动复制')
      } finally {
        document.body.removeChild(input)
      }
    },
    // 创建新集合
    createNewCollection() {
      let _that = this
      if (!_that.dialogShow.createCollection) {
        _that.dialogShow.createCollection = true
        return
      }
      if (_that.beginDialogSubmit('createCollection')) {
        return
      }
      Api.CreateCollection(_that.dialogData.createCollection, function (res) {
        _that.endDialogSubmit('createCollection')
        if (res.ErrCode === 0) {
          _that.dialogShow.createCollection = false
          let newCollection = _that.normalizeCollectionNode(res.Data)
          _that.pushUniqueByKey(_that.treeData, newCollection, 'uniqueid')
          _that.syncTreeSortCacheFromTree()
          // 新建集合成功后，自动打开该集合 tab
          _that.$nextTick(() => {
            _that.openWorkspaceTab(newCollection, { reload: true })
          })
        } else {
          _that.$message.error(res.ErrMsg)
        }
      })
    },
    // 创建新文件夹
    createNewDir(data, data2) {
      console.log('数据', data, data2)
      let _that = this
      if (!_that.dialogShow.createDir) {
        if (data !== undefined && data !== null) {
          _that.openWorkspaceTab(data, { reload: false })
        }
        _that.dialogShow.createDir = true
        _that.dialogData.createDir = {}
        return
      }
      console.log('选择的', _that.selectedItem)
      if (_that.beginDialogSubmit('createDir')) {
        return
      }
      _that.dialogData.createDir.collection_id = _that.selectedItem.id
      Api.CreateDir(_that.dialogData.createDir, function (res) {
        _that.endDialogSubmit('createDir')
        if (res.ErrCode === 0) {
          _that.dialogShow.createDir = false
          _that.refreshCollectionFolders(_that.dialogData.createDir.collection_id).then(async () => {
            _that.syncTreeSortCacheFromTree()
            // 新建文件夹成功后，自动打开该文件夹 tab
            const newFolder = _that.normalizeFolderNode(res.Data, _that.dialogData.createDir.collection_id)
            await _that.openWorkspaceTab(newFolder, { reload: true })
          })
        } else {
          _that.$message.error(res.ErrMsg)
        }
      })
    },
    // 运行全部
    executeAll() {
      // 实现运行全部接口的逻辑
    },

    // 导出集合
    exportCollection() {
      // 实现导出集合的逻辑
    },

    // 执行单个接口
    executeApi(api) {
      // 实现执行单个接口的逻辑
    },

    // 处理集合更新
    handleCollectionUpdate(collection) {
      let _that = this
      Api.CreateCollection(collection, function (res) {
        if (res.ErrCode === 0) {
          for (let i in _that.treeData) {
            if (parseInt(collection.id) === parseInt(_that.treeData[i].id)) {
              _that.syncCollectionNodeFields(_that.treeData[i], res.Data || collection)
              // 更新工作区 tab 数据
              _that.upsertWorkspaceTabData(_that.treeData[i])
            }
          }
        } else {
          _that.$message.error(res.ErrMsg)
        }
      })
    },
    //删除集合
    handleCollectionDelete(collection) {
      let _that = this
      this.$confirm('确定要删除这个集合吗？此操作不可恢复。', '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        Api.DeleteCollection(collection, function (res) {
          if (res.ErrCode === 0) {
            _that.treeData = ArrayUtil.DeleteValueByStringKey(_that.treeData, 'uniqueid', collection.uniqueid)
            // 关闭该集合相关的所有 tab
            _that.closeWorkspaceTabsByCollection(collection.id)
            _that.syncTreeSortCacheFromTree()
            _that.$message.success('删除成功')
          } else {
            _that.$message.error(res.ErrMsg)
          }
        })
      }).catch(() => {
        _that.$message.info('已取消删除')
      })
    },

    // 处理文件夹更新
    handleFolderUpdate(folder) {
      console.log('更新文件夹', folder)
      let _that = this
      const folderNode = _that.findFolderNode(folder.collection_id, folder.id)
      if (folderNode) {
        _that.syncFolderNodeFields(folderNode, folder)
        _that.upsertWorkspaceTabData(folderNode)
      } else {
        _that.upsertWorkspaceTabData(folder)
      }
    },

    handleFolderDelete: function (folder) {
      console.log('删除文件夹', folder)
      let _that = this
      _that.$confirm('确定要删除这个文件夹吗？此操作不可恢复。', '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        Api.DeleteDir(folder, function (res) {
          if (res.ErrCode === 0) {
            // 关闭该文件夹及其下所有接口的 tab
            _that.closeWorkspaceTabsByFolder(folder.id)
            _that.refreshCollectionFolders(folder.collection_id).finally(() => {
              _that.syncTreeSortCacheFromTree()
            })
            _that.$message.success('删除成功')
          } else {
            _that.$message.error(res.ErrMsg)
          }
        })
      }).catch(() => {
        _that.$message.info('已取消删除')
      })
    },
    handleFolderCreateApi: function () {
      let _that = this
      if (!_that.dialogShow.createApi) {
        _that.dialogShow.createApi = true
        _that.dialogData.createApi.curlData = ''
        return
      }
      if (_that.beginDialogSubmit('createApi')) {
        return
      }
      _that.dialogData.createApi.folder_id = _that.selectedItem.id
      _that.dialogData.createApi.collection_id = _that.selectedItem.collection_id
      Api.CreateApi(_that.dialogData.createApi, function (res) {
        _that.endDialogSubmit('createApi')
        _that.dialogShow.createApi = false
        if (res.ErrCode !== 0) {
          _that.$message.error(res.ErrMsg)
          return
        }
        let newApi = res.Data
        _that.refreshFolderApis(_that.dialogData.createApi.collection_id, _that.dialogData.createApi.folder_id).then(async () => {
          _that.syncTreeSortCacheFromTree()
          // 新建接口成功后，自动打开新接口 tab
          const newApiNode = _that.normalizeApiNode(newApi, newApi.folder_id, newApi.collection_id)
          _that.upsertWorkspaceTabData(_that.findFolderNode(newApi.collection_id, newApi.folder_id))
          await _that.openWorkspaceTab(newApiNode, { reload: true })
        })
      })
    },
    // 处理接口更新
    handleApiUpdate(api) {
      console.log('更新api', api)
      // 实现接口更新逻辑
      let _that = this
      Api.CreateApi({
        id: api.id,
        folder_id: api.folder_id,
        collection_id: api.collection_id,
        name: api.name,
        method: api.method,
        url: api.url,
        protocol: api.protocol,
        desc: api.desc,
        headers: api.header_list,
        query_params: api.query_params_data,
        content_type: api.content_type,
        body_form: api.body_form_data,
        body_json: api.body_json_data,
        env_id: api.env_id,
        response_take: api.response_take_data,
        take_result: api.take_result_data,
      }, function (res) {
        if (res.ErrCode !== 0) {
          _that.$message.error(res.ErrMsg)
          return
        }
        let newApi = res.Data
        const currentApiNode = _that.findApiNode(api.collection_id, api.folder_id, api.id)
        if (currentApiNode) {
          _that.syncApiNodeFields(currentApiNode, newApi)
        }
        // 更新工作区 tab 数据
        _that.upsertWorkspaceTabData(_that.normalizeApiNode(newApi, newApi.folder_id, newApi.collection_id))
      })
    },

    // 处理API操作
    handleApiAction(command, data) {
      let _that = this
      if (command === 'copy_api') {
        _that.openCopyApiDialog(data)
      } else if (command === 'delete_api') {
        _that.$confirm(`确定要删除接口 "${data.name}" 吗？`, '确认删除', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(() => {
          Api.DeleteApi(data, function (res) {
            if (res.ErrCode === 0) {
              // 关闭该接口的 tab
              const tabKey = _that.buildWorkspaceTabKey(data)
              _that.closeWorkspaceTab(tabKey)
              _that.refreshFolderApis(data.collection_id, data.folder_id).finally(() => {
                _that.syncTreeSortCacheFromTree()
              })
              _that.$message.success('删除成功')
            } else {
              _that.$message.error(res.ErrMsg)
            }
          })
        }).catch(() => {
          // 用户点击取消，不做任何操作
          _that.$message.info('已取消删除')
        })
      } else if (command === 'down_api') {
        Api.ApiWeightDown(data, function (res) {
          if (res.ErrCode === 0) {
            _that.refreshFolderApis(data.collection_id, data.folder_id).then(() => {
              _that.upsertWorkspaceTabData(_that.findFolderNode(data.collection_id, data.folder_id))
            })
            _that.$message.success('移动成功')
          } else {
            _that.$message.error(res.ErrMsg)
          }
        })
      }
    },

    // 打开复制接口对话框
    openCopyApiDialog(api) {
      // 复制API数据到复制对话框
      this.dialogData.copyApi = JSON.parse(JSON.stringify(api))
      this.dialogData.copyApi.id = 0
      this.dialogData.copyApi.name = api.name + '-复制'
      this.dialogShow.copyApi = true
    },
    // 打开JSON导入对话框
    openJsonImportDialog(collection) {
      this.dialogData.jsonImport.collection_id = collection.id
      this.dialogData.jsonImport.json = ''
      this.dialogShow.jsonImport = true
    },
    // JSON导入
    apiImportJson() {
      let _that = this
      if (_that.beginDialogSubmit('jsonImport')) {
        return
      }
      if (!this.dialogData.jsonImport.collection_id) {
        _that.endDialogSubmit('jsonImport')
        _that.$message.error('请选择集合')
        return
      }
      if (!this.dialogData.jsonImport.json) {
        _that.endDialogSubmit('jsonImport')
        _that.$message.error('请输入JSON数据')
        return
      }
      try {
        // Validate JSON format
        JSON.parse(this.dialogData.jsonImport.json)
      } catch (e) {
        _that.endDialogSubmit('jsonImport')
        _that.$message.error('JSON格式错误，请检查输入')
        return
      }
      Api.ApiImportJson({
        collection_id: this.dialogData.jsonImport.collection_id,
        json: this.dialogData.jsonImport.json
      }, function (res) {
        _that.endDialogSubmit('jsonImport')
        if (res.ErrCode === 0) {
          _that.$message.success('导入成功')
          _that.dialogShow.jsonImport = false
          _that.loadCollectionData()
        } else {
          _that.$message.error(res.ErrMsg)
        }
      })
    },
    // 复制接口
    copyApi() {
      let _that = this
      if (_that.beginDialogSubmit('copyApi')) {
        return
      }
      Api.CreateApi(_that.dialogData.copyApi, function (res) {
        _that.endDialogSubmit('copyApi')
        _that.dialogShow.copyApi = false
        if (res.ErrCode !== 0) {
          _that.$message.error(res.ErrMsg)
          return
        }
        let newApi = res.Data
        _that.refreshFolderApis(_that.dialogData.copyApi.collection_id, _that.dialogData.copyApi.folder_id).then(async () => {
          _that.syncTreeSortCacheFromTree()
          // 复制接口成功后，自动打开新接口 tab
          const newApiNode = _that.normalizeApiNode(newApi, newApi.folder_id, newApi.collection_id)
          _that.upsertWorkspaceTabData(_that.findFolderNode(newApi.collection_id, newApi.folder_id))
          await _that.openWorkspaceTab(newApiNode, { reload: true })
        })
        _that.$message.success('接口复制成功')
      })
    }
  }
}
</script>

<style scoped>
.collection-container {
  display: flex;
  height: 100%;
  min-height: 100%;
  min-width: 0;
  background-color: transparent;
  width: 100%;
  box-sizing: border-box;
  gap: 12px;
  color: #4a4a4a;
}

.collection-container.resizing {
  user-select: none;
}

.left-sidebar {
  display: flex;
  min-height: 0;
  background: #fff;
  border: 1px solid #e8e8e0;
  border-radius: 12px;
  flex-direction: column;
  width: 280px;
  min-width: 250px;
  max-width: 350px;
  flex-shrink: 0;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(54, 74, 54, 0.05);
}

.collection-section {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-height: 0;
}

.panel-resizer {
  width: 8px;
  flex-shrink: 0;
  cursor: col-resize;
  position: relative;
  border-radius: 6px;
  transition: background-color 0.2s ease;
}

.panel-resizer::before {
  content: '';
  position: absolute;
  left: 3px;
  top: 10px;
  bottom: 10px;
  width: 2px;
  border-radius: 2px;
  background: #d6dfd2;
}

.panel-resizer:hover::before {
  background: #9fb59a;
}

.section-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  width: 100%;
}

.section-header-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.toolbar-btn {
  padding: 6px 10px;
}

.toolbar-btn-small {
  padding: 4px 8px;
  font-size: 12px;
}

.toolbar-btn-mini {
  padding: 4px 7px;
  font-size: 12px;
}

.skill-install-popover {
  padding: 2px;
}

.skill-install-title {
  font-size: 14px;
  font-weight: 600;
  color: #4f804f;
  margin-bottom: 6px;
}

.skill-install-text {
  font-size: 13px;
  color: #5e6b57;
  margin-bottom: 8px;
}

.skill-install-feature {
  font-size: 12px;
  color: #4f804f;
  margin-bottom: 8px;
  line-height: 1.5;
}

.skill-install-url {
  margin-bottom: 8px;
}

.skill-install-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.section-header {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid #ecece4;
  background: #f7f7f2;
  font-weight: 600;
  cursor: pointer;
  user-select: none;
  color: #4a4a4a;
}

.section-header-title {
  line-height: 1.2;
}

.collection-list {
  flex: 1;
  overflow-y: auto;
  overflow-x: auto;
  min-height: 0;
  padding: 8px;
  padding-bottom: 0;
}

.tree-wrapper {
  min-width: 100%;
  display: block;
}

.collection-list .el-tree {
  min-width: 100%;
  white-space: nowrap;
  color: #4a4a4a;
}

.tree-node {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 4px 0;
  white-space: nowrap;
  position: relative;
}

.node-icon {
  margin-right: 6px;
  color: #5a8a5a;
  flex-shrink: 0;
}

.node-label {
  flex: 1;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
  cursor: default;
  max-width: calc(100% - 40px);
  margin-right: 10px;
}

.node-actions {
  opacity: 0;
  transition: opacity 0.3s;
  white-space: nowrap;
  position: absolute;
  right: 0;
  background: #fff;
  z-index: 1;
  padding-left: 10px;
  padding-right: 5px;
}

.tree-node:hover .node-actions {
  opacity: 1;
}

.collection-list .el-tree-node__content {
  position: relative;
  height: auto;
  min-height: 32px;
  line-height: 24px;
  padding-right: 0;
  white-space: nowrap;
  overflow: visible;
  border-radius: 8px;
}

.collection-list :deep(.el-tree-node__content:hover) {
  background: #f3f7ef;
}

.archive-section {
  border-top: 1px solid #ecece4;
}

.archive-list {
  max-height: 200px;
  overflow-y: auto;
  padding: 8px;
}

.archive-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 4px;
  margin-bottom: 4px;
}

.archive-item:hover {
  background-color: #f3f7ef;
}

.archive-item .el-icon {
  margin-right: 8px;
  color: #7a8773;
}

.collapse-icon {
  transition: transform 0.3s;
}

.rotate-180 {
  transform: rotate(180deg);
}

/* 右侧面板样式 */
.right-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid #e8e8e0;
  border-radius: 12px;
  min-width: 0;
  overflow: hidden;
  box-shadow: 0 4px 12px rgba(54, 74, 54, 0.05);
}

/* 工作区 Tab 样式 */
.workspace-tabs {
  background: #f7f7f2;
  border-bottom: 1px solid #ecece4;
  padding: 0 12px;
}

.workspace-tabs :deep(.el-tabs__header) {
  margin: 0;
  border-bottom: none;
}

.workspace-tabs :deep(.el-tabs__nav-wrap) {
  padding: 8px 0 0 0;
}

.workspace-tabs :deep(.el-tabs__item) {
  height: 32px;
  line-height: 32px;
  padding: 0 12px;
  border-radius: 6px 6px 0 0;
  background: #fff;
  border: 1px solid #e8e8e0;
  border-bottom: none;
  margin-right: 4px;
  color: #5e6b57;
  font-size: 13px;
}

.workspace-tabs :deep(.el-tabs__item.is-active) {
  color: #4f804f;
  background: #fff;
  border-bottom: 1px solid #fff;
}

.workspace-tabs :deep(.el-tabs__item:hover) {
  color: #4f804f;
}

.workspace-tabs :deep(.el-tabs__nav-prev),
.workspace-tabs :deep(.el-tabs__nav-next) {
  line-height: 40px;
}

.workspace-tab-label {
  display: flex;
  align-items: center;
  gap: 4px;
}

.workspace-tab-label .tab-icon {
  font-size: 14px;
}

.workspace-tab-label .tab-title {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #ecece4;
  background: #f7f7f2;
}

.header-left {
  display: flex;
  align-items: center;
}

.selected-info {
  display: flex;
  align-items: center;
}

.info-icon {
  font-size: 24px;
  color: #5a8a5a;
  margin-right: 12px;
}

.info-content {
  display: flex;
  flex-direction: column;
}

.info-name {
  font-size: 16px;
  font-weight: 600;
  color: #4a4a4a;
}

.info-desc {
  font-size: 12px;
  color: #7a8773;
  margin-top: 4px;
}

.no-selection {
  color: #7a8773;
  font-style: italic;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.panel-content {
  flex: 1;
  padding: 12px 16px;
  overflow-y: auto;
  background: #fff;
}

.panel-content--flush {
  padding: 0;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}

.panel-content :deep(.el-tabs--card > .el-tabs__header) {
  border-bottom-color: #ecece4;
}

.panel-content :deep(.el-tabs--card > .el-tabs__header .el-tabs__item) {
  border-color: #ecece4;
  color: #5e6b57;
}

.panel-content :deep(.el-tabs--card > .el-tabs__header .el-tabs__item.is-active) {
  color: #4f804f;
  background: #f6f8f3;
}

.panel-content :deep(.el-button--primary) {
  --el-button-bg-color: #5a8a5a;
  --el-button-border-color: #5a8a5a;
  --el-button-hover-bg-color: #4f804f;
  --el-button-hover-border-color: #4f804f;
}

.panel-content :deep(.el-tag--success) {
  --el-tag-bg-color: #eef7ea;
  --el-tag-border-color: #cfe0c8;
  --el-tag-text-color: #4f804f;
}

.panel-content :deep(.el-tag--primary) {
  --el-tag-bg-color: #eef4ea;
  --el-tag-border-color: #cfe0c8;
  --el-tag-text-color: #4f804f;
}

@media (max-width: 1200px) {
  .collection-container {
    flex-direction: column;
    height: auto;
    min-height: 0;
  }

  .left-sidebar {
    width: 100% !important;
    max-width: 100%;
    min-width: 0;
  }

  .panel-resizer {
    display: none;
  }

  .right-panel {
    min-height: 500px;
  }
}

</style>




