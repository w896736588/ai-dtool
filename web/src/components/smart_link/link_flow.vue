<template>
  <div class="smart-process-canvas-container">
    <div class="left-sidebar">
      <div class="search-box">
        <el-input
            v-model="state.searchQuery"
            clearable
            placeholder="鎼滅储鎵ц閫昏緫"
            @input="searchList"
        />
      </div>
      <div class="add-btn">
        <GitActionButton @click="createNewProcess">鏂板鎵ц閫昏緫</GitActionButton>&nbsp;
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
                title="纭畾鍒犻櫎姝ゆ墽琛岄€昏緫鍚楋紵"
                @confirm="deleteProcess(process.id)"
            >
              <template #reference>
                <GitActionButton
                    compact
                    variant="danger"
                    @click.stop
                >鍒犻櫎
                </GitActionButton>
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
            <GitActionButton variant="info" @click="editProcessName">缂栬緫</GitActionButton>
            <GitActionButton @click="addNewItem">鏂板鎵ц閫昏緫瀛愰」</GitActionButton>
            <GitActionButton variant="info" @click="resetView">閲嶇疆瑙嗗浘</GitActionButton>
            <GitActionButton variant="info" @click="fitView">閫傚簲鐢诲竷</GitActionButton>
            <GitActionButton variant="info" @click="zoomIn">鏀惧ぇ</GitActionButton>
            <GitActionButton variant="info" @click="zoomOut">缂╁皬</GitActionButton>
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
            <!-- 鑷畾涔夎妭鐐?-->
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
                    <GitActionButton
                        compact
                        variant="info"
                        size="small"
                        @click.stop="editItem(nodeProps.data.item)"
                    >
                      缂栬緫
                    </GitActionButton>
                    <el-popconfirm
                        title="纭畾鍒犻櫎姝ゆ墽琛岄€昏緫瀛愰」鍚楋紵"
                        @confirm="deleteItem(nodeProps.data.item.id)"
                    >
                      <template #reference>
                        <GitActionButton
                            compact
                            variant="danger"
                            size="small"
                            @click.stop
                        >鍒犻櫎
                        </GitActionButton>
                      </template>
                    </el-popconfirm>
                  </div>
                </div>
                <div class="node-details">
                  <div v-if="nodeProps.data.item.locator !== ''" class="detail-item">
                    <span class="detail-label">瀹氫綅:</span>
                    <span class="detail-value">{{ nodeProps.data.item.locator }}</span>
                  </div>
                  <div v-if="nodeProps.data.item.out_key !== ''" class="detail-item">
                    <span class="detail-label">杈撳嚭鍊?</span>
                    <span class="detail-value">{{ nodeProps.data.item.out_key }}</span>
                  </div>
                  <div v-if="nodeProps.data.item.check_key !== ''" class="detail-item">
                    <span class="detail-label">鍒ゆ柇:</span>
                    <span class="detail-value">{{ nodeProps.data.item.check_key }}</span>
                  </div>
                  <div v-if="nodeProps.data.item.value !== ''" class="detail-item">
                    <span class="detail-label">鍊?</span>
                    <span class="detail-value">{{ nodeProps.data.item.value }}</span>
                  </div>
                  <div v-if="nodeProps.data.item.out_key !== ''" class="detail-item">
                    <span class="detail-label">输出到替换列表:</span>
                    <span class="detail-value">{{ nodeProps.data.item.append_to_replace === '0' ? '否' : '是' }}</span>
                  </div>
                  <div v-if="parseInt(nodeProps.data.item.wait_mills) > 0" class="detail-item">
                    <span class="detail-label">绛夊緟鏃堕暱:</span>
                    <span class="detail-value">{{ nodeProps.data.item.wait_mills }}ms</span>
                  </div>
                </div>
                <!-- 杩炴帴鐐?-->
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

            <!-- 鑳屾櫙缃戞牸 -->
            <Background :gap="15" :size="1" />

            <!-- 鎺у埗鎸夐挳 -->
            <Controls :show-interactive="false" />
          </VueFlow>
        </div>
      </template>
      <div v-else class="empty-tip">
        璇烽€夋嫨鎴栧垱寤轰竴涓墽琛岄€昏緫
      </div>
    </div>

    <!-- 缂栬緫鎵ц閫昏緫鍚嶇О瀵硅瘽妗?-->
    <el-dialog v-model="state.dialogProcessName" title="缂栬緫鎵ц閫昏緫鍚嶇О" width="30%">
      <el-input v-model="state.editingProcessName"/>
      <template #footer>
        <GitActionButton @click="state.dialogProcessName = false">鍙栨秷</GitActionButton>
        <GitActionButton @click="saveProcessName">淇濆瓨</GitActionButton>
      </template>
    </el-dialog>

    <!-- 缂栬緫鎵ц閫昏緫瀛愰」瀵硅瘽妗?-->
    <el-dialog v-model="state.dialogProcessItem" :title="state.editingItem.id ? `编辑执行逻辑子项 #${state.editingItem.id}` : '新增执行逻辑子项'" width="70%">
      <ProcessItemEditor ref="processItemEditorRef" v-model="state.editingItem" :process-item-options="state.processItems" />
      <template #footer>
        <GitActionButton @click="state.dialogProcessItem = false">鍙栨秷</GitActionButton>
        <GitActionButton @click="saveProcessItem">淇濆瓨</GitActionButton>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import API from '@/utils/base/smart_link_proces'
import ProcessItemEditor from '@/components/smart_link/ProcessItemEditor.vue'
import GitActionButton from '@/components/base/GitActionButton.vue'

export default {
  components: {
    VueFlow,
    Background,
    Controls,
    Handle,
    ProcessItemEditor,
    GitActionButton,
  },
  setup(props, { emit }) {
    const processItemEditorRef = ref(null)
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

    // Vue Flow 鐩稿叧鐘舵€?
    const { nodes, edges, setNodes, setEdges, fitView, zoomIn, zoomOut } = useVueFlow()
    const resetView = () => {
      setNodes(nodes.value.map(node => ({
        ...node,
        position: { x: node.position.x, y: node.position.y }
      })))
    }

    // 鑺傜偣鍜岃竟鐨勮浆鎹㈠嚱鏁?
    const convertItemsToNodesAndEdges = (items) => {
      const newNodes = []
      const newEdges = []

      // 鍒涘缓鑺傜偣
      items.forEach(item => {
        newNodes.push({
          id: item.id.toString(),
          type: 'custom-node',
          position: {
            x: item.x || (item.weight * 200) || 100,
            y: item.y || (item.id * 100) || 100
          },
          data: {
            item: { ...item }, // 鉁?鍙紶蹇呰瀛楁锛屽埆浼犳暣涓?item
            label: `${item.name} (${item.type})`
          },
          draggable: true
        })
      })

      // 鍒涘缓杈癸紙杩炴帴绾匡級
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
        name: `鏂版墽琛岄€昏緫 ${state.processes.length + 1}`
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
      // 璁＄畻鏂拌妭鐐圭殑浣嶇疆锛堟斁鍦ㄧ敾甯冨彸渚э級
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

    // 鍙屽嚮杩炵嚎锛氬脊鍑虹‘璁ゆ
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
      const isValid = processItemEditorRef.value ? processItemEditorRef.value.validateForSave() : true
      if (!isValid) {
        ElMessage.error('请先修正表单中的格式问题，再保存流程项。')
        return
      }
      // 保存节点位置 / Persist current node position before save.
      const currentNode = nodes.value.find(n => n.id === state.editingItem.id.toString())
      if (currentNode) {
        state.editingItem.x = currentNode.position.x
        state.editingItem.y = currentNode.position.y
      }
      state.editingItem.smart_link_process_id = state.activeProcess.id
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

    // Vue Flow 浜嬩欢澶勭悊
    const onNodeDragStop = (event) => {
      // 鑺傜偣鎷栨嫿鍋滄鏃舵洿鏂颁綅缃?
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

      // 鉁?璋冪敤鍚庣鎺ュ彛寤虹珛鍏崇郴
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
            // 鉁?涓嶅啀鍒锋柊鏁翠釜鍥?
          })
        }
      })
    }

    // 鍒濆鍖?
    onMounted(() => {
      fetchProcesses()
    })

    const changeToLinks = function () {
      emit('changeModelToLinks')
    }

    return {
      state,
      processItemEditorRef,
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

<style scoped src="@/css/components/smart_link/link_flow.css"></style>
