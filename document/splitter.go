package document

// Splitter interface defines methods for splitting text into chunks
type Splitter interface {
	SplitText(text string) ([]string, error)
}

// SplitDocuments splits multiple documents using a splitter
func SplitDocuments(splitter Splitter, documents []Document) ([]Document, error) {
	texts := make([]string, len(documents))
	metadatas := make([]map[string]interface{}, len(documents))

	for i, doc := range documents {
		texts[i] = doc.PageContent
		metadatas[i] = doc.Metadata
	}

	return CreateDocuments(splitter, texts, metadatas)
}

// CreateDocuments creates documents from texts and metadata
func CreateDocuments(splitter Splitter, texts []string, metadatas []map[string]interface{}) ([]Document, error) {
	if len(metadatas) == 0 {
		metadatas = make([]map[string]interface{}, len(texts))
		for i := range metadatas {
			metadatas[i] = make(map[string]interface{})
		}
	}

	if len(texts) != len(metadatas) {
		return nil, ErrMetadataTextMismatch
	}

	var documents []Document

	for i := range texts {
		chunks, err := splitter.SplitText(texts[i])
		if err != nil {
			return nil, err
		}

		for _, chunk := range chunks {
			doc := Document{
				PageContent: chunk,
				Metadata:    copyMetadata(metadatas[i]),
			}
			documents = append(documents, doc)
		}
	}

	return documents, nil
}

func copyMetadata(metadata map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		copy[k] = v
	}
	return copy
}
