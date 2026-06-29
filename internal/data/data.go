package data

import (
	"helloworld/internal/data/base"

	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(base.NewData, NewTodoRepo)

// Data keeps the old internal/data.Data name while the implementation lives in base.
type Data = base.Data
