## Telegramm BOT for GitLab v0.1-alpha

### WARNING
This is first **Alpha** version, all in one

### Capabilities
**Push events**  
This events will be triggered by a push to the repository  

**Tag push events**  
This events will be triggered when a new tag is pushed to the repository  

**Issues events**  
This events will be triggered when an issue is created/updated/close/reopen  

### Build

```sh
git clone https://gitlab.com/POStroi/gitlab-telegram-bot.git
cd gitlab-telegram-bot
go get github.com/Syfaro/telegram-bot-api github.com/julienschmidt/httprouter github.com/mattn/go-sqlite3
go build -ldflags "-s -w" .
```
### Run
```sh
./gitlab-telegram-bot &
```

### Settings
Default config path `./bot.cfg`, JSON format
```json
{
"bot_api": "BOT_API_KEY",
"hook_key": "RANDOM_KEY",
"bot_admin": "Privileged_USER_for_BOT",
"listen": "0.0.0.0:3000",
"database": "./chat.db"
}
```
**bot_api** - Api key obtained from [@BotFather][1]  
**hook_key** - Auth. key - random letter diget combination  
**bot_admin** - telegram user, admin for bot. Nikname without "@"  
**listen** - IP and port  
**database** - SQLite3 database path

### Settings GitLab WEB hook
`URL` = `http://IP:PORT/hook/hook_key`

### Enabling event publishing to a telegram channel

1. Add your  bot to common channel or open a private channel  
2. Use the /start_hook `ARGS` command to start a publicate events to channel,  
`ARGS` should be equivalent to the repository name.  
3. Eat a cookie and enjoy spam in channel :D

### Docker

```sh
docker pull sysalex/gitlab-telegram-bot

docker run -t -i -p 3000:8081 -d \
--env BOT_API_KEY="BOT_API_KEY" \
--env BOT_HOOK_KEY="BOT_HOOK_KEY" \
--env BOT_ADMIN="ADMIN_USER_NAME" sysalex/gitlab-telegram-bot
```
**hook url**
`URL` = `http://IP:8081/hook/BOT_HOOK_KEY`

## Contributing
:-*

## License
Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

[1]: https://core.telegram.org/bots#6-botfather
