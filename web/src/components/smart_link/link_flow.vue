<template>
  <div class="smart-process-canvas-container">
    <div class="left-sidebar">
      <div class="search-box">
        <el-input
            v-model="state.searchQuery"
            clearable
            placeholder="搜索执行逻辑"
            @input="searchList"
        />
      </div>
      <div class="add-btn">
        <el-button class="process-action-btn process-action-btn--primary" @click="createNewProcess">新增执行逻辑</el-button>&nbsp;
        <el-link type="primary" @click="changeToLinks">切换到执行</el-link>
      </div>
      <div class="process-list">
        <el-scrollbar>
          <div
              v-for="process in state.filteredProcesses"
              :key="process.id"
              :class="{ active: state.activeProcess && state.activeProcess.id === process.id }"
              class="process-item"
              @click="selectProcess(process.id)"
          >
            <span>#{{ process.id }} {{ process.name }}</span>
            <el-popconfirm
                title="确定删除此执行逻辑吗？"
                @confirm="deleteProcess(process.id)"
            >
              <template #reference>
                <el-button
                    class="process-action-btn process-action-btn--danger process-action-btn--compact"
                    @click.stop
                >删除
                </el-button>
              </template>
            </el-popconfirm>
          </div>
        </el-scrollbar>
      </div>
    </div>

    <div class="right-content">
      <template v-if="state.activeProcess">
        <div class="canvas-header">
          <h2>{{ state.activeProcess.name }}</h2>
          <div class="header-actions">
            <el-button class="process-action-btn process-action-btn--ghost" @click="editProcessName">编辑</el-button>
            <el-button class="process-action-btn process-action-btn--primary" @click="addNewItem">新增执行逻辑子项</el-button>
            <el-button @click="resetView">重置视图</el-button>
            <el-button @click="fitView">适应画布</el-button>
            <el-button @click="zoomIn">放大</el-button>
            <el-button @click="zoomOut">缩小</el-button>
          </div>
        </div>

        <div class="canvas-wrapper">
          <VueFlow
              :nodes="nodes"
              :edges="edges"
              :default-viewport="{ zoom: 1 }"
              @node-drag-stop="onNodeDragStop"
              @connect="onConnect"
              @edges-delete="onEdgesDelete"
              @edge-double-click="onEdgeDoubleClick"
              class="vue-flow-canvas"
              :fit-view-on-init="true"
              :min-zoom="0.1"
              :max-zoom="3"
          >
            <!-- 自定义节点 -->
            <template #node-custom-node="nodeProps">
              <div
                  class="custom-node"
                  :class="{ selected: nodeProps.selected }"
                  @dblclick="editItem(nodeProps.data.item)"
              >
                <div class="node-header">
                  <div class="node-title">
                    <span class="node-id">#{{ nodeProps.data.item.id }}</span>
                    <span class="node-name">{{ nodeProps.data.item.name }}</span>
                    <span class="node-type">{{ nodeProps.data.item.type }}</span>
                  </div>
                  <div class="node-actions">
                    <el-button
                        class="process-action-btn process-action-btn--ghost process-action-btn--compact"
                        size="small"
                        @click.stop="editItem(nodeProps.data.item)"
                    >
                      编辑
                    </el-button>
                    <el-popconfirm
                        title="确定删除此执行逻辑子项吗？"
                        @confirm="deleteItem(nodeProps.data.item.id)"
                    >
                      <template #reference>
                        <el-button
                            class="process-action-btn process-action-btn--danger process-action-btn--compact"
                            size="small"
                            @click.stop
                        >删除
                        </el-button>
                      </template>
                    </el-popconfirm>
                  </div>
                </div>
                <div class="node-details">
                  <div v-if="nodeProps.data.item.locator !== ''" class="detail-item">
                    <span class="detail-label">定位:</span>
                    <span class="detail-value">{{ nodeProps.data.item.locator }}</span>
                  </div>
                  <div v-if="nodeProps.data.item.out_key !== ''" class="detail-item">
                    <span class="detail-label">输出值:</span>
                    <span class="detail-value">{{ nodeProps.data.item.out_key }}</span>
                  </div>
                  <div v-if="nodeProps.data.item.check_key !== ''" class="detail-item">
                    <span class="detail-label">判断:</span>
                    <span class="detail-value">{{ nodeProps.data.item.check_key }}</span>
                  </div>
                  <div v-if="nodeProps.data.item.value !== ''" class="detail-item">
                    <span class="detail-label">值:</span>
                    <span class="detail-value">{{ nodeProps.data.item.value }}</span>
                  </div>
                  <div v-if="nodeProps.data.item.out_key !== ''" class="detail-item">
                    <span class="detail-label">输出到替换列表:</span>
                    <span class="detail-value">{{ nodeProps.data.item.append_to_replace === '0' ? '否' : '是' }}</span>
                  </div>
                  <div v-if="parseInt(nodeProps.data.item.wait_mills) > 0" class="detail-item">
                    <span class="detail-label">等待时长:</span>
                    <span class="detail-value">{{ nodeProps.data.item.wait_mills }}ms</span>
                  </div>
                </div>
                <!-- 连接点 -->
                <Handle
                    type="source"
                    :position="Position.Right"
                    id="right"
                    style="background: #555; width: 10px; height: 10px; border-radius: 5px;"
                />
                <Handle
                    type="target"
                    :position="Position.Left"
                    id="left"
                    style="background: #555; width: 10px; height: 10px; border-radius: 5px;"
                />
              </div>
            </template>

            <!-- 背景网格 -->
            <Background :gap="15" :size="1" />

            <!-- 控制按钮 -->
            <Controls :show-interactive="false" />
          </VueFlow>
        </div>
      </template>
      <div v-else class="empty-tip">
        请选择或创建一个执行逻辑
      </div>
    </div>

    <!-- 编辑执行逻辑名称对话框 -->
    <el-dialog v-model="state.dialogProcessName" title="编辑执行逻辑名称" width="30%">
      <el-input v-model="state.editingProcessName"/>
      <template #footer>
        <el-button @click="state.dialogProcessName = false">取消</el-button>
        <el-button type="primary" @click="saveProcessName">保存</el-button>
      </template>
    </el-dialog>

    <!-- 编辑执行逻辑子项对话框 -->
    <el-dialog v-model="state.dialogProcessItem" :title="state.editingItem.id ? '编辑执行逻辑子项' : '新增执行逻辑子项'" width="70%">
      <ProcessItemEditor v-model="state.editingItem" :process-item-options="state.processItems" />
      <template #footer>
        <el-button @click="state.dialogProcessItem = false">取消</el-button>
        <el-button type="primary" @click="saveProcessItem">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { reactive, onMounted, ref, watch } from 'vue'
import { VueFlow, useVueFlow, Position, Handle } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { Controls } from '@vue-flow/controls'
import '@vue-flow/core/dist/style.css'
import '@vue-flow/core/dist/theme-default.css'
import '@vue-flow/controls/dist/style.css'
import { ElMessageBox } from 'element-plus'
import API from '@/utils/base/smart_link_proces'
import ProcessItemEditor from '@/components/smart_link/ProcessItemEditor.vue'

export default {
  components: {
    VueFlow,
    Background,
    Controls,
    Handle,
    ProcessItemEditor,
  },
  setup(props, { emit }) {
    const state = reactive({
      searchQuery: '',
      processes: [],
      filteredProcesses: [],
      activeProcess: null,
      processItems: [],
      dialogProcessName: false,
      editingProcessName: '',
      dialogProcessItem: false,
      editingItem: {
        id: 0,
        name: '',
        smart_link_process_id: 0,
        type: '',
        locator: '',
        wait_mills: 0,
        tip: '',
        value: '',
        out_key: '',
        check_key: '',
        weight: 0,
        domain_limit: '',
        append_to_replace: '0',
        is_async: '0',
        is_error_continue: '0',
        next_ids: 0,
      }
    })

    // Vue Flow 相关状态
    const { nodes, edges, setNodes, setEdges, fitView, zoomIn, zoomOut } = useVueFlow()
    const resetView = () => {
      setNodes(nodes.value.map(node => ({
        ...node,
        position: { x: node.position.x, y: node.position.y }
      })))
    }

    // 节点和边的转换函数
    const convertItemsToNodesAndEdges = (items) => {
      const newNodes = []
      const newEdges = []

      // 创建节点
      items.forEach(item => {
        newNodes.push({
          id: item.id.toString(),
          type: 'custom-node',
          position: {
            x: item.x || (item.weight * 200) || 100,
            y: item.y || (item.id * 100) || 100
          },
          data: {
            item: { ...item }, // ✅ 只传必要字段，别传整个 item
            label: `${item.name} (${item.type})`
          },
          draggable: true
        })
      })

      // 创建边（连接线）
      items.forEach(item => {
        if (item.next_ids && item.next_ids.toString().trim() !== '') {
          const nextIds = item.next_ids.toString().split(',').map(id => id.trim()).filter(id => id)
          nextIds.forEach(nextId => {
            if (nextId && items.some(i => i.id.toString() === nextId)) {
              newEdges.push({
                id: `e${item.id}-${nextId}`,
                source: item.id.toString(),
                target: nextId,
                type: 'default',
                animated: true,
                style: { stroke: '#555', strokeWidth: 2 }
              })
            }
          })
        }
      })

      setNodes(newNodes)
      setEdges(newEdges)
    }

    // Methods
    const fetchProcesses = function () {
      API.SmartProcessList(function (response) {
        if (response && response.Data) {
          state.processes = response.Data.list
          state.filteredProcesses = response.Data.list
          if (state.processes.length > 0) {
            state.activeProcess = state.processes[0]
            fetchProcessItems(state.activeProcess.id)
          }
        }
      })
    }

    const searchList = function () {
      if (state.searchQuery !== '') {
        state.filteredProcesses = state.processes.filter(process =>
            process.name.includes(state.searchQuery)
        )
        state.activeProcess = null
        setNodes([])
        setEdges([])
      } else {
        state.filteredProcesses = state.processes
      }

      if (state.filteredProcesses.length > 0) {
        setTimeout(() => {
          selectProcess(state.filteredProcesses[0].id)
        }, 0)
      }
    }

    const createNewProcess = function () {
      const newProcess = {
        id: 0,
        name: `新执行逻辑 ${state.processes.length + 1}`
      }
      API.SmartProcessAdd(newProcess, function (response) {
        newProcess.id = response.Data.id
        state.processes.unshift(newProcess)
        state.activeProcess = newProcess
        searchList()
      })
    }

    const ProcessSetRelation = function (prevId, nextId) {
      API.SmartProcessSetRelation({prev_id: prevId, next_id: nextId}, function (response) {
      })
    }

    const ProcessCancelRelation = function (prevId, nextId) {
      API.SmartProcessCancelRelation({prev_id: prevId, next_id: nextId}, function (response) {
      })
    }

    const ProcessSetPosition = function (id, x, y) {
      API.SmartProcessSetPosition({id: id, x: x, y: y}, function (response) {
      })
    }

    const selectProcess = function (id) {
      state.activeProcess = state.processes.find(p => p.id === id)
      fetchProcessItems(id)
    }

    const deleteProcess = function (id) {
      API.SmartProcessDelete({id}, function () {
        fetchProcesses()
      })
    }

    const editProcessName = function () {
      state.editingProcessName = state.activeProcess.name
      state.dialogProcessName = true
    }

    const saveProcessName = function () {
      state.activeProcess.name = state.editingProcessName
      API.SmartProcessAdd(state.activeProcess, function () {
        state.dialogProcessName = false
      })
    }

    const fetchProcessItems = function (processId) {
      API.SmartProcessItemList({smart_link_process_id: processId}, function (response) {
        if (response && response.Data) {
          state.processItems = response.Data.list
          convertItemsToNodesAndEdges(state.processItems)
          setTimeout(() => {
            fitView()
          }, 100)
        }
      })
    }

    const addNewItem = function () {
      // 计算新节点的位置（放在画布右侧）
      let maxX = 0
      let maxY = 0
      if (nodes.value.length > 0) {
        maxX = Math.max(...nodes.value.map(n => n.position.x))
        maxY = Math.max(...nodes.value.map(n => n.position.y))
      }

      state.editingItem = {
        id: 0,
        name: '',
        smart_link_process_id: state.activeProcess.id,
        type: '',
        locator: '',
        wait_mills: 3000,
        tip: '',
        value: '',
        out_key: '',
        check_key: '',
        weight: state.processItems.length > 0 ?
            Math.max(...state.processItems.map(i => i.weight)) + 1 : 0,
        domain_limit: '',
        append_to_replace: '0',
        is_async: '0',
        is_error_continue: '0',
        next_ids: 0,
        x: maxX + 200,
        y: maxY + 50
      }
      state.dialogProcessItem = true
    }

    const editItem = function (item) {
      state.editingItem = JSON.parse(JSON.stringify(item))
      state.dialogProcessItem = true
    }

    // 双击连线：弹出确认框
    const onEdgeDoubleClick = (edge) => {
      ElMessageBox.confirm('确定要取消这条连线吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
          .then(() => {
            ProcessCancelRelation(edge.edge.source, edge.edge.target)
            setEdges(edges.value.filter(e => e.id !== edge.edge.id))
          })
          .catch(() => {})
    }

    const saveProcessItem = function () {
      // 保存节点位置
      const currentNode = nodes.value.find(n => n.id === state.editingItem.id.toString())
      if (currentNode) {
        state.editingItem.x = currentNode.position.x
        state.editingItem.y = currentNode.position.y
      }

      API.SmartProcessItemAdd(state.editingItem, function () {
        state.dialogProcessItem = false
        fetchProcessItems(state.activeProcess.id)
      })
    }

    const deleteItem = function (id) {
      API.SmartProcessItemDelete({id}, function () {
        fetchProcessItems(state.activeProcess.id)
      })
    }

    // Vue Flow 事件处理
    const onNodeDragStop = (event) => {
      // 节点拖拽停止时更新位置
      const itemId = parseInt(event.node.id)
      ProcessSetPosition(itemId, event.node.position.x, event.node.position.y)
    }

    const onConnect = (params) => {
      const newEdge = {
        ...params,
        id: `e${params.source}-${params.target}`,
        type: 'default',
        animated: true,
        style: { stroke: '#555', strokeWidth: 2 }
      }

      setEdges([...edges.value, newEdge])

      // ✅ 调用后端接口建立关系
      ProcessSetRelation(params.source, params.target)
    }

    const onEdgesDelete = (connections) => {
      connections.forEach(connection => {
        const sourceId = connection.source
        const targetId = connection.target

        const sourceItem = state.processItems.find(item => item.id.toString() === sourceId)
        if (sourceItem) {
          let nextIds = sourceItem.next_ids ? sourceItem.next_ids.toString().split(',').map(id => id.trim()) : []
          nextIds = nextIds.filter(id => id !== targetId)
          sourceItem.next_ids = nextIds.join(',')

          API.SmartProcessItemAdd(sourceItem, () => {
            // ✅ 不再刷新整个图
          })
        }
      })
    }

    // 初始化
    onMounted(() => {
      fetchProcesses()
    })

    const changeToLinks = function () {
      emit('changeModelToLinks')
    }

    return {
      state,
      searchList,
      createNewProcess,
      selectProcess,
      deleteProcess,
      editProcessName,
      saveProcessName,
      addNewItem,
      editItem,
      saveProcessItem,
      deleteItem,
      changeToLinks,
      ProcessCancelRelation,
      ProcessSetRelation,
      ProcessSetPosition,
      nodes,
      edges,
      Position,
      fitView,
      zoomIn,
      zoomOut,
      resetView,
      onNodeDragStop,
      onConnect,
      onEdgesDelete,
      onEdgeDoubleClick,
    }
  }
}
</script>

<style scoped>
.smart-process-canvas-container {
  display: flex;
  height: 100vh;
  font-size: 14px;
}

.left-sidebar {
  width: 300px;
  border-right: 1px solid #e6e6e6;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.search-box {
  padding: 15px;
}

.add-btn {
  padding: 0 15px 15px;
}

.process-list {
  flex: 1;
  overflow: hidden;
}

.process-item {
  padding: 12px 15px;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.process-item:hover {
  background-color: #f5f5f5;
}

.process-item.active {
  background-color: #e6f7ff;
}

.process-action-btn {
  border-radius: 999px;
  font-weight: 600;
  transition: all 0.2s ease;
}

.process-action-btn--compact {
  padding: 6px 12px;
}

.process-action-btn--primary {
  color: #2f6b45;
  border-color: #b8d7c1;
  background: linear-gradient(180deg, #f4fbf6 0%, #e8f5ec 100%);
}

.process-action-btn--primary:hover {
  color: #255639;
  border-color: #97c2a5;
  background: linear-gradient(180deg, #edf8f0 0%, #dceee2 100%);
}

.process-action-btn--ghost {
  color: #315f8d;
  border-color: #c9d9eb;
  background: #f3f8fd;
}

.process-action-btn--ghost:hover {
  color: #274c72;
  border-color: #adc6df;
  background: #e8f1fb;
}

.process-action-btn--danger {
  color: #b44343;
  border-color: #efc4c4;
  background: #fff4f4;
}

.process-action-btn--danger:hover {
  color: #963737;
  border-color: #e5a7a7;
  background: #feeaea;
}

.right-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.canvas-header {
  display: flex;
  align-items: center;
  padding: 15px 20px;
  border-bottom: 1px solid #e6e6e6;
  background-color: #f8f9fa;
}

.canvas-header h2 {
  margin: 0;
  flex: 1;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.canvas-wrapper {
  flex: 1;
  position: relative;
  overflow: hidden;
}

.vue-flow-canvas {
  width: 100%;
  height: 100%;
}

/* 自定义节点样式 */
.custom-node {
  background: white;
  border: 2px solid #ddd;
  border-radius: 8px;
  padding: 12px;
  min-width: 250px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  transition: all 0.2s ease;
  max-width : 150px;
}

.custom-node.selected {
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.3);
}

.node-header {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid #eee;
}

.node-title {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.node-id {
  font-weight: bold;
  color: #666;
  font-size: 12px;
}

.node-name {
  font-weight: bold;
  margin-right: 8px;
}

.node-type {
  background: #f0f2f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  color: #555;
}

.node-actions {
  display: flex;
  gap: 5px;
}

.node-details {
  font-size: 12px;
  color: #666;
}

.detail-item {
  margin: 4px 0;
  display: flex;
  align-items: flex-start;
}

.detail-label {
  font-weight: bold;
  min-width: 80px;
  color: #333;
}

.detail-value {
  flex: 1;
  word-break: break-all;
}

.start-node-indicator {
  margin-top: 8px;
  text-align: right;
}

.empty-tip {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #999;
  font-size: 14px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .smart-process-canvas-container {
    flex-direction: column;
  }

  .left-sidebar {
    width: 100%;
    height: auto;
    max-height: 300px;
  }

  .canvas-header {
    flex-wrap: wrap;
  }

  .header-actions {
    margin-top: 10px;
    width: 100%;
    justify-content: space-between;
  }
}
</style>
