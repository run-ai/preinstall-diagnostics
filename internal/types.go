package internal

type TestResult struct {
	Name    string `json:"name,omitempty"`
	Result  bool   `json:"result,omitempty"`
	Message string `json:"message,omitempty"`
}
