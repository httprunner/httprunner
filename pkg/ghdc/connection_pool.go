package ghdc

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type ConnectionPool struct {
	mu       sync.Mutex
	conns    chan net.Conn
	host     string
	port     string
	maxConns int
}

func newConnectionPool(host string, port string, maxConns int) (*ConnectionPool, error) {
	pool := &ConnectionPool{
		conns:    make(chan net.Conn, maxConns),
		host:     host,
		port:     port,
		maxConns: maxConns,
	}

	for i := 0; i < maxConns; i++ {
		conn, err := pool.newConnection()
		if err != nil {
			return nil, err
		}
		pool.conns <- conn
	}

	return pool, nil
}

func (p *ConnectionPool) newConnection() (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", p.host, p.port), 5*time.Second)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (p *ConnectionPool) getConnection() (net.Conn, error) {
	select {
	case conn := <-p.conns:
		return conn, nil
	default:
		// 如果池中没有可用连接，创建一个新的连接
		return p.newConnection()
	}
}

func (p *ConnectionPool) releaseConnection(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case p.conns <- conn:
		// 放回连接池成功
	default:
		// 连接池已满，关闭连接
		conn.Close()
	}
}

func (p *ConnectionPool) close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.conns)
	for conn := range p.conns {
		conn.Close()
	}
}
