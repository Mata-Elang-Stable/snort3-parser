package internal

import (
	"fmt"
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("[INFO] MQTT Client Connected.")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("[WARN] Connection Lost: %s\n", err.Error())
}

func InitMQTT(host string, port int, username string, password string) mqtt.Client {
	var appID = uuid.New()
	var broker = fmt.Sprintf("tcp://%s:%d", host, port)
	options := mqtt.NewClientOptions()
	options.AddBroker(broker)
	options.SetClientID(appID.String())
	if username != "" && password != "" {
		options.SetUsername(username)
		options.SetPassword(password)
	}
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	return mqtt.NewClient(options)
}
