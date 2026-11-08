package rgui

// Signal holds reactive state. Use signals and their bindings on the UI thread.
type Signal[T any] struct {
	value     T
	observers subscriptions
}

type subscriptions map[*binding]struct{}

// NewSignal creates a signal with the given initial value.
func NewSignal[T any](value T) *Signal[T] {
	return &Signal[T]{value: value}
}

// Get reads the value and tracks it when a text binding is evaluating.
func (s *Signal[T]) Get() T {
	if activeBinding != nil {
		if s.observers == nil {
			s.observers = make(subscriptions)
		}
		s.observers[activeBinding] = struct{}{}
		activeBinding.dependencies[&s.observers] = struct{}{}
	}
	return s.value
}

// Set updates the value immediately and queues dependent bindings for the frame.
func (s *Signal[T]) Set(value T) {
	s.value = value
	for observer := range s.observers {
		observer.pending[observer] = struct{}{}
	}
}

var activeBinding *binding

type binding struct {
	evaluate     func()
	dependencies map[*subscriptions]struct{}
	pending      subscriptions
}

func (b *binding) evaluateTracked() {
	b.unsubscribe()
	previous := activeBinding
	activeBinding = b
	defer func() { activeBinding = previous }()
	b.evaluate()
}

func (b *binding) unsubscribe() {
	for observers := range b.dependencies {
		delete(*observers, b)
	}
	clear(b.dependencies)
}

func (b *binding) dispose() {
	b.unsubscribe()
	delete(b.pending, b)
}
