package sendspace

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type SendspaceSessionToken struct {
	Result string `xml:"result"`
	Token  string `xml:"token"`
}

func RetrieveSendspaceToken() (SendspaceSessionToken, error) {
	sendspaceKey := os.Getenv("SENDSPACE_KEY")
	log.Println("Sendspace API key env: " + sendspaceKey)
	resp, err := http.Get("http://api.sendspace.com/rest/?method=auth.createtoken&api_key=" + sendspaceKey + "&api_version=1.0&response_format=xml&app_version=0.1")
	body, err := ioutil.ReadAll(resp.Body)
	log.Println(string(body))
	var token SendspaceSessionToken
	err = xml.Unmarshal(body, &token)
	return token, err
}
