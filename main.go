package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

type Payload struct {
	Topic    string `json:"topic"`
	Type     string `json:"type"`
	Duration string `json:"duration"`
	TimeZone string `json:"timezone"`
	Password string `json:"password"`
	Agenda   string `json:"agenda"`
}

type ZoomApi struct {
	JoinUrl string `json:"join_url"`
}
type SlackPayload struct {
	UserName  string `json:"username"`
	IconEmoji string `json:"icon_emoji"`
	Channel   string `json:"channel"`
	Text      string `json:"text"`
}

func fetch_daily_person(gas_url string) {
	req, _ := http.NewRequest("GET", gas_url, nil)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{ServerName: "script.google.com"},
	}
	client := &http.Client{
		Transport: tr,
	}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	/* 	encode_res := bufio.NewScanner(transform.NewReader(res.Body, japanese.ShiftJIS.NewEncoder()))
	 */
	contents, err := ioutil.ReadAll(res.Body)
	fmt.Println(string(contents))
}

func notify_slack(join_url string) {
	path := os.Getenv("WEBHOOK_URL")

	payload := SlackPayload{
		UserName:  "デイリーお知らせbot",
		IconEmoji: ":spiral_calendar_pad",
		Channel:   "#会議担当者の通知",
		Text:      "aaa",
	}
	json_payload, _ := json.Marshal(payload)

	/* res, err := http.PostForm(path, payload) */
	req, err := http.NewRequest("POST", path, bytes.NewReader(json_payload))
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.StatusCode)
}

func main() {
	loadEnv()
	payload := Payload{
		Topic:    "デイリー",
		Type:     "1",
		Duration: "40",
		TimeZone: "Asia/Tokyo",
		Password: "",
		Agenda:   "進捗報告",
	}
	payload_json, _ := json.Marshal(payload)
	path := "https://api.zoom.us/v2/users/" + os.Getenv("USER_ID") + "/meetings"

	connect, _ := http.NewRequest("POST", path, bytes.NewBuffer(payload_json))

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{ServerName: "api.zoom.us"},
	}
	client := &http.Client{
		Transport: tr,
	}

	payloadForJwt := jwt.MapClaims{
		"iss": os.Getenv("API_KEY"),
		"exp": time.Now().Add(36000).UnixNano(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payloadForJwt)
	tokenString, err := token.SignedString([]byte(os.Getenv("API_SECRET")))

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("Authorization", "Bearer"+tokenString)
	connect.Header = header
	req, err := client.Do(connect)
	if err != nil {
		log.Fatal(err)
	}
	contents, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	join_url := []byte(string(contents))
	var z ZoomApi
	json.Unmarshal(join_url, &z)
	/* fmt.Println(z.JoinUrl) */
	fetch_daily_person(os.Getenv("SPREAD_SHEET_URL"))
	notify_slack(string(join_url))
}

func loadEnv() {

	// ここで.envファイル全体を読み込みます。
	// この読み込み処理がないと、個々の環境変数が取得出来ません。
	// 読み込めなかったら err にエラーが入ります。
	err := godotenv.Load(".env")

	//もし err がnilではないなら、"読み込み出来ませんでした"が出力されます。
	if err != nil {
		fmt.Printf("読み込み出来ませんでした: %v", err)
	}

	//.envの SAMPLE_MESSAGEを取得して、messageに代入します。
	message := os.Getenv("SAMPLE_MESSAGE")

	fmt.Println(message)
}
