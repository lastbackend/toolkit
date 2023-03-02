package runtime

const (
	MetaOptionEnvPrefix   = "envPrefix"
	MetaOptionVersion     = "version"
	MetaOptionDescription = "description"
)

type Option interface {
	Name() string
	Value() string
}

type Options []Option

type option struct {
	name  string
	value string
}

func (o *option) Name() string {
	return o.name
}

func (o *option) Value() string {
	return o.value
}

func WithEnvPrefix(prefix string) Option {
	return &option{name: MetaOptionEnvPrefix, value: prefix}
}

func WithVersion(version string) Option {
	return &option{name: MetaOptionVersion, value: version}
}

func WithDescription(description string) Option {
	return &option{name: MetaOptionDescription, value: description}
}
