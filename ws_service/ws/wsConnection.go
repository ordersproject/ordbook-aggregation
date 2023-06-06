package ws

import (
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net"
	"sync"
	"time"
)

type CData struct {
	Msg []byte
	OpCode ws.OpCode
}

type Connection struct {
	wsConn    net.Conn
	writeChan chan []byte
	readChan  chan CData
	closeChan chan byte
	isClose   bool
	mutex     *sync.Mutex
}

func InitConnect(conn net.Conn) *Connection {
	connect := &Connection{
		wsConn:    conn,
		writeChan: make(chan []byte, 1000),//
		readChan:  make(chan CData, 1000),//
		closeChan: make(chan byte, 1),
		isClose:   false,
		mutex:     new(sync.Mutex),
	}
	//go loop
	go connect.readLoop()
	go connect.writeLoop()
	return connect
}

func (c *Connection) readLoop() {
	for {
		msg, opCode, err := wsutil.ReadClientData(c.wsConn)
		if  err != nil {
			fmt.Println("readLoop err :", err)
			c.close()
			return
		}
		select {
		case c.readChan <- CData{msg, opCode}:
		case <-c.closeChan:
			c.close()
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (c *Connection) writeLoop() {
	var msg []byte
	for {
		select {
		case msg = <-c.writeChan:
		case <-c.closeChan:
			c.close()
			return
		}
		err := wsutil.WriteServerMessage(c.wsConn, ws.OpText, msg)
		if err != nil {
			fmt.Println("writeLoop err :", err)
			c.close()
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (c *Connection) close()  {
	err := c.wsConn.Close()
	c.mutex.Lock()
	if !c.isClose {
		close(c.closeChan)
		c.isClose = true
	}
	c.mutex.Unlock()
	if err != nil {
		//fmt.Printf("CO[%s], close err:%s\n", c.wsConn.RemoteAddr().String(), err.Error())
	}
}

func (c *Connection) ReadMessage() ([]byte, ws.OpCode, error) {
	var (
		cData CData
		//data []byte
	)
	select{
	case cData = <- c.readChan:
	case <- c.closeChan:
		return nil, 0, errors.New("connection is closed")
	}
	return cData.Msg, cData.OpCode, nil
}

func (c *Connection) WriteMessage(data []byte) error {
	select{
	case c.writeChan <- data:
	case <- c.closeChan:
		return errors.New("connection is closed")
	}
	return nil
}
