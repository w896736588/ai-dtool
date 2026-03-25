package controller

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

type toolPortProcessItem struct {
	PID      int    `json:"pid"`
	Command  string `json:"command"`
	Protocol string `json:"protocol"`
	Address  string `json:"address"`
}

func ToolPortProcessList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	port, err := normalizeToolPort(dataMap[`port`])
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	items, err := queryToolPortProcesses(port)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`port`:  port,
		`items`: items,
	})
}

func ToolPortProcessKill(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	pid := cast.ToInt(dataMap[`pid`])
	if pid <= 0 {
		gsgin.GinResponseError(c, `pid不能为空`, nil)
		return
	}

	if err := killToolProcess(pid); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`pid`: pid,
	})
}

func normalizeToolPort(raw any) (int, error) {
	port := cast.ToInt(strings.TrimSpace(cast.ToString(raw)))
	if port <= 0 || port > 65535 {
		return 0, errors.New(`端口必须在 1-65535 之间`)
	}
	return port, nil
}

func queryToolPortProcesses(port int) ([]toolPortProcessItem, error) {
	switch runtime.GOOS {
	case `windows`:
		return queryToolPortProcessesWindows(port)
	case `linux`, `darwin`:
		return queryToolPortProcessesUnix(port)
	default:
		return nil, fmt.Errorf(`当前系统暂不支持: %s`, runtime.GOOS)
	}
}

func queryToolPortProcessesWindows(port int) ([]toolPortProcessItem, error) {
	raw, err := runToolCommand(`netstat`, `-ano`, `-p`, `tcp`)
	if err != nil {
		return nil, fmt.Errorf(`查询端口占用失败: %w`, err)
	}

	pidNameMap := make(map[int]string)
	for _, pid := range collectWindowsPIDList(raw, port) {
		name, nameErr := queryWindowsProcessName(pid)
		if nameErr == nil {
			pidNameMap[pid] = name
		}
	}

	return parseWindowsPortProcessRows(raw, port, pidNameMap), nil
}

func queryToolPortProcessesUnix(port int) ([]toolPortProcessItem, error) {
	raw, err := runToolCommand(`lsof`, `-nP`, fmt.Sprintf(`-iTCP:%d`, port), `-sTCP:LISTEN`)
	if err != nil {
		return nil, fmt.Errorf(`查询端口占用失败，请确认系统已安装 lsof: %w`, err)
	}
	return parseUnixPortProcessRows(raw), nil
}

func killToolProcess(pid int) error {
	switch runtime.GOOS {
	case `windows`:
		_, err := runToolCommand(`taskkill`, `/PID`, strconv.Itoa(pid), `/F`)
		if err != nil {
			return fmt.Errorf(`结束进程失败: %w`, err)
		}
		return nil
	case `linux`, `darwin`:
		_, err := runToolCommand(`kill`, `-9`, strconv.Itoa(pid))
		if err != nil {
			return fmt.Errorf(`结束进程失败: %w`, err)
		}
		return nil
	default:
		return fmt.Errorf(`当前系统暂不支持: %s`, runtime.GOOS)
	}
}

func runToolCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == `` {
			errMsg = err.Error()
		}
		return ``, errors.New(errMsg)
	}
	return stdout.String(), nil
}

func collectWindowsPIDList(raw string, port int) []int {
	rows := parseWindowsPortProcessRows(raw, port, nil)
	seen := make(map[int]struct{}, len(rows))
	list := make([]int, 0, len(rows))
	for _, row := range rows {
		if row.PID <= 0 {
			continue
		}
		if _, ok := seen[row.PID]; ok {
			continue
		}
		seen[row.PID] = struct{}{}
		list = append(list, row.PID)
	}
	sort.Ints(list)
	return list
}

func queryWindowsProcessName(pid int) (string, error) {
	raw, err := runToolCommand(`tasklist`, `/FI`, fmt.Sprintf(`PID eq %d`, pid), `/FO`, `CSV`, `/NH`)
	if err != nil {
		return ``, err
	}
	reader := csv.NewReader(strings.NewReader(strings.TrimSpace(raw)))
	record, err := reader.Read()
	if err != nil || len(record) == 0 {
		return ``, errors.New(`tasklist 输出解析失败`)
	}
	if strings.EqualFold(record[0], `INFO: No tasks are running which match the specified criteria.`) {
		return ``, errors.New(`未找到进程`)
	}
	return strings.TrimSpace(record[0]), nil
}

func parseWindowsPortProcessRows(raw string, port int, pidNameMap map[int]string) []toolPortProcessItem {
	lines := strings.Split(raw, "\n")
	list := make([]toolPortProcessItem, 0, len(lines))
	for _, line := range lines {
		clean := strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))
		if clean == `` {
			continue
		}
		fields := strings.Fields(clean)
		if len(fields) < 5 {
			continue
		}
		if !strings.EqualFold(fields[0], `TCP`) {
			continue
		}
		if !strings.EqualFold(fields[3], `LISTENING`) {
			continue
		}
		if extractPortFromAddress(fields[1]) != port {
			continue
		}
		pid, err := strconv.Atoi(fields[4])
		if err != nil {
			continue
		}
		item := toolPortProcessItem{
			PID:      pid,
			Command:  strings.TrimSpace(pidNameMap[pid]),
			Protocol: strings.ToLower(fields[0]),
			Address:  fields[1],
		}
		list = append(list, item)
	}
	return list
}

func parseUnixPortProcessRows(raw string) []toolPortProcessItem {
	lines := strings.Split(raw, "\n")
	list := make([]toolPortProcessItem, 0, len(lines))
	for _, line := range lines {
		clean := strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))
		if clean == `` || strings.HasPrefix(clean, `COMMAND `) {
			continue
		}
		fields := strings.Fields(clean)
		if len(fields) < 9 {
			continue
		}
		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		list = append(list, toolPortProcessItem{
			PID:      pid,
			Command:  fields[0],
			Protocol: strings.ToLower(fields[len(fields)-2-1]),
			Address:  fields[len(fields)-2],
		})
	}
	return list
}

func extractPortFromAddress(address string) int {
	address = strings.TrimSpace(address)
	index := strings.LastIndex(address, `:`)
	if index <= 0 || index >= len(address)-1 {
		return 0
	}
	port, _ := strconv.Atoi(address[index+1:])
	return port
}
