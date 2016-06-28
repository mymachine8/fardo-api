package slack

import (
	"encoding/json"
	"net/http"
	"log"
	"bytes"
)

type Level uint8

const (
	// PanicLevel level, highest level of severity. Logs and then calls panic with the
	// message passed to Debug, Info, ...
	PanicLevel Level = iota
	// FatalLevel level. Logs and then calls `os.Exit(1)`. It will exit even if the
	// logging level is set to Panic.
	FatalLevel
	// ErrorLevel level. Logs. Used for errors that should definitely be noted.
	// Commonly used for hooks to send errors to an error tracking service.
	ErrorLevel
	// WarnLevel level. Non-critical entries that deserve eyes.
	WarnLevel
	// InfoLevel level. General operational entries about what's going on inside the
	// application.
	InfoLevel
	// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
	DebugLevel
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Attachment struct {
	Fallback   *string  `json:"fallback"`
	Color      *string  `json:"color"`
	PreText    *string  `json:"pretext"`
	AuthorName *string  `json:"author_name"`
	AuthorLink *string  `json:"author_link"`
	AuthorIcon *string  `json:"author_icon"`
	Title      *string  `json:"title"`
	TitleLink  *string  `json:"title_link"`
	Text       *string  `json:"text"`
	ImageUrl   *string  `json:"image_url"`
	Fields     []*Field `json:"fields"`
}

func (attachment *Attachment) AddField(field Field) *Attachment {
	attachment.Fields = append(attachment.Fields, &field)
	return attachment
}

func getColor(level Level) string {
	var color string;
	switch level {
	case DebugLevel:
		color = "#9B30FF"
	case InfoLevel:
		color = "good"
	case ErrorLevel, FatalLevel, PanicLevel:
		color = "danger"
	default:
		color = "warning"
	}
	return color;
}

func Send(level Level, text string) {
	if (text == "") {
		return;
	}
	payload := make(map[string]interface{})
	payload["parse"] = "full"
	payload["username"] = "golang"
	payload["channel"] = "#bugs"
	payload["iconEmoji"] = ":ghost:"
	payload["color"] = getColor(level);
	payload["text"] = text

	sendMessage("https://hooks.slack.com/services/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", payload)
}

func sendMessage(webhookUrl string, payload map[string]interface{}) {
	data, _ := json.Marshal(payload)
	r := bytes.NewReader(data)
	_, err := http.Post(webhookUrl, "application/json", r);
	if err != nil {
		log.Print(err)
	}
}