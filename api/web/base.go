package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var errMissRequiredParams = errors.New("miss required params")

func writeError(writer http.ResponseWriter, err error) {
	writer.Header().Add("Content-Type", "application/json")

	_, _ = writer.Write([]byte(fmt.Sprintf("{\"error:\" %s}", err.Error())))
}

func writeJson(writer http.ResponseWriter, v interface{}) {
	marshal, err := json.Marshal(v)
	if err != nil {
		writeError(writer, err)
	} else {
		writer.Header().Add("Content-Type", "application/json")

		_, _ = writer.Write(marshal)
	}
}
