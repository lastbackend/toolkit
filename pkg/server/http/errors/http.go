package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var GrpcErrorHandlerFunc = HTTP.ParseGrpcError
var HTTP Http

type Http struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (Http) Unauthorized(w http.ResponseWriter, msg ...string) {
	HTTP.getUnauthorized(msg...).send(w)
}

func (Http) Forbidden(w http.ResponseWriter, msg ...string) {
	HTTP.getForbidden(msg...).send(w)
}

func (Http) NotAllowed(w http.ResponseWriter, msg ...string) {
	HTTP.getNotAllowed(msg...).send(w)
}

func (Http) BadRequest(w http.ResponseWriter, msg ...string) {
	HTTP.getBadRequest(msg...).send(w)
}

func (Http) NotFound(w http.ResponseWriter, args ...string) {
	HTTP.getNotFound(args...).send(w)
}

func (Http) InternalServerError(w http.ResponseWriter, msg ...string) {
	HTTP.getInternalServerError(msg...).send(w)
}

func (Http) BadGateway(w http.ResponseWriter) {
	HTTP.getBadGateway().send(w)
}

func (Http) PaymentRequired(w http.ResponseWriter, msg ...string) {
	HTTP.getPaymentRequired(msg...).send(w)
}

func (Http) NotImplemented(w http.ResponseWriter, msg ...string) {
	HTTP.getPaymentRequired(msg...).send(w)
}

func (Http) BadParameter(w http.ResponseWriter, args ...string) {
	HTTP.getBadParameter(args...).send(w)
}

func (Http) InvalidJSON(w http.ResponseWriter, msg ...string) {
	HTTP.getIncorrectJSON(msg...).send(w)
}

func (Http) InvalidXML(w http.ResponseWriter, msg ...string) {
	HTTP.getIncorrectXML(msg...).send(w)
}

func (h Http) send(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(h.Code)
	response, _ := json.Marshal(h)
	w.Write(response)
}

// ===================================================================================================================
// ============================================= INTERNAL HELPER METHODS =============================================
// ===================================================================================================================

func (Http) getUnauthorized(msg ...string) *Http {
	return getHttpError(http.StatusUnauthorized, msg...)
}

func (Http) getForbidden(msg ...string) *Http {
	return getHttpError(http.StatusForbidden, msg...)
}

func (Http) getNotAllowed(msg ...string) *Http {
	return getHttpError(http.StatusMethodNotAllowed, msg...)
}

func (Http) getPaymentRequired(msg ...string) *Http {
	return getHttpError(http.StatusPaymentRequired, msg...)
}

func (Http) getUnknown(msg ...string) *Http {
	return getHttpError(http.StatusInternalServerError, msg...)
}

func (Http) getInternalServerError(msg ...string) *Http {
	return getHttpError(http.StatusInternalServerError, msg...)
}

func (Http) getBadGateway() *Http {
	return getHttpError(http.StatusBadGateway)
}

func (Http) getNotImplemented(msg ...string) *Http {
	return getHttpError(http.StatusNotImplemented, msg...)
}

func (Http) getBadRequest(msg ...string) *Http {
	return getHttpError(http.StatusBadRequest, msg...)
}

func (Http) getNotFound(args ...string) *Http {
	message := "Not Found"
	for i, a := range args {
		switch i {
		case 0:
			message = fmt.Sprintf("%s not found", toUpperFirstChar(a))
		default:
			panic("Wrong parameter count: (is allowed from 0 to 1)")
		}
	}
	return &Http{
		Code:    http.StatusNotFound,
		Status:  http.StatusText(http.StatusNotFound),
		Message: message,
	}
}

func (Http) getNotUnique(name string) *Http {
	return &Http{
		Code:    http.StatusBadRequest,
		Status:  StatusNotUnique,
		Message: fmt.Sprintf("%s is already in use", toUpperFirstChar(name)),
	}
}

func (Http) getIncorrectJSON(msg ...string) *Http {
	message := "Incorrect json"
	for i, m := range msg {
		switch i {
		case 0:
			message = m
		default:
			panic("Wrong parameter count: (is allowed from 0 to 1)")
		}
	}
	return &Http{
		Code:    http.StatusBadRequest,
		Status:  StatusIncorrectJson,
		Message: message,
	}
}

func (Http) getIncorrectXML(msg ...string) *Http {
	message := "Incorrect json"
	for i, m := range msg {
		switch i {
		case 0:
			message = m
		default:
			panic("Wrong parameter count: (is allowed from 0 to 1)")
		}
	}
	return &Http{
		Code:    http.StatusBadRequest,
		Status:  StatusIncorrectXml,
		Message: message,
	}
}

func (Http) getAllocatedParameter(args ...string) *Http {
	message := "Value is in use"
	for i, a := range args {
		switch i {
		case 0:
			message = fmt.Sprintf("%s is already in use", toUpperFirstChar(a))
		default:
			panic("Wrong parameter count: (is allowed from 0 to 1)")
		}
	}
	return &Http{
		Code:    http.StatusBadRequest,
		Status:  StatusBadParameter,
		Message: message,
	}
}

func (Http) getBadParameter(args ...string) *Http {
	message := "Bad parameter"
	for i, a := range args {
		switch i {
		case 0:
			message = fmt.Sprintf("Bad %s parameter", a)
		default:
			panic("Wrong parameter count: (is allowed from 0 to 1)")
		}
	}
	return &Http{
		Code:    http.StatusBadRequest,
		Status:  StatusBadParameter,
		Message: message,
	}
}

func getHttpError(code int, msg ...string) *Http {
	httpStatus := http.StatusText(code)
	httpMessage := http.StatusText(code)

	for i, m := range msg {
		switch i {
		case 0:
			httpStatus = m
		default:
			panic("Wrong parameter count: (is allowed from 0 to 1)")
		}
	}
	return &Http{
		Code:    code,
		Status:  httpMessage,
		Message: httpStatus,
	}
}

func (h Http) ParseGrpcError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)
	if !ok {
		h.InternalServerError(w, err.Error())
		return
	}
	getHttpError(httpStatusFromCode(st.Code()), st.Message()).send(w)
}

func httpStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return http.StatusRequestTimeout
	case codes.Unknown:
		return http.StatusInternalServerError
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		// Note, this deliberately
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	}
	return http.StatusInternalServerError
}
