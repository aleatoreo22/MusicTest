// Package providers define a interface comum que todo provedor de
// música (Spotify, SoundCloud, YouTube) deve implementar. Isso
// desacopla os handlers e a rota /providers dos tipos concretos.
package providers

import (
	"context"
	"time"

	"music-unified-backend/internal/models"
)

// Provider é o contrato unificado de um provedor de música.
// Cada implementação é responsável por seu próprio mecanismo de
// autenticação, cache e formato de resposta específico.
type Provider interface {
	// Name retorna o identificador curto do provider ("spotify",
	// "soundcloud", "youtube"). Usado em /providers e como Source
	// nos tracks retornados.
	Name() string

	// Enabled retorna true se o provider tem configuração mínima
	// para funcionar (credenciais, binário no PATH, etc.).
	Enabled() bool

	// Search retorna até `limit` faixas correspondentes à query.
	// limit <= 0 significa "use o default do provider".
	Search(ctx context.Context, query string, limit int) ([]models.Track, error)

	// Stream resolve a URL reproduzível de uma faixa e quando ela expira.
	// Providers que não expõem stream direto (ex: Spotify) podem
	// retornar ErrNotSupported.
	Stream(ctx context.Context, id string) (url string, expiresAt time.Time, err error)
}
