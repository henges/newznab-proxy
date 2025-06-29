package newznab

import (
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/gorilla/schema"
	"github.com/henges/newznab-proxy/xmlutil"
)

type Server struct {
	impl              ServerImplementation
	validateAPIKey    bool
	getAllowedAPIKeys func() ([]string, error)
	middlewares       []Middleware
}

type Middleware func(handler http.Handler) http.Handler

type serverOptions struct {
	middlewares       []Middleware
	getAllowedAPIKeys func() ([]string, error)
}

type ServerOption func(options *serverOptions)

func WithAPIKeyValidation(keyProvider func() ([]string, error)) ServerOption {
	return func(options *serverOptions) {
		options.getAllowedAPIKeys = keyProvider
	}
}

func WithMiddleware(m ...Middleware) ServerOption {
	return func(options *serverOptions) {
		options.middlewares = m
	}
}

func NewServer(impl ServerImplementation, opts ...ServerOption) *Server {
	options := &serverOptions{}
	for _, o := range opts {
		o(options)
	}

	ret := &Server{
		impl:              impl,
		validateAPIKey:    options.getAllowedAPIKeys != nil,
		getAllowedAPIKeys: options.getAllowedAPIKeys,
		middlewares:       options.middlewares,
	}
	return ret
}

func (s *Server) Handler() http.Handler {

	m := http.NewServeMux()
	return s.HandlerWithMux(m)
}

func (s *Server) HandlerWithMux(m *http.ServeMux) http.Handler {

	m.Handle("GET /api", http.HandlerFunc(s.apiHandler))
	m.Handle("GET /getnzb/{id}", http.HandlerFunc(s.getNZB))
	var ret http.Handler = m
	for _, middle := range s.middlewares {
		ret = middle(ret)
	}
	m.Handle("GET /healthz", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	return ret
}

func (s *Server) apiHandler(rw http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		respondError(rw, http.StatusBadRequest, err)
		return
	}
	reqType := r.Form.Get("t")
	if reqType == "" {
		respondErrorString(rw, http.StatusBadRequest, "t parameter must be provided")
		return
	}
	apiKey := r.Form.Get("apikey")
	if s.validateAPIKey {
		keys, err := s.getAllowedAPIKeys()
		if err != nil {
			respondError(rw, http.StatusInternalServerError, err)
			return
		}
		if !slices.Contains(keys, apiKey) {
			respondErrorString(rw, http.StatusUnauthorized, "Unauthorized")
			return
		}
	}

	// Rest of the implementation is delegated to handler funcs
	switch reqType {
	case "search":
		s.search(rw, r)
	default:
		respondErrorString(rw, http.StatusNotImplemented, fmt.Sprint("method", reqType, "not implemented"))
	}
}

func (s *Server) getNZB(rw http.ResponseWriter, r *http.Request) {

	value := r.PathValue("id")
	if value == "" {
		respondErrorString(rw, http.StatusBadRequest, "an NZB id must be provided")
		return
	}

	nzb, err := s.impl.GetNZB(r.Context(), value)
	if err != nil {
		var srvErr ServerError
		if errors.As(err, &srvErr) {
			respondXML(rw, srvErr)
			return
		}
		respondError(rw, http.StatusInternalServerError, err)
		return
	}
	respondNZB(rw, nzb)
}

var decoder = schema.NewDecoder()

func (s *Server) search(rw http.ResponseWriter, r *http.Request) {

	decoder.IgnoreUnknownKeys(true)
	var p SearchParams
	err := decoder.Decode(&p, r.Form)
	if err != nil {
		respondError(rw, http.StatusBadRequest, err)
		return
	}
	res, err := s.impl.Search(r.Context(), p)
	if err != nil {
		var srvErr ServerError
		if errors.As(err, &srvErr) {
			respondXML(rw, srvErr)
			return
		}
		respondError(rw, http.StatusInternalServerError, err)
		return
	}
	respondXML(rw, res)
}

func respondXML(rw http.ResponseWriter, v any) {
	rw.Header().Set("Content-Type", "application/xml")
	rw.WriteHeader(http.StatusOK)
	writeXML(rw, v)
}

func respondNZB(rw http.ResponseWriter, nzb NZB) {
	rw.Header().Set("Content-Type", "application/x-nzb")
	rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", nzb.Filename))
	rw.WriteHeader(http.StatusOK)
	rw.Write(nzb.Data)
}

func writeXML(rw http.ResponseWriter, v any) {
	bytes, _ := xmlutil.Marshal(v)
	rw.Write(bytes)
}

func respondError(rw http.ResponseWriter, code int, err error) {

	respondErrorString(rw, code, err.Error())
}

func respondErrorString(rw http.ResponseWriter, code int, err string) {

	rw.Header().Set("Content-Type", "application/xml")
	rw.WriteHeader(http.StatusOK)
	resp := ServerError{
		Code:        code,
		Description: err,
	}
	writeXML(rw, resp)
}
