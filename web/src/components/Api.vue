<template>
  <div class="collection-container">
    <!-- 左侧集合列表 -->
    <div class="left-sidebar">
      <!-- 集合列表区域 -->
      <div class="collection-section">
        <div class="section-header">
          <span>集合列表</span>
          <div>
            <el-button link type="primary" @click="createNewCollection">新建集合</el-button>
            <el-button link type="primary" @click="drawerVisibleMarkdown = true">文档</el-button>
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
                node-key="uniqueid"
                tabindex="0" @keyup="handleKeyUp"
                @node-click="handleNodeClick"
                @node-expand="handleNodeExpand"
                @node-collapse="handleNodeCollapse"
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

  <el-dialog v-model="dialogShow.createCollection" title="创建集合" width="500">
    <el-form>
      <el-form-item :label-width="80" label="集合名称">
        <el-input v-model="dialogData.createCollection.name" autocomplete="off" @keydown="keydownCreateNewCollection"/>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogShow.createCollection = false">取消</el-button>
        <el-button type="primary" @click="createNewCollection">保存</el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="dialogShow.createDir" title="创建文件夹" width="500" tabindex="0" @keydown="keydownNewDir">
    <el-form>
      <el-form-item :label-width="80" label="文件夹名称">
        <el-input v-model="dialogData.createDir.name" autocomplete="off" @keydown="keydownNewDir"/>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogShow.createDir = false">取消</el-button>
        <el-button type="primary" @click="createNewDir">保存</el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="dialogShow.createApi" title="创建接口" width="700" tabindex="0" @keydown="keywodnCreateApi">
    <el-tabs v-model="createApiType" tabindex="0" @keydown="keywodnCreateApi">
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

  <el-dialog v-model="dialogShow.copyApi" title="复制接口" width="500">
    <el-form>
      <el-form-item :label-width="80" label="接口名称">
        <el-input v-model="dialogData.copyApi.name" autocomplete="off" placeholder="请输入新的接口名称" @keyup="copyApiKeyup"/>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogShow.copyApi = false">取消</el-button>
        <el-button type="primary" @click="copyApi">保存</el-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="dialogShow.jsonImport" title="通过JSON导入" width="800">
    <el-form :model="dialogData.jsonImport" label-width="120px">
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
import {FolderOpened, Folder, Document, ArrowDown, ArrowUp, More} from '@element-plus/icons-vue'
import CollectionBasicInfo from './api/CollectionBasicInfo'
import CollectionEnvironment from './api/CollectionEnvironment'
import CollectionPermission from './api/CollectionPermission'
import FolderDetail from './api/FolderDetail'
import ApiDetail from './api/ApiDetail'
import Markdown from '@/components/Markdown.vue'
import Api from '@/utils/base/api'
import Array from '@/utils/base/array'
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
      keyup: null,
    }
  },
  mounted() {
    this.loadCollectionData()
    this.loadArchivedItems()
  },
  methods: {


    // 加载集合数据
    loadCollectionData() {
      let _that = this
      Api.Collections({}, function (res) {
        if (res.ErrCode === 0) {
          _that.treeData = res.Data.list
          _that.initTreeExpansion()
        } else {
          _that.$message.error(res.ErrMsg)
        }
      })
    },

    // 初始化树展开状态
    initTreeExpansion() {
      let _that = this
      let expandedKeys = store.getStore('collection_expanded_keys')
      if (expandedKeys) {
        try {
          expandedKeys = JSON.parse(expandedKeys)
          _that.$nextTick(() => {
            expandedKeys.forEach(key => {
              const node = _that.$refs.collectionTreeRef.getNode(key)
              if (node) {
                node.expand()
              }
            })
          })
        } catch (e) {
          _that.expandAllNodes()
        }
      } else {
        _that.expandAllNodes()
      }
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
      })
    },

    // 处理节点展开
    handleNodeExpand(data) {
      let _that = this
      _that.updateExpandedCache(data.uniqueid, true)
    },

    // 处理节点折叠
    handleNodeCollapse(data) {
      let _that = this
      _that.updateExpandedCache(data.uniqueid, false)
    },

    // 更新展开缓存
    updateExpandedCache(nodeKey, isExpanded) {
      let _that = this
      let expandedKeys = store.getStore('collection_expanded_keys')
      if (expandedKeys) {
        try {
          expandedKeys = JSON.parse(expandedKeys)
        } catch (e) {
          expandedKeys = []
        }
      } else {
        expandedKeys = []
      }

      if (isExpanded) {
        if (!expandedKeys.includes(nodeKey)) {
          expandedKeys.push(nodeKey)
        }
      } else {
        expandedKeys = expandedKeys.filter(key => key !== nodeKey)
      }

      store.setStore('collection_expanded_keys', JSON.stringify(expandedKeys))
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
      })
    },
    // 处理归档项点击
    handleArchiveItemClick(item) {
      this.selectedItem = item
    },
    keydownCreateNewCollection(e){
      if(e.key === 'Enter'){
        this.createNewCollection()
      }
    },
    // 创建新集合
    createNewCollection() {
      let _that = this
      if (!_that.dialogShow.createCollection) {
        _that.dialogShow.createCollection = true
        return
      }
      Api.CreateCollection(_that.dialogData.createCollection, function (res) {
        if (res.ErrCode === 0) {
          _that.dialogShow.createCollection = false
          let newCollection = res.Data
          newCollection.children = []
          _that.treeData.push(newCollection)
        } else {
          _that.$message.error(res.ErrMsg)
        }
      })
    },
    keydownNewDir(e){
      console.log(e)
      if(e.key === 'Enter'){
        this.createNewDir()
      }
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
      _that.dialogData.createDir.collection_id = _that.selectedItem.id
      Api.CreateDir(_that.dialogData.createDir, function (res) {
        if (res.ErrCode === 0) {
          _that.dialogShow.createDir = false
          let newDir = res.Data
          newDir.children = []
          for (let i in _that.treeData) {
            if (parseInt(_that.dialogData.createDir.collection_id) === parseInt(_that.treeData[i].id)) {
              _that.treeData[i].children.push(newDir)
            }
          }
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
            _that.treeData = Array.DeleteValueByStringKey(_that.treeData, 'uniqueid', collection.uniqueid)
            //如果是删除的集合
            if (_that.selectedItem && _that.selectedItem.type === 'collection') {
              if (_that.selectedItem.id === collection.id) {
                _that.selectedItem = {}
              }
            }
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
              _that.treeData[i].children = Array.DeleteValueByStringKey(_that.treeData[i].children, 'uniqueid', folder.uniqueid)
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
            _that.$message.success('删除成功')
          } else {
            _that.$message.error(res.ErrMsg)
          }
        })
      }).catch(() => {
        _that.$message.info('已取消删除')
      })
    },
    keywodnCreateApi(e){
      if(e.key === 'Enter'){
        this.handleFolderCreateApi()
      }
    },
    handleFolderCreateApi: function () {
      let _that = this
      if (!_that.dialogShow.createApi) {
        _that.dialogShow.createApi = true
        _that.dialogData.createApi.curlData = ''
        return
      }
      _that.dialogData.createApi.folder_id = _that.selectedItem.id
      _that.dialogData.createApi.collection_id = _that.selectedItem.collection_id
      Api.CreateApi(_that.dialogData.createApi, function (res) {
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
            _that.treeData[i].children[j].children.push(newApi)
          }
        }
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
                  _that.treeData[i].children[j].children = Array.DeleteValueByStringKey(folderInfo.children, 'uniqueid', data.uniqueid)
                }
              }
              if (_that.selectedItem.uniqueid === data.uniqueid) {//如果当前删除的api是选中的api 那么置空当前选项
                _that.selectedItem = {}
              }
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
      if (!this.dialogData.jsonImport.collection_id) {
        _that.$message.error('请选择集合')
        return
      }
      if (!this.dialogData.jsonImport.json) {
        _that.$message.error('请输入JSON数据')
        return
      }
      try {
        // Validate JSON format
        JSON.parse(this.dialogData.jsonImport.json)
      } catch (e) {
        _that.$message.error('JSON格式错误，请检查输入')
        return
      }
      Api.ApiImportJson({
        collection_id: this.dialogData.jsonImport.collection_id,
        json: this.dialogData.jsonImport.json
      }, function (res) {
        if (res.ErrCode === 0) {
          _that.$message.success('导入成功')
          _that.dialogShow.jsonImport = false
          _that.loadCollectionData()
        } else {
          _that.$message.error(res.ErrMsg)
        }
      })
    },
    copyApiKeyup: function (event) {
      let _that = this
      if (event.key === 'Enter') {
        _that.copyApi()
        event.preventDefault()
      }
    },
    // 复制接口
    copyApi() {
      let _that = this
      Api.CreateApi(_that.dialogData.copyApi, function (res) {
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
            _that.treeData[i].children[j].children.push(newApi)
          }
        }
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
  height: 100vh;
  background-color: #f5f7fa;
}

/* 左侧样式 */
.left-sidebar {
  background: white;
  border-right: 1px solid #e4e7ed;
  flex-direction: column;
  margin-left: -20px;
  width: 20%;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
  font-weight: 600;
  cursor: pointer;
  user-select: none;
}

.collection-list {
  flex: 1;
  overflow-y: auto;
  overflow-x: auto; /* 添加横向滚动条 */
  padding: 8px;
  padding-bottom: 0; /* 避免底部被滚动条遮挡 */
}

/* 为el-tree添加一个包装容器来实现横向滚动 */
.tree-wrapper {
  min-width: 100%; /* 改为100%以适应容器 */
  display: block; /* 改为block以适应容器 */
}

/* 为el-tree容器添加横向滚动支持 */
.collection-list .el-tree {
  min-width: 100%;
  white-space: nowrap; /* 防止内容换行 */
}

/* 树节点样式 */
.tree-node {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 4px 0;
  white-space: nowrap; /* 确保内容不换行 */
  position: relative;
}

.node-icon {
  margin-right: 6px;
  color: #409eff;
  flex-shrink: 0; /* 防止图标被压缩 */
}

.node-label {
  flex: 1;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis; /* 超出显示省略号 */
  min-width: 0; /* 允许收缩 */
  cursor: default; /* 显示为默认光标，表示可悬停查看 */
  max-width: calc(100% - 40px); /* 确保为右侧操作按钮预留足够空间 */
  margin-right: 10px; /* 与操作按钮之间的小间隙 */
}

.node-actions {
  opacity: 0;
  transition: opacity 0.3s;
  white-space: nowrap; /* 确保按钮不换行 */
  position: absolute;
  right: 0;
  background: white; /* 确保背景色覆盖其他内容 */
  z-index: 1; /* 确保在最上层 */
  padding-left: 10px; /* 与标签内容分离 */
  padding-right: 5px; /* 右侧留白 */
}

.tree-node:hover .node-actions {
  opacity: 1;
}

/* 修正el-tree节点的宽度问题 */
.collection-list .el-tree-node__content {
  position: relative;
  height: auto;
  min-height: 32px;
  line-height: 24px;
  padding-right: 0; /* 移除默认的右边距 */
  white-space: nowrap;
  overflow: visible; /* 确保操作按钮可见 */
}

/* 归档列表区域 */
.archive-section {
  border-top: 1px solid #e4e7ed;
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
  background-color: #f5f7fa;
}

.archive-item .el-icon {
  margin-right: 8px;
  color: #909399;
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
  background: white;
  width: 80%;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #e4e7ed;
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
  color: #409eff;
  margin-right: 12px;
}

.info-content {
  display: flex;
  flex-direction: column;
}

.info-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.info-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.no-selection {
  color: #909399;
  font-style: italic;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.panel-content {
  flex: 1;
  padding-left: 20px;
  padding-top: -15px;
  overflow-y: auto;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}

</style>



