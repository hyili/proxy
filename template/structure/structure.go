package structure

import (
	"context"
	"fmt"
	"net"
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
}

func ConnectionPairPoolInit(connPairPool *ConnectionPairPool) {
	connPairPool.pool = make(map[int]*ConnectionPairPoolSlot)
	connPairPool.emptySlots = make(chan int)
	connPairPool.maxUnusedId = 0
}

func (connPairPool *ConnectionPairPool) Add(cid int, sid int) int {
	var id int

	select {
	case id, _ = <-connPairPool.emptySlots:
	default:
		id = connPairPool.maxUnusedId
		connPairPool.maxUnusedId += 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	connPair := ConnectionPair{
		Cid:    cid,
		Sid:    sid,
		Ctx:    ctx,
		Cancel: cancel,
	}
	connPairPool.pool[id] = &ConnectionPairPoolSlot{
		id:       id,
		connPair: connPair,
	}

	return id
}

func (connPairPool *ConnectionPairPool) Get(id int) ConnectionPair {
	target := connPairPool.pool[id]

	return target.connPair
}

func (connPairPool *ConnectionPairPool) Recycle(id int) {
	delete(connPairPool.pool, id)
	connPairPool.emptySlots <- id
}

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
		fmt.Println(id)
	}

	connPool.pool[id] = &ConnectionPoolSlot{
		id:   id,
		conn: conn,
	}

	return id
}

func (connPool *ConnectionPool) Get(id int) net.Conn {
	target := connPool.pool[id]

	return target.conn
}

func (connPool *ConnectionPool) Recycle(id int) {
	delete(connPool.pool, id)
	connPool.emptySlots <- id
}

func (connPool *ConnectionPool) DrainPool() {
	for id := range connPool.pool {
		connPool.Recycle(id)
	}
}

func (connPool *ConnectionPool) ClosePool() {
}
