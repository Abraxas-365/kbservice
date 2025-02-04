package kb

import (
	"github.com/Abraxas-365/kbservice/datasource"
	"github.com/Abraxas-365/kbservice/embedding"
	"github.com/Abraxas-365/kbservice/llm"
	"github.com/Abraxas-365/kbservice/vectorstore"
)

type KnowledgeBase struct {
	embedder   embedding.Embedder
	vStore     vectorstore.VectorStore
	store      vectorstore.Store
	llm        *llm.LLM
	datasource datasource.DataSource
}
