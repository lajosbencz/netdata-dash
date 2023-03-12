package main

import (
	"strings"

	"github.com/gammazero/nexus/v3/wamp"
)

type authz struct {
	rootRole string
}

func isProtectedProcedure(name string) bool {
	return strings.HasPrefix(name, "wamp.")
}

func isProtectedTopic(name string) bool {
	return strings.HasPrefix(name, "wamp.")
}

func (a *authz) Authorize(sess *wamp.Session, msg wamp.Message) (bool, error) {
	role := wamp.OptionString(sess.Details, "authrole")
	// if the role is root, allow everything
	if role == a.rootRole && a.rootRole != "" {
		return true, nil
	}
	// for every other role, be restrictive
	if m, ok := msg.(*wamp.Call); ok {
		if !isProtectedProcedure(string(m.Procedure)) {
			return true, nil
		}
	}
	if m, ok := msg.(*wamp.Subscribe); ok {
		if !isProtectedTopic(string(m.Topic)) {
			return true, nil
		}
	}
	if _, ok := msg.(*wamp.Unsubscribe); ok {
		return true, nil
	}
	// if _, ok := msg.(*wamp.Publish); ok {
	// 	return false, nil
	// }
	// if _, ok := msg.(*wamp.Register); ok {
	// 	return false, nil
	// }
	return false, nil
}
