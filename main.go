package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/nxadm/tail"
)

const usage = `Usage of Snort3 Parser:
 -h, --help
    Show this usage
 -f, --snort-alert-file
    Snort v3 JSON Log Alert File Path
 -h, --mqtt-host
    MQTT Broker Host (default: localhost)
 -p, --mqtt-port
    MQTT Broker Port (default: 1883)
 -t, --topic
    MQTT Topic to send data into (default: mataelang/sensor/v3/<machine-id>)
`

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	log.Println("MQTT Client Connected.")
}

var connectionLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	log.Printf("Connection Lost: %s\n", err.Error())
}

func main() {
	machineID, err := machineid.ID()
	if err != nil {
		log.Println("Cannot get machine unique ID")
		machineID = "anonymous"
	}
	var (
		mqttBrokerHost     string
		mqttBrokerPort     string
		mqttTopic          string
		snortAlertFilePath string
		errorCount         = 0
		successCount       = 0
		messageCount       = 0
	)

	flag.StringVar(&mqttBrokerHost, "host", "127.0.0.1", "MQTT Broker Host")
	flag.StringVar(&mqttBrokerHost, "H", "127.0.0.1", "MQTT Broker Host")
	flag.StringVar(&mqttBrokerPort, "port", "1883", "MQTT Broker Port")
	flag.StringVar(&mqttBrokerPort, "P", "1883", "MQTT Broker Port")
	flag.StringVar(&mqttTopic, "topic", "mataelang/sensor/v3/<machine-id>", "MQTT Broker Topic")
	flag.StringVar(&mqttTopic, "t", "mataelang/sensor/v3/<machine-id>", "MQTT Broker Topic")
	flag.StringVar(&snortAlertFilePath, "snort-alert-path", "", "Snort v3 JSON Log Alert File Path")
	flag.StringVar(&snortAlertFilePath, "f", "", "Snort v3 JSON Log Alert File Path")
	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	if snortAlertFilePath == "" {
		fmt.Printf("snort-alert-path cannot be null. Check the required parameter. Exiting.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	mqttTopic = strings.ReplaceAll(mqttTopic, "<machine-id>", machineID)

	log.Println("MQTT Broker Host\t: " + mqttBrokerHost)
	log.Println("MQTT Broker Port\t: " + mqttBrokerPort)
	log.Println("MQTT Broker Topic\t: " + mqttTopic)

	log.Println("Snort Alert File Path\t: " + snortAlertFilePath)

	log.Print("Checking snort alert file is exist...")
	if _, err := os.Stat(snortAlertFilePath); errors.Is(err, os.ErrNotExist) {
		log.Println("\nSnort alert file at " + snortAlertFilePath + ", does not exist.")
		log.Fatalln("Cannot continue, exiting.")
	}
	log.Println("\tOk, found.")

	var broker = "tcp://" + mqttBrokerHost + ":" + mqttBrokerPort

	options := mqtt.NewClientOptions()
	options.AddBroker(broker)
	options.SetClientID("mataelang_sensor_snort_v3_" + machineID)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	client := mqtt.NewClient(options)
	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		log.Fatalln(token.Error())
	}

	messages := make(chan string)

	// Create a tail process
	t, err := tail.TailFile(
		snortAlertFilePath, tail.Config{Follow: true})
	if err != nil {
		log.Fatalln(err)
	}

	// Create routine for sending message from messages channel
	go func() {
		for textLine := range messages {
			messageCount += 1
			log.Printf("Sending snort log... ")
			token = client.Publish(mqttTopic, 0, false, textLine)
			if token.Wait() && token.Error() != nil {
				errorCount += 1
				continue
			}
			fmt.Printf("[ok]\n")
			successCount += 1
		}
	}()

	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("Total=" + strconv.Itoa(messageCount) + "\tSuccess=" + strconv.Itoa(successCount) + "\tFailed=" + strconv.Itoa(errorCount) + "\tError Rate=" + strconv.Itoa((errorCount/messageCount)*100) + "%")
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	// Send the message to channel
	for line := range t.Lines {
		messages <- line.Text
	}
}
