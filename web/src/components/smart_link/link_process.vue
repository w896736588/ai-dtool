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
        <GitActionButton compact sizeMode="compact-small" @click="createNewProcess">新增执行逻辑</GitActionButton>
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
            <div class="process-item-actions">
              <GitActionButton
                  compact
                  sizeMode="compact-small"
                  variant="info"
                  @click.stop="openCopyProcessDialog(process)"
              >复制
              </GitActionButton>
              <el-popconfirm
                  title="确定删除此执行逻辑吗？"
                  @confirm="deleteProcess(process.id)"
              >
                <template #reference>
                  <GitActionButton
                      compact
                      sizeMode="compact-small"
                      variant="danger"
                      @click.stop
                  >删除
                  </GitActionButton>
                </template>
              </el-popconfirm>
            </div>
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
            <GitActionButton compact sizeMode="compact-small" variant="info" @click="copyProcessSql(state.activeProcess)">复制SQL</GitActionButton>
            <GitActionButton compact sizeMode="compact-small" variant="info" @click="editProcessName">编辑</GitActionButton>
            <GitActionButton compact sizeMode="compact-small" @click="addNewItem">新增执行逻辑子项</GitActionButton>
          </div>
        </div>
        <div class="process-items-wrapper">
          <el-scrollbar class="process-items-scroll">
            <div class="process-table-header">
              <div class="process-table-header__drag"></div>
              <div class="process-table-header__name">节点</div>
              <div class="process-table-header__detail">关键配置</div>
              <div class="process-table-header__actions">操作</div>
            </div>
            <draggable
                v-model="state.processItems"
                handle=".drag-handle"
                item-key="id"
                @end="handleSortEnd"
            >
              <template #item="{ element }">
                <div class="process-table-row">
                  <div class="process-table-row__drag">
                    <el-icon class="drag-handle">
                      <Menu/>
                    </el-icon>
                  </div>
                  <div class="process-table-row__name">
                    <div class="item-title-row">
                      <span class="item-id">#{{ element.id }}</span>
                      <span class="item-name">{{ element.name }}</span>
                      <span class="item-type">{{ element.type }}</span>
                    </div>
                  </div>
                  <div class="process-table-row__detail">
                    <div
                        v-for="detail in getCompactProcessItemDetails(element)"
                        :key="`${element.id}-${detail.key}`"
                        class="item-detail-inline"
                    >
                      <span class="item-detail-inline__label">{{ detail.label }}</span>
                      <span class="item-detail-inline__value">{{ detail.value }}</span>
                    </div>
                  </div>
                  <div class="process-table-row__actions">
                    <el-dropdown
                        trigger="click"
                        @command="(command) => handleRowActionCommand(element, command)"
                    >
                      <GitActionButton compact sizeMode="compact-small" variant="info">
                        操作
                      </GitActionButton>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item command="copy_new">复制新增</el-dropdown-item>
                          <el-dropdown-item command="copy_from">复制节点</el-dropdown-item>
                          <el-dropdown-item command="edit">编辑</el-dropdown-item>
                          <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
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

    <el-dialog v-model="state.dialogCopyProcess" title="复制执行流程" :width="DEFAULT_PROCESS_DIALOG_WIDTH">
      <el-input v-model="state.copyProcessName" placeholder="请输入新执行流程名称" />
      <template #footer>
        <GitActionButton @click="state.dialogCopyProcess = false">取消</GitActionButton>
        <GitActionButton @click="confirmCopyProcess">保存</GitActionButton>
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

    <el-dialog v-model="state.dialogCopyFromProcessItem" title="复制其他流程节点" width="520px">
      <div class="copy-process-dialog">
        <el-form label-width="110px">
          <el-form-item label="选择执行流程">
            <el-select
              v-model="state.copySourceProcessId"
              placeholder="请选择执行流程"
              style="width: 100%"
              @change="handleCopySourceProcessChange"
            >
              <el-option
                v-for="process in state.processes"
                :key="`copy-process-${process.id}`"
                :label="`#${process.id} ${process.name}`"
                :value="process.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="选择节点">
            <el-select
              v-model="state.copySourceItemId"
              placeholder="请选择节点"
              style="width: 100%"
              :disabled="state.copySourceItems.length === 0"
            >
              <el-option
                v-for="item in state.copySourceItems"
                :key="`copy-item-${item.id}`"
                :label="`#${item.id} ${item.name} (${item.type})`"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
        </el-form>
        <div class="copy-process-dialog__tip">确认后会覆盖当前节点配置，保留当前节点的 ID、所属执行流程和权重。</div>
      </div>
      <template #footer>
        <GitActionButton @click="state.dialogCopyFromProcessItem = false">取消</GitActionButton>
        <GitActionButton @click="confirmCopyFromProcessItem">确定覆盖</GitActionButton>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import {reactive, onMounted, ref} from 'vue'
import draggable from 'vuedraggable'
import {Menu} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
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
      dialogCopyProcess: false,
      copySourceProcess: null,
      copyProcessName: '',
      dialogProcessItem: false,
      dialogCopyFromProcessItem: false,
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
      },
      copyTargetItem: null,
      copySourceProcessId: null,
      copySourceItemId: null,
      copySourceItems: [],
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

    const smartProcessAddAsync = function (payload) {
      return new Promise((resolve) => {
        API.SmartProcessAdd(payload, function (response) {
          resolve(response && response.Data ? response.Data : {})
        })
      })
    }

    const smartProcessItemListAsync = function (processId) {
      return new Promise((resolve) => {
        API.SmartProcessItemList({ smart_link_process_id: processId }, function (response) {
          resolve(response && response.Data ? (response.Data.list || []) : [])
        })
      })
    }

    const smartProcessItemAddAsync = function (payload) {
      return new Promise((resolve) => {
        API.SmartProcessItemAdd(payload, function (response) {
          resolve(response && response.Data ? response.Data : {})
        })
      })
    }

    const writeClipboard = async function (text) {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(text)
        return
      }
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.setAttribute('readonly', 'readonly')
      textarea.style.position = 'fixed'
      textarea.style.top = '-9999px'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }

    const sqlEscape = function (value) {
      return String(value ?? '').replace(/'/g, "''")
    }

    const sqlString = function (value) {
      return `'${sqlEscape(value)}'`
    }

    const sqlInt = function (value, defaultValue = 0) {
      const numberValue = Number(value)
      return Number.isFinite(numberValue) ? String(Math.trunc(numberValue)) : String(defaultValue)
    }

    const buildProcessExportSql = function (process, itemList) {
      const normalizedProcess = {
        name: process?.name ?? '',
        status: process?.status ?? 1,
        create_time: process?.create_time ?? 0,
        update_time: process?.update_time ?? 0,
      }
      const processColumns = ['name', 'status', 'create_time', 'update_time']
      const processValues = [
        sqlString(normalizedProcess.name),
        sqlInt(normalizedProcess.status, 1),
        sqlInt(normalizedProcess.create_time),
        sqlInt(normalizedProcess.update_time),
      ]
      const lines = [
        '-- 开启事务',
        'BEGIN TRANSACTION;',
        '',
        '-- 1. 插入主表',
        `INSERT INTO "tbl_smart_link_process" (${processColumns.map((column) => `"${column}"`).join(', ')})`,
        `VALUES (${processValues.join(', ')});`,
        '',
        '-- 2. 创建临时表保存刚插入的ID',
        'CREATE TEMP TABLE temp_last_id AS SELECT last_insert_rowid() as saved_id;',
      ]

      const itemColumns = [
        'name',
        'smart_link_process_id',
        'type',
        'locator',
        'tip',
        'value',
        'out_key',
        'check_key',
        'weight',
        'domain_limit',
        'status',
        'create_time',
        'update_time',
        'wait_mills',
        'append_to_replace',
        'is_async',
        'is_error_continue',
        'next_ids',
        'x',
        'y',
      ]
      lines.push('', '-- 3. 插入所有子表记录，使用临时表中保存的ID')
      itemList.forEach((item) => {
        const itemValues = [
          sqlString(item?.name),
          '(SELECT saved_id FROM temp_last_id)',
          sqlString(item?.type),
          sqlString(item?.locator),
          sqlString(item?.tip),
          sqlString(item?.value),
          sqlString(item?.out_key),
          sqlString(item?.check_key),
          sqlInt(item?.weight),
          sqlString(item?.domain_limit),
          sqlInt(item?.status, 1),
          sqlInt(item?.create_time),
          sqlInt(item?.update_time),
          sqlInt(item?.wait_mills, 3000),
          sqlString(item?.append_to_replace ?? '0'),
          sqlString(item?.is_async ?? '0'),
          sqlString(item?.is_error_continue ?? '0'),
          sqlString(item?.next_ids),
          sqlInt(item?.x),
          sqlInt(item?.y),
        ]
        lines.push(`INSERT INTO "main"."tbl_smart_link_process_item" (${itemColumns.map((column) => `"${column}"`).join(', ')})`)
        lines.push(`VALUES (${itemValues.join(', ')});`)
      })

      lines.push('', '-- 4. 删除临时表', 'DROP TABLE temp_last_id;', '', '-- 5. 提交事务', 'COMMIT;')
      return lines.join('\n')
    }

    const copyProcessSql = async function (process) {
      try {
        const itemList = await smartProcessItemListAsync(process.id)
        const sql = buildProcessExportSql(process, itemList)
        await writeClipboard(sql)
        ElMessage.success('SQL 已复制到剪贴板。')
      } catch (error) {
        ElMessage.error(`复制 SQL 失败：${error && error.message ? error.message : '未知错误'}`)
      }
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

    const openCopyProcessDialog = function (process) {
      state.copySourceProcess = JSON.parse(JSON.stringify(process))
      state.copyProcessName = `${process.name}-复制`
      state.dialogCopyProcess = true
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

    const confirmCopyProcess = async function () {
      const sourceProcess = state.copySourceProcess
      const nextName = String(state.copyProcessName || '').trim()
      if (!sourceProcess || !sourceProcess.id) {
        ElMessage.error('没有可复制的执行流程。')
        return
      }
      if (!nextName) {
        ElMessage.error('请输入新执行流程名称。')
        return
      }

      const createdProcess = await smartProcessAddAsync({
        id: 0,
        name: nextName,
      })
      if (!createdProcess.id) {
        ElMessage.error('复制执行流程失败：创建新流程失败。')
        return
      }

      const sourceItemList = await smartProcessItemListAsync(sourceProcess.id)
      const idMap = {}
      const createdPairs = []

      for (const sourceItem of sourceItemList) {
        const createdItem = await smartProcessItemAddAsync({
          ...JSON.parse(JSON.stringify(sourceItem)),
          id: 0,
          smart_link_process_id: createdProcess.id,
          next_ids: '',
        })
        if (createdItem.id) {
          idMap[sourceItem.id] = createdItem.id
          createdPairs.push({
            source: sourceItem,
            created: createdItem,
          })
        }
      }

      for (const pair of createdPairs) {
        const nextIds = String(pair.source.next_ids || '')
          .split(',')
          .map((item) => String(item || '').trim())
          .filter(Boolean)
          .map((item) => idMap[item])
          .filter(Boolean)
          .join(',')
        await smartProcessItemAddAsync({
          ...JSON.parse(JSON.stringify(pair.created)),
          next_ids: nextIds,
        })
      }

      state.dialogCopyProcess = false
      state.copySourceProcess = null
      state.copyProcessName = ''
      fetchProcesses()
      ElMessage.success('执行流程复制成功。')
    }

    const fetchProcessItems = function (processId) {
      API.SmartProcessItemList({smart_link_process_id: processId}, function (response) {
        if (response && response.Data) {
          state.processItems = response.Data.list
        }
      })
    }

    const fetchProcessItemsByProcessId = function (processId, callback) {
      API.SmartProcessItemList({smart_link_process_id: processId}, function (response) {
        const list = response && response.Data ? (response.Data.list || []) : []
        if (typeof callback === 'function') {
          callback(list)
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

    const openCopyFromProcessDialog = function (targetItem) {
      state.copyTargetItem = JSON.parse(JSON.stringify(targetItem))
      state.copySourceProcessId = state.activeProcess ? state.activeProcess.id : null
      state.copySourceItemId = null
      state.copySourceItems = []
      state.dialogCopyFromProcessItem = true
      if (state.copySourceProcessId) {
        handleCopySourceProcessChange(state.copySourceProcessId)
      }
    }

    const handleCopySourceProcessChange = function (processId) {
      state.copySourceItemId = null
      if (!processId) {
        state.copySourceItems = []
        return
      }
      fetchProcessItemsByProcessId(processId, function (list) {
        state.copySourceItems = list
      })
    }

    const confirmCopyFromProcessItem = function () {
      if (!state.copyTargetItem || !state.copySourceProcessId || !state.copySourceItemId) {
        ElMessage.error('请先选择执行流程和节点。')
        return
      }
      const sourceItem = state.copySourceItems.find((item) => item.id === state.copySourceItemId)
      if (!sourceItem) {
        ElMessage.error('没有找到要复制的节点，请重新选择。')
        return
      }
      state.editingItem = {
        ...JSON.parse(JSON.stringify(sourceItem)),
        id: state.copyTargetItem.id,
        smart_link_process_id: state.copyTargetItem.smart_link_process_id,
        weight: state.copyTargetItem.weight,
        x: state.copyTargetItem.x,
        y: state.copyTargetItem.y,
      }
      state.dialogCopyFromProcessItem = false
      state.dialogProcessItem = true
    }

    const saveProcessItem = function () {
      const isValid = processItemEditorRef.value ? processItemEditorRef.value.validateForSave() : true
      if (!isValid) {
        ElMessage.error('请先修正表单中的格式问题，再保存流程项。')
        return
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

    // handleRowActionCommand 统一处理表格行下拉动作，避免右侧操作按钮过度占宽。
    // Handle row action dropdown commands so the action column stays compact.
    const handleRowActionCommand = async function (item, command) {
      if (command === 'copy_new') {
        addNewItem(item)
        return
      }
      if (command === 'copy_from') {
        openCopyFromProcessDialog(item)
        return
      }
      if (command === 'edit') {
        editItem(item)
        return
      }
      if (command === 'delete') {
        try {
          await ElMessageBox.confirm('确定删除此执行逻辑子项吗？', '确认删除', {
            type: 'warning',
            confirmButtonText: '删除',
            cancelButtonText: '取消',
          })
          deleteItem(item.id)
        } catch (error) {
          // 用户取消删除时无需额外提示。 // No extra feedback is needed when the user cancels.
        }
      }
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
      getCompactProcessItemDetails: function (item) {
        return buildProcessItemDisplayDetails(item)
          .slice(0, 3)
          .map((detail) => ({
            key: detail.key,
            label: detail.label,
            value: Array.isArray(detail.lines) ? detail.lines.join(' / ') : '',
          }))
      },
      searchList,
      createNewProcess,
      selectProcess,
      deleteProcess,
      copyProcessSql,
      openCopyProcessDialog,
      editProcessName,
      saveProcessName,
      confirmCopyProcess,
      addNewItem,
      openCopyFromProcessDialog,
      handleCopySourceProcessChange,
      confirmCopyFromProcessItem,
      editItem,
      saveProcessItem,
      deleteItem,
      handleRowActionCommand,
      handleSortEnd,
      changeToLinks,
      ProcessCancelRelation,
      ProcessSetRelation,
      ProcessSetPosition,
    }
  }
}
</script>

<style scoped src="@/css/components/smart_link/link_process.css"></style>
