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

package contract

import (
	"context"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

// Sender 定义了向云传输字节的函数约定
type Sender func(data []byte) bool

// Notifier 定义了通知云新添加设备元数据的函数约定
type Notifier func(event *models.Event) bool

// Receiver 定义处理从云接收到的南向命令的函数约定。
type Receiver func(command string)

// Marshaller defines function contract for marshalling type to []byte; supports unit testing.
//Marshaller 为marshalling type  to []blyte定义函数约定;支持单元测试。
type Marshaller func(v interface{}) ([]byte, error)

// CleanUp 为执行结束清理活动定义函数契约。
type CleanUp func()

// MetadataClient 定义了与EdgeX core-metadata服务交互的接口;定义以方便单元测试。
type MetadataClient interface {
	// DeviceForName loads the device for the specified name
	DeviceForName(name string, ctx context.Context) (models.Device, error)
}

// EdgeXContext 定义了与应用程序交互的接口定义以方便单元测试。
type EdgeXContext interface {
	MarkAsPushed() error
}
