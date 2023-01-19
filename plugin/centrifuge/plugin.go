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
	"fmt"
	"strings"
	"sync"

	"github.com/centrifugal/centrifuge-go"
	"github.com/lastbackend/toolkit"
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
	p.envPrefix = service.Meta().GetEnvPrefix()
	p.service = service.Meta().GetName()
	p.opts.Config = new(centrifuge.Config)
	err := p.Register(service, opts)
	if err != nil {
		return nil
	}
	return p
}

// Register - registers the plugin implements storage using Rabbitmq as a broker service
func (p *plugin) Register(app toolkit.Service, opts *Options) error {
	p.prefix = opts.Name
	if p.prefix == "" {
		p.prefix = defaultPrefix
	}

	p.addFlags(app)

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

func (p *plugin) addFlags(app toolkit.Service) {
	app.CLI().AddStringFlag(p.withPrefix("endpoint"), &p.opts.Endpoint).
		Env(p.generateEnvName("ENDPOINT")).
		Usage("Set connection endpoint").
		Default(defaultWSURL)

	app.CLI().AddStringFlag(p.withPrefix("token"), &p.opts.Token).
		Env(p.generateEnvName("TOKEN")).
		Usage("Set token for a connection authentication.")

	app.CLI().AddStringFlag(p.withPrefix("name"), &p.opts.Name).
		Env(p.generateEnvName("NAME")).
		Usage("Set allows setting client name")

	app.CLI().AddStringFlag(p.withPrefix("version"), &p.opts.Version).
		Env(p.generateEnvName("VERSION")).
		Usage("Set allows setting client version")

	app.CLI().AddDurationFlag(p.withPrefix("read-timeout"), &p.opts.ReadTimeout).
		Env(p.generateEnvName("READ_TIMEOUT")).
		Usage("Set ReadTimeout is how long to wait read operations to complete.. (Default: 5 * time.Second.)")

	app.CLI().AddDurationFlag(p.withPrefix("write-timeout"), &p.opts.WriteTimeout).
		Env(p.generateEnvName("WRITE_TIMEOUT")).
		Usage("Set WriteTimeout is Websocket write timeout. (Default: 1 * time.Second.)")

	app.CLI().AddDurationFlag(p.withPrefix("handshake-timeout"), &p.opts.HandshakeTimeout).
		Env(p.generateEnvName("HANDSHAKE_TIMEOUT")).
		Usage("Set HandshakeTimeout specifies the duration for the handshake to complete. (Default: 1 * time.Second.")

	app.CLI().AddDurationFlag(p.withPrefix("max-server-ping-delay"), &p.opts.MaxServerPingDelay).
		Env(p.generateEnvName("MAX_SERVER_PING_DELAY")).
		Usage("Set MaxServerPingDelay used to set maximum delay of ping from server.. (Default: 10 * time.Second.")
}

func (p *plugin) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", p.prefix, name)
}

func (p *plugin) generateEnvName(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", p.prefix, strings.Replace(name, "-", "_", -1)))
}

func (p *plugin) generateWithEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", p.envPrefix, p.generateEnvName(name)))
}
