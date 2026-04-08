<template>
  <div class="api-documentation">
    <!-- 左侧导航菜单 -->
    <div class="sidebar">
      <div class="sidebar-header">
        <h3>接口列表</h3>
        <pl-button size="small" type="primary" @click="copyAllApisAsMarkdown">
          复制所有接口(Markdown)
        </pl-button>
        <pl-button v-if="folderId" size="small" @click="openDocumentPage">
          在新窗口查看完整文档
        </pl-button>
      </div>
      <div class="sidebar-content">
        <ul class="api-menu">
          <li
              v-for="api in apis"
              :key="api.id"
              :class="{ active: activeApiId === api.id }"
              @click="scrollToApi(api.id)"
          >
            <!-- 移除了 el-tag，只保留名称 -->
            <span class="api-name">{{ api.name }}</span>
          </li>
        </ul>
      </div>
    </div>

    <!-- 右侧API详情内容（保持不变） -->
    <div class="api-content">
      <div class="content-wrapper">
        <div
            v-for="api in apis"
            :id="'api-' + api.id"
            :key="api.id"
            class="api-section"
        >
          <div class="api-header">
            <el-tag :type="getMethodTagType(api.method)" class="method-tag">
              {{ api.method }}
            </el-tag>
            <h3 class="api-title">{{ api.name }}</h3>
          </div>

          <el-descriptions :column="1" border>
            <el-descriptions-item label="请求URL">
              {{ api.method + ' ' + formatApiUrl(api.url) }}
            </el-descriptions-item>
            <el-descriptions-item label="请求类型" v-if="api.method !== 'GET'">
              <el-tag type="info">{{ api.content_type }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatTimestamp(api.create_time) }}
            </el-descriptions-item>
          </el-descriptions>
          <h4 v-if="api.desc">描述</h4>
          <div
              class="markdown-content"
              v-html="renderMarkdown(api.desc)"
              v-if="api.desc"
          ></div>

          <!-- 请求头 -->
          <h4 v-if="getParsedHeaders(api) && Object.keys(getParsedHeaders(api)).length > 0">请求头</h4>
          <div v-if="getParsedHeaders(api) && Object.keys(getParsedHeaders(api)).length > 0" class="section-content">
            <key-value-view :data="getParsedHeaders(api)"/>
          </div>

          <!-- 请求参数 -->
          <h4 v-if="getParsedQueryParams(api) && getParsedQueryParams(api).length > 0">请求参数</h4>
          <div v-if="getParsedQueryParams(api) && getParsedQueryParams(api).length > 0" class="section-content">
            <el-table :data="getParsedQueryParams(api)" style="width: 100%">
              <el-table-column label="字段名" prop="field" width="200"></el-table-column>
              <el-table-column label="类型" prop="type" width="120"></el-table-column>
              <el-table-column label="值" prop="value" show-overflow-tooltip></el-table-column>
              <el-table-column label="描述" prop="description"></el-table-column>
            </el-table>
          </div>

          <!-- 请求体 -->
          <h4 v-if="hasBodyContent(api)">请求体</h4>
          <div v-if="hasBodyContent(api)" class="section-content">
            <div v-if="api.content_type === 'application/json' && getParsedBodyJson(api) && Object.keys(getParsedBodyJson(api)).length > 0">
              <pre class="json-body">{{ formatJsonBody(getParsedBodyJson(api)) }}</pre>
            </div>
            <div v-else-if="(api.content_type === 'application/x-www-form-urlencoded' || api.content_type === 'multipart/form-data') && getParsedBodyForm(api) && getParsedBodyForm(api).length > 0">
              <div class="body-forms-container">
                <el-table :data="getParsedBodyForm(api)" style="width: 100%">
                  <el-table-column label="字段名" prop="field" width="200"></el-table-column>
                  <el-table-column label="类型" prop="type" width="120"></el-table-column>
                  <el-table-column label="值" prop="value" show-overflow-tooltip></el-table-column>
                  <el-table-column label="描述" prop="description"></el-table-column>
                </el-table>
              </div>
            </div>
            <div v-else-if="api.body_raw && api.body_raw !== ''">
              <pre class="raw-body">{{ api.body_raw }}</pre>
            </div>
            <div v-else class="no-content">无请求体内容</div>
          </div>

          <!-- 请求体 -->
          <h4 v-if="hasLastResult(api)">请求结果</h4>
          <div v-if="hasLastResult(api)" class="section-content">
            <pre class="json-body">{{ formatJsonBody(api.last_result) }}</pre>
          </div>

          <!-- 结果提取 -->
<!--          <div v-if="getParsedResponseTake(api) && getParsedResponseTake(api).length > 0" class="section-content">-->
<!--            <h4>结果提取</h4>-->
<!--            <div v-for="(item, index) in getParsedResponseTake(api)" :key="index" class="response-take-item">-->
<!--              <div><strong>变量名:</strong> {{ item.key }}</div>-->
<!--              <div><strong>提取路径:</strong> {{ item.path }}</div>-->
<!--              <div><strong>提取方式:</strong> {{ item.type || 'default' }}</div>-->
<!--            </div>-->
<!--          </div>-->

          <div class="section-divider"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import KeyValueView from './KeyValueView.vue'
import KeyValueEditor from "@/components/api/KeyValueEditor.vue";
import Base from '@/utils/base'
import { marked } from 'marked';

export default {
  name: 'ApiDocumentation',
  components: {
    KeyValueEditor,
    KeyValueView
  },
  props: {
    apis: {
      type: Array,
      default: () => []
    },
    folderName: {
      type: String,
      default: ''
    },
    folderId: {
      type: Number,
      default: null
    }
  },
  data() {
    return {
      activeApiId: null,
      localIp: ''
    }
  },
  mounted() {
    if (this.apis.length > 0) {
      this.activeApiId = this.apis[0].id
    }
    this.$nextTick(() => {
      const contentElement = document.querySelector('.api-content')
      if (contentElement) {
        contentElement.addEventListener('scroll', this.handleScroll)
      }
    })
    this.loadLocalIp()
  },
  beforeUnmount() {
    const contentElement = document.querySelector('.api-content')
    if (contentElement) {
      contentElement.removeEventListener('scroll', this.handleScroll)
    }
  },
  methods: {
    loadLocalIp() {
      const _this = this
      Base.BasePost('/api/BaseLogin', {}, function (res) {
        console.log('BaseLogin response:', res)
        if (res.ErrCode === 0 && res.Data && res.Data.local_ip) {
          console.log('Setting localIp:', res.Data.local_ip)
          _this.localIp = res.Data.local_ip
        }
      })
    },

    openDocumentPage() {
      const _this = this
      let url = `#/ApiDocument/${this.folderId}`
      if(_this.localIp !== ''){
        if (process.env.NODE_ENV === 'production') {
          url = 'http://' + _this.localIp + ':17170/' + url
        }else{
          url = 'http://' + _this.localIp + ':8080/' + url
        }
      }
      
      window.open(url, '_blank')
    },

    formatApiUrl(url) {
      if (!url) return url
      console.log('formatApiUrl - original:', url, 'localIp:', this.localIp)
      // Replace $Url$ with localIp if available
      if (this.localIp && url.includes('$Url$')) {
        const formatted = url.replace(/\$Url\$/g, this.localIp)
        console.log('formatApiUrl - formatted:', formatted)
        return formatted
      }
      // If no localIp, remove $Url$
      const formatted = url.replace(/\$Url\$/g, '')
      console.log('formatApiUrl - no localIp, formatted:', formatted)
      return formatted
    },

    getMethodTagType(method) {
      const types = {
        GET: 'success',
        POST: 'warning',
        PUT: 'primary',
        DELETE: 'danger',
        PATCH: 'info'
      }
      return types[method] || 'info'
    },

    getParsedHeaders(api) {
      if (!api || !api.headers) return {}
      try {
        return JSON.parse(api.headers)
      } catch {
        return {}
      }
    },

    getParsedQueryParams(api) {
      if (!api || !api.query_params) return []
      try {
        const params = JSON.parse(api.query_params)
        return Array.isArray(params) ? params : []
      } catch {
        return []
      }
    },

    getParsedBodyJson(api) {
      if (!api || !api.body_json) return {}
      try {
        return JSON.parse(api.body_json)
      } catch {
        return {}
      }
    },

    getParsedBodyForm(api) {
      if (!api || !api.body_form) return []
      try {
        const form = JSON.parse(api.body_form)
        return Array.isArray(form) ? form : []
      } catch {
        return []
      }
    },
    renderMarkdown(content) {
      if (!content) {
        return '';
      }

      // 配置 marked
      marked.setOptions({
        highlight: function(code, language) {
          const validLang = hljs.getLanguage(language) ? language : 'plaintext';
          return hljs.highlight(code, { language: validLang }).value;
        },
        langPrefix: 'hljs language-',
        breaks: true, // 转换换行符为 <br>
        gfm: true,    // 启用 GitHub Flavored Markdown
      });

      // 渲染 Markdown
      const html = marked(content);
      return `<div class="markdown-body">${html}</div>`;
    },
    getParsedResponseTake(api) {
      if (!api || !api.response_take) return []
      try {
        const take = JSON.parse(api.response_take)
        return Array.isArray(take) ? take : []
      } catch {
        return []
      }
    },
    getParsedTakeResult(api) {
      if (!api || !api.take_result) return []
      try {
        const take = JSON.parse(api.take_result)
        return Array.isArray(take) ? take : []
      } catch {
        return []
      }
    },

    hasBodyContent(api) {
      return (api &&
          ((api.content_type === 'application/json' && this.getParsedBodyJson(api) && Object.keys(this.getParsedBodyJson(api)).length > 0) ||
              ((api.content_type === 'application/x-www-form-urlencoded' || api.content_type === 'multipart/form-data') && this.getParsedBodyForm(api) && this.getParsedBodyForm(api).length > 0) ||
              (api.body_raw && api.body_raw !== '')))
    },

    hasLastResult(api) {
      return api && (api.last_result !== '') && api.last_result !== '{}'
    },

    formatJsonBody(body) {
      try {
        if (typeof body === 'string') {
          const parsed = JSON.parse(body)
          return JSON.stringify(parsed, null, 2)
        } else if (typeof body === 'object') {
          return JSON.stringify(body, null, 2)
        }
        return body
      } catch (e) {
        return body
      }
    },

    formatTimestamp(timestamp) {
      if (!timestamp) return '未知时间'
      return new Date(timestamp * 1000).toLocaleString()
    },

    scrollToApi(apiId) {
      const element = document.getElementById('api-' + apiId)
      if (element) {
        element.scrollIntoView({behavior: 'smooth'})
        this.activeApiId = apiId
      }
    },

    handleScroll() {
      const sections = document.querySelectorAll('.api-section')
      let current = null

      sections.forEach(section => {
        const rect = section.getBoundingClientRect()
        if (rect.top <= 100 && rect.bottom >= 100) {
          current = section.id.replace('api-', '')
        }
      })

      if (current) {
        this.activeApiId = parseInt(current)
      }
    },

    // 新增：一键复制所有接口信息为Markdown格式
    copyAllApisAsMarkdown() {
      if (!this.apis || this.apis.length === 0) {
        this.$message.warning('暂无接口可复制')
        return
      }

      const markdownLines = [];

      // 添加标题
      markdownLines.push(`## ${this.folderName} 接口文档\n`);

      // 遍历每个API生成Markdown内容
      this.apis.forEach((api, index) => {
        // API 标题
        markdownLines.push(`### ${index + 1}. ${api.name}`);
        markdownLines.push('');

        // 基本信息表格
        markdownLines.push('| 项目 | 详情 |');
        markdownLines.push('| --- | --- |');
        let apiUrl = this.formatApiUrl(api.url)
        markdownLines.push(`| 请求URL | \`${api.method}\` \`${apiUrl}\` |`);
        if (this.hasBodyContent(api)) {
          markdownLines.push(`| 请求类型 | \`${api.content_type}\` |`);
        }
        markdownLines.push(`| 创建时间 | ${this.formatTimestamp(api.create_time)} |`);
        markdownLines.push('');

        //描述
        if(api.desc !== ''){
          markdownLines.push(`备注 `);
          markdownLines.push(...api.desc.split('\n'));
          markdownLines.push('');
        }

        // 请求头
        // const headers = this.getParsedHeaders(api);
        // if (headers && Object.keys(headers).length > 0) {
        //   markdownLines.push('### 请求头');
        //   markdownLines.push('');
        //   markdownLines.push('| 字段 | 值 |');
        //   markdownLines.push('| --- | --- |');
        //   Object.entries(headers).forEach(([key, value]) => {
        //     markdownLines.push(`| ${key} | ${value} |`);
        //   });
        //   markdownLines.push('');
        // }

        // 请求参数
        const queryParams = this.getParsedQueryParams(api);
        if (queryParams && queryParams.length > 0) {
          markdownLines.push('请求参数');
          markdownLines.push('');
          markdownLines.push('| 字段名 | 类型 | 值 | 描述 |');
          markdownLines.push('| --- | --- | --- | --- |');
          queryParams.forEach(param => {
            markdownLines.push(`| ${param.field || ''} | ${param.type || ''} | ${param.value || ''} | ${Base.FormatEnterToMarkdown(param.description || '')} |`);
          });
          markdownLines.push('');
        }

        // 请求体
        if (this.hasBodyContent(api)) {
          markdownLines.push('请求体');
          markdownLines.push('');

          if (api.content_type === 'application/json' && this.getParsedBodyJson(api) && Object.keys(this.getParsedBodyJson(api)).length > 0) {
            markdownLines.push('json');
            markdownLines.push(this.formatJsonBody(this.getParsedBodyJson(api)));
            markdownLines.push('');
            markdownLines.push('');
          } else if ((api.content_type === 'application/x-www-form-urlencoded' || api.content_type === 'multipart/form-data') && this.getParsedBodyForm(api) && this.getParsedBodyForm(api).length > 0) {
            markdownLines.push('| 字段名 | 类型 | 值 | 描述 |');
            markdownLines.push('| --- | --- | --- | --- |');
            this.getParsedBodyForm(api).forEach(param => {
              markdownLines.push(`| ${param.field || ''} | ${param.type || ''} | ${param.value || ''} | ${Base.FormatEnterToMarkdown(param.description || '')} |`);
            });
            markdownLines.push('');
          } else if (api.body_raw && api.body_raw !== '') {
            markdownLines.push('\n' + api.body_raw + '\n```');
            markdownLines.push('');
          }
        }
        // 结果提取
        // const responseTake = this.getParsedResponseTake(api);
        // if (responseTake && responseTake.length > 0) {
        //   markdownLines.push('### 结果提取');
        //   markdownLines.push('');
        //   markdownLines.push('| 变量名 | 提取路径 | 提取方式 |');
        //   markdownLines.push('| --- | --- | --- |');
        //   responseTake.forEach(item => {
        //     markdownLines.push(`| ${item.key || ''} | ${item.path || ''} | ${item.type || 'default'} |`);
        //   });
        //   markdownLines.push('');
        // }

        // 结果说明文档
        const takeResultData = this.getParsedTakeResult(api);
        if (takeResultData && takeResultData.length > 0) {
          markdownLines.push('返回结果说明');
          markdownLines.push('');
          markdownLines.push('| 字段名 | 类型 | 说明 |');
          markdownLines.push('| --- | --- | --- |');
          takeResultData.forEach(item => {
            markdownLines.push(`| ${item.key || ''} | ${item.type || ''} | ${item.desc} |`);
          });
          markdownLines.push('');
        }

        // 请求结果
        // if (this.hasLastResult(api)) {
        //   api.last_result_data = JSON.parse(api.last_result)
        //   markdownLines.push('### 请求结果');
        //   markdownLines.push('');
        //   markdownLines.push('```json');
        //   markdownLines.push(this.formatJsonBody(api.last_result_data.result));
        //   markdownLines.push('```');
        //   markdownLines.push('');
        // }

        // 分隔线（除最后一个API外）
        if (index < this.apis.length - 1) {
          markdownLines.push('---');
          markdownLines.push('');
        }
      });

      const markdownText = markdownLines.join('\n');

      navigator.clipboard.writeText(markdownText).then(() => {
        this.$message.success('已复制所有接口信息为Markdown格式到剪贴板')
      }).catch(err => {
        console.error('复制失败:', err)
        this.$message.error('复制失败，请手动复制')
      })
    }
  }
}
</script>

<style scoped> .api-documentation {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: #f3f6f2;
}

.sidebar {
  width: 240px;
  background: #fff;
  border-right: 1px solid #e6ece0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sidebar-header {
  padding: 16px 12px;
  border-bottom: 1px solid #e6ece0;
  background: #f7f9f5;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sidebar-header h3 {
  margin: 0;
  padding: 0;
  font-size: 14px;
  color: #303133;
  font-weight: 600;
}

.sidebar-header :deep(.el-button) {
  width: 100%;
  display: block;
  margin-left: 0;
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 6px 0;
}

.api-menu {
  list-style: none;
  padding: 0;
  margin: 0;
}

.api-menu li {
  padding: 8px 16px;
  cursor: pointer;
  font-size: 13px; /* 字体缩小 */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  border-left: 2px solid transparent;
  transition: all 0.2s;
}

.api-menu li:hover {
  background-color: #f4faf2;
}

.api-menu li.active {
  background-color: #edf6ea;
  border-left: 2px solid #5a8a5a;
  color: #4f7d4f;
  font-weight: 500;
}

.api-content {
  flex: 1;
  overflow-y: auto;
  background: #f3f6f2;
}

/* 其余样式保持不变... */
.content-wrapper {
  padding: 20px;
}

.api-section {
  background: #fff;
  margin-bottom: 20px;
  padding: 20px;
  border: 1px solid #e8eee5;
  border-radius: 12px;
  box-shadow: 0 6px 18px rgba(80, 110, 80, 0.08);
}

.api-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e6ece0;
}

.method-tag {
  font-weight: bold;
  min-width: 60px;
  text-align: center;
}

.api-title {
  margin: 0;
  color: #303133;
  font-size: 20px;
}

.section-content {
  margin: 10px;
}

.section-content h4 {
  margin-top: 0;
  margin-bottom: 10px;
  color: #4f5f4b;
  border-bottom: 1px solid #e6ece0;
  padding-bottom: 5px;
}

.json-body {
  background: #2d2d2d;
  color: #f8f8f2;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #2f3a2f;
  overflow: auto;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  white-space: pre;
}

.raw-body {
  background: #f7f9f5;
  border: 1px solid #e6ece0;
  padding: 16px;
  border-radius: 8px;
  overflow: auto;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  white-space: pre-wrap;
}

.no-content {
  color: #909399;
  font-style: italic;
  padding: 10px 0;
}

.response-take-item {
  background: #f7f9f5;
  border: 1px solid #e6ece0;
  padding: 10px;
  border-radius: 8px;
  margin-bottom: 8px;
}

.response-take-item > div {
  margin-bottom: 4px;
}

.response-take-item > div:last-child {
  margin-bottom: 0;
}

.section-divider {
  height: 1px;
  background: #e6ece0;
  margin: 30px 0;
}

/* 滚动条样式 */
.sidebar::-webkit-scrollbar, .api-content::-webkit-scrollbar {
  width: 6px;
}

.sidebar::-webkit-scrollbar-track, .api-content::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.sidebar::-webkit-scrollbar-thumb, .api-content::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.sidebar::-webkit-scrollbar-thumb:hover, .api-content::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
} </style>

