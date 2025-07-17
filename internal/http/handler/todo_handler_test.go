// File: internal/http/handler/todo_handler_test.go
/* per lanciare il test andare nella cartella radice del progetto ed eseguire
$go test ./... -v */
package handler

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"todolist-api-v2/internal/store"
)

// setupTestAPI è una funzione helper che costruisce un'API completa in-memory per i test.
// Crea uno store, un handler e un router chi, proprio come in main.go.
// Restituisce il router (che possiamo usare per inviare richieste) e una funzione di teardown.
func setupTestAPI(t *testing.T) (http.Handler, func()) {
	testFile := "handler_test_todos.json"

	// 1. Crea lo store
	s, err := store.New(testFile)
	require.NoError(t, err)

	// 2. Crea l'handler
	h := NewTodoHandler(s)

	// 3. Crea il router e registra le rotte
	// Questo è il passo FONDAMENTALE per testare handler che usano parametri URL.
	r := chi.NewRouter()
	r.Route("/todos", func(r chi.Router) {
		r.Get("/", http.HandlerFunc(h.GetAll))
		r.Post("/", http.HandlerFunc(h.Create))
		r.Route("/{todoID}", func(r chi.Router) {
			r.Get("/", http.HandlerFunc(h.GetByID))
			r.Put("/", http.HandlerFunc(h.Update))
			r.Delete("/", http.HandlerFunc(h.Delete))
		})
	})

	// 4. Definisci la funzione di pulizia
	teardown := func() {
		os.Remove(testFile)
	}

	return r, teardown
}

// TestTodoHandlers copre l'intero ciclo di vita di un todo attraverso l'API.
func TestTodoHandlers(t *testing.T) {
	// Setup
	router, teardown := setupTestAPI(t)
	defer teardown()

	var createdTodoID int // Variabile per passare l'ID tra i sotto-test

	// === Test 1: Creare un nuovo Todo (POST /todos) ===
	t.Run("POST /todos - Success", func(t *testing.T) {
		// Preparazione richiesta
		payload := `{"title":"Testare gli handler"}`
		req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Esecuzione
		router.ServeHTTP(rr, req)

		// Verifica
		assert.Equal(t, http.StatusCreated, rr.Code)

		var respBody store.Todo
		err := json.Unmarshal(rr.Body.Bytes(), &respBody)
		require.NoError(t, err)
		assert.Equal(t, "Testare gli handler", respBody.Title)
		assert.Equal(t, "not completed", respBody.Completed)
		assert.NotZero(t, respBody.ID) // L'ID dovrebbe essere stato assegnato

		createdTodoID = respBody.ID // Salviamo l'ID per i test successivi
	})

	// === Test 2: Ottenere il Todo appena creato (GET /todos/{id}) ===
	t.Run("GET /todos/{id} - Success", func(t *testing.T) {
		require.NotZero(t, createdTodoID, "L'ID del todo creato non può essere zero per questo test")

		req := httptest.NewRequest(http.MethodGet, "/todos/"+strconv.Itoa(createdTodoID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var respBody store.Todo
		err := json.Unmarshal(rr.Body.Bytes(), &respBody)
		require.NoError(t, err)
		assert.Equal(t, createdTodoID, respBody.ID)
		assert.Equal(t, "Testare gli handler", respBody.Title)
	})

	// === Test 3: Ottenere tutti i Todo (GET /todos) ===
	t.Run("GET /todos - Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/todos", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var respBody []store.Todo
		err := json.Unmarshal(rr.Body.Bytes(), &respBody)
		require.NoError(t, err)
		assert.Len(t, respBody, 1, "Dovrebbe esserci un solo todo nella lista")
		assert.Equal(t, createdTodoID, respBody[0].ID)
	})

	// === Test 4: Aggiornare il Todo (PUT /todos/{id}) ===
	t.Run("PUT /todos/{id} - Success", func(t *testing.T) {
		payload := `{"title":"Titolo aggiornato dagli handler","completed":"completed"}`
		req := httptest.NewRequest(http.MethodPut, "/todos/"+strconv.Itoa(createdTodoID), bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var respBody store.Todo
		err := json.Unmarshal(rr.Body.Bytes(), &respBody)
		require.NoError(t, err)
		assert.Equal(t, "Titolo aggiornato dagli handler", respBody.Title)
		assert.Equal(t, "completed", respBody.Completed)
	})

	// === Test 5: Tentare di ottenere un todo inesistente (GET /todos/{id}) ===
	t.Run("GET /todos/{id} - Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/todos/9999", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	// === Test 6: Cancellare il Todo (DELETE /todos/{id}) ===
	t.Run("DELETE /todos/{id} - Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/todos/"+strconv.Itoa(createdTodoID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
	})

	// === Test 7: Verificare che il Todo sia stato cancellato (GET /todos/{id}) ===
	t.Run("GET /todos/{id} after DELETE - Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/todos/"+strconv.Itoa(createdTodoID), nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

// Puoi aggiungere altri Test... per i casi di fallimento specifici
func TestCreateTodoHandler_Failures(t *testing.T) {
	router, teardown := setupTestAPI(t)
	defer teardown()

	t.Run("POST /todos - Invalid JSON", func(t *testing.T) {
		payload := `{"title":"Testare gli handler"` // JSON malformato
		req := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	// Aggiungi altri test per titoli vuoti, etc.
}
