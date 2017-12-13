package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/summadb/summadb/database"
	"github.com/summadb/summadb/types"
	"github.com/summadb/summadb/utils"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var SEP = []byte{' '}

func handlewebsocket(db *database.SummaDB, w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed to upgrade to ws.", "err", err)
		http.Error(w, err.Error(), 500)
		return
	}
	for {
		mt, bmessage, err := c.ReadMessage()
		if err != nil {
			log.Error("ws read error.", "err", err, "message", string(bmessage), "mt", mt)
			continue
		}
		method, messageId, body, err := parseMessage(bmessage)
		if err != nil {
			log.Error("ws parsing error.", "message", string(bmessage), "err", err)
			continue
		}

		answer := func(response []byte) { send(c, []byte("answer"), messageId, response) }

		var args Arguments
		if err = json.Unmarshal(body, &args); err != nil {
			log.Error("failed to parse arguments.",
				"message", string(bmessage),
				"err", err)
			answer(jsonError("failed to parse arguments"))
			continue
		}

		switch method {
		case "rev":
			rev, err := db.Rev(args.Path)
			if err != nil {
				answer(jsonError(err.Error()))
				continue
			}
			answer(utils.JSONString(rev))
		case "read":
			tree, err := db.Read(args.Path)
			if err != nil {
				answer(jsonError(err.Error()))
				continue
			}
			resp, err := tree.MarshalJSON()
			if err != nil {
				answer(jsonError(err.Error()))
				continue
			}
			answer(resp)
		case "records":
			records, err := db.Query(args.Path, database.QueryParams{
				KeyStart:   args.KeyStart,
				KeyEnd:     args.KeyEnd,
				Descending: args.Descending,
				Limit:      args.Limit,
			})
			if err != nil {
				answer(jsonError(err.Error()))
				continue
			}
			resp, err := json.Marshal(records)
			if err != nil {
				answer(jsonError(err.Error()))
				continue
			}
			answer(resp)
		case "set":
			err := db.Set(args.Path, args.Record)
			if err != nil {
				answer(jsonError(err.Error()))
				continue
			}
			answer(jsonSuccess())
		case "merge":
			err := db.Merge(args.Path, args.Record)
			if err != nil {
				answer(jsonError(err.Error()))
				continue
			}
			answer(jsonSuccess())
		case "delete":
			err := db.Delete(args.Path, args.Rev)
			if err != nil {
				answer(jsonError(err.Error()))
				continue
			}
			answer(jsonSuccess())
		case "replicate":
			// enter replication state. lock everything until the replication completes.
			replicationId := string(messageId)
			log.Debug("accepting replication.", "id", replicationId)
			err := acceptReplication(c, db, args.Path, replicationId)
			log.Debug("replication ended", "id", replicationId, "err", err)
		default:
			log.Error("ws unknown method.", "message", string(bmessage))
			answer(jsonError("unknown method " + method))
			continue
		}
	}
	defer c.Close()
}

type Arguments struct {
	Path       types.Path `json:"path"`
	Record     types.Tree `json:"record"`
	Rev        string     `json:"rev"`
	KeyStart   string     `json:"key_start"`
	KeyEnd     string     `json:"key_end"`
	Descending bool       `json:"descending"`
	Limit      int        `json:limit`
}

func send(c *websocket.Conn, args ...[]byte) {
	body := bytes.Join(args, []byte{' '})
	c.WriteMessage(1, body)
}

func parseMessage(bmessage []byte) (method string, messageId []byte, body []byte, err error) {
	parts := bytes.SplitN(bmessage, SEP, 3)
	if len(parts) != 3 {
		err = errors.New("should have 3 parts, has " + strings.Itoa(len(parts)))
		return
	}

	method = string(parts[0])
	messageId = parts[1]
	body := parts[2]
	return
}
