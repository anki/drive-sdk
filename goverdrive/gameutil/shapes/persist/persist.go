// Copyright 2017 Anki, Inc.
// Author: gwenz@anki.com

// Package persist helps manage game shapes that persist for >1 tick.
package persist

import (
	_ "fmt"

	"github.com/anki/goverdrive/phys"
	"github.com/anki/goverdrive/viz"
)

// shapeNode is a linked-list node for a persistent shape. A linked list is used
// because it has O(1) add/remove time.
type shapeNode struct {
	shape   *viz.GameShape
	tExpire phys.SimTime
	next    *shapeNode
}

// Manager manages persistent shapes
type Manager struct {
	head    *shapeNode
	tail    *shapeNode
	numElem uint
}

// New returns a manager with no shapes
func New() *Manager {
	return &Manager{head: nil, tail: nil, numElem: 0}
}

// Add adds a GameShape to the persistence manager. The duration to persist is
// in milliseconds (for convenience).
func (m *Manager) Add(now phys.SimTime, msDur uint, shape *viz.GameShape) {
	node := shapeNode{
		shape:   shape,
		tExpire: now + (phys.SimTime(msDur) * phys.SimMillisecond),
		next:    nil,
	}

	// add to end of linked list
	if m.tail == nil { // empty
		m.head = &node
		m.tail = &node
	} else {
		m.tail.next = &node
		m.tail = &node
	}
	m.numElem++
}

// Update removes all GameShape objects that have expired, and returns a list of
// shapes that have not yet expired.
func (m *Manager) Update(now phys.SimTime) *[]*viz.GameShape {
	vizShapes := make([]*viz.GameShape, 0, m.numElem)
	var prev *shapeNode = nil
	for cur := m.head; cur != nil; {
		if cur.tExpire > now {
			vizShapes = append(vizShapes, cur.shape)
			prev = cur
		} else {
			// shape expired => do not draw, and remove it from linked list
			// no change to prev
			m.numElem--
			if cur.next == nil {
				m.tail = prev
			}
			if prev == nil {
				m.head = cur.next
			} else {
				prev.next = cur.next
			}
		}
		cur = cur.next
	}
	return &vizShapes
}
