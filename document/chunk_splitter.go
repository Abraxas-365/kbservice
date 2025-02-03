package document

import "strings"

type CharacterSplitter struct {
	ChunkSize    int
	ChunkOverlap int
	Separator    string
}

func NewCharacterSplitter(chunkSize int, chunkOverlap int, separator string) *CharacterSplitter {
	if separator == "" {
		separator = " "
	}

	return &CharacterSplitter{
		ChunkSize:    chunkSize,
		ChunkOverlap: chunkOverlap,
		Separator:    separator,
	}
}

func (cs *CharacterSplitter) SplitText(text string) ([]string, error) {
	if text == "" {
		return nil, nil
	}

	parts := strings.Split(text, cs.Separator)
	var chunks []string
	currentChunk := strings.Builder{}

	for i := 0; i < len(parts); i++ {
		if currentChunk.Len()+len(parts[i])+1 > cs.ChunkSize {
			if currentChunk.Len() > 0 {
				chunks = append(chunks, strings.TrimSpace(currentChunk.String()))

				if cs.ChunkOverlap > 0 {
					overlapText := currentChunk.String()
					if len(overlapText) > cs.ChunkOverlap {
						overlapText = overlapText[len(overlapText)-cs.ChunkOverlap:]
					}
					currentChunk.Reset()
					currentChunk.WriteString(overlapText)
				} else {
					currentChunk.Reset()
				}
			}
		}

		if currentChunk.Len() > 0 {
			currentChunk.WriteString(cs.Separator)
		}
		currentChunk.WriteString(parts[i])
	}

	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	return chunks, nil
}
