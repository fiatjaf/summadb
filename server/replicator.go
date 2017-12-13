package server

import (
	"errors"
	"time"

	"github.com/summadb/summadb/database"
	"github.com/summadb/summadb/types"
	"golang.org/x/net/websocket"
)

const (
	WAITING_REVS = iota
	SENDING_REVS
	SENDING_DIFF_AND_VALUES
	WAITING_DIFF_AND_VALUES
	WAITING_DIFF
	WAITING_VALUES
	SENDING_VALUES
)

func acceptReplication(
	c *websocket.Conn,
	db *database.SummaDB,
	path types.Path,
	replicationId string,
) error {

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(time.Second * 60)
		timeout <- true
	}()

	rpl := database.Replicator(db, path)
	step := SENDING_REVS

	rsend := func(tag string, val []byte) { send(c, []byte(tag), []byte(replicationId), val) }

	// fetch revs and send them -- meaning we've accepted this replication attempt
	allrevs, err := rpl.AllRevs()
	if err != nil {
		return err
	}
	for _, pathrev := range allrevs {
		j, _ := pathrev.MarshalJSON()
		rsend("pathrev", j)
	}
	rsend("endpathrev")
	step = WAITING_DIFF_AND_VALUES

	var diff []string       // a list of paths
	var values []types.Tree // a list of trees without children, only leafs

	for {
		received := make(chan []byte, 1)
		go func() {
			_, bmessage, _ := c.ReadMessage()
			received <- bmessage
		}()
		select {
		case <-timeout:
			return errors.New("timed out.")
		case bmessage := <-received:
			method, messageId, body, err := parseMessage(bmessage)
			if err != nil {
				log.Error("ws parsing error.", "message", string(bmessage), "err", err)
				continue
			}

			if messageId != replicationId {
				log.Info("unexpected messageId while replicating.", "message", string(bmessage))
				continue
			}

			switch method {
			case "diff":
				diff = append(diff, string(body))
			case "enddiff":
				if step == WAITING_DIFF_AND_VALUES {
					step = WAITING_VALUES
				} else if step == WAITING_DIFF {
					break
				}
			case "value":
				t := &Tree{}
				t.UnmarshalJSON(body)
				values = append(values, t)
			case "endvalue":
				if step == WAITING_DIFF_AND_VALUES {
					step = WAITING_DIFF
				} else if step == WAITING_VALUES {
					break
				}
			default:
				log.Info("unexpected method while replicating.", "message", string(bmessage))
				continue
			}
		}
	}

	step = SENDING_VALUES

	// fetch values requested by the other db and send them
	// ...

	// apply changes received
	// ...

	return nil
}
