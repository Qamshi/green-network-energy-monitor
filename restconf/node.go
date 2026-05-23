package restconf

import (
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/source"
)

func Node(mgmt *Server, ypath source.Opener) node.Node {
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (node.Node, error) {
			return nil, nil
		},
		OnField: func(r node.FieldRequest, hnd *node.ValueHandle) error {
			return nil
		},
	}
}