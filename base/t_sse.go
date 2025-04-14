package base

import "C"
import (
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"strings"
)

type TSse struct {
	Sse *gsgin.TSse
}

type ChunkType string

const ChunkEnter ChunkType = `enter`
const ChunkNum ChunkType = `num`

type Chunk struct {
	Type ChunkType //num \n
	Num  int
}

func (h *TSse) SendMsg(sseClient, msg string, delayMills int) error {
	data := _struct.StreamData{
		Choices: []struct {
			Delta struct {
				Content string `json:"content"`
				Role    string `json:"role"`
			} `json:"delta"`
		}{
			{
				Delta: struct {
					Content string `json:"content"`
					Role    string `json:"role"`
				}{
					Content: msg,
					Role:    "",
				},
			},
		},
	}
	_ = h.Sse.Send(sseClient, `data: `+gstool.JsonEncode(data), delayMills)
	return nil
}

func (h *TSse) SendMsgChunk(sseClient, msg string, chunkT Chunk, delayMills int) error {
	var chunkList []string
	if chunkT.Type == ChunkNum {
		chunkList = gstool.SChunks(msg, chunkT.Num)
	} else {
		chunkList = strings.Split(msg, "\n")
	}
	nums := len(chunkList)
	for k, chunk := range chunkList {
		if k+1 == nums {
			chunk += "\n"
		}
		data := _struct.StreamData{
			Choices: []struct {
				Delta struct {
					Content string `json:"content"`
					Role    string `json:"role"`
				} `json:"delta"`
			}{
				{
					Delta: struct {
						Content string `json:"content"`
						Role    string `json:"role"`
					}{
						Content: chunk,
						Role:    "",
					},
				},
			},
		}
		_ = h.Sse.Send(sseClient, `data: `+gstool.JsonEncode(data), delayMills)
	}
	return nil
}
