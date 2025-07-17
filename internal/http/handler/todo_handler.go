package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv" // Pacchetto per la conversione di stringhe
	"todolist-api-v2/internal/store"

	"github.com/go-chi/chi/v5"
	//qui definiremo lo store e i suoi metodi
)

// qui metteremo gli handler

//TodoHandler collega gli handler HTTP conlo store

type TodoHandler struct {
	Store *store.Store
}

// crea un nuovo handler con una dipendenza dallo store
func NewTodoHandler(s *store.Store) *TodoHandler {
	return &TodoHandler{
		Store: s,
	}
}

// GetAll è l'handler per GET /todos.
// Nota il ricevitore (h *TodoHandler). Questo lega la funzione alla struct.
func (h *TodoHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// 1. Chiama la logica di business (la cucina).
	todos := h.Store.GetAll()

	// 2. Prepara e invia la risposta HTTP (il cameriere serve il piatto).
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK

	if err := json.NewEncoder(w).Encode(todos); err != nil {
		// Se c'è un errore qui, è un problema del server.
		// Potremmo loggarlo, ma per ora rispondiamo con un errore generico.
		// (In realtà è difficile che json.NewEncoder fallisca con una slice valida)
		http.Error(w, "Errore durante la codifica della risposta", http.StatusInternalServerError)
	}
}

func (h *TodoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// === PASSO 1: Estrarre il parametro dall'URL ===
	// chi.URLParam prende la richiesta (r) e il nome del parametro
	// che abbiamo definito nella rotta ("todoID").
	// Restituisce SEMPRE una stringa.
	idStr := chi.URLParam(r, "todoID")

	// === PASSO 2: Convertire la stringa in intero ===
	// strconv.Atoi ("ASCII to Integer") è la funzione standard per questo.
	// Restituisce l'intero e un potenziale errore.
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// Se la conversione fallisce, significa che il client ha inviato un ID non valido
		// (es. /todos/abc). Questa è una "Bad Request".
		http.Error(w, "ID non valido, deve essere un numero intero", http.StatusBadRequest) // 400
		return                                                                              // Interrompiamo l'esecuzione dell'handler.
	}
	getedTodo, statusSearch := h.Store.GetByID(id)

	if statusSearch {
		fmt.Println("trovato")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(getedTodo)
	} else {
		fmt.Println("non trovato")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		//json.NewEncoder(w).Encode("Elemento non presente nella lista")
		http.Error(w, "Elemento non presente nella lista", http.StatusNotFound)

	}

}

// gestisce le richieste POST /todos
func (h *TodoHandler) Create(w http.ResponseWriter, r *http.Request) {
	// 1. Definiamo una struct per decodificare il JSON in arrivo.
	//    Ci aspettiamo solo il campo 'title' dal client.
	var input struct {
		Title string `json:"title"`
	}

	// 2. Decodifichiamo il corpo della richiesta.
	//    json.NewDecoder legge da r.Body (la richiesta) e Decode popola
	//    la nostra struct 'input'.
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// Se il JSON è malformato o mancante, è un errore del client.
		http.Error(w, "Corpo della richiesta JSON non valido", http.StatusBadRequest) // 400 Bad Request
		return
	}

	// 3. Facciamo una validazione di base.
	if input.Title == "" {
		http.Error(w, "Il campo 'title' non può essere vuoto", http.StatusBadRequest)
		return
	}

	// 4. Chiamiamo lo store per creare effettivamente il todo.
	createdTodo := h.Store.Create(input.Title)

	// 5. Rispondiamo al client.
	w.Header().Set("Content-Type", "application/json")
	// Impostiamo lo status code a 201 Created, che è lo standard per POST andati a buon fine.
	w.WriteHeader(http.StatusCreated)
	// Inviamo il JSON del todo appena creato come corpo della risposta.
	json.NewEncoder(w).Encode(createdTodo)
}

func (h *TodoHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "todoID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// Se la conversione fallisce, significa che il client ha inviato un ID non valido
		// (es. /todos/abc). Questa è una "Bad Request".
		http.Error(w, "ID non valido, deve essere un numero intero", http.StatusBadRequest) // 400
		return                                                                              // Interrompiamo l'esecuzione dell'handler.
	}

	var input struct {
		Title     string `json:"title"`
		Completed string `json:"completed"`
	}

	// 2. Decodifichiamo il corpo della richiesta.
	//    json.NewDecoder legge da r.Body (la richiesta) e Decode popola
	//    la nostra struct 'input'.
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// Se il JSON è malformato o mancante, è un errore del client.
		http.Error(w, "Corpo della richiesta JSON non valido", http.StatusBadRequest) // 400 Bad Request
		return
	}

	updatedTodo, statusSearch := h.Store.Update(id, input.Title, input.Completed)
	if statusSearch {
		fmt.Println("trovato")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedTodo)
	} else {
		fmt.Println("non trovato")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		http.Error(w, "Elemento non presente nella lista", http.StatusNotFound)

	}

}

func (h *TodoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "todoID")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID non valido, deve essere un intero", http.StatusBadRequest)
		return
	}

	resultDelete := h.Store.Delete(id)

	if resultDelete {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		json.NewEncoder(w).Encode("Eliminato")
	} else {
		http.Error(w, "Todo non trovato", http.StatusNotFound)
		return
	}
}
