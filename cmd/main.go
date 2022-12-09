package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/denisbrodbeck/machineid"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mata-elang-stable/snort3-parser/internal"
	"github.com/nxadm/tail"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func main() {
	machineID, err := machineid.ID()
	if err != nil {
		log.Println("[WARN] Cannot get machine unique ID, set machine-id to \"anonymous\"")
		machineID = "anonymous"
	}

	var (
		mqttBrokerHost, mqttBrokerUsername, mqttBrokerPassword, mqttTopic string
		sensorID, snortAlertFilePath                                      string
		mqttBrokerPort                                                    int
		noCLI, verboseLog                                                 = false, false
		successCount, messageCount                                        = 0, 0
		statsIntervalSec                                                  = 10
	)

	flag.StringVar(&mqttBrokerHost, "H", "127.0.0.1", "MQTT Broker Host")
	flag.IntVar(&mqttBrokerPort, "P", 1883, "MQTT Broker Port")
	flag.StringVar(&mqttBrokerUsername, "u", "", "MQTT Broker Username")
	flag.StringVar(&mqttBrokerPassword, "p", "", "MQTT Broker Password")
	flag.StringVar(&sensorID, "s", "<machine-id>", "Sensor ID")
	flag.StringVar(&mqttTopic, "t", "mataelang/sensor/v3/<machine-id>", "MQTT Broker Topic")
	flag.StringVar(&snortAlertFilePath, "f", "/var/log/snort/alert_json.txt", "Snort v3 JSON Log Alert File Path")
	flag.BoolVar(&noCLI, "b", false, "Wheter to use flag or environment variable")
	flag.BoolVar(&verboseLog, "v", false, "Verbose payload to stdout")
	flag.IntVar(&statsIntervalSec, "d", 10, "Log Statistics interval in second")
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

		mqttBrokerUsername = os.Getenv("MQTT_USERNAME")
		mqttBrokerPassword = os.Getenv("MQTT_PASSWORD")

		mqttTopic = os.Getenv("MQTT_TOPIC")
		if mqttTopic == "" {
			mqttTopic = "mataelang/sensor/v3/<machine-id>"
		}

		sensorID = os.Getenv("SENSOR_ID")
		if sensorID == "" {
			sensorID = "<machine-id>"
		}
	}

	if _, err := internal.ValidatePort(mqttBrokerPort); err != nil {
		log.Fatal(err)
	}

	if snortAlertFilePath == "" {
		log.Printf("[ERROR] Snort Alert Path cannot be null. Exiting.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("[INFO] Loading configuration.")

	sensorID = strings.ReplaceAll(sensorID, "<machine-id>", machineID)
	mqttTopic = strings.ReplaceAll(mqttTopic, "<machine-id>", machineID)
	mqttTopic = strings.ReplaceAll(mqttTopic, "<sensor-id>", sensorID)

	log.Printf("[INFO] Sensor ID\t\t: %s\n", sensorID)
	log.Printf("[INFO] MQTT Broker Host\t: %s\n", mqttBrokerHost)
	log.Printf("[INFO] MQTT Broker Port\t: %d\n", mqttBrokerPort)
	log.Printf("[INFO] MQTT Broker Topic\t: %s\n", mqttTopic)
	log.Printf("[INFO] Snort Alert Path\t: %s\n", snortAlertFilePath)

	log.Printf("[INFO] Checking snort alert file is exist...\n")
	if _, err := os.Stat(snortAlertFilePath); errors.Is(err, os.ErrNotExist) {
		log.Printf("[ERROR] The snort alert file at %s does not exist\n", snortAlertFilePath)
		log.Fatalln("[ERROR] Cannot continue, exiting")
	}
	log.Printf("[INFO] Snort alert file exist\n")

	var client mqtt.Client = internal.InitMQTT(mqttBrokerHost, mqttBrokerPort, mqttBrokerUsername, mqttBrokerPassword)
	var token mqtt.Token = client.Connect()

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

	log.Printf("[INFO] Start sending logs")

	p := message.NewPrinter(language.AmericanEnglish)

	// Create routine for sending message from messages channel
	go func() {
		for textLine := range messages {
			messageCount += 1
			payload, err := json.Marshal(textLine)
			if err != nil {
				log.Println(err)
			}
			k := client.Publish(mqttTopic, 0, true, payload)
			if k.Wait() && k.Error() != nil {
				log.Println(k.Error())
				continue
			}
			successCount += 1
		}
	}()

	ticker := time.NewTicker(time.Duration(statsIntervalSec) * time.Second)
	tickerLog := time.NewTicker(time.Duration(60) * time.Second)

	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				tempMessageCount, tempSuccessCount := messageCount, successCount
				messageCount, successCount = 0, 0
				if successCount > messageCount {
					messageCount = successCount
				}
				tempErrorCount := tempMessageCount - tempSuccessCount
				tempMessageRate := tempMessageCount / statsIntervalSec

				log.Printf("[STATS] Total=%d\tSuccess=%d\tFailed=%d\tAvgRate=%s message/second\n",
					tempMessageCount, tempSuccessCount, tempErrorCount, p.Sprintf("%v", tempMessageRate))

			case <-tickerLog.C:
				files, err := filepath.Glob(fmt.Sprintf("%s.*", snortAlertFilePath))
				if err != nil {
					log.Printf("[INFO] No rotated log file found.")
					continue
				}
				for _, f := range files {
					if err := os.Remove(f); err != nil {
						log.Printf("[WARN] Cannot remove %s file", f)
						continue
					}
					log.Printf("[INFO] File %s is removed.", f)
				}
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

		if verboseLog {
			log.Printf("[DEBUG] PAYLOAD - %s\n", fmt.Sprint(payload))
		}

		messages <- payload
	}
}
