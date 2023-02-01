package sender

import (
	"context"

	"github.com/iunary/sesify/internal/compaign"
)

type Provider interface {
	Send(ctx context.Context, comp *compaign.Compaign) compaign.Result
}
