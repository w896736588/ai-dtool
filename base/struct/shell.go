package _struct

import "dev_tool/base/define"

type Chunk struct {
	Type  define.ChunkType //num \n
	Num   int
	Split string //分割符
}
