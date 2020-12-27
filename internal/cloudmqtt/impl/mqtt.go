/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

package impl

import (
	"crypto/tls"
	"fmt"
	"os"
	"time"

	mqttlib "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/michaelestrin/cloudmqtt/internal/cloudmqtt/contract"
)

const qosAtLeastOnce = 1

// mqtt是包装单向MQTTS实现的接收器。
type mqtt struct {
	loggingClient  logger.LoggingClient
	client         mqttlib.Client
	eventTopic     string
	newDeviceTopic string
	commandTopic   string
	receiver       contract.Receiver
}

// NewMqttInstanceForCloud 是一个构造函数，它返回一个为基于云的mqtt配置的mqtt接收器
func NewMqttInstanceForCloud(
	loggingClient logger.LoggingClient,
	certFile string,
	keyFile string,
	clientId string,
	userName string,
	password string,
	server string,
	eventTopic string,
	newDeviceTopic string,
	commandTopic string,
	receiver contract.Receiver) (q *mqtt) {

	q = &mqtt{
		loggingClient:  loggingClient,
		eventTopic:     eventTopic,
		newDeviceTopic: newDeviceTopic,
		commandTopic:   commandTopic,
		receiver:       receiver,
	}

	tlsConfig := &tls.Config{}
	if len(certFile) > 0 && len(keyFile) > 0 {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			q.loggingClient.Error(fmt.Sprintf("mqtt mqttInstanceForCloud LoadX509KeyPair failed: %v", err))
			os.Exit(-1)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	} else {
		tlsConfig.ClientAuth = tls.NoClientCert
		tlsConfig.ClientCAs = nil
	}

	options := mqttlib.ClientOptions{
		ClientID:             clientId,
		Username:             userName,
		Password:             password,
		CleanSession:         true,
		AutoReconnect:        true,
		MaxReconnectInterval: 1 * time.Second,
		KeepAlive:            int64(30 * time.Second),
		TLSConfig:            tlsConfig,
	}
	options.AddBroker(server)
	q.client = mqttlib.NewClient(&options)

	if token := q.client.Connect(); token.Wait() && token.Error() != nil {
		q.loggingClient.Error(fmt.Sprintf("mqtt mqttInstanceForCloud Connect failed: %v", token.Error()))
		os.Exit(-1)
	}

	if token := q.client.Subscribe(commandTopic, 1, q.receive); token.Wait() && token.Error() != nil {
		q.loggingClient.Error(fmt.Sprintf("mqtt mqttInstanceForCloud Subscribe failed: %v", token.Error()))
		os.Exit(-1)
	}
	return
}

// receive delegates handling of southbound command to provided receiver contract implementation.
// recevice 委托处理南向命令提供的接收方约定实现。
func (q *mqtt) receive(client mqttlib.Client, message mqttlib.Message) {
	q.receiver(string(message.Payload()))
}

// send 函数在指定的北向MQTT topic 上发布内容
func send(q *mqtt, topicName string, content []byte) bool {
	if token := q.client.Publish(topicName, qosAtLeastOnce, false, content); token.Wait() && token.Error() != nil {
		q.loggingClient.Warn("mqtt send to " + topicName + " failed (" + token.Error().Error() + ")")
		return false
	}
	return true
}

// EventSender 将内容传输到北向的MQTT事件topic。
func (q *mqtt) EventSender(content []byte) bool {
	return send(q, q.eventTopic, content)
}

// NewDeviceSender 方法将内容传输到北向的MQTT新设备topic。
func (q *mqtt) NewDeviceSender(content []byte) bool {
	return send(q, q.newDeviceTopic, content)
}

func (q *mqtt) CleanUp() {
	if token := q.client.Unsubscribe(q.commandTopic); token.Wait() && token.Error() != nil {
		q.loggingClient.Error(fmt.Sprintf("mqtt mqttInstanceForCloud Unsubscribe failed: %v", token.Error()))
	}
}
