package error

import (
	"errors"
	"fmt"
)

type IngressNotFoundError struct {
	errMsg string
}

func (e IngressNotFoundError) Error() string {
	return e.errMsg
}

func IsIngressNotFoundError(e error) bool {
	var err IngressNotFoundError
	return errors.As(e, &err)
}

func NewIngressNotFoundError(errMsg string) error {
	return IngressNotFoundError{
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

func NewMissIngressAnnotationsError(ann, name, namespace string) error {
	return MissIngressAnnotationsError{
		errMsg: fmt.Sprintf(fmt.Sprintf("miss annotations '%s', ingress '%s' in namespace '%s'", ann, name, namespace)),
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

type InvalidIngressAnnotationsError struct {
	errMsg string
}

func (e InvalidIngressAnnotationsError) Error() string {
	return e.errMsg
}

func IsInvalidIngressAnnotationsError(e error) bool {
	var err InvalidIngressAnnotationsError
	return errors.As(e, &err)
}

func NewInvalidIngressAnnotationsError(ann, name, namespace string) error {
	return InvalidIngressAnnotationsError{
		errMsg: fmt.Sprintf("ingress annotation '%s' contain invalid value, ingress '%s' in namespace '%s'", ann, name, namespace),
	}
}

type AnnotationValidationFailedError struct {
	errMsg string
}

func (e AnnotationValidationFailedError) Error() string {
	return e.errMsg
}

func IsAnnotationValidationFailedError(e error) bool {
	var err AnnotationValidationFailedError
	return errors.As(e, &err)
}

func NewAnnotationValidationFailedError(ann, name, namespace string) error {
	return AnnotationValidationFailedError{
		errMsg: fmt.Sprintf("the value verification of the annotation for '%s' is invalid, ingress '%s' in namespace '%s'", ann, name, namespace),
	}
}

type InvalidSvcPortError struct {
	errMsg string
}

func (e InvalidSvcPortError) Error() string {
	return e.errMsg
}

func IsInvalidSvcPortError(e error) bool {
	var err InvalidSvcPortError
	return errors.As(e, &err)
}

func NewInvalidSvcPortError(svc, name, namespace string) error {
	return InvalidSvcPortError{
		errMsg: fmt.Sprintf("invalid service port, service '%s', ingress '%s' in namespace '%s'", svc, name, namespace),
	}
}
