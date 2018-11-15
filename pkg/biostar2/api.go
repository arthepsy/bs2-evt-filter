package biostar2

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func NewAPI(url string, username string, password string) *API {
	b := &API{url: url, username: username, password: password}
	b.authorized = false
	b.sessionID = ""
	return b
}

func (b *API) SessionID() string {
	return b.sessionID
}

func (r *API) log(f string, v ...interface{}) {
	log.Printf("[b2api] "+f, v...)
}

func (b *API) doHttpRequest(req *http.Request) (*http.Response, bool) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		b.log("http.client failed with: %v", err)
		return nil, false
	}
	return resp, true
}

func (b *API) getHttpRequest(apiUrl string, data []byte) (*http.Request, context.CancelFunc) {
	url := fmt.Sprintf("%s%s", strings.TrimRight(b.url, "/"), apiUrl)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		b.log("http.request failed with: %v", err)
		return nil, nil
	}
	req.Header.Set("Content-Type", "application/json")
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*7)
	req = req.WithContext(ctx)
	return req, cancel
}

func (b *API) logHttpFailure(resp *http.Response) {
	body, _ := ioutil.ReadAll(resp.Body)
	r, ok := ParseResponse(body)
	s := strings.TrimSpace(resp.Status)
	if ok {
		b.log("error, status: %s, code: %s, msg: %s", s, r.Code, r.Message)
	} else {
		b.log("error, status: %s", s)
	}
}
func (b *API) Auth() bool {
	b.sessionID = ""
	uw := &UserAuthWrapper{UserAuth{Username: b.username, Password: b.password}}
	j, err := json.Marshal(uw)
	if err != nil {
		b.log("json.marshal failed with: %v", err)
		return false
	}
	req, cancel := b.getHttpRequest("/api/login", j)
	if cancel == nil {
		return false
	}
	defer cancel()
	resp, ok := b.doHttpRequest(req)
	if !ok {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		hv, ok := resp.Header["Bs-Session-Id"]
		if ok && len(hv) > 0 {
			b.sessionID = hv[0]
			b.log("session retrieved successfuly\n")
			return true
		}
		b.log("session id not found")
		return false
	} else {
		b.logHttpFailure(resp)
		return false
	}
}

func (b *API) StartEvents() bool {
	req, cancel := b.getHttpRequest("/api/events/start", []byte(""))
	if cancel == nil {
		return false
	}
	defer cancel()
	req.Header.Set("bs-session-id", b.sessionID)
	resp, ok := b.doHttpRequest(req)
	if !ok {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		b.log("events started")
		return true
	} else {
		b.logHttpFailure(resp)
		return false
	}
}

func (b *API) wslog(f string, v ...interface{}) {
	log.Printf("[b2wsapi] "+f, v...)
}

func (b *API) WebSocket(recv chan<- []byte, send <-chan []byte, done <-chan struct{}) {
	url := fmt.Sprintf("%s/wsapi", strings.TrimRight(b.url, "/"))
	url = strings.Replace(url, "http", "ws", 1)

	b.wslog("connecting to %s", url)
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c, _, err := dialer.Dial(url, nil)
	if err != nil {
		b.wslog("connection error: %v", err)
		return
	}
	b.wslog("connected to %s\n", url)
	defer c.Close()

	rdone := make(chan struct{})
	go func() {
		defer close(rdone)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				b.wslog("read error: %v", err)
				return
			}
			message = bytes.TrimSpace(message)
			recv <- message
		}
	}()

	for {
		select {
		case msg := <-send:
			c.WriteMessage(websocket.TextMessage, msg)
		case <-rdone:
			b.wslog("disconnect (read)")
			return
		case <-done:
			b.wslog("disconnect (done)")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				b.wslog("write error: %v", err)
				return
			}
			select {
			case <-rdone:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
