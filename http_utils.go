package main

const (
	MESSAGE_ROLE_USER      = "user"
	MESSAGE_ROLE_ASSISTANT = "assistant"

	GORQ_REQ_URL = "https://api.groq.com/openai/v1/chat/completions"

	PROMPT = `You are a network graph maker who extracts terms and their relations from a given context.
	You are provided with a context chunk (delimited by ) Your task is to extract the ontology
	of terms mentioned in the given context. These terms should represent the key concepts as per the context.
	Thought 1: While traversing through each sentence, Think about the key terms mentioned in it.
		Terms may include object, entity, location, organization, person,
		condition, acronym, documents, service, concept, etc.
		Terms should be as atomistic as possible
	Thought 2: Think about how these terms can have one on one relation with other terms.
	   Terms that are mentioned in the same sentence or the same paragraph are typically related to each other.
	   Terms can be related to many other terms\n\n"
	Thought 3: Find out the relation between each such related pair of terms.
	Format your output as a list of json. Each element of the list contains a pair of terms
	and the relation between them, like the follwing. don't change following json structure. Edge relation
	should be short:
	{
		"data": [
			{
				"node_1": "A concept from extracted ontology",
				"node_2": "A related concept from extracted ontology",
				"edge": "relationship between the two concepts, node_1 and node_2 in less than 5 words"
			}, {...}\n
		]
	}
	
	Content:
	-------------------------
	`
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
