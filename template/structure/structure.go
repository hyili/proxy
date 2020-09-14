package structure

import (
	"context"
	"net"
	"sync"
)

// Connection Pair (Proxy) & Connection Pair (Proxy) Pool
type ConnectionPair struct {
	Cid    int
	Sid    int
	Ctx    context.Context
	Cancel func()
}

type ConnectionPairPoolSlot struct {
	id       int
	connPair ConnectionPair
}

type ConnectionPairPool struct {
	pool        map[int]*ConnectionPairPoolSlot
	emptySlots  chan int
	maxUnusedId int
	lock        sync.Mutex
}

func ConnectionPairPoolInit(connPairPool *ConnectionPairPool) {
	connPairPool.pool = make(map[int]*ConnectionPairPoolSlot)
	connPairPool.emptySlots = make(chan int)
	connPairPool.maxUnusedId = 0
}

func (connPairPool *ConnectionPairPool) Add(cid int, sid int) int {
	var id int

	/*
		select {
		case id, _ = <-connPairPool.emptySlots:
		default:
			id = connPairPool.maxUnusedId
			connPairPool.maxUnusedId += 1
		}
	*/
	id = connPairPool.maxUnusedId
	connPairPool.maxUnusedId += 1

	connPair := ConnectionPair{
		Cid: cid,
		Sid: sid,
	}
	connPairPool.lock.Lock()
	connPairPool.pool[id] = &ConnectionPairPoolSlot{
		id:       id,
		connPair: connPair,
	}
	connPairPool.lock.Unlock()

	return id
}

func (connPairPool *ConnectionPairPool) Get(id int) ConnectionPair {
	connPairPool.lock.Lock()
	target := connPairPool.pool[id]
	connPairPool.lock.Unlock()

	return target.connPair
}

func (connPairPool *ConnectionPairPool) Recycle(id int) {
	connPairPool.lock.Lock()
	delete(connPairPool.pool, id)
	connPairPool.lock.Unlock()
	connPairPool.emptySlots <- id
}

// TODO: lock issue
func (connPairPool *ConnectionPairPool) DrainPool() {
	for id := range connPairPool.pool {
		connPairPool.Recycle(id)
	}
}

// Connection & Connection Pool
type ConnectionPoolSlot struct {
	id   int
	conn net.Conn
}

type ConnectionPool struct {
	pool        map[int]*ConnectionPoolSlot
	emptySlots  chan int
	maxUnusedId int
	lock        sync.Mutex
}

func ConnectionPoolInit(connPool *ConnectionPool) {
	connPool.pool = make(map[int]*ConnectionPoolSlot)
	connPool.emptySlots = make(chan int)
	connPool.maxUnusedId = 0
}

func (connPool *ConnectionPool) Add(conn net.Conn) int {
	var id int

	select {
	case id, _ = <-connPool.emptySlots:
	default:
		id = connPool.maxUnusedId
		connPool.maxUnusedId += 1
	}

	connPool.lock.Lock()
	connPool.pool[id] = &ConnectionPoolSlot{
		id:   id,
		conn: conn,
	}
	connPool.lock.Unlock()

	return id
}

func (connPool *ConnectionPool) Get(id int) net.Conn {
	connPool.lock.Lock()
	target := connPool.pool[id]
	connPool.lock.Unlock()

	return target.conn
}

func (connPool *ConnectionPool) Recycle(id int) {
	connPool.lock.Lock()
	delete(connPool.pool, id)
	connPool.lock.Unlock()
	connPool.emptySlots <- id
}

// TODO: lock issue
func (connPool *ConnectionPool) DrainPool() {
	for id := range connPool.pool {
		connPool.Recycle(id)
	}
}

func (connPool *ConnectionPool) ClosePool() {
}
