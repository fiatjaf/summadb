package server

import (
	"bytes"
	"encoding/json"
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

		parts := bytes.SplitN(bmessage, SEP, 3)
		if len(parts) != 3 {
			log.Error("ws invalid number of parts.",
				"parts", parts)
			continue
		}

		method := string(parts[0])
		messageId := parts[1]
		var args Arguments
		if err = json.Unmarshal(parts[2], &args); err != nil {
			log.Error("failed to parse arguments.",
				"message", string(bmessage),
				"err", err)
			answer(c, messageId, jsonError("failed to parse arguments"))
			continue
		}

		switch method {
		case "rev":
			rev, err := db.Rev(args.Path)
			if err != nil {
				answer(c, messageId, jsonError(err.Error()))
				continue
			}
			answer(c, messageId, utils.JSONString(rev))
		case "read":
			tree, err := db.Read(args.Path)
			if err != nil {
				answer(c, messageId, jsonError(err.Error()))
				continue
			}
			resp, err := tree.MarshalJSON()
			if err != nil {
				answer(c, messageId, jsonError(err.Error()))
				continue
			}
			answer(c, messageId, resp)
		case "records":
			records, err := db.Query(args.Path, database.QueryParams{
				KeyStart:   args.KeyStart,
				KeyEnd:     args.KeyEnd,
				Descending: args.Descending,
				Limit:      args.Limit,
			})
			if err != nil {
				answer(c, messageId, jsonError(err.Error()))
				continue
			}
			resp, err := json.Marshal(records)
			if err != nil {
				answer(c, messageId, jsonError(err.Error()))
				continue
			}
			answer(c, messageId, resp)
		case "set":
			err := db.Set(args.Path, args.Record)
			if err != nil {
				answer(c, messageId, jsonError(err.Error()))
				continue
			}
			answer(c, messageId, jsonSuccess())
		case "merge":
			err := db.Merge(args.Path, args.Record)
			if err != nil {
				answer(c, messageId, jsonError(err.Error()))
				continue
			}
			answer(c, messageId, jsonSuccess())
		case "delete":
			err := db.Delete(args.Path, args.Rev)
			if err != nil {
				answer(c, messageId, jsonError(err.Error()))
				continue
			}
			answer(c, messageId, jsonSuccess())
		default:
			log.Error("ws unknown method.", "message", string(bmessage))
			answer(c, messageId, jsonError("unknown method "+method))
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

func answer(c *websocket.Conn, messageId []byte, response []byte) {
	body := append(messageId, ' ')
	body = append(body, response...)
	c.WriteMessage(1, body)
}
