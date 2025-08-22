package argsparser

import (
	"testing"
)

type setStruct struct {
	Key   string `arg:"pos:1"`
	Value string `arg:"pos:2"`

	GET bool

	SetKey struct {
		NX bool
		XX bool
	} `arg:"enum"`

	Expire struct {
		Key     string `arg:"enum-key"`
		EX      int    `arg:"auto"`
		PX      int    `arg:",unimplemented"`
		EXAT    int    `arg:"enum-value,unimplemented"`
		PXAT    int
		KEEPTTL bool
	} `arg:"enum"`
}

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Error("expect no error but got:", err)
	}
}

func expectEqual[T comparable](t *testing.T, expected, actual T) {
	if expected != actual {
		t.Errorf("expect equal\nexepcted=%+v\n  actual=%+v", expected, actual)
	}
}

func expectNil[T any](t *testing.T, v *T) {
	if v != nil {
		t.Error("expect nil but got value")
	}
}

func expectNoNilEqual[T comparable](t *testing.T, expected T, v *T) {
	if v == nil {
		t.Error("expect no nil but got value")
		return
	}
	expectEqual(t, expected, *v)
}

func Test_extractTag(t *testing.T) {
	_, err := extractTag[setStruct]()
	expectNoError(t, err)
}

func Test_Parse(t *testing.T) {
	args := []string{"SET", "v_key", "k_value", "GET", "NX", "PX", "10000"}

	c, err := Parse[setStruct](args)

	expected := setStruct{}
	expected.Key = "v_key"
	expected.Value = "k_value"
	expected.GET = true
	expected.SetKey.NX = true
	expected.Expire.Key = "PX"
	expected.Expire.PX = 10000

	expectNoError(t, err)
	expectEqual(t, expected, c)
}

func Test_ParseVariadic(t *testing.T) {
	args := []string{"RPUSH", "key", "v1", "v2", "v3", "v4"}

	type rpush struct {
		Key    string   `arg:"pos:1"`
		Values []string `arg:"pos:2,variadic"`
	}

	c, err := Parse[rpush](args)

	vs := []string{"v1", "v2", "v3", "v4"}

	expectNoError(t, err)
	expectEqual(t, "key", c.Key)
	expectEqual(t, len(vs), len(c.Values))

	t.Log(vs, c.Values)
}

func Test_ParseOptionalPositionPointer(t *testing.T) {
	type lpop struct {
		Key   string `arg:"pos:1"`
		Count *int   `arg:"pos:2,optional"`
	}

	c1, err1 := Parse[lpop]([]string{"LPOP", "key_1"})

	expectNoError(t, err1)
	expectEqual(t, "key_1", c1.Key)
	expectNil(t, c1.Count)

	c2, err2 := Parse[lpop]([]string{"LPOP", "key_2", "12"})

	expectNoError(t, err2)
	expectEqual(t, "key_2", c2.Key)
	expectNoNilEqual(t, 12, c2.Count)
}
