package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func main() {

	token := "xoxb-5686869092359-5742892319394-DHVYnC4YEG9M1gsZ4IMxA84d"
	appToken := "xapp-1-A05MUPC098T-5750387648388-e149aa33939d1fec5eac6bd80f68f2d44c67bbfda4b4f23c9191a5825a5c8d46"
	// Create a new client to slack by giving token
	// Set debug to true while developing
	// Also add a ApplicationToken option to the client
	client := slack.New(token, slack.OptionDebug(true), slack.OptionAppLevelToken(appToken))
	// go-slack comes with a SocketMode package that we need to use that accepts a Slack client and outputs a Socket mode client instead
	socketClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		// Option to set a custom logger
		socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	// Create a context that can be used to cancel goroutine
	ctx, cancel := context.WithCancel(context.Background())
	// Make this cancel called properly in a real program , graceful shutdown etc
	defer cancel()

	go func(ctx context.Context, client *slack.Client, socketClient *socketmode.Client) {

		// Create a for loop that selects either the context cancellation or the events incomming
		for {
			select {
			// inscase context cancel is called exit the goroutine
			case <-ctx.Done():
				log.Println("Shutting down socketmode listener")
				return
			case event := <-socketClient.Events:

				// We have a new Events, let's type switch the event
				// Add more use cases here if you want to listen to other events.
				switch event.Type {
				// handle EventAPI events
				case socketmode.EventTypeEventsAPI:

					// The Event sent on the channel is not the same as the EventAPI events so we need to type cast it
					eventsAPIEvent, ok := event.Data.(slackevents.EventsAPIEvent)
					if !ok {
						log.Printf("Could not type cast the event to the EventsAPIEvent: %v\n", event)
						continue
					}
					// We need to send an Acknowledge to the slack server
					socketClient.Ack(*event.Request)
					// Now we have an Events API event, but this event type can in turn be many types, so we actually need another type switch
					err := handleEventMessage(eventsAPIEvent)
					if err != nil {
						// Replace with actual err handeling
						log.Fatal(err)
					}
				// Handle Slash Commands
				case socketmode.EventTypeSlashCommand:
					// Just like before, type cast to the correct event type, this time a SlashEvent
					command, ok := event.Data.(slack.SlashCommand)
					if !ok {
						log.Printf("Could not type cast the message to a SlashCommand: %v\n", command)
						continue
					}
					// Dont forget to acknowledge the request
					socketClient.Ack(*event.Request)
					// handleSlashCommand will take care of the command
					err := handleSlashCommand(command, client)
					if err != nil {
						log.Fatal(err)
					}

				}
			}

		}
	}(ctx, client, socketClient)

	socketClient.Run()

}

// handleSlashCommand will take a slash command and route to the appropriate function
func handleSlashCommand(command slack.SlashCommand, client *slack.Client) error {
	fmt.Println("COMMAND", command.Command)
	// We need to switch depending on the command
	switch command.Command {
	case "/hello":
		// This was a hello command, so pass it along to the proper function
		return handleHelloCommand(command, client)
	}

	return nil
}

func handleEventMessage(event slackevents.EventsAPIEvent) error {
	switch event.Type {
	// First we check if this is an CallbackEvent
	case slackevents.CallbackEvent:

		innerEvent := event.InnerEvent
		// Yet Another Type switch on the actual Data to see if its an AppMentionEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// The application has been mentioned since this Event is a Mention event
			log.Println(ev)
		}
	default:
		return errors.New("unsupported event type")
	}
	return nil
}

// handleHelloCommand will take care of /hello submissions
func handleHelloCommand(command slack.SlashCommand, client *slack.Client) error {

	blocks := createFormBlocks()

	modalPayload := buildModal(blocks)

	_, err := client.OpenView(command.TriggerID, modalPayload)
	if err != nil {
		log.Println("Error opening modal:", err)
	} else {
		fmt.Println("Modal opened successfully")
	}

	return nil

}

func createFormBlocks() []slack.Block {
	blocks := []slack.Block{
		slack.NewInputBlock(
			"name",
			slack.NewTextBlockObject(slack.PlainTextType, "Name:", false, false),
			nil,
			slack.NewPlainTextInputBlockElement(
				slack.NewTextBlockObject(slack.PlainTextType, "Enter the employee's name", false, false),
				"name_input",
			),
		),
		slack.NewInputBlock(
			"cinc",
			slack.NewTextBlockObject(slack.PlainTextType, "CINC:", false, false),
			nil,
			slack.NewPlainTextInputBlockElement(
				slack.NewTextBlockObject(slack.PlainTextType, "Enter the CINC number", false, false),
				"cinc_input",
			),
		),
		slack.NewInputBlock(
			"dob",
			slack.NewTextBlockObject(slack.PlainTextType, "Date of Birth:", false, false),
			nil,
			slack.NewDatePickerBlockElement(
				"dob_input",
			),
		),
		slack.NewInputBlock(
			"gender",
			slack.NewTextBlockObject(slack.PlainTextType, "Gender:", false, false),
			nil,
			slack.NewPlainTextInputBlockElement(
				slack.NewTextBlockObject(slack.PlainTextType, "Enter the gender", false, false),
				"gender_input",
			),
		),
		slack.NewActionBlock(
			"image",
			slack.NewButtonBlockElement(
				"image_upload_button",
				"Upload Image",
				slack.NewTextBlockObject(slack.PlainTextType, "Upload an image file", false, false),
			).WithStyle(slack.StylePrimary),
		),
		slack.NewInputBlock(
			"personal_email",
			slack.NewTextBlockObject(slack.PlainTextType, "Personal Email:", false, false),
			nil,
			slack.NewPlainTextInputBlockElement(
				slack.NewTextBlockObject(slack.PlainTextType, "Enter the personal email", false, false),
				"personal_email_input",
			),
		),
		slack.NewInputBlock(
			"mobile",
			slack.NewTextBlockObject(slack.PlainTextType, "Mobile Number:", false, false),
			nil,
			slack.NewPlainTextInputBlockElement(
				slack.NewTextBlockObject(slack.PlainTextType, "Enter the mobile number", false, false),
				"mobile_input",
			),
		),
		slack.NewInputBlock(
			"home_address",
			slack.NewTextBlockObject(slack.PlainTextType, "Home Address:", false, false),
			nil,
			slack.NewPlainTextInputBlockElement(
				slack.NewTextBlockObject(slack.PlainTextType, "Enter the home address", false, false),
				"home_address_input",
			),
		),
		slack.NewInputBlock(
			"emergency_contact",
			slack.NewTextBlockObject(slack.PlainTextType, "Emergency Contact:", false, false),
			nil,
			slack.NewPlainTextInputBlockElement(
				slack.NewTextBlockObject(slack.PlainTextType, "Enter the emergency contact", false, false),
				"emergency_contact_input",
			),
		),
	}

	return blocks
}

func buildModal(blocks []slack.Block) slack.ModalViewRequest {
	return slack.ModalViewRequest{
		Type:       slack.VTModal,
		CallbackID: "employee_form",
		Title:      slack.NewTextBlockObject(slack.PlainTextType, "Employee Form", false, false), // Update the title text
		Blocks:     slack.Blocks{BlockSet: blocks},                                               // Wrap blocks in a slack.Blocks struct
		Close:      slack.NewTextBlockObject(slack.PlainTextType, "Close", false, false),
		Submit:     slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false),
	}
}
