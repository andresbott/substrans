package llmtranslate

import (
	"bytes"
	"context"
	"github.com/tmc/langchaingo/llms"
	"text/template"

	"github.com/tmc/langchaingo/llms/ollama"
)

// TranslateLine represents the payload for translation

// Translator is responsible for connecting to Ollama and translating text
type Translator struct {
	client *ollama.LLM
}

const ModelLlama3 = "llama3"
const ModelLlama32 = "llama3.2"
const ModelPhi35mini = "phi3.5"
const ModelPhi4 = "phi4:14b"
const ModelGemma3 = "gemma3:12b"
const defaultUrl = "http://127.0.0.1:11434"

// NewTranslator creates a new Translator instance
func NewTranslator(model, url string) (*Translator, error) {

	if url == "" {
		url = defaultUrl
	}

	llm, err := ollama.New(ollama.WithModel(model), ollama.WithServerURL(url))
	if err != nil {
		return nil, err
	}
	t := &Translator{
		client: llm,
	}
	return t, nil
}

// chatMsg represents a message with context and translation payload
type chatMsg struct {
	PrevContext []string
	PostContext []string
	Line        string
	Lang        string
}

var tmpl = `
Given the context as follows:
{{range .PrevContext}}- {{.}}
{{end}}
{{.Line}}
{{range .PostContext}}- {{.}}
{{end}}

translate the line:
'{{.Line}}' 
into {{.Lang}}

Please make sure to only say the translated line, no babling or explanation, don't print the context.
`

// FormatMessage formats the message for translation using the Go template engine
func (c *chatMsg) FormatMessage() (string, error) {
	msg, err := template.New("message").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = msg.Execute(&buf, c)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

const LangEs = "spanish from spain"

const systemPromt = `You are a professional translator with deep knowledge of different languages and phrasing.
your task is to translate movie subtitles from one language to another trying to keep the sentiment of the conversation and the tone as close as possible to the original. 
You will be given a context and a line to translate in that context`

// Translate translates the given text to the specified language
func (t *Translator) Translate(ctx context.Context, prevContext, postContext []string, translateLine, lang string) (string, error) {

	msg := chatMsg{
		PrevContext: prevContext,
		PostContext: postContext,
		Line:        translateLine,
		Lang:        lang,
	}
	parsedMsg, err := msg.FormatMessage()
	if err != nil {
		return "", err
	}

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPromt),
		llms.TextParts(llms.ChatMessageTypeHuman, parsedMsg),
	}
	resp, err := t.client.GenerateContent(ctx, content, llms.WithTemperature(0.3))
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Content, nil
}
