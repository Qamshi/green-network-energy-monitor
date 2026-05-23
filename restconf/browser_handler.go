package restconf

import (
	"context"
	"net/http"

	"github.com/freeconf/yang/meta"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
)

type browserHandler struct {
	browser *node.Browser
}

const EventTimeFormat = "2006-01-02T15:04:05-07:00"

type ProxyContextKey string

var RemoteIpAddressKey = ProxyContextKey("FC_REMOTE_IP")

type MimeType string

const (
	YangDataJsonMimeType1 = MimeType("application/yang-data+json")
	PlainJsonMimeType     = MimeType("application/json")
	TextStreamMimeType    = MimeType("text/event-stream")
)

const SimplifiedComplianceParam = "simplified"

type ComplianceContextKeyType string

var ComplianceContextKey = ComplianceContextKeyType("RESTCONF_COMPLIANCE")

func (hndlr *browserHandler) ServeHTTP(compliance ComplianceOptions, ctx context.Context, w http.ResponseWriter, r *http.Request, endpointId int) {
	var err error
	sel := hndlr.browser.RootWithContext(ctx)
	defer sel.Release()
	acceptType := MimeType(r.Header.Get("Accept"))

	if target, findErr := sel.Find(r.URL.EscapedPath()); findErr == nil {
		if target == nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		defer target.Release()

		switch r.Method {
		case "GET":
			w.Header().Set("Content-Type", string(PlainJsonMimeType))
			err = target.InsertInto((&nodeutil.JSONWtr{Out: w}).Node())
		case "POST":
			payload := nodeutil.ReadJSONIO(r.Body)
			if meta.IsAction(target.Meta()) {
				_, err = target.Action(payload)
			} else {
				editable, _ := target.Constrain("content=config")
				err = editable.InsertFrom(payload)
			}
		}
	}

	if err != nil {
		handleErr(compliance, err, r, w, acceptType)
	}
}

func (m MimeType) IsXml() bool { return false }
func (m MimeType) IsRfc() bool { return false }