package error

import (
	"errors"
	"fmt"
	v1 "k8s.io/api/networking/v1"
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

type InvalidIngressPathError struct {
	errMsg string
}

func (e InvalidIngressPathError) Error() string {
	return e.errMsg
}

func NewInvalidIngressPathError(path, name, namespace string) error {
	return InvalidSvcPortError{
		errMsg: fmt.Sprintf("invalid path '%s', ingress '%s' in namespace '%s'", path, name, namespace),
	}
}

type MissIngressFieldError struct {
	errMsg string
}

func (e MissIngressFieldError) Error() string {
	return e.errMsg
}

func NewMissIngressFieldValueError(field, name, namespace string) error {
	return MissIngressFieldError{
		errMsg: fmt.Sprintf("miss %s value, ingress '%s' in namespace '%s'", field, name, namespace),
	}
}

type NotFoundTlsHostError struct {
	errMsg string
}

func (e NotFoundTlsHostError) Error() string {
	return e.errMsg
}

func NewNotFoundTlsHostError(name, namespace string) error {
	return NotFoundTlsHostError{
		errMsg: fmt.Sprintf("'.spec.host' does not exist in '.spec.tls', ingress '%s' in namespace '%s'", name, namespace),
	}
}

type DuplicateHostError struct {
	errMsg string
}

func (e DuplicateHostError) Error() string {
	return e.errMsg
}

func NewDuplicateHostError(name, namespace string) error {
	return DuplicateHostError{
		errMsg: fmt.Sprintf("'.spec.host' duplicate, ingress '%s' in namespace '%s'", name, namespace),
	}
}

type DuplicatePathError struct {
	errMsg string
}

func (e DuplicatePathError) Error() string {
	return e.errMsg
}

func NewDuplicatePathError(name, namespace string) error {
	return DuplicatePathError{
		errMsg: fmt.Sprintf("'.spec.Host.path' duplicate, ingress '%s' in namespace '%s'", name, namespace),
	}
}

type InvalidIngressValueError struct {
	errMsg string
}

func (e InvalidIngressValueError) Error() string {
	return e.errMsg
}

func NewInvalidIngressValueError(name, namespace string) error {
	return InvalidIngressValueError{
		errMsg: fmt.Sprintf("'.spec.Host.path' contain invalid value, have regular expressions been enabled?, ingress '%s' in namespace '%s'", name, namespace),
	}
}

type SetPathTypeError struct {
	errMsg string
}

func (e SetPathTypeError) Error() string {
	return e.errMsg
}

func NewSetPathTypeError(name, namespace string) error {
	return SetPathTypeError{
		errMsg: fmt.Sprintf("'pathType' field setting error, how to make the path field a regular expression, pathType should be set to: %v, ingress '%s' in namespace '%s'", v1.PathTypeImplementationSpecific, name, namespace),
	}
}

type KubernetesResourcesNotFoundError struct {
	errMsg string
}

func (e KubernetesResourcesNotFoundError) Error() string {
	return e.errMsg
}

func NewKubernetesResourcesNotFoundError(resource, name, namespace string) error {
	return KubernetesResourcesNotFoundError{
		errMsg: fmt.Sprintf("%s '%s' resource not found  not found, namespace '%s'", resource, name, namespace),
	}
}
