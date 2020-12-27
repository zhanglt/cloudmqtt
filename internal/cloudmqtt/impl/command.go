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
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

// receiver 是一个接收器，提供一个记录入站命令的示例接收器
type commandHandler struct {
	loggingClient logger.LoggingClient
}

// NewCommandHandler 是一个返回receiver实例的构造函数。
func NewCommandHandler(loggingClient logger.LoggingClient) *commandHandler {
	return &commandHandler{
		loggingClient: loggingClient,
	}
}

// receivedCommandLogMessage 功能对接收到的命令进行格式化并返回日志消息。
func receivedCommandLogMessage(command string) string {
	return fmt.Sprintf("command received: %s", command)
}

// Receiver method implements Receiver contract; it logs the incoming command but could be extended to interpret
// the string content and call an endpoint on the EdgeX core-command service.
// Receiver方法实现Receiver协定；它记录传入的命令,可以扩展为解释字符串内容并调用EdgeX core-command服务上的endpoint。
func (c *commandHandler) Receiver(command string) {
	c.loggingClient.Debug(receivedCommandLogMessage(command))

	// translate incoming command string into call to appropriate core-command endpoint
}
