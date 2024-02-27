package main

import (
	"context"
	"flag"
	"fmt"
	caldav "mycaldav/pkg/caldav_client"
	"os"
	"os/exec"
)

func resetTerminal() {
	cmd := exec.Command("reset")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	var caldavUrl string
	var clientName string
	var clientToken string
	flag.StringVar(&caldavUrl, "url", "https://caldav.yandex.ru", "CalDAV server url")
	flag.StringVar(&clientName, "name", "test", "Server port ")
	flag.StringVar(&clientToken, "token", "test", "Your id")
	flag.Parse()
	client, err := caldav.NewCaldavClient(clientName, clientToken, caldavUrl, context.Background())
	if err != nil {
		fmt.Println(err)
	}
	//client.GetCalendarsNames(context.Background())
	defer resetTerminal()
	for {
		var command string
		fmt.Scan(&command)
		switch command {
		case "GetCalendarsNames":
			calendarNames := client.GetCalendarsNames(context.Background())
			fmt.Println(calendarNames)
		case "GetCalendars":
			resp, err := client.GetCalendars(context.Background(), "Test")
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(resp)
			}
		case "q":
			resetTerminal()
			os.Exit(0)
		default:
			fmt.Println("Error. Command not recognized")
		}
	}
}
