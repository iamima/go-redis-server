package livedb

import (
	"testing"
	"bytes"
)

var testData map[string]interface{} = map[string]interface{}{
	"foo":	"bar",
	"h":	map[string]string {
		"ga": "bu",
	},
}



func TestNewFrom(t *testing.T) {
	if !NewFrom(testData).FieldEquals("foo", "bar") {
		t.Fatalf("NewFrom")
	}
}

func TestSET(t *testing.T) {
	db := New()
	db.SET("foo", "bar")
	if !db.FieldEquals("foo", "bar") {
		t.Fatalf("%s != %s", db.data["foo"], "bar")
	}
}

func TestApplySET(t *testing.T) {
	db := New()
	_, err := db.Apply("SET", "foo", "bar")
	if err != nil {
		t.Fatal(err)
	}
	if !db.FieldEquals("foo", "bar") {
		t.Fatal("GET/SET")
	}
}

func testGETExists(t *testing.T) {
	db := NewFrom(map[string]interface{}{"foo":"bar"})
	if value, err := db.GET("foo"); err != nil {
		t.Fatal(err)
	} else if !stringEqual(value, "bar") {
		t.Fatalf("GET returned %s instead of bar", value)
	}
}

func TestGETNonExist(t *testing.T) {
	db := New()
	if value, err := db.GET("foo"); err != nil {
		t.Fatal("GET on non-existing key should return nil, not an error")
	} else if value != nil {
		t.Fatal("GET on non-existing key should return nil")
	}
}

func TestGETonHash(t *testing.T) {
	db := New()
	db.HSET("foo", "k", "hello")
	if _, err := db.GET("foo"); err == nil {
		t.Fatal("GET on hash should return an error")
	}
}

func TestHSET(t *testing.T) {
	db := New()
	db.HSET("foo", "k", "hello")
	if !db.Equals(NewFromJSON([]byte("{\"foo\": {\"k\": \"hello\"}}"))) {
		t.Fatalf("Wrong db state after HSET: %#v\n", db.data)
	}
}

func TestHGET(t *testing.T) {
	db := NewFrom(testData)
	if value, err := db.HGET("h", "ga"); err != nil {
		t.Fatal(err)
	} else if !stringEqual(value, "bu") {
		t.Fatalf("HGET returned %s intead of %s", value, "bu")
	}
}

func TestHGETNonExistingHash(t *testing.T) {
	db := New()
	if value, err := db.HGET("foo", "k"); err != nil {
		t.Fatal("HGET on non-existing hash should return null")
	} else if value != nil {
		t.Fatal("HGET on non-existing hash should return null")
	}
}

func TestHGETNonExistingKey(t *testing.T) {
	db := New()
	if value, err := db.HGET("foo", "k"); err != nil {
		t.Fatal(err)
	} else if value != nil {
		t.Fatal("HGET on non-existing key should return null")
	}
}

func TestLoadJSON1(t *testing.T) {
	db := New()
	if n, err := db.LoadJSON([]byte("{\"foo\": \"bar\"}")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatalf("LoadJSON returned %d instead of 1", n)
	}
	if value, err := db.GET("foo"); err != nil {
		t.Fatal(err)
	} else if value == nil {
		t.Fatalf("LoadJSON didn't store a value")
	} else if *value != "bar" {
		t.Fatalf("LoadJSON stored the wrong value (%s)", *value)
	}
}

func TestLoadJSON2(t *testing.T) {
	db := New()
	if n, err := db.LoadJSON([]byte("{\"foo\": {\"ga\": \"bu\"}}")); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatalf("LoadJSON returned %d instead of 1", n)
	}
	if !db.FieldEquals("foo", map[string]string{"ga": "bu"}) {
		t.Fatalf("Wrong DB state after LoadJSON: %#v\n", db.data)
	}
}

func TestReplicateStart(t *testing.T) {
	wire := new(bytes.Buffer)
	in := NewFrom(testData)
	in.ReplicateTo(wire)
	out := New()
	if _, err := out.ReplicateFrom(wire); err != nil {
		t.Fatal(err)
	}
	if !in.Equals(out) {
		t.Fatalf("%s != %s", in, out)
	}
}

func TestReplicateFromSET(t *testing.T) {
	var wire bytes.Buffer
	New().ReplicateTo(&wire)
	wire.WriteString("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")
	db := New()
	if n, err := db.ReplicateFrom(&wire); err != nil {
		t.Fatal(err)
	} else if n != 1 {
		t.Fatalf("Replicate returned %d instead of 1", n)
	}
}

func TestReplicateWrongArgSize(t *testing.T) {
	db := New()
	input := new(bytes.Buffer)
	New().ReplicateTo(input)
	input.WriteString("*3\r\n$3\r\nSET\r\n$4\r\nfoo\r\n$3\r\nbar\r\n")
	if _, err := db.ReplicateFrom(input); err == nil {
		t.Fatalf("Wrong command in replication stream should trigger an error")
	}
}

func TestNewDump(t *testing.T) {
	in := NewFrom(testData)
	dump := NewDump(in.data)
	if !stringEqual(dump.Strings["foo"], "bar") {
		t.Fatalf("Dump didn't preserve string")
	}
	if !hashEqual(dump.Hashes["h"], map[string]string{"ga":"bu"}) {
		t.Fatal("Dump didn't preserve hash")
	}
}

func TestDumpData(t *testing.T) {
	in := NewFrom(testData)
	dump := NewDump(in.data)
	data := dump.Data()
	if !stringEqual(data["foo"], "bar") {
		t.Fatalf("Dump didn't preserve string")
	}
	if !hashEqual(data["h"].(map[string]string), map[string]string{"ga":"bu"}) {
		t.Fatal("Dump didn't preserve hash")
	}
}

func TestDump(t *testing.T) {
	in := NewFrom(testData)
	dump := NewDump(in.data)
	out := NewFrom(dump.Data())
	t.Logf("--> %#v\n", *(dump.Data()["foo"].(*string)))
	if !out.FieldEquals("foo", "bar") {
		t.Errorf("Dump didn't preserve string: %s", out.data["foo"])
	}
	if !out.FieldEquals("h", map[string]string{"ga":"bu"}) {
		t.Errorf("Dump didn't preserve hash")
	}
}

func TestLoadData(t *testing.T) {
	out := NewFrom(testData)
	if !out.FieldEquals("foo", "bar") {
		t.Errorf("NewFrom didn't preserve string: %s", out.data["foo"])
	}
	if !out.FieldEquals("h", map[string]string{"ga":"bu"}) {
		t.Errorf("NewFrom didn't preserve hash")
	}
}
