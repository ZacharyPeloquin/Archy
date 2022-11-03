package levelFaster

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	firebase "firebase.google.com/go"
)

// Payload struct that is expected
type Payload struct {
	ServerId string `json:"server_id"`
	UserId   string `json:"user_id"`
}

// User expected format in firestore
type FirestoreUser struct {
	Username           string `firestore:"username"`
	Level              int    `firestore:"level"`
	AvatarUrl          string `firestore:"avatar_url"`
	ExpTowardNextLevel int    `firestore:"exp_toward_next_level"`
	TotalExp           int    `firestore:"total_exp"`
}

// Return the level image url
func Level(w http.ResponseWriter, r *http.Request) {
	var payload Payload

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		panic(err)
	}

	if len(payload.ServerId) == 0 ||
		len(payload.UserId) == 0 {
		fmt.Fprint(w, "Something went wrong")
		return
	}

	// Instanciate the firestore context
	ctx := context.Background()
	conf := &firebase.Config{ProjectID: "archy-f06ed"}
	app, err := firebase.NewApp(ctx, conf)

	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	userRef, err := client.Collection("servers").Doc(payload.ServerId).Collection("users").Doc(payload.UserId).Get(ctx)
	if err != nil {
		panic(err)
	}
	var user FirestoreUser

	err = userRef.DataTo(&user)
	if err != nil {
		panic(err)
	}

	rank, _ := client.Collection("servers").Doc(payload.ServerId).Collection("users").Where("total_exp", ">", user.TotalExp).Documents(ctx).GetAll()

	re := regexp.MustCompile("avatars/(.*).png")
	var subMatch = re.FindStringSubmatch(user.AvatarUrl)
	if len(subMatch) < 2 {
		re := regexp.MustCompile("avatars/(.*).gif")
		subMatch = re.FindStringSubmatch(user.AvatarUrl)
	}

	var avatarId = subMatch[1]

	var levelExpNeeded float64 = (5 * math.Pow(float64(user.Level), 2)) + (50 * float64(user.Level)) + 100

	var formatedUrl strings.Builder
	formatedUrl.WriteString("https://us-central1-archy-f06ed.cloudfunctions.net/nextjs/api/bar?")

	formatedUrl.WriteString("username=")
	formatedUrl.WriteString(url.QueryEscape(user.Username))

	formatedUrl.WriteString("&rank=")
	formatedUrl.WriteString(url.QueryEscape(strconv.FormatInt(int64(len(rank)+1), 10)))

	formatedUrl.WriteString("&level=")
	formatedUrl.WriteString(url.QueryEscape(strconv.FormatInt(int64(user.Level), 10)))

	formatedUrl.WriteString("&avatar_url=")
	formatedUrl.WriteString(url.QueryEscape(avatarId))

	formatedUrl.WriteString("&exp_toward_next_level=")
	formatedUrl.WriteString(url.QueryEscape(strconv.FormatInt(int64(user.ExpTowardNextLevel), 10)))

	formatedUrl.WriteString("&level_exp_needed=")
	formatedUrl.WriteString(url.QueryEscape(strconv.FormatInt(int64(levelExpNeeded), 10)))

	print(formatedUrl.String())
	fmt.Fprint(w, formatedUrl.String())
}
