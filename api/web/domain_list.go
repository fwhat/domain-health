package web

import (
	"github.com/qjues/domain-health/store"
	"net/http"
)

func DomainList(writer http.ResponseWriter, request *http.Request) {
	list := store.GetDomainStore().ReadAllDomainListNoError()

	writeJson(writer, list)
}
