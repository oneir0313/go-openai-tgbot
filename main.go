package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/joho/godotenv/autoload"
	openai "github.com/sashabaranov/go-openai"
	tele "gopkg.in/telebot.v3"
)

var (
	tgToken string
	openaiToken string
	stream bool
	aiModel string
)

func init() {
	var defaultAiMdoel string
	if aiModelEnv, ok := os.LookupEnv("OPENAPI_MODEL"); ok {
		defaultAiMdoel = aiModelEnv
	} else {
		defaultAiMdoel = openai.GPT3Dot5Turbo
	}
	flag.StringVar(&tgToken, "tg-bot-token", os.Getenv("TGBOT_TOKEN"), "telegram bot token")
	flag.StringVar(&openaiToken, "openai-token", os.Getenv("OPENAPI_TOKEN"), "chat gpt token")
	flag.StringVar(&aiModel, "openai-mdoel", defaultAiMdoel, "chat gpt model")
	flag.BoolVar(&stream, "stream", false, "use streaming")
}


func main() {
	flag.Parse()
	pref := tele.Settings{
		Token:  tgToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	client := openai.NewClient(openaiToken)


	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	if stream {
		fmt.Println("Handle stream")
		b.Handle(tele.OnText, func(c tele.Context) error {
			// All the text messages that weren't
			// captured by existing handlers.
			ctx := context.Background()
			ch := make(chan string)
			done := make(chan struct{}, 1)
			wg := &sync.WaitGroup{}
			wg.Add(2)
			var (
				user = c.Sender()
				text = c.Text()
			)
			go func () {
				defer wg.Done()
	
				req := openai.ChatCompletionRequest{
					Model:     aiModel,
					MaxTokens: 20,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleUser,
							Content: text,
						},
					},
					Stream: true,
				}
				stream, err := client.CreateChatCompletionStream(ctx, req)
				if err != nil {
					close(ch)
					fmt.Printf("CompletionStream error: %v\n", err)
					return
				}
				defer stream.Close()
			
				for {
					response, err := stream.Recv()
					if errors.Is(err, io.EOF) {
						done <- struct{}{}
						fmt.Println("Stream finished")
						return
					}
			
					if err != nil {
						fmt.Printf("Stream error: %v\n", err)
						return
					}
					
					ch <- response.Choices[0].Delta.Content
					fmt.Printf("Stream response: %v\n", response)
				}
			}()
	
			go func() {
				defer wg.Done()
				var lastResp string
				var sent *tele.Message
				send := func(msg string) error {
					defer func() {
						lastResp = msg
					}()
			
					if len(strings.Trim(msg, "\n")) == 0 {
						time.Sleep(time.Microsecond * 50)
						return nil
					}
			
					if sent == nil {
						var err error
			
						sent, err = b.Send(user, msg)
						return err
					}
			
					if msg == lastResp {
						return nil
					}
					msg = lastResp + msg
					if _, err := b.Edit(sent, msg); err != nil {
						return err
					}
			
					return nil
				}
				
				for {
					select {
					case <- done:
						return
					case msg := <- ch:
						err := send(msg)
						if err != nil {
							fmt.Printf("telegram bot send: %s\n", err.Error())
						}
					}
				}
			}()
		
	
			wg.Wait()
	
			return nil
		})
	} else {
		fmt.Println("Handle without streaming")
		b.Handle(tele.OnText, func(c tele.Context) error {
			var (
				text = c.Text()
			)
			
			resp, err := client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model: aiModel,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleUser,
							Content: text,
						},
					},
				},
			)
		
			if err != nil {
				errResp := fmt.Sprintf("ChatCompletion error: %v\n", err)
				return c.Send(errResp)
			}
			return c.Send(resp.Choices[0].Message.Content)
		})
	}
	fmt.Printf("openai tgbot start, your model is: %s", aiModel)

	b.Start()
}