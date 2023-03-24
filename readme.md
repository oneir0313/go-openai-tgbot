# OpenAI TG BOT

simple, chat to gpt from telegram bot
### Go Run

    $ go run main.go \
    -tg-bot-token={YOUR_TELEGRAM_BOT_TOKEN} \ 
    -openai-token={YOUR_OPENAI_TOKEN} \
    -stream 
    
> `-stream` is optional if you want to run with streaming

### Docker Build

    docker build . -t openai-tgbot

### Docker Run


    docker run --restart=on-failure:5 -d \
        -e OPENAPI_TOKEN={YOUR_OPENAI_TOKEN} \
        -e TGBOT_TOKEN={YOUR_TELEGRAM_BOT_TOKEN} \
        --name openai-tgbot \
        openai-tgbot