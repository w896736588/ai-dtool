package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_curl"
	"github.com/spf13/cast"
)

const (
	// CodeTypeCurlChrome 中文：Chrome 风格 curl。 English: Chrome-style curl snippet.
	CodeTypeCurlChrome = `curl bash(chrome)`
	// CodeTypeCurlApifox 中文：Apifox 风格 shell curl。 English: Apifox-style shell curl snippet.
	CodeTypeCurlApifox = `curl shell(apifox)`
	// CodeTypeFetch 中文：JavaScript fetch 示例。 English: JavaScript fetch snippet.
	CodeTypeFetch = `JavaScript fetch`
	// CodeTypeAxios 中文：JavaScript axios 示例。 English: JavaScript axios snippet.
	CodeTypeAxios = `JavaScript axios`
	// CodeTypePython 中文：Python requests 示例。 English: Python requests snippet.
	CodeTypePython = `Python requests`
	// CodeTypePHP 中文：PHP cURL 示例。 English: PHP cURL snippet.
	CodeTypePHP = `PHP cURL`
	// CodeTypeGolang 中文：Golang net/http 示例。 English: Golang net/http snippet.
	CodeTypeGolang = `Golang net/http`
	// CodeTypePostman 中文：Postman Collection 导入片段。 English: Postman collection snippet.
	CodeTypePostman = `Postman collection`
)

// SupportedCodeTypes 中文：代码 tab 支持的常见代码类型列表。 English: Supported snippet types for the code tab.
var SupportedCodeTypes = []string{
	CodeTypeCurlChrome,
	CodeTypeCurlApifox,
	CodeTypeFetch,
	CodeTypeAxios,
	CodeTypePython,
	CodeTypePHP,
	CodeTypeGolang,
	CodeTypePostman,
}

type codePair struct {
	Key   string
	Value string
}

// GenerateCode 中文：按指定类型生成代码片段。 English: Generate a code snippet for the requested type.
func (h *Api) GenerateCode(codeType string) string {
	h.ReplaceEnv()
	switch codeType {
	case CodeTypeCurlChrome:
		return h.buildCurlSnippet(false)
	case CodeTypeCurlApifox:
		return h.buildCurlSnippet(true)
	case CodeTypeFetch:
		return h.buildFetchSnippet()
	case CodeTypeAxios:
		return h.buildAxiosSnippet()
	case CodeTypePython:
		return h.buildPythonRequestsSnippet()
	case CodeTypePHP:
		return h.buildPHPCurlSnippet()
	case CodeTypeGolang:
		return h.buildGolangSnippet()
	case CodeTypePostman:
		return h.buildPostmanCollectionSnippet()
	default:
		return h.buildCurlSnippet(false)
	}
}

// ToChromeCurlBash 中文：兼容旧调用的 Chrome curl 生成入口。 English: Backward-compatible Chrome curl entrypoint.
func (h *Api) ToChromeCurlBash() string {
	return h.GenerateCode(CodeTypeCurlChrome)
}

// requestBodyText 中文：获取当前请求体文本。 English: Resolve the raw request payload text for the current request.
func (h *Api) requestBodyText() string {
	if h.CurlStruct.ContentType == define.ContentTypeJson {
		return h.CurlStruct.BodyJson
	}
	if h.CurlStruct.ContentType == define.ContentTypeText || h.CurlStruct.ContentType == define.ContentTypeRaw {
		return h.CurlStruct.BodyRaw
	}
	return ``
}

// sortedHeaders 中文：按 key 排序请求头，保证输出稳定。 English: Sort headers by key for stable snippet output.
func (h *Api) sortedHeaders() []codePair {
	keys := make([]string, 0, len(h.CurlStruct.Headers))
	for key := range h.CurlStruct.Headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	headers := make([]codePair, 0, len(keys))
	for _, key := range keys {
		headers = append(headers, codePair{Key: key, Value: h.CurlStruct.Headers[key]})
	}
	return headers
}

// sortedBodyForm 中文：按字段稳定排序表单项。 English: Sort form fields to keep snippet output deterministic.
func (h *Api) sortedBodyForm() []p_curl.KeyValue {
	bodyForm := make([]p_curl.KeyValue, len(h.CurlStruct.BodyForm))
	copy(bodyForm, h.CurlStruct.BodyForm)
	sort.SliceStable(bodyForm, func(i, j int) bool {
		if bodyForm[i].Field == bodyForm[j].Field {
			return bodyForm[i].Value < bodyForm[j].Value
		}
		return bodyForm[i].Field < bodyForm[j].Field
	})
	return bodyForm
}

// buildCurlSnippet 中文：生成 curl 代码，兼容 Chrome/Apifox 两种风格。 English: Build curl code for Chrome/Apifox styles.
func (h *Api) buildCurlSnippet(withLocation bool) string {
	lines := make([]string, 0)
	method := strings.ToUpper(h.CurlStruct.Method)
	if method == `` {
		method = http.MethodGet
	}
	firstLine := fmt.Sprintf("curl --request %s '%s' \\", method, escapeSingleQuote(h.CurlStruct.Url))
	if withLocation {
		firstLine = fmt.Sprintf("curl --location --request %s '%s' \\", method, escapeSingleQuote(h.CurlStruct.Url))
	}
	lines = append(lines, firstLine)
	for _, header := range h.sortedHeaders() {
		lines = append(lines, fmt.Sprintf("  --header '%s: %s' \\", escapeSingleQuote(header.Key), escapeSingleQuote(header.Value)))
	}

	switch h.CurlStruct.ContentType {
	case define.ContentTypeForm:
		for _, item := range h.sortedBodyForm() {
			lines = append(lines, fmt.Sprintf("  --data-urlencode '%s=%s' \\", escapeSingleQuote(item.Field), escapeSingleQuote(cast.ToString(normalizeBodyValue(item)))))
		}
	case define.ContentTypeMultiForm:
		for _, item := range h.sortedBodyForm() {
			if item.Type == p_curl.FieldTypeFile || item.Type == `file` {
				lines = append(lines, fmt.Sprintf("  --form '%s=@%s' \\", escapeSingleQuote(item.Field), escapeSingleQuote(item.Value)))
				continue
			}
			lines = append(lines, fmt.Sprintf("  --form '%s=%s' \\", escapeSingleQuote(item.Field), escapeSingleQuote(cast.ToString(normalizeBodyValue(item)))))
		}
	case define.ContentTypeJson, define.ContentTypeText, define.ContentTypeRaw:
		bodyText := h.requestBodyText()
		if strings.TrimSpace(bodyText) != `` {
			lines = append(lines, fmt.Sprintf("  --data-raw '%s' \\", escapeSingleQuote(bodyText)))
		}
	}
	if len(lines) == 0 {
		return ``
	}
	lines[len(lines)-1] = strings.TrimSuffix(lines[len(lines)-1], " \\")
	return strings.Join(lines, "\n")
}

// buildFetchSnippet 中文：生成 JavaScript fetch 示例。 English: Build a JavaScript fetch example.
func (h *Api) buildFetchSnippet() string {
	lines := []string{
		fmt.Sprintf("const url = '%s';", escapeSingleQuote(h.CurlStruct.Url)),
	}
	headers := h.sortedHeaders()
	if len(headers) > 0 {
		lines = append(lines, "const headers = {")
		for _, header := range headers {
			lines = append(lines, fmt.Sprintf("  '%s': '%s',", escapeSingleQuote(header.Key), escapeSingleQuote(header.Value)))
		}
		lines = append(lines, "};", "")
	}
	bodyLines := h.buildJSBodyLines()
	lines = append(lines, bodyLines...)
	lines = append(lines, "fetch(url, {")
	lines = append(lines, fmt.Sprintf("  method: '%s',", strings.ToUpper(h.CurlStruct.Method)))
	if len(headers) > 0 {
		lines = append(lines, "  headers,")
	}
	if bodyRef := h.jsBodyRef(); bodyRef != `` {
		lines = append(lines, fmt.Sprintf("  body: %s,", bodyRef))
	}
	lines = append(lines, "})")
	lines = append(lines, "  .then((response) => response.text())")
	lines = append(lines, "  .then((result) => console.log(result))")
	lines = append(lines, "  .catch((error) => console.error(error));")
	return strings.Join(lines, "\n")
}

// buildAxiosSnippet 中文：生成 JavaScript axios 示例。 English: Build a JavaScript axios example.
func (h *Api) buildAxiosSnippet() string {
	lines := []string{
		"import axios from 'axios';",
		"",
	}
	bodyLines := h.buildJSBodyLines()
	lines = append(lines, bodyLines...)
	lines = append(lines, "const config = {")
	lines = append(lines, fmt.Sprintf("  method: '%s',", strings.ToLower(h.CurlStruct.Method)))
	lines = append(lines, fmt.Sprintf("  url: '%s',", escapeSingleQuote(h.CurlStruct.Url)))
	headers := h.sortedHeaders()
	if len(headers) > 0 {
		lines = append(lines, "  headers: {")
		for _, header := range headers {
			lines = append(lines, fmt.Sprintf("    '%s': '%s',", escapeSingleQuote(header.Key), escapeSingleQuote(header.Value)))
		}
		lines = append(lines, "  },")
	}
	if bodyRef := h.jsBodyRef(); bodyRef != `` {
		lines = append(lines, fmt.Sprintf("  data: %s,", bodyRef))
	}
	lines = append(lines, "};", "", "axios.request(config)")
	lines = append(lines, "  .then((response) => {")
	lines = append(lines, "    console.log(response.data);")
	lines = append(lines, "  })")
	lines = append(lines, "  .catch((error) => {")
	lines = append(lines, "    console.error(error);")
	lines = append(lines, "  });")
	return strings.Join(lines, "\n")
}

// buildPythonRequestsSnippet 中文：生成 Python requests 示例。 English: Build a Python requests example.
func (h *Api) buildPythonRequestsSnippet() string {
	imports := []string{"import requests"}
	if h.CurlStruct.ContentType == define.ContentTypeJson && strings.TrimSpace(h.requestBodyText()) != `` {
		imports = append(imports, "import json")
	}
	lines := append(imports, "", fmt.Sprintf("url = %q", h.CurlStruct.Url))
	headers := h.sortedHeaders()
	if len(headers) > 0 {
		lines = append(lines, "headers = {")
		for _, header := range headers {
			lines = append(lines, fmt.Sprintf("    %q: %q,", header.Key, header.Value))
		}
		lines = append(lines, "}", "")
	}
	bodyLines, requestArg := h.buildPythonBodyLines()
	lines = append(lines, bodyLines...)
	requestLine := fmt.Sprintf("response = requests.%s(url", strings.ToLower(h.CurlStruct.Method))
	if len(headers) > 0 {
		requestLine += ", headers=headers"
	}
	if requestArg != `` {
		requestLine += ", " + requestArg
	}
	requestLine += ")"
	lines = append(lines, requestLine, "print(response.text)")
	return strings.Join(lines, "\n")
}

// buildPHPCurlSnippet 中文：生成 PHP cURL 示例。 English: Build a PHP cURL example.
func (h *Api) buildPHPCurlSnippet() string {
	lines := []string{
		"<?php",
		"",
		"$curl = curl_init();",
		"",
		"curl_setopt_array($curl, [",
		fmt.Sprintf("    CURLOPT_URL => '%s',", escapeSingleQuote(h.CurlStruct.Url)),
		"    CURLOPT_RETURNTRANSFER => true,",
		fmt.Sprintf("    CURLOPT_CUSTOMREQUEST => '%s',", strings.ToUpper(h.CurlStruct.Method)),
	}
	if bodyLine := h.buildPHPBodyOption(); bodyLine != `` {
		lines = append(lines, bodyLine)
	}
	headers := h.sortedHeaders()
	if len(headers) > 0 {
		lines = append(lines, "    CURLOPT_HTTPHEADER => [")
		for _, header := range headers {
			lines = append(lines, fmt.Sprintf("        '%s: %s',", escapeSingleQuote(header.Key), escapeSingleQuote(header.Value)))
		}
		lines = append(lines, "    ],")
	}
	lines = append(lines, "]);", "", "$response = curl_exec($curl);", "$error = curl_error($curl);", "", "curl_close($curl);", "", "if ($error) {", "    echo $error;", "} else {", "    echo $response;", "}")
	return strings.Join(lines, "\n")
}

// buildGolangSnippet 中文：生成 Golang net/http 示例。 English: Build a Golang net/http example.
func (h *Api) buildGolangSnippet() string {
	lines := []string{
		"package main",
		"",
		"import (",
		`	"bytes"`,
		`	"fmt"`,
		`	"io"`,
		`	"net/http"`,
		")",
		"",
		"func main() {",
	}
	if bodyLine := h.buildGoBodyDeclaration(); bodyLine != `` {
		lines = append(lines, bodyLine, "")
	}
	bodyVar := h.goBodyVar()
	if bodyVar == `` {
		bodyVar = "nil"
	}
	lines = append(lines, fmt.Sprintf(`	req, err := http.NewRequest("%s", "%s", %s)`, strings.ToUpper(h.CurlStruct.Method), h.CurlStruct.Url, bodyVar))
	lines = append(lines, "	if err != nil {", "		panic(err)", "	}")
	for _, header := range h.sortedHeaders() {
		lines = append(lines, fmt.Sprintf(`	req.Header.Set("%s", "%s")`, header.Key, header.Value))
	}
	lines = append(lines, "", "	client := &http.Client{}", "	resp, err := client.Do(req)", "	if err != nil {", "		panic(err)", "	}", "	defer resp.Body.Close()", "", "	body, err := io.ReadAll(resp.Body)", "	if err != nil {", "		panic(err)", "	}", "", "	fmt.Println(string(body))", "}")
	return strings.Join(lines, "\n")
}

// buildPostmanCollectionSnippet 中文：生成 Postman Collection v2.1 的单接口片段。 English: Build a single-request Postman Collection v2.1 snippet.
func (h *Api) buildPostmanCollectionSnippet() string {
	bodyMode := `raw`
	bodyValue := h.requestBodyText()
	postmanBody := map[string]any{}
	switch h.CurlStruct.ContentType {
	case define.ContentTypeForm:
		bodyMode = "urlencoded"
		items := make([]map[string]string, 0, len(h.CurlStruct.BodyForm))
		for _, item := range h.sortedBodyForm() {
			items = append(items, map[string]string{
				"key":   item.Field,
				"value": cast.ToString(normalizeBodyValue(item)),
				"type":  "text",
			})
		}
		postmanBody["mode"] = bodyMode
		postmanBody["urlencoded"] = items
	case define.ContentTypeMultiForm:
		bodyMode = "formdata"
		items := make([]map[string]string, 0, len(h.CurlStruct.BodyForm))
		for _, item := range h.sortedBodyForm() {
			itemType := "text"
			if item.Type == p_curl.FieldTypeFile || item.Type == `file` {
				itemType = "file"
			}
			items = append(items, map[string]string{
				"key":   item.Field,
				"value": cast.ToString(normalizeBodyValue(item)),
				"type":  itemType,
			})
		}
		postmanBody["mode"] = bodyMode
		postmanBody["formdata"] = items
	default:
		postmanBody["mode"] = bodyMode
		postmanBody["raw"] = bodyValue
	}

	headers := make([]map[string]string, 0)
	for _, header := range h.sortedHeaders() {
		headers = append(headers, map[string]string{
			"key":   header.Key,
			"value": header.Value,
			"type":  "text",
		})
	}

	postmanData := map[string]any{
		"info": map[string]any{
			"name":   "Generated Request",
			"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		"item": []map[string]any{
			{
				"name": "Generated Request",
				"request": map[string]any{
					"method": strings.ToUpper(h.CurlStruct.Method),
					"header": headers,
					"body":   postmanBody,
					"url":    h.CurlStruct.Url,
				},
			},
		},
	}
	jsonBytes, err := json.MarshalIndent(postmanData, "", "  ")
	if err != nil {
		return ``
	}
	return string(jsonBytes)
}

// buildJSBodyLines 中文：为 fetch/axios 生成请求体变量。 English: Build request body variables for fetch/axios snippets.
func (h *Api) buildJSBodyLines() []string {
	switch h.CurlStruct.ContentType {
	case define.ContentTypeForm:
		lines := []string{"const formData = new URLSearchParams();"}
		for _, item := range h.sortedBodyForm() {
			lines = append(lines, fmt.Sprintf("formData.append('%s', '%s');", escapeSingleQuote(item.Field), escapeSingleQuote(cast.ToString(normalizeBodyValue(item)))))
		}
		return append(lines, "")
	case define.ContentTypeMultiForm:
		lines := []string{"const formData = new FormData();"}
		for _, item := range h.sortedBodyForm() {
			if item.Type == p_curl.FieldTypeFile || item.Type == `file` {
				lines = append(lines, "// 中文：文件字段需要替换为真实文件对象。 English: Replace with a real File/Blob object.")
				lines = append(lines, fmt.Sprintf("formData.append('%s', new Blob(), '%s');", escapeSingleQuote(item.Field), escapeSingleQuote(item.Value)))
				continue
			}
			lines = append(lines, fmt.Sprintf("formData.append('%s', '%s');", escapeSingleQuote(item.Field), escapeSingleQuote(cast.ToString(normalizeBodyValue(item)))))
		}
		return append(lines, "")
	case define.ContentTypeJson, define.ContentTypeText, define.ContentTypeRaw:
		bodyText := h.requestBodyText()
		if strings.TrimSpace(bodyText) == `` {
			return nil
		}
		return []string{
			fmt.Sprintf("const body = `%s`;", escapeTemplateLiteral(bodyText)),
			"",
		}
	default:
		return nil
	}
}

// jsBodyRef 中文：返回 fetch/axios 里请求体变量名。 English: Return the request body variable reference for JS snippets.
func (h *Api) jsBodyRef() string {
	switch h.CurlStruct.ContentType {
	case define.ContentTypeForm, define.ContentTypeMultiForm:
		return "formData"
	case define.ContentTypeJson, define.ContentTypeText, define.ContentTypeRaw:
		if strings.TrimSpace(h.requestBodyText()) != `` {
			return "body"
		}
	}
	return ``
}

// buildGoBodyDeclaration 中文：生成 Go 请求体变量声明。 English: Build the Go request body variable declaration.
func (h *Api) buildGoBodyDeclaration() string {
	switch h.CurlStruct.ContentType {
	case define.ContentTypeJson, define.ContentTypeText, define.ContentTypeRaw:
		bodyText := h.requestBodyText()
		if strings.TrimSpace(bodyText) == `` {
			return ``
		}
		return fmt.Sprintf("	bodyReader := bytes.NewBufferString(%q)", bodyText)
	case define.ContentTypeForm:
		values := make([]string, 0, len(h.CurlStruct.BodyForm))
		for _, item := range h.sortedBodyForm() {
			values = append(values, fmt.Sprintf("%s=%s", item.Field, cast.ToString(normalizeBodyValue(item))))
		}
		return fmt.Sprintf("	bodyReader := bytes.NewBufferString(%q)", strings.Join(values, "&"))
	case define.ContentTypeMultiForm:
		lines := []string{
			"	bodyReader := bytes.NewBufferString(\"\")",
			"	// 中文：multipart/form-data 里如果包含文件，需要改成 multipart.Writer 动态组装。 English: Replace with multipart.Writer when real files are needed.",
		}
		return strings.Join(lines, "\n")
	default:
		return ``
	}
}

// goBodyVar 中文：返回 Go 示例里的 body 变量名。 English: Return the Go request body variable name.
func (h *Api) goBodyVar() string {
	switch h.CurlStruct.ContentType {
	case define.ContentTypeJson, define.ContentTypeText, define.ContentTypeRaw, define.ContentTypeForm, define.ContentTypeMultiForm:
		return "bodyReader"
	default:
		return ``
	}
}

// buildPythonBodyLines 中文：生成 Python 请求体定义和调用参数。 English: Build Python body declarations and request argument.
func (h *Api) buildPythonBodyLines() ([]string, string) {
	switch h.CurlStruct.ContentType {
	case define.ContentTypeJson:
		bodyText := strings.TrimSpace(h.requestBodyText())
		if bodyText == `` {
			return nil, ``
		}
		return []string{
			fmt.Sprintf("payload = json.loads(r'''%s''')", escapePythonTripleSingle(bodyText)),
			"",
		}, "json=payload"
	case define.ContentTypeForm:
		lines := []string{"payload = {"}
		for _, item := range h.sortedBodyForm() {
			lines = append(lines, fmt.Sprintf("    %q: %q,", item.Field, cast.ToString(normalizeBodyValue(item))))
		}
		lines = append(lines, "}", "")
		return lines, "data=payload"
	case define.ContentTypeMultiForm:
		lines := []string{"data = {"}
		for _, item := range h.sortedBodyForm() {
			if item.Type == p_curl.FieldTypeFile || item.Type == `file` {
				continue
			}
			lines = append(lines, fmt.Sprintf("    %q: %q,", item.Field, cast.ToString(normalizeBodyValue(item))))
		}
		lines = append(lines, "}")
		fileLines := []string{"files = {"}
		hasFile := false
		for _, item := range h.sortedBodyForm() {
			if item.Type != p_curl.FieldTypeFile && item.Type != `file` {
				continue
			}
			hasFile = true
			fileLines = append(fileLines, fmt.Sprintf("    %q: open(%q, 'rb'),", item.Field, item.Value))
		}
		fileLines = append(fileLines, "}", "")
		if hasFile {
			return append(append(lines, ""), fileLines...), "data=data, files=files"
		}
		return append(lines, ""), "data=data"
	case define.ContentTypeText, define.ContentTypeRaw:
		bodyText := h.requestBodyText()
		if strings.TrimSpace(bodyText) == `` {
			return nil, ``
		}
		return []string{
			fmt.Sprintf("payload = %q", bodyText),
			"",
		}, "data=payload"
	default:
		return nil, ``
	}
}

// buildPHPBodyOption 中文：生成 PHP CURLOPT_POSTFIELDS 片段。 English: Build PHP CURLOPT_POSTFIELDS snippet.
func (h *Api) buildPHPBodyOption() string {
	switch h.CurlStruct.ContentType {
	case define.ContentTypeJson, define.ContentTypeText, define.ContentTypeRaw:
		bodyText := h.requestBodyText()
		if strings.TrimSpace(bodyText) == `` {
			return ``
		}
		return fmt.Sprintf("    CURLOPT_POSTFIELDS => %q,", bodyText)
	case define.ContentTypeForm:
		payload := make([]string, 0)
		for _, item := range h.sortedBodyForm() {
			payload = append(payload, fmt.Sprintf("%s=%s", item.Field, cast.ToString(normalizeBodyValue(item))))
		}
		return fmt.Sprintf("    CURLOPT_POSTFIELDS => %q,", strings.Join(payload, "&"))
	case define.ContentTypeMultiForm:
		lines := []string{"    CURLOPT_POSTFIELDS => ["}
		for _, item := range h.sortedBodyForm() {
			if item.Type == p_curl.FieldTypeFile || item.Type == `file` {
				lines = append(lines, fmt.Sprintf("        '%s' => new CURLFile('%s'),", escapeSingleQuote(item.Field), escapeSingleQuote(item.Value)))
				continue
			}
			lines = append(lines, fmt.Sprintf("        '%s' => '%s',", escapeSingleQuote(item.Field), escapeSingleQuote(cast.ToString(normalizeBodyValue(item)))))
		}
		lines = append(lines, "    ],")
		return strings.Join(lines, "\n")
	}
	return ``
}

// normalizeBodyValue 中文：把表单值转换成更贴近真实语义的类型。 English: Normalize form values into semantically closer types.
func normalizeBodyValue(item p_curl.KeyValue) any {
	switch item.Type {
	case p_curl.FieldTypeInt, `integer`:
		return cast.ToInt(item.Value)
	case p_curl.FieldTypeFloat:
		return cast.ToFloat64(item.Value)
	case p_curl.FieldTypeBool, `boolean`:
		return cast.ToBool(item.Value)
	default:
		return item.Value
	}
}

// escapeSingleQuote 中文：转义单引号，避免 shell/php 字符串截断。 English: Escape single quotes to keep shell/php strings valid.
func escapeSingleQuote(value string) string {
	return strings.ReplaceAll(value, `'`, `'\''`)
}

// escapeTemplateLiteral 中文：转义 JS 模板字符串里的特殊字符。 English: Escape special characters in JS template literals.
func escapeTemplateLiteral(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "`", "\\`")
	value = strings.ReplaceAll(value, "${", "\\${")
	return value
}

// escapePythonTripleSingle 中文：转义 Python 三引号单引号字符串。 English: Escape Python triple-single-quoted strings.
func escapePythonTripleSingle(value string) string {
	if !json.Valid([]byte(value)) {
		return strings.ReplaceAll(value, `'''`, `\'\'\'`)
	}
	return strings.ReplaceAll(value, `'''`, `\'\'\'`)
}
