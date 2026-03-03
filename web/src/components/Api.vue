<template>
  <div class="collection-container" :class="{ resizing: sidebarResizing }">
    <!-- 左侧集合列表 -->
    <div class="left-sidebar" :style="{ width: sidebarWidth + 'px' }">
      <!-- 集合列表区域 -->
      <div class="collection-section">
        <div class="section-header">
          <span class="section-header-title">集合列表</span>
          <div class="section-header-actions">
            <el-button class="toolbar-btn toolbar-btn-small" size="small" type="primary" plain @click="openCreateCollectionDialog">
              <el-icon><Plus /></el-icon>新建集合
            </el-button>
            <el-button class="toolbar-btn toolbar-btn-small" size="small" type="primary" plain @click="drawerVisibleMarkdown = true">
              <el-icon><QuestionFilled /></el-icon>文档
            </el-button>
            <el-popover placement="bottom-end" :width="360" trigger="click">
              <template #reference>
                <el-button class="toolbar-btn toolbar-btn-mini" size="small" type="info" plain>
                  <el-icon><Tools /></el-icon>Skills安装
                </el-button>
              </template>
              <div class="skill-install-popover">
                <div class="skill-install-title">dtool-api-import-update 安装</div>
                <div class="skill-install-text">把下面 ZIP 地址发给编辑器 AI，让它按 ZIP 安装 skills。</div>
                <div class="skill-install-feature">
                  功能说明：安装后可通过 AI 直接在此工具中生成接口（支持集合/文件夹选择、按 URI 导入或更新）。
                </div>
                <el-input :model-value="skillInstallZipUrl" readonly class="skill-install-url" />
                <div class="skill-install-actions">
                  <el-button size="small" type="primary" plain @click="copyText(skillInstallZipUrl, 'ZIP 地址已复制')">复制 ZIP 地址</el-button>
                  <el-button size="small" type="primary" plain @click="copyText(skillInstallPrompt, 'AI 安装提示已复制')">复制安装提示</el-button>
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
                <div class="tree-node">
                  <span class="node-icon">
                    <el-icon v-if="data.type === 'collection'"><Files/></el-icon>
                    <el-icon v-else-if="data.type === 'folder'"><Folder/></el-icon>
                    <!--                    <el-icon v-else><Document /></el-icon>-->
                  </span>
                  <span v-if="data.type === 'folder'" :title="node.label + '(' + (data.children ? data.children.length : 0) + ')'" class="node-label" style="font-weight: 500;">{{
                      node.label + '(' + (data.children ? data.children.length : 0) + ')'
                    }}</span>
                  <span v-if="data.type === 'collection'" :title="node.label + '(' + (data.children ? data.children.length : 0) + ')'" class="node-label" style="font-weight: 800;">{{
                      node.label + '(' + (data.children ? data.children.length : 0) + ')'
                    }}</span>
                  <span v-if="data.type === 'api'" :title="node.label" class="node-label">
                    <el-tag v-if="data.method === 'GET'" size="small" type="success">G</el-tag>
                    <el-tag v-if="data.method === 'POST'" size="small" type="primary">P</el-tag>
                    {{ node.label }}
                  </span>
                  <span v-if="data.type === 'collection'" class="node-actions">
                    <el-button link type="primary" @click.stop="toggleCollection(data)">
                      <el-dropdown>
                      <el-button link type="primary" @click.stop>
                        <el-icon><More/></el-icon>
                      </el-button>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item command="copy_api" icon="CopyDocument" @click="createNewDir(data)">创建文件夹</el-dropdown-item>
                          <el-dropdown-item command="json_import" icon="Upload" @click="openJsonImportDialog(data)">通过json导入</el-dropdown-item>
                          <el-dropdown-item command="delete_collection" icon="Delete" @click="handleCollectionDelete(data)" style="color:red;">删除集合</el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                      </el-dropdown>
                    </el-button>
                  </span>
                  <span v-else-if="data.type === 'api'" class="node-actions">
                    <el-dropdown>
                      <el-button link type="primary" @click.stop>
                        <el-icon><More/></el-icon>
                      </el-button>
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
                      <el-button link type="primary" @click.stop>
                        <el-icon><More/></el-icon>
                      </el-button>
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
      <!-- 顶部信息栏 -->
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
          <el-button v-if="selectedItem.type === 'collection'" type="primary" @click="createNewDir(selectedItem)">创建文件夹
          </el-button>
          <el-button type="primary" @click="executeAll">运行全部</el-button>
          <el-button @click="exportCollection">导出</el-button>
        </div>
      </div>

      <!-- 内容区域 -->
      <div class="panel-content">
        <!-- 集合设置 -->
        <div v-if="selectedItem && selectedItem.type === 'collection'" class="collection-settings">
          <el-tabs v-model="collectionActiveTab" type="card">
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
              :handleCreateApi="handleFolderCreateApi"
              @delete="handleFolderDelete"
              @update="handleFolderUpdate"
          />
        </div>

        <!-- 接口设置 -->
        <div v-else-if="selectedItem && selectedItem.type === 'api'" class="api-settings">
          <api-detail
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
        <el-button @click="dialogShow.createCollection = false">取消</el-button>
        <el-button type="primary" @click="createNewCollection">保存</el-button>
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
        <el-button @click="dialogShow.createDir = false">取消</el-button>
        <el-button type="primary" @click="createNewDir">保存</el-button>
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
        <el-button @click="dialogShow.createApi = false">取消</el-button>
        <el-button type="primary" @click="handleFolderCreateApi">保存</el-button>
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
        <el-button @click="dialogShow.copyApi = false">取消</el-button>
        <el-button type="primary" @click="copyApi">保存</el-button>
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
        <el-button @click="dialogShow.jsonImport = false">取消</el-button>
        <el-button type="primary" @click="apiImportJson">导入</el-button>
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
import {FolderOpened, Folder, Document, ArrowDown, ArrowUp, More, Plus, QuestionFilled, Tools} from '@element-plus/icons-vue'
import CollectionBasicInfo from './api/CollectionBasicInfo'
import CollectionEnvironment from './api/CollectionEnvironment'
import CollectionPermission from './api/CollectionPermission'
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
    ArrowDown,
    ArrowUp,
    More,
    Plus,
    QuestionFilled,
    Tools,
    CollectionBasicInfo,
    CollectionEnvironment,
    CollectionPermission,
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
        label: 'name'
      },
      createApiType: 'params',
      defaultExpandedKeys: [],

      // 归档相关
      archiveExpanded: false,
      archivedItems: [],

      // 选中项
      selectedItem: null,

      // 文档drawer
      drawerVisibleMarkdown: false,
      markdownType: 'api',
      skillInstallZipUrl: 'https://gitee.com/Sxiaobai/skills/raw/master/dtool-api-import-update.zip',
      skillInstallPrompt: '请安装这个 skills zip： https://gitee.com/Sxiaobai/skills/raw/master/dtool-api-import-update.zip',

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


    // 加载集合数据
    loadCollectionData() {
      let _that = this
      Api.Collections({}, function (res) {
        if (res.ErrCode === 0) {
          _that.treeData = res.Data.list
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
      const expandedState = _that.getExpandedStateCache()
      if (!Array.isArray(expandedState) || expandedState.length === 0) {
        _that.expandAllNodes()
        return
      }
      _that.$nextTick(() => {
        _that.applyExpandedStateFromCache(expandedState)
      })
    },

    // 展开所有节点
    expandAllNodes() {
      let _that = this
      _that.$nextTick(() => {
        _that.$refs.collectionTreeRef.store._getAllNodes().forEach(node => {
          if (node.level > 0) {
            node.expand()
          }
        })
        // 首次默认展开后，立即写入缓存，保证刷新可恢复
        _that.syncExpandedStateCacheFromTree()
      })
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
        return []
      }
      try {
        const cacheData = JSON.parse(cacheText)
        return Array.isArray(cacheData) ? cacheData : []
      } catch (e) {
        return []
      }
    },
    // 写入树展开缓存
    setExpandedStateCache(expandedState) {
      store.setStore('collection_tree_expand_state', JSON.stringify(expandedState))
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
    applyExpandedStateFromCache(expandedState) {
      let _that = this
      if (!_that.$refs.collectionTreeRef) {
        return
      }
      _that.treeData.forEach(collection => {
        const collectionKey = _that.buildExpandStateKey(collection)
        if (collectionKey && expandedState.includes(collectionKey)) {
          const collectionNode = _that.$refs.collectionTreeRef.getNode(collection.uniqueid)
          if (collectionNode) {
            collectionNode.expand()
          }
        }
        if (!Array.isArray(collection.children)) {
          return
        }
        collection.children.forEach(folder => {
          const folderKey = _that.buildExpandStateKey(folder)
          if (folderKey && expandedState.includes(folderKey)) {
            const folderNode = _that.$refs.collectionTreeRef.getNode(folder.uniqueid)
            if (folderNode) {
              folderNode.expand()
            }
          }
        })
      })
    },
    // 更新展开缓存
    updateExpandedCache(data, isExpanded) {
      let _that = this
      const nodeKey = _that.buildExpandStateKey(data)
      if (!nodeKey) {
        return
      }
      let expandedState = _that.getExpandedStateCache()
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
      this.$refs.collectionTreeRef.updateKeyChildren(collection.id, collection.children)
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

        } else if ((key1 === 'Control' && key2 === 'Enter') || (key1 === 'Enter' && key2 === 'Control')) {
          if (_that.selectedItem.type === 'api') {
            _that.$nextTick(() => {
              _that.$refs.refApiDetail.handleExecute();
            });
          }

        }
      }, 500)
    },
    // 处理节点点击
    handleNodeClick(data) {
      let _that = this
      if (data.type && data.type !== "api") {
        const node = this.$refs.collectionTreeRef.getNode(data.uniqueid)
        if (node) {
          if (!node.expanded) {
            node.expand()
          }
          if (data.type === 'folder') {
            _that.fillCollectionApis(data.collection_id, data.id)
          }
        }

      } else if (data.type && data.type === 'api') {
        _that.$nextTick(() => {
          _that.$refs.refApiDetail.InitApiDetail(data);
        });
      }
      _that.selectedItem = data
    },
    fillCollectionApis: function (collection_id, dir_id) {
      let _that = this
      Api.Apis({
        collection_id: collection_id,
        dir_id: dir_id
      }, function (res) {
        if (res.ErrCode !== 0) {
          _that.$message.error(res.ErrMsg)
          return
        }
        for (let i in _that.treeData) {
          if (parseInt(collection_id) === parseInt(_that.treeData[i].id)) {
            for (let j in _that.treeData[i].children) {
              if (parseInt(dir_id) === parseInt(_that.treeData[i].children[j].id)) {
                _that.treeData[i].children[j].children = res.Data.list
              }
            }
          }
        }
        // 拉取接口后按缓存顺序恢复
        _that.applyFolderApiSort(collection_id, dir_id)
      })
    },
    // 处理归档项点击
    handleArchiveItemClick(item) {
      this.selectedItem = item
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
          let newCollection = res.Data
          newCollection.children = []
          _that.pushUniqueByKey(_that.treeData, newCollection, 'uniqueid')
          _that.syncTreeSortCacheFromTree()
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
          _that.selectedItem = data
          _that.$refs.collectionTreeRef.setCurrentKey(_that.selectedItem.uniqueid)
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
          let newDir = res.Data
          newDir.children = []
          for (let i in _that.treeData) {
            if (parseInt(_that.dialogData.createDir.collection_id) === parseInt(_that.treeData[i].id)) {
              _that.pushUniqueByKey(_that.treeData[i].children, newDir, 'uniqueid')
            }
          }
          _that.syncTreeSortCacheFromTree()
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
              _that.treeData[i] = collection
              _that.selectedItem = collection
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
            //如果是删除的集合
            if (_that.selectedItem && _that.selectedItem.type === 'collection') {
              if (_that.selectedItem.id === collection.id) {
                _that.selectedItem = {}
              }
            }
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
      for (let i in _that.treeData) {
        if (parseInt(folder.collection_id) !== parseInt(_that.treeData[i].id)) {
          continue
        }
        for (let j in _that.treeData[i].children) {
          if (parseInt(folder.id) !== parseInt(_that.treeData[i].children[j].id)) {
            continue
          }
          // Use Object.assign to ensure reactivity
          Object.assign(_that.treeData[i].children[j], {
            name: folder.name,
            desc: folder.desc
          })
          break
        }
      }
      // Update the selected item if it's the same folder
      if (_that.selectedItem && parseInt(_that.selectedItem.id) === parseInt(folder.id)) {
        Object.assign(_that.selectedItem, {
          name: folder.name,
          desc: folder.desc
        })
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
            for (let i in _that.treeData) {
              if (parseInt(folder.collection_id) !== parseInt(_that.treeData[i].id)) {
                continue
              }
              _that.treeData[i].children = ArrayUtil.DeleteValueByStringKey(_that.treeData[i].children, 'uniqueid', folder.uniqueid)
            }

            //如果是删除的集合
            if(_that.selectedItem){
              if (_that.selectedItem.type === 'folder') {
                if (_that.selectedItem.id === folder.id) {
                  _that.selectedItem = {}
                }
              } else if (_that.selectedItem.type === 'api') {
                if (_that.selectedItem.folder_id === folder.id) {
                  _that.selectedItem = {}
                }
              }
            }
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
        for (let i in _that.treeData) {
          if (parseInt(_that.dialogData.createApi.collection_id) !== parseInt(_that.treeData[i].id)) {
            continue
          }
          for (let j in _that.treeData[i].children) {
            if (parseInt(_that.dialogData.createApi.folder_id) !== parseInt(_that.treeData[i].children[j].id)) {
              continue
            }
            _that.pushUniqueByKey(_that.treeData[i].children[j].children, newApi, 'uniqueid')
          }
        }
        _that.syncTreeSortCacheFromTree()
        _that.selectedItem = newApi
        _that.$nextTick(function () {
          _that.$refs.refApiDetail.InitApiDetail(newApi)
          _that.$refs.collectionTreeRef.setCurrentKey(newApi.uniqueid)
          console.log('设置当前选中的菜单为', newApi.uniqueid)
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
        for (let i in _that.treeData) {
          if (parseInt(api.collection_id) !== parseInt(_that.treeData[i].id)) {
            continue
          }
          for (let j in _that.treeData[i].children) {
            if (parseInt(api.folder_id) !== parseInt(_that.treeData[i].children[j].id)) {
              continue
            }
            for (let k in _that.treeData[i].children[j].children) {
              if (parseInt(api.id) !== parseInt(_that.treeData[i].children[j].children[k].id)) {
                continue
              }
              _that.treeData[i].children[j].children[k] = newApi
            }
          }
        }
        _that.selectedItem = newApi
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
              for (let i in _that.treeData) {
                let collectionInfo = _that.treeData[i]
                if (parseInt(data.collection_id) !== parseInt(collectionInfo.id)) { //集合
                  continue
                }
                for (let j in collectionInfo.children) { //文件夹
                  let folderInfo = collectionInfo.children[j]
                  if (parseInt(data.folder_id) !== parseInt(folderInfo.id)) {
                    continue
                  }
                  _that.treeData[i].children[j].children = ArrayUtil.DeleteValueByStringKey(folderInfo.children, 'uniqueid', data.uniqueid)
                }
              }
              if (_that.selectedItem.uniqueid === data.uniqueid) {//如果当前删除的api是选中的api 那么置空当前选项
                _that.selectedItem = {}
              }
              _that.syncTreeSortCacheFromTree()
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
            _that.fillCollectionApis(data.collection_id, data.folder_id)
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
        // 将新接口添加到对应文件夹下
        for (let i in _that.treeData) {
          if (parseInt(_that.dialogData.copyApi.collection_id) !== parseInt(_that.treeData[i].id)) {
            continue
          }
          for (let j in _that.treeData[i].children) {
            if (parseInt(_that.dialogData.copyApi.folder_id) !== parseInt(_that.treeData[i].children[j].id)) {
              continue
            }
            _that.pushUniqueByKey(_that.treeData[i].children[j].children, newApi, 'uniqueid')
          }
        }
        _that.syncTreeSortCacheFromTree()
        _that.selectedItem = newApi
        _that.$nextTick(() => {
          _that.$refs.collectionTreeRef.setCurrentKey(newApi.uniqueid)
        })
        _that.$refs.refApiDetail.InitApiDetail(newApi);
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



