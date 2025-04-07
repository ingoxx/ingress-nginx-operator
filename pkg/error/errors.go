package error

import "errors"

type IngressNotError struct {
	errMsg string
}

func (e IngressNotError) Error() string {
	return e.errMsg
}

func IsIngressNotError(e error) bool {
	var err IngressNotError
	return errors.As(e, &err)
}

func NewIngressNotError(errMsg string) error {
	return IngressNotError{
		errMsg: errMsg,
	}
}

type BackendNotFoundError struct {
	errMsg string
}

func (e BackendNotFoundError) Error() string {
	return e.errMsg
}

func IsBackendNotFoundError(e error) bool {
	var err BackendNotFoundError
	return errors.As(e, &err)
}

func NewBackendNotFoundError(errMsg string) error {
	return BackendNotFoundError{
		errMsg: errMsg,
	}
}

// CtrNotFoundError ingress的annotations或ingressClass有没有选择当前控制器
type CtrNotFoundError struct {
	errMsg string
}

func (e CtrNotFoundError) Error() string {
	return e.errMsg
}

func IsCtrNotFoundError(e error) bool {
	var err CtrNotFoundError
	return errors.As(e, &err)
}

func NewCtrNotFoundError(errMsg string) error {
	return CtrNotFoundError{
		errMsg: errMsg,
	}
}

type MissIngressAnnotationsError struct {
	errMsg string
}

func (e MissIngressAnnotationsError) Error() string {
	return e.errMsg
}

func IsMissIngressAnnotationsError(e error) bool {
	var err MissIngressAnnotationsError
	return errors.As(e, &err)
}

func NewMissIngressAnnotationsError(errMsg string) error {
	return MissIngressAnnotationsError{
		errMsg: errMsg,
	}
}

type IngressAnnotationsContentError struct {
	errMsg string
}

func (e IngressAnnotationsContentError) Error() string {
	return e.errMsg
}

func IsIngressAnnotationsContentError(e error) bool {
	var err IngressAnnotationsContentError
	return errors.As(e, &err)
}

func NewIngressAnnotationsContentError(errMsg string) error {
	return IngressAnnotationsContentError{
		errMsg: errMsg,
	}
}
