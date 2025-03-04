package crud

type createRequest[K any, V any] struct {
	key   K
	value V
	ack   chan bool
}

type readResult[V any] struct {
	value V
	ok    bool
}

type readRequest[K any, V any] struct {
	key      K
	receiver chan readResult[V]
}

type updateRequest[K any, V any] struct {
	key    K
	update func(V) V
	ack    chan bool
}

type deleteRequest[K any] struct {
	key K
	ack chan bool
}

type snapshotRequest[K comparable, V any] struct {
	receiver chan map[K]V
}

func New[K comparable, V any]() Crud[K, V] {
	crud := Crud[K, V]{
		values:    make(map[K]V),
		creates:   make(chan createRequest[K, V]),
		reads:     make(chan readRequest[K, V]),
		updates:   make(chan updateRequest[K, V]),
		deletes:   make(chan deleteRequest[K]),
		snapshots: make(chan snapshotRequest[K, V]),
		quit:      make(chan bool),
	}
	go crud.worker()
	return crud
}

type Crud[K comparable, V any] struct {
	values    map[K]V
	creates   chan createRequest[K, V]
	reads     chan readRequest[K, V]
	updates   chan updateRequest[K, V]
	deletes   chan deleteRequest[K]
	snapshots chan snapshotRequest[K, V]
	quit      chan bool
}

func (c *Crud[K, V]) worker() {
	for {
		select {
		case <-c.quit:
			return
		case create := <-c.creates:
			_, ok := c.values[create.key]
			if !ok {
				c.values[create.key] = create.value
			}
			create.ack <- !ok
		case read := <-c.reads:
			var result readResult[V]
			result.value, result.ok = c.values[read.key]
			read.receiver <- result
		case update := <-c.updates:
			val, ok := c.values[update.key]
			if ok {
				c.values[update.key] = update.update(val)
			}
			update.ack <- ok
		case del := <-c.deletes:
			_, ok := c.values[del.key]
			delete(c.values, del.key)
			del.ack <- ok
		case snapshot := <-c.snapshots:
			ret := make(map[K]V, len(c.values))
			for k, v := range c.values {
				ret[k] = v
			}
			snapshot.receiver <- ret
		}
	}
}

func (c *Crud[K, V]) Close() {
	close(c.quit)
}

func (c *Crud[K, V]) Create(key K, value V) bool {
	ack := make(chan bool)
	c.creates <- createRequest[K, V]{
		key:   key,
		value: value,
		ack:   ack,
	}
	return <-ack
}

func (c *Crud[K, V]) Read(key K) (V, bool) {
	receiver := make(chan readResult[V])
	c.reads <- readRequest[K, V]{
		key:      key,
		receiver: receiver,
	}
	val := <-receiver
	return val.value, val.ok
}

func (c *Crud[K, V]) Update(key K, update func(V) V) bool {
	ack := make(chan bool)
	c.updates <- updateRequest[K, V]{
		key:    key,
		update: update,
		ack:    ack,
	}
	return <-ack
}

func (c *Crud[K, V]) Delete(key K) bool {
	ack := make(chan bool)
	c.deletes <- deleteRequest[K]{
		key: key,
		ack: ack,
	}
	return <-ack
}

func (c *Crud[K, V]) Snapshot() map[K]V {
	receiver := make(chan map[K]V)
	c.snapshots <- snapshotRequest[K, V]{
		receiver: receiver,
	}
	return <-receiver
}
