package misc

import (
	"github.com/gsrai/go-ionic/internal/clients/covalent"
	"github.com/gsrai/go-ionic/internal/types"
)

type LogEventQuery struct {
	types.InputCSVRecord
	StartBlock, EndBlock covalent.Block
}
