package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

type TSse struct {
	Sse *gsgin.TSse
}

func (h *TSse) SendMsg(sseClient, contentType string, msg any, delayMills int) error {
	data := define.SseData{
		SseClientId: sseClient,
		Data:        msg,
		Type:        contentType,
	}
	_ = h.Sse.Send(define.SseIdDistribute, `data: `+gstool.JsonEncode(data), delayMills)
	return nil
}

func (h *TSse) SendMsgChunk(sseClient, msg string, chunkT _struct.Chunk, delayMills int) error {
	var chunkList []string
	split := ``
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
		_ = h.Sse.Send(define.SseIdDistribute, `data: `+gstool.JsonEncode(data), delayMills)
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
		_ = h.Sse.Send(define.SseIdDistribute, `data: `+gstool.JsonEncode(data), delayMills)
	}
	return nil
}
