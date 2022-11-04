package main

import (
	"encoding/json"
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
	"github.com/mata-elang-stable/snort3-parser/internal"
	"github.com/nxadm/tail"
)

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
		mqttBrokerPort     int
		mqttTopic          string
		snortAlertFilePath string
		sensorID           string
		mqttClientID       string
		noCLI              = false
		errorCount         = 0
		successCount       = 0
		messageCount       = 0
	)

	flag.StringVar(&mqttBrokerHost, "H", "127.0.0.1", "MQTT Broker Host")
	flag.Var(internal.PortVar(&mqttBrokerPort), "P", "MQTT Broker Port")
	flag.StringVar(&sensorID, "s", "<machine-id>", "Sensor ID")
	flag.StringVar(&mqttTopic, "t", "mataelang/sensor/v3/<machine-id>", "MQTT Broker Topic")
	flag.StringVar(&snortAlertFilePath, "f", "/var/log/snort/alert_json.txt", "Snort v3 JSON Log Alert File Path")
	flag.BoolVar(&noCLI, "b", true, "Wheter to use flag or environment variable")
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()

	if noCLI {
		snortAlertFilePath = os.Getenv("SNORT_ALERT_FILE_PATH")
		if snortAlertFilePath == "" {
			snortAlertFilePath = "/var/log/snort/alert_json.txt"
		}
		mqttBrokerHost = os.Getenv("MQTT_HOST")
		mqttBrokerPort, err = strconv.Atoi(os.Getenv("MQTT_PORT"))
		if err != nil {
			mqttBrokerPort = 1883
		}
		mqttTopic = os.Getenv("MQTT_TOPIC")
		if mqttTopic == "" {
			mqttTopic = "mataelang/sensor/v3/<machine-id>"
		}
		
		sensorID = os.Getenv("SENSOR_ID")
		if sensorID == "" {
			sensorID = "<machine-id>"
		}
	}

	if snortAlertFilePath == "" {
		fmt.Printf("snort-alert-path cannot be null. Check the required parameter. Exiting.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	mqttTopic = strings.ReplaceAll(mqttTopic, "<machine-id>", machineID)
	sensorID = strings.ReplaceAll(sensorID, "<machine-id>", machineID)
	mqttClientID = fmt.Sprintf("mataelang_sensor_snort_v3_%s", machineID)

	log.Printf("MQTT Broker Host\t: %s\n", mqttBrokerHost)
	log.Printf("MQTT Broker Port\t: %d\n", mqttBrokerPort)
	log.Printf("MQTT Broker Topic\t: %s\n", mqttTopic)
	log.Printf("Snort Alert File Path\t: %s\n", snortAlertFilePath)

	log.Printf("Checking snort alert file is exist...\n")
	if _, err := os.Stat(snortAlertFilePath); errors.Is(err, os.ErrNotExist) {
		log.Printf("\nSnort alert file at %s, does not exist.\n", snortAlertFilePath)
		log.Fatalln("Cannot continue, exiting.")
	}
	log.Printf("Snort alert file exist.\n")

	var broker = fmt.Sprintf("tcp://%s:%d", mqttBrokerHost, mqttBrokerPort)

	options := mqtt.NewClientOptions()
	options.AddBroker(broker)
	options.SetClientID(mqttClientID)
	options.OnConnect = connectHandler
	options.OnConnectionLost = connectionLostHandler

	client := mqtt.NewClient(options)
	token := client.Connect()

	if token.Wait() && token.Error() != nil {
		log.Fatalln(token.Error())
	}

	messages := make(chan map[string]interface{})

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
			payload, err := json.Marshal(textLine)
			if err != nil {
				log.Println(err)
			}
			token = client.Publish(mqttTopic, 0, false, payload)
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
		var payload map[string]interface{}
		err = json.Unmarshal([]byte(line.Text), &payload)
		if err != nil {
			log.Printf("ERROR - Cannot parse event log")
			continue
		}
		payload["sensor_id"] = sensorID
		log.Println(payload)
		messages <- payload
	}
}
