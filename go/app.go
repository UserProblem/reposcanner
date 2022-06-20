package swagger

import "log"

type App struct {
	Router    HttpRouter
	RepoStore RepoStore
}

func (a *App) Initialize() {
	a.Router = a.NewRouter()
	a.ClearStores()
}

func (a *App) ClearStores() {
	rs, err := NewRepoStore()
	if err != nil {
		log.Fatal("Cannot initialize repository data store.\n")
	}
	a.RepoStore = *rs
}
