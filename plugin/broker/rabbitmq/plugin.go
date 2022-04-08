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
	"context"
	"fmt"
	"github.com/lastbackend/toolkit"
	"github.com/streadway/amqp"
	"strings"
)

const (
	defaultPrefix = "amqp"
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
	prefix string
	opts   brokerOptions

	broker *broker
}

func NewPlugin(app toolkit.Service, opts *Options) Plugin {
	p := new(plugin)
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
	p.broker, err = newBroker(p.opts)
	if err != nil {
		return err
	}
	return p.broker.Connect()
}

func (p *plugin) Stop() error {
	return p.broker.Disconnect()
}

func (p *plugin) Channel() (*amqp.Channel, error) {
	return p.broker.conn.Channel()
}

func (p *plugin) withPrefix(name string) string {
	return fmt.Sprintf("%s-%s", p.prefix, name)
}

func (p *plugin) withEnvPrefix(name string) string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", p.prefix, name))
}

func (p *plugin) addFlags(app toolkit.Service) {
	app.CLI().AddStringFlag(p.withPrefix("endpoint"), &p.opts.Endpoint).
		Env(p.withEnvPrefix("ENDPOINT")).
		Usage("Rabbitmq connection string (Ex: amqp://guest:guest@127.0.0.1:5672)").
		Required()

	app.CLI().AddBoolFlag(p.withPrefix("tls-verify"), &p.opts.TLSVerify).
		Env(p.withEnvPrefix("TLS_VERIFY")).
		Usage("Sets the tls verify")

	app.CLI().AddStringFlag(p.withPrefix("tls-cert"), &p.opts.TLSCert).
		Env(p.withEnvPrefix("TLS_CERT")).
		Usage("Sets the certificate file")

	app.CLI().AddStringFlag(p.withPrefix("tls-key"), &p.opts.TLSKey).
		Env(p.withEnvPrefix("TLS_KEY")).
		Usage("Sets the private key file")
}
