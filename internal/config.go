package internal

type mqttConfig struct {
	host string
	port int
	username string
	password string
}

type Config struct {
	mqtt mqttConfig
	
}