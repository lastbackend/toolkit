package meta

import (
	"regexp"
	"strings"
)

type Meta struct {
	name        string
	version     string
	description string
	prefix      string
}

func (m *Meta) SetName(name string) *Meta {
	m.name = name
	return m
}

func (m *Meta) GetName() string {
	return m.name
}

func (m *Meta) GetSlug() string {
	slug := regexp.MustCompile(`[^_a-zA-Z0-9 ]+`).ReplaceAllString(m.name, "_")
	return slug
}

func (m *Meta) SetVersion(version string) *Meta {
	m.version = version
	return m
}

func (m *Meta) GetVersion() string {
	return m.version
}

func (m *Meta) SetDescription(description string) *Meta {
	m.description = description
	return m
}

func (m *Meta) GetDescription() string {
	return m.description
}

func (m *Meta) SetEnvPrefix(prefix string) *Meta {
	m.prefix = strings.ToUpper(regexp.MustCompile(`[^_a-zA-Z0-9 ]+`).ReplaceAllString(prefix, ""))
	return m
}

func (m *Meta) GetEnvPrefix() string {
	return m.prefix
}
