package store

import (
	"encoding/json"
	//"fmt"
	"os"
	"sync"
)

// definiamo la struct Todo, lo facciamo qui perchè è strettamente
// legata allo store.
type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed string `json:"completed"`
}

/*Store gestisce l'accesso ai dati dei Todo*/
type Store struct {
	mu       sync.RWMutex //RWMutex è più performante per letture multiple
	todos    map[int]Todo // mappa per accesso veloce tramite ID
	nextID   int
	filePath string
}

// crea e inizializza una nuova istanza dello store.
func New(filePath string) (*Store, error) {
	s := &Store{
		todos:    make(map[int]Todo),
		nextID:   1,
		filePath: filePath,
	}
	return s, s.load()
}

// carica i dati sul file json
func (s *Store) load() error {
	// RLock permette a più lettori di accedere contemporaneamente
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil //il file non esiste e va bene, partiamo da zero
		}
		return err //altro errore di lettura
	}

	var todos []Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return err
	}

	//popoliamo la mappa e troviamo il nextID corretto
	for _, t := range todos {
		s.todos[t.ID] = t
		if t.ID >= s.nextID {
			s.nextID = t.ID + 1
		}
	}
	return nil

}

// scirve lo stato corrente dello store su file
func (s *Store) save() error {
	s.mu.RLock() // basta un lock di lettura per creare la slice temporanea
	defer s.mu.RUnlock()

	return s.saveInternal() //chiama la versione interna per evitare deadlock
}

// saveInternal fa il lavoro sporco, ma PRESUPPONE che un lock
// sia già stato acquisito dal chiamante.
func (s *Store) saveInternal() error {
	todos := make([]Todo, 0, len(s.todos))
	for _, t := range s.todos {
		todos = append(todos, t)
	}

	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

// restituisce una slice di tutti i todo
func (s *Store) GetAll() []Todo {
	s.mu.RLock() //Lock in lettura, più goroutine possono leggere contemporaneamente
	defer s.mu.RUnlock()

	//crea una slice con la dimensione esatta della mappa
	allTodos := make([]Todo, 0, len(s.todos))

	//itera sulla mappa e aggiunge ogni todo alla slice
	for _, todo := range s.todos {
		allTodos = append(allTodos, todo)

	}
	return allTodos
}

func (s *Store) GetByID(ID int) (Todo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, ok := s.todos[ID]
	//fmt.Println(result)

	return result, ok

}

func (s *Store) Create(title string) Todo {
	// Usiamo un Lock() completo perché stiamo per modificare i dati (nextID e la mappa).
	s.mu.Lock()
	defer s.mu.Unlock()

	//creiamo la nuova struct todo
	newTodo := Todo{
		ID:        s.nextID,
		Title:     title,
		Completed: "not completed",
	}

	// aggiungiamo il nuovo elemento alla mappa in memoria
	s.todos[newTodo.ID] = newTodo
	s.nextID++

	// Salviamo lo stato aggiornato su file.
	// In un'app reale, potremmo voler gestire l'errore di salvataggio
	// in modo più granulare, ma per ora lo ignoriamo per semplicità.
	// La gestione errori robusta è un ottimo argomento per una fase successiva!
	s.saveInternal()

	return newTodo
}

func (s *Store) Update(ID int, title string, completed string) (Todo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eleTodo, ok := s.todos[ID]
	if !ok {
		return Todo{}, false
	}
	newTitle := title
	newCompleted := completed
	if title == "" {
		newTitle = eleTodo.Title
	}
	if completed == "" {
		newCompleted = eleTodo.Completed
	}

	newTodo := Todo{
		ID:        eleTodo.ID,
		Title:     newTitle,
		Completed: newCompleted,
	}

	s.todos[ID] = newTodo
	s.saveInternal()
	return newTodo, ok

}

func (s *Store) Delete(ID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exist := s.todos[ID]

	if exist {
		delete(s.todos, ID)
		s.saveInternal()
		return true
	} else {
		return false
	}
}
