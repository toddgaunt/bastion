package content

type Store interface {
	Get(key string) (Article, bool)
	GetAll(pinned bool) []Article
	GetDetails() Details
}

type Details struct {
	Name        string
	Description string
	Style       string
}
