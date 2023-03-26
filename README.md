## Simple Golang Chat-GPT (OpenAI) Telegram Bot

#### Step 1:

Create a telegram bot via https://t.me/BotFather and get it's api token

#### Step 2:

Register for an open ai account and grab the api token

#### Step 3:
create your .env file `cp .env-example .env`  then complete it by inserting your tokens.


#### Step 4:
Telegram wants an SSL endpoint to deliver bot requests too your server. 

You have some options here. 

- Services like CloudFlare and its proxy option with Flexible ssl mode. (Recommended)
- Use letsencrypt service for free ssl (needs renew every 3 months)
- Self Signed ssl on your server and sending the certificate it via setWebhook api call. (needs code change)

We used CloudFlare.

And here is a simple nginx configuration for the subdomain pointing to the server. change `8000` to value of `golangPort` on `.env` file:
```
server {
    server_name yoursub.domain.com;
    listen 80;
    listen [::]:80;
    server_tokens off;
    location / {
        proxy_pass http://localhost:8000/webhook;
    }
}
```


#### Step 5:
run `docker-compose up -d` the bot is ready!
