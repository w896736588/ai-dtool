<template>
  <div class="markdown-container">
    <!-- 左侧文档列表 -->
    <el-container class="full-height">
      <el-aside class="sidebar" width="280px">
        <div class="sidebar-header">
          <el-input
              v-model="state.searchQuery"
              class="search-input"
              clearable
              placeholder="搜索文档..."
              @input="searchList"
          >
            <template #append>
              <pl-button type="primary" @click="createNewDoc">新增</pl-button>
            </template>
          </el-input>
        </div>

        <el-scrollbar class="doc-list-scroll">
          <!-- 使用 draggable 组件替换原来的 el-menu -->
          <draggable
              v-model="state.filteredDocs"
              item-key="id"
              handle=".drag-handle"
              @end="handleSortEnd"
          >
            <template #item="{ element }">
              <div
                  class="doc-item"
                  :class="{ active: state.activeDoc.id === element.id }"
                  @click="selectDoc(element.id)"
              >
                <el-icon class="drag-handle"><Menu /></el-icon>
                <span class="doc-title">{{ element.name }}</span>
                <el-popconfirm
                    title="确定删除此文档吗？"
                    @confirm="deleteDoc(element.id)"
                >
                  <template #reference>
                    <pl-button
                        class="doc-delete-btn"
                        type="text"
                        @click.stop
                    >删除</pl-button>
                  </template>
                </el-popconfirm>
              </div>
            </template>
          </draggable>
        </el-scrollbar>
      </el-aside>

      <!-- 右侧编辑器 -->
      <el-main class="editor-main">
        <div v-if="state.activeDoc" class="editor-container" @keydown.ctrl.s.prevent="saveDoc">
          <div class="editor-header">
            <el-input
                v-model="state.activeDoc.name"
                class="title-input"
                placeholder="文档标题"
                @input="changeContent"
            />
            <el-tag v-loading="state.isSaving">
              <span v-if="state.isSave === 0" style="color: red;">未保存</span>
              <span v-else style="color: green;">已保存</span>
            </el-tag>
            &nbsp;
            <pl-button
                class="save-btn"
                type="primary"
                @click="saveDoc"
            >
              保存
            </pl-button>
            <pl-button @click="ShowHistoryList">变更记录</pl-button>
          </div>
          <MdEditor v-model="state.activeDoc.content" @blur="changeContent" :onSave="changeContent" />
        </div>

        <el-empty v-else description="请选择或创建文档"/>
      </el-main>
    </el-container>

  </div>

  <el-dialog v-model="state.dialogMarkdownHistory" title="变更历史" width="60%">
    <el-form >
      <el-table :data="state.markdownHistoryList" style="width: 100%">
        <el-table-column prop="create_time_desc" label="变动时间" width="200" />
        <el-table-column prop="change_desc" label="变更简要" />
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <pl-button type="primary" link @click="ShowDiff(scope.row)">查看变更</pl-button>
            <el-tooltip content="删除" placement="top">
              <el-popconfirm
                  cancel-button-text="取消"
                  confirm-button-text="删除"
                  icon-color="#626AEF"
                  title="确定删除吗?"
                  @confirm="DeleteHistory(scope.row)"
              >
                <template #reference>
                  <pl-button link type="danger" >删除记录</pl-button>
                </template>
              </el-popconfirm>
            </el-tooltip>

          </template>
        </el-table-column>
      </el-table>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="state.dialogMarkdownHistory = false">取消</pl-button>
        <pl-button type="primary" @click="starSave">
          保存
        </pl-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog v-model="state.dialogMarkdownDiff" title="文档对比" width="60%">
    <diff :old-text="state.oldCode" :new-text="state.newCode" v-model="state.dialogMarkdownDiff" ></diff>
  </el-dialog>
</template>

<script>
import {ref, computed, onMounted, onUnmounted, reactive} from 'vue';
import { MdEditor } from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';
import API from '@/utils/base/markdown'
import DiffCode from '@/components/base/diff_markwodn.vue'
import draggable from 'vuedraggable';

export default {
  components: {
    MdEditor, diff: DiffCode,draggable,
  },
  props: {
    markdownType: {
      type: String,
      default: 'normal'
    },
  },
  setup(props) {
    const state = reactive({
      dialogMarkdownHistory : false,
      dialogMarkdownDiff : false,
      markdownHistoryList : [],
      oldCode: `function hello() {\n  console.log('Hello, world!');\n}`,
      newCode: `function hello(name) {\n  console.log('Hello, ' + name);\n}`,
      isSaving : false,//是否处于保存中
      isSave : 1,//是否已经保存 1已经保存 0未保存
      searchQuery: '',//当前搜索的内容
      docs: [], //所有文档列表
      filteredDocs: [], //显示的文档列表
      activeDoc: {}, //当前选中的文档
    })
    // 新增的排序方法
    const MarkdownSort = function(markdown_ids, callBack) {
      API.MarkdownSort(markdown_ids , function(response) {
        if (callBack) callBack(response);
      });
    };

    // 拖拽结束处理
    const handleSortEnd = function() {
      const markdownIds = state.filteredDocs.map(doc => doc.id).join(',');
      MarkdownSort(markdownIds, function() {
        // 排序成功后的回调
        fetchDocs(); // 重新获取文档列表以确保顺序正确
      });
    };

    const toolbar = 'undo redo clear | h bold italic strikethrough quote | ul ol table hr | link image code';
    let autoSaveInterval = null;
    const searchList = function () {
      if (state.searchQuery !== '') {
        state.filteredDocs = []
        state.activeDoc = {}
        for (let i = 0; i < state.docs.length; i++) {
          if(state.docs[i].name.indexOf(state.searchQuery) > -1 || state.docs[i].content.indexOf(state.searchQuery) > -1){
            state.filteredDocs.push(state.docs[i])
          }
        }
      } else {
        state.filteredDocs = state.docs
      }
      if(state.filteredDocs.length > 0 && (!state.activeDoc || !state.activeDoc.id || parseInt(state.activeDoc.id) <= 0)){
        state.activeDoc = state.filteredDocs[0]
      }
    }
    console.log('类型' , props.markdownType)
    // Methods
    const fetchDocs = function () {
      API.MarkdownList(props.markdownType ,function (response) {
        if (response && response.Data) {
          state.docs = response.Data;
          console.log(state.docs)
          if (state.docs.length > 0) {
            state.activeDoc = state.docs[0]
            setTimeout(function (){
              state.isSave = 1
            } , 500)
          }
          searchList()
          if (state.docs.length === 0) {
            createNewDoc();
          }
        }
      });
    };

    const createNewDoc = function(){
      const newDoc = {
        id: 0,
        name: `新文档 ${state.docs.length + 1}`,
        content: '# 新文档\n\n从这里开始编辑...',
        createdAt: new Date(),
        updatedAt: new Date()
      };
      API.MarkdownAdd(newDoc.id,props.markdownType, newDoc.name, newDoc.content, function (response) {
        newDoc.id = response.Data.id
        state.docs.unshift(newDoc);
        state.activeDoc = newDoc
        searchList()
        //saveToLocalStorage();
      });

    };

    const selectDoc = function(id) {
      saveDoc()
      for(let doc of state.docs){
        if(doc.id === id){
          state.activeDoc = doc
          break;
        }
      }
    };

    const deleteDoc = async (id) => {
      API.MarkdownDel(id, function (response) {
        fetchDocs()
        //saveToLocalStorage();
      });
    };

    const changeContent = function() {
      state.isSave = 0
      state.isSaving = false
    }

    const saveDoc = function() {
      if (!state.activeDoc.name){
        return;
      }
      state.isSaving = true
      state.activeDoc.updatedAt = new Date();
      API.MarkdownAdd(
          state.activeDoc.id,
          props.markdownType,
          state.activeDoc.name,
          state.activeDoc.content, function () {
            setTimeout(function (){
              state.isSave = 1
              state.isSaving = false
            } , 500)

          }
      );
    };

    // Auto-save every 5 seconds
    const startAutoSave = () => {
      autoSaveInterval = setInterval(() => {
        //已保存时不处理
        if(state.isSave !== 0){
          return
        }
        saveDoc();
      }, 60000); // 5 seconds
    };

    // Initialize
    onMounted(() => {
      fetchDocs();
      startAutoSave();
    });

    // Clean up interval when component is unmounted
    onUnmounted(() => {
      if (autoSaveInterval) {
        clearInterval(autoSaveInterval);
      }
    });

    const ShowHistoryList = function (){
      state.dialogMarkdownHistory = true
      API.MarkdownHistoryList(
          state.activeDoc.id,function (response) {
            state.markdownHistoryList = response.Data
          }
      );
    }

    const ShowDiff = function (row){
      console.log(row)
      state.dialogMarkdownDiff = true
      state.oldCode = row.old_content
      state.newCode = row.new_content
    }

    const DeleteHistory = function (row){
      console.log(row)
      API.MarkdownHistoryDel(
          row.id,function (response) {
           ShowHistoryList()
          }
      );
    }

    return {
      toolbar,
      createNewDoc,
      selectDoc,
      deleteDoc,
      saveDoc,
      searchList,
      state,
      changeContent,
      ShowDiff,
      DeleteHistory,
      ShowHistoryList,
      handleSortEnd,
    };
  }
};
</script>

<style scoped lang="scss" src="@/css/components/Markdown.scss"></style>
