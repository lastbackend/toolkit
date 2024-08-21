// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: github.com/calenducky/backend/proto/toolkit/options/annotations.proto

package annotations

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on Plugin with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Plugin) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Plugin with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in PluginMultiError, or nil if none found.
func (m *Plugin) ValidateAll() error {
	return m.validate(true)
}

func (m *Plugin) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Plugin

	// no validation rules for Prefix

	if len(errors) > 0 {
		return PluginMultiError(errors)
	}

	return nil
}

// PluginMultiError is an error wrapping multiple validation errors returned by
// Plugin.ValidateAll() if the designated constraints aren't met.
type PluginMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m PluginMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m PluginMultiError) AllErrors() []error { return m }

// PluginValidationError is the validation error returned by Plugin.Validate if
// the designated constraints aren't met.
type PluginValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e PluginValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e PluginValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e PluginValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e PluginValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e PluginValidationError) ErrorName() string { return "PluginValidationError" }

// Error satisfies the builtin error interface
func (e PluginValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sPlugin.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = PluginValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = PluginValidationError{}

// Validate checks the field values on Service with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Service) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Service with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in ServiceMultiError, or nil if none found.
func (m *Service) ValidateAll() error {
	return m.validate(true)
}

func (m *Service) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Service

	// no validation rules for Package

	if len(errors) > 0 {
		return ServiceMultiError(errors)
	}

	return nil
}

// ServiceMultiError is an error wrapping multiple validation errors returned
// by Service.ValidateAll() if the designated constraints aren't met.
type ServiceMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ServiceMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ServiceMultiError) AllErrors() []error { return m }

// ServiceValidationError is the validation error returned by Service.Validate
// if the designated constraints aren't met.
type ServiceValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ServiceValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ServiceValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ServiceValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ServiceValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ServiceValidationError) ErrorName() string { return "ServiceValidationError" }

// Error satisfies the builtin error interface
func (e ServiceValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sService.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ServiceValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ServiceValidationError{}

// Validate checks the field values on Runtime with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Runtime) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Runtime with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in RuntimeMultiError, or nil if none found.
func (m *Runtime) ValidateAll() error {
	return m.validate(true)
}

func (m *Runtime) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	for idx, item := range m.GetPlugins() {
		_, _ = idx, item

		if all {
			switch v := interface{}(item).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, RuntimeValidationError{
						field:  fmt.Sprintf("Plugins[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, RuntimeValidationError{
						field:  fmt.Sprintf("Plugins[%v]", idx),
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(item).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return RuntimeValidationError{
					field:  fmt.Sprintf("Plugins[%v]", idx),
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	}

	if len(errors) > 0 {
		return RuntimeMultiError(errors)
	}

	return nil
}

// RuntimeMultiError is an error wrapping multiple validation errors returned
// by Runtime.ValidateAll() if the designated constraints aren't met.
type RuntimeMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m RuntimeMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m RuntimeMultiError) AllErrors() []error { return m }

// RuntimeValidationError is the validation error returned by Runtime.Validate
// if the designated constraints aren't met.
type RuntimeValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RuntimeValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RuntimeValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RuntimeValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RuntimeValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RuntimeValidationError) ErrorName() string { return "RuntimeValidationError" }

// Error satisfies the builtin error interface
func (e RuntimeValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRuntime.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RuntimeValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RuntimeValidationError{}

// Validate checks the field values on TestSpec with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *TestSpec) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on TestSpec with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in TestSpecMultiError, or nil
// if none found.
func (m *TestSpec) ValidateAll() error {
	return m.validate(true)
}

func (m *TestSpec) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if all {
		switch v := interface{}(m.GetMockery()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, TestSpecValidationError{
					field:  "Mockery",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, TestSpecValidationError{
					field:  "Mockery",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetMockery()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return TestSpecValidationError{
				field:  "Mockery",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return TestSpecMultiError(errors)
	}

	return nil
}

// TestSpecMultiError is an error wrapping multiple validation errors returned
// by TestSpec.ValidateAll() if the designated constraints aren't met.
type TestSpecMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m TestSpecMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m TestSpecMultiError) AllErrors() []error { return m }

// TestSpecValidationError is the validation error returned by
// TestSpec.Validate if the designated constraints aren't met.
type TestSpecValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e TestSpecValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e TestSpecValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e TestSpecValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e TestSpecValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e TestSpecValidationError) ErrorName() string { return "TestSpecValidationError" }

// Error satisfies the builtin error interface
func (e TestSpecValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sTestSpec.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = TestSpecValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = TestSpecValidationError{}

// Validate checks the field values on MockeryTestsSpec with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *MockeryTestsSpec) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on MockeryTestsSpec with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// MockeryTestsSpecMultiError, or nil if none found.
func (m *MockeryTestsSpec) ValidateAll() error {
	return m.validate(true)
}

func (m *MockeryTestsSpec) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Package

	if len(errors) > 0 {
		return MockeryTestsSpecMultiError(errors)
	}

	return nil
}

// MockeryTestsSpecMultiError is an error wrapping multiple validation errors
// returned by MockeryTestsSpec.ValidateAll() if the designated constraints
// aren't met.
type MockeryTestsSpecMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m MockeryTestsSpecMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m MockeryTestsSpecMultiError) AllErrors() []error { return m }

// MockeryTestsSpecValidationError is the validation error returned by
// MockeryTestsSpec.Validate if the designated constraints aren't met.
type MockeryTestsSpecValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e MockeryTestsSpecValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e MockeryTestsSpecValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e MockeryTestsSpecValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e MockeryTestsSpecValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e MockeryTestsSpecValidationError) ErrorName() string { return "MockeryTestsSpecValidationError" }

// Error satisfies the builtin error interface
func (e MockeryTestsSpecValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sMockeryTestsSpec.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = MockeryTestsSpecValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = MockeryTestsSpecValidationError{}

// Validate checks the field values on Server with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Server) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Server with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in ServerMultiError, or nil if none found.
func (m *Server) ValidateAll() error {
	return m.validate(true)
}

func (m *Server) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	if len(errors) > 0 {
		return ServerMultiError(errors)
	}

	return nil
}

// ServerMultiError is an error wrapping multiple validation errors returned by
// Server.ValidateAll() if the designated constraints aren't met.
type ServerMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m ServerMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m ServerMultiError) AllErrors() []error { return m }

// ServerValidationError is the validation error returned by Server.Validate if
// the designated constraints aren't met.
type ServerValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e ServerValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e ServerValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e ServerValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e ServerValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e ServerValidationError) ErrorName() string { return "ServerValidationError" }

// Error satisfies the builtin error interface
func (e ServerValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sServer.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = ServerValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = ServerValidationError{}

// Validate checks the field values on Route with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *Route) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on Route with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in RouteMultiError, or nil if none found.
func (m *Route) ValidateAll() error {
	return m.validate(true)
}

func (m *Route) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	switch v := m.Server.(type) {
	case *Route_HttpProxy:
		if v == nil {
			err := RouteValidationError{
				field:  "Server",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if all {
			switch v := interface{}(m.GetHttpProxy()).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, RouteValidationError{
						field:  "HttpProxy",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, RouteValidationError{
						field:  "HttpProxy",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(m.GetHttpProxy()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return RouteValidationError{
					field:  "HttpProxy",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *Route_WebsocketProxy:
		if v == nil {
			err := RouteValidationError{
				field:  "Server",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}

		if all {
			switch v := interface{}(m.GetWebsocketProxy()).(type) {
			case interface{ ValidateAll() error }:
				if err := v.ValidateAll(); err != nil {
					errors = append(errors, RouteValidationError{
						field:  "WebsocketProxy",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			case interface{ Validate() error }:
				if err := v.Validate(); err != nil {
					errors = append(errors, RouteValidationError{
						field:  "WebsocketProxy",
						reason: "embedded message failed validation",
						cause:  err,
					})
				}
			}
		} else if v, ok := interface{}(m.GetWebsocketProxy()).(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return RouteValidationError{
					field:  "WebsocketProxy",
					reason: "embedded message failed validation",
					cause:  err,
				}
			}
		}

	case *Route_Websocket:
		if v == nil {
			err := RouteValidationError{
				field:  "Server",
				reason: "oneof value cannot be a typed-nil",
			}
			if !all {
				return err
			}
			errors = append(errors, err)
		}
		// no validation rules for Websocket
	default:
		_ = v // ensures v is used
	}

	if len(errors) > 0 {
		return RouteMultiError(errors)
	}

	return nil
}

// RouteMultiError is an error wrapping multiple validation errors returned by
// Route.ValidateAll() if the designated constraints aren't met.
type RouteMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m RouteMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m RouteMultiError) AllErrors() []error { return m }

// RouteValidationError is the validation error returned by Route.Validate if
// the designated constraints aren't met.
type RouteValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e RouteValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e RouteValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e RouteValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e RouteValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e RouteValidationError) ErrorName() string { return "RouteValidationError" }

// Error satisfies the builtin error interface
func (e RouteValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sRoute.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = RouteValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = RouteValidationError{}

// Validate checks the field values on HttpProxy with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *HttpProxy) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on HttpProxy with the rules defined in
// the proto definition for this message. If any rules are violated, the
// result is a list of violation errors wrapped in HttpProxyMultiError, or nil
// if none found.
func (m *HttpProxy) ValidateAll() error {
	return m.validate(true)
}

func (m *HttpProxy) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Service

	// no validation rules for Method

	if len(errors) > 0 {
		return HttpProxyMultiError(errors)
	}

	return nil
}

// HttpProxyMultiError is an error wrapping multiple validation errors returned
// by HttpProxy.ValidateAll() if the designated constraints aren't met.
type HttpProxyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m HttpProxyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m HttpProxyMultiError) AllErrors() []error { return m }

// HttpProxyValidationError is the validation error returned by
// HttpProxy.Validate if the designated constraints aren't met.
type HttpProxyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e HttpProxyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e HttpProxyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e HttpProxyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e HttpProxyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e HttpProxyValidationError) ErrorName() string { return "HttpProxyValidationError" }

// Error satisfies the builtin error interface
func (e HttpProxyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sHttpProxy.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = HttpProxyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = HttpProxyValidationError{}

// Validate checks the field values on WsProxy with the rules defined in the
// proto definition for this message. If any rules are violated, the first
// error encountered is returned, or nil if there are no violations.
func (m *WsProxy) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on WsProxy with the rules defined in the
// proto definition for this message. If any rules are violated, the result is
// a list of violation errors wrapped in WsProxyMultiError, or nil if none found.
func (m *WsProxy) ValidateAll() error {
	return m.validate(true)
}

func (m *WsProxy) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Service

	// no validation rules for Method

	if len(errors) > 0 {
		return WsProxyMultiError(errors)
	}

	return nil
}

// WsProxyMultiError is an error wrapping multiple validation errors returned
// by WsProxy.ValidateAll() if the designated constraints aren't met.
type WsProxyMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m WsProxyMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m WsProxyMultiError) AllErrors() []error { return m }

// WsProxyValidationError is the validation error returned by WsProxy.Validate
// if the designated constraints aren't met.
type WsProxyValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e WsProxyValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e WsProxyValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e WsProxyValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e WsProxyValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e WsProxyValidationError) ErrorName() string { return "WsProxyValidationError" }

// Error satisfies the builtin error interface
func (e WsProxyValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sWsProxy.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = WsProxyValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = WsProxyValidationError{}
