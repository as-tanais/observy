package pool

// Resettable интерфейс что задать ограничения для Pool.
type Resettable interface {
	Reset()
}

// Pool структура с generic-параметром, который ограничен типами с методом Reset()
type Pool[T Resettable] struct {
	Items []T
}

// New конструктор который создаёт и возвращает указатель на структуру Pool;
func New[T Resettable]() *Pool[T] {
	return &Pool[T]{}
}

// Put метод который кладет объект в Pool
func (p *Pool[T]) Put(i T) {
	p.Items = append(p.Items, i)
}

// Get метод который возращает объект из Pool
func (p *Pool[T]) Get() T {
	if len(p.Items) == 0 {
		return *new(T)
	}
	output := p.Items[len(p.Items)-1]
	output.Reset()
	p.Items = p.Items[:len(p.Items)-1]

	return output
}
