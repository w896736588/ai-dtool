<template>
  <div class="e2e-page-container">
    <!-- 顶部标签页 -->
    <div class="e2e-header-tabs">
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="分组管理" name="group">
          <template #label>
            <span class="tab-label">
              <el-icon><FolderOpened /></el-icon>
              分组管理
            </span>
          </template>
        </el-tab-pane>
        <el-tab-pane label="用例管理" name="case">
          <template #label>
            <span class="tab-label">
              <el-icon><Document /></el-icon>
              用例管理
            </span>
          </template>
        </el-tab-pane>
        <el-tab-pane label="执行记录" name="run">
          <template #label>
            <span class="tab-label">
              <el-icon><VideoPlay /></el-icon>
              执行记录
            </span>
          </template>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 分组管理面板 -->
    <div v-if="activeTab === 'group'" class="e2e-group-panel">
      <div class="panel-toolbar">
        <el-button type="primary" @click="showGroupDialog('create')">
          <el-icon><Plus /></el-icon>
          新建分组
        </el-button>
      </div>

      <el-table :data="groupList" v-loading="groupLoading" stripe border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="分组名称" min-width="200">
          <template #default="{ row }">
            <span class="group-name">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="case_count" label="用例数" width="100" align="center" />
        <el-table-column prop="notification_enabled" label="通知" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.notification_enabled ? 'success' : 'info'" size="small">
              {{ row.notification_enabled ? '已启用' : '未启用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="showGroupDialog('edit', row)">编辑</el-button>
            <el-button size="small" type="primary" link @click="showCasesByGroup(row)">查看用例</el-button>
            <el-button size="small" type="danger" link @click="deleteGroup(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 用例管理面板 -->
    <div v-if="activeTab === 'case'" class="e2e-case-panel">
      <div class="panel-toolbar">
        <el-select v-model="caseFilterGroupId" placeholder="选择分组" clearable @change="loadCaseList" class="group-filter">
          <el-option v-for="g in groupList" :key="g.id" :label="g.name" :value="g.id" />
        </el-select>
        <el-button type="primary" @click="showCaseDialog('create')">
          <el-icon><Plus /></el-icon>
          新建用例
        </el-button>
        <el-button type="success" @click="openRecorderDialog">
          <el-icon><VideoCamera /></el-icon>
          开始录制
        </el-button>
      </div>

      <el-table :data="caseList" v-loading="caseLoading" stripe border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="用例名称" min-width="200" />
        <el-table-column prop="group_name" label="所属分组" width="150" />
        <el-table-column prop="env_url" label="环境URL" min-width="200" show-overflow-tooltip />
        <el-table-column prop="step_count" label="步骤数" width="80" align="center" />
        <el-table-column prop="assertion_count" label="断言数" width="80" align="center" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getCaseStatusType(row.status)" size="small">
              {{ getCaseStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="创建时间" width="180" />
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="executeCase(row)">执行</el-button>
            <el-button size="small" type="primary" link @click="showCaseDialog('edit', row)">编辑</el-button>
            <el-button size="small" type="info" link @click="copyCase(row)">复制</el-button>
            <el-button size="small" type="danger" link @click="deleteCase(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="casePage"
          v-model:page-size="casePageSize"
          :total="caseTotal"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadCaseList"
          @current-change="loadCaseList"
        />
      </div>
    </div>

    <!-- 执行记录面板 -->
    <div v-if="activeTab === 'run'" class="e2e-run-panel">
      <div class="panel-toolbar">
        <el-select v-model="runFilterGroupId" placeholder="选择分组" clearable @change="loadRunList" class="group-filter">
          <el-option v-for="g in groupList" :key="g.id" :label="g.name" :value="g.id" />
        </el-select>
        <el-select v-model="runFilterStatus" placeholder="执行状态" clearable @change="loadRunList" class="status-filter">
          <el-option label="全部" value="" />
          <el-option label="成功" value="passed" />
          <el-option label="失败" value="failed" />
          <el-option label="运行中" value="running" />
          <el-option label="待执行" value="pending" />
        </el-select>
        <el-button @click="loadRunList">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>

      <el-table :data="runList" v-loading="runLoading" stripe border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="case_name" label="用例名称" min-width="180" />
        <el-table-column prop="group_name" label="分组" width="120" />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getRunStatusType(row.status)" size="small">
              {{ getRunStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="total_steps" label="总步骤" width="80" align="center" />
        <el-table-column prop="passed_steps" label="通过" width="80" align="center">
          <template #default="{ row }">
            <span class="text-success">{{ row.passed_steps || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="failed_steps" label="失败" width="80" align="center">
          <template #default="{ row }">
            <span class="text-danger">{{ row.failed_steps || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="duration_ms" label="耗时" width="100" align="center">
          <template #default="{ row }">
            {{ formatDuration(row.duration_ms) }}
          </template>
        </el-table-column>
        <el-table-column prop="create_time" label="开始时间" width="180" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" link @click="showRunDetail(row)">详情</el-button>
            <el-button v-if="row.status === 'running'" size="small" type="danger" link @click="stopRun(row)">停止</el-button>
            <el-button size="small" type="info" link @click="showRunRequests(row)">请求追踪</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="runPage"
          v-model:page-size="runPageSize"
          :total="runTotal"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadRunList"
          @current-change="loadRunList"
        />
      </div>
    </div>

    <!-- 分组编辑弹窗 -->
    <el-dialog v-model="groupDialogVisible" :title="groupDialogTitle" width="500px">
      <el-form :model="groupForm" label-width="100px">
        <el-form-item label="分组名称" required>
          <el-input v-model="groupForm.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="通知">
          <el-switch v-model="groupForm.notification_enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="groupDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="groupFormSaving" @click="saveGroup">确定</el-button>
      </template>
    </el-dialog>

    <!-- 用例编辑弹窗 -->
    <el-dialog v-model="caseDialogVisible" :title="caseDialogTitle" width="80%" top="5vh">
      <el-form :model="caseForm" label-width="100px" class="case-form">
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="用例名称" required>
              <el-input v-model="caseForm.name" placeholder="请输入用例名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="所属分组" required>
              <el-select v-model="caseForm.group_id" placeholder="选择分组">
                <el-option v-for="g in groupList" :key="g.id" :label="g.name" :value="g.id" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="环境URL" required>
              <el-input v-model="caseForm.env_url" placeholder="https://example.com" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="基础路径">
              <el-input v-model="caseForm.env_base_url" placeholder="/api/v1" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="全局变量">
              <el-input
                v-model="caseForm.variables"
                type="textarea"
                :rows="3"
                placeholder='{"key": "value"}'
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="步骤配置" required>
              <div class="steps-editor">
                <div v-for="(step, idx) in caseForm.steps" :key="idx" class="step-item">
                  <span class="step-index">{{ idx + 1 }}</span>
                  <el-select v-model="step.type" placeholder="步骤类型" class="step-type">
                    <el-option label="打开环境" value="open_env" />
                    <el-option label="点击" value="click" />
                    <el-option label="输入" value="input" />
                    <el-option label="等待" value="wait" />
                    <el-option label="悬停" value="hover" />
                    <el-option label="下拉选择" value="select" />
                    <el-option label="页面导航" value="navigation" />
                  </el-select>
                  <el-input v-model="step.selector" placeholder="选择器" class="step-selector" />
                  <el-input v-model="step.value" placeholder="值" class="step-value" />
                  <el-button type="danger" link @click="removeStep(idx)">删除</el-button>
                </div>
                <el-button type="primary" plain @click="addStep">
                  <el-icon><Plus /></el-icon>
                  添加步骤
                </el-button>
              </div>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="断言配置">
              <div class="assertions-editor">
                <div v-for="(assertion, idx) in caseForm.assertions" :key="idx" class="assertion-item">
                  <span class="assertion-index">{{ idx + 1 }}</span>
                  <el-select v-model="assertion.type" placeholder="断言类型" class="assertion-type">
                    <el-option label="文本断言" value="text" />
                    <el-option label="元素断言" value="element" />
                    <el-option label="URL模式" value="url_pattern" />
                    <el-option label="API响应" value="api_response" />
                    <el-option label="API请求" value="api_request" />
                  </el-select>
                  <el-input v-model="assertion.target" placeholder="目标" class="assertion-target" />
                  <el-input v-model="assertion.expected" placeholder="期望值" class="assertion-expected" />
                  <el-button type="danger" link @click="removeAssertion(idx)">删除</el-button>
                </div>
                <el-button type="primary" plain @click="addAssertion">
                  <el-icon><Plus /></el-icon>
                  添加断言
                </el-button>
              </div>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="24">
            <el-form-item label="执行后通知">
              <el-switch v-model="caseForm.notification_enabled" />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <el-button @click="caseDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="caseFormSaving" @click="saveCase">确定</el-button>
      </template>
    </el-dialog>

    <!-- 执行详情弹窗 -->
    <el-dialog v-model="runDetailVisible" title="执行详情" width="90%" top="5vh">
      <div v-if="runDetail" class="run-detail">
        <div class="detail-header">
          <el-descriptions :column="3" border>
            <el-descriptions-item label="用例名称">{{ runDetail.case_name }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="getRunStatusType(runDetail.status)">{{ getRunStatusText(runDetail.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="耗时">{{ formatDuration(runDetail.duration_ms) }}</el-descriptions-item>
            <el-descriptions-item label="开始时间">{{ runDetail.create_time }}</el-descriptions-item>
            <el-descriptions-item label="通过/总数">{{ runDetail.passed_steps }}/{{ runDetail.total_steps }}</el-descriptions-item>
            <el-descriptions-item label="失败数">
              <span class="text-danger">{{ runDetail.failed_steps }}</span>
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <div class="detail-steps">
          <h4>步骤详情</h4>
          <el-table :data="runDetail.steps" border size="small">
            <el-table-column prop="step_index" label="#" width="60" align="center" />
            <el-table-column prop="step_type" label="类型" width="120" />
            <el-table-column prop="selector" label="选择器" min-width="200" show-overflow-tooltip />
            <el-table-column prop="value" label="值" min-width="150" show-overflow-tooltip />
            <el-table-column prop="status" label="状态" width="100" align="center">
              <template #default="{ row }">
                <el-tag :type="row.status === 'passed' ? 'success' : 'danger'" size="small">
                  {{ row.status === 'passed' ? '通过' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="error_msg" label="错误信息" min-width="200" show-overflow-tooltip />
            <el-table-column prop="duration_ms" label="耗时" width="100" align="center">
              <template #default="{ row }">
                {{ formatDuration(row.duration_ms) }}
              </template>
            </el-table-column>
            <el-table-column label="截图" width="80" align="center">
              <template #default="{ row }">
                <el-button v-if="row.screenshot" size="small" type="primary" link @click="showScreenshot(row)">查看</el-button>
                <span v-else>-</span>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-dialog>

    <!-- 请求追踪弹窗 -->
    <el-dialog v-model="requestsDialogVisible" title="请求追踪" width="90%" top="5vh">
      <div v-if="selectedRunId" class="requests-list">
        <el-table :data="requestList" v-loading="requestsLoading" border size="small" max-height="60vh">
          <el-table-column prop="url" label="URL" min-width="300" show-overflow-tooltip />
          <el-table-column prop="method" label="方法" width="80" align="center">
            <template #default="{ row }">
              <el-tag size="small">{{ row.method }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status_code" label="状态码" width="80" align="center" />
          <el-table-column prop="step_index" label="关联步骤" width="100" align="center" />
          <el-table-column prop="request_time" label="请求时间" width="180" />
          <el-table-column label="操作" width="120" align="center">
            <template #default="{ row }">
              <el-button size="small" type="primary" link @click="showRequestDetail(row)">详情</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>

    <!-- 请求详情弹窗 -->
    <el-dialog v-model="requestDetailVisible" title="请求详情" width="80%" top="5vh">
      <div v-if="requestDetail" class="request-detail">
        <el-tabs>
          <el-tab-pane label="请求信息" name="request">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="URL">{{ requestDetail.url }}</el-descriptions-item>
              <el-descriptions-item label="方法">{{ requestDetail.method }}</el-descriptions-item>
              <el-descriptions-item label="状态码">{{ requestDetail.status_code }}</el-descriptions-item>
              <el-descriptions-item label="请求时间">{{ requestDetail.request_time }}</el-descriptions-item>
            </el-descriptions>
            <div class="detail-section">
              <h5>请求头</h5>
              <pre class="code-block">{{ formatJson(requestDetail.request_headers) }}</pre>
            </div>
            <div class="detail-section">
              <h5>请求体</h5>
              <pre class="code-block">{{ formatJson(requestDetail.request_body) }}</pre>
            </div>
          </el-tab-pane>
          <el-tab-pane label="响应信息" name="response">
            <div class="detail-section">
              <h5>响应头</h5>
              <pre class="code-block">{{ formatJson(requestDetail.response_headers) }}</pre>
            </div>
            <div class="detail-section">
              <h5>响应体</h5>
              <pre class="code-block">{{ formatJson(requestDetail.response_body) }}</pre>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-dialog>

    <!-- 录制入口对话框：tab 切换 启动录制 / 导入 JSON -->
    <el-dialog v-model="recorderDialogVisible" title="录制 / 导入用例步骤" width="640px">
      <el-tabs v-model="recorderTab">
        <el-tab-pane label="启动录制会话" name="live">
          <el-form :model="recorderForm" label-width="100px">
            <el-form-item label="会话名" required>
              <el-input v-model="recorderForm.session_name" placeholder="录制会话名称" />
            </el-form-item>
            <el-form-item label="所属分组" required>
              <el-select v-model="recorderForm.group_id" placeholder="选择分组">
                <el-option v-for="g in groupList" :key="g.id" :label="g.name" :value="g.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="选择链接" required>
              <el-select v-model="recorderForm.smart_link_id" filterable placeholder="选择 smart_link 链接" @change="onSmartLinkPick">
                <el-option v-for="opt in smartLinkOptions" :key="opt.id" :label="opt.label" :value="opt.id" />
              </el-select>
            </el-form-item>
            <el-form-item label="选择账号">
              <el-select v-model="recorderForm.user_name" :disabled="!smartLinkUserOptions.length">
                <el-option v-for="u in smartLinkUserOptions" :key="u" :label="u" :value="u" />
              </el-select>
            </el-form-item>
            <el-form-item label="关联用例">
              <el-select v-model="recorderForm.case_id" placeholder="可选：关联到现有用例" clearable filterable>
                <el-option
                  v-for="c in caseList"
                  :key="c.id"
                  :label="`[${c.group_name || ''}] ${c.name}`"
                  :value="c.id"
                />
              </el-select>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
      <!--
        关键：el-dialog 的 #footer 必须挂在 el-dialog 直接 children，不能藏在 el-tabs > el-tab-pane 里
        （el-tab-pane 没有 footer 插槽，tab 内的 footer 模板会被静默忽略）。
        「导入录制 JSON」已经从启动弹窗里挪到 sessionDialogVisible 会话详情弹窗中。
      -->
      <template #footer>
        <el-button @click="recorderDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="recorderStarting" @click="startRecording">启动录制</el-button>
      </template>
    </el-dialog>

    <!-- 步骤确认弹窗 -->
    <StepConfirmDialog
      v-if="pendingStep"
      v-model:visible="stepConfirmVisible"
      :step="pendingStep"
      :session-id="recorderSession ? recorderSession.session_id : ''"
      title="确认录制步骤"
      @confirmed="confirmPendingStep"
      @cancelled="cancelPendingStep"
    />

    <!-- 会话详情对话框 -->
    <el-dialog v-model="sessionDialogVisible" :title="sessionDialogTitle" width="90%" top="3vh">
      <div v-if="currentSession" class="session-detail">
        <el-descriptions :column="3" border size="small">
          <el-descriptions-item label="会话 ID">{{ currentSession.id }}</el-descriptions-item>
          <el-descriptions-item label="业务 ID">{{ currentSession.session_id }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ currentSession.status }}</el-descriptions-item>
          <el-descriptions-item label="环境URL" :span="3">{{ currentSession.env_url }}</el-descriptions-item>
        </el-descriptions>

        <div class="session-toolbar">
          <el-button @click="replayWholeSession" :loading="replayingAll">
            <el-icon><VideoPlay /></el-icon>
            整段回放
          </el-button>
          <el-button type="primary" @click="openCommitDialog">
            <el-icon><Check /></el-icon>
            提交到用例
          </el-button>
        </div>

        <!--
          导入步骤 JSON：从 dtool-record-*.json 文件或剪贴板导入步骤，追加到当前会话的步骤表。
          不创建新会话，不 commit 用例；导入完成后用户继续走"提交到用例"。
        -->
        <el-collapse v-model="importPanelOpen" class="session-import-panel">
          <el-collapse-item name="import" title="导入步骤 JSON（追加到当前会话）">
            <el-alert
              type="info"
              :closable="false"
              show-icon
              style="margin-bottom: 8px;"
              title="如何获取 JSON"
              description="在录制浏览器里点 toolbar 的「结束并下载」会自动下载 dtool-record-*.json；或点「复制 JSON」拿到剪贴板内容。"
            />
            <el-form label-width="100px">
              <el-form-item label="JSON 内容">
                <el-input
                  v-model="importJsonText"
                  type="textarea"
                  :rows="6"
                  placeholder='粘贴 recorder toolbar 导出的 JSON，或点下方"选择文件"导入 .json 文件'
                />
              </el-form-item>
              <el-form-item>
                <input ref="importJsonFileInput" type="file" accept=".json,application/json" style="display:none" @change="onImportJsonFile" />
                <el-button size="small" @click="$refs.importJsonFileInput.click()">选择 .json 文件</el-button>
                <el-button size="small" type="primary" :loading="importParsing" @click="parseImportJson">解析并预览</el-button>
                <el-button size="small" @click="pasteFromClipboard">从剪贴板粘贴</el-button>
              </el-form-item>
              <el-form-item v-if="importParseResult" label="预览">
                <div class="import-preview">
                  <el-tag size="small">解析成功：{{ (importParseResult.steps || []).length }} 步</el-tag>
                  <el-button size="small" type="success" :loading="importAppending" @click="appendImportStepsToSession">追加到当前会话</el-button>
                </div>
              </el-form-item>
            </el-form>
          </el-collapse-item>
        </el-collapse>

        <el-table :data="currentSession.steps || []" border stripe size="small" max-height="50vh">
          <el-table-column type="index" label="#" width="60" />
          <el-table-column prop="type" label="类型" width="180">
            <template #default="{ row }">
              <el-tag size="small">{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="180" />
          <el-table-column label="配置" min-width="220">
            <template #default="{ row }">
              <code class="config-cell">{{ formatJson(row.config) }}</code>
            </template>
          </el-table-column>
          <el-table-column label="等待(ms)" width="80" prop="wait_after_ms" />
          <el-table-column label="断言" width="80" align="center">
            <template #default="{ row }">
              <el-tag v-if="(row.assertions || []).length" size="small" type="warning">
                {{ (row.assertions || []).length }}
              </el-tag>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button size="small" type="primary" link @click="replayOneStep(row)">回放</el-button>
              <el-button size="small" type="warning" link @click="editRecordedStep(row)">编辑</el-button>
              <el-button size="small" type="danger" link @click="deleteRecordedStep(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-dialog>

    <!-- 录制会话提交对话框 -->
    <el-dialog v-model="commitDialogVisible" title="提交录制为用例" width="520px">
      <el-form :model="commitForm" label-width="100px">
        <el-form-item label="目标分组" required>
          <el-select v-model="commitForm.group_id" placeholder="选择分组">
            <el-option v-for="g in groupList" :key="g.id" :label="g.name" :value="g.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="用例名称">
          <el-input v-model="commitForm.name" placeholder="留空则使用会话名" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="commitForm.tags" placeholder="逗号分隔" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="commitDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="commitToCase" :loading="committing">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import base from '../utils/base'
import {
  FolderOpened,
  Document,
  VideoPlay,
  VideoCamera,
  Plus,
  Refresh,
  Check,
  Delete,
} from '@element-plus/icons-vue'
import StepConfirmDialog from './e2e/StepConfirmDialog.vue'

export default {
  name: 'E2e',
  components: {
    FolderOpened,
    Document,
    VideoPlay,
    VideoCamera,
    Plus,
    Refresh,
    Check,
    Delete,
    StepConfirmDialog,
  },
  data() {
    return {
      activeTab: 'group',
      // 分组相关
      groupList: [],
      groupLoading: false,
      groupDialogVisible: false,
      groupDialogTitle: '新建分组',
      groupForm: {
        id: null,
        name: '',
        notification_enabled: false,
      },
      groupFormSaving: false,
      // 用例相关
      caseList: [],
      caseLoading: false,
      caseDialogVisible: false,
      caseDialogTitle: '新建用例',
      caseForm: {
        id: null,
        group_id: null,
        name: '',
        env_url: '',
        env_base_url: '',
        variables: '',
        steps: [],
        assertions: [],
        notification_enabled: false,
      },
      caseFormSaving: false,
      caseFilterGroupId: null,
      casePage: 1,
      casePageSize: 20,
      caseTotal: 0,
      // 执行记录相关
      runList: [],
      runLoading: false,
      runDetailVisible: false,
      runDetail: null,
      runFilterGroupId: null,
      runFilterStatus: '',
      runPage: 1,
      runPageSize: 20,
      runTotal: 0,
      // 请求追踪相关
      requestsDialogVisible: false,
      selectedRunId: null,
      requestList: [],
      requestsLoading: false,
      requestDetailVisible: false,
      requestDetail: null,

      // ===== 录制功能 =====
      recorderDialogVisible: false,
      recorderTab: 'live', // live | import
      recorderForm: {
        session_name: '',
        group_id: null,
        smart_link_id: null,
        user_name: '',
        case_id: null,
      },
      smartLinkOptions: [],
      smartLinkUserOptions: [],
      recorderStarting: false,
      recorderSession: null, // { id, session_id, status }
      toolbarVisible: true,
      toolbarRecording: true,
      toolbarMode: 'click',
      recordedSteps: [],
      pendingStep: null,
      stepConfirmVisible: false,
      // 导入 JSON
      importJsonText: '',
      importParseResult: null, // {schema, steps, ...} | null
      importParsing: false,
      importAppending: false,
      importPanelOpen: [],
      sessionDialogVisible: false,
      sessionDialogTitle: '录制会话详情',
      currentSession: null,
      replayingAll: false,
      commitDialogVisible: false,
      commitForm: {
        group_id: null,
        name: '',
        tags: '',
      },
      committing: false,
    }
  },
  created() {
    this.loadGroupList()
    this.loadCaseList()
    this.loadRunList()
  },
  methods: {
    // ============ 分组管理 ============
    async loadGroupList() {
      this.groupLoading = true
      base.BasePost('/api/e2e/group/list', { page: 1, page_size: 100 }, (res) => {
        this.groupLoading = false
        if (res && res.ErrCode === 0) {
          this.groupList = res.Data?.list || []
        }
      })
    },
    showGroupDialog(type, row = null) {
      if (type === 'create') {
        this.groupDialogTitle = '新建分组'
        this.groupForm = { id: null, name: '', notification_enabled: false }
      } else {
        this.groupDialogTitle = '编辑分组'
        this.groupForm = { ...row }
      }
      this.groupDialogVisible = true
    },
    saveGroup() {
      if (!this.groupForm.name?.trim()) {
        this.$message.warning('请输入分组名称')
        return
      }
      this.groupFormSaving = true
      const api = this.groupForm.id ? '/api/e2e/group/update' : '/api/e2e/group/create'
      base.BasePost(api, this.groupForm, (res) => {
        this.groupFormSaving = false
        if (res && res.ErrCode === 0) {
          this.groupDialogVisible = false
          this.loadGroupList()
          this.$message.success('保存成功')
        } else {
          this.$message.error(res?.ErrMsg || '保存失败')
        }
      })
    },
    deleteGroup(row) {
      this.$confirm(`确定删除分组「${row.name}」吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        base.BasePost('/api/e2e/group/delete', { id: row.id }, (res) => {
          if (res && res.ErrCode === 0) {
            this.loadGroupList()
            this.$message.success('删除成功')
          } else {
            this.$message.error(res?.ErrMsg || '删除失败')
          }
        })
      }).catch(() => {})
    },
    showCasesByGroup(row) {
      this.caseFilterGroupId = row.id
      this.activeTab = 'case'
    },

    // ============ 用例管理 ============
    async loadCaseList() {
      this.caseLoading = true
      const params = {
        page: this.casePage,
        page_size: this.casePageSize,
      }
      if (this.caseFilterGroupId) {
        params.group_id = this.caseFilterGroupId
      }
      base.BasePost('/api/e2e/case/list', params, (res) => {
        this.caseLoading = false
        if (res && res.ErrCode === 0) {
          this.caseList = res.Data?.list || []
          this.caseTotal = res.Data?.pagination?.total || 0
        }
      })
    },
    showCaseDialog(type, row = null) {
      if (type === 'create') {
        this.caseDialogTitle = '新建用例'
        this.caseForm = {
          id: null,
          group_id: this.caseFilterGroupId,
          name: '',
          env_url: '',
          env_base_url: '',
          variables: '',
          steps: [],
          assertions: [],
          notification_enabled: false,
        }
      } else {
        this.caseDialogTitle = '编辑用例'
        const parsedCase = { ...row }
        try {
          if (parsedCase.variables && typeof parsedCase.variables === 'string') {
            parsedCase.variables = parsedCase.variables
          } else {
            parsedCase.variables = JSON.stringify(parsedCase.variables || {}, null, 2)
          }
        } catch (e) {
          parsedCase.variables = ''
        }
        this.caseForm = parsedCase
      }
      this.caseDialogVisible = true
    },
    addStep() {
      this.caseForm.steps.push({ type: 'open_env', selector: '', value: '' })
    },
    removeStep(idx) {
      this.caseForm.steps.splice(idx, 1)
    },
    addAssertion() {
      this.caseForm.assertions.push({ type: 'text', target: '', expected: '' })
    },
    removeAssertion(idx) {
      this.caseForm.assertions.splice(idx, 1)
    },
    saveCase() {
      if (!this.caseForm.name?.trim()) {
        this.$message.warning('请输入用例名称')
        return
      }
      if (!this.caseForm.group_id) {
        this.$message.warning('请选择所属分组')
        return
      }
      if (!this.caseForm.env_url?.trim()) {
        this.$message.warning('请输入环境URL')
        return
      }
      if (!this.caseForm.steps || this.caseForm.steps.length === 0) {
        this.$message.warning('请至少添加一个步骤')
        return
      }
      this.caseFormSaving = true
      const api = this.caseForm.id ? '/api/e2e/case/update' : '/api/e2e/case/create'
      base.BasePost(api, this.caseForm, (res) => {
        this.caseFormSaving = false
        if (res && res.ErrCode === 0) {
          this.caseDialogVisible = false
          this.loadCaseList()
          this.$message.success('保存成功')
        } else {
          this.$message.error(res?.ErrMsg || '保存失败')
        }
      })
    },
    copyCase(row) {
      const copied = { ...row, id: null, name: `${row.name} (副本)` }
      delete copied.create_time
      delete copied.update_time
      this.caseForm = copied
      this.caseDialogTitle = '复制用例'
      this.caseDialogVisible = true
    },
    deleteCase(row) {
      this.$confirm(`确定删除用例「${row.name}」吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        base.BasePost('/api/e2e/case/delete', { id: row.id }, (res) => {
          if (res && res.ErrCode === 0) {
            this.loadCaseList()
            this.$message.success('删除成功')
          } else {
            this.$message.error(res?.ErrMsg || '删除失败')
          }
        })
      }).catch(() => {})
    },
    executeCase(row) {
      this.$confirm(`确定执行用例「${row.name}」吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'info',
      }).then(() => {
        base.BasePost('/api/e2e/run/execute', { case_id: row.id }, (res) => {
          if (res && res.ErrCode === 0) {
            this.$message.success('执行已开始')
            this.activeTab = 'run'
            this.loadRunList()
          } else {
            this.$message.error(res?.ErrMsg || '执行失败')
          }
        })
      }).catch(() => {})
    },

    // ============ 执行记录 ============
    async loadRunList() {
      this.runLoading = true
      const params = {
        page: this.runPage,
        page_size: this.runPageSize,
      }
      if (this.runFilterGroupId) {
        params.group_id = this.runFilterGroupId
      }
      if (this.runFilterStatus) {
        params.status = this.runFilterStatus
      }
      base.BasePost('/api/e2e/run/list', params, (res) => {
        this.runLoading = false
        if (res && res.ErrCode === 0) {
          this.runList = res.Data?.list || []
          this.runTotal = res.Data?.pagination?.total || 0
        }
      })
    },
    showRunDetail(row) {
      base.BasePost('/api/e2e/run/detail', { run_id: row.id }, (res) => {
        if (res && res.ErrCode === 0) {
          this.runDetail = res.Data
          this.runDetailVisible = true
        } else {
          this.$message.error(res?.ErrMsg || '加载失败')
        }
      })
    },
    stopRun(row) {
      this.$confirm(`确定停止执行「${row.case_name}」吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        base.BasePost('/api/e2e/run/stop', { run_id: row.id }, (res) => {
          if (res && res.ErrCode === 0) {
            this.$message.success('已停止')
            this.loadRunList()
          } else {
            this.$message.error(res?.ErrMsg || '停止失败')
          }
        })
      }).catch(() => {})
    },
    showRunRequests(row) {
      this.selectedRunId = row.id
      this.requestsDialogVisible = true
      this.loadRequestList(row.id)
    },
    loadRequestList(runId) {
      this.requestsLoading = true
      base.BaseGet(`/api/e2e/run/${runId}/requests`, {}, (res) => {
        this.requestsLoading = false
        if (res && res.ErrCode === 0) {
          this.requestList = res.Data?.requests || []
        }
      })
    },
    showRequestDetail(row) {
      base.BaseGet(`/api/e2e/run/${this.selectedRunId}/request/${row.id}`, {}, (res) => {
        if (res && res.ErrCode === 0) {
          this.requestDetail = res.Data
          this.requestDetailVisible = true
        } else {
          this.$message.error(res?.ErrMsg || '加载失败')
        }
      })
    },

    // ============ 辅助方法 ============
    handleTabChange(tab) {
      // 切换时刷新数据
      if (tab === 'group') {
        this.loadGroupList()
      } else if (tab === 'case') {
        this.loadCaseList()
      } else if (tab === 'run') {
        this.loadRunList()
      }
    },
    getCaseStatusType(status) {
      const map = { draft: 'info', active: 'success', disabled: 'warning' }
      return map[status] || 'info'
    },
    getCaseStatusText(status) {
      const map = { draft: '草稿', active: '启用', disabled: '禁用' }
      return map[status] || status
    },
    getRunStatusType(status) {
      const map = { passed: 'success', failed: 'danger', running: 'warning', pending: 'info' }
      return map[status] || 'info'
    },
    getRunStatusText(status) {
      const map = { passed: '成功', failed: '失败', running: '运行中', pending: '待执行' }
      return map[status] || status
    },
    formatDuration(ms) {
      if (!ms) return '-'
      if (ms < 1000) return `${ms}ms`
      if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`
      return `${(ms / 60000).toFixed(1)}m`
    },
    formatJson(str) {
      if (!str) return ''
      if (typeof str === 'object') return JSON.stringify(str, null, 2)
      try {
        return JSON.stringify(JSON.parse(str), null, 2)
      } catch (e) {
        return str
      }
    },
    showScreenshot(row) {
      if (row.screenshot) {
        window.open(row.screenshot, '_blank')
      }
    },

    // ============ 录制功能 ============
    openRecorderDialog() {
      // 默认填一个会话名
      const now = new Date()
      this.recorderForm = {
        session_name: `录制-${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')} ${String(now.getHours()).padStart(2, '0')}${String(now.getMinutes()).padStart(2, '0')}`,
        group_id: this.caseFilterGroupId || null,
        smart_link_id: null,
        user_name: '',
        case_id: null,
      }
      this.smartLinkOptions = []
      this.smartLinkUserOptions = []
      // 拉取 smart_link 列表，附带 userList
      base.BasePost('/api/SmartLinkItemList', {}, (res) => {
        if (res && res.ErrCode === 0) {
          const list = (res.Data && res.Data.smart_link_list) || []
          this.smartLinkOptions = list.map((it) => ({
            id: it.id,
            label: it.label,
            userList: Array.isArray(it.userList) ? it.userList : [],
          }))
        }
      })
      this.recorderDialogVisible = true
    },

    onSmartLinkPick() {
      const opt = this.smartLinkOptions.find((o) => o.id === this.recorderForm.smart_link_id)
      this.smartLinkUserOptions = (opt && opt.userList && opt.userList.map((u) => u.user_name)) || []
      if (this.smartLinkUserOptions.length === 1) this.recorderForm.user_name = this.smartLinkUserOptions[0]
    },

    startRecording() {
      if (!this.recorderForm.session_name?.trim()) {
        this.$message.warning('请填写会话名')
        return
      }
      if (!this.recorderForm.group_id) {
        this.$message.warning('请选择分组')
        return
      }
      if (!this.recorderForm.smart_link_id) {
        this.$message.warning('请选择链接')
        return
      }
      this.recorderStarting = true
      base.BasePost('/api/e2e/record/open', {
        smart_link_id: this.recorderForm.smart_link_id,
        link_id: this.recorderForm.smart_link_id,
        user_name: this.recorderForm.user_name || '',
        session_name: this.recorderForm.session_name,
        group_id: this.recorderForm.group_id,
        case_id: this.recorderForm.case_id || 0,
      }, (res) => {
        this.recorderStarting = false
        if (!(res && res.ErrCode === 0)) {
          this.$message.error(res?.ErrMsg || '启动失败')
          return
        }
        this.recorderSession = res.Data
        this.recordedSteps = []
        this.toolbarMode = 'click'
        this.toolbarRecording = true
        this.recorderDialogVisible = false
        this.$message.success('录制会话已创建，浏览器由 smart_link 接管')
        this.openSessionDialog()
      })
    },

    onToolbarModeChange(m) {
      this.toolbarMode = m
      // 切换到步骤列表时刷新
      if (m === 'list') this.loadRecordedSteps()
    },

    toggleRecorderRecording() {
      this.toolbarRecording = !this.toolbarRecording
      this.$message.info(this.toolbarRecording ? '已继续录制' : '已暂停录制')
    },

    closeRecorder() {
      if (this.recorderSession && this.recordedSteps.length === 0) {
        this.$confirm('当前会话没有步骤，确定关闭吗？', '提示', { type: 'warning' }).then(() => {
          this._doCloseRecorder()
        }).catch(() => {})
      } else {
        this._doCloseRecorder()
      }
    },

    _doCloseRecorder() {
      if (this.recorderSession) {
        base.BasePost('/api/e2e/record/session/delete', { id: this.recorderSession.session_id }, () => {})
      }
      this.toolbarVisible = false
      this.recorderSession = null
      this.recordedSteps = []
    },

    loadRecordedSteps() {
      if (!this.recorderSession) return
      base.BasePost('/api/e2e/record/session/get', { id: this.recorderSession.session_id }, (res) => {
        if (res && res.ErrCode === 0) {
          this.currentSession = res.Data
          this.recordedSteps = res.Data?.steps || []
        }
      })
    },

    /**
     * 录制工具条按钮被外部触发后调用此方法（实际由前端录制脚本注入）
     * 简化版：提供一个手动添加步骤的入口（用于开发调试）
     */
    addRecordedStepManually(stepType, config, description) {
      if (!this.recorderSession) {
        this.$message.warning('请先启动录制会话')
        return
      }
      if (!this.toolbarRecording) {
        this.$message.warning('已暂停，请先继续录制')
        return
      }
      const step = {
        id: 'stp_' + Date.now() + '_' + Math.floor(Math.random() * 1000),
        type: stepType,
        version: '1.0',
        description: description || '',
        wait_after_ms: 200,
        config: config || {},
        recorded_at: Date.now(),
      }
      base.BasePost('/api/e2e/record/step/add', {
        session_id: this.recorderSession.session_id,
        step,
      }, (res) => {
        if (res && res.ErrCode === 0) {
          this.pendingStep = res.Data?.step || step
          this.stepConfirmVisible = true
          this.loadRecordedSteps()
        } else {
          this.$message.error(res?.ErrMsg || '追加步骤失败')
        }
      })
    },

    // ============ 导入录制 JSON ============
    // onImportJsonFile 用户选了 .json 文件：读 FileReader 内容回填 textarea
    onImportJsonFile(ev) {
      const f = ev.target.files && ev.target.files[0]
      if (!f) return
      const reader = new FileReader()
      reader.onload = () => {
        this.importJsonText = String(reader.result || '')
        this.parseImportJson()
      }
      reader.onerror = () => this.$message.error('读取文件失败')
      reader.readAsText(f)
      // 清空 value 允许选同一文件
      ev.target.value = ''
    },
    // pasteFromClipboard 从 navigator.clipboard 读取
    async pasteFromClipboard() {
      try {
        const text = await navigator.clipboard.readText()
        this.importJsonText = text || ''
        this.parseImportJson()
      } catch (e) {
        this.$message.warning('剪贴板读取失败：请检查浏览器权限或直接粘贴')
      }
    },
    // parseImportJson 把 textarea 文本解析成 {schema, steps, ...}
    parseImportJson() {
      const text = (this.importJsonText || '').trim()
      if (!text) { this.$message.warning('JSON 内容为空'); return }
      this.importParsing = true
      try {
        const obj = JSON.parse(text)
        const steps = Array.isArray(obj.steps) ? obj.steps
          : Array.isArray(obj) ? obj
            : null
        if (!steps) throw new Error('JSON 缺少 steps 数组')
        for (const s of steps) {
          if (!s || typeof s !== 'object' || !s.type) {
            throw new Error('步骤缺少 type 字段')
          }
        }
        this.importParseResult = obj
        this.$message.success(`解析成功：${steps.length} 步`)
      } catch (e) {
        this.importParseResult = null
        this.$message.error('解析失败：' + e.message)
      } finally {
        this.importParsing = false
      }
    },
    // appendImportStepsToSession 把解析后的 steps 逐条 append 到当前会话（不再创建新会话、不再直接 commit 用例）。
    // 导入完成后用户继续点 "提交到用例" 走原 commit 流程。
    appendImportStepsToSession() {
      if (!this.importParseResult) { this.$message.warning('请先解析 JSON'); return }
      if (!this.recorderSession) { this.$message.warning('当前没有打开的录制会话'); return }
      const steps = this.importParseResult.steps || []
      if (!steps.length) { this.$message.warning('JSON 中没有步骤'); return }
      this.importAppending = true
      const sessionId = this.recorderSession.session_id
      const self = this
      let i = 0
      let failed = 0
      const next = () => {
        if (i >= steps.length) return finish()
        const s = steps[i++]
        base.BasePost('/api/e2e/record/step/add', {
          session_id: sessionId,
          step: {
            id: 'stp_' + Date.now() + '_' + Math.floor(Math.random() * 10000),
            type: s.type,
            version: s.version || '1.0',
            description: s.description || '',
            wait_after_ms: s.wait_after_ms || 0,
            config: s.config || {},
            recorded_at: s.recorded_at || Date.now(),
          },
        }, (r) => {
          if (!(r && r.ErrCode === 0)) failed++
          next()
        })
      }
      const finish = () => {
        self.importAppending = false
        if (failed > 0) {
          self.$message.warning(`已追加：${steps.length - failed}/${steps.length} 步（${failed} 失败）`)
        } else {
          self.$message.success(`已追加 ${steps.length} 步到当前会话`)
        }
        self.importJsonText = ''
        self.importParseResult = null
        self.loadRecordedSteps()
      }
      next()
    },

    confirmPendingStep(updatedStep) {
      if (!this.pendingStep || !this.recorderSession) return
      base.BasePost('/api/e2e/record/step/update', {
        session_id: this.recorderSession.session_id,
        step_id: this.pendingStep.id,
        step: updatedStep,
      }, (res) => {
        if (res && res.ErrCode === 0) {
          this.$message.success('步骤已确认')
          this.stepConfirmVisible = false
          this.pendingStep = null
          this.loadRecordedSteps()
        } else {
          this.$message.error(res?.ErrMsg || '更新步骤失败')
        }
      })
    },

    cancelPendingStep() {
      // 取消即删除该步骤
      if (this.pendingStep && this.recorderSession) {
        base.BasePost('/api/e2e/record/step/delete', {
          session_id: this.recorderSession.session_id,
          step_id: this.pendingStep.id,
        }, () => {
          this.pendingStep = null
          this.loadRecordedSteps()
        })
      } else {
        this.pendingStep = null
      }
    },

    openSessionDialog() {
      this.sessionDialogVisible = true
      this.sessionDialogTitle = `录制会话详情 #${this.recorderSession?.id || ''}`
      this.loadRecordedSteps()
    },

    replayOneStep(row) {
      if (!this.recorderSession) return
      this.$message.info(`回放步骤 ${row.type}...`)
      base.BasePost('/api/e2e/record/step/replay', {
        session_id: this.recorderSession.session_id,
        step_id: row.id,
      }, (res) => {
        if (res && res.ErrCode === 0) {
          if (res.Data?.success) this.$message.success('回放成功')
          else this.$message.error('回放失败：' + (res.Data?.error || ''))
        } else {
          this.$message.error(res?.ErrMsg || '请求失败')
        }
      })
    },

    editRecordedStep(row) {
      this.pendingStep = row
      this.stepConfirmVisible = true
    },

    deleteRecordedStep(row) {
      this.$confirm('确定删除该步骤吗？', '提示', { type: 'warning' }).then(() => {
        base.BasePost('/api/e2e/record/step/delete', {
          session_id: this.recorderSession.session_id,
          step_id: row.id,
        }, (res) => {
          if (res && res.ErrCode === 0) {
            this.$message.success('已删除')
            this.loadRecordedSteps()
          }
        })
      }).catch(() => {})
    },

    replayWholeSession() {
      if (!this.recorderSession) return
      this.replayingAll = true
      base.BasePost('/api/e2e/record/session/replay', {
        session_id: this.recorderSession.session_id,
        start_index: 0,
        continue_on_error: true,
      }, (res) => {
        this.replayingAll = false
        if (res && res.ErrCode === 0) {
          if (res.Data?.success) this.$message.success('整段回放成功')
          else this.$message.warning('整段回放完成，存在失败：' + (res.Data?.error || ''))
        } else {
          this.$message.error(res?.ErrMsg || '回放失败')
        }
      })
    },

    openCommitDialog() {
      if (!this.recorderSession || !this.currentSession) {
        this.$message.warning('会话未加载')
        return
      }
      this.commitForm = {
        group_id: this.currentSession.group_id || this.caseFilterGroupId,
        name: '',
        tags: '',
      }
      this.commitDialogVisible = true
    },

    commitToCase() {
      if (!this.commitForm.group_id) {
        this.$message.warning('请选择目标分组')
        return
      }
      this.committing = true
      base.BasePost('/api/e2e/record/commit', {
        session_id: this.recorderSession.session_id,
        group_id: this.commitForm.group_id,
        name: this.commitForm.name,
        tags: this.commitForm.tags,
      }, (res) => {
        this.committing = false
        if (res && res.ErrCode === 0) {
          this.$message.success(`已提交为用例 #${res.Data.case_id}`)
          this.commitDialogVisible = false
          this.sessionDialogVisible = false
          this.toolbarVisible = false
          this.recorderSession = null
          this.recordedSteps = []
          this.loadCaseList()
        } else {
          this.$message.error(res?.ErrMsg || '提交失败')
        }
      })
    },
  },
}
</script>

<style scoped>
.e2e-page-container {
  padding: 16px;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.e2e-header-tabs {
  margin-bottom: 16px;
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 6px;
}

.panel-toolbar {
  margin-bottom: 16px;
  display: flex;
  gap: 12px;
  align-items: center;
}

.group-filter,
.status-filter {
  width: 200px;
}

.pagination-wrapper {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.group-name {
  font-weight: 500;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}

.case-form {
  max-height: 60vh;
  overflow-y: auto;
}

.steps-editor,
.assertions-editor {
  border: 1px dashed #dcdfe6;
  border-radius: 4px;
  padding: 12px;
  background: #fafafa;
}

.step-item,
.assertion-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.step-index,
.assertion-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: #409eff;
  color: #fff;
  border-radius: 50%;
  font-size: 12px;
  flex-shrink: 0;
}

.step-type,
.step-selector,
.step-value {
  flex: 1;
  min-width: 100px;
}

.assertion-type,
.assertion-target,
.assertion-expected {
  flex: 1;
  min-width: 100px;
}

.run-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-header {
  padding: 0;
}

.detail-steps {
  margin-top: 16px;
}

.detail-steps h4 {
  margin-bottom: 12px;
  color: #303133;
}

.detail-section {
  margin-top: 16px;
}

.detail-section h5 {
  margin-bottom: 8px;
  color: #606266;
}

.code-block {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.6;
  max-height: 300px;
  overflow-y: auto;
}

.requests-list {
  min-height: 200px;
}

.session-detail {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.session-toolbar {
  display: flex;
  gap: 8px;
  margin: 10px 0;
}
.config-cell {
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 11px;
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
