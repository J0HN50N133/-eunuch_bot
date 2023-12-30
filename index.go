package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"text/template"
	"time"

	// "github.com/tencentyun/scf-go-lib/cloudfunction"
	"github.com/google/go-github/v57/github"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
	"github.com/tencentyun/scf-go-lib/events"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type DefineEvent struct {
	Key1 string `json:"key1"`
	Key2 string `json:"key2"`
}

var (
	tgBotToken        string
	ghToken           string
	myChatID          int64
	bot               *tgbotapi.BotAPI
	microBlogTemplate *template.Template
)

func init() {
	tgBotToken = os.Getenv("TG_TOKEN")
	ghToken = os.Getenv("GH_TOKEN")
	chatIDStr := os.Getenv("CHATID")
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	myChatID = chatID

	var errBot error
	bot, errBot = tgbotapi.NewBotAPI(tgBotToken)
	if errBot != nil {
		log.Fatal(errBot)
	}
	s, err := os.ReadFile("./microblog_template")
	if err != nil {
		log.Fatal(err)
	}
	microBlogTemplate, err = template.New("microBlogTemplate").Parse(string(s))
	if err != nil {
		log.Fatal(err)
	}
}

type MicroBlog struct {
	Content   string
	CreatedOn time.Time
}

func renderMicroblogTemplate(m MicroBlog) (string, error) {
	var tpl bytes.Buffer
	if err := microBlogTemplate.Execute(&tpl, m); err != nil {
		return "", err
	}
	return tpl.String(), nil
}
func uploadToGithub(m MicroBlog) error {
	microBlogContent, err := renderMicroblogTemplate(m)
	if err != nil {
		return err
	}
	fileName := fmt.Sprintf("%s.md", m.CreatedOn.Format("2006-01-02-15-04-05"))
	ghClient := github.NewClient(nil).WithAuthToken(ghToken)
	commitMsg := "Add microblog " + fileName
	_, _, err = ghClient.Repositories.CreateFile(context.Background(), "J0HN50N133", "J0HN50N133.github.io", "_microblogs/"+fileName, &github.RepositoryContentFileOptions{
		Message: &commitMsg,
		Content: []byte(microBlogContent),
	})
	return err
}

func handleUpdate(update tgbotapi.Update) error {
	chatID := update.Message.Chat.ID
	createdTime := time.Now()
	if chatID != myChatID {
		msg := tgbotapi.NewMessage(chatID, "Sorry, you are not my master.")
		_, err := bot.Send(msg)
		return err
	}
	m := MicroBlog{
		Content:   update.Message.Text,
		CreatedOn: createdTime,
	}
	err := uploadToGithub(m)
	if err != nil {
		msg := tgbotapi.NewMessage(myChatID, "Fail: "+err.Error())
		bot.Send(msg)
		return err
	}
	msg := tgbotapi.NewMessage(myChatID, "Success")
	_, err = bot.Send(msg)
	return err
}
func mainHandler(ctx context.Context, event events.APIGatewayRequest) (string, error) {
	eventBody := event.Body
	var update tgbotapi.Update
	fmt.Println(eventBody)
	if err := json.Unmarshal([]byte(eventBody), &update); err != nil {
		fmt.Println(err)
	}
	err := handleUpdate(update)
	if err != nil {
		return "Error", err
	}
	return "Success", nil
}

func main() {
	cloudfunction.Start(mainHandler)
	// template_test()
}

// package main

// import (
// 	"context"
// 	"fmt"

// 	"github.com/tencentyun/scf-go-lib/cloudfunction"
// )

// type DefineEvent struct {
// 	// test event define
// 	Key1 string `json:"key1"`
// 	Key2 string `json:"key2"`
// }

// func hello(ctx context.Context, event DefineEvent) (string, error) {
// 	fmt.Println("key1:", event.Key1)
// 	fmt.Println("key2:", event.Key2)
// 	return fmt.Sprintf("Hello %s!", event.Key1), nil
// }

// func main() {
// 	// Make the handler available for Remote Procedure Call by Cloud Function
// 	cloudfunction.Start(hello)
// }
