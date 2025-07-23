package store

import (
	"fmt"
	// Import "blank" per il driver. L'underscore dice a Go di eseguire
	// solo la funzione di init() del pacchetto, che lo registra.
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// definiamo la struct Todo, lo facciamo qui perchè è strettamente
// legata allo store.
type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed string `json:"completed"`
}

/*Store gestisce l'accesso ai dati dei Todo*/
/*
type Store struct {
	mu       sync.RWMutex //RWMutex è più performante per letture multiple
	todos    map[int]Todo // mappa per accesso veloce tramite ID
	nextID   int
	filePath string
}
*/
//versione per db sqlite
type Store struct {
	db *sql.DB
}

// crea e inizializza una nuova istanza dello store.
/*
func New(filePath string) (*Store, error) {
	s := &Store{
		todos:    make(map[int]Todo),
		nextID:   1,
		filePath: filePath,
	}
	return s, s.load()
}
*/

// New crea una nuova istanza dello Store e inizializza il database.
func New(dbPath string) (*Store, error) {
	// Apriamo la connessione al database. Se il file non esiste, viene creato.
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("errore nell'aprire il db: %w", err)
	}

	// Ping verifica che la connessione sia effettivamente valida.
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("errore nel ping del db: %w", err)
	}

	// Creiamo la tabella se non esiste.
	if err := createTable(db); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

// createTable esegue la query per creare la nostra tabella 'todos'.
func createTable(db *sql.DB) error {
	// Usiamo TEXT per i campi stringa e INTEGER PRIMARY KEY AUTOINCREMENT
	// per un ID che si auto-incrementa.
	query := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		completed TEXT NOT NULL
	);`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("errore nella creazione della tabella: %w", err)
	}
	return nil
}

/*
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
*/

/*
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
*/

// restituisce una slice di tutti i todo
/*
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
*/
func (s *Store) GetAll() ([]Todo, error) {
	query := "SELECT * FROM todos"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("errore nella query get all: %w", err)
	}
	defer rows.Close() //fondamentale per rilasciare la connessione al database

	// Creiamo una slice per contenere i risultati.
	var todos []Todo

	// Iteriamo su tutte le righe restituite.
	for rows.Next() {
		var t Todo
		// Scan mappa le colonne della riga corrente nei campi della nostra struct.
		if err := rows.Scan(&t.ID, &t.Title, &t.Completed); err != nil {
			// Se una riga dà errore, logghiamo e continuiamo, o restituiamo l'errore.
			return nil, fmt.Errorf("errore nello scan di una riga: %w", err)
		}
		todos = append(todos, t)
	}

	// Controlliamo se ci sono stati errori durante l'iterazione.
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("errore durante l'iterazione delle righe: %w", err)
	}

	return todos, nil
}

/*
func (s *Store) GetByID(ID int) (Todo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, ok := s.todos[ID]
	//fmt.Println(result)

	return result, ok

}
*/

func (s *Store) GetByID(ID int) (Todo, error) {
	query := "SELECT * FROM todos WHERE id=?"

	var newEle Todo
	err := s.db.QueryRow(query, ID).Scan(&newEle)
	if err != nil {
		return Todo{}, fmt.Errorf("errore nel ritornare l'elemento cercato: %w", err)
	}

	return newEle, nil

}

/*
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
*/

/*metodo create con sql*/
func (s *Store) Create(title string) (Todo, error) {
	initialStatus := "not completed"

	// returning id ci ritorna l'id appena generato
	query := "INSERT INTO todos (title, completed) VALUES (?,?) RETURNING id"

	var newID int
	/* usiamo QueryRow che è perfetta quando come ritorno ci aspettiamo una sola riga */
	err := s.db.QueryRow(query, title, initialStatus).Scan(&newID)
	if err != nil {
		return Todo{}, fmt.Errorf("errore nell'inserimento del todo: %w", err)

	}

	newTodo := Todo{
		ID:        newID,
		Title:     title,
		Completed: initialStatus,
	}

	return newTodo, nil
}

/*
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
*/

func (s *Store) Update(ID int, title string, completed string) (Todo, error) {
	query := "UPDATE todos SET title = ?, completed = ? WHERE id = ?"
	_, err := s.db.Exec(query, title, completed, ID)
	if err != nil {
		return Todo{}, fmt.Errorf("errore nell'update dell'elemento: %w", err)
	}

	return s.GetByID(ID)

}

/*
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
*/

func (s *Store) Delete(ID int) error {
	query := "DELETE FROM todos WHERE id = ?"
	result, err := s.db.Exec(query, ID)
	if err != nil {
		return fmt.Errorf("errore nella cancellazione: %w", err)

	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("errore nel recuperare le righe modificate dopo la cancellazione: %w", err)
	}

	if rowsAffected == 0 {
		// Nessuna riga cancellata significa ID non trovato.
		return sql.ErrNoRows
	}

	return nil // Successo! Non c'è nulla da restituire.
}
