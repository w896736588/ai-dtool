package main

import (
	"bytes"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type agentSmartLinkLastStore struct {
	serverURL string
	safeToken string
	client    *http.Client
}

// newAgentSmartLinkLastStore 创建 agent 侧的历史目录存储适配器。
// agent 不直连配置库，所有 tbl_smart_link_last 读写都通过服务端接口完成。
func newAgentSmartLinkLastStore(serverURL, safeToken string) *agentSmartLinkLastStore {
	return &agentSmartLinkLastStore{
		serverURL: strings.TrimRight(serverURL, "/"),
		safeToken: strings.TrimSpace(safeToken),
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

// GetLastUserDataIndex 查询指定用户和域名上次使用的数据目录索引。
func (h *agentSmartLinkLastStore) GetLastUserDataIndex(userName, domain string) (int, error) {
	resp, err := h.do(define.AgentSmartLinkLastRequest{
		Action:   define.AgentSmartLinkLastActionGetLast,
		UserName: userName,
		Domain:   domain,
	})
	if err != nil {
		return 0, err
	}
	return resp.UserDataIndex, nil
}

// ExistDomainUserDataIndex 判断某个域名是否已经占用了指定数据目录索引。
func (h *agentSmartLinkLastStore) ExistDomainUserDataIndex(domain string, userDataIndex int) (bool, error) {
	resp, err := h.do(define.AgentSmartLinkLastRequest{
		Action:        define.AgentSmartLinkLastActionExists,
		Domain:        domain,
		UserDataIndex: userDataIndex,
	})
	if err != nil {
		return false, err
	}
	return resp.Exists, nil
}

// UpsertLastUserDataIndex 记录当前任务实际使用的数据目录索引，供下次复用登录态。
func (h *agentSmartLinkLastStore) UpsertLastUserDataIndex(smartLinkID int, userName, domain string, userDataIndex int) error {
	_, err := h.do(define.AgentSmartLinkLastRequest{
		Action:        define.AgentSmartLinkLastActionUpsert,
		SmartLinkID:   smartLinkID,
		UserName:      userName,
		Domain:        domain,
		UserDataIndex: userDataIndex,
	})
	return err
}

// do 统一调用服务端代理接口，并解析项目标准响应结构。
func (h *agentSmartLinkLastStore) do(payload define.AgentSmartLinkLastRequest) (define.AgentSmartLinkLastResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return define.AgentSmartLinkLastResponse{}, err
	}
	request, err := http.NewRequest(http.MethodPost, h.serverURL+"/api/smart-link/agent/last-user-data", bytes.NewReader(body))
	if err != nil {
		return define.AgentSmartLinkLastResponse{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	if h.safeToken != "" {
		request.Header.Set("Token", h.safeToken)
	}
	response, err := h.client.Do(request)
	if err != nil {
		return define.AgentSmartLinkLastResponse{}, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return define.AgentSmartLinkLastResponse{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return define.AgentSmartLinkLastResponse{}, fmt.Errorf("last-user-data status=%d body=%s", response.StatusCode, string(responseBody))
	}
	result := struct {
		ErrCode int                               `json:"ErrCode"`
		ErrMsg  string                            `json:"ErrMsg"`
		Data    define.AgentSmartLinkLastResponse `json:"Data"`
	}{}
	if err = json.Unmarshal(responseBody, &result); err != nil {
		return define.AgentSmartLinkLastResponse{}, err
	}
	if result.ErrCode != 0 {
		if result.ErrMsg == "" {
			result.ErrMsg = "last-user-data接口返回失败"
		}
		return define.AgentSmartLinkLastResponse{}, errors.New(result.ErrMsg)
	}
	return result.Data, nil
}
