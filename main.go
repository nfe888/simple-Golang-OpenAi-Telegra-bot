package main

import (
    "strconv"
    "bytes"
    "fmt"
    "os"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "github.com/gorilla/mux"
)

const (
    openAIAPI = "https://api.openai.com/v1/completions"
)

var  openAiToken,telegramBotToken string

type Update struct {
    UpdateID int `json:"update_id"`
    Message  struct {
        MessageID int `json:"message_id"`
        From      struct {
            ID           int    `json:"id"`
            FirstName    string `json:"first_name"`
            LastName     string `json:"last_name"`
            UserName     string `json:"username"`
        } `json:"from"`
        Chat struct {
            ID    int    `json:"id"`
            Type  string `json:"type"`
            Title string `json:"title"`
        } `json:"chat"`
        Text string `json:"text"`
    } `json:"message"`
}

type OpenAiResponse struct {
    ID       string `json:"id"`
    Choices  []struct {
        Text string `json:"text"`
    } `json:"choices"`
}
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    body, _ := ioutil.ReadAll(r.Body)

    var update Update
    client := &http.Client{}
    json.Unmarshal(body, &update)

    if update.Message.Text == "" {
        return
    }
    text :=""
    if update.Message.Text == "/START"{
	text = "Welcome, You can start chatting with Haj Jipit."
    } else{


    openAIReqBody := map[string]interface{}{
        "prompt": update.Message.Text,
        "model": "text-davinci-003",
        "temperature": 0.9,
        "max_tokens": 200,
        "user" : strconv.Itoa(update.Message.From.ID),
    }
    openAIReqBodyJSON, _ := json.Marshal(openAIReqBody)
    openAIReq, _ := http.NewRequest("POST", openAIAPI, bytes.NewBuffer(openAIReqBodyJSON))
    openAIReq.Header.Set("Content-Type", "application/json")
    openAIReq.Header.Set("Authorization", "Bearer "+openAiToken)
    
    openAIResp, _ := client.Do(openAIReq)
    openAIRespBody, _ := ioutil.ReadAll(openAIResp.Body)
    defer openAIResp.Body.Close()

    var openAIResponse OpenAiResponse
    fmt.Println(string(openAIRespBody))
    json.Unmarshal(openAIRespBody, &openAIResponse)
    text = openAIResponse.Choices[0].Text
}
    sendMessageReqBody := map[string]interface{}{
        "chat_id": update.Message.Chat.ID,
        "text": text,
    }
    sendMessageReqBodyJSON, _ := json.Marshal(sendMessageReqBody)
 sendMessageReq, _ := http.NewRequest("POST", "https://api.telegram.org/bot"+telegramBotToken+"/sendMessage", bytes.NewBuffer(sendMessageReqBodyJSON))
    sendMessageReq.Header.Set("Content-Type", "application/json")
    sendMessageResp, _ := client.Do(sendMessageReq)
    defer sendMessageResp.Body.Close()


}

func main() {
    openAiToken=os.Getenv("openAiToken")
    telegramBotToken=os.Getenv("telegramBotToken")
    router := mux.NewRouter()
    router.HandleFunc("/webhook", handleWebhook).Methods("POST")

    http.ListenAndServe(":8000", router)
}

