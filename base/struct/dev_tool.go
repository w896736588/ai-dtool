package _struct

type DevToolEvent struct {
	Args []struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"args"`
	ExecutionContextId int `json:"executionContextId"`
	StackTrace         struct {
		CallFrames []struct {
			ColumnNumber int    `json:"columnNumber"`
			FunctionName string `json:"functionName"`
			LineNumber   int    `json:"lineNumber"`
			ScriptId     string `json:"scriptId"`
			Url          string `json:"url"`
		} `json:"callFrames"`
	} `json:"stackTrace"`
	Timestamp float64 `json:"timestamp"`
	Type      string  `json:"type"`
}
