package store

import (
	"os"
	"testing"
	// La nostra libreria di assertion
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestStore è una funzione helper che crea uno store pulito per ogni test.
// Restituisce lo store e una funzione di "teardown" per pulire dopo il test.

func setupTestStore(t *testing.T) (*Store, func()) {
	// usiamo un file di test temporaneo per non sporcare il nostro todos.json
	testFile := "test_todos.json"

	store, err := New(testFile)
	// require è come assert, ma usa t.Fatal se il check fallisce.
	// Se non riusciamo a creare lo store, non ha senso continuare il test.
	require.NoError(t, err, "La creazione dello store non dovrebbe fallire")

	// La funzione di teardown viene restituita e chiamata alla fine del test
	// usando 'defer'.
	teardown := func() {
		os.Remove(testFile)
	}

	return store, teardown
}

// Test per il ciclo di vita completo di un Todo.
func TestTodoLifecycle(t *testing.T) {
	store, teardown := setupTestStore(t)
	defer teardown() // Assicura che la pulizia venga eseguita alla fine del test.

	// usiamo t.Run per raggruppare i sottotest
	t.Run("1. Create Todo", func(t *testing.T) {
		// Azione
		created := store.Create("Test di creazione")

		// verifica assertion
		assert.Equal(t, 1, created.ID, "L'ID del primo Todo dovrebbe essere 1")
		assert.Equal(t, "Test di creazione", created.Title, "Il titolo non corrisponde")
		assert.Equal(t, "not completed", created.Completed, "Un nuovo Todo non dovrebbe essere completato")
	})

	t.Run("2. Get Todo By ID", func(t *testing.T) {
		// Azione
		todo, found := store.GetByID(1)

		// Verifica
		assert.True(t, found, "Il todo con ID 1 dovrebbe essere trovato")
		assert.Equal(t, 1, todo.ID)
		assert.Equal(t, "Test di creazione", todo.Title)
	})

	t.Run("3. Get a non-existent Todo", func(t *testing.T) {
		// Azione
		_, found := store.GetByID(999)

		// Verifica
		assert.False(t, found, "Un todo con ID 999 non dovrebbe esistere")
	})

	t.Run("4. Update Todo", func(t *testing.T) {
		//Azione
		updated, found := store.Update(1, "Titolo aggiornato", "completed")

		//verifica
		assert.True(t, found)
		assert.Equal(t, "Titolo aggiornato", updated.Title)
		assert.Equal(t, "completed", updated.Completed)

		// contro verifica: rileggiamo il dato per essere sicuri
		reRead, _ := store.GetByID(1)
		assert.Equal(t, "Titolo aggiornato", reRead.Title)

	})

	t.Run("5. Delete Todo", func(t *testing.T) {
		//Azione
		found := store.Delete(1)

		//verifica
		assert.True(t, found, "Il Delete dovrebbe avere successo per un id esistente")

		// contro verifica
		_, foundAfterDelete := store.GetByID(1)
		assert.False(t, foundAfterDelete, "Il todo non dovrebbe più esistere dopo la cancellazione")

	})
}
