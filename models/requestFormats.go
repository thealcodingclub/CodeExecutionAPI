package models

type ExecuteRequest struct {
	Language  string   `json:"language"`
	Code      string   `json:"code"`
	Timeout   int      `json:"timeout"`
	MaxMemory int      `json:"max_memory"`
	Inputs    []string `json:"inputs,omitempty"`
}

type ExecuteResponse struct {
	Output     string `json:"output"`
	Error      string `json:"error"`
	MemoryUsed string `json:"memory_used"`
	CpuTime    string `json:"cpu_time"`
}
