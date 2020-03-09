package mediafire

type MediafireSessionToken struct {
	Action     string `xml:"action"`
	Token      string `xml:"session_token"`
	Secret_key string `xml:"secret_key"`
	Time       string `xml:"time"`
	Result     string `xml:"result"`
}
