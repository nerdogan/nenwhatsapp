package main

import (
	"context"
	"fmt"
	"mime"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var client *whatsmeow.Client

func eventHandler(evt interface{}) {

	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("  ... ")
		fmt.Println(" ....")
		fmt.Println("Received a message!", v.Message.GetConversation())
		img := v.Message.GetImageMessage()
		if img != nil {
			data, err := client.Download(img)
			if err != nil {
				fmt.Printf("Failed to download image: %v", err)
				return
			}
			exts, _ := mime.ExtensionsByType(img.GetMimetype())
			path := fmt.Sprintf("%s-%s%s", "deneme", v.Info.ID, exts[0])
			err = os.WriteFile(path, data, 0600)
			if err != nil {
				fmt.Printf("Failed to save image: %v", err)
				return
			}
			fmt.Printf("Saved image in message to %s", path)
		}
		fmt.Println("...... ")

	}
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite
	container, err := sqlstore.New("sqlite3", "file:examplestore.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	clientLog := waLog.Stdout("Client", "DEBUG", true)
	client = whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(eventHandler)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}

	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	for {
		_, err = client.SendMessage(context.Background(), types.JID{
			User:   "905533155775",
			Server: types.DefaultUserServer,
		}, "", &waProto.Message{
			Conversation: proto.String("Deneme "),
		})
		time.Sleep(100 * time.Second)

	}

	client.Disconnect()
}
