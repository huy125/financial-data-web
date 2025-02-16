package store

type Store struct {
	db *DB

	user	*userService
	stocks	*stockService
	metrics	*metricService
}

func New(db *DB) *Store {
	store := &Store{
		db: db,
	}

	store.user = &userService{db: db}
	store.stocks = &stockService{db: db}
	store.metrics = &metricService{db: db}

	return store
}
