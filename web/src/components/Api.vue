<template>
  <div class="collection-container" :class="{ resizing: sidebarResizing }">
    <!-- 左侧集合列表 -->
    <div class="left-sidebar" :style="{ width: sidebarWidth + 'px' }">
      <!-- 集合列表区域 -->
      <div class="collection-section">
        <div class="section-header">
          <span class="section-header-title">集合列表</span>
          <div v-if="!filterFolderId" class="section-header-actions">
            <pl-button class="toolbar-btn toolbar-btn-small" size="small" type="primary" plain @click="openCreateCollectionDialog">
              <el-icon><Plus /></el-icon>新建集合
            </pl-button>
            <pl-button class="toolbar-btn toolbar-btn-small" size="small" @click="openCleanupDialog">
              清理
            </pl-button>
            <pl-button class="toolbar-btn toolbar-btn-small" size="small" type="primary" plain @click="drawerVisibleMarkdown = true">
              <el-icon><QuestionFilled /></el-icon>文档
            </pl-button>
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
                :draggable="!filterFolderId"
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
                    <el-icon v-else-if="data.type === 'archive'"><FolderOpened/></el-icon>
                    <el-icon v-else-if="data.type === 'folder'"><Folder/></el-icon>
                    <!--                    <el-icon v-else><Document /></el-icon>-->
                  </span>
                  <span v-if="data.type === 'folder'" :title="node.label + '(' + getNodeChildCount(data) + ')'" class="node-label node-label--folder">{{
                      node.label + '(' + getNodeChildCount(data) + ')'
                    }}</span>
                  <span v-if="data.type === 'collection'" :title="node.label + '(' + getNodeChildCount(data) + ')'" class="node-label node-label--collection">{{
                      node.label + '(' + getNodeChildCount(data) + ')'
                    }}</span>
                  <span v-if="data.type === 'archive'" :title="node.label + '(' + getNodeChildCount(data) + ')'" class="node-label node-label--archive">{{
                      node.label + '(' + getNodeChildCount(data) + ')'
                    }}</span>
                  <span v-if="data.type === 'api'" :title="node.label" class="node-label">
                    <el-tag v-if="data.method === 'GET'" size="small" type="success">G</el-tag>
                    <el-tag v-if="data.method === 'POST'" size="small" type="primary">P</el-tag>
                    {{ node.label }}
                  </span>
                  <span v-if="data.type === 'collection' && !filterFolderId" class="node-actions">
                    <el-dropdown>
                      <pl-button class="node-action-trigger" link type="primary" @click.stop>
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
                  </span>
                  <span v-else-if="data.type === 'api'" class="node-actions">
                    <el-dropdown>
                      <pl-button class="node-action-trigger" link type="primary" @click.stop>
                        <el-icon><More/></el-icon>
                      </pl-button>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item command="copy_api" icon="CopyDocument" @click="handleApiAction('copy_api' , data)">复制接口</el-dropdown-item>
                          <el-dropdown-item command="move_api" icon="FolderOpened" @click="handleApiAction('move_api' , data)">迁移</el-dropdown-item>
                          <el-dropdown-item command="delete_api" icon="Move" @click="handleApiAction('delete_api' , data)">删除接口</el-dropdown-item>
                          <el-dropdown-item command="down_api" icon="Move" @click="handleApiAction('down_api' , data)">下移</el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                  </span>
                  <span v-else-if="data.type === 'folder' && !data.is_archived && !filterFolderId" class="node-actions">
                    <el-dropdown>
                      <pl-button class="node-action-trigger" link type="primary" @click.stop>
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
                  <span v-else-if="data.type === 'folder' && data.is_archived" class="node-actions">
                    <el-dropdown>
                      <pl-button class="node-action-trigger" link type="primary" @click.stop>
                        <el-icon><More/></el-icon>
                      </pl-button>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item icon="RefreshLeft" @click="handleRestoreFolder(data)">恢复</el-dropdown-item>
                          <el-dropdown-item icon="Delete" style="color:red;" @click="handlePermanentDeleteFolder(data)">永久删除</el-dropdown-item>
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

  <el-dialog v-model="dialogShow.moveApi" title="迁移接口" width="560" @keydown.enter.prevent="handleDialogEnter('moveApi', $event)">
    <el-form :model="dialogData.moveApi" label-width="100px" @submit.prevent>
      <el-form-item label="接口名称">
        <el-input :model-value="dialogData.moveApi.name" readonly />
      </el-form-item>
      <el-form-item label="当前位置">
        <el-input :model-value="dialogData.moveApi.current_path" readonly />
      </el-form-item>
      <el-form-item label="目标文件夹">
        <el-select
          v-model="dialogData.moveApi.folder_id"
          filterable
          clearable
          placeholder="请选择目标文件夹"
          class="dialog-select-full"
          :loading="moveApiFolderOptionsLoading"
        >
          <el-option-group
            v-for="group in moveApiFolderOptions"
            :key="group.collection_id"
            :label="group.collection_name"
          >
            <el-option
              v-for="folder in group.folders"
              :key="folder.id"
              :label="folder.optionLabel"
              :value="folder.id"
              :disabled="folder.disabled"
            />
          </el-option-group>
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="dialogShow.moveApi = false">取消</pl-button>
        <pl-button type="primary" @click="moveApi">确认迁移</pl-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="dialogShow.jsonImport" title="通过JSON导入" width="800" @keydown.enter.prevent="handleDialogEnter('jsonImport', $event)">
    <el-form :model="dialogData.jsonImport" label-width="120px" @submit.prevent>
      <el-form-item label="选择集合">
        <el-select v-model="dialogData.jsonImport.collection_id" placeholder="请选择集合" class="dialog-select-full">
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

  <el-dialog v-model="dialogShow.cleanupFolders" title="清理文件夹" width="760" @closed="resetCleanupSelection">
    <div class="cleanup-dialog">
      <div class="cleanup-toolbar">
        <div class="cleanup-toolbar-main">
          <span class="cleanup-toolbar-label">默认显示</span>
          <el-input-number v-model="dialogData.cleanupFolders.days" :min="1" :max="3650" controls-position="right" />
          <span class="cleanup-toolbar-label">天未使用</span>
          <pl-button type="primary" :loading="cleanupLoading" @click="checkCleanupFolders">检查</pl-button>
        </div>
        <el-checkbox
          v-if="dialogData.cleanupFolders.candidates.length > 0"
          v-model="dialogData.cleanupFolders.selectAll"
          @change="toggleCleanupSelectAll"
        >
          全选
        </el-checkbox>
      </div>

      <div class="cleanup-result">
        <el-empty v-if="!cleanupLoading && dialogData.cleanupFolders.checkedAt === 0" description="点击检查后列出可清理文件夹" />
        <el-empty v-else-if="!cleanupLoading && dialogData.cleanupFolders.candidates.length === 0" description="没有符合条件的文件夹" />
        <div v-else class="cleanup-list">
          <label v-for="folder in dialogData.cleanupFolders.candidates" :key="folder.id" class="cleanup-item">
            <el-checkbox v-model="dialogData.cleanupFolders.selectedIds" :label="folder.id">
              <span class="cleanup-item-title">{{ folder.collection_name }} / {{ folder.name }}</span>
            </el-checkbox>
            <div class="cleanup-item-meta">
              <span>接口数 {{ folder.child_count }}</span>
              <span>最近活动 {{ formatTimestamp(folder.latest_active_time) }}</span>
              <span>最近执行 {{ formatTimestamp(folder.last_run_time) }}</span>
              <span>最近修改 {{ formatTimestamp(Math.max(Number(folder.update_time || 0), Number(folder.last_api_update_time || 0))) }}</span>
            </div>
          </label>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="dialogShow.cleanupFolders = false">取消</pl-button>
        <pl-button type="danger" :disabled="dialogData.cleanupFolders.selectedIds.length === 0" :loading="cleanupDeleting" @click="confirmCleanupDelete">删除</pl-button>
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
import {FolderOpened, Folder, Document, More, Plus, QuestionFilled, Tools, CopyDocument, RefreshLeft} from '@element-plus/icons-vue'
import CollectionBasicInfo from './api/CollectionBasicInfo'
import CollectionEnvironment from './api/CollectionEnvironment'
import FolderDetail from './api/FolderDetail'
import ApiDetail from './api/ApiDetail'
import Markdown from '@/components/Markdown.vue'
import Api from '@/utils/base/api'
import Base from '@/utils/base'
import ArrayUtil from '@/utils/base/array'
import KeyDebounceDetector from "@/utils/base/keyup"
import store from "@/utils/base/store";
import sseDistribute from '@/utils/base/sse_distribute'

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
    CopyDocument,
    RefreshLeft,
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
        moveApi: false, //迁移接口弹窗
        jsonImport: false, //JSON导入弹窗
        cleanupFolders: false, //清理文件夹弹窗
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
          body_raw: '',
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
          body_raw: '',
        },
        moveApi: {
          api_id: '',
          name: '',
          folder_id: '',
          current_folder_id: '',
          current_collection_id: '',
          current_path: '',
        },
        jsonImport: {
          collection_id: '',
          json: '',
        },
        cleanupFolders: {
          days: 7,
          candidates: [],
          selectedIds: [],
          selectAll: false,
          checkedAt: 0,
        }
      },
      moveApiFolderOptions: [],
      moveApiFolderOptionsLoading: false,
      cleanupLoading: false,
      cleanupDeleting: false,
      // 弹窗保存防重入，避免回车和点击导致重复提交
      dialogSubmitting: {
        createCollection: false,
        createDir: false,
        createApi: false,
        copyApi: false,
        moveApi: false,
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
      // 过滤模式：仅显示指定文件夹及其下的接口（从 TaskWorkflow 跳转时启用）
      filterFolderId: 0,
      filterCollectionId: 0,
    }
  },
  mounted() {
    this.sidebarResizeMoveHandler = this.handleSidebarResize.bind(this)
    this.sidebarResizeUpHandler = this.stopSidebarResize.bind(this)
    // 初始化左侧面板宽度缓存
    this.loadSidebarWidthCache()
    // 读取路由参数：过滤模式（仅显示指定文件夹）
    const query = this.$route.query
    this.filterCollectionId = parseInt(query.collection_id || 0)
    this.filterFolderId = parseInt(query.filter_folder_id || 0)
    this.loadCollectionData()
    // 过滤模式下不加载归档
    if (!this.filterFolderId) {
      this.loadArchivedItems()
    }
    // 注册 API 数据变更 SSE 回调
    sseDistribute.RegisterReceive('api_data_change', this.handleApiChangeSSE)
  },
  beforeUnmount() {
    // 组件销毁时兜底移除拖拽事件
    this.stopSidebarResize()
    // 注销 API 数据变更 SSE 回调
    sseDistribute.UnRegisterReceive('api_data_change')
  },
  methods: {

    // ==================== SSE 数据变更回调 ====================

    handleApiChangeSSE(data) {
      if (!data || !data.change_type) {
        return
      }
      // 跳过自身触发的变更，避免重复刷新
      if (data.source_client_id && data.source_client_id === sseDistribute.GetSseClientId()) {
        return
      }
      const changeType = data.change_type
      const collectionId = data.collection_id
      const folderId = data.folder_id

      if (changeType === 'collection_created' || changeType === 'collection_updated' || changeType === 'collection_deleted') {
        this.loadCollectionData()
      } else if (changeType === 'folder_created' || changeType === 'folder_updated' || changeType === 'folder_deleted') {
        if (collectionId) {
          this.refreshCollectionFolders(collectionId)
        }
      } else if (changeType === 'api_created' || changeType === 'api_updated' || changeType === 'api_deleted' || changeType === 'api_weight_changed') {
        if (collectionId && folderId) {
          this.refreshFolderApis(collectionId, folderId)
        }
      } else if (changeType === 'api_moved') {
        // 刷新源文件夹和目标文件夹
        if (data.old_folder_id && collectionId) {
          this.refreshFolderApis(collectionId, data.old_folder_id)
        }
        if (folderId && collectionId) {
          this.refreshFolderApis(collectionId, folderId)
        }
      } else if (changeType === 'batch_imported') {
        if (collectionId) {
          this.refreshCollectionFolders(collectionId)
        }
      } else if (changeType === 'folder_archived') {
        if (collectionId) {
          this.refreshCollectionFolders(collectionId)
        }
        this.refreshArchiveNode()
      } else if (changeType === 'folder_restored') {
        if (collectionId) {
          this.refreshCollectionFolders(collectionId)
        }
        this.refreshArchiveNode()
      } else if (changeType === 'folder_permanent_deleted') {
        this.refreshArchiveNode()
      }
    },

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

      // 同步接口详情数据（已加载的接口 tab 需要手动触发 InitApiDetail，否则 Vue 复用组件导致内容不变）
      if (tab.type === 'api' && tab.loaded && tab.data) {
        await this.$nextTick()
        if (this.$refs.refApiDetail) {
          this.$refs.refApiDetail.InitApiDetail(tab.data)
        }
      }
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
    formatTimestamp(timestamp) {
      const value = Number(timestamp || 0)
      if (value <= 0) {
        return '未执行'
      }
      const date = new Date(value * 1000)
      if (Number.isNaN(date.getTime())) {
        return '未知'
      }
      const pad = (num) => String(num).padStart(2, '0')
      return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`
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
        env_id: source.env_id,
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
        // 过滤模式：仅保留指定文件夹
        if (this.filterFolderId > 0) {
          collectionNode.children = collectionNode.children.filter((folder) => Number(folder.id) === this.filterFolderId)
        }
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
    // 批量加载多个集合的文件夹（一次请求），用于恢复展开状态等场景
    async loadCollectionFoldersBatch(collectionNodes) {
      const pending = (collectionNodes || []).filter((node) => node && !node.loading && !node.loaded)
      if (pending.length === 0) {
        return
      }
      pending.forEach((node) => { node.loading = true })
      try {
        const collectionIds = pending.map((node) => node.id)
        const respData = await this.requestApi('CollectionFoldersBatchBasic', {
          collection_ids: collectionIds,
        })
        const grouped = respData.data || {}
        const sortCache = this.getTreeSortCache()
        for (const collectionNode of pending) {
          const folderList = grouped[String(collectionNode.id)] || []
          collectionNode.children = folderList.map((folder) => this.normalizeFolderNode(folder, collectionNode.id))
          if (this.filterFolderId > 0) {
            collectionNode.children = collectionNode.children.filter((folder) => Number(folder.id) === this.filterFolderId)
          }
          collectionNode.child_count = collectionNode.children.length
          collectionNode.isLeaf = collectionNode.child_count <= 0
          collectionNode.loaded = true
          this.sortListByIdOrder(collectionNode.children, sortCache.folders[String(collectionNode.id)] || [])
          this.syncTreeNodeChildren(collectionNode.uniqueid, collectionNode.children)
        }
      } catch (error) {
        this.$message.error(error.message || '批量加载集合文件夹失败')
      } finally {
        pending.forEach((node) => { node.loading = false })
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
    // 批量加载多个文件夹的接口（一次请求），用于恢复展开状态等场景
    async loadFolderApisBatch(folderNodes) {
      const pending = (folderNodes || []).filter((node) => node && !node.loading && !node.loaded)
      if (pending.length === 0) {
        return
      }
      pending.forEach((node) => { node.loading = true })
      try {
        const folderIds = pending.map((node) => node.id)
        const respData = await this.requestApi('FolderApisBatchBasic', {
          folder_ids: folderIds,
        })
        const grouped = respData.data || {}
        for (const folderNode of pending) {
          const apiList = grouped[String(folderNode.id)] || []
          folderNode.children = apiList.map((api) => this.normalizeApiNode(api, folderNode.id, folderNode.collection_id))
          folderNode.child_count = folderNode.children.length
          folderNode.isLeaf = folderNode.child_count <= 0
          folderNode.loaded = true
          this.applyFolderApiSort(folderNode.collection_id, folderNode.id)
          this.syncTreeNodeChildren(folderNode.uniqueid, folderNode.children)
        }
      } catch (error) {
        this.$message.error(error.message || '批量加载文件夹接口失败')
      } finally {
        pending.forEach((node) => { node.loading = false })
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
        // 查找文件夹的环境配置，用于接口无环境时继承
        const folderNode = this.findFolderNode(apiNode.collection_id, apiNode.folder_id)
        if (folderNode && folderNode.env_id) {
          detail.folder_env_id = folderNode.env_id
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
          let list = (res.Data.list || []).map((collection) => _that.normalizeCollectionNode(collection))
          // 过滤模式：仅保留指定集合
          if (_that.filterCollectionId > 0) {
            list = list.filter((collection) => Number(collection.id) === _that.filterCollectionId)
          }
          _that.treeData = list
          // 加载集合树后按本地缓存恢复排序
          _that.applyTreeSortCache()
          // 过滤模式下不追加归档节点
          if (!_that.filterFolderId) {
            // 末尾追加虚拟归档集合节点
            _that.pushArchiveNode(res.Data.archive_count || 0)
          }
          _that.initTreeExpansion()
          // 处理从任务清单跳转过来的初始导航
          _that.handleInitialNavigation()
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
    // 处理从任务清单跳转过来的初始导航（通过 query 参数定位集合/文件夹）
    async handleInitialNavigation() {
      const query = this.$route.query
      const collectionId = parseInt(query.collection_id || 0)
      if (collectionId <= 0) return

      await this.$nextTick()

      const collectionNode = this.findCollectionNode(collectionId)
      if (!collectionNode) return

      // 加载集合下的文件夹
      await this.ensureCollectionFoldersLoaded(collectionNode)

      const treeRef = this.$refs.collectionTreeRef
      if (!treeRef) return

      // 展开集合节点
      const collectionTreeNode = treeRef.getNode(collectionNode.uniqueid)
      if (collectionTreeNode) {
        collectionTreeNode.expand()
      }

      const folderId = parseInt(query.folder_id || 0)
      if (folderId > 0) {
        const folderNode = this.findFolderNode(collectionId, folderId)
        if (folderNode) {
          treeRef.setCurrentKey(folderNode.uniqueid)
          this.$nextTick(() => {
            setTimeout(() => {
              const el = treeRef.$el?.querySelector('.el-tree-node.is-current')
              if (el) {
                el.scrollIntoView({ block: 'center' })
              }
            }, 300)
          })
          this.openWorkspaceTab(folderNode, { reload: true })
          return
        }
      }
      treeRef.setCurrentKey(collectionNode.uniqueid)
      this.openWorkspaceTab(collectionNode, { reload: true })
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
      if (data.type === 'archive') {
        resolve(await this.loadArchiveFolders(data))
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
      // 筛选出需要恢复展开的集合节点
      const collectionsToExpand = []
      for (const collection of _that.treeData) {
        const collectionKey = _that.buildExpandStateKey(collection)
        if (collectionKey && expandedKeys.includes(collectionKey)) {
          collectionsToExpand.push(collection)
        }
      }
      // 一次批量请求加载所有集合的文件夹
      if (collectionsToExpand.length > 0) {
        await _that.loadCollectionFoldersBatch(collectionsToExpand)
        for (const collection of collectionsToExpand) {
          const collectionNode = _that.$refs.collectionTreeRef.getNode(collection.uniqueid)
          if (collectionNode) {
            collectionNode.expand()
          }
        }
      }
      // 收集所有需要恢复展开的文件夹节点
      const foldersToExpand = []
      for (const collection of _that.treeData) {
        if (!Array.isArray(collection.children)) {
          continue
        }
        for (const folder of collection.children) {
          const folderKey = _that.buildExpandStateKey(folder)
          if (folderKey && expandedKeys.includes(folderKey)) {
            foldersToExpand.push(folder)
          }
        }
      }
      // 一次批量请求加载所有文件夹的接口
      if (foldersToExpand.length > 0) {
        await _that.loadFolderApisBatch(foldersToExpand)
        for (const folder of foldersToExpand) {
          const folderNode = _that.$refs.collectionTreeRef.getNode(folder.uniqueid)
          if (folderNode) {
            folderNode.expand()
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
      if (dragData.type === 'archive' || dragData.is_archived) {
        return false
      }
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
      if (!data || (data.type !== 'collection' && data.type !== 'folder' && data.type !== 'archive')) {
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
        moveApi: () => this.moveApi(),
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
    copyApiHostAndToken() {
      const apiHost = Base.GetApiHost()
      const token = Base.GetSafeToken()
      const lines = []
      lines.push('API Host: ' + apiHost)
      if (token) {
        lines.push('Token: ' + token)
      }
      const text = lines.join('\n')
      this.copyText(text, 'API地址和Token已复制')
    },
    async buildMoveApiFolderOptions(api) {
      this.moveApiFolderOptionsLoading = true
      try {
        const groups = []
        for (const collection of this.treeData) {
          const folders = await this.ensureCollectionFoldersLoaded(collection)
          if (!Array.isArray(folders) || folders.length === 0) {
            continue
          }
          groups.push({
            collection_id: collection.id,
            collection_name: collection.name,
            folders: folders.map((folder) => ({
              id: folder.id,
              name: folder.name,
              collection_id: folder.collection_id,
              collection_name: collection.name,
              optionLabel: `${folder.name} (${collection.name})`,
              disabled: parseInt(folder.id) === parseInt(api.folder_id),
            })),
          })
        }
        this.moveApiFolderOptions = groups
        return groups
      } finally {
        this.moveApiFolderOptionsLoading = false
      }
    },
    findMoveApiTargetFolder(folderId) {
      for (const group of this.moveApiFolderOptions) {
        if (!Array.isArray(group.folders)) {
          continue
        }
        const targetFolder = group.folders.find((folder) => parseInt(folder.id) === parseInt(folderId))
        if (targetFolder) {
          return targetFolder
        }
      }
      return null
    },
    buildApiFolderPath(api) {
      const collectionNode = this.findCollectionNode(api.collection_id)
      const folderNode = this.findFolderNode(api.collection_id, api.folder_id)
      const collectionName = (collectionNode && collectionNode.name) || api.collection_name || api.collection_id
      const folderName = (folderNode && folderNode.name) || api.folder_name || api.folder_id
      return `${collectionName} / ${folderName}`
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
    escapeHtml(text) {
      return String(text ?? '')
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;')
    },
    confirmDeleteAction(entityType, entityName, onConfirm) {
      const safeType = this.escapeHtml(entityType)
      const safeName = this.escapeHtml(entityName)
      const message = `
        <div>
          <div>此操作不可恢复，请确认是否继续删除。</div>
          <div class="confirm-delete-message">
            ${safeType}：${safeName}
          </div>
        </div>
      `
      this.$confirm(message, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
        dangerouslyUseHTMLString: true,
      }).then(() => {
        onConfirm()
      }).catch(() => {
        this.$message.info('已取消删除')
      })
    },
    openCleanupDialog() {
      this.dialogShow.cleanupFolders = true
      if (this.dialogData.cleanupFolders.checkedAt === 0) {
        this.checkCleanupFolders()
      }
    },
    resetCleanupSelection() {
      this.dialogData.cleanupFolders.selectedIds = []
      this.dialogData.cleanupFolders.selectAll = false
    },
    toggleCleanupSelectAll(checked) {
      this.dialogData.cleanupFolders.selectedIds = checked
        ? this.dialogData.cleanupFolders.candidates.map((item) => item.id)
        : []
    },
    async checkCleanupFolders() {
      const days = Number(this.dialogData.cleanupFolders.days || 7)
      if (days <= 0) {
        this.$message.error('天数必须大于 0')
        return
      }
      this.cleanupLoading = true
      try {
        const data = await this.requestApi('CleanupCandidateFolders', { days })
        this.dialogData.cleanupFolders.candidates = Array.isArray(data.list) ? data.list : []
        this.dialogData.cleanupFolders.selectedIds = []
        this.dialogData.cleanupFolders.selectAll = false
        this.dialogData.cleanupFolders.checkedAt = Date.now()
      } catch (error) {
        this.$message.error(error.message || '检查失败')
      } finally {
        this.cleanupLoading = false
      }
    },
    confirmCleanupDelete() {
      const selectedFolders = this.dialogData.cleanupFolders.candidates.filter((item) =>
        this.dialogData.cleanupFolders.selectedIds.includes(item.id)
      )
      if (selectedFolders.length === 0) {
        this.$message.error('请选择要删除的文件夹')
        return
      }
      this.$confirm(`删除后进入回收站，确定归档选中的 ${selectedFolders.length} 个文件夹吗？`, '批量删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        this.executeCleanupDelete(selectedFolders)
      }).catch(() => {
        this.$message.info('已取消删除')
      })
    },
    async executeCleanupDelete(selectedFolders) {
      this.cleanupDeleting = true
      try {
        const folderIds = selectedFolders.map((item) => item.id)
        await this.requestApi('CleanupArchiveFolders', { folder_ids: folderIds })
        const collectionIds = [...new Set(selectedFolders.map((item) => Number(item.collection_id || 0)).filter((id) => id > 0))]
        selectedFolders.forEach((folder) => this.closeWorkspaceTabsByFolder(folder.id))
        await Promise.all(collectionIds.map((collectionId) => this.refreshCollectionFolders(collectionId)))
        await this.refreshArchiveNode()
        this.syncTreeSortCacheFromTree()
        this.$message.success('已移入回收站')
        await this.checkCleanupFolders()
      } catch (error) {
        this.$message.error(error.message || '删除失败')
      } finally {
        this.cleanupDeleting = false
      }
    },
    //删除集合
    handleCollectionDelete(collection) {
      let _that = this
      _that.confirmDeleteAction('集合', collection.name || collection.label || `#${collection.id}`, () => {
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
      let _that = this
      const safeName = _that.escapeHtml(folder.name || folder.label || `#${folder.id}`)
      const message = `
        <div>
          <div>此文件夹将移入归档，可随时恢复。</div>
          <div class="confirm-archive-message">
            文件夹：${safeName}
          </div>
        </div>
      `
      _that.$confirm(message, '移入归档', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
        dangerouslyUseHTMLString: true,
      }).then(() => {
        Api.DeleteDir(folder, function (res) {
          if (res.ErrCode === 0) {
            _that.closeWorkspaceTabsByFolder(folder.id)
            _that.refreshCollectionFolders(folder.collection_id).finally(() => {
              _that.syncTreeSortCacheFromTree()
              _that.refreshArchiveNode()
            })
            _that.$message.success('已移入归档')
          } else {
            _that.$message.error(res.ErrMsg)
          }
        })
      }).catch(() => {
        _that.$message.info('已取消')
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
      const createApiParams = { ..._that.dialogData.createApi }
      if (createApiParams.body_json && typeof createApiParams.body_json === 'object') {
        createApiParams.body_json = JSON.stringify(createApiParams.body_json)
      }
      Api.CreateApi(createApiParams, function (res) {
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
      const bodyJsonStr = api.body_json_data && typeof api.body_json_data === 'object'
        ? JSON.stringify(api.body_json_data)
        : api.body_json_data
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
        body_json: bodyJsonStr,
        body_raw: api.body_raw_data,
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
      } else if (command === 'move_api') {
        _that.openMoveApiDialog(data)
      } else if (command === 'delete_api') {
        _that.confirmDeleteAction('接口', data.name || data.label || `#${data.id}`, () => {
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
    async openCopyApiDialog(api) {
      let detail = null
      try {
        const detailResponse = await this.requestApi('ApisDetailByIds', { ids: [api.id] })
        detail = Array.isArray(detailResponse.list) ? detailResponse.list[0] : null
      } catch (error) {
        this.$message.warning(error.message || '加载接口详情失败，将尝试复制当前已加载内容')
      }
      const copySource = detail || api
      // 复制API数据到复制对话框
      this.dialogData.copyApi = JSON.parse(JSON.stringify(copySource))
      this.dialogData.copyApi.id = 0
      this.dialogData.copyApi.name = api.name + '-复制'
      this.dialogShow.copyApi = true
    },
    async openMoveApiDialog(api) {
      this.dialogData.moveApi = {
        api_id: api.id,
        name: api.name || '',
        folder_id: '',
        current_folder_id: api.folder_id,
        current_collection_id: api.collection_id,
        current_path: this.buildApiFolderPath(api),
      }
      this.dialogShow.moveApi = true
      try {
        const groups = await this.buildMoveApiFolderOptions(api)
        const hasAvailableTarget = groups.some((group) => Array.isArray(group.folders) && group.folders.some((folder) => !folder.disabled))
        if (!hasAvailableTarget) {
          this.$message.warning('当前没有可迁移的目标文件夹')
        }
      } catch (error) {
        this.$message.error(error.message || '加载目标文件夹失败')
      }
    },
    async moveApi() {
      let _that = this
      if (_that.beginDialogSubmit('moveApi')) {
        return
      }
      const moveForm = _that.dialogData.moveApi
      const targetFolder = _that.findMoveApiTargetFolder(moveForm.folder_id)
      if (!moveForm.api_id) {
        _that.endDialogSubmit('moveApi')
        _that.$message.error('请选择要迁移的接口')
        return
      }
      if (!targetFolder) {
        _that.endDialogSubmit('moveApi')
        _that.$message.error('请选择目标文件夹')
        return
      }
      try {
        await _that.requestApi('ApiMove', {
          api_id: moveForm.api_id,
          folder_id: moveForm.folder_id,
        })
        const sourceCollectionId = moveForm.current_collection_id
        const sourceFolderId = moveForm.current_folder_id
        const targetCollectionId = targetFolder.collection_id
        const targetFolderId = targetFolder.id
        const targetCollectionNode = _that.findCollectionNode(targetCollectionId)
        if (targetCollectionNode) {
          await _that.ensureCollectionFoldersLoaded(targetCollectionNode)
        }
        const refreshTasks = []
        if (parseInt(sourceCollectionId) !== parseInt(targetCollectionId) || parseInt(sourceFolderId) !== parseInt(targetFolderId)) {
          refreshTasks.push(_that.refreshFolderApis(sourceCollectionId, sourceFolderId))
          refreshTasks.push(_that.refreshFolderApis(targetCollectionId, targetFolderId))
        }
        await Promise.all(refreshTasks)
        const moveApiNode = _that.normalizeApiNode({
          ...(_that.getWorkspaceTabByKey(`api:${moveForm.api_id}`)?.data || {}),
          id: moveForm.api_id,
          name: moveForm.name,
          folder_id: targetFolderId,
          collection_id: targetCollectionId,
        }, targetFolderId, targetCollectionId)
        _that.upsertWorkspaceTabData(moveApiNode)
        const activeTab = _that.getWorkspaceTabByKey(`api:${moveForm.api_id}`)
        if (activeTab) {
          activeTab.loaded = false
        }
        _that.dialogShow.moveApi = false
        _that.syncTreeSortCacheFromTree()
        if (_that.activeTabKey === `api:${moveForm.api_id}`) {
          await _that.activateWorkspaceTab(`api:${moveForm.api_id}`, { reload: true })
        }
        _that.$message.success('迁移成功')
      } catch (error) {
        _that.$message.error(error.message || '迁移失败')
      } finally {
        _that.endDialogSubmit('moveApi')
      }
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
      const copyApiParams = { ..._that.dialogData.copyApi }
      if (copyApiParams.body_json && typeof copyApiParams.body_json === 'object') {
        copyApiParams.body_json = JSON.stringify(copyApiParams.body_json)
      }
      Api.CreateApi(copyApiParams, function (res) {
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
    },
    // 追加虚拟归档集合节点到树末尾
    pushArchiveNode(archiveCount) {
      let _that = this
      const archiveNode = {
        id: -1,
        uniqueid: 'archive:-1',
        type: 'archive',
        name: '归档',
        children: [],
        isLeaf: false,
        loaded: false,
        loading: false,
        child_count: archiveCount || 0,
      }
      _that.pushUniqueByKey(_that.treeData, archiveNode, 'uniqueid')
    },
    // 加载归档文件夹列表
    async loadArchiveFolders(archiveNode) {
      if (!archiveNode || archiveNode.loading) {
        return archiveNode.children || []
      }
      archiveNode.loading = true
      try {
        const data = await this.requestApi('ArchiveFolderList', {})
        archiveNode.children = (data.list || []).map((folder) => {
          return {
            ...this.normalizeFolderNode(folder, 0),
            is_archived: true,
            original_collection_id: folder.original_collection_id,
          }
        })
        archiveNode.child_count = archiveNode.children.length
        archiveNode.isLeaf = archiveNode.child_count <= 0
        archiveNode.loaded = true
        this.syncTreeNodeChildren(archiveNode.uniqueid, archiveNode.children)
        return archiveNode.children
      } finally {
        archiveNode.loading = false
      }
    },
    // 刷新归档节点
    async refreshArchiveNode() {
      const archiveNode = this.treeData.find((node) => node.type === 'archive')
      if (!archiveNode) {
        return
      }
      archiveNode.loaded = false
      archiveNode.children = []
      this.syncTreeNodeChildren(archiveNode.uniqueid, [])
      const treeRef = this.$refs.collectionTreeRef
      if (treeRef) {
        const treeNode = treeRef.getNode(archiveNode)
        if (treeNode && treeNode.expanded) {
          await this.loadArchiveFolders(archiveNode)
        }
      }
    },
    // 恢复归档文件夹
    handleRestoreFolder(folder) {
      let _that = this
      Api.RestoreFolder({ id: folder.id }, function (res) {
        if (res.ErrCode === 0) {
          _that.closeWorkspaceTabsByFolder(folder.id)
          _that.refreshArchiveNode()
          const collectionId = folder.original_collection_id
          if (collectionId) {
            _that.refreshCollectionFolders(collectionId)
          }
          _that.$message.success('恢复成功')
        } else {
          _that.$message.error(res.ErrMsg)
        }
      })
    },
    // 永久删除归档文件夹
    handlePermanentDeleteFolder(folder) {
      let _that = this
      _that.confirmDeleteAction('归档文件夹', folder.name || folder.label || `#${folder.id}`, () => {
        Api.PermanentDeleteDir({ id: folder.id }, function (res) {
          if (res.ErrCode === 0) {
            _that.closeWorkspaceTabsByFolder(folder.id)
            _that.refreshArchiveNode()
            _that.$message.success('已永久删除')
          } else {
            _that.$message.error(res.ErrMsg)
          }
        })
      })
    }
  }
}
</script>

<style scoped src="@/css/components/Api.css"></style>




