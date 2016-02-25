// Copyright 2016 Daniel Harrison. All Rights Reserved.

package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
)

func ExampleBuilder() {
	f := func(w *Builder) error {
		w.Add("baz", 7)
		return nil
	}

	g := func(w *ListBuilder) error {
		w.Add(1).Add(2).Add(3)
		return nil
	}

	h := func(w *Builder) error {
		w.AddObjectFunc("corge", f).
			AddObjectFunc("grault", func(w *Builder) error {
			w.AddListFunc("garply", g)
			return nil
		})
		return nil
	}

	var buf bytes.Buffer
	NewBuilder(&buf).
		Add("foo", "bar").
		AddObjectFunc("quz", f).
		AddListFunc("quux", g).
		AddObjectFunc("waldo", h).
		AddAll("1", "one", "2", "two").
		Close()
	fmt.Println(buf.String())
}

var benchLoad = 1000

func BenchmarkStdlib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := make(map[string]int)
		for i := 0; i < benchLoad; i++ {
			m[strconv.Itoa(i)] = i
		}
		bytes, err := json.Marshal(m)
		if err != nil {
			b.Fatal(err)
		}
		b.SetBytes(int64(len(bytes)))
	}
}

func BenchmarkBuilder(b *testing.B) {
	b.ReportAllocs()
	var buf bytes.Buffer
	for i := 0; i < b.N; i++ {
		j := NewBuilder(&buf)
		for i := 0; i < benchLoad; i++ {
			j.Add(strconv.Itoa(i), i)
		}
		j.Close()
		if j.Err != nil {
			b.Fatal(j.Err)
		}
		b.SetBytes(int64(len(buf.Bytes())))
		buf.Reset()
	}
}

func f(w *Builder) error {
	w.Add("baz", 7)
	return nil
}

func g(w *ListBuilder) error {
	w.Add(1).Add(2).Add(3)
	return nil
}

func h(w *Builder) error {
	w.AddObjectFunc("corge", f).
		AddObjectFunc("grault", func(w *Builder) error {
		w.AddListFunc("garply", g)
		return nil
	})
	return nil
}

var jsonTests = []struct {
	out string
	fn  func(*Builder)
}{
	{`{}`, func(j *Builder) {}},

	{`{"foo":"bar"}`, func(j *Builder) { j.Add("foo", "bar") }},
	{`{"foo":7}`, func(j *Builder) { j.Add("foo", 7) }},
	{`{"foo":false}`, func(j *Builder) { j.Add("foo", false) }},
	{`{"foo":6.2}`, func(j *Builder) { j.Add("foo", 6.2) }},

	{`{"foo":[1,2]}`, func(j *Builder) { j.Add("foo", []int{1, 2}) }},
	{`{"foo":["bar","baz"]}`, func(j *Builder) { j.Add("foo", []string{"bar", "baz"}) }},

	{`{"foo":{"a":7,"b":"bar"}}`, func(j *Builder) {
		j.Add("foo", struct {
			A int    `json:"a"`
			B string `json:"b"`
		}{7, "bar"})
	}},
	{`{"foo":[{"a":7,"b":"bar"},{"a":1,"b":"baz"}]}`, func(j *Builder) {
		j.Add("foo", []struct {
			A int    `json:"a"`
			B string `json:"b"`
		}{{7, "bar"}, {1, "baz"}})
	}},

	{`{"foo":{"bar":7}}`, func(j *Builder) { s := j.AddObject("foo"); s.Add("bar", 7); s.Close() }},
	{`{"foo":["bar",7]}`, func(j *Builder) { s := j.AddList("foo"); s.AddAll("bar", 7); s.Close() }},

	{`{"foo":{"baz":7}}`, func(j *Builder) { j.AddObjectFunc("foo", f) }},
	{`{"foo":[1,2,3]}`, func(j *Builder) { j.AddListFunc("foo", g) }},
	{`{"foo":{"corge":{"baz":7},"grault":{"garply":[1,2,3]}}}`, func(j *Builder) { j.AddObjectFunc("foo", h) }},
}

func TestBuilder(t *testing.T) {
	for i, jsonTest := range jsonTests {
		var buf bytes.Buffer
		j := NewBuilder(&buf)
		jsonTest.fn(j)
		j.Close()
		if j.Err != nil {
			t.Errorf("%d Unexpected error <%d>", j.Err.Error())
		}
		if got := buf.String(); got != jsonTest.out {
			t.Errorf("%d have <%s> want <%s>", i, got, jsonTest.out)
		}
	}
}

var jsonListTests = []struct {
	out string
	fn  func(*ListBuilder)
}{
	{`[]`, func(j *ListBuilder) {}},

	{`["foo",7]`, func(j *ListBuilder) { j.Add("foo").Add(7) }},
	{`[false,6.2]`, func(j *ListBuilder) { j.Add(false).Add(6.2) }},
	{`["foo",7]`, func(j *ListBuilder) { j.AddAll("foo", 7) }},

	{`[{"a":1,"b":"baz"}]`, func(j *ListBuilder) {
		j.Add(struct {
			A int    `json:"a"`
			B string `json:"b"`
		}{1, "baz"})
	}},
	{`[[{"a":7,"b":"bar"},{"a":1,"b":"baz"}]]`, func(j *ListBuilder) {
		j.Add([]struct {
			A int    `json:"a"`
			B string `json:"b"`
		}{{7, "bar"}, {1, "baz"}})
	}},

	{`[{"baz":7}]`, func(j *ListBuilder) { j.AddObjectFunc(f) }},
	{`[[1,2,3]]`, func(j *ListBuilder) { j.AddListFunc(g) }},
	{`[{"corge":{"baz":7},"grault":{"garply":[1,2,3]}}]`, func(j *ListBuilder) { j.AddObjectFunc(h) }},
}

func TestListBuilder(t *testing.T) {
	for i, jsonListTest := range jsonListTests {
		var buf bytes.Buffer
		j := NewListBuilder(&buf)
		jsonListTest.fn(j)
		j.Close()
		if j.Err != nil {
			t.Errorf("%d Unexpected error <%d>", j.Err.Error())
		}
		if got := buf.String(); got != jsonListTest.out {
			t.Errorf("%d have <%s> want <%s>", i, got, jsonListTest.out)
		}
	}
}

func TestUnclosedSubBuilder(t *testing.T) {
	var buf bytes.Buffer
	j := NewBuilder(&buf).Add("1", 1)
	j.AddObject("2").Add("3", 3)
	j.Add("4", 4)
	if j.Err == nil {
		t.Error("Expected error")
	}
}
