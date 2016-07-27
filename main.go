package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"net"
	"net/http"
)

type Payload struct {
	After string `json:"after"`
}

var mchan chan string

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func notifier(mchan <-chan string, Email string, Password string, ChannelID string, Repo string) {
	log.Printf("Logging into: Email: %s ChannelID: %s\n", Email, ChannelID)
	Token := ""

	dg, err := discordgo.New(Email, Password, Token)
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}

	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}
	for {
		msg := <-mchan
		dg.ChannelMessageSend(ChannelID, msg)
	}
}

func Endpoint(w http.ResponseWriter, r *http.Request) {
	var parsed Payload
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&parsed)
	if err != nil {
		log.Println("Failed to parse json.")
	}

	log.Printf("New commit: %s\n", parsed.After)
	mchan <- fmt.Sprintf("New commit: https://%s/commit/", parsed.After)
	fmt.Fprintf(w, "OK.")

}

func main() {
	Email := flag.String("email", "", "email")
	Password := flag.String("pass", "", "password")
	ChannelID := flag.String("channel", "", "channelID")
	Repo := flag.String("repo", "", "Repo url eg. github.com/pielover88888/bowtf")

	flag.Parse()

	mchan = make(chan string)
	go notifier(mchan, *Email, *Password, *ChannelID, *Repo)

	log.Println("Starting endpoint on port :1313")
	fmt.Printf("Add webhook http://%s:1313/endpoint to https://%s/settings/hooks/new\n", GetLocalIP(), *Repo)
	http.HandleFunc("/endpoint", Endpoint)
	http.ListenAndServe(":1313", nil)
}
