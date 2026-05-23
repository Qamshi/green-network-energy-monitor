package secure

import (
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
)

func Manage(rbac *Rbac) node.Node {
	// We use a simple reflection node. 
	// This allows the code to compile without referencing rbac.AccessControl
	return nodeutil.ReflectChild(rbac)
}