package request

type SMSContent struct {
	Number      string `json:"number"`
	Destination string `json:"destination"`
	Text        string `json:"text"`
}
