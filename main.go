package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

var reqChan chan map[string]string

func main() {
	reqChan = make(chan map[string]string, 10)

	intrptChan := make(chan os.Signal)
	signal.Notify(intrptChan, os.Interrupt)
	go func() {
		<-intrptChan
		close(reqChan)
		os.Exit(0)
	}()

	http.HandleFunc("/", handler)
	http.ListenAndServe("localhost:3000", nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	reqMap := map[string]string{}
	err = json.Unmarshal(reqBody, &reqMap)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	go worker(reqChan)

	reqChan <- reqMap

	w.Write([]byte("Success"))
}

type vt struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type reqStr struct {
	Event            string        `json:"event"`
	Event_type       string        `json:"event_type"`
	App_id           string        `json:"app_id"`
	User_id          string        `json:"user_id"`
	Message_id       string        `json:"message_id"`
	Page_title       string        `json:"page_title"`
	Page_url         string        `json:"page_url"`
	Browser_language string        `json:"browser_language"`
	Screen_size      string        `json:"screen_size"`
	Attributes       map[string]vt `json:"attributes"`
	Traits           map[string]vt `json:"traits"`
}

func worker(reqChan <-chan map[string]string) {
	reqMap := <-reqChan

	attributesMap := make(map[string]vt)
	for i := 1; ; i++ {
		if atrKey, ok := reqMap["atrk"+strconv.Itoa(i)]; ok {
			attributesMap[atrKey] = vt{
				Value: reqMap["atrv"+strconv.Itoa(i)],
				Type:  reqMap["atrt"+strconv.Itoa(i)],
			}
		} else {
			break
		}
	}

	traitsMap := make(map[string]vt)
	for i := 1; ; i++ {
		if traitKey, ok := reqMap["uatrk"+strconv.Itoa(i)]; ok {
			traitsMap[traitKey] = vt{
				Value: reqMap["uatrv"+strconv.Itoa(i)],
				Type:  reqMap["uatrt"+strconv.Itoa(i)],
			}
		} else {
			break
		}
	}

	reqStr := reqStr{
		Event:            reqMap["ev"],
		Event_type:       reqMap["et"],
		App_id:           reqMap["id"],
		User_id:          reqMap["uid"],
		Message_id:       reqMap["mid"],
		Page_title:       reqMap["t"],
		Page_url:         reqMap["p"],
		Browser_language: reqMap["l"],
		Screen_size:      reqMap["sc"],
		Attributes:       attributesMap,
		Traits:           traitsMap,
	}

	jsonReq, _ := json.Marshal(reqStr)
	fmt.Println(string(jsonReq))

	bodyReader := strings.NewReader(string(jsonReq))
	req, _ := http.NewRequest("POST", "https://webhook.site/819ecea9-f5f8-42c4-9641-575e20866467", bodyReader)
	http.DefaultClient.Do(req)
}
