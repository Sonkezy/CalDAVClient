package main

import (
	"flag"
	"fmt"
	caldav "mycaldav/pkg/caldav_client"
	"os"
	"os/exec"
)

func resetTerminal() {
	cmd := exec.Command("reset")
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("Reset termnal error: ", err)
	}
}

func main() {
	var caldavURL string
	var clientName string
	var clientToken string
	flag.StringVar(&caldavURL, "url", "https://caldav.yandex.ru", "CalDAV server url ")
	flag.StringVar(&clientName, "name", "test", "Server port ")
	flag.StringVar(&clientToken, "token", "test", "Your id")
	flag.Parse()
	client, err := caldav.NewCaldavClient(clientName, clientToken, caldavURL)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Hello")
	defer resetTerminal()
ui_loop:
	for {
		var command string
		_, err := fmt.Scan(&command)
		if err != nil {
			fmt.Println(err)
		}
		switch command {
		case "GetCalendarsNames":
			calendarNames := client.GetCalendarsNames()
			fmt.Println(calendarNames)
		case "GetCalendars":
			resp, err := client.GetCalendars()
			if err != nil {
				fmt.Println(err)
			} else {
				client.OutputEvents(resp)
			}
		case "CreateEvent":
			client.CreateEvent()
		case "q":
			resetTerminal()
			break ui_loop
		default:
			fmt.Println("Error. Command not recognized")
		}
	}
}
