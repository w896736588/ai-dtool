package main

import (
	"bytes"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	neturl "net/url"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gorilla/websocket"
)

func buildScrapeResultUploadRequest(endpoint, taskID, fileName string, content []byte, safeToken string) (*http.Request, error) {
	body := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("task_id", taskID); err != nil {
		return nil, err
	}
	fileWriter, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}
	if _, err = fileWriter.Write(content); err != nil {
		return nil, err
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if strings.TrimSpace(safeToken) != "" {
		request.Header.Set("Token", strings.TrimSpace(safeToken))
	}
	return request, nil
}

// WsClient WebSocket 客户端
type WsClient struct {
	config      Config
	conn        *websocket.Conn
	mu          sync.Mutex
	taskHandler func(msg define.AgentWsMessage)
	stopChan    chan struct{}
	reconnectMu sync.Mutex
}

// NewWsClient 创建 WebSocket 客户端
func NewWsClient(cfg Config) *WsClient {
	return &WsClient{
		config:   cfg,
		stopChan: make(chan struct{}),
	}
}

// SetTaskHandler 设置任务消息处理器
func (w *WsClient) SetTaskHandler(handler func(msg define.AgentWsMessage)) {
	w.taskHandler = handler
}

// Connect 建立 WebSocket 连接
func (w *WsClient) Connect() error {
	return w.connectWithRetry()
}

// connectWithRetry 带指数退避的重连
func (w *WsClient) connectWithRetry() error {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		url := fmt.Sprintf("%s/api/agent/ws?client_id=%s",
			w.wsURL(), w.config.ClientID)

		gstool.FmtPrintlnLogTime(`正在连接WebSocket url=%s`, url)

		conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
		if err == nil {
			w.mu.Lock()
			w.conn = conn
			w.mu.Unlock()

			gstool.FmtPrintlnLogTime(`WebSocket连接建立成功`)

			// 发送 hello
			w.Send(define.AgentWsMessage{
				Type: define.AgentWsMsgHello,
				Data: define.AgentHelloData{
					ClientVersion: w.config.ClientVersion,
					Hostname:      getHostname(),
					Os:            getOs(),
					Arch:          getArch(),
					UserName:      getUsername(),
					RuntimeReady:  false,
				},
			})

			// 启动读循环和心跳
			go w.readLoop()
			go w.heartbeatLoop()

			return nil
		}

		// 打印详细错误信息
		errDetail := err.Error()
		if resp != nil {
			errDetail = fmt.Sprintf("%s (HTTP %d)", errDetail, resp.StatusCode)
			// 尝试读取响应体
			if resp.Body != nil {
				buf := make([]byte, 1024)
				n, _ := resp.Body.Read(buf)
				if n > 0 {
					errDetail = fmt.Sprintf("%s body=%s", errDetail, string(buf[:n]))
				}
				resp.Body.Close()
			}
		}

		gstool.FmtPrintlnLogTime(`WebSocket连接失败: %s, %d秒后重试`, errDetail, backoff/time.Second)

		select {
		case <-time.After(backoff):
			backoff = backoff * 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		case <-w.stopChan:
			return fmt.Errorf("已关闭")
		}
	}
}

// readLoop 读消息循环
func (w *WsClient) readLoop() {
	defer func() {
		w.reconnectMu.Lock()
		defer w.reconnectMu.Unlock()
		gstool.FmtPrintlnLogTime(`WebSocket读循环退出，准备重连`)
		go w.connectWithRetry()
	}()

	for {
		w.mu.Lock()
		conn := w.conn
		w.mu.Unlock()

		if conn == nil {
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			gstool.FmtPrintlnLogTime(`WebSocket读错误: %s`, err.Error())
			return
		}

		var msg define.AgentWsMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			gstool.FmtPrintlnLogTime(`WebSocket消息解析失败: %s`, err.Error())
			continue
		}

		switch msg.Type {
		case define.AgentWsMsgTaskExecute:
			gstool.FmtPrintlnLogTime(`收到服务端任务下发 type=%s task_id=%s sse_distribute_id=%s`, msg.Type, msg.TaskID, msg.SseDistributeId)
			if w.taskHandler != nil {
				go w.taskHandler(msg)
			}
		case define.AgentWsMsgReadyCheck:
			// 响应就绪探测
			w.Send(define.AgentWsMessage{
				Type: define.AgentWsMsgHeartbeat,
				Data: define.AgentHeartbeatData{
					RuntimeReady: true,
				},
			})
		default:
			gstool.FmtPrintlnLogTime(`收到服务端消息 type=%s`, msg.Type)
		}
	}
}

// heartbeatLoop 心跳循环
func (w *WsClient) heartbeatLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.Send(define.AgentWsMessage{
				Type: define.AgentWsMsgHeartbeat,
				Data: define.AgentHeartbeatData{
					RuntimeReady:  true,
					CurrentTaskID: "",
				},
			})
		case <-w.stopChan:
			return
		}
	}
}

// Send 发送消息（线程安全）
func (w *WsClient) Send(msg define.AgentWsMessage) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conn == nil {
		gstool.FmtPrintlnLogTime(`发送消息被跳过，WebSocket未连接 type=%s task_id=%s`, msg.Type, msg.TaskID)
		return nil
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	gstool.FmtPrintlnLogTime(`发送消息到服务端 type=%s task_id=%s sse_distribute_id=%s`, msg.Type, msg.TaskID, msg.SseDistributeId)
	return w.conn.WriteMessage(websocket.TextMessage, data)
}

// SendStatus 发送客户端状态
func (w *WsClient) SendStatus(status string) {
	w.Send(define.AgentWsMessage{
		Type:     define.AgentWsMsgTaskStatus,
		ClientID: w.config.ClientID,
		Data: define.AgentTaskStatusData{
			Status: status,
		},
	})
}

// SendTaskLog 发送任务日志
func (w *WsClient) SendTaskLog(taskID, sseDistributeId, name, message string) {
	w.Send(define.AgentWsMessage{
		Type:            define.AgentWsMsgTaskLog,
		ClientID:        w.config.ClientID,
		TaskID:          taskID,
		SseDistributeId: sseDistributeId,
		Data: define.AgentTaskLogData{
			Name:    name,
			Message: message,
		},
	})
}

// SendTaskStatus 发送任务状态
func (w *WsClient) SendTaskStatus(taskID, sseDistributeId, status string) {
	w.Send(define.AgentWsMessage{
		Type:            define.AgentWsMsgTaskStatus,
		ClientID:        w.config.ClientID,
		TaskID:          taskID,
		SseDistributeId: sseDistributeId,
		Data: define.AgentTaskStatusData{
			Status: status,
		},
	})
}

// SendTaskResult 发送任务最终结果
func (w *WsClient) SendTaskResult(taskID, sseDistributeId, status, errMsg string) {
	w.SendTaskResultData(taskID, sseDistributeId, define.AgentTaskResultData{
		Status:       status,
		ErrorMessage: errMsg,
		FinishTime:   time.Now().Unix(),
	})
}

// SendTaskResultData 发送带扩展结果字段的任务最终结果。
func (w *WsClient) SendTaskResultData(taskID, sseDistributeId string, result define.AgentTaskResultData) {
	w.Send(define.AgentWsMessage{
		Type:            define.AgentWsMsgTaskResult,
		ClientID:        w.config.ClientID,
		TaskID:          taskID,
		SseDistributeId: sseDistributeId,
		Data:            result,
	})
}

// UploadScrapeResultFile 上传抓取任务 ZIP 结果文件到服务端。
func (w *WsClient) UploadScrapeResultFile(taskID, fileName string, content []byte, safeToken string) (*define.AgentTaskResultFileUploadResponse, error) {
	gstool.FmtPrintlnLogTime(`开始上传抓取结果文件 task_id=%s file_name=%s size=%d token_set=%t`, taskID, fileName, len(content), strings.TrimSpace(safeToken) != "")
	endpoint, err := neturl.Parse(strings.TrimRight(w.config.ServerURL, "/") + "/api/smart-link/task/result-file")
	if err != nil {
		return nil, err
	}
	gstool.FmtPrintlnLogTime(`抓取结果上传目标 task_id=%s endpoint=%s`, taskID, endpoint.String())
	request, err := buildScrapeResultUploadRequest(endpoint.String(), taskID, fileName, content, safeToken)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 60 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	gstool.FmtPrintlnLogTime(`抓取结果上传响应 task_id=%s status=%d body_len=%d`, taskID, response.StatusCode, len(responseBody))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("upload result file status=%d body=%s", response.StatusCode, string(responseBody))
	}
	result := struct {
		ErrCode int                                      `json:"ErrCode"`
		ErrMsg  string                                   `json:"ErrMsg"`
		Data    define.AgentTaskResultFileUploadResponse `json:"Data"`
	}{}
	if err = json.Unmarshal(responseBody, &result); err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	gstool.FmtPrintlnLogTime(`抓取结果上传完成 task_id=%s saved_file_name=%s download_url=%s`, taskID, result.Data.FileName, result.Data.DownloadURL)
	return &result.Data, nil
}

// Close 关闭连接
func (w *WsClient) Close() {
	close(w.stopChan)
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conn != nil {
		_ = w.conn.Close()
	}
}

// wsURL 将 HTTP URL 转换为 WebSocket URL
func (w *WsClient) wsURL() string {
	url := w.config.ServerURL
	if len(url) > 7 && url[:7] == "http://" {
		return "ws://" + url[7:]
	}
	if len(url) > 8 && url[:8] == "https://" {
		return "wss://" + url[8:]
	}
	return "ws://" + url
}
