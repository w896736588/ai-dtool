package base

import (
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"os"
)

type TJas struct {
	Regis  map[string]string
	JsData map[string]string
}

func (h *TJas) Load() {
	for name, path := range h.Regis {
		workErr := gstool.DirWalk(path, func(path string, info os.FileInfo, err error) {
			if info.IsDir() {
				return
			}
			if err != nil {
				gstool.FmtPrintlnLogTime(`加载%s %s目录失败 %s`, name, path, err.Error())
				return
			}
			content, contentErr := gstool.FileGetContent(path)
			if contentErr != nil {
				gstool.FmtPrintlnLogTime(`读取文件内容失败%s %s`, path, contentErr.Error())
				return
			}
			h.JsData[name+"/"+info.Name()] = content
		})
		if workErr != nil {
			gstool.FmtPrintlnLogTime(`获取文件夹下面的文件失败%s %s`, path, workErr.Error())
		}
	}
}

func (h *TJas) Get(name, fileName string) string {
	if data, ok := h.JsData[name+"/"+fileName]; ok {
		return data
	}
	return ``
}
