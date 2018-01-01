package etherdelta

import (
	"encoding/json"
	"errors"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"log"
	"time"
)

const ED_WS_URL = "wss://socket.etherdelta.com/socket.io/?EIO=3&transport=websocket"

func NewWSClient(isConnected chan bool) WSClient {
	sockclient, err := gosocketio.Dial(
		ED_WS_URL,
		transport.GetDefaultWebsocketTransport(),
	)

	client := WSClient{
		client: sockclient,
	}

	if err != nil {
		isConnected <- false
		log.Println("Error connecting to EtherDelta websocket URI:", err)
		return client
	}

	err = client.client.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		isConnected <- true
		log.Println("Connected to EtherDelta websocket.")
	})

	if err != nil {
		isConnected <- false
		log.Println(err)
		return client
	}

	return client
}

func (client WSClient) EmitRequest(topic string, requestBody *WSEmitBody) error {
	message, err := json.Marshal(requestBody)

	if err != nil {
		return err
	}

	msg := string(message)

	err = client.client.Emit(topic, requestBody)

	if err != nil {
		return err
	}

	log.Printf("Emitted EtherDelta websocket request for \"%s\" topic with payload %s", topic, msg)

	return nil
}

func (client WSClient) EmitListenOnceAndClose(topic string, requestBody *WSEmitBody, messageChan chan *WSResponse, emitTopic string) {
	result := &WSResponse{}
	expired := false

	go func() {
		err := client.client.On(topic, func(h *gosocketio.Channel, message Message) {
			log.Printf(`Got websocket data for "%s" topic`, topic)
			//log.Println(message)

			if !expired {
				result.Message = message
				messageChan <- result
				client.client.Close()
				close(messageChan)
				expired = true
			}
		})

		if err != nil {
			if !expired {
				result.Error = err
				messageChan <- result
				client.client.Close()
				close(messageChan)
				expired = true
			}
		}

		if emitTopic == "message" {
			err = client.PostOrder(requestBody.Order)
		} else {
			err = client.EmitRequest(emitTopic, requestBody)
		}

		if err != nil {
			if !expired {
				result.Error = err
				messageChan <- result
				client.client.Close()
				close(messageChan)
				expired = true
			}
		}
	}()

	go func() {
		select {
		case <-time.After(time.Second * 60):
			if result.Error == nil && !expired {
				result.Error = errors.New("Websocket response timeout")
				messageChan <- result
				client.client.Close()
				close(messageChan)
				expired = true
			}
		}
	}()
}

func (client WSClient) PostOrder(order OrderPost) error {
	message, err := json.Marshal(order)

	if err != nil {
		return err
	}

	msg := string(message)

	topic := "message"
	err = client.client.Emit(topic, order)

	if err != nil {
		return err
	}

	log.Printf("Emitted EtherDelta websocket request for \"%s\" topic with order payload %s", topic, msg)

	return nil
}
