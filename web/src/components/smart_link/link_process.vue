<template>
  <div class="smart-process-container">
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
        <div class="process-header">
          <h2>{{ state.activeProcess.name }}</h2>&nbsp;
          <el-button class="process-action-btn process-action-btn--ghost" @click="editProcessName">编辑</el-button>
          <el-button class="process-action-btn process-action-btn--primary" @click="addNewItem">新增执行逻辑子项</el-button>
        </div>
        <div class="process-items-wrapper">
          <el-scrollbar class="process-items-scroll">
            <draggable
                v-model="state.processItems"
                handle=".drag-handle"
                item-key="id"
                @end="handleSortEnd"
            >
              <template #item="{ element }">
                <div class="process-item-card">
                  <div class="item-header">
                    <el-icon class="drag-handle">
                      <Menu/>
                    </el-icon>
                    <span>#{{ element.id }} {{ element.name }}  {{ element.type }}</span>
                    <div class="item-actions">
                      <el-button class="process-action-btn process-action-btn--primary process-action-btn--compact" @click="addNewItem(element)">新增复制</el-button>
                      <el-button class="process-action-btn process-action-btn--ghost process-action-btn--compact" @click="editItem(element)">编辑</el-button>
                      <el-popconfirm
                          title="确定删除此执行逻辑子项吗？"
                          @confirm="deleteItem(element.id)"
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
                  </div>
                  <div class="item-details">
                    <div>
                      <span v-if="element.locator !== ''"> 定位: {{ element.locator }}</span>
                      <span v-if="element.out_key !== ''"> 输出值: {{ element.out_key }}</span>
                      <span v-if="element.check_key !== ''"> 判断: {{ element.check_key }}</span>
                      <span v-if="element.out_key !== ''"> 输出到替换列表: {{ element.append_to_replace }}</span>
                      <span v-if="parseInt(element.wait_mills) > 0"> 等待时长: {{ element.wait_mills }}ms</span>
                    </div>
                  </div>
                </div>
              </template>
            </draggable>
          </el-scrollbar>
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
import {reactive, onMounted} from 'vue'
import draggable from 'vuedraggable'
import {Menu} from '@element-plus/icons-vue'
import API from '@/utils/base/smart_link_proces'
import ProcessItemEditor from '@/components/smart_link/ProcessItemEditor.vue'

export default {
  components: {
    draggable,
    Menu,
    ProcessItemEditor,
  },
  setup(props, {emit}) {
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
        append_to_replace : '0', //1需要加入到替换列表
        is_async : '0',
        is_error_continue : '0', //遇到错误是否继续
        next_ids : 0, //下一个节点的id,多个用逗号分割
        is_start : 0, //是否为开始节点，1是
      }
    })

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
        // 搜索时清空右侧内容
        state.activeProcess = null
        state.processItems = []
      } else {
        state.filteredProcesses = state.processes
      }

      // 搜索完成后如果有匹配项，默认显示第一个
      if (state.filteredProcesses.length > 0) {
        // 使用setTimeout确保DOM更新完成后再触发
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

    //关联节点
    const ProcessSetRelation = function (prevId , nextId) {
      API.SmartProcessSetRelation({prev_id: prevId, next_id: nextId}, function (response) {
      })
    }

    //取消关联节点
    const ProcessCancelRelation = function (prevId , nextId) {
      API.SmartProcessCancelRelation({prev_id: prevId, next_id: nextId}, function (response) {
      })
    }

    //设置节点位置
    const ProcessSetPosition = function (id , x , y) {
      API.SmartProcessSetPosition({id: id, x:x,y:y}, function (response) {
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
        }
      })
    }

    const addNewItem = function (copyItem) {
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
        append_to_replace : '0', //1需要加入到替换列表
        is_async : '0',
        is_error_continue : '0', //遇到错误是否继续
        next_ids : 0, //下一个节点的id,多个用逗号分割
        is_start : 0, //是否为开始节点，1是
      }
      if (copyItem){
        state.editingItem.name = copyItem.name + '-复制'
        state.editingItem.smart_link_process_id = copyItem.smart_link_process_id
        state.editingItem.type = copyItem.type
        state.editingItem.locator = copyItem.locator
        state.editingItem.wait_mills = copyItem.wait_mills
        state.editingItem.tip = copyItem.tip
        state.editingItem.value = copyItem.value
        state.editingItem.out_key = copyItem.out_key
        state.editingItem.check_key = copyItem.check_key
        state.editingItem.domain_limit = copyItem.domain_limit
        state.editingItem.append_to_replace = copyItem.append_to_replace
        state.editingItem.is_async = copyItem.is_async
        state.editingItem.is_error_continue = copyItem.is_error_continue
        state.editingItem.next_ids = copyItem.next_ids
        state.editingItem.is_start = copyItem.is_start
      }
      state.dialogProcessItem = true
    }

    const editItem = function (item) {
      state.editingItem = JSON.parse(JSON.stringify(item))
      state.dialogProcessItem = true
    }

    const saveProcessItem = function () {
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

    const handleSortEnd = function () {
      const itemIds = state.processItems.map(item => item.id).join(',')
      API.SmartProcessItemSort({
        smart_link_process_id: state.activeProcess.id,
        smart_link_process_item_ids: itemIds
      }, function () {
        // 可以在这里添加排序成功后的回调
      })
    }

    // Initialize
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
      handleSortEnd,
      changeToLinks,
      ProcessCancelRelation,
      ProcessSetRelation,
      ProcessSetPosition,
    }
  }
}
</script>

<style scoped>
.smart-process-container {
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
  padding: 20px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.process-header {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
  flex-shrink: 0;
}

.add-item-btn {
  flex-shrink: 0;
  margin-bottom: 15px;
}

.process-items-wrapper {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.process-items-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}

.process-item-card {
  border: 1px solid #e6e6e6;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 10px;
  background-color: #fff;
}

.item-header {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

.item-header .drag-handle {
  margin-right: 10px;
  cursor: move;
}

.item-header span {
  flex: 1;
  font-weight: bold;
}

.item-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.item-details {
  display: flex;
  flex-direction: column; /* 改为垂直排列 */
  gap: 8px; /* 设置行间距 */
  margin-top: 8px;
}

.item-details div {
  width: 100%; /* 每项占满整行 */
  white-space: normal; /* 允许换行 */
  overflow: visible; /* 显示全部内容 */
  text-overflow: clip; /* 不使用省略号 */
  word-break: break-all; /* 长单词或URL可以换行 */
}

.add-item-btn {
  margin-top: 15px;
}

.empty-tip {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #999;
  font-size: 14px;
}
</style>
