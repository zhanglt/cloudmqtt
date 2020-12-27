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
	"context"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/michaelestrin/cloudmqtt/internal/cloudmqtt/contract"
)

// notify 是一个封装元数据查询转发实现的接收器。
type notify struct {
	loggingClient  logger.LoggingClient
	send           contract.Sender
	marshal        contract.Marshaller
	metadataClient contract.MetadataClient
}

// NewNotifier 是一个构造函数，它返回一个被配置为与EdgeX core-metadata实例通信的notify实例
func NewNotifier(
	loggingClient logger.LoggingClient,
	send contract.Sender,
	marshal contract.Marshaller,
	metadataClient contract.MetadataClient) *notify {

	return &notify{
		loggingClient:  loggingClient,
		send:           send,
		marshal:        marshal,
		metadataClient: metadataClient,
	}
}

// deviceCallFailedLogMessage 函数格式化并返回设备调用失败时的日志消息。
func deviceCallFailedLogMessage(eventId string, errorMessage string) string {
	return fmt.Sprintf("device call failed for %s (%s)", eventId, errorMessage)
}

// marshalFailedLogMessage 函数格式化并返回当试图封送类型失败时的日志消息。
func marshalFailedLogMessage(eventId string, errorMessage string) string {
	return fmt.Sprintf("marshal failed for %s (%s)", eventId, errorMessage)
}

// Notify 方法实现通知器约定;它查询EdgeX core-metadata实例以获得特定设备的元数据，并将结果向北向转发
func (n *notify) Notify(event *models.Event) bool {
	result, err := n.metadataClient.DeviceForName(event.Device, context.Background())
	if err != nil {
		n.loggingClient.Error(deviceCallFailedLogMessage(event.ID, err.Error()))
		return false
	}

	bytes, err := n.marshal(result)
	if err != nil {
		n.loggingClient.Error(marshalFailedLogMessage(event.ID, err.Error()))
		return false
	}

	return n.send(bytes)
}
