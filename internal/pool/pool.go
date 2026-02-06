package pool

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	Items []T
}

func New[T Resettable]() *Pool[T] {
	return &Pool[T]{}
}

func (p *Pool[T]) Put(i T) {
	p.Items = append(p.Items, i)
}

func (p *Pool[T]) Get() T {
	if len(p.Items) == 0 {
		return *new(T)
	}
	output := p.Items[len(p.Items)-1]
	output.Reset()
	p.Items = p.Items[:len(p.Items)-1]

	return output
}
