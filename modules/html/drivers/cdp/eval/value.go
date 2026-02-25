package eval

import cdpruntime "github.com/mafredri/cdp/protocol/runtime"

type RemoteValue interface {
	RemoteID() cdpruntime.RemoteObjectID
}
