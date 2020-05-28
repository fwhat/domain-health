package web

import (
	"encoding/json"
	"github.com/qjues/domain-health/store"
	"github.com/qjues/domain-health/store/model"
	"net/http"
	"net/url"
)

type domainAddReq struct {
	Address string
}

type domainAddRes struct {
	Success bool
}

func DomainAdd(writer http.ResponseWriter, request *http.Request) {
	reqData := &domainAddReq{}
	err := json.NewDecoder(request.Body).Decode(&reqData)
	if err != nil {
		writeError(writer, err)

		return
	}

	if reqData.Address == "" {
		writeError(writer, errMissRequiredParams)

		return
	}

	address, err := url.Parse(reqData.Address)

	if err != nil {
		writeError(writer, err)

		return
	}

	domain := model.NewDomain()
	domain.Address = address.String()
	domain.From = "user"

	store.GetDomainStore().SaveDomainInfo(domain)

	writeJson(writer, &domainAddRes{Success: true})
}
