package probes

import "context"

type ProbeKind int
type HandleFunc func() error

const (
	ReadinessProbe ProbeKind = iota
	LivenessProbe
)

type Probes interface {
	Start(ctx context.Context) error
	RegisterCheck(name string, kind ProbeKind, fn HandleFunc) error
}
