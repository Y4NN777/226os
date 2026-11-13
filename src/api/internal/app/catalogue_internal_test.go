package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/discover"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

func TestCatalogueStaysUnimplementedWithoutAHandler(t *testing.T) {
	s := NewServer()
	id := uuid.MustParse("00000000-0000-7000-8000-0000000000a1")
	for _, call := range []func(*httptest.ResponseRecorder){
		func(rec *httptest.ResponseRecorder) {
			s.ListResources(rec, httptest.NewRequest(http.MethodGet, "/", nil), apigen.ListResourcesParams{})
		},
		func(rec *httptest.ResponseRecorder) {
			s.GetResource(rec, httptest.NewRequest(http.MethodGet, "/", nil), id)
		},
		func(rec *httptest.ResponseRecorder) {
			s.ListResourceSources(rec, httptest.NewRequest(http.MethodGet, "/", nil), id)
		},
	} {
		rec := httptest.NewRecorder()
		call(rec)
		if rec.Code != http.StatusNotImplemented {
			t.Fatalf("status %d", rec.Code)
		}
	}
}

func TestCatalogueMethodsDelegateWhenSet(t *testing.T) {
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer(WithCatalogue(discover.NewHandler(&fakePool{}, c, nil)))
	rec := httptest.NewRecorder()
	s.ListResources(rec, httptest.NewRequest(http.MethodGet, "/", nil), apigen.ListResourcesParams{})
	if rec.Code == http.StatusNotImplemented {
		t.Fatal("catalogue was not wired")
	}
	id := uuid.MustParse("00000000-0000-7000-8000-0000000000a1")
	rec = httptest.NewRecorder()
	s.GetResource(rec, httptest.NewRequest(http.MethodGet, "/", nil), id)
	if rec.Code == http.StatusNotImplemented {
		t.Fatal("get was not wired")
	}
	rec = httptest.NewRecorder()
	s.ListResourceSources(rec, httptest.NewRequest(http.MethodGet, "/", nil), id)
	if rec.Code == http.StatusNotImplemented {
		t.Fatal("sources was not wired")
	}
}
