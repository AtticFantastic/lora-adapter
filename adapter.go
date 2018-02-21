package lora

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/containous/traefik/log"
	"github.com/eclipse/paho.mqtt.golang"
)

// Adapter implements a MQTT pub-sub adapter between LoRa Sever and Mainflux
type Adapter struct {
	conn mqtt.Client
	// Are we connecting to LoRa Server or to Mainflux
	isLora bool
	mutex  sync.RWMutex
}

const (
	loraServerTopic string = "application/+/node/+/rx"
	mainfluxTopic   string = "/lora"
)

var (
	loraAdapter     *Adapter
	mainfluxAdapter *Adaper
)

// NewAdapter creates a new Adapter
func NewAdapter(server, username, password string, isLora bool) (*Adapter, error) {
	b := Adapter{}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(server)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetOnConnectHandler(b.onConnected)
	opts.SetConnectionLostHandler(b.onConnectionLost)

	log.WithField("server", server).Info("backend: connecting to mqtt broker")
	b.conn = mqtt.NewClient(opts)
	if token := b.conn.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	b.isLora = isLora

	return &b, nil
}

// Send MQTT message
func (b *Adapter) SendMQTTMsg(topic string, data []byte) error {
	if token := b.conn.Publish(topic, 0, false, data); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

// Close closes the backend
func (b *Adapter) Close() {
	b.conn.Disconnect(250) // wait 250 milisec to complete pending actions
}

// Subscribe to lora server messages
func (b *Adapter) Sub() error {
	switch b.isLora {
	case true:
		if s := b.conn.Subscribe(loraServerTopic, 0, b.MessageHandler); s.Wait() && s.Error() != nil {
			log.Info("Failed to subscribe, err: %v\n", s.Error())
			return s.Error()
		}
	case false:
		// For now we do not SUB to Mainflux
		break
	}

	return nil
}

// Handler for received messages from loraserver
func (b *Adapter) MessageHandler(c mqtt.Client, msg mqtt.Message) {
	switch b.isLora {
	case true:
		// Mainflux backend is subscribed to LoRa Network Server and recieves LoRa messages
		u := LoraMessage{}
		errStatus := json.Unmarshal(msg.Payload(), &u)
		if errStatus != nil {
			log.Errorf("\nerror: decode json failed")
			log.Errorf(errStatus.Error())
			return
		}

		fmt.Printf("\n <-- RCVD DATA: %s\n", u.Data)
		data, err := base64.StdEncoding.DecodeString(u.Data)
		if err != nil {
			log.Errorf("\nerror: decode base64 failed")
		}

		mainfluxBackend.SendMQTTMsg(mainfluxTopic, data)
	case false:
		// LoRa backend is not currently subsctibed to Mainflux MQTT broker
		break
	}

}

func (b *Adapter) onConnected(c mqtt.Client) {
	defer b.mutex.RUnlock()
	b.mutex.RLock()
	log.Info("backend: connected to mqtt broker")
}

func (b *Backend) onConnectionLost(c mqtt.Client, reason error) {
	log.Errorf("backend: mqtt connection error: %s", reason)
}
