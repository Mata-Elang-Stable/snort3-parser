package main

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nxadm/tail"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("MQTT Client Connected.")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connection Lost: %s\n", err.Error())
}

func main() {
	var broker = "tcp://192.168.1.121:1883"

	var default_topic = "mataelang/sensor/v3"

	options := mqtt.NewClientOptions()
	options.AddBroker(broker)
	options.SetClientID("mataelang_sensor_snort_v3")
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	client := mqtt.NewClient(options)
	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	topic := default_topic

	filePath := "/home/fadhilyori/Projects/mata-elang/stable/snort3/snort_data/alert_json.txt"
	// filePath := "/home/fadhilyori/Projects/mata-elang/stable/snort3/test.log"

	messages := make(chan string)

	// Create a tail process
	t, err := tail.TailFile(
		filePath, tail.Config{Follow: true})
	if err != nil {
		panic(err)
	}

	// Create routine for sending message from messages channel
	go func() {
		for textLine := range messages {
			log.Printf("Sending snort log... ")
			token = client.Publish(topic, 0, false, textLine)
			token.Wait()
			log.Printf("[ok]\n")
		}
	}()

	// Send the message to channel
	for line := range t.Lines {
		messages <- line.Text
	}
}
