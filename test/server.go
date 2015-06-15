package test

import (
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/adams-sarah/test2doc/doc"
)

// resources = map[uri]Resource
var resources = map[string]*doc.Resource{}

type Server struct {
	*httptest.Server
	doc *doc.Doc
}

// TODO: filter out 404 responses
func NewServer(handler http.Handler, pkgDir string) (s *Server, err error) {
	outDoc, err := doc.NewDoc(pkgDir)
	if err != nil {
		return s, err
	}

	httptestServer := httptest.NewServer(handleAndRecord(handler, outDoc))

	return &Server{
		httptestServer,
		outDoc,
	}, nil
}

func (s *Server) Finish() {
	s.Close()

	for _, r := range resources {
		s.doc.AddResource(r)
	}

	err := s.doc.Write()
	if err != nil {
		panic(err.Error())
	}
}

func handleAndRecord(handler http.Handler, outDoc *doc.Doc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		u := doc.NewURL(req)
		path := u.ParameterizedPath

		// setup
		if resources[path] == nil {
			resources[path] = doc.NewResource(u)
		}

		// copy request body into Request object
		docReq, err := doc.NewRequest(req)
		if err != nil {
			log.Println("Error:", err.Error())
			return
		}

		// record response
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)

		// store response body in Response object
		docResp := doc.NewResponse(resp)

		// add Action to Resource's list of Actions
		method := doc.HTTPMethod(req.Method)
		action, err := doc.NewAction(method, docReq, docResp)
		if err != nil {
			log.Println("Error:", err.Error())
			return
		}

		resources[path].AddAction(action)

		// copy response over to w
		w.WriteHeader(resp.Code)
		doc.CopyHeader(w.Header(), resp.Header())
		w.Write(resp.Body.Bytes())
	}
}
