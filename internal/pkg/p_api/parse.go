package p_api

import (
	"dev_tool/base/define"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// CurlCommand 表示解析后的curl命令结构
type CurlCommand struct {
	URL            string
	Method         string
	Headers        map[string]string
	ContentType    string
	Data           string
	FormDataFields map[string]string // 专门存储multipart/form-data字段
	Form           []KeyValue
	Cookies        string
	UserAgent      string
	Timeout        int
	FollowRedirect bool
	Verbose        bool
	Insecure       bool
}

// ParseCurlCommand 解析curl命令字符串
func ParseCurlCommand(curlCmd string) (*CurlCommand, error) {
	cmd := &CurlCommand{
		Method:         "GET",
		Headers:        make(map[string]string),
		Form:           make([]KeyValue, 0),
		FormDataFields: make(map[string]string),
		Timeout:        30,
		FollowRedirect: true,
	}

	// 移除多余的空格和换行符
	curlCmd = strings.TrimSpace(curlCmd)
	curlCmd = regexp.MustCompile(`\\\s*\n\s*`).ReplaceAllString(curlCmd, " ")
	curlCmd = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(curlCmd), " ")

	// 分割命令参数
	args := splitCurlArgs(curlCmd)

	i := 0
	for i < len(args) {
		arg := args[i]

		switch arg {
		case "curl":
			// 跳过curl命令本身
		case "-X", "--request":
			i++
			if i < len(args) {
				cmd.Method = strings.ToUpper(args[i])
			}
		case "-H", "--header":
			i++
			if i < len(args) {
				header := args[i]
				parts := strings.SplitN(header, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					cmd.Headers[key] = value
					// 检查是否是Content-Type头部
					if strings.ToLower(key) == "content-type" {
						cmd.ContentType = value
					}
				}
			}
		case "-d", "--data", "--data-raw", "--data-ascii":
			i++
			if i < len(args) {
				cmd.Data = args[i]
				if cmd.Method == "GET" {
					cmd.Method = "POST"
				}

				// 如果没有显式设置Content-Type，根据数据类型设置默认值
				if cmd.ContentType == "" {
					cmd.ContentType = "application/x-www-form-urlencoded"
					cmd.Headers["Content-Type"] = cmd.ContentType
				}

				// 检查是否是multipart/form-data格式
				if contentType, exists := cmd.Headers["Content-Type"]; exists &&
					strings.Contains(strings.ToLower(contentType), "multipart/form-data") {
					// 处理$'...'格式的转义序列
					processedData := processStringEscapes(cmd.Data)
					cmd.parseMultipartFormData(processedData)
				}
			}
		case "--data-urlencode":
			i++
			if i < len(args) {
				cmd.Data = args[i]
				if cmd.Method == "GET" {
					cmd.Method = "POST"
				}

				// 如果没有显式设置Content-Type，设置默认值
				if cmd.ContentType == "" {
					cmd.ContentType = "application/x-www-form-urlencoded"
					cmd.Headers["Content-Type"] = cmd.ContentType
				}
			}
		case "-F", "--form":
			i++
			if i < len(args) {
				formData := args[i]
				field := cmd.parseFormField(formData)
				cmd.Form = append(cmd.Form, field)
				if cmd.Method == "GET" {
					cmd.Method = "POST"
				}

				// 设置multipart/form-data的Content-Type
				if cmd.ContentType == "" || !strings.Contains(strings.ToLower(cmd.ContentType), "multipart/form-data") {
					// 生成boundary
					boundary := cmd.generateBoundary()
					cmd.ContentType = fmt.Sprintf("multipart/form-data; boundary=%s", boundary)
					cmd.Headers["Content-Type"] = cmd.ContentType
				}
			}
		case "-u", "--user":
			i++
			if i < len(args) {
				userPass := args[i]
				// Basic认证需要base64编码
				cmd.Headers["Authorization"] = fmt.Sprintf("Basic %s", userPass)
			}
		case "-c", "--cookie-jar":
			i++
			// 忽略cookie jar文件路径
		case "-b", "--cookie":
			i++
			if i < len(args) {
				cmd.Cookies = args[i]
			}
		case "-A", "--user-agent":
			i++
			if i < len(args) {
				cmd.UserAgent = args[i]
			}
		case "-m", "--max-time", "--connect-timeout":
			i++
			if i < len(args) {
				if timeout, err := parseInt(args[i]); err == nil {
					cmd.Timeout = timeout
				}
			}
		case "-L", "--location":
			cmd.FollowRedirect = true
		case "-v", "--verbose":
			cmd.Verbose = true
		case "-k", "--insecure":
			cmd.Insecure = true
		case "--compressed":
			cmd.Headers["Accept-Encoding"] = "deflate, gzip"
		case "--basic":
			// 如果Authorization头已经存在，则不覆盖
			if _, exists := cmd.Headers["Authorization"]; !exists {
				cmd.Headers["Authorization"] = "Basic"
			}
		case "--digest":
			// 如果Authorization头已经存在，则不覆盖
			if _, exists := cmd.Headers["Authorization"]; !exists {
				cmd.Headers["Authorization"] = "Digest"
			}
		default:
			// 检查是否是URL（不以-开头且包含http/https）
			if strings.HasPrefix(arg, "http") {
				cmd.URL = arg
			} else if !strings.HasPrefix(arg, "-") && !isValidFlag(arg) {
				// 可能是URL
				cmd.URL = arg
			}
		}
		i++
	}

	// 如果没有设置Content-Type但有数据，设置默认值
	if cmd.ContentType == "" && cmd.Data != "" {
		cmd.ContentType = "application/x-www-form-urlencoded"
		cmd.Headers["Content-Type"] = cmd.ContentType
	}

	// 验证URL是否为空
	if cmd.URL == "" {
		return nil, fmt.Errorf("no URL found in curl command")
	}
	if strings.Contains(cmd.ContentType, define.ContentTypeForm) {
		cmd.ContentType = define.ContentTypeForm
	} else if strings.Contains(cmd.ContentType, define.ContentTypeJson) {
		cmd.ContentType = define.ContentTypeJson
	} else if strings.Contains(cmd.ContentType, define.ContentTypeMultiForm) {
		cmd.ContentType = define.ContentTypeMultiForm
	} else if strings.Contains(cmd.ContentType, define.ContentTypeRaw) {
		cmd.ContentType = define.ContentTypeRaw
	} else if strings.Contains(cmd.ContentType, define.ContentTypeText) {
		cmd.ContentType = define.ContentTypeText
	}
	return cmd, nil
}

// generateBoundary 生成multipart/form-data的boundary
func (c *CurlCommand) generateBoundary() string {
	// 简单的boundary生成，实际curl可能使用更复杂的算法
	return fmt.Sprintf("----WebKitFormBoundary%d", len(c.FormDataFields)+len(c.Form))
}

// processStringEscapes 处理字符串中的转义序列，如$'...'中的\r\n
func processStringEscapes(s string) string {
	// 检查是否是$'...'格式
	if strings.HasPrefix(s, "$'") && strings.HasSuffix(s, "'") {
		s = s[2 : len(s)-1] // 去掉$'和'
	}

	// 处理常见的转义序列
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	s = strings.ReplaceAll(s, "\\\"", "\"")

	return s
}

// parseMultipartFormData 解析multipart/form-data格式的数据
func (c *CurlCommand) parseMultipartFormData(data string) {
	// 查找boundary
	boundary := ""
	if contentType, exists := c.Headers["Content-Type"]; exists {
		re := regexp.MustCompile(`boundary=([^;\s]+)`)
		matches := re.FindStringSubmatch(contentType)
		if len(matches) > 1 {
			boundary = matches[1]
			// 如果boundary以"开始和结束，需要去掉引号
			if strings.HasPrefix(boundary, `"`) && strings.HasSuffix(boundary, `"`) {
				boundary = boundary[1 : len(boundary)-1]
			}
		}
	}

	if boundary == "" {
		return
	}

	// 分割数据 - 按边界分割
	parts := strings.Split(data, "--"+boundary)

	for _, part := range parts {
		// 跳过边界标记和结束标记
		if strings.TrimSpace(part) == "" || strings.TrimSpace(part) == "--" {
			continue
		}

		// 检查是否是结束标记
		if strings.HasSuffix(strings.TrimSpace(part), "--") {
			part = strings.TrimSuffix(strings.TrimSpace(part), "--")
		}

		// 按行分割
		lines := strings.Split(part, "\n")
		if len(lines) == 0 {
			continue
		}

		// 找到Content-Disposition行
		var fieldName string
		var valueLines []string

		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(strings.ToLower(trimmedLine), "content-disposition:") {
				// 提取name
				nameRe := regexp.MustCompile(`name="([^"]+)"`)
				nameMatches := nameRe.FindStringSubmatch(trimmedLine)
				if len(nameMatches) > 1 {
					fieldName = nameMatches[1]
				}
				// 跳过头部，从下一行开始是值
				valueLines = lines[i+1:]
				break
			}
		}

		// 寻找空行，空行后是实际值
		if fieldName != "" {
			for i, line := range valueLines {
				if strings.TrimSpace(line) == "" {
					// 空行之后是实际值
					actualValue := strings.Join(valueLines[i+1:], "\n")
					// 去除末尾的空行
					actualValue = strings.TrimRight(actualValue, "\n\r ")
					c.FormDataFields[fieldName] = actualValue
					break
				}
			}
		}
	}
}

// splitCurlArgs 将curl命令分割为参数列表，处理引号
func splitCurlArgs(cmd string) []string {
	var args []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	i := 0

	for i < len(cmd) {
		char := cmd[i]
		if char == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			i++
			continue
		}
		if char == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			i++
			continue
		}
		if char == ' ' && !inSingleQuote && !inDoubleQuote {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
			i++
			continue
		}
		current.WriteByte(char)
		i++
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

// parseInt 尝试将字符串转换为整数
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// isValidFlag 检查字符串是否是有效的curl标志
func isValidFlag(s string) bool {
	if len(s) < 2 || s[0] != '-' {
		return false
	}
	// 检查是否是有效的curl标志格式
	return true
}

// String 返回curl命令的字符串表示
func (c *CurlCommand) String() string {
	jsonStr := fmt.Sprintf(`{
  "URL": "%s",
  "Method": "%s",
  "Headers": %s,
  "ContentType": "%s",
  "Data": %s,
  "FormDataFields": %s,
  "Form": %s,
  "Cookies": "%s",
  "UserAgent": "%s",
  "Timeout": %d,
  "FollowRedirect": %t,
  "Verbose": %t,
  "Insecure": %t
}`, c.URL, c.Method, formatMap(c.Headers), c.ContentType, formatString(c.Data), formatMap(c.FormDataFields), formatSlice(c.Form), c.Cookies, c.UserAgent, c.Timeout, c.FollowRedirect, c.Verbose, c.Insecure)
	return jsonStr
}

// formatString 格式化字符串为JSON字符串
func formatString(s string) string {
	// 转义特殊字符
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return fmt.Sprintf(`"%s"`, s)
}

// formatMap 格式化map为JSON字符串
func formatMap(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}

	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ","
		}
		result += fmt.Sprintf(`"%s":%s`, k, formatString(v))
		first = false
	}
	result += "}"
	return result
}

// formatSlice 格式化KeyValue切片为JSON字符串
func formatSlice(s []KeyValue) string {
	if len(s) == 0 {
		return "[]"
	}

	result := "["
	for i, kv := range s {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`{"Field":%s,"Type":%s,"Value":%s,"Description":%s}`,
			formatString(kv.Field), formatString(kv.Type), formatString(kv.Value), formatString(kv.Description))
	}
	result += "]"
	return result
}

// ToCurlCommand 生成等效的curl命令字符串
func (c *CurlCommand) ToCurlCommand() string {
	var result strings.Builder
	result.WriteString("curl")
	if c.Method != "GET" {
		result.WriteString(fmt.Sprintf(" -X %s", c.Method))
	}

	// 按顺序添加头部，保留原始顺序
	// 首先添加非Content-Type头部
	for key, value := range c.Headers {
		if strings.ToLower(key) != "content-type" {
			// 转义引号以避免命令注入
			escapedValue := strings.ReplaceAll(value, "\"", "\\\"")
			result.WriteString(fmt.Sprintf(" -H \"%s: %s\"", key, escapedValue))
		}
	}

	// 然后处理Content-Type头部和表单数据
	contentType, hasContentType := c.Headers["Content-Type"]
	if hasContentType && strings.Contains(strings.ToLower(contentType), "multipart/form-data") {
		// 如果是multipart/form-data，只添加Content-Type头部
		escapedValue := strings.ReplaceAll(contentType, "\"", "\\\"")
		result.WriteString(fmt.Sprintf(" -H \"Content-Type: %s\"", escapedValue))

		// 添加所有表单字段
		for fieldName, fieldValue := range c.FormDataFields {
			// 转义特殊字符
			escapedField := strings.ReplaceAll(fieldName, "\"", "\\\"")
			escapedValue := strings.ReplaceAll(fieldValue, "\"", "\\\"")
			result.WriteString(fmt.Sprintf(" -F \"%s=%s\"", escapedField, escapedValue))
		}
	} else {
		// 对于非multipart/form-data，按原样添加Content-Type头部
		if hasContentType {
			escapedValue := strings.ReplaceAll(contentType, "\"", "\\\"")
			result.WriteString(fmt.Sprintf(" -H \"Content-Type: %s\"", escapedValue))
		}

		// 添加其他表单字段（如果有的话）
		if len(c.Form) > 0 {
			for _, formField := range c.Form {
				escapedValue := strings.ReplaceAll(formField.Value, "\"", "\\\"")
				escapedField := strings.ReplaceAll(formField.Field, "\"", "\\\"")

				switch formField.Type {
				case FieldTypeFile:
					result.WriteString(fmt.Sprintf(" -F \"%s=@%s\"", escapedField, escapedValue))
				case FieldTypeInt, FieldTypeFloat, FieldTypeBool, FieldTypeString:
					// 对于字符串，如果包含空格或特殊字符，需要加引号
					if strings.ContainsAny(formField.Value, " \t\"'&|<>") {
						result.WriteString(fmt.Sprintf(" -F \"%s=%s\"", escapedField, escapedValue))
					} else {
						result.WriteString(fmt.Sprintf(" -F %s=%s", escapedField, escapedValue))
					}
				default:
					result.WriteString(fmt.Sprintf(" -F \"%s=%s\"", escapedField, escapedValue))
				}
			}
		}

		// 添加原始数据（如果不是multipart/form-data）
		if c.Data != "" {
			escapedData := strings.ReplaceAll(c.Data, "'", "'\"'\"'")
			result.WriteString(fmt.Sprintf(" -d '%s'", escapedData))
		}
	}

	if c.Cookies != "" {
		escapedCookies := strings.ReplaceAll(c.Cookies, "'", "'\"'\"'")
		result.WriteString(fmt.Sprintf(" -b '%s'", escapedCookies))
	}
	if c.UserAgent != "" {
		escapedUserAgent := strings.ReplaceAll(c.UserAgent, "'", "'\"'\"'")
		result.WriteString(fmt.Sprintf(" -A '%s'", escapedUserAgent))
	}
	if c.Timeout != 30 {
		result.WriteString(fmt.Sprintf(" -m %d", c.Timeout))
	}
	if !c.FollowRedirect {
		result.WriteString(" --no-location")
	}
	if c.Verbose {
		result.WriteString(" -v")
	}
	if c.Insecure {
		result.WriteString(" -k")
	}
	// 转义URL以避免命令注入
	escapedURL := strings.ReplaceAll(c.URL, "'", "'\"'\"'")
	result.WriteString(fmt.Sprintf(" '%s'", escapedURL))
	return result.String()
}

// parseFormField 修复版 - 正确处理引号
func (c *CurlCommand) parseFormField(formData string) KeyValue {
	field := KeyValue{
		Field:       "",
		Type:        FieldTypeString,
		Value:       "",
		Description: "",
	}

	// 1. 文件上传：field=@filename;type=content/type
	if strings.Contains(formData, "=@") {
		parts := strings.SplitN(formData, "=@", 2)
		if len(parts) == 2 {
			field.Field = strings.Trim(parts[0], `"`) // 去除字段名的引号
			field.Type = FieldTypeFile

			// 处理可能的内容类型
			fileParts := strings.SplitN(parts[1], ";type=", 2)
			field.Value = strings.Trim(fileParts[0], `"`) // 去除文件路径的引号
			if len(fileParts) > 1 {
				field.Description = fmt.Sprintf("File upload with content-type: %s", fileParts[1])
			} else {
				field.Description = fmt.Sprintf("File upload: %s", fileParts[0])
			}
		}
	} else if strings.Contains(formData, "=") { // 2. 普通键值对：field=value
		parts := strings.SplitN(formData, "=", 2)
		if len(parts) == 2 {
			field.Field = strings.Trim(parts[0], `"`) // 去除字段名的引号
			fieldValue := strings.Trim(parts[1], `"`) // 去除字段值的引号

			// 处理转义字符
			fieldValue = processFormValue(fieldValue)
			field.Value = fieldValue
			field.Type = c.getValueType(fieldValue)
			field.Description = fmt.Sprintf("Form field: %s", field.Field)
		}
	} else { // 3. 只有字段名
		field.Field = strings.Trim(formData, `"`)
		field.Type = FieldTypeString
		field.Description = fmt.Sprintf("Form field: %s", formData)
	}
	return field
}

// processFormValue 处理表单字段值中的转义字符
func processFormValue(value string) string {
	// 如果值被引号包围，去掉引号
	value = strings.Trim(value, `"`)

	// 处理常见的转义序列
	value = strings.ReplaceAll(value, `\\`, `\`)
	value = strings.ReplaceAll(value, `\"`, `"`)
	value = strings.ReplaceAll(value, `\n`, "\n")
	value = strings.ReplaceAll(value, `\r`, "\r")
	value = strings.ReplaceAll(value, `\t`, "\t")

	return value
}

// getValueType 判断值的类型
func (c *CurlCommand) getValueType(value string) string {
	// 检查是否是整数
	if _, err := strconv.Atoi(value); err == nil {
		return FieldTypeInt
	}
	// 检查是否是浮点数
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return FieldTypeFloat
	}
	// 检查是否是布尔值
	if value == "true" || value == "false" {
		return FieldTypeBool
	}
	return FieldTypeString
}
