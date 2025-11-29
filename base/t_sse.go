package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"strings"

	"gitee.com/Sxiaobai/gs/gstool"
	"gitee.com/Sxiaobai/gs/v2/gsgin"
)

type TSse struct {
	Sse *gsgin.TSse
}

func (h *TSse) SendMsg(sseClient, contentType string, msg any, delayMills int, sseIds ...string) error {
	data := define.SseData{
		SseClientId: sseClient,
		Type:        contentType,
		Data:        msg,
	}
	if len(sseIds) == 0 {
		sseIds = define.SseClientIds
	}
	for _, sseId := range sseIds {
		_ = h.Sse.Send(sseId, `data: `+gstool.JsonEncode(data), delayMills)
	}
	return nil
}

func (h *TSse) SendMsgChunk(sseClient, msg string, chunkT _struct.Chunk, delayMills int, sseIds ...string) error {
	var chunkList []string
	split := ``
	if len(sseIds) == 0 {
		sseIds = define.SseClientIds
	}
	if chunkT.Type == define.ChunkNum {
		if chunkT.Num == 0 {
			chunkList = append(chunkList, msg)
		} else {
			chunkList = gstool.SChunks(msg, chunkT.Num)
		}

	} else if chunkT.Type == define.ChunkEnter {
		if chunkT.Split == `` {
			split = "\n"
		}
		chunkList = strings.Split(msg, split)
	} else if chunkT.Type == define.ChunkR {
		if chunkT.Split == `` {
			split = "\r"
		}
		chunkList = strings.Split(msg, split)
	}
	nums := len(chunkList)
	for k, chunk := range chunkList {
		if k+1 == nums {
			chunk += "\n"
		}
		data := define.SseData{
			SseClientId: sseClient,
			Data:        chunk,
			Type:        define.SseContentTypeMsg,
		}

		for _, sseId := range sseIds {
			_ = h.Sse.Send(sseId, `data: `+gstool.JsonEncode(data), delayMills)
		}

	}
	return nil
}

func (h *TSse) SendMsgChunkList(sseClient string, chunkList []string, delayMills int) error {
	nums := len(chunkList)
	for k, chunk := range chunkList {
		if k+1 == nums {
			chunk += "\n"
		}
		data := define.SseData{
			SseClientId: sseClient,
			Data:        chunk,
			Type:        define.SseContentTypeMsg,
		}
		for _, sseClientId := range define.SseClientIds {
			_ = h.Sse.Send(sseClientId, `data: `+gstool.JsonEncode(data), delayMills)
		}
	}
	return nil
}
