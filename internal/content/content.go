package content

type Store interface {
	GetDetails() Details
	Get(key string) (Article, error)
	GetAll(pinned bool) []Article
	Update(key string, doc Document) error
}

type Details struct {
	Name        string
	Description string
	Style       string
}
