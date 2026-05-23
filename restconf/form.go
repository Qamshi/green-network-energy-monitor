package restconf

import (
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/freeconf/yang/meta"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/val"
)

func isMultiPartForm(hdrs http.Header) bool {
	return strings.HasPrefix(hdrs.Get("Content-Type"), "multipart/form-data")
}

func formNode(req *http.Request) (node.Node, error) {
	err := req.ParseMultipartForm(10000)
	if err != nil {
		return nil, err
	}
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			entry, found := req.MultipartForm.File[r.Meta.Ident()]
			if !found || len(entry) == 0 {
				return nil, nil
			}
			if meta.IsList(r.Meta) {
				return formListNode(entry), nil
			}
			return nil, nil
		},
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			sval := req.FormValue(r.Meta.Ident())
			if sval != "" {
				var err error
				hnd.Val, err = node.NewValue(r.Meta.Type(), sval)
				return err
			}
			return nil
		},
	}, nil
}

func formListNode(files []*multipart.FileHeader) node.Node {
	return &nodeutil.Basic{
		OnNext: func(r node.ListRequest) (node.Node, []val.Value, error) {
			if r.Row >= len(files) {
				return nil, nil, nil
			}
			f, _ := files[r.Row].Open()
			defer f.Close()
			return nodeutil.ReadJSONIO(f), nil, nil
		},
	}
}