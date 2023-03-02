/*
Copyright [2014] - [2023] The Last.Backend authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package centrifuge

import (
	"context"
	"github.com/centrifugal/centrifuge-go"
	"github.com/lastbackend/toolkit"
	"github.com/lastbackend/toolkit/pkg/config"
	"sync"
	"time"
)

const (
	defaultPrefix = "centrifuge"
	defaultWSURL  = "ws://localhost:8000/connection/websocket"
)

type Plugin interface {
	toolkit.Plugin

	Client() *centrifuge.Client

	Register(app toolkit.Service, opts *Options) error
}

type Options struct {
	Name string
}

type Config struct {
	Endpoint           string        `env:"ENDPOINT" comment:"Set connection endpoint (ws://localhost:8000/connection/websocket)"`
	Token              string        `env:"TOKEN" comment:"Set token for a connection authentication."`
	Name               string        `env:"NAME" comment:"Set allows setting client name"`
	Version            string        `env:"VERSION" comment:"Set allows setting client version"`
	ReadTimeout        time.Duration `env:"READ_TIMEOUT" comment:"Set ReadTimeout is how long to wait read operations to complete.. (Default: 5 * time.Second.)"`
	WriteTimeout       time.Duration `env:"WRITE_TIMEOUT" comment:"Set WriteTimeout is Websocket write timeout. (Default: 1 * time.Second.)"`
	HandshakeTimeout   time.Duration `env:"HANDSHAKE_TIMEOUT" comment:"Set HandshakeTimeout specifies the duration for the handshake to complete. (Default: 1 * time.Second."`
	MaxServerPingDelay time.Duration `env:"MAX_SERVER_PING_DELAY" comment:"Set MaxServerPingDelay used to set maximum delay of ping from server.. (Default: 10 * time.Second."`
}

type options struct {
	*centrifuge.Config

	// Endpoint for connection to public WS server
	Endpoint string
}

type plugin struct {
	sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	prefix    string
	envPrefix string
	service   string

	opts options

	client *centrifuge.Client
}

func NewPlugin(service toolkit.Service, opts *Options) Plugin {
	p := new(plugin)

	p.prefix = opts.Name
	if p.prefix == "" {
		p.prefix = defaultPrefix
	}

	if err := config.Parse(&p.opts, p.prefix); err != nil {
		return nil
	}

	err := p.Register(service, opts)
	if err != nil {
		return nil
	}
	return p
}

// Register - registers the plugin implements message distribution using Centrifuge as a broker service
func (p *plugin) Register(app toolkit.Service, _ *Options) error {

	if err := app.PluginRegister(p); err != nil {
		return err
	}

	return nil
}

func (p *plugin) Start(ctx context.Context) error {

	p.ctx, p.cancel = context.WithCancel(ctx)

	if p.opts.Endpoint == "" {
		p.opts.Endpoint = defaultWSURL
	}

	p.client = centrifuge.NewJsonClient(p.opts.Endpoint, *p.opts.Config)

	return p.client.Connect()
}

func (p *plugin) Client() *centrifuge.Client {
	return p.client
}

func (p *plugin) Stop() error {
	p.client.Close()
	return nil
}
