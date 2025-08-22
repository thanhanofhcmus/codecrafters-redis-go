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

func expectNoError(t *testing.T, err error) bool {
	if err != nil {
		t.Error("expect no error but got:", err)
		return false
	}
	return true
}

func expectEqual[T comparable](t *testing.T, expected, actual T) bool {
	if expected != actual {
		t.Errorf("expect equal\nexepcted=%+v\n  actual=%+v", expected, actual)
		return false
	}
	return true
}

func expectEqualSlice[T comparable](t *testing.T, expected, actual []T) bool {
	if len(expected) != len(actual) {
		t.Errorf("expect array length equal\nexepcted=%+v\n  actual=%+v", len(expected), len(actual))
		return false
	}
	for idx := range len(expected) {
		e, a := expected[idx], actual[idx]
		if !expectEqual(t, e, a) {
			return false
		}
	}
	return true
}

func expectNil[T any](t *testing.T, v *T) bool {
	if v != nil {
		t.Error("expect nil but got value")
		return false
	}
	return true
}

func expectNoNilEqual[T comparable](t *testing.T, expected T, v *T) bool {
	if v == nil {
		t.Error("expect no nil but got value")
		return false
	}
	return expectEqual(t, expected, *v)
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

	expectNoError(t, err)
	expectEqual(t, "key", c.Key)
	expectEqualSlice(t, []string{"v1", "v2", "v3", "v4"}, c.Values)
}

func Test_ParseVariadicRequiredAfter(t *testing.T) {
	args := []string{"BLPOP", "l1", "l2", "l3", "120", "another"}

	type blpop struct {
		Key      string   `arg:"pos:1"`
		RestKeys []string `arg:"pos:2,variadic"`
		Timeout  int      `arg:"pos:3"`
		TestKey  string   `arg:"pos:4"`
	}

	c, err := Parse[blpop](args)

	expectNoError(t, err)
	expectEqual(t, "l1", c.Key)
	expectEqualSlice(t, []string{"l2", "l3"}, c.RestKeys)
	expectEqual(t, 120, c.Timeout)
	expectEqual(t, "another", c.TestKey)
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
