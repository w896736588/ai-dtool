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
        <GitActionButton @click="createNewProcess">新增执行逻辑</GitActionButton>
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
            <div class="process-item-main">
              <span class="process-item-id">#{{ process.id }}</span>
              <span class="process-item-name">{{ process.name }}</span>
            </div>
            <el-popconfirm
                title="确定删除此执行逻辑吗？"
                @confirm="deleteProcess(process.id)"
            >
              <template #reference>
                <GitActionButton
                    compact
                    variant="danger"
                    @click.stop
                >删除
                </GitActionButton>
              </template>
            </el-popconfirm>
          </div>
        </el-scrollbar>
      </div>
    </div>
    <div class="right-content">
      <template v-if="state.activeProcess">
        <div class="process-header">
          <div class="process-header-main">
            <div class="process-header-eyebrow">执行逻辑</div>
            <h2>{{ state.activeProcess.name }}</h2>
          </div>
          <div class="process-header-actions">
            <GitActionButton variant="info" @click="editProcessName">编辑</GitActionButton>
            <GitActionButton @click="addNewItem">新增执行逻辑子项</GitActionButton>
          </div>
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
                    <div class="item-title-group">
                      <div class="item-title-row">
                        <span class="item-id">#{{ element.id }}</span>
                        <span class="item-name">{{ element.name }}</span>
                        <span class="item-type">{{ element.type }}</span>
                      </div>
                    </div>
                    <div class="item-actions">
                      <GitActionButton compact @click="addNewItem(element)">新增复制</GitActionButton>
                      <GitActionButton compact variant="info" @click="editItem(element)">编辑</GitActionButton>
                      <el-popconfirm
                          title="确定删除此执行逻辑子项吗？"
                          @confirm="deleteItem(element.id)"
                      >
                        <template #reference>
                          <GitActionButton
                              compact
                              variant="danger"
                              @click.stop
                          >删除
                          </GitActionButton>
                        </template>
                      </el-popconfirm>
                    </div>
                  </div>
                  <div class="item-meta-list">
                    <span v-if="element.tip" class="item-meta-chip item-meta-chip--info">提示: {{ element.tip }}</span>
                    <span v-if="element.domain_limit" class="item-meta-chip">域名限制: {{ element.domain_limit }}</span>
                  </div>
                  <div class="item-details">
                    <div
                        v-for="detail in getProcessItemDetails(element)"
                        :key="`${element.id}-${detail.key}`"
                        :class="['item-detail-row', `item-detail-row--${detail.emphasis}`]"
                    >
                      <div class="item-detail-label">{{ detail.label }}</div>
                      <div class="item-detail-content">
                        <template v-if="detail.lines.length === 1">
                          <span>{{ detail.lines[0] }}</span>
                        </template>
                        <div v-else class="item-detail-multiline">
                          <div
                              v-for="(line, index) in detail.lines"
                              :key="`${detail.key}-${index}`"
                              class="item-detail-line"
                          >
                            {{ line }}
                          </div>
                        </div>
                      </div>
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
    <el-dialog v-model="state.dialogProcessName" title="编辑执行逻辑名称" :width="DEFAULT_PROCESS_DIALOG_WIDTH">
      <el-input v-model="state.editingProcessName"/>
      <template #footer>
        <GitActionButton @click="state.dialogProcessName = false">取消</GitActionButton>
        <GitActionButton @click="saveProcessName">保存</GitActionButton>
      </template>
    </el-dialog>

    <!-- 编辑执行逻辑子项对话框 -->
    <el-dialog v-model="state.dialogProcessItem" :title="state.editingItem.id ? `编辑执行逻辑子项 #${state.editingItem.id}` : '新增执行逻辑子项'" :width="DEFAULT_PROCESS_ITEM_DIALOG_WIDTH">
      <ProcessItemEditor ref="processItemEditorRef" v-model="state.editingItem" :process-item-options="state.processItems" />
      <template #footer>
        <GitActionButton @click="state.dialogProcessItem = false">取消</GitActionButton>
        <GitActionButton @click="saveProcessItem">保存</GitActionButton>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import {reactive, onMounted, ref} from 'vue'
import draggable from 'vuedraggable'
import {Menu} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import API from '@/utils/base/smart_link_proces'
import ProcessItemEditor from '@/components/smart_link/ProcessItemEditor.vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import processDisplay from '@/utils/smart_link_process_display.cjs'

// PROCESS_ITEM_WAIT_DEFAULT 表示流程项等待时长默认值。
const PROCESS_ITEM_WAIT_DEFAULT = 3000

// DEFAULT_APPEND_TO_REPLACE 表示默认不写入替换列表。
const DEFAULT_APPEND_TO_REPLACE = '0'
// DEFAULT_FLAG_DISABLED 表示默认关闭异步和错误继续等布尔开关。
const DEFAULT_FLAG_DISABLED = '0'
// DEFAULT_NEXT_IDS 表示默认没有后继节点。
const DEFAULT_NEXT_IDS = 0
// DEFAULT_IS_START 表示默认不是开始节点。
const DEFAULT_IS_START = 0
// DEFAULT_PROCESS_DIALOG_WIDTH 表示编辑流程名弹窗宽度。
const DEFAULT_PROCESS_DIALOG_WIDTH = '30%'
// DEFAULT_PROCESS_ITEM_DIALOG_WIDTH 表示编辑流程项弹窗宽度。
const DEFAULT_PROCESS_ITEM_DIALOG_WIDTH = '70%'

const {
  buildProcessItemDisplayDetails,
} = processDisplay

export default {
  components: {
    draggable,
    Menu,
    ProcessItemEditor,
    GitActionButton,
  },
  setup(props, {emit}) {
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
        append_to_replace : DEFAULT_APPEND_TO_REPLACE, //1需要加入到替换列表
        is_async : DEFAULT_FLAG_DISABLED,
        is_error_continue : DEFAULT_FLAG_DISABLED, //遇到错误是否继续
        next_ids : DEFAULT_NEXT_IDS, //下一个节点的id,多个用逗号分割
        is_start : DEFAULT_IS_START, //是否为开始节点，1是
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
        wait_mills: PROCESS_ITEM_WAIT_DEFAULT,
        tip: '',
        value: '',
        out_key: '',
        check_key: '',
        weight: state.processItems.length > 0 ?
            Math.max(...state.processItems.map(i => i.weight)) + 1 : 0,
        domain_limit: '',
        append_to_replace : DEFAULT_APPEND_TO_REPLACE, //1需要加入到替换列表
        is_async : DEFAULT_FLAG_DISABLED,
        is_error_continue : DEFAULT_FLAG_DISABLED, //遇到错误是否继续
        next_ids : DEFAULT_NEXT_IDS, //下一个节点的id,多个用逗号分割
        is_start : DEFAULT_IS_START, //是否为开始节点，1是
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
      const isValid = processItemEditorRef.value ? processItemEditorRef.value.validateForSave() : true
      if (!isValid) {
        ElMessage.error('请先修正表单中的格式问题，再保存流程项。')
        return
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
      DEFAULT_PROCESS_DIALOG_WIDTH,
      DEFAULT_PROCESS_ITEM_DIALOG_WIDTH,
      state,
      processItemEditorRef,
      getProcessItemDetails: buildProcessItemDisplayDetails,
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
  height: 100%;
  font-size: 13px;
  color: #4a4a4a;
}

.left-sidebar {
  width: 300px;
  border-right: 1px solid #e6e8de;
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f5f6f0;
}

.search-box {
  padding: 16px 16px 12px;
}

.add-btn {
  padding: 0 16px 14px;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.process-list {
  flex: 1;
  overflow: hidden;
  padding-bottom: 8px;
}

.process-item {
  padding: 10px 12px;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  border: 1px solid transparent;
  border-radius: 8px;
  margin: 6px 10px 0;
  transition: background-color 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.process-item:hover {
  background: #eef4ea;
  border-color: #dbe6d4;
}

.process-item.active {
  background: #e7f1e3;
  border-color: #d4e4c3;
  box-shadow: none;
}

.process-item-main {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.process-item-id {
  color: #5f7c53;
  font-size: 12px;
  font-weight: 600;
}

.process-item-name {
  min-width: 0;
  color: #4a4a4a;
  font-size: 14px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.right-content {
  flex: 1;
  padding: 16px 18px 18px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  background: #fafaf7;
}

.process-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 14px;
  padding: 2px 2px 14px;
  margin-bottom: 4px;
  flex-shrink: 0;
}

.process-header-main {
  min-width: 0;
}

.process-header-eyebrow {
  margin-bottom: 4px;
  color: #82917d;
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.04em;
}

.process-header h2 {
  margin: 0;
  color: #4a4a4a;
  font-size: 18px;
  font-weight: 600;
  line-height: 1.35;
}

.process-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
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
  border: 1px solid #e6e8de;
  border-radius: 10px;
  padding: 14px;
  margin-bottom: 12px;
  background: #fff;
  box-shadow: none;
}

.item-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.item-header .drag-handle {
  color: #556655;
  cursor: move;
  font-size: 14px;
}

.item-title-group {
  flex: 1;
  min-width: 0;
}

.item-title-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.item-id {
  color: #4f6546;
  font-size: 12px;
  font-weight: 600;
}

.item-name {
  color: #4a4a4a;
  font-size: 16px;
  font-weight: 600;
  line-height: 1.35;
}

.item-type {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 999px;
  background: #f2f6ee;
  color: #556655;
  font-size: 12px;
  font-weight: 500;
}

.item-actions {
  display: flex;
  gap: 8px;
  margin-left: auto;
  flex-wrap: wrap;
}

.item-meta-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 10px;
}

.item-meta-chip {
  display: inline-flex;
  align-items: center;
  padding: 4px 8px;
  border-radius: 999px;
  background: #f4f6f1;
  color: #5f7059;
  font-size: 12px;
  line-height: 1;
  font-weight: 400;
}

.item-meta-chip--info {
  background: #eef4ea;
  color: #4c7048;
}

.item-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 6px;
}

.item-detail-row {
  display: grid;
  grid-template-columns: 96px minmax(0, 1fr);
  gap: 10px;
  align-items: start;
  padding: 8px 10px;
  border-radius: 8px;
  background: #fafaf8;
  border: 1px solid #f0f1ec;
}

.item-detail-row--accent .item-detail-content {
  color: #4f804f;
  font-weight: 500;
}

.item-detail-row--block {
  background: #f8f9f6;
}

.item-detail-label {
  color: #6f7f68;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.6;
}

.item-detail-content {
  min-width: 0;
  color: #4a4a4a;
  font-size: 13px;
  line-height: 1.6;
  word-break: break-all;
}

.item-detail-multiline {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.item-detail-line {
  padding: 6px 8px;
  border-radius: 8px;
  background: #ffffff;
  border: 1px solid #eef0ea;
}

.empty-tip {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: #879483;
  font-size: 14px;
  border: 1px dashed #d8dfd3;
  border-radius: 10px;
  background: #ffffff;
}

:deep(.search-box .el-input__wrapper),
:deep(.el-dialog .el-input__wrapper) {
  border-radius: 10px;
  background: #ffffff;
  box-shadow: 0 0 0 1px #dde4d8 inset;
}

:deep(.add-btn .git-action-button),
:deep(.process-header-actions .git-action-button) {
  height: 32px;
  padding: 6px 12px;
  font-size: 12px;
}

:deep(.process-item .git-action-button),
:deep(.item-actions .git-action-button) {
  font-size: 12px;
}

:deep(.search-box .el-input__wrapper.is-focus),
:deep(.el-dialog .el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #97b595 inset;
}

@media (max-width: 960px) {
  .smart-process-container {
    flex-direction: column;
  }

  .left-sidebar {
    width: 100%;
    height: 300px;
    border-right: none;
    border-bottom: 1px solid #e6e8de;
  }

  .process-header {
    flex-direction: column;
  }

  .process-header-actions {
    width: 100%;
  }

  .item-header {
    align-items: flex-start;
    flex-wrap: wrap;
  }

  .item-actions {
    margin-left: 0;
  }

  .item-detail-row {
    grid-template-columns: 1fr;
    gap: 6px;
  }
}
</style>
