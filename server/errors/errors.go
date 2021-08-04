package errors

import (
	"github.com/pkg/errors"

	"net/http"
	"strings"
)

const (
	StatusNotFound            = "Not Found"
	StatusBadParameter        = "Bad Parameter"
	StatusInUse               = "In use"
	StatusBadRequest          = "Bad Request"
	StatusUnknown             = "Unknown"
	StatusIncorrectXml        = "Incorrect Xml"
	StatusIncorrectJson       = "Incorrect Json"
	StatusNotUnique           = "Not Unique"
	StatusInternalServerError = "Internal Server Error"
	StatusForbidden           = "Forbidden"
	StatusNotAllowed          = "Not Allowed"
	ArgumentIsEmpty           = "Argument Is Empty"
)

const (
	ErrorInviteNotFound = iota
	ErrorUsernameInUse
	ErrorEmailInUse
	ErrorAccountNotFound
	ErrorDeviceNotFound
	ErrorReleaseNotFound
	ErrorMediaPlanNotFound
	ErrorScheduleNotFound
	ErrorVpnServerNotFound
	ErrorCommercialScheduleNotFound
	ErrorAdvertisingCampaignNotFound
	ErrorMediaFileNotFound
	ErrorTagNotFound
	ErrorRecoveryTokenExpired
	ErrorNotUnique
	ErrorAlreadyAttached
	ErrorNoAttached
	ErrorUnknown
)

type ErrorCode int

var errorMessages = map[ErrorCode]string{
	ErrorInviteNotFound:              "Invite not found",
	ErrorUsernameInUse:               "Username in use",
	ErrorEmailInUse:                  "Email in use",
	ErrorAccountNotFound:             "OwnerId not found",
	ErrorDeviceNotFound:              "Device not found",
	ErrorReleaseNotFound:             "Release not found",
	ErrorScheduleNotFound:            "Schedules not found",
	ErrorVpnServerNotFound:           "Vpn server not found",
	ErrorMediaPlanNotFound:           "Media plan not found",
	ErrorCommercialScheduleNotFound:  "Commercial schedule not found",
	ErrorAdvertisingCampaignNotFound: "AdvertisingCampaign campaign not found",
	ErrorMediaFileNotFound:           "Media file not found",
	ErrorTagNotFound:                 "Tag not found",
	ErrorRecoveryTokenExpired:        "Recovery token expired",
	ErrorNotUnique:                   "Not unique",
	ErrorAlreadyAttached:             "Already attached",
	ErrorNoAttached:                  "Mo attached",
	ErrorUnknown:                     "Unknown error not found",
}

func GetError(code ErrorCode) error {
	if message, ok := errorMessages[code]; ok {
		return errors.New(message)
	} else {
		return errors.New(errorMessages[ErrorUnknown])
	}
}

func GetErrorMessage(code ErrorCode) string {
	if message, ok := errorMessages[code]; ok {
		return message
	} else {
		return errorMessages[ErrorUnknown]
	}
}

type Err struct {
	Code   string
	Attr   string
	origin error
	http   *Http
}

func BadParameter(attr string, e ...error) *Err {
	return &Err{
		Code:   StatusBadParameter,
		Attr:   attr,
		origin: getError(attr+": bad parameter", e...),
		http:   HTTP.getBadParameter(attr),
	}
}

func IncorrectJSON(e ...error) *Err {
	return &Err{
		Code:   StatusIncorrectJson,
		origin: getError("incorrect json", e...),
		http:   HTTP.getIncorrectJSON(),
	}
}

func IncorrectXML(e ...error) *Err {
	return &Err{
		Code:   StatusIncorrectXml,
		origin: getError("incorrect xml", e...),
		http:   HTTP.getIncorrectJSON(),
	}
}

func Forbidden(e ...error) *Err {
	return &Err{
		Code:   StatusForbidden,
		origin: getError("forbidden", e...),
		http:   HTTP.getForbidden(),
	}
}

func NotAllowed(e ...error) *Err {
	return &Err{
		Code:   StatusNotAllowed,
		origin: getError("not allowed", e...),
		http:   HTTP.getNotAllowed(),
	}
}

func Unknown(e ...error) *Err {
	return &Err{
		Code:   StatusUnknown,
		origin: getError("unknown error", e...),
		http:   HTTP.getUnknown(),
	}
}

func (e *Err) Err() error {
	return e.origin
}

func (e *Err) Http(w http.ResponseWriter) {
	e.http.send(w)
}

func (e *Err) SetMessage(s string) *Err {
	e.http.Message = s
	return e
}

type err struct {
	s string
}

func New(text string) *err {
	return &err{text}
}

func (e *err) Error() string {
	return e.s
}

func (e *err) Unauthorized(err ...error) *Err {
	return &Err{
		Code:   http.StatusText(http.StatusUnauthorized),
		origin: getError(joinNameAndMessage(e.s, "access denied"), err...),
		http:   HTTP.getUnauthorized(),
	}
}

func (e *err) NotFound(err ...error) *Err {
	return &Err{
		Code:   http.StatusText(http.StatusNotFound),
		origin: getError(joinNameAndMessage(e.s, "not found"), err...),
		http:   HTTP.getNotFound(e.s),
	}
}

func (e *err) InternalServerError(err ...error) *Err {
	return &Err{
		Code:   http.StatusText(http.StatusInternalServerError),
		origin: getError(joinNameAndMessage(e.s, "internal server error"), err...),
		http:   HTTP.getInternalServerError(e.s),
	}
}

func (e *err) NotUnique(attr string, err ...error) *Err {
	return &Err{
		Code:   StatusNotUnique,
		origin: getError(joinNameAndMessage(e.s, strings.ToLower(attr)+" not unique"), err...),
		http:   HTTP.getNotUnique(strings.ToLower(attr)),
	}
}

func (e *err) Allocated(attr string, err ...error) *Err {
	return &Err{
		Code:   StatusInUse,
		Attr:   attr,
		origin: getError(joinNameAndMessage(e.s, strings.ToLower(attr))+" is in use", err...),
		http:   HTTP.getAllocatedParameter(strings.ToLower(attr)),
	}
}

func (e *err) BadParameter(attr string, err ...error) *Err {
	return &Err{
		Code:   StatusBadParameter,
		Attr:   attr,
		origin: getError(joinNameAndMessage(e.s, "bad parameter "+strings.ToLower(attr)), err...),
		http:   HTTP.getBadParameter(attr),
	}
}

func (e *err) BadRequest(msg string, err ...error) *Err {
	return &Err{
		Code:   StatusBadParameter,
		origin: getError(joinNameAndMessage(e.s, msg), err...),
		http:   HTTP.getBadRequest(msg),
	}
}

func (e *err) IncorrectJSON(err ...error) *Err {
	return &Err{
		Code:   StatusIncorrectJson,
		origin: getError(joinNameAndMessage(e.s, "incorrect json"), err...),
		http:   HTTP.getIncorrectJSON(),
	}
}

func (e *err) IncorrectXML(err ...error) *Err {
	return &Err{
		Code:   StatusIncorrectJson,
		origin: getError(joinNameAndMessage(e.s, "incorrect xml"), err...),
		http:   HTTP.getIncorrectXML(),
	}
}

func (e *err) Forbidden(err ...error) *Err {
	return &Err{
		Code:   StatusForbidden,
		origin: getError(joinNameAndMessage(e.s, "forbidden"), err...),
		http:   HTTP.getForbidden(),
	}
}

func (e *err) NotAllowed(err ...error) *Err {
	return &Err{
		Code:   StatusNotAllowed,
		origin: getError(joinNameAndMessage(e.s, "not allowed"), err...),
		http:   HTTP.getNotAllowed(),
	}
}

func (e *err) Unknown(err ...error) *Err {
	return &Err{
		Code:   StatusUnknown,
		origin: getError(joinNameAndMessage(e.s, "unknown error"), err...),
		http:   HTTP.getUnknown(),
	}
}

func getError(msg string, err ...error) error {
	if len(err) == 0 {
		return errors.New(msg)
	} else {
		return err[0]
	}
}

func joinNameAndMessage(name, message string) string {
	return toUpperFirstChar(name) + ": " + message
}

func toUpperFirstChar(srt string) string {
	return strings.ToUpper(srt[0:1]) + srt[1:]
}
