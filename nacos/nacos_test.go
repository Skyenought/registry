// Copyright 2021 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package nacos

import (
	"context"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/stretchr/testify/assert"
)

var namingClient = getNamingClient()

// getNamingClient use to config for naming_client by default.
func getNamingClient() naming_client.INamingClient {
	// create ServerConfig
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848, constant.WithContextPath("/nacos")),
	}

	// create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithCustomLogger(nil),
		constant.WithNamespaceId(""),
		constant.WithTimeoutMs(50000),
		constant.WithUpdateCacheWhenEmpty(true),
		constant.WithNotLoadCacheAtStart(true),
	)

	// create naming client
	newClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(err)
	}
	return newClient
}

func TestWithTag(t *testing.T) {
	var opts1 []config.Option

	opts1 = append(opts1, server.WithRegistry(NewNacosRegistry(namingClient), &registry.Info{
		ServiceName: "demo.hertz-contrib.test1",
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:7512"),
		Weight:      10,
	}))
	opts1 = append(opts1, server.WithHostPorts("127.0.0.1:7512"))
	srv1 := server.New(opts1...)
	srv1.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		ctx.String(200, "pong1")
	})

	go srv1.Spin()

	time.Sleep(2 * time.Second)

	cli, _ := client.NewClient()
	r := NewNacosResolver(namingClient)
	cli.Use(sd.Discovery(r))

	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFunc()

	status, body, err := cli.Get(ctx, nil,
		"http://demo.hertz-contrib.test1/ping",
		config.WithSD(true),
	)
	assert.Nil(t, err)
	assert.Equal(t, 200, status)
	assert.Equal(t, "pong1", string(body))
}
