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
	flag.Parse()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Unable to load environments: %s", err)
	}

	text, err := pdfToText(*inputPath)
	if err != nil {
		log.Fatalf("unable to convert pdf file to text: %s", err)
	}

	apiKey := os.Getenv("API_KEY")
	chsize := 4000
	ntext := len(text)
	nchunks := (ntext / chsize) + 1

	data := ResponseMessage{}

	for i := 0; i < nchunks; i++ {
		end := (i + 1) * chsize
		if end > ntext {
			end = ntext
		}
		prompt := PROMPT + text[i*chsize:end]

		d, err := GorqSendRequest(prompt, apiKey, *llmModel)
		if err != nil {
			log.Println(err)
			i--
			continue
		}

		data.Data = append(data.Data, d.Data...)
	}

	err = visualizeData(data)
	if err != nil {
		log.Fatalf("unable to visualize data: %s", err)
	}
}

func pdfToText(pdfPath string) (string, error) {
	cmd := exec.Command("pdftotext", pdfPath, "-")
	out, err := cmd.Output()

	return string(out), err
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

	if res.StatusCode == http.StatusOK {
		json.Unmarshal([]byte(gorqRes.Choices[0].Message.Content), &resMsg)
	} else {
		return resMsg, fmt.Errorf("unable to get response from this chunk")
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
	layout=sfdp;
	overlap=scale;

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
