package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
)

func main() {
	llmModel := flag.String("model", "llama3-8b-8192", "input data")
	inputPath := flag.String("i", "", "input data")
	promptOnly := flag.Bool("promptonly", false, "Just print prompt to use in another LLM models")
	flag.Parse()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Unable to load environments: %s", err)
	}

	prompt, err := generatePrompt(*inputPath)
	if err != nil {
		log.Fatalf("unable to convert pdf file to text: %s", err)
	}

	if *promptOnly {
		fmt.Println(prompt)
		return
	}

	data, err := GorqSendRequest(prompt, os.Getenv("API_KEY"), *llmModel)
	if err != nil {
		log.Fatalf("unable to get response from gorq: %s", err)
	}

	err = visualizeData(data)
	if err != nil {
		log.Fatalf("unable to visualize data: %s", err)
	}
}

func generatePrompt(pdfPath string) (string, error) {
	prompt := `You are a network graph maker who extracts terms and their relations from a given context.
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

	cmd := exec.Command("pdftotext", pdfPath, "-")
	out, err := cmd.Output()

	return prompt + string(out), err
}

func GorqSendRequest(content string, token string, model string) (ResponseMessage, error) {
	body := GorqRequest{
		Messages: []Message{{
			Role:    MESSAGE_ROLE_USER,
			Content: content,
		}},
		Model:          model,
		Temperature:    1,
		MaxTokens:      1024,
		TopP:           1,
		ResponseFormat: &ResponseFormat{Type: "json_object"},
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		log.Fatalf("unable to marshal body: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, GORQ_REQ_URL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		log.Fatalf("unable to create request: %s", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("unable to send request: %s", err)
	}

	resBytes, _ := io.ReadAll(res.Body)

	gorqRes := GorqResponse{}
	resMsg := ResponseMessage{}

	json.Unmarshal(resBytes, &gorqRes)

	if res.StatusCode != http.StatusOK {
		json.Unmarshal([]byte(gorqRes.Error.FailedGenerated), &resMsg)
		log.Println(gorqRes.Error.Message)
		fmt.Println(gorqRes.Error.FailedGenerated)
		fmt.Print(resMsg)

	} else {
		json.Unmarshal([]byte(gorqRes.Choices[0].Message.Content), &resMsg)
	}

	return resMsg, nil
}

func visualizeData(data ResponseMessage) error {
	nodes := make(map[string]string)
	for _, d := range data.Data {
		nodes[d.Node1] = d.Node1
		nodes[d.Node2] = d.Node2
	}

	graphConf := struct {
		Nodes map[string]string
		Edges []Edge
	}{Nodes: nodes, Edges: data.Data}

	tmpl := template.New("graph")
	tmpl.Parse(`digraph G {
	layout=circo;

	{{ range $node := .Nodes }}
	"{{ $node }}" [label="{{ $node }}"]
	{{- end }}

	{{ range $edge := .Edges }}
	"{{ $edge.Node1 }}" -> "{{ $edge.Node2 }}" [label="{{ $edge.Edge }}"]
	{{- end }}
}
`)
	tmpl.Execute(os.Stdout, graphConf)

	return nil
}
