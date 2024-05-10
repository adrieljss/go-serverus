package result

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// err is nil if the result is Ok
type Result struct {
	Err      error  `json:"-"`
	Success  bool   `json:"success"`
	Status   int16  `json:"status"`
	Type     string `json:"type,omitempty"`
	Message  string `json:"message,omitempty"`
	Content  any    `json:"content,omitempty"`
	Metadata any    `json:"metadata"`
}

type Error = Result

func (res *Result) ok() bool {
	return res.Err == nil
}

func (res *Result) SendJSON(ctx *gin.Context) {
	ctx.JSON(int(res.Status), res)
}

func Ok(status int16, content any) *Result {
	return &Result{
		Err:     nil,
		Success: true,
		Status:  status,
		Content: content,
	}
}

func Err(status int16, err error, errType string, message string) *Result {
	return &Result{
		Err:     err,
		Success: false,
		Status:  status,
		Type:    errType,
		Message: message,
	}
}

func ErrWithMetadata(status int16, err error, errType string, message string, metadata map[string]string) *Result {
	return &Result{
		Err:      err,
		Success:  false,
		Status:   status,
		Type:     errType,
		Message:  message,
		Metadata: metadata,
	}
}

// only accepts error validation, returns status 400
func ErrValidate(errvalidate error) *Result {
	return &Result{
		Err:      errvalidate,
		Success:  false,
		Status:   400,
		Type:     "BAD_REQUEST",
		Metadata: errvalidate,
	}
}

func ErrBodyBind() *Result {
	return Err(400, nil, "INVALID_BODY", "invalid json body")
}

// also logs server error with
//
//	logrus.WithError(err).Error("something went wrong")
//
// so no need to do logging
func ServerErr(err error) *Result {
	logrus.WithError(err).Error("something went wrong")
	return Err(500, err, "UNKNOWN_ERROR", "something went wrong with the server")
}
