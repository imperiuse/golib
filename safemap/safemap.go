package safemap

type SafeMap interface {
	Each(EachFunc) int
	Set(interface{}, interface{})
	Del(interface{})
	Get(interface{}) (interface{}, bool)

	GetInterface(interface{}) interface{}
	GetInt(interface{}) *int
	GetInt64(interface{}) *int64
	GetString(interface{}) *string
	GetBool(interface{}) *bool

	Len() int
	Update(interface{}, UpdateFunc)
	Close() map[interface{}]interface{}

	MultiGet(keys map[interface{}]interface{}) map[interface{}]interface{}
	MultiSet(keys map[interface{}]interface{})

	SumI(interface{}, int)
	SumF(interface{}, float64)
	Inc(interface{})
	Dec(interface{})
}

type itemRes struct {
	key   interface{}
	value interface{}
	found bool
}

type commandData struct {
	itemRes
	action  commandAction
	result  chan<- itemRes
	items   map[interface{}]interface{}
	data    chan<- map[interface{}]interface{}
	updater UpdateFunc
}

type safeMap chan commandData
type commandAction int

const (
	rem commandAction = iota
	end
	get
	mget
	mset
	set
	length
	update
	each
)

type UpdateFunc func(interface{}, bool) interface{}
type EachFunc func(key interface{}, val interface{}, cnt int) bool

func (sm safeMap) SumF(key interface{}, delta float64) {
	sm.Update(key, func(val interface{}, found bool) interface{} {
		if found {
			return val.(float64) + delta
		}
		return delta
	})
}

func (sm safeMap) SumI(key interface{}, delta int) {
	sm.Update(key, func(val interface{}, found bool) interface{} {
		if found {
			return val.(int) + delta
		}
		return delta
	})
}

func (sm safeMap) Inc(key interface{}) {
	sm.SumI(key, 1)
}

func (sm safeMap) Dec(key interface{}) {
	sm.SumI(key, -1)
}

func (sm safeMap) Each(fn EachFunc) int {
	n := 0
	reply := make(chan itemRes, sm.Len())
	sm <- commandData{action: each, result: reply}
	for itm := range reply {
		if itm.found == false {
			break
		}

		if !fn(itm.key, itm.value, n) {
			break
		}
		n++
	}
	close(reply)
	return n
}

func (sm safeMap) Update(key interface{}, fn UpdateFunc) {
	sm <- commandData{action: update, updater: fn, itemRes: itemRes{key: key}}
}

func (sm safeMap) Set(key interface{}, value interface{}) {
	sm <- commandData{action: set, itemRes: itemRes{key: key, value: value}}
}

func (sm safeMap) Del(key interface{}) {
	sm <- commandData{action: rem, itemRes: itemRes{key: key}}
}

func (sm safeMap) Get(key interface{}) (value interface{}, found bool) {
	reply := make(chan itemRes)
	sm <- commandData{action: get, result: reply, itemRes: itemRes{key: key}}
	result := <-reply
	close(reply)
	return result.value, result.found
}

func (sm safeMap) GetInterface(key interface{}) interface{} {
	if value, ok := sm.Get(key); ok {
		return value
	}
	return nil
}

func (sm safeMap) GetBool(key interface{}) *bool {
	if vI, ok := sm.Get(key); ok {
		if v, ok := vI.(bool); ok {
			return &v
		}
	}
	return nil
}

func (sm safeMap) GetInt(key interface{}) *int {
	if vI, ok := sm.Get(key); ok {
		if v, ok := vI.(int); ok {
			return &v
		}
	}
	return nil
}

func (sm safeMap) GetInt64(key interface{}) *int64 {
	if vI, ok := sm.Get(key); ok {
		if v, ok := vI.(int64); ok {
			return &v
		}
	}
	return nil
}

func (sm safeMap) GetString(key interface{}) *string {
	if vI, ok := sm.Get(key); ok {
		if v, ok := vI.(string); ok {
			return &v
		}
	}
	return nil
}

func (sm safeMap) Len() int {
	reply := make(chan itemRes)
	sm <- commandData{action: length, result: reply}
	return (<-reply).value.(int)
}

func (sm safeMap) Close() map[interface{}]interface{} {
	reply := make(chan map[interface{}]interface{})
	sm <- commandData{action: end, data: reply}
	return <-reply
}

func (sm safeMap) MultiGet(keys map[interface{}]interface{}) map[interface{}]interface{} {
	reply := make(chan map[interface{}]interface{})
	sm <- commandData{action: mget, items: keys, data: reply}
	return <-reply
}

func (sm safeMap) MultiSet(keys map[interface{}]interface{}) {
	sm <- commandData{action: mset, items: keys}
}

func (sm safeMap) run() {
	store := make(map[interface{}]interface{})
	for command := range sm {
		switch command.action {
		case set:
			store[command.key] = command.value
		case rem:
			delete(store, command.key)
		case get:
			value, found := store[command.key]
			command.result <- itemRes{nil, value, found}
		case length:
			command.result <- itemRes{nil, len(store), true}
		case update:
			value, found := store[command.key]
			store[command.key] = command.updater(value, found)
		case end:
			close(sm)
			command.data <- store
		case mget:
			out := make(map[interface{}]interface{})
			for key := range command.items {
				if val, f := store[key]; f {
					out[key] = val
				}
			}
			command.data <- out
		case mset:
			for key, val := range command.items {
				store[key] = val
			}
		case each:
			func() {
				defer func() { recover() }()
				for i := range store {
					command.result <- itemRes{key: i, value: store[i], found: true}
				}
				command.result <- itemRes{nil, nil, false}
			}()
		}
	}
}

func New(bufsize int) SafeMap {
	sm := make(safeMap, bufsize) // тип safeMap chan commandData
	go sm.run()
	return sm
}
