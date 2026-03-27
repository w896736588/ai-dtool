<template>
  <el-form :model="localItem" label-width="180px" class="process-item-editor">
    <el-form-item label="名称" :error="fieldError('name')">
      <el-input v-model="localItem.name" />
    </el-form-item>

    <el-form-item label="类型" :error="fieldError('type')">
      <el-select v-model="localItem.type" placeholder="请选择类型" style="width: 100%">
        <el-option
          v-for="option in processTypeOptions"
          :key="option.value"
          :label="option.label"
          :value="option.value"
        />
      </el-select>
    </el-form-item>

    <el-form-item label="前端执行提示">
      <el-input v-model="localItem.tip" placeholder="可选，展示给执行中的用户提示" />
    </el-form-item>

    <template v-if="showField('locator')">
      <el-form-item :label="fieldLabel('locator')" :error="fieldError('locator')">
        <div class="list-editor">
          <div v-if="localItem.type === 'text_content'" class="locator-config-panel">
            <div
              v-for="(item, index) in formMeta.text_content_locators"
              :key="item.id"
              class="base-locator-card"
            >
              <div class="base-locator-card__header">
                <div class="base-locator-card__title">提取规则 {{ index + 1 }}</div>
                <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeTextContentLocator(index)">
                  删除
                </GitActionButton>
              </div>
              <div class="base-locator-card__summary">{{ describeBaseLocator(item.base_locator) }}</div>
              <div class="base-locator-card__footer">
                <GitActionButton compact size="small" native-type="button" @click="openBaseLocatorDialog('text_content', index, `编辑提取规则 ${index + 1}`)">
                  编辑定位
                </GitActionButton>
                <div class="base-locator-card__action">
                  <span class="base-locator-card__action-label">存在时</span>
                  <el-select v-model="item.on_found" size="small" class="base-locator-card__action-select">
                    <el-option label="返回其提取" value="extract_text" />
                    <el-option label="返回空值" value="return_empty" />
                  </el-select>
                </div>
              </div>
            </div>
            <GitActionButton compact size="small" native-type="button" @click="addTextContentLocator">
              新增提取规则
            </GitActionButton>
          </div>

          <div v-else-if="localItem.type === 'click' || localItem.type === 'input'" class="locator-config-panel">
            <div class="locator-config-panel__header">
              <div class="locator-config-panel__title">操作类型</div>
              <el-select v-model="formMeta.action_strategy" class="locator-config-panel__select">
                <el-option label="任意一个元素存在时执行" value="first_found_do_action" />
              </el-select>
            </div>
            <div
              v-for="(item, index) in formMeta.action_locators"
              :key="item.id"
              class="base-locator-card"
            >
              <div class="base-locator-card__header">
                <div class="base-locator-card__title">基础定位 {{ index + 1 }}</div>
                <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeActionLocator(index)">
                  删除
                </GitActionButton>
              </div>
              <div class="base-locator-card__summary">{{ describeBaseLocator(item.base_locator) }}</div>
              <GitActionButton compact size="small" native-type="button" @click="openBaseLocatorDialog('action', index, `编辑基础定位 ${index + 1}`)">
                编辑定位
              </GitActionButton>
            </div>
            <GitActionButton compact size="small" native-type="button" @click="addActionLocator">
              新增基础定位
            </GitActionButton>
          </div>

          <div v-else-if="useAdvancedLocatorEditor" class="structured-locator-editor">
            <div class="structured-locator-card">
              <div class="structured-locator-card__header">
                <div>
                  <div class="structured-locator-card__title">高级定位配置</div>
                  <div class="structured-locator-card__desc">
                    直接映射当前后端已支持的结构化 Locator 能力，可配置 has / has_not / 文本过滤 / chain。
                  </div>
                </div>
                <div class="structured-locator-card__badge">text_content</div>
              </div>

              <div class="structured-locator-section">
                <div class="structured-locator-section__title">1. 主元素</div>
                <div class="structured-locator-grid">
                  <el-select v-model="formMeta.locator_advanced_form.kind" placeholder="请选择查找方式">
                    <el-option
                      v-for="option in structuredLocatorKindOptions"
                      :key="`advanced-${option.value}`"
                      :label="option.label"
                      :value="option.value"
                    />
                  </el-select>
                  <el-input
                    v-model="formMeta.locator_advanced_form.value"
                    :placeholder="getStructuredLocatorPrimaryPlaceholder(formMeta.locator_advanced_form.kind)"
                  />
                </div>
              </div>

              <div class="structured-locator-section">
                <div class="structured-locator-section__title">2. 过滤条件</div>
                <div class="structured-locator-grid">
                  <el-input v-model="formMeta.locator_advanced_form.has_text" placeholder="包含文本，例如 已登录" />
                  <el-input v-model="formMeta.locator_advanced_form.has_not_text" placeholder="不包含文本，例如 去登录" />
                </div>
                <div class="structured-locator-grid structured-locator-grid--stacked">
                  <div class="structured-locator-filter-pair">
                    <el-select v-model="formMeta.locator_advanced_form.has_kind" placeholder="包含子元素的查找方式">
                      <el-option
                        v-for="option in structuredLocatorKindOptions"
                        :key="`has-${option.value}`"
                        :label="option.label"
                        :value="option.value"
                      />
                    </el-select>
                    <el-input v-model="formMeta.locator_advanced_form.has_value" placeholder="必须包含的子元素，例如 .profile-name" />
                  </div>
                  <div class="structured-locator-filter-pair">
                    <el-select v-model="formMeta.locator_advanced_form.has_not_kind" placeholder="不包含子元素的查找方式">
                      <el-option
                        v-for="option in structuredLocatorKindOptions"
                        :key="`has-not-${option.value}`"
                        :label="option.label"
                        :value="option.value"
                      />
                    </el-select>
                    <el-input v-model="formMeta.locator_advanced_form.has_not_value" placeholder="不能包含的子元素，例如 .btn.login_as_reg_btn" />
                  </div>
                </div>
                <div class="structured-locator-inline-field">
                  <div class="structured-locator-inline-field__label">
                    <div class="structured-locator-inline-field__title">可见性</div>
                    <div class="structured-locator-inline-field__desc">不限制时留空；仅当后端已有 visible 过滤需求时使用。</div>
                  </div>
                  <div class="structured-locator-inline-field__control">
                    <el-select v-model="formMeta.locator_advanced_form.visible" placeholder="不限制">
                      <el-option label="不限制" value="" />
                      <el-option label="必须可见" value="true" />
                      <el-option label="必须不可见" value="false" />
                    </el-select>
                  </div>
                </div>
              </div>

              <div class="structured-locator-section">
                <div class="structured-locator-section__title">3. 向下继续查找</div>
                <div class="structured-locator-grid">
                  <el-select v-model="formMeta.locator_advanced_form.chain_kind" placeholder="子节点查找方式">
                    <el-option
                      v-for="option in structuredLocatorKindOptions"
                      :key="`chain-${option.value}`"
                      :label="option.label"
                      :value="option.value"
                    />
                  </el-select>
                  <el-input v-model="formMeta.locator_advanced_form.chain_value" placeholder="可选，例如 .nickname" />
                </div>
              </div>

              <div class="structured-locator-section">
                <div class="structured-locator-section__title">4. 匹配规则</div>
                <div class="structured-locator-switches">
                  <div class="structured-locator-switch-card">
                    <div class="structured-locator-switch-card__label">文字完全一致</div>
                    <div class="structured-locator-switch-card__control">
                      <span class="structured-locator-switch-card__state">{{ formMeta.locator_advanced_form.exact ? '完全一致' : '允许模糊包含' }}</span>
                      <el-switch v-model="formMeta.locator_advanced_form.exact" />
                    </div>
                  </div>
                  <div class="structured-locator-switch-card">
                    <div class="structured-locator-switch-card__label">要求元素不存在</div>
                    <div class="structured-locator-switch-card__control">
                      <span class="structured-locator-switch-card__state">{{ formMeta.locator_advanced_form.negate ? '要求不存在' : '要求存在' }}</span>
                      <el-switch v-model="formMeta.locator_advanced_form.negate" />
                    </div>
                  </div>
                </div>

                <div class="structured-locator-grid">
                  <el-select v-model="formMeta.locator_advanced_form.pick_mode" placeholder="多个结果时怎么处理">
                    <el-option label="按默认方式处理" value="none" />
                    <el-option label="只取第一个" value="first" />
                    <el-option label="只取最后一个" value="last" />
                    <el-option label="取第 N 个" value="nth" />
                  </el-select>
                  <el-input-number
                    v-model="formMeta.locator_advanced_form.nth"
                    :min="0"
                    :controls="false"
                    class="plain-number-input"
                    :disabled="formMeta.locator_advanced_form.pick_mode !== 'nth'"
                  />
                </div>

                <div class="structured-locator-inline-field">
                  <div class="structured-locator-inline-field__label">
                    <div class="structured-locator-inline-field__title">最多等待多久</div>
                  </div>
                  <div class="structured-locator-inline-field__control">
                    <el-input-number
                      v-model="formMeta.locator_advanced_form.timeout_mills"
                      :min="0"
                      :controls="false"
                      class="plain-number-input"
                      placeholder="3000"
                    />
                    <span class="structured-locator-inline-field__unit">ms</span>
                  </div>
                </div>
              </div>

              <div class="structured-locator-preview">
                <div class="structured-locator-preview__title">系统生成的结构化配置</div>
                <el-input
                  :model-value="formMeta.locator_structured"
                  type="textarea"
                  :rows="10"
                  readonly
                />
              </div>
            </div>
          </div>

          <div v-else class="structured-locator-editor">
            <div class="structured-locator-card">
              <div class="structured-locator-card__header">
                <div>
                  <div class="structured-locator-card__title">定位配置</div>
                  <div class="structured-locator-card__desc">
                    按页面上你能看到的内容来选，系统会自动帮你生成底层定位配置。
                  </div>
                </div>
                <div class="structured-locator-card__badge">推荐</div>
              </div>

              <div class="structured-locator-section">
                <div class="structured-locator-section__title">1. 先选要找什么</div>
                <div class="structured-locator-grid">
                  <el-select v-model="formMeta.locator_structured_form.kind" placeholder="请选择查找方式">
                    <el-option
                      v-for="option in structuredLocatorKindOptions"
                      :key="option.value"
                      :label="option.label"
                      :value="option.value"
                    />
                  </el-select>
                  <el-input
                    v-model="structuredLocatorPrimaryValue"
                    :placeholder="structuredLocatorPrimaryPlaceholder"
                  />
                </div>

                <div v-if="showStructuredLocatorTargetText" class="structured-locator-grid structured-locator-grid--secondary">
                  <el-input
                    v-model="formMeta.locator_structured_form.target_text"
                    :placeholder="structuredLocatorTargetPlaceholder"
                  />
                </div>

                <div class="structured-locator-tip">
                  {{ structuredLocatorScenarioHint }}
                </div>
              </div>

              <div class="structured-locator-section">
                <div class="structured-locator-section__title">2. 再决定匹配规则</div>
                <div class="structured-locator-switches">
                  <div class="structured-locator-switch-card">
                    <div class="structured-locator-switch-card__label">文字要不要完全一致</div>
                    <div class="structured-locator-switch-card__control">
                      <span class="structured-locator-switch-card__state">{{ formMeta.locator_structured_form.exact ? '完全一致' : '允许模糊包含' }}</span>
                      <el-switch v-model="formMeta.locator_structured_form.exact" />
                    </div>
                  </div>
                  <div class="structured-locator-switch-card">
                    <div class="structured-locator-switch-card__label">这个元素必须存在吗</div>
                    <div class="structured-locator-switch-card__control">
                      <span class="structured-locator-switch-card__state">{{ formMeta.locator_structured_form.negate ? '要求不存在' : '要求存在' }}</span>
                      <el-switch v-model="formMeta.locator_structured_form.negate" />
                    </div>
                  </div>
                </div>

                <div class="structured-locator-inline-field">
                  <div class="structured-locator-inline-field__label">
                    <div class="structured-locator-inline-field__title">最多等待多久</div>
                    <div class="structured-locator-inline-field__desc">通常保持默认即可，页面慢时再适当调大。</div>
                  </div>
                  <div class="structured-locator-inline-field__control">
                    <el-input-number
                      v-model="formMeta.locator_structured_form.timeout_mills"
                      :min="0"
                      :controls="false"
                      class="plain-number-input"
                      placeholder="3000"
                    />
                    <span class="structured-locator-inline-field__unit">ms</span>
                  </div>
                </div>
              </div>

              <div class="structured-locator-section">
                <div class="structured-locator-section__title">3. 如果找到多个结果，怎么处理</div>
                <div class="structured-locator-grid">
                  <el-select v-model="formMeta.locator_structured_form.pick_mode" placeholder="多个结果时怎么处理">
                    <el-option label="按默认方式处理" value="none" />
                    <el-option label="只取第一个" value="first" />
                    <el-option label="只取最后一个" value="last" />
                    <el-option label="取第 N 个" value="nth" />
                  </el-select>
                  <div class="structured-locator-inline-field structured-locator-inline-field--compact">
                    <div class="structured-locator-inline-field__label">
                      <div class="structured-locator-inline-field__title">第几个结果</div>
                    </div>
                    <div class="structured-locator-inline-field__control">
                      <el-input-number
                        v-model="formMeta.locator_structured_form.nth"
                        :min="0"
                        :controls="false"
                        class="plain-number-input"
                        :disabled="formMeta.locator_structured_form.pick_mode !== 'nth'"
                      />
                    </div>
                  </div>
                </div>
              </div>

              <div class="structured-locator-preview">
                <div class="structured-locator-preview__title">系统生成的结构化配置</div>
                <el-input
                  :model-value="formMeta.locator_structured"
                  type="textarea"
                  :rows="8"
                  readonly
                />
              </div>
            </div>
          </div>

        </div>
      </el-form-item>
    </template>

    <template v-if="showField('secondary_locator')">
      <el-form-item :label="fieldLabel('secondary_locator')" :error="fieldError('secondary_locator')">
        <el-input v-model="formMeta.secondary_locator" :placeholder="fieldPlaceholder('secondary_locator')" />
        <div class="field-guide">{{ fieldGuide('secondary_locator') }}</div>
      </el-form-item>
    </template>

    <template v-if="showField('tertiary_locator')">
      <el-form-item :label="fieldLabel('tertiary_locator')" :error="fieldError('tertiary_locator')">
        <el-input v-model="formMeta.tertiary_locator" :placeholder="fieldPlaceholder('tertiary_locator')" />
        <div class="field-guide">{{ fieldGuide('tertiary_locator') }}</div>
      </el-form-item>
    </template>

    <template v-if="localItem.type === 'redirect_uri'">
      <el-form-item label="跳转地址" :error="fieldError('value')">
        <el-input v-model="formMeta.value" placeholder="例如 /login 或 https://example.com/login" />
        <div class="field-guide">{{ fieldGuide('value') }}</div>
      </el-form-item>

      <el-form-item label="跳转后等待地址" :error="fieldError('register_response_urls')">
        <div class="list-editor">
          <div v-for="(item, index) in formMeta.register_response_urls" :key="item.uid" class="response-url-row">
            <el-input v-model="item.url" placeholder="等待地址，例如 /home 或 https://example.com/home" />
            <el-input-number v-model="item.wait_second" :min="1" :controls="false" class="plain-number-input" />
            <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeRegisterResponseUrl(index)">
              删除
            </GitActionButton>
          </div>
          <GitActionButton compact size="small" native-type="button" @click="addRegisterResponseUrl">
            新增等待地址
          </GitActionButton>
          <div class="field-guide">{{ fieldGuide('register_response_urls') }}</div>
        </div>
      </el-form-item>
    </template>

    <template v-else-if="showField('value')">
      <el-form-item :label="fieldLabel('value')" :error="fieldError('value')">
        <el-input
          v-model="formMeta.value"
          :placeholder="fieldPlaceholder('value')"
          type="textarea"
          :rows="textareaRows('value')"
        />
        <div class="field-guide">{{ fieldGuide('value') }}</div>
      </el-form-item>
    </template>

    <template v-if="localItem.type === 'wait_url'">
      <el-form-item label="等待地址" :error="fieldError('response_url')">
        <el-input v-model="formMeta.response_url" :placeholder="fieldPlaceholder('response_url')" />
        <div class="field-guide">{{ fieldGuide('response_url') }}</div>
      </el-form-item>

      <el-form-item label="等待秒数" :error="fieldError('wait_second')">
        <el-input-number v-model="formMeta.wait_second" :min="1" :controls="false" class="plain-number-input" />
      </el-form-item>
    </template>

    <template v-else>
      <template v-if="showField('wait_second')">
        <el-form-item :label="fieldLabel('wait_second')" :error="fieldError('wait_second')">
          <el-input-number v-model="formMeta.wait_second" :min="1" :controls="false" class="plain-number-input" />
        </el-form-item>
      </template>

      <template v-if="showField('response_url')">
        <el-form-item :label="fieldLabel('response_url')" :error="fieldError('response_url')">
          <el-input v-model="formMeta.response_url" :placeholder="fieldPlaceholder('response_url')" />
          <div class="field-guide">{{ fieldGuide('response_url') }}</div>
        </el-form-item>
      </template>

      <template v-if="localItem.type !== 'redirect_uri' && showField('register_response_urls')">
        <el-form-item :label="fieldLabel('register_response_urls')" :error="fieldError('register_response_urls')">
          <div class="list-editor">
            <div v-for="(item, index) in formMeta.register_response_urls" :key="item.uid" class="response-url-row">
              <el-input v-model="item.url" placeholder="等待地址" />
              <el-input-number v-model="item.wait_second" :min="1" :controls="false" class="plain-number-input" />
              <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeRegisterResponseUrl(index)">
                删除
              </GitActionButton>
            </div>
            <GitActionButton compact size="small" native-type="button" @click="addRegisterResponseUrl">
              新增等待地址
            </GitActionButton>
            <div class="field-guide">{{ fieldGuide('register_response_urls') }}</div>
          </div>
        </el-form-item>
      </template>
    </template>

    <template v-if="showField('out_key') && localItem.type !== 'input'">
      <el-form-item :label="fieldLabel('out_key')" :error="fieldError('out_key')">
        <el-input v-model="formMeta.out_key" placeholder="例如 {login_state}" />
        <div class="field-guide">{{ fieldGuide('out_key') }}</div>
      </el-form-item>
    </template>

    <template v-if="showField('check_key')">
      <el-form-item :label="fieldLabel('check_key')" :error="fieldError('check_key')">
        <div class="list-editor">
          <el-select v-model="formMeta.check_mode" class="check-mode-select">
            <el-option label="可以执行" value="none" />
            <el-option label="按前面结果判断" value="bool" />
            <el-option label="按内容比较判断" value="compare" />
          </el-select>

          <template v-if="formMeta.check_mode === 'bool'">
            <div v-for="(item, index) in formMeta.check_rule_list" :key="item.uid" class="check-rule-row">
              <el-select v-model="item.key" filterable placeholder="请选择前面节点的输出">
                <el-option
                  v-for="option in checkKeyOptions"
                  :key="`${option.value}-${option.label}`"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <el-select v-model="item.expect" class="check-rule-row__mode">
                <el-option label="必须为真" value="true" />
                <el-option label="必须为假" value="false" />
              </el-select>
              <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeCheckRule(index)">
                删除
              </GitActionButton>
            </div>
            <GitActionButton
              compact
              size="small"
              native-type="button"
              :disabled="checkKeyOptions.length === 0 || usedCheckKeyCount >= checkKeyOptions.length"
              @click="addCheckRule"
            >
              新增判断条件
            </GitActionButton>
          </template>

          <template v-if="formMeta.check_mode === 'compare'">
            <div class="compare-rule-row">
              <el-select v-model="formMeta.compare_rule.left" filterable placeholder="请选择左侧输出">
                <el-option
                  v-for="option in checkKeyOptions"
                  :key="`left-${option.value}-${option.label}`"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <el-select v-model="formMeta.compare_rule.operator" class="compare-rule-row__operator">
                <el-option label="等于" value="==" />
                <el-option label="不等于" value="!=" />
              </el-select>
              <div class="compare-rule-right">
                <el-input
                  v-model="formMeta.compare_rule.right"
                  placeholder="请输入固定字符串，或点击下方快捷填入注入值"
                />
                <div class="compare-rule-right__actions">
                  <GitActionButton
                    v-for="option in compareRightOptions"
                    :key="`right-${option.value}`"
                    compact
                    size="small"
                    native-type="button"
                    @click="applyCompareRightQuickPick(option.value)"
                  >
                    {{ option.label }}
                  </GitActionButton>
                </div>
              </div>
            </div>
          </template>
        </div>
        <div class="field-guide">{{ fieldGuide('check_key') }}</div>
      </el-form-item>
    </template>

    <template v-if="showField('wait_count')">
      <el-form-item :label="fieldLabel('wait_count')" :error="fieldError('wait_count')">
        <el-input-number v-model="formMeta.wait_count" :min="1" :controls="false" class="plain-number-input" />
      </el-form-item>
    </template>

    <template v-if="showField('delete_mode')">
      <el-form-item :label="fieldLabel('delete_mode')">
        <el-select v-model="formMeta.delete_mode" style="width: 100%">
          <el-option label="按 class 删除" value="class" />
        </el-select>
      </el-form-item>
    </template>

    <template v-if="showField('bool_result_rules')">
      <el-form-item label="主元素定位规则" :error="fieldError('bool_result_rules')">
        <div class="list-editor">
          <div v-for="(item, index) in formMeta.bool_result_rules" :key="item.uid" class="bool-result-rule-card">
            <div class="bool-result-rule-card__header">
              <div class="bool-result-rule-card__title">规则 {{ index + 1 }}</div>
              <div class="bool-result-rule-card__actions">
                <el-select v-model="item.on_found" class="bool-result-rule-card__result">
                  <el-option label="命中返回 true" :value="true" />
                  <el-option label="命中返回 false" :value="false" />
                </el-select>
                <GitActionButton compact size="small" variant="danger" native-type="button" @click="removeBoolResultRule(index)">
                  删除
                </GitActionButton>
              </div>
            </div>
            <div class="base-locator-card__summary">{{ describeBaseLocator(item.base_locator) }}</div>
            <GitActionButton compact size="small" native-type="button" @click="openBaseLocatorDialog('bool_result', index, `编辑规则 ${index + 1} 定位`)">
              编辑定位
            </GitActionButton>
          </div>

          <GitActionButton compact size="small" native-type="button" @click="addBoolResultRule">
            新增基础定位
          </GitActionButton>
          <div class="field-guide">{{ fieldGuide('bool_result_rules') }}</div>
        </div>
      </el-form-item>
    </template>

    <el-form-item label="权重">
      <el-input-number v-model="localItem.weight" :min="0" />
    </el-form-item>

    <el-form-item label="等待时长(ms)" :error="fieldError('wait_mills')">
      <el-input-number v-model="localItem.wait_mills" :min="0" />
    </el-form-item>

    <el-form-item label="域名限制" :error="fieldError('domain_limit')">
      <el-input v-model="localItem.domain_limit" placeholder="可选，例如 example.com" />
      <div class="field-guide">{{ fieldGuide('domain_limit') }}</div>
    </el-form-item>

    <el-form-item v-if="allowAppendToReplace" label="输出追加到替换列表">
      <el-select v-model="localItem.append_to_replace" style="width: 100%">
        <el-option label="追加" value="1" />
        <el-option label="不追加" value="0" />
      </el-select>
    </el-form-item>

    <el-form-item label="执行方式">
      <el-select v-model="localItem.is_async" style="width: 100%">
        <el-option label="同步" value="0" />
        <el-option label="异步" value="1" />
      </el-select>
    </el-form-item>

    <el-form-item label="出错后是否继续">
      <el-select v-model="localItem.is_error_continue" style="width: 100%">
        <el-option label="中断" value="0" />
        <el-option label="继续" value="1" />
      </el-select>
    </el-form-item>
  </el-form>
  <el-dialog v-model="baseLocatorDialog.visible" :title="baseLocatorDialog.title || '编辑基础定位'" width="60%">
    <div class="structured-locator-editor">
      <div class="structured-locator-card">
        <div class="structured-locator-card__header">
          <div>
            <div class="structured-locator-card__title">基础定位配置</div>
            <div class="structured-locator-card__desc">保存后会折叠成按钮展示，避免占用过多空间。</div>
          </div>
        </div>

        <div class="structured-locator-inline-field">
          <div class="structured-locator-inline-field__label">
            <div class="structured-locator-inline-field__title">编辑模式</div>
            <div class="structured-locator-inline-field__desc">普通场景优先用简单模式；复杂页面再切高级模式。</div>
          </div>
          <div class="structured-locator-inline-field__control">
            <el-select v-model="baseLocatorDialog.draft.locator_editor_mode">
              <el-option label="简单模式" value="simple" />
              <el-option label="高级模式" value="advanced" />
            </el-select>
          </div>
        </div>

        <template v-if="baseLocatorDialog.draft.locator_editor_mode === 'advanced'">
          <div class="structured-locator-section">
            <div class="structured-locator-section__title">1. 主元素</div>
            <div class="structured-locator-grid">
              <el-select v-model="baseLocatorDialog.draft.locator_advanced_form.kind" placeholder="请选择查找方式">
                <el-option
                  v-for="option in structuredLocatorKindOptions"
                  :key="`dialog-advanced-${option.value}`"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <el-input
                v-model="baseLocatorDialog.draft.locator_advanced_form.value"
                :placeholder="getStructuredLocatorPrimaryPlaceholder(baseLocatorDialog.draft.locator_advanced_form.kind)"
              />
            </div>
          </div>
          <div class="structured-locator-section">
            <div class="structured-locator-section__title">2. 过滤条件</div>
            <div class="structured-locator-grid">
              <el-input v-model="baseLocatorDialog.draft.locator_advanced_form.has_text" placeholder="包含文本" />
              <el-input v-model="baseLocatorDialog.draft.locator_advanced_form.has_not_text" placeholder="不包含文本" />
            </div>
            <div class="structured-locator-grid structured-locator-grid--stacked">
              <div class="structured-locator-filter-pair">
                <el-select v-model="baseLocatorDialog.draft.locator_advanced_form.has_kind" placeholder="包含子元素查找方式">
                  <el-option
                    v-for="option in structuredLocatorKindOptions"
                    :key="`dialog-has-${option.value}`"
                    :label="option.label"
                    :value="option.value"
                  />
                </el-select>
                <el-input v-model="baseLocatorDialog.draft.locator_advanced_form.has_value" placeholder="必须包含的子元素" />
              </div>
              <div class="structured-locator-filter-pair">
                <el-select v-model="baseLocatorDialog.draft.locator_advanced_form.has_not_kind" placeholder="不包含子元素查找方式">
                  <el-option
                    v-for="option in structuredLocatorKindOptions"
                    :key="`dialog-has-not-${option.value}`"
                    :label="option.label"
                    :value="option.value"
                  />
                </el-select>
                <el-input v-model="baseLocatorDialog.draft.locator_advanced_form.has_not_value" placeholder="不能包含的子元素" />
              </div>
            </div>
          </div>
        </template>

        <template v-else>
          <div class="structured-locator-section">
            <div class="structured-locator-section__title">1. 先选要找什么</div>
            <div class="structured-locator-grid">
              <el-select v-model="baseLocatorDialog.draft.locator_structured_form.kind" placeholder="请选择查找方式">
                <el-option
                  v-for="option in structuredLocatorKindOptions"
                  :key="`dialog-${option.value}`"
                  :label="option.label"
                  :value="option.value"
                />
              </el-select>
              <el-input
                v-model="baseLocatorDialog.draft.locator_structured_form.value"
                :placeholder="getStructuredLocatorPrimaryPlaceholder(baseLocatorDialog.draft.locator_structured_form.kind)"
              />
            </div>
          </div>
        </template>
      </div>
    </div>
    <template #footer>
      <GitActionButton @click="baseLocatorDialog.visible = false">取消</GitActionButton>
      <GitActionButton @click="saveBaseLocatorDialog">保存定位</GitActionButton>
    </template>
  </el-dialog>
</template>

<script>
import GitActionButton from '@/components/base/GitActionButton.vue'

const {
  PROCESS_ITEM_FIELD_GUIDES,
  PROCESS_TYPE_FIELDS,
  parseCheckConfig,
  parseRedirectUriValue,
  parseWaitUrlValue,
  serializeCheckConfig,
  serializeRedirectUriValue,
  serializeWaitUrlValue,
  validateProcessItemForm,
} = require('../../utils/smart_link_process_validation.cjs')
const {
  buildAdvancedLocatorPayload,
  buildSimpleLocatorPayload,
  createAdvancedLocatorForm,
  createSimpleLocatorForm,
  deserializeLocatorEditorState,
  parseStructuredLocatorPayload,
  stringifyLocatorPayload,
} = require('../../utils/smart_link_locator_form.cjs')
const {
  buildLocatorConfigByType,
  createBaseLocatorMeta,
  deserializeLocatorConfigToFormMeta,
  isLocatorConfigPayload,
} = require('../../utils/smart_link_locator_config.cjs')
const {
  formatStructuredLocator,
} = require('../../utils/smart_link_process_display.cjs')

const createDefaultItem = () => ({
  id: 0,
  name: '',
  smart_link_process_id: 0,
  type: '',
  locator: '',
  wait_mills: 3000,
  tip: '',
  value: '',
  out_key: '',
  check_key: '',
  weight: 0,
  domain_limit: '',
  append_to_replace: '0',
  is_async: '0',
  is_error_continue: '0',
  next_ids: '',
  x: 0,
  y: 0,
})

const PROCESS_TYPE_OPTIONS = [
  { label: '提取元素内容 text_content', value: 'text_content' },
  { label: '跳转 redirect_uri', value: 'redirect_uri' },
  { label: '等待接口完成 wait_url', value: 'wait_url' },
  { label: '等待毫秒 wait', value: 'wait' },
  { label: '判断输出 bool_result', value: 'bool_result' },
  { label: '判断存在 bool_exist', value: 'bool_exist' },
  { label: '点击元素 click', value: 'click' },
  { label: '输入信息 input', value: 'input' },
  { label: '关闭页面 close', value: 'close' },
  { label: '不存在时等待 no_exist_wait', value: 'no_exist_wait' },
  { label: '提取 canvas 图片 canvas_image', value: 'canvas_image' },
  { label: '输入账号密码 login_username_password', value: 'login_username_password' },
  { label: '删除元素 delete_element', value: 'delete_element' },
]

function safeParseJson(text, fallback) {
  if (!text) return fallback
  try {
    return JSON.parse(text)
  } catch (error) {
    return fallback
  }
}

function createRegisterUrl() {
  return {
    uid: `response-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    url: '',
    wait_second: 10,
  }
}

function createLocatorRow() {
  return {
    uid: `locator-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    value: '',
    exist_mode: 'exist',
    match_mode: 'all',
  }
}

function createBoolResultRule() {
  return {
    uid: `bool-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    id: `bool-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    base_locator: createBaseLocatorMeta(),
    on_found: true,
    locator_structured: '',
    locator_structured_form: createStructuredLocatorForm(),
    locator_advanced_form: createAdvancedStructuredLocatorForm(),
    return: true,
  }
}

function createActionLocator(role = '') {
  return {
    id: `base-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    role,
    base_locator: createBaseLocatorMeta(),
  }
}

function createTextContentRule() {
  return {
    id: `text-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    on_found: 'extract_text',
    base_locator: createBaseLocatorMeta(),
  }
}

function createCheckRule() {
  return {
    uid: `check-${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
    key: '',
    expect: 'true',
  }
}

function createCompareRule() {
  return {
    left: '',
    operator: '==',
    right: '',
  }
}

function createStructuredLocatorForm() {
  return createSimpleLocatorForm()
}

function createAdvancedStructuredLocatorForm() {
  return createAdvancedLocatorForm()
}

// normalizeTokenLabel 用于统一下拉标签展示，兼容已带大括号的旧 out_key。
// normalizeTokenLabel normalizes option labels and preserves legacy wrapped out_key values.
function normalizeTokenLabel(value) {
  const normalizedValue = String(value || '').trim()
  if (!normalizedValue) return ''
  return normalizedValue.startsWith('{') && normalizedValue.endsWith('}')
    ? normalizedValue
    : `{${normalizedValue}}`
}

function withRegisterUid(list) {
  return (Array.isArray(list) ? list : []).map((item) => ({
    uid: item.uid || createRegisterUrl().uid,
    url: item.url || '',
    wait_second: Number(item.wait_second || 10),
  }))
}

function withCheckRuleUid(list) {
  return (Array.isArray(list) ? list : []).map((item) => ({
    uid: item.uid || createCheckRule().uid,
    key: item.key || '',
    expect: item.expect || 'true',
  }))
}

export default {
  name: 'ProcessItemEditor',
  components: {
    GitActionButton,
  },
  props: {
    modelValue: {
      type: Object,
      default: () => createDefaultItem(),
    },
    processItemOptions: {
      type: Array,
      default: () => [],
    },
  },
  emits: ['update:modelValue'],
  data() {
    return {
      localItem: createDefaultItem(),
      formMeta: {
        locator_list: [],
        locator_joiner: 'structured',
        locator_raw: '',
        locator_editor_mode: 'simple',
        locator_structured: '',
        locator_structured_form: createStructuredLocatorForm(),
        locator_advanced_form: createAdvancedStructuredLocatorForm(),
        secondary_locator: '',
        tertiary_locator: '',
        value: '',
        out_key: '',
        check_key: '',
        check_mode: 'none',
        check_rule_list: [],
        compare_rule: createCompareRule(),
        wait_second: 10,
        wait_count: 3,
        response_url: '',
        delete_mode: 'class',
        register_response_urls: [],
        bool_result_rules: [],
        text_content_locators: [createTextContentRule()],
        action_strategy: 'first_found_do_action',
        action_locators: [createActionLocator()],
      },
      baseLocatorDialog: {
        visible: false,
        targetType: '',
        targetIndex: -1,
        title: '',
        draft: createBaseLocatorMeta(),
      },
      syncingFromParent: false,
      lastSerializedSignature: '',
      fieldErrors: {},
      processTypeOptions: PROCESS_TYPE_OPTIONS,
    }
  },
  computed: {
    currentFields() {
      return PROCESS_TYPE_FIELDS[this.localItem.type] || []
    },
    allowAppendToReplace() {
      return this.localItem.type !== 'click' && this.localItem.type !== 'delete_element'
    },
    useLocatorExpressionEditor() {
      return this.showField('locator') && this.supportLocatorExpression(this.localItem.type)
    },
    structuredLocatorKindOptions() {
      return [
        { label: '按按钮文字查找', value: 'button_text' },
        { label: '按页面文字查找', value: 'text' },
        { label: '按输入框标签查找', value: 'label' },
        { label: '按占位提示查找', value: 'placeholder' },
        { label: '按图片说明文字查找', value: 'alt_text' },
        { label: '按标题提示查找', value: 'title' },
        { label: '按测试标识查找', value: 'test_id' },
        { label: '按 CSS / XPath 查找', value: 'css' },
      ]
    },
    showStructuredLocatorTargetText() {
      return this.formMeta.locator_structured_form.kind === 'button_text'
    },
    useAdvancedLocatorEditor() {
      return this.localItem.type === 'text_content'
    },
    useLocatorConfigEditor() {
      return this.localItem.type === 'text_content' || this.localItem.type === 'click' || this.localItem.type === 'input'
    },
    structuredLocatorPrimaryPlaceholder() {
      const kind = this.formMeta.locator_structured_form.kind
      if (kind === 'button_text') return '例如 提交、登录、确定'
      if (kind === 'text') return '例如 登录成功、请选择门店'
      if (kind === 'label') return '例如 用户名、手机号'
      if (kind === 'placeholder') return '例如 请输入用户名'
      if (kind === 'alt_text') return '例如 商品图片、头像'
      if (kind === 'title') return '例如 复制、刷新'
      if (kind === 'test_id') return '例如 submit-btn'
      return '例如 .dialog .confirm-btn 或 //button[text()="提交"]'
    },
    structuredLocatorTargetPlaceholder() {
      return '如果不是普通按钮，可改成链接、菜单项等显示文字'
    },
    structuredLocatorScenarioHint() {
      return this.getStructuredLocatorScenarioHint(this.formMeta.locator_structured_form.kind)
    },
    structuredLocatorPrimaryValue: {
      get() {
        const form = this.formMeta.locator_structured_form || createStructuredLocatorForm()
        return form.value
      },
      set(value) {
        this.formMeta.locator_structured_form.value = value
      },
    },
    useStructuredLocatorTextarea() {
      return true
    },
    checkKeyOptions() {
      const processList = Array.isArray(this.processItemOptions) ? this.processItemOptions : []
      const currentIndex = processList.findIndex(item => String(item.id) === String(this.localItem.id))
      const availableList = currentIndex >= 0 ? processList.slice(0, currentIndex) : processList
      return availableList
        .filter(item => String(item.out_key || '').trim() !== '')
        .map(item => ({
          value: String(item.out_key || '').trim(),
          label: `${item.name || item.type || '未命名节点'}.${normalizeTokenLabel(item.out_key)}`,
        }))
    },
    compareRightOptions() {
      return [
        { value: '{user_name}', label: '当前账号用户名 {user_name}' },
        { value: '{password}', label: '当前账号密码 {password}' },
      ]
    },
    showTextContentLocatorSummary() {
      return false
    },
    usedCheckKeyCount() {
      return (Array.isArray(this.formMeta.check_rule_list) ? this.formMeta.check_rule_list : [])
        .map(item => String(item && item.key || '').trim())
        .filter(Boolean)
        .length
    },
  },
  watch: {
    modelValue: {
      deep: true,
      immediate: true,
      handler(value) {
        this.syncFromModel(value || createDefaultItem())
      },
    },
    localItem: {
      deep: true,
      handler() {
        this.emitChange()
      },
    },
    formMeta: {
      deep: true,
      handler() {
        this.emitChange()
      },
    },
    'localItem.type'(nextType, prevType) {
      if (!this.syncingFromParent && nextType !== prevType) {
        this.resetMetaForType(nextType)
      }
    },
    'formMeta.locator_structured_form': {
      deep: true,
      handler() {
        if (this.syncingFromParent || this.useAdvancedLocatorEditor) return
        this.formMeta.locator_structured = this.stringifyStructuredLocatorPayload(this.buildStructuredLocatorPayload())
      },
    },
    'formMeta.locator_advanced_form': {
      deep: true,
      handler() {
        if (this.syncingFromParent || !this.useAdvancedLocatorEditor) return
        this.formMeta.locator_structured = this.stringifyStructuredLocatorPayload(this.buildAdvancedLocatorPayload())
      },
    },
  },
  methods: {
    createSignature(payload) {
      return JSON.stringify(payload || {})
    },
    syncFromModel(value) {
      const normalizedValue = {
        ...createDefaultItem(),
        ...JSON.parse(JSON.stringify(value || {})),
        next_ids: (value && value.next_ids) || '',
        append_to_replace: String((value && value.append_to_replace) ?? '0'),
        is_async: String((value && value.is_async) ?? '0'),
        is_error_continue: String((value && value.is_error_continue) ?? '0'),
      }
      const incomingSignature = this.createSignature(normalizedValue)
      if (!this.syncingFromParent && incomingSignature === this.lastSerializedSignature) {
        return
      }
      this.syncingFromParent = true
      this.localItem = normalizedValue
      this.formMeta = this.deserializeMeta(this.localItem)
      this.fieldErrors = {}
      this.lastSerializedSignature = this.createSignature(this.serializeItem())
      this.$nextTick(() => {
        this.syncingFromParent = false
      })
    },
    resetMetaForType(type) {
      this.formMeta = this.deserializeMeta({
        ...this.localItem,
        type,
        locator: '',
        value: '',
        out_key: '',
        check_key: '',
      })
      this.fieldErrors = {}
    },
    deserializeMeta(item) {
      // deserializeMeta 负责把后端存储结构转换成前端表单状态。
      // deserializeMeta converts backend payloads into editable form state.
      const meta = {
        locator_list: [],
        locator_joiner: 'structured',
        locator_raw: '',
        locator_editor_mode: item.type === 'text_content' ? 'advanced' : 'simple',
        locator_structured: '',
        locator_structured_form: createStructuredLocatorForm(),
        locator_advanced_form: createAdvancedStructuredLocatorForm(),
        secondary_locator: '',
        tertiary_locator: '',
        value: item.value || '',
        out_key: item.out_key || '',
        check_key: item.check_key || '',
        check_mode: 'none',
        check_rule_list: [],
        compare_rule: createCompareRule(),
        wait_second: 10,
        wait_count: 3,
        response_url: '',
        delete_mode: item.value || 'class',
        register_response_urls: [],
        bool_result_rules: [],
        text_content_locators: [createTextContentRule()],
        action_strategy: 'first_found_do_action',
        action_locators: [createActionLocator()],
      }

      if (item.type === 'bool_result' && isLocatorConfigPayload(item.locator)) {
        const configMeta = deserializeLocatorConfigToFormMeta(item.locator) || {}
        meta.bool_result_rules = Array.isArray(configMeta.bool_result_rules)
          ? configMeta.bool_result_rules.map((rule) => ({
            ...createBoolResultRule(),
            id: rule.id || '',
            on_found: rule.on_found !== false,
            return: rule.on_found !== false,
            base_locator: rule.base_locator || createBaseLocatorMeta(),
          }))
          : []
      } else if (item.type === 'bool_result') {
        const rules = safeParseJson(item.locator, [])
        meta.bool_result_rules = Array.isArray(rules)
          ? rules.map(rule => {
            const baseRule = createBoolResultRule()
            const structuredLocator = parseStructuredLocatorPayload(rule && rule.locator)
            return {
              ...baseRule,
              locator_structured: structuredLocator ? JSON.stringify(structuredLocator, null, 2) : '',
              locator_structured_form: structuredLocator ? this.deserializeStructuredLocatorForm(structuredLocator) : createStructuredLocatorForm(),
              locator_advanced_form: structuredLocator ? this.deserializeAdvancedLocatorForm(structuredLocator) : createAdvancedStructuredLocatorForm(),
              return: rule.return !== false,
            }
          })
          : []
      } else if (item.type === 'text_content' && isLocatorConfigPayload(item.locator)) {
        const configMeta = deserializeLocatorConfigToFormMeta(item.locator) || {}
        meta.text_content_locators = Array.isArray(configMeta.text_content_locators) && configMeta.text_content_locators.length > 0
          ? configMeta.text_content_locators
          : [createTextContentRule()]
      } else if (item.type === 'wait_url') {
        Object.assign(meta, parseWaitUrlValue(item.value))
      } else if (item.type === 'redirect_uri') {
        const redirectMeta = parseRedirectUriValue(item.value)
        meta.value = redirectMeta.value
        meta.register_response_urls = withRegisterUid(redirectMeta.register_response_urls)
      } else if (item.type === 'no_exist_wait') {
        const [waitSecond, waitCount] = String(item.value || '').split('|')
        meta.wait_second = Number(waitSecond || 10)
        meta.wait_count = Number(waitCount || 3)
        const structuredLocator = parseStructuredLocatorPayload(item.locator)
        if (structuredLocator) {
          meta.locator_structured_form = this.deserializeStructuredLocatorForm(structuredLocator)
          meta.locator_structured = JSON.stringify(structuredLocator, null, 2)
        }
      } else if (item.type === 'login_username_password') {
        const parts = String(item.locator || '').split('||')
        meta.locator_list = parts[0] ? [{ uid: createLocatorRow().uid, value: parts[0] }] : [createLocatorRow()]
        meta.secondary_locator = parts[1] || ''
        meta.tertiary_locator = parts[2] || ''
      } else if (item.type === 'delete_element') {
        meta.locator_list = this.decodeLocatorList(item.locator, '|')
      } else if ((item.type === 'click' || item.type === 'input') && isLocatorConfigPayload(item.locator)) {
        const configMeta = deserializeLocatorConfigToFormMeta(item.locator) || {}
        meta.action_strategy = configMeta.action_strategy || 'first_found_do_action'
        meta.action_locators = Array.isArray(configMeta.action_locators) && configMeta.action_locators.length > 0
          ? configMeta.action_locators
          : [createActionLocator()]
      } else if (this.showTypeField(item.type, 'locator')) {
        const structuredLocator = parseStructuredLocatorPayload(item.locator)
        if (structuredLocator) {
          const editorState = deserializeLocatorEditorState(structuredLocator, {
            preferAdvanced: item.type === 'text_content',
          })
          meta.locator_editor_mode = editorState.mode
          meta.locator_structured_form = editorState.simpleForm
          meta.locator_advanced_form = editorState.advancedForm
          meta.locator_structured = JSON.stringify(structuredLocator, null, 2)
        }
      }

      if (this.showTypeField(item.type, 'check_key')) {
        const checkConfig = parseCheckConfig(item.check_key || '')
        meta.check_mode = checkConfig.mode
        meta.check_rule_list = withCheckRuleUid(checkConfig.bool_rules)
        meta.compare_rule = {
          ...createCompareRule(),
          ...(checkConfig.compare_rule || {}),
        }
      }

      if (this.showTypeField(item.type, 'locator') && !String(meta.locator_structured || '').trim()) {
        meta.locator_structured = item.type === 'text_content'
          ? this.stringifyStructuredLocatorPayload(this.buildAdvancedLocatorPayloadByForm(meta.locator_advanced_form))
          : this.stringifyStructuredLocatorPayload(this.buildStructuredLocatorPayloadByForm(meta.locator_structured_form))
      }

      return meta
    },
    supportLocatorExpression(type) {
      return this.showTypeField(type, 'locator')
        && type !== 'delete_element'
        && type !== 'login_username_password'
    },
    decodeLocatorList(rawLocator, separator) {
      const list = String(rawLocator || '')
        .split(separator)
        .map(v => v.trim())
        .filter(Boolean)
        .map(v => ({ uid: createLocatorRow().uid, value: v }))
      return list.length > 0 ? list : [createLocatorRow()]
    },
    decodeLocatorExpression(rawLocator) {
      const normalizedLocator = String(rawLocator || '').trim()
      if (!normalizedLocator) {
        return {
          locator_joiner: 'single',
          locator_raw: '',
          locator_list: [createLocatorRow()],
        }
      }

      const hasAnd = normalizedLocator.includes('&&')
      const hasOr = normalizedLocator.includes('||')
      if (hasAnd && hasOr) {
        return {
          locator_joiner: 'raw',
          locator_raw: normalizedLocator,
          locator_list: [createLocatorRow()],
        }
      }

      const separator = hasAnd ? '&&' : (hasOr ? '||' : '')
      const segments = separator ? normalizedLocator.split(separator) : [normalizedLocator]
      const locatorList = segments.map(segment => this.parseLocatorSegment(segment)).filter(Boolean)

      return {
        locator_joiner: hasAnd ? 'and' : (hasOr ? 'or' : 'single'),
        locator_raw: normalizedLocator,
        locator_list: locatorList.length > 0 ? locatorList : [createLocatorRow()],
      }
    },
    parseLocatorSegment(segment) {
      const normalizedSegment = String(segment || '').trim()
      if (!normalizedSegment) return null

      const partList = normalizedSegment.split('|').map(item => item.trim()).filter(Boolean)
      let locatorValue = partList[0] || ''
      const existMode = locatorValue.startsWith('!') ? 'not_exist' : 'exist'
      if (existMode === 'not_exist') {
        locatorValue = locatorValue.slice(1)
      }

      return {
        uid: createLocatorRow().uid,
        value: locatorValue,
        exist_mode: existMode,
        match_mode: partList.includes('first') ? 'first' : 'all',
      }
    },
    shouldShowStructuredLocatorTargetText(kind) {
      return kind === 'button_text'
    },
    getStructuredLocatorPrimaryPlaceholder(kind) {
      if (kind === 'button_text') return '例如 提交、登录、确定'
      if (kind === 'text') return '例如 登录成功、请选择门店'
      if (kind === 'label') return '例如 用户名、手机号'
      if (kind === 'placeholder') return '例如 请输入用户名'
      if (kind === 'alt_text') return '例如 商品图片、头像'
      if (kind === 'title') return '例如 复制、刷新'
      if (kind === 'test_id') return '例如 submit-btn'
      return '例如 .dialog .confirm-btn 或 //button[text()="提交"]'
    },
    getStructuredLocatorScenarioHint(kind) {
      if (kind === 'button_text') return '适合页面上有明确按钮文案的场景，最适合普通用户配置。'
      if (kind === 'text') return '适合根据页面上看得到的文字定位，例如提示语、标题、状态文案。'
      if (kind === 'label') return '适合输入框左侧或上方有固定标签文字的场景。'
      if (kind === 'placeholder') return '适合输入框内部有占位提示语的场景。'
      if (kind === 'alt_text') return '适合图片、图标有替代说明文字的场景。'
      if (kind === 'title') return '适合鼠标悬停提示或 title 属性稳定的元素。'
      if (kind === 'test_id') return '适合研发已经在页面里埋了稳定 test_id 的场景。'
      return '适合复杂页面或无法通过文字稳定定位的场景，推荐由研发或熟悉选择器的人填写。'
    },
    deserializeStructuredLocatorForm(locatorValue) {
      const form = createStructuredLocatorForm()
      const payload = locatorValue && typeof locatorValue === 'object' ? locatorValue : {}
      const spec = payload.spec && typeof payload.spec === 'object' ? payload.spec : {}
      const options = spec.options && typeof spec.options === 'object' ? spec.options : {}
      const pick = spec.pick && typeof spec.pick === 'object' ? spec.pick : {}

      form.method = spec.method || 'locator'
      form.value = spec.value || ''
      form.target_text = options.name || ''
      form.exact = Boolean(options.exact)
      form.negate = Boolean(spec.negate)
      form.timeout_mills = Number(spec.timeout_mills ?? 3000)

      if (spec.method === 'role' && spec.value === 'button') {
        form.kind = 'button_text'
      } else if (spec.method === 'text') {
        form.kind = 'text'
      } else if (spec.method === 'label') {
        form.kind = 'label'
      } else if (spec.method === 'placeholder') {
        form.kind = 'placeholder'
      } else if (spec.method === 'alt_text') {
        form.kind = 'alt_text'
      } else if (spec.method === 'title') {
        form.kind = 'title'
      } else if (spec.method === 'test_id') {
        form.kind = 'test_id'
      } else {
        form.kind = 'css'
      }

      if (pick.first) {
        form.pick_mode = 'first'
      } else if (pick.last) {
        form.pick_mode = 'last'
      } else if (Number.isInteger(Number(pick.nth))) {
        form.pick_mode = 'nth'
        form.nth = Number(pick.nth)
      }

      return form
    },
    buildStructuredLocatorPayload() {
      return this.buildStructuredLocatorPayloadByForm(this.formMeta.locator_structured_form)
    },
    buildAdvancedLocatorPayload() {
      return this.buildAdvancedLocatorPayloadByForm(this.formMeta.locator_advanced_form)
    },
    buildStructuredLocatorPayloadByForm(formValue) {
      return buildSimpleLocatorPayload(formValue || createStructuredLocatorForm())
    },
    buildAdvancedLocatorPayloadByForm(formValue) {
      return buildAdvancedLocatorPayload(formValue || createAdvancedStructuredLocatorForm())
    },
    deserializeAdvancedLocatorForm(locatorValue) {
      return deserializeLocatorEditorState(locatorValue, { preferAdvanced: true }).advancedForm
    },
    stringifyStructuredLocatorPayload(payload) {
      return stringifyLocatorPayload(payload)
    },
    serializeLocatorExpression() {
      return this.stringifyStructuredLocatorPayload(
        this.useAdvancedLocatorEditor ? this.buildAdvancedLocatorPayload() : this.buildStructuredLocatorPayload()
      )
    },
    parseStructuredLocatorValue() {
      return this.useAdvancedLocatorEditor ? this.buildAdvancedLocatorPayload() : this.buildStructuredLocatorPayload()
    },
    serializeItem() {
      // serializeItem 负责把前端表单重新编码成后端需要的流程项结构。
      // serializeItem serializes editable form state back into backend process payloads.
      const item = {
        ...this.localItem,
        next_ids: '',
      }
      const checkKeyExpression = serializeCheckConfig({
        mode: this.formMeta.check_mode,
        bool_rules: this.formMeta.check_rule_list,
        compare_rule: this.formMeta.compare_rule,
      })

      if (item.type === 'bool_result') {
        item.locator = this.formMeta.bool_result_rules.length > 0
          ? JSON.stringify(buildLocatorConfigByType('bool_result', this.formMeta))
          : ''
        item.value = ''
        item.out_key = this.formMeta.out_key
        item.check_key = checkKeyExpression
      } else if (item.type === 'text_content') {
        item.locator = JSON.stringify(buildLocatorConfigByType('text_content', this.formMeta))
        item.value = ''
        item.out_key = this.formMeta.out_key
        item.check_key = checkKeyExpression
      } else if (item.type === 'wait_url') {
        item.locator = ''
        item.value = serializeWaitUrlValue(this.formMeta)
        item.out_key = ''
        item.check_key = checkKeyExpression
      } else if (item.type === 'redirect_uri') {
        item.locator = ''
        item.value = serializeRedirectUriValue(this.formMeta)
        item.out_key = ''
        item.check_key = checkKeyExpression
      } else if (item.type === 'no_exist_wait') {
        item.locator = this.serializeLocatorExpression()
        item.value = `${Number(this.formMeta.wait_second || 10)}|${Number(this.formMeta.wait_count || 3)}`
        item.out_key = this.formMeta.out_key
        item.check_key = checkKeyExpression
      } else if (item.type === 'login_username_password') {
        const userLocator = this.formMeta.locator_list[0] ? this.formMeta.locator_list[0].value : ''
        item.locator = [userLocator, this.formMeta.secondary_locator, this.formMeta.tertiary_locator].filter(Boolean).join('||')
        item.value = ''
        item.out_key = ''
        item.check_key = ''
      } else if (item.type === 'delete_element') {
        item.locator = this.formMeta.locator_list.map(v => v.value.trim()).filter(Boolean).join('|')
        item.value = this.formMeta.delete_mode
        item.out_key = ''
        item.check_key = ''
      } else if (item.type === 'click' || item.type === 'input') {
        item.locator = JSON.stringify(buildLocatorConfigByType(item.type, this.formMeta))
        item.value = this.formMeta.value
        item.out_key = this.formMeta.out_key
        item.check_key = checkKeyExpression
      } else {
        if (this.showTypeField(item.type, 'locator')) {
          item.locator = this.supportLocatorExpression(item.type)
            ? this.serializeLocatorExpression()
            : this.formMeta.locator_list.map(v => v.value.trim()).filter(Boolean).join('||')
        } else {
          item.locator = ''
        }
        item.value = this.formMeta.value
        item.out_key = this.formMeta.out_key
        item.check_key = checkKeyExpression
      }

      return item
    },
    emitChange() {
      if (this.syncingFromParent) return
      const serializedItem = this.serializeItem()
      if (Object.keys(this.fieldErrors).length > 0) {
        this.runValidation()
      }
      const nextSignature = this.createSignature(serializedItem)
      if (nextSignature === this.lastSerializedSignature) {
        return
      }
      this.lastSerializedSignature = nextSignature
      this.$emit('update:modelValue', serializedItem)
    },
    showTypeField(type, fieldName) {
      return (PROCESS_TYPE_FIELDS[type] || []).includes(fieldName)
    },
    showField(fieldName) {
      return this.currentFields.includes(fieldName)
    },
    fieldGuide(fieldName) {
      if (fieldName === 'locator') {
        return ''
      }
      if (fieldName === 'check_key') {
        if (this.formMeta.check_mode === 'compare') {
          return '内容比较会生成类似 {login_user}!={user_name}、{login_user}!={password} 或 {sign_in_btn}==Sign in 的条件。左侧选前面节点的输出，比较类型选“等于/不等于”，右侧既可以直接选择注入值，也可以输入固定字符串。'
        }
        if (this.formMeta.check_mode === 'bool') {
          return '结果判断会生成类似 {need_login}、{need_login}&&{qrcode_dialog} 或 {need_login}&&{need_change_to_password} 的条件。后端这里只支持多个条件用 && 同时成立，不支持或条件；单独一个条件就表示该输出为 true 时执行。'
        }
        return '可以不限制执行，也可以按前面节点的结果判断，或按文本内容比较判断。'
      }
      if (fieldName === 'bool_result_rules') {
        return '每条规则都使用高级结构化定位。命中该规则时，按右侧 true / false 返回结果。'
      }
      return PROCESS_ITEM_FIELD_GUIDES[fieldName] || ''
    },
    fieldError(fieldName) {
      return this.fieldErrors[fieldName] || ''
    },
    runValidation() {
      const validationResult = validateProcessItemForm({
        item: this.serializeItem(),
        formMeta: this.formMeta,
      })
      this.fieldErrors = validationResult.fieldErrors || {}
      return validationResult
    },
    validateForSave() {
      return this.runValidation().valid
    },
    fieldLabel(fieldName) {
      const labels = {
        locator: '主元素定位',
        secondary_locator: '密码框定位',
        tertiary_locator: '提交按钮定位',
        value: '值',
        out_key: '输出键',
        check_key: '是否执行判断',
        wait_second: '等待秒数',
        wait_count: '轮询次数',
        response_url: '等待地址',
        delete_mode: '删除类型',
        register_response_urls: '跳转后等待地址',
      }
      if (this.localItem.type === 'login_username_password' && fieldName === 'locator') {
        return '用户名框定位'
      }
      return labels[fieldName] || fieldName
    },
    fieldPlaceholder(fieldName) {
      if (fieldName === 'locator') return '例如 .username 或 .login-btn'
      if (fieldName === 'secondary_locator') return '例如 #password'
      if (fieldName === 'tertiary_locator') return '例如 .submit-btn'
      if (this.localItem.type === 'input' && fieldName === 'value') return '支持 {user_name} / {password} / {rand}'
      if (fieldName === 'response_url') return '例如 {scheme}://{domain}/api/login'
      return ''
    },
    textareaRows(fieldName) {
      return fieldName === 'value' ? 2 : 1
    },
    addRegisterResponseUrl() {
      this.formMeta.register_response_urls.push(createRegisterUrl())
    },
    removeRegisterResponseUrl(index) {
      this.formMeta.register_response_urls.splice(index, 1)
    },
    addLocatorRow() {
      this.formMeta.locator_list.push(createLocatorRow())
    },
    removeLocatorRow(index) {
      this.formMeta.locator_list.splice(index, 1)
    },
    addBoolResultRule() {
      this.formMeta.bool_result_rules.push(createBoolResultRule())
    },
    removeBoolResultRule(index) {
      this.formMeta.bool_result_rules.splice(index, 1)
    },
    addCheckRule() {
      this.formMeta.check_rule_list.push(createCheckRule())
    },
    removeCheckRule(index) {
      this.formMeta.check_rule_list.splice(index, 1)
    },
    // applyCompareRightQuickPick 用于把注入变量快捷写入比较右值输入框。
    // applyCompareRightQuickPick applies injected-variable shortcuts into the compare right input.
    applyCompareRightQuickPick(value) {
      this.formMeta.compare_rule.right = String(value || '')
    },
    describeBaseLocator(baseLocator) {
      const payload = baseLocator && baseLocator.locator_editor_mode === 'advanced'
        ? this.buildAdvancedLocatorPayloadByForm(baseLocator.locator_advanced_form)
        : this.buildStructuredLocatorPayloadByForm(baseLocator && baseLocator.locator_structured_form)
      return formatStructuredLocator(payload)
    },
    openBaseLocatorDialog(targetType, index, title) {
      let source = createBaseLocatorMeta()
      if (targetType === 'bool_result') {
        source = this.formMeta.bool_result_rules[index] && this.formMeta.bool_result_rules[index].base_locator
          ? this.formMeta.bool_result_rules[index].base_locator
          : createBaseLocatorMeta()
      } else if (targetType === 'text_content') {
        source = this.formMeta.text_content_locators[index] && this.formMeta.text_content_locators[index].base_locator
          ? this.formMeta.text_content_locators[index].base_locator
          : createBaseLocatorMeta()
      } else if (targetType === 'action') {
        source = this.formMeta.action_locators[index] && this.formMeta.action_locators[index].base_locator
          ? this.formMeta.action_locators[index].base_locator
          : createBaseLocatorMeta()
      }
      this.baseLocatorDialog = {
        visible: true,
        targetType,
        targetIndex: index,
        title,
        draft: JSON.parse(JSON.stringify(source || createBaseLocatorMeta())),
      }
    },
    saveBaseLocatorDialog() {
      const draft = JSON.parse(JSON.stringify(this.baseLocatorDialog.draft || createBaseLocatorMeta()))
      if (this.baseLocatorDialog.targetType === 'bool_result' && this.formMeta.bool_result_rules[this.baseLocatorDialog.targetIndex]) {
        this.formMeta.bool_result_rules[this.baseLocatorDialog.targetIndex].base_locator = draft
      } else if (this.baseLocatorDialog.targetType === 'text_content' && this.formMeta.text_content_locators[this.baseLocatorDialog.targetIndex]) {
        this.formMeta.text_content_locators[this.baseLocatorDialog.targetIndex].base_locator = draft
      } else if (this.baseLocatorDialog.targetType === 'action' && this.formMeta.action_locators[this.baseLocatorDialog.targetIndex]) {
        this.formMeta.action_locators[this.baseLocatorDialog.targetIndex].base_locator = draft
      }
      this.baseLocatorDialog.visible = false
    },
    addActionLocator() {
      this.formMeta.action_locators.push(createActionLocator())
    },
    addTextContentLocator() {
      this.formMeta.text_content_locators.push(createTextContentRule())
    },
    removeTextContentLocator(index) {
      this.formMeta.text_content_locators.splice(index, 1)
    },
    removeActionLocator(index) {
      this.formMeta.action_locators.splice(index, 1)
    },
  },
}
</script>

<style scoped>
.list-editor {
  width: 100%;
}

.locator-expression-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}

.locator-expression-toolbar__select {
  width: 220px;
}

.locator-expression-toolbar__tip,
.field-guide {
  color: #6b7b68;
  font-size: 12px;
  line-height: 1.5;
}

.locator-config-panel {
  display: grid;
  gap: 12px;
}

.locator-config-panel__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  background: #f7faf4;
  border: 1px solid #dfe8d7;
  border-radius: 10px;
}

.locator-config-panel__title {
  color: #456238;
  font-size: 13px;
  font-weight: 600;
}

.locator-config-panel__select {
  width: 280px;
}

.base-locator-card {
  padding: 14px;
  background: #fcfdfb;
  border: 1px solid #d8e3d0;
  border-radius: 12px;
}

.base-locator-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.base-locator-card__title {
  color: #355128;
  font-size: 13px;
  font-weight: 600;
}

.base-locator-card__summary {
  margin: 10px 0;
  color: #6b7b68;
  font-size: 12px;
  line-height: 1.6;
  word-break: break-word;
}

.base-locator-card__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.base-locator-card__action {
  display: flex;
  align-items: center;
  gap: 8px;
}

.base-locator-card__action-label {
  color: #4b6540;
  font-size: 12px;
}

.base-locator-card__action-select {
  width: 180px;
}

.field-guide {
  margin-top: 8px;
}

.locator-purpose-card {
  margin-bottom: 10px;
  padding: 10px 12px;
  background: #f6f8f4;
  border: 1px solid #dfe8d7;
  border-radius: 8px;
}

.locator-purpose-card__title {
  color: #48653a;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 4px;
}

.locator-purpose-card__text {
  color: #6b7b68;
  font-size: 12px;
  line-height: 1.6;
}

.locator-behavior-summary {
  margin-top: 10px;
  padding: 10px 12px;
  background: #fffdf5;
  border: 1px solid #efe4b0;
  border-radius: 8px;
}

.locator-behavior-summary__title {
  color: #8a6b13;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 4px;
}

.locator-behavior-summary__text {
  color: #7b714d;
  font-size: 12px;
  line-height: 1.6;
}

.structured-locator-editor {
  margin-top: 4px;
}

.structured-locator-card {
  padding: 16px;
  background: linear-gradient(180deg, #fcfdfb 0%, #f6f8f4 100%);
  border: 1px solid #d8e3d0;
  border-radius: 14px;
  box-shadow: 0 8px 18px rgba(89, 113, 72, 0.08);
}

.structured-locator-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.structured-locator-card__title {
  color: #355128;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.4;
}

.structured-locator-card__desc {
  margin-top: 4px;
  color: #6a7d60;
  font-size: 12px;
  line-height: 1.6;
}

.structured-locator-card__badge {
  flex: 0 0 auto;
  padding: 4px 10px;
  color: #48653a;
  font-size: 12px;
  font-weight: 600;
  background: #e7f1df;
  border-radius: 999px;
}

.structured-locator-section {
  padding: 14px;
  margin-bottom: 12px;
  background: rgba(255, 255, 255, 0.88);
  border: 1px solid #e4ebde;
  border-radius: 12px;
}

.structured-locator-section:last-child {
  margin-bottom: 0;
}

.structured-locator-section__title {
  margin-bottom: 10px;
  color: #456238;
  font-size: 13px;
  font-weight: 600;
}

.structured-locator-grid {
  display: grid;
  grid-template-columns: minmax(240px, 320px) minmax(0, 1fr);
  gap: 10px;
  align-items: center;
}

.structured-locator-grid + .structured-locator-grid {
  margin-top: 10px;
}

.structured-locator-grid--secondary {
  grid-template-columns: minmax(0, 1fr);
}

.structured-locator-grid--stacked {
  grid-template-columns: minmax(0, 1fr);
}

.structured-locator-filter-pair {
  display: grid;
  grid-template-columns: minmax(220px, 280px) minmax(0, 1fr);
  gap: 10px;
  align-items: center;
}

.structured-locator-tip {
  margin-top: 10px;
  padding: 10px 12px;
  color: #607355;
  font-size: 12px;
  line-height: 1.6;
  background: #f7faf4;
  border-radius: 10px;
}

.structured-locator-switches {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.structured-locator-switch-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 14px;
  background: #f9fbf7;
  border: 1px solid #e3eadc;
  border-radius: 10px;
}

.structured-locator-switch-card__label {
  color: #4e6442;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.5;
}

.structured-locator-switch-card__control {
  display: flex;
  align-items: center;
  gap: 8px;
}

.structured-locator-switch-card__state {
  color: #6c7c63;
  font-size: 12px;
  white-space: nowrap;
}

.structured-locator-inline-field {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  padding: 12px 14px;
  background: #f9fbf7;
  border: 1px solid #e3eadc;
  border-radius: 10px;
}

.structured-locator-inline-field--compact {
  min-height: 46px;
}

.structured-locator-inline-field__label {
  min-width: 0;
}

.structured-locator-inline-field__title {
  color: #4e6442;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.5;
}

.structured-locator-inline-field__desc {
  margin-top: 2px;
  color: #7a8970;
  font-size: 12px;
  line-height: 1.5;
}

.structured-locator-inline-field__control {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 0 0 180px;
}

.structured-locator-inline-field__unit {
  color: #69805d;
  font-size: 12px;
  font-weight: 600;
}

.structured-locator-preview {
  margin-top: 4px;
}

.locator-expression-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 130px 130px 60px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.locator-expression-row__mode {
  width: 100%;
}

.list-editor__row,
.response-url-row {
  display: grid;
  grid-template-columns: 1fr 120px 60px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.list-editor__row {
  grid-template-columns: 1fr 60px;
}

.bool-result-row {
  display: grid;
  grid-template-columns: 1fr 140px 60px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.bool-result-rule-card {
  padding: 12px;
  margin-bottom: 12px;
  background: #fbfcfa;
  border: 1px solid #e2eadc;
  border-radius: 12px;
}

.bool-result-rule-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.bool-result-rule-card__title {
  color: #48653a;
  font-size: 13px;
  font-weight: 700;
}

.bool-result-rule-card__actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.bool-result-rule-card__mode {
  width: 140px;
}

.bool-result-rule-card__result {
  width: 160px;
}

.structured-locator-card--nested {
  padding: 12px;
  background: linear-gradient(180deg, #ffffff 0%, #f7faf5 100%);
  box-shadow: none;
}

.check-rule-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 140px 60px;
  gap: 10px;
  align-items: center;
  margin-bottom: 10px;
}

.check-rule-row__mode {
  width: 100%;
}

.check-mode-select {
  width: 220px;
  margin-bottom: 10px;
}

.compare-rule-row {
  display: grid;
  grid-template-columns: minmax(260px, 1fr) 140px minmax(320px, 1fr);
  gap: 10px;
  align-items: start;
  margin-bottom: 10px;
}

.compare-rule-row__operator {
  width: 100%;
}

.compare-rule-right {
  min-width: 0;
}

.compare-rule-right__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}

.plain-number-input {
  width: 100%;
}
</style>
