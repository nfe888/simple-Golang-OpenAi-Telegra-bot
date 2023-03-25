package main

import (
	"context"
	"log"
	"strconv"

	// "strconv"
	"bytes"
	"fmt"
	"os"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/go-redis/redis/v8"
	"time"
	"github.com/joho/godotenv"
	"strings"
)

const (
	openAIAPI = "https://api.openai.com/v1/chat/completions"
)

var ctx = context.Background()

var openAiToken, telegramBotToken string

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			UserName  string `json:"username"`
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
	ID      string `json:"id"`
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

var redisClient *redis.Client

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	var update Update
	client := &http.Client{}
	json.Unmarshal(body, &update)

	if update.Message.Text == "" {
		return
	}
	text := ""
	if strings.ToUpper(update.Message.Text) == "/START" {
		text = "Welcome, You can start chatting with Haj Jipit."
	} else {

		var messages []Message
		list := redisClient.LRange(ctx, "userHistory:"+strconv.Itoa(update.Message.Chat.ID), 0, -1)
		fmt.Printf("olds: %+v\n", list.Val())
		for _, value := range list.Val() {
			var oldMessage Message
			json.Unmarshal( []byte(value), &oldMessage)
			messages = append(messages, oldMessage)
		}
		var newMessage = Message{Content: update.Message.Text, Role: "user"}
		messages = append(messages, newMessage)
		fmt.Printf("messages: %+v\n", messages)

		openAIReqBody := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": messages,
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
		text = openAIResponse.Choices[0].Message.Content
		messageJson, _ := json.Marshal(newMessage)
		replyJson, _ := json.Marshal(openAIResponse.Choices[0].Message)
		redisClient.RPush(ctx, "userHistory:"+strconv.Itoa(update.Message.Chat.ID), messageJson, replyJson)

	}
	sendMessageReqBody := map[string]interface{}{
		"chat_id": update.Message.Chat.ID,
		"text":    text,
		""
	}
	sendMessageReqBodyJSON, _ := json.Marshal(sendMessageReqBody)
	sendMessageReq, _ := http.NewRequest("POST", "https://api.telegram.org/bot"+telegramBotToken+"/sendMessage", bytes.NewBuffer(sendMessageReqBodyJSON))
	sendMessageReq.Header.Set("Content-Type", "application/json")
	sendMessageResp, _ := client.Do(sendMessageReq)
	defer sendMessageResp.Body.Close()

}

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Some error occured on reading . env file. Err: %s", err)
	}

	openAiToken = os.Getenv("openAiToken")
	telegramBotToken = os.Getenv("telegramBotToken")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: os.Getenv("redisPassword"),
		DB:       0,
	})

	for {

		if _, err := redisClient.Ping(ctx).Result(); err == nil {
			break
			// log.Fatal(err)
		}
		time.Sleep(2 * time.Second)

	}
	log.Println("after redis ")

	router := mux.NewRouter()
	router.HandleFunc("/webhook", handleWebhook).Methods("POST")

	http.ListenAndServe(":8000", router)
}
