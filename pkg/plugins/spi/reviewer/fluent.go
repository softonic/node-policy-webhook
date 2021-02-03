package reviewer

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	admission_api "k8s.io/api/admission/v1"
	"k8s.io/client-go/dynamic"

	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/k8s"
)

type responseType int8

const (
	proceed   responseType = iota
	allowed   responseType = iota
	failure   responseType = iota
	jsonpatch responseType = iota
)

type (
	BaseStage struct {
		Error error
		responseType

		Logger logr.Logger
		*k8s.Interface
	}

	GivenStage struct {
		BaseStage
		*admission_api.AdmissionRequest
	}

	RequestedObjectStage struct {
		*GivenStage
	}

	PatcherStage struct {
		*GivenStage
	}

	WhenStage struct {
		*GivenStage
		Patcher
		Patch []PatchOperation
	}

	ThenStage struct {
		*WhenStage
	}

	EndStage struct {
		*ThenStage
	}
)

/**
 * Base
 */

func (s *BaseStage) setResponseType(cause error, value responseType) *BaseStage {
	if cause != nil {
		s.Error = cause
	}
	s.responseType = value
	return s
}

func (s *BaseStage) CanContinue() bool {
	return s.responseType == proceed
}

func (s *BaseStage) Allow(cause error) *BaseStage {
	return s.setResponseType(cause, allowed)
}

func (s *BaseStage) Fail(cause error) *BaseStage {
	return s.setResponseType(cause, failure)
}

func (s *BaseStage) JsonPatch() *BaseStage {
	return s.setResponseType(nil, jsonpatch)
}

/**
 * Given
 */

func Given(logger logr.Logger, itf dynamic.Interface) *GivenStage {
	return &GivenStage{
		BaseStage: BaseStage{
			Logger:    logger,
			Interface: k8s.NewInterface(itf),
		},
	}
}

func (g *GivenStage) Request(request *admission_api.AdmissionRequest) *GivenStage {
	g.AdmissionRequest = request
	g.Logger = g.Logger.WithValues("kind", g.AdmissionRequest.RequestKind)
	return g
}

func (g *GivenStage) An() *GivenStage {
	return g
}

func (g *GivenStage) A() *GivenStage {
	return g
}

func (g *GivenStage) The() *GivenStage {
	return g
}

func (g *GivenStage) Or() *GivenStage {
	return g
}

func (g *GivenStage) And() *GivenStage {
	return g
}

func (g *GivenStage) Group() *GivenStage {
	return g
}

func (g *GivenStage) End() *GivenStage {
	return g
}

func (g *GivenStage) RequestedObject() *RequestedObjectStage {
	return &RequestedObjectStage{g}
}

func (s *RequestedObjectStage) NamespaceIsNot(name string) *RequestedObjectStage {
	s.Logger = s.Logger.WithValues("namespace", s.AdmissionRequest.Namespace)
	if s.AdmissionRequest.Namespace == name {
		s.Allow(errors.Errorf("namespace is %s", name))
	}
	return s
}

func (s *RequestedObjectStage) IsNotNull() *RequestedObjectStage {
	if s.AdmissionRequest.Object.Raw == nil {
		s.Allow(errors.New("Request object raw is nil"))
	}
	return s
}

func (s *RequestedObjectStage) The() *RequestedObjectStage {
	return s
}

func (s *RequestedObjectStage) And() *RequestedObjectStage {
	return s
}

func (s *RequestedObjectStage) Or() *RequestedObjectStage {
	return s
}

func (s *RequestedObjectStage) End() *GivenStage {
	return s.GivenStage
}

/**
 * When stage
 */

func (g *GivenStage) When(hook Hook) *WhenStage {
	return hook.Review(g)
}

func (w *WhenStage) I() *WhenStage {
	if !w.CanContinue() {
		return w
	}
	return w
}

func (w *WhenStage) PatchTheRequest() *WhenStage {
	if !w.CanContinue() {
		return w
	}
	if !w.CanContinue() {
		return w
	}
	w.Patch, w.Error = w.Patcher.Create()
	if w.Error != nil {
		w.Fail(errors.WithMessage(w.Error, "Can't patch object"))
		return w
	}
	w.JsonPatch()
	return w
}

func (w *WhenStage) Then() *ThenStage {
	return &ThenStage{
		WhenStage: w,
	}
}

/**
 * Then stage
 */

func (t *ThenStage) I() *ThenStage {
	return t
}

func (t *ThenStage) Can() *ThenStage {
	return t
}

func (t *ThenStage) ReturnThePatch() *ThenStage {
	return t
}

func (t *ThenStage) OrElse() *ThenStage {
	return t
}

func (t *ThenStage) ReturnTheStatus() *ThenStage {
	return t
}

/**
 * End
 */
func (t *ThenStage) End() *EndStage {
	return &EndStage{t}
}

func (e *EndStage) Response() *admission_api.AdmissionResponse {
	switch e.responseType {
	case allowed:
		e.Logger.Info("allowing")
		return newAllowedResponse(e.AdmissionRequest, e.Error)
	case failure:
		e.Logger.Error(e.Error, "denying")
		return newFailureResponse(e.AdmissionRequest, e.Error)
	case jsonpatch:
		e.Logger.WithValues("patch", e.Patch).Info("patching")
		return newJSONPatchResponse(e.AdmissionRequest, e.Patch)
	}
	return newAllowedResponse(e.AdmissionRequest, errors.New("should never reach this code"))
}
