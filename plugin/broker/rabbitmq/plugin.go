/*
Copyright [2014] - [2021] The Last.Backend authors.

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

	Channel() (*amqp.Channel, error)
	Register(app toolkit.Service, opts *Options) error
}

type Options struct {
	Name string
}

type plugin struct {
	prefix    string
	envPrefix string
	opts      brokerOptions

	broker *broker
}

func NewPlugin(app toolkit.Service, opts *Options) Plugin {
	p := new(plugin)
	p.envPrefix = app.Meta().GetEnvPrefix()
	err := p.Register(app, opts)
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

func (p *plugin) Start(context.Context) (err error) {

	if p.opts.Endpoint == "" {
		config := p.getAMQPConfig()
		if config.Host == "" {
			return fmt.Errorf("%s flag or %s environment variable required but not set",
				p.withPrefix("endpoint"), p.generatetWithEnvPrefix(envHostName))
		}
		p.opts.Endpoint = config.getConnectionString()
	}

	p.broker, err = newBroker(p.opts)
	if err != nil {
		return err
	}

	return p.broker.Connect()
}

func (p *plugin) Stop() error {
	return p.broker.Disconnect()
}

// Call - send message with delivery guarantee
func (p *plugin) Call() error {
	return nil
}

// Cast - send message without delivery guarantee
func (p *plugin) Cast() error {
	return nil
}

func (p *plugin) Publish() error {
	return nil
}

func (p *plugin) Subscribe() error {
	return nil
}

func (p *plugin) Channel() (*amqp.Channel, error) {
	return p.broker.conn.Channel()
}

func (p *plugin) genUsage() string {
	return fmt.Sprintf(`Rabbitmq connection string (Ex: amqp://guest:guest@127.0.0.1:5672)
or use environment variables: 
	%s - The host to connect to (required), 
	%s - The port to bind to (default: 5672), 
	%s - The username to connect with (not required), 
	%s - The password to connect with (not required)`,
		p.generatetWithEnvPrefix(envHostName), p.generatetWithEnvPrefix(envPortName),
		p.generatetWithEnvPrefix(envUserName), p.generatetWithEnvPrefix(envPasswordName))
}

func (p *plugin) addFlags(app toolkit.Service) {
	app.CLI().AddStringFlag(p.withPrefix("endpoint"), &p.opts.Endpoint).
		Env(p.generateEnvName("ENDPOINT")).
		Usage(p.genUsage()).
		Required()

	app.CLI().AddBoolFlag(p.withPrefix("tls-verify"), &p.opts.TLSVerify).
		Env(p.generateEnvName("TLS_VERIFY")).
		Usage("Sets the tls verify")

	app.CLI().AddStringFlag(p.withPrefix("tls-cert"), &p.opts.TLSCert).
		Env(p.generateEnvName("TLS_CERT")).
		Usage("Sets the certificate file")

	app.CLI().AddStringFlag(p.withPrefix("tls-key"), &p.opts.TLSKey).
		Env(p.generateEnvName("TLS_KEY")).
		Usage("Sets the private key file")
}

type amqpConfig struct {
	Host     string
	Port     int32
	Username string
	Password string
}

func (c *amqpConfig) getConnectionString() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d",
		c.Username, c.Password, c.Host, c.Port)
}

func (p *plugin) getAMQPConfig() amqpConfig {
	config := amqpConfig{Port: defaultPort}

	if host, ok := os.LookupEnv(p.generatetWithEnvPrefix(envHostName)); ok {
		config.Host = host
	}
	if port, ok := os.LookupEnv(p.generatetWithEnvPrefix(envPortName)); ok {
		if value, err := strconv.ParseInt(port, 10, 32); err == nil {
			config.Port = int32(value)
		}
	}
	if user, ok := os.LookupEnv(p.generatetWithEnvPrefix(envUserName)); ok {
		config.Username = user
	}
	if password, ok := os.LookupEnv(p.generatetWithEnvPrefix(envPasswordName)); ok {
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

func (p *plugin) generatetWithEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", p.envPrefix, p.generateEnvName(name)))
}
