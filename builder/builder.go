//go:generate moq -out mock.go . Builder

package builder

type Builder interface {
	Build() error
}

type builder struct{}

func (b *builder) Build() error {
	return nil
}

func New() Builder {
	return &builder{}
}
