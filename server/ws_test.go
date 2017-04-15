package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/summadb/summadb/database"
	. "github.com/summadb/summadb/utils"
	. "gopkg.in/check.v1"
)

func TestAll(t *testing.T) {
	TestingT(t)
}

type ServerSuite struct{}

var _ = Suite(&ServerSuite{})

func (s *ServerSuite) TestWebSocket(c *C) {
	var err error
	db := database.Open("/tmp/summadb-test-websocket")
	defer db.Erase()
	h := &Handler{db}
	srv := httptest.NewServer(h)
	defer srv.Close()
	d := websocket.Dialer{}

	conn, resp, err := d.Dial("ws://"+srv.Listener.Addr().String()+"/", nil)

	c.Assert(err, IsNil)

	got := resp.StatusCode
	c.Assert(got, Equals, http.StatusSwitchingProtocols)

	conn.WriteMessage(1, []byte(`set 1 {"path":["cidades","petrolina"],"record":{"nome":"petrolina","uf":"pernambuco"}}`))
	conn.WriteMessage(1, []byte(`records 2 {"path":["cidades"],"key_start":"k"}`))
	for i := 0; i < 2; i++ {
		_, m, err := conn.ReadMessage()
		c.Assert(err, IsNil)
		spl := bytes.SplitN(m, []byte{' '}, 2)
		c.Assert(spl, HasLen, 2)
		if m[0] == '1' {
			c.Assert(spl[1], JSONEquals, jsonSuccess())
		} else if m[0] == '2' {
			c.Assert(string(spl[1]), StartsWith, `[{"_key":"petrolina"`)
		}
	}

	conn.WriteMessage(1, []byte(`rev 3 {}`))
	_, m, _ := conn.ReadMessage()
	rev := strings.Split(string(m), " ")[1]
	conn.WriteMessage(1, []byte(`merge 4 {"record":{"_rev": `+rev+`, "cidades": {"juazeiro": {"nome":"juazeiro","uf":"bahia"}}}}`))
	_, m, _ = conn.ReadMessage()
	c.Assert(bytes.SplitN(m, []byte{' '}, 2)[1], JSONEquals, jsonSuccess())
	conn.WriteMessage(1, []byte(`merge 5 {"record":{"_rev":"2-owiqwqenqwe", "cidades": {"campinagrande": {"nome":"campina grande","uf":"paraíba"}}}}`))
	_, m, _ = conn.ReadMessage()
	c.Assert(string(bytes.SplitN(m, []byte{' '}, 2)[1]), StartsWith, `{"error":"mismatched revs`)
	conn.WriteMessage(1, []byte(`read 6 {"path":["cidades"]}`))
	_, m, _ = conn.ReadMessage()
	var read map[string]interface{}
	err = json.Unmarshal(bytes.SplitN(m, []byte{' '}, 2)[1], &read)
	c.Assert(err, IsNil)
	c.Assert(read["_key"], Equals, "cidades")
	_, ok := read["petrolina"]
	c.Assert(ok, Equals, true)
	_, ok = read["juazeiro"]
	c.Assert(ok, Equals, true)
	_, ok = read["campinagrande"]
	c.Assert(ok, Equals, false)
}
