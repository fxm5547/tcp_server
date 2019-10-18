package tcp_server

import (
	"bytes"
	"net"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func buildTestServer() *server {
	return New("localhost:9999")
}

func Test_accepting_new_client_callback(t *testing.T) {
	server := buildTestServer()

	var messageReceived bool
	var messageText []byte
	var newClient bool
	var connectinClosed bool

	server.OnNewClient(func(c *Client) {
		newClient = true
	})
	server.OnSplitMessage(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			// We have a full newline-terminated line.
			return i + 1, data[:i+1], nil
		}
		if atEOF {
			return 0, nil, nil
		}
		// Request more data.
		return 0, nil, nil
	})
	server.OnNewMessage(func(c *Client, message []byte) {
		t.Log("OnNewMessage", string(message))
		messageReceived = true
		messageText = message
	})
	server.OnClientConnectionClosed(func(c *Client, err error) {
		t.Log("OnClientConnectionClosed")
		if err != nil {
			t.Log("err: ", err.Error())
		}
		connectinClosed = true
	})
	go server.Listen()

	// Wait for server
	// If test fails - increase this value
	time.Sleep(10 * time.Millisecond)

	conn, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		t.Fatal("Failed to connect to test server")
	}
	conn.Write([]byte("Test message\n"))
	conn.Close()

	// Wait for server
	time.Sleep(10 * time.Millisecond)

	Convey("Messages should be equal", t, func() {
		So(string(messageText), ShouldEqual, "Test message\n")
	})
	Convey("It should receive new client callback", t, func() {
		So(newClient, ShouldEqual, true)
	})
	Convey("It should receive message callback", t, func() {
		So(messageReceived, ShouldEqual, true)
	})
	Convey("It should receive connection closed callback", t, func() {
		So(connectinClosed, ShouldEqual, true)
	})
}
