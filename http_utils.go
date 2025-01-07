package main

const (
	MESSAGE_ROLE_USER      = "user"
	MESSAGE_ROLE_ASSISTANT = "assistant"

	GORQ_REQ_URL = "https://api.groq.com/openai/v1/chat/completions"
)

type ResponseFormat struct {
	Type string `json:"type"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GorqRequest struct {
	Messages       []Message       `json:"messages"`
	Model          string          `json:"model"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Temperature    float64         `json:"temperature,omitempty"`
	TopP           int             `json:"top_p,omitempty"`
	Stream         bool            `json:"stream,omitempty"`
	Stop           string          `json:"stop,omitempty"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

func NewGorqRequest(content string) *GorqRequest {
	return &GorqRequest{
		Messages: []Message{{
			Role:    MESSAGE_ROLE_USER,
			Content: content,
		}},
		Model:          "llama3-8b-8192",
		Temperature:    1,
		MaxTokens:      1024,
		TopP:           1,
		ResponseFormat: &ResponseFormat{Type: "json_object"},
	}
}

type Edge struct {
	Node1 string `json:"node_1"`
	Node2 string `json:"node_2"`
	Edge  string `json:"edge"`
}

type ResponseMessage struct {
	Data []Edge `json:"data"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	Delta        Message `json:"delta"`
	FinishReason string  `json:"finish_reason"`
}

type GorqResponseError struct {
	Message         string `json:"message"`
	Type            string `json:"type"`
	Code            string `json:"code"`
	FailedGenerated string `json:"failed_generation"`
}

type GorqResponse struct {
	Error             GorqResponseError `json:"error"`
	ID                string            `json:"id"`
	Object            string            `json:"object"`
	Created           int64             `json:"created"`
	Model             string            `json:"model"`
	SystemFingerprint string            `json:"system_fingerprint"`
	Choices           []Choice          `json:"choices"`
}
