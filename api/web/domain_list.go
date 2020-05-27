package web

import (
	"github.com/Dowte/domain-health/store"
	"net/http"
)

func DomainList(writer http.ResponseWriter, request *http.Request) {
	list := store.GetDomainStore().ReadAllDomainListNoError()

	writeJson(writer, list)
}
