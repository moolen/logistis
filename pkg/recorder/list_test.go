package recorder

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moolen/logistis/pkg/store"
)

func TestFormatDiff(t *testing.T) {
	ev := &store.Event{
		ID:        uuid.New().String(),
		Group:     "apps",
		Kind:      "Deployment",
		Namespace: "kube-system",
		Name:      "coredns",
		Timestamp: time.Now(),
		Object: []byte(`{
"apiVersion": "v1",
"kind": "Pod",
"1": "1",
"2": "2",
"3": "3",
"metadata": {
  "labels": {
    "foo": "bar",
	"baz": "bang",
	"other": "ok"
  }},
"spec": {
  "1": "1",
  "2": "2",
  "3": "3",
  "containers": [
    {
        "name": "foo",
        "image": "example:1"
	}
  ]
}
}`),
		OldObject: []byte(`{
"apiVersion": "v1",
"kind": "Pod",
"1": "1",
"2": "2",
"3": "3",
"metadata": {
  "labels": {
    "foo": "bar",
	"baz": "shring",
	"dafuq": "shrizz",
	"dafuq": "shrizz",
	"dafuq": "shrizz"
  }},
"spec": {
"1": "1",
"2": "2",
"3": "3",
  "containers": [
    {
        "name": "foo",
        "image": "example:0.1"
	}
  ]
}
}`),
	}
	df, err := diffEvent(ev)
	if err != nil {
		t.Error(err)
	}
	t.Logf("raw diff: %s", df)
	df = formatDiff(df, 3)
	t.Logf("processed diff: %s", df)
	t.Logf("%#v", df)
	if df != "   \"kind\": \"Pod\",\n   \"metadata\": {\n     \"labels\": {\n\x1b[30;41m-      \"baz\": \"shring\",\x1b[0m\n\x1b[30;42m+      \"baz\": \"bang\",\x1b[0m\n\x1b[30;41m-      \"dafuq\": \"shrizz\",\x1b[0m\n       \"foo\": \"bar\"\n\x1b[30;42m+      \"other\": \"ok\"\x1b[0m\n     }\n   },\n   \"spec\": {\n[...]\n     \"3\": \"3\",\n     \"containers\": [\n       0: {\n\x1b[30;41m-        \"image\": \"example:0.1\",\x1b[0m\n\x1b[30;42m+        \"image\": \"example:1\",\x1b[0m\n         \"name\": \"foo\"\n       }\n     ]\n" {
		t.Fail()
	}
}
