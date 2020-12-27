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

package cloudmqtt

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/michaelestrin/cloudmqtt/internal/cloudmqtt/contract"
)

// transport是一个包装通用事件和元数据导出适配器的接收器。
type transport struct {
	loggingClient                logger.LoggingClient
	sendFailureWaitInNanoseconds time.Duration
	send                         contract.Sender
	notify                       contract.Notifier
	marshal                      contract.Marshaller
	cleanUp                      contract.CleanUp
	wg                           sync.WaitGroup
	events                       chan *models.Event
}

// NewTransport is a constructor that returns a configured transport receiver whose Run() method can be included in
// a call to the EdgeX Applications Functions SDK's SetFunctionsPipeline() method.
// NewTransport是一个构造函数，它返回一个配置的传输接收器，它的Run()方法
// 可以包含在EdgeX应用程序函数SDK的SetFunctionsPipeline()方法的调用中
func NewTransport(
	loggingClient logger.LoggingClient,
	sendFailureWaitInNanoseconds time.Duration,
	send contract.Sender,
	notify contract.Notifier,
	marshal contract.Marshaller,
	cleanUp contract.CleanUp) *transport {

	t := &transport{
		loggingClient:                loggingClient,
		sendFailureWaitInNanoseconds: sendFailureWaitInNanoseconds,
		send:                         send,
		notify:                       notify,
		marshal:                      marshal,
		cleanUp:                      cleanUp,
		events:                       make(chan *models.Event, 16),
	}
	t.wg.Add(1)
	go t.newDeviceHandler()
	return t
}

//detectedNewDeviceLogMessage 函数对检测到的新设备进行格式化并返回日志消息。
func detectedNewDeviceLogMessage(deviceName string) string {
	return fmt.Sprintf("detected new device %s", deviceName)
}

// newDeviceHandler方法被构造函数作为goroutine执行，负责跟踪已知的设备并为新设备调用通知器实现。
func (t *transport) newDeviceHandler() {
	defer t.wg.Done()

	devices := make(map[string]bool)
	for event := range t.events {
		_, ok := devices[event.Device]
		if !ok {
			if t.notify(event) {
				devices[event.Device] = true
				t.loggingClient.Debug(detectedNewDeviceLogMessage(event.Device))
			}
		}
	}
}

// newDeviceHandler 函数格式化并返回当试图封送类型失败时的日志消息。
func marshalFailedLogMessage(eventId string, errorMessage string) string {
	return fmt.Sprintf("marshal failed for %s (%s)", eventId, errorMessage)
}

// sentLogMessage 函数格式化并返回成功向北向发送事件的日志消息。
func sentLogMessage(eventId string) string {
	return fmt.Sprintf("sent for %s", eventId)
}

// handleEvent方法向北向传输事件
func (t *transport) handleEvent(EdgeXContext contract.EdgeXContext, event *models.Event) {
	bytes, err := t.marshal(event)
	if err != nil {
		t.loggingClient.Warn(marshalFailedLogMessage(event.ID, err.Error()))
		return
	}

	for {
		if t.send(bytes) {
			t.loggingClient.Debug(sentLogMessage(event.ID))

			err = EdgeXContext.MarkAsPushed()
			if err != nil {
				t.loggingClient.Error(err.Error())
			}

			return
		}
		time.Sleep(t.sendFailureWaitInNanoseconds)
	}
}

// run方法是内部实现，由公共访问的run()委托;实施以方便单元测试
func (t *transport) run(EdgeXContext contract.EdgeXContext, params ...interface{}) (bool, interface{}) {
	for _, param := range params {
		if event, ok := param.(models.Event); ok {
			t.events <- &event
			t.handleEvent(EdgeXContext, &event)
		}
	}
	return true, params
}

// Run 方法是EdgeX应用程序函数sdk兼容的函数，可以包含在调用它的SetFunctionsPipeline()方法中。
func (t *transport) Run(EdgeXContext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	return t.run(EdgeXContext, params[0])
}

// CleanUp 方法确保newDeviceHandler() goroutine已完成。
func (t *transport) CleanUp() {
	close(t.events)
	t.wg.Wait()
	t.cleanUp()
}
