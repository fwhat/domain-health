package web

import (
	"github.com/Dowte/domain-health/store"
	"net/http"
)

type DomainDeleteRes struct {
	Success bool
}

func DomainDelete(writer http.ResponseWriter, request *http.Request) {
	uuid := request.URL.Query().Get("uuid")
	if uuid == "" {
		writeError(writer, errMissRequiredParams)
		return
	}

	writeJson(writer, &DomainDeleteRes{
		Success: store.GetDomainStore().DeleteDomainByAddress(uuid),
	})
}
