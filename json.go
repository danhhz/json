// Copyright 2016 Daniel Harrison. All Rights Reserved.

// Package json implements streaming encoding of JSON objects.
//
// All of the heavy lifting is delegated to the stdlib encoding/json, so look
// there for details on the encoding: https://golang.org/pkg/encoding/json/
package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type writerState int

const startState = 0
const openedState = 1
const closedState = 2

var openBraceBytes = []byte{'{'}
var closeBraceBytes = []byte{'}'}
var openBracketBytes = []byte{'['}
var closeBracketBytes = []byte{']'}
var colonBytes = []byte{':'}
var commaBytes = []byte{','}

// BuilderFunc represents the creation of a JSON object.
type BuilderFunc func(*Builder) error

// ListBuilderFunc represents the creation of a JSON list.
type ListBuilderFunc func(*ListBuilder) error

// A Builder writes JSON objects to an output stream, without needing it all to
// be in memory at once.
type Builder struct {
	state writerState
	w     io.Writer
	e     encoder
	subB  builderCommon
	Err   error
}

// NewBuilder returns a new encoder that writes to w.
func NewBuilder(w io.Writer) *Builder {
	b := &Builder{startState, w, newEncoder(w), nil, nil}
	b.init()
	return b
}

func (b *Builder) init() {
	if b.state != startState {
		b.Err = errors.New("Builder init'd after being mutated")
	}
	b.write(openBraceBytes)
}

func (b *Builder) write(x []byte) {
	if b.Err == nil {
		_, b.Err = b.w.Write(x)
	}
}

func (b *Builder) checkSub() error {
	if b.Err == nil && b.subB != nil {
		if err := b.subB.err(); err != nil {
			b.Err = err
		} else if !b.subB.closed() {
			b.Err = errors.New("A sub-Builder was not closed")
		}
		b.subB = nil
	}
	return b.Err
}

func (b *Builder) preadd(key string) error {
	if b.state == closedState {
		b.Err = errors.New("Builder mutated after Close()")
	}
	if err := b.checkSub(); err != nil {
		return err
	}

	if b.state == startState {
		b.state = openedState
	} else {
		b.write(commaBytes)
	}

	b.Err = b.e.encode(key)
	b.write(colonBytes)
	return b.Err
}

// Add emits a single key value pair to the stream.
func (b *Builder) Add(key string, value interface{}) *Builder {
	if b.preadd(key) != nil {
		return b
	}

	b.Err = b.e.encode(value)
	return b
}

// AddAll emits many key value pairs to the stream.
//
// The args represent a key, then a value, then a key, and so on. There must be
// an even number of args and the keys must all be strings.
func (b *Builder) AddAll(args ...interface{}) *Builder {
	if len(args)%2 != 0 {
		b.Err = errors.New("AddAll takes an even number of args")
		return b
	}
	for i := 0; i < len(args) && b.Err == nil; i += 2 {
		if key, ok := args[i].(string); ok {
			b.Add(key, args[i+1])
		} else {
			b.Err = fmt.Errorf("Arg %d was not a string %s: %s", i, reflect.TypeOf(args[i]), args[i])
			return b
		}
	}
	return b
}

// AddObject returns a builder for a JSON object value with the given key.
//
// Close() must be called on the sub-object before using this builder again.
func (b *Builder) AddObject(key string) *Builder {
	b.preadd(key)
	subB := &Builder{0, b.w, b.e, nil, nil}
	subB.init()
	b.subB = subB
	return subB
}

// AddList returns a builder for a JSON list value with the given key.
//
// Close() must be called on the sub-list before using this builder again.
func (b *Builder) AddList(key string) *ListBuilder {
	b.preadd(key)
	subB := &ListBuilder{0, b.w, b.e, nil, nil}
	subB.init()
	b.subB = subB
	return subB
}

// AddObjectFunc emits a JSON object value (computed from f) with the given key.
func (b *Builder) AddObjectFunc(key string, f BuilderFunc) *Builder {
	if b.preadd(key) != nil {
		return b
	}

	subB := Builder{0, b.w, b.e, nil, nil}
	subB.init()
	b.Err = f(&subB)
	if b.Err == nil {
		b.Err = subB.Err
	}
	subB.Close()
	return b
}

// AddListFunc emits a JSON list value (computed from f) with the given key.
func (b *Builder) AddListFunc(key string, f ListBuilderFunc) *Builder {
	if b.preadd(key) != nil {
		return b
	}

	subB := ListBuilder{0, b.w, b.e, nil, nil}
	subB.init()
	b.Err = f(&subB)
	if b.Err == nil {
		b.Err = subB.Err
	}
	subB.Close()
	return b
}

// Close finalizes this JSON object and must be called for it to be complete.
//
// After Close is called, nothing else on this object may be called except Err.
func (b *Builder) Close() *Builder {
	if b.state == closedState {
		b.Err = errors.New("ListBuilder mutated after Close()")
		return b
	}
	if b.checkSub() != nil {
		return b
	}

	b.write(closeBraceBytes)
	b.state = closedState
	return b
}

func (b *Builder) closed() bool {
	return b.state == closedState
}

func (b *Builder) err() error {
	return b.Err
}

// A ListBuilder writes JSON lists to an output stream, without needing it all
// to be in memory at once.
type ListBuilder struct {
	state writerState
	w     io.Writer
	e     encoder
	subB  builderCommon
	Err   error
}

// NewListBuilder returns a new encoder that writes to w.
func NewListBuilder(w io.Writer) *ListBuilder {
	b := &ListBuilder{startState, w, newEncoder(w), nil, nil}
	b.init()
	return b
}

func (b *ListBuilder) init() {
	if b.state != startState {
		b.Err = errors.New("ListBuilder init'd after being mutated")
	}
	b.write(openBracketBytes)
}

func (b *ListBuilder) write(x []byte) {
	if b.Err == nil {
		_, b.Err = b.w.Write(x)
	}
}

func (b *ListBuilder) checkSub() error {
	if b.Err == nil && b.subB != nil {
		if err := b.subB.err(); err != nil {
			b.Err = err
		} else if !b.subB.closed() {
			b.Err = errors.New("A sub-Builder was not closed")
		}
		b.subB = nil
	}
	return b.Err
}

func (b *ListBuilder) preadd() error {
	if b.state == closedState {
		b.Err = errors.New("ListWriter mutated after Close()")
	}
	if err := b.checkSub(); err != nil {
		return err
	}

	if b.state == startState {
		b.state = openedState
	} else {
		b.write(commaBytes)
	}
	return b.Err
}

// Add emits a single value to the stream.
func (b *ListBuilder) Add(value interface{}) *ListBuilder {
	if b.preadd() != nil {
		return b
	}

	b.Err = b.e.encode(value)
	return b
}

// AddAll emits many values to the stream.
func (b *ListBuilder) AddAll(args ...interface{}) *ListBuilder {
	for i := 0; i < len(args) && b.Err == nil; i++ {
		b.Add(args[i])
	}
	return b
}

// AddObject returns a builder for a JSON object value inserted as the next
// element.
//
// Close() must be called on the sub-object before using this builder again.
func (b *ListBuilder) AddObject() *Builder {
	if b.preadd() != nil {
		return nil
	}
	subB := &Builder{0, b.w, b.e, nil, nil}
	subB.init()
	return subB
}

// AddList returns a builder for a JSON list value inserted as the next element.
//
// Close() must be called on the sub-list before using this builder again.
func (b *ListBuilder) AddList() *ListBuilder {
	b.preadd()
	subB := &ListBuilder{0, b.w, b.e, nil, nil}
	subB.init()
	b.subB = subB
	return subB
}

// AddObjectFunc emits a JSON object value (computed from f) as the next
// element.
func (b *ListBuilder) AddObjectFunc(f BuilderFunc) *ListBuilder {
	if b.preadd() != nil {
		return b
	}

	subB := Builder{0, b.w, b.e, nil, nil}
	subB.init()
	b.Err = f(&subB)
	if b.Err == nil {
		b.Err = subB.Err
	}
	subB.Close()
	return b
}

// AddListFunc emits a JSON list value (computed from f) as the next element.
func (b *ListBuilder) AddListFunc(f ListBuilderFunc) *ListBuilder {
	if b.preadd() != nil {
		return b
	}

	subB := ListBuilder{0, b.w, b.e, nil, nil}
	subB.init()
	b.Err = f(&subB)
	if b.Err == nil {
		b.Err = subB.Err
	}
	subB.Close()
	return b
}

// Close finalizes this JSON object and must be called for it to be complete.
//
// After Close is called, nothing else on this object may be called except Err.
func (b *ListBuilder) Close() *ListBuilder {
	if b.state == closedState {
		b.Err = errors.New("ListWriter mutated after Close()")
		return b
	}
	if b.checkSub() != nil {
		return b
	}

	b.write(closeBracketBytes)
	b.state = closedState
	return b
}

func (b *ListBuilder) closed() bool {
	return b.state == closedState
}

func (b *ListBuilder) err() error {
	return b.Err
}

type builderCommon interface {
	closed() bool
	err() error
}

type encoder interface {
	encode(arg interface{}) error
}

func newEncoder(w io.Writer) encoder {
	return basicEncoder{w}
	// TODO(dan): This removes a ton of garbage overhead (enough to make it faster
	// than the stdlib in benchmarks), but the trimTrailingNewlineWriter is
	// probably too likely to be broken by stdlib changes. Make a decision on
	// which to use and delete the encoder abstraction.
	// return newStreamingEncoder(w)
}

type basicEncoder struct {
	io.Writer
}

func (b basicEncoder) encode(arg interface{}) error {
	bytes, err := json.Marshal(arg)
	if err != nil {
		return nil
	}
	_, err = b.Write(bytes)
	return err
}

type streamingEncoder struct {
	*json.Encoder
}

func newStreamingEncoder(w io.Writer) streamingEncoder {
	return streamingEncoder{json.NewEncoder(trimTrailingNewlineWriter{w})}
}
func (b streamingEncoder) encode(arg interface{}) error {
	return b.Encode(arg)
}

type trimTrailingNewlineWriter struct {
	w io.Writer
}

func (h trimTrailingNewlineWriter) Write(p []byte) (n int, err error) {
	if p[len(p)-1] == '\n' {
		return h.w.Write(p[0 : len(p)-1])
	}
	return h.w.Write(p)
}
