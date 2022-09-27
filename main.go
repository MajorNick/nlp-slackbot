package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/krognol/go-wolfram"
	"github.com/shomali11/slacker"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go/v2"
)

func printCommandEvents(analyticsChan <-chan *slacker.CommandEvent) {
	for event := range analyticsChan {
		fmt.Println("Command Event:")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Problem In loading Env : %v", err)
	}

	client := witai.NewClient(os.Getenv("WIT-AI-TOKEN"))

	bot := slacker.NewClient(os.Getenv("SLACK-BOT-TOKEN"), os.Getenv("SLACK-APP-TOKEN"))
	go printCommandEvents(bot.CommandEvents())
	wolframClient := &wolfram.Client{os.Getenv("WOLFRAM-APP-ID")}

	bot.Command("query for bot - <message>", &slacker.CommandDefinition{
		Description: "send any questions to wolfram",
		Examples:    []string{"Who is president of Georgia"},
		Handler: func(botCtx slacker.BotContext, request slacker.Request, response slacker.ResponseWriter) {
			query := request.Param("message")

			msg, _ := client.Parse(&witai.MessageRequest{
				Query: query,
			})
			data, err := json.MarshalIndent(msg, "", "    ")
			if err != nil {
				log.Fatalf("Can't marshal message for wolfram: %v", err)
			}

			tmp := string(data[:])
			value := gjson.Get(tmp, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")

			question := value.String()
			ans, err := wolframClient.GetSpokentAnswerQuery(question, wolfram.Metric, 1000)
			if err != nil {
				log.Fatalf("Error in getting answer: %v", err)
			}
			response.Reply(ans)
		},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = bot.Listen(ctx)

	if err != nil {
		log.Fatalf("Problem In Listening Bot : %v", err)
	}
}
