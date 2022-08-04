package httphandlers

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"net/http"
	"os"
)

const SPRICEVALO = 0
const POTRDILO_O_SOLANJU = 1

type Document struct {
	sql.Document
	ExporterName string
}

func (server *httpImpl) FetchAllDocuments(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant") {
		WriteForbiddenJWT(w)
		return
	}
	documents, err := server.db.GetAllDocuments()
	if err != nil {
		return
	}
	documentsJson := make([]Document, 0)
	for i := 0; i < len(documents); i++ {
		user, err := server.db.GetUser(documents[i].ExportedBy)
		if err != nil {
			continue
		}
		documentsJson = append(documentsJson, Document{
			Document:     documents[i],
			ExporterName: user.Name,
		})
	}
	WriteJSON(w, Response{Data: documentsJson, Success: true}, http.StatusOK)
}

func (server *httpImpl) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	jwt, err := sql.CheckJWT(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(jwt["role"] == "admin" || jwt["role"] == "principal" || jwt["role"] == "principal assistant") {
		WriteForbiddenJWT(w)
		return
	}

	documentId := r.FormValue("documentId")

	err = os.Remove(fmt.Sprintf("documents/%s.pdf", documentId))
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while deleting the document", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	server.db.DeleteDocument(documentId)

	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}
