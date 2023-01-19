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

package rabbitmq

import (
	"github.com/lastbackend/toolkit"
	"github.com/streadway/amqp"
	"sync"

	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultPrefix = "amqp"
	defaultPort   = 5672
)

const (
	envHostName     = "HOST"
	envPortName     = "PORT"
	envUserName     = "USERNAME"
	envPasswordName = "PASSWORD"
)

type Plugin interface {
	toolkit.Plugin

	Publish(event string, payload []byte, opts *PublishOptions) error
	Subscribe(service, event string, handler Handler, opts *SubscribeOptions) (Subscriber, error)
	Channel() (*amqp.Channel, error)

	Register(app toolkit.Service, opts *Options) error
}

type Options struct {
	Name string
}

type plugin struct {
	sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc

	prefix    string
	envPrefix string
	service   string

	opts   brokerOptions
	broker *broker
	probe  toolkit.Probe

	subscribers map[string]bool
}

func NewPlugin(service toolkit.Service, opts *Options) Plugin {
	p := new(plugin)
	p.envPrefix = service.Meta().GetEnvPrefix()
	p.service = service.Meta().GetName()
	p.probe = service.Probe()
	p.subscribers = make(map[string]bool, 0)
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
		config := p.getAMQPConfig()
		if config.Host == "" {
			return fmt.Errorf("%s flag or %s environment variable required but not set",
				p.withPrefix("endpoint"), p.generateWithEnvPrefix(envHostName))
		}
		p.opts.Endpoint = config.getConnectionString()
	}

	p.opts.DefaultExchange = &Exchange{
		Name:    p.service,
		Durable: true,
	}

	p.broker = newBroker(p.opts)

	if err := p.broker.Connect(); err != nil {
		return err
	}

	p.probe.AddReadinessFunc(p.prefix, func() error {
		return p.broker.Connected()
	})

	return nil
}

func (p *plugin) Stop() error {
	return p.broker.Disconnect()
}

func (p *plugin) Publish(event string, payload []byte, opts *PublishOptions) error {
	return p.broker.Publish(p.service, event, payload, opts)
}

type Handler func(ctx context.Context, payload []byte)

func (p *plugin) Subscribe(service, event string, handler Handler, opts *SubscribeOptions) (Subscriber, error) {
	queue := fmt.Sprintf("%s:events", p.service)
	key := fmt.Sprintf("%s:%s", queue, event)

	p.RLock()
	_, exists := p.subscribers[key]
	p.RUnlock()

	if exists {
		return nil, fmt.Errorf("handler already set for event: %s", event)
	}

	p.RLock()
	p.subscribers[key] = true
	p.RUnlock()

	fn := func(ctx context.Context, name string, data []byte) {
		if event != name {
			return
		}
		handler(ctx, data)
	}

	if opts == nil {
		opts = new(SubscribeOptions)
	}

	return p.broker.Subscribe(service, queue, fn, opts)
}

func (p *plugin) Channel() (*amqp.Channel, error) {
	return p.broker.Channel()
}

func (p *plugin) genUsage() string {
	return fmt.Sprintf(`Rabbitmq connection string (Ex: amqp://guest:guest@127.0.0.1:5672)
or use environment variables: 
	%s - The host to connect to (required), 
	%s - The port to bind to (default: 5672), 
	%s - The username to connect with (not required, guest by default), 
	%s - The password to connect with (not required, guest by default)`,
		p.generateWithEnvPrefix(envHostName), p.generateWithEnvPrefix(envPortName),
		p.generateWithEnvPrefix(envUserName), p.generateWithEnvPrefix(envPasswordName))
}

func (p *plugin) addFlags(app toolkit.Service) {
	app.CLI().AddStringFlag(p.withPrefix("endpoint"), &p.opts.Endpoint).
		Env(p.generateEnvName("ENDPOINT")).
		Usage(p.genUsage())

	app.CLI().AddBoolFlag(p.withPrefix("tls-verify"), &p.opts.TLSVerify).
		Env(p.generateEnvName("TLS_VERIFY")).
		Usage("Sets the tls verify")

	app.CLI().AddStringFlag(p.withPrefix("tls-cert"), &p.opts.TLSCert).
		Env(p.generateEnvName("TLS_CERT")).
		Usage("Sets the certificate file")

	app.CLI().AddStringFlag(p.withPrefix("tls-key"), &p.opts.TLSKey).
		Env(p.generateEnvName("TLS_KEY")).
		Usage("Sets the private key file")

	app.CLI().AddIntFlag(p.withPrefix("prefetch-count"), &p.opts.PrefetchCount).
		Env(p.generateEnvName("PREFETCH_COUNT")).
		Default(DefaultPrefetchCount).
		Usage("Sets the prefetch count")

	app.CLI().AddBoolFlag(p.withPrefix("prefetch-global"), &p.opts.PrefetchGlobal).
		Env(p.generateEnvName("PREFETCH_GLOBAL")).
		Usage("Sets the prefetch global")
}

type amqpConfig struct {
	Host     string
	Port     int32
	Username string
	Password string
}

func (c *amqpConfig) getConnectionString() string {
	if c.Username == "" {
		c.Username = "guest"
	}
	if c.Password == "" {
		c.Password = "guest"
	}
	return fmt.Sprintf("amqp://%s:%s@%s:%delivery", c.Username, c.Password, c.Host, c.Port)
}

func (p *plugin) getAMQPConfig() amqpConfig {
	config := amqpConfig{Port: defaultPort}

	if host, ok := os.LookupEnv(p.generateWithEnvPrefix(envHostName)); ok {
		config.Host = host
	}
	if port, ok := os.LookupEnv(p.generateWithEnvPrefix(envPortName)); ok {
		if value, err := strconv.ParseInt(port, 10, 32); err == nil {
			config.Port = int32(value)
		}
	}
	if user, ok := os.LookupEnv(p.generateWithEnvPrefix(envUserName)); ok {
		config.Username = user
	}
	if password, ok := os.LookupEnv(p.generateWithEnvPrefix(envPasswordName)); ok {
		config.Password = password
	}

	return config
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
