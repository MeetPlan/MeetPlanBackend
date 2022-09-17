package httphandlers

import (
	"fmt"
	"github.com/MeetPlan/MeetPlanBackend/sql"
	"net/http"
	"os"
)

const SPRICEVALO = 0
const POTRDILO_O_SOLANJU = 1
const RESETIRANJE_GESLA = 2
const POTRDILO_O_SAMOTESTIRANJU = 3

type Document struct {
	sql.Document
	ExporterName string
}

func (server *httpImpl) FetchAllDocuments(w http.ResponseWriter, r *http.Request) {
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}
	documents, err := server.db.GetAllDocuments()
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while fetching documents", Error: err.Error(), Success: false}, http.StatusInternalServerError)
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
	user, err := server.db.CheckToken(GetAuthorizationJWT(r))
	if err != nil {
		WriteForbiddenJWT(w)
		return
	}
	if !(user.Role == ADMIN || user.Role == PRINCIPAL || user.Role == PRINCIPAL_ASSISTANT) {
		WriteForbiddenJWT(w)
		return
	}

	documentId := r.FormValue("documentId")

	document, err := server.db.GetDocument(documentId)
	if err != nil {
		WriteJSON(w, Response{Data: "Document not found", Error: err.Error(), Success: false}, http.StatusNotFound)
		return
	}

	err = os.Remove(fmt.Sprintf("documents/%s.pdf", document.ID))
	if err != nil {
		WriteJSON(w, Response{Data: "Failed while deleting the document", Error: err.Error(), Success: false}, http.StatusInternalServerError)
		return
	}

	server.db.DeleteDocument(documentId)

	WriteJSON(w, Response{Data: "OK", Success: true}, http.StatusOK)
}
