//***************************************
// Author: Samolyk Alexey
// E-mail: root@sysalex.com
// Twitter/Skype and other: @POS_troi
//***************************************

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Syfaro/telegram-bot-api"
	"github.com/julienschmidt/httprouter"
	_ "github.com/mattn/go-sqlite3"
)

type (
	// Структура Gitlab JSON
	GitLabHook struct {
		ObjectKind       string           `json:"object_kind"`
		Ref              string           `json:"ref"`
		UserName         string           `json:"user_name"`
		Repository       Repository       `json:"repository"`
		Commit           []Commit         `json:"commits"`
		TotalCommits     int              `json:"total_commits_count"`
		User             User             `json:"user"`
		ObjectAttributes ObjectAttributes `json:"object_attributes"`
	}
	Repository struct {
		Homepage string `json:"homepage"`
		Name     string `json:"name"`
	}
	Commit struct {
		Id      string `json:"id"`
		Message string `json:"message"`
		Url     string `json:"url"`
		Author  Author `json:"author"`
	}
	Author struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	User struct {
		Name string `json:"name"`
	}
	ObjectAttributes struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		State       string `json:"state"`
		URL         string `json:"url"`
		Action      string `json:"action"`
	}
	// -------------
	HookMessage struct {
		Text string
	}

	// config
	Config struct {
		Api      string `json:"bot_api"`
		HookKey  string `json:"hook_key"`
		Admin    string `json:"bot_admin"`
		Listen   string `json:"listen"`
		DataBase string `json:"database"`
	}
)

var (
	webhookResponse chan HookMessage
	RoomID          int64
	RoomTitle       string
	Command         string
	CommandArg      string
	UserName        string
	FromUserID      int
	Message         tgbotapi.Chattable
	RepositoryName  string
	BotConfig       Config
	RoomList        map[string][]string
	DB              *sql.DB
)

func main() {
	config := os.Args[1]
	if config != "" {
		BotConfig = Config{}.loadConfig(config) // Load config file
	}else{
		BotConfig = Config{}.loadConfig("./bot.cfg") // Load config file
	}

	DB = initDB(BotConfig.DataBase)              // init data base
	RoomList = getRepositoryList(DB)             // Get added repositories and room for push event message
	// Hook chanel
	webhookResponse = make(chan HookMessage, 5)
	// Connect to Telegram
	tgbot, err := tgbotapi.NewBotAPI(BotConfig.Api)
	checkErr(err)

	ucfg := tgbotapi.NewUpdate(0)
	ucfg.Timeout = 60
	updates, err := tgbot.GetUpdatesChan(ucfg)
	// Run WEB
	go WebHook()
	// Magick
	for {
		select {
		// Get chanel update
		case update := <-updates:

			// bot command
			Command = update.Message.Command()
			// bot command arg
			CommandArg = update.Message.CommandArguments()
			// room ID
			RoomID = update.Message.Chat.ID
			// room title
			RoomTitle = update.Message.Chat.Title
			// Message Username
			UserName = update.Message.From.UserName
			// Message user ID
			FromUserID = update.Message.From.ID

			// Wait command "/start_hook ARGS", ARGS - GitLab repository name
			// example /start_hook gitlab-ce
			if Command == "start_hook" && UserName == BotConfig.Admin {
				msg, list := addNewRepositoryToRoom(DB, CommandArg, RoomTitle, RoomID)
				RoomList = list
				Message := tgbotapi.NewMessage(RoomID, msg)
				Message.DisableWebPagePreview = true
				Message.ParseMode = tgbotapi.ModeHTML
				tgbot.Send(Message)
			}
			// WebHook
		case hook := <-webhookResponse:
			for _, id := range RoomList[RepositoryName] { // get room id for repository
				roomID, _ := strconv.Atoi(id)
				Message := tgbotapi.NewMessage(int64(roomID), hook.Text)
				Message.DisableWebPagePreview = true
				Message.ParseMode = tgbotapi.ModeHTML
				tgbot.Send(Message)
			}
		}
	}

}

// WEB hook
func WebHook() {
	r := httprouter.New()
	// location "/"
	r.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprint(w, "Sorry!\n")
	})
	// location "/hook/$hook_key" - $hook_key - from bot.cfg
	// Get POST request from GitLab
	hookURL := fmt.Sprintf("/hook/%s", BotConfig.HookKey)
	r.POST(hookURL, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		u := GitLabHook{}
		json.NewDecoder(r.Body).Decode(&u)
		webhookResponse <- MakeMessage(u)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(201)
	})

	http.ListenAndServe(BotConfig.Listen, r)

}

// Make Telegram message for PUSH/TAG/ISSUE events
func MakeMessage(hook GitLabHook) HookMessage {
	var message string
	switch {
	case hook.ObjectKind == "push":
		message = commitMesage(hook)
	case hook.ObjectKind == "tag_push":
		message = tagPushMesage(hook)
	case hook.ObjectKind == "issue":
		message = issueMesage(hook)
	}

	RepositoryName = hook.Repository.Name
	return HookMessage{message}
}

// Make Commit message text
func commitMesage(hook GitLabHook) (message string) {
	message = fmt.Sprintf("<b>New commits</b>\nProject: <a href=\"%s\">%s</a>\nUser: %s\n\n",
		hook.Repository.Homepage, hook.Repository.Name, hook.UserName)
	for _, commit := range hook.Commit {
		message += fmt.Sprintf("> <a href=\"%s\">%s</a> \n", commit.Url, strings.TrimRight(commit.Message, "\n"))
	}
	message += "\nTotal commits: " + strconv.Itoa(hook.TotalCommits) + "\n"
	return
}

// Make TAG message text
func tagPushMesage(hook GitLabHook) (message string) {
	tagName := strings.Replace(hook.Ref, "refs/tags/", "", -1)
	message = fmt.Sprintf("<b>New tag</b>\nProject: <a href=\"%s\">%s</a>\nUser: %s\nTag: <a href=\"%s/tags/%s\">%s</a>",
		hook.Repository.Homepage, hook.Repository.Name, hook.UserName, hook.Repository.Homepage, tagName, tagName)
	return
}

// Make Issue message text
func issueMesage(hook GitLabHook) (message string) {
	var issueState string
	switch {
	case hook.ObjectAttributes.Action == "open":
		issueState = "Open new"
	case hook.ObjectAttributes.Action == "close":
		issueState = "Close"
	case hook.ObjectAttributes.Action == "reopen":
		issueState = "Reopen"
	case hook.ObjectAttributes.Action == "update":
		issueState = "Update"

	}
	message = fmt.Sprintf("<b>%s Issue</b>\nProject: <a href=\"%s\">%s</a>\nUser: %s\n<a href=\"%s\">%s</a>\n",
		issueState, hook.Repository.Homepage, hook.Repository.Name, hook.User.Name, hook.ObjectAttributes.URL, hook.ObjectAttributes.Title)

	return
}

// Load config file
func (config Config) loadConfig(confFile string) Config {
	data, err := ioutil.ReadFile(confFile)
	checkErr(err)

	json.Unmarshal(data, &config)
	return config
}

///////////////////////////////////////////////////
// SQL
//////////////////////////////////////////////////

// Connect to SQLite3 data base
func initDB(dbPath string) *sql.DB {
	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		log.Fatal(err)
	}
	// Data base file no exist? Create
	if _, err := os.Stat(BotConfig.DataBase); err != nil {
		sqlStmt := `
		create table repository (id integer not null primary key AUTOINCREMENT, repository_name sting, repository_id integer);
		create table room (id integer not null primary key AUTOINCREMENT, room_title string, room_id integer, repository_id integer);
		`
		_, err = db.Exec(sqlStmt)
		checkErr(err)

	}
	return db
}

// Get repository list
// In type sql.DB connection
//               Repo name    Room ID's
// out map, map[Repository1:[1111 2222] Repository2:[1111 2222]]
func getRepositoryList(db *sql.DB) map[string][]string {
	repository := make(map[string][]string)
	rows, err := db.Query("select repository.repository_name, group_concat(room.room_id) as list from repository join room on room.repository_id = repository.id GROUP BY repository.repository_name ")
	checkErr(err)
	defer rows.Close()
	for rows.Next() {
		var repository_name string
		var list string
		err = rows.Scan(&repository_name, &list)
		checkErr(err)
		repository[repository_name] = strings.Split(list, ",")
	}
	err = rows.Err()
	checkErr(err)
	return repository
}

// Add new repository to room ( telegram chat)
// In type sql.DB connection,
//    repositoryName - equal to name in gitlab
//    roomID - ID Telegram room (chat)
// out error

func addNewRepositoryToRoom(db *sql.DB, repositoryName string, roomTitle string, roomID int64) (message string, list map[string][]string) {

	id, err := checkRepository(db, repositoryName)
	if err != nil {
		repoid, err := addRepository(db, repositoryName)
		checkErr(err)
		err = addRoom(db, roomTitle, roomID, repoid)
		checkErr(err)
		message = fmt.Sprintf("<b>Repository: %s</b>\nThe repository has been added to the current room.\n<b>Room title:</b> %s\n<b>Room ID:</b> %d\n", repositoryName, roomTitle, roomID)
	} else {
		_, err := checkRoom(db, id, roomID)
		if err != nil {
			err = addRoom(db, roomTitle, roomID, id)
			checkErr(err)
			message = fmt.Sprintf("The repository %s has been added to the current room (ID %d)", repositoryName, roomID)
		} else {
			message = fmt.Sprintf("Repository %s is already in this room", repositoryName)
		}

	}
	list = getRepositoryList(db)
	return
}

func checkRepository(db *sql.DB, repositoryName string) (id int64, err error) {
	sqlQuery := `
		select id from repository where repository_name = ?
	`
	query, err := db.Prepare(sqlQuery)
	defer query.Close()
	err = query.QueryRow(repositoryName).Scan(&id)

	return
}

func checkRoom(db *sql.DB, repositoryId int64, roomID int64) (id int64, err error) {
	sqlQuery := `
		select id from room where repository_id = ? and room_id = ?
	`
	query, err := db.Prepare(sqlQuery)
	defer query.Close()
	err = query.QueryRow(repositoryId, roomID).Scan(&id)
	return
}

func addRepository(db *sql.DB, repositoryName string) (lastID int64, err error) {
	sqlQuery := `
		insert into repository(repository_name) values(?)
	`
	repository, err := db.Exec(sqlQuery, repositoryName)
	lastID, err = repository.LastInsertId()
	return
}

func addRoom(db *sql.DB, roomTitle string, roomID int64, repositoryID int64) (err error) {
	sqlQuery := `
		insert into room(room_title,room_id, repository_id) values(?,?,?)
	`
	_, err = db.Exec(sqlQuery, roomTitle, roomID, repositoryID)
	return
}

///////////////////////////////////////////////////
// Util
//////////////////////////////////////////////////
// Error check
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
