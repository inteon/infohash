package infohash

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func checkResult(t *testing.T, err error, obj1, obj2 interface{}) {
	t.Helper()

	if reflect.DeepEqual(obj1, obj2) {
		if err != nil {
			t.Fatal(err)
		}

		return
	}

	if err == nil {
		t.Error("the objects are not equal, but the error is nil")
		return
	}

	fcErr, ok := err.(FieldChangedError)
	if !ok {
		t.Error("the error is not FieldChangedError")
	}

	if fcErr.Field == "" {
		t.Errorf("the field name is empty, obj1: %v, obj2: %v", obj1, obj2)
	}

	value1 := reflect.ValueOf(obj1).FieldByName(fcErr.Field).Interface()
	value2 := reflect.ValueOf(obj2).FieldByName(fcErr.Field).Interface()

	if reflect.DeepEqual(value1, value2) {
		t.Errorf("the field %q's value is equal, but the error is not nil", fcErr.Field)
	}
}

func TestHashStruct(t *testing.T) {
	type testStruct struct {
		Field1 string `infohash:"Field1"`
		Field2 string `infohash:"Field2"`
		Field3 string `infohash:"Field3"`
		Field4 string `infohash:"Field4"`
		Field5 string `infohash:"Field5"`
	}

	test1 := testStruct{
		Field1: "test1",
		Field2: "test2",
		Field3: "test3",
		Field4: "test4",
		Field5: "test5",
	}

	hash, err := HashStruct(&test1)
	if err != nil {
		t.Fatal(err)
	}

	type testCase func(obj *testStruct)

	testCases := []testCase{
		func(obj *testStruct) {},
		func(obj *testStruct) {
			obj.Field1 = "Field1 changed"
		},
		func(obj *testStruct) {
			obj.Field5 = "Field5 changed"
		},
		func(obj *testStruct) {
			obj.Field4 = "Field4 changed"
		},
	}

	for _, tc := range testCases {
		test2 := test1
		tc(&test2)

		checkResult(t, CompareHashStruct(&test2, hash), test1, test2)
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestHashStructRandom(t *testing.T) {
	type testStruct struct {
		Field1  string   `infohash:"Field1"`
		Field2  []string `infohash:"Field2"`
		Field3  int      `infohash:"Field3"`
		Field4  string   `infohash:"Field4"`
		Field5  float32  `infohash:"Field5"`
		Field6  float32  `infohash:"Field6"`
		Field7  float32  `infohash:"Field7"`
		Field8  float32  `infohash:"Field8"`
		Field9  float32  `infohash:"Field9"`
		Field10 float32  `infohash:"Field10"`
		Field11 float32  `infohash:"Field11"`
		Field12 float32  `infohash:"Field12"`
	}

	test1 := testStruct{
		Field1: "test1",
		Field2: []string{"test2", "test3"},
		Field3: 123,
		Field4: "test4",
		Field5: 123.456,
	}

	hash, err := HashStruct(&test1)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		test2 := test1
		test2.Field4 = randStringRunes(10)

		checkResult(t, CompareHashStruct(&test2, hash), test1, test2)
	}
}

func TestHashStructRandomMultipleFields(t *testing.T) {
	type testStruct struct {
		Field1 string   `infohash:"Field1"`
		Field2 []string `infohash:"Field2"`
		Field3 int      `infohash:"Field3"`
		Field4 string   `infohash:"Field4"`
		Field5 float32  `infohash:"Field5"`
	}

	test1 := testStruct{
		Field1: "test1",
		Field2: []string{"test2", "test3"},
		Field3: 123,
		Field4: "test4",
		Field5: 123.456,
	}

	hash, err := HashStruct(&test1)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		test2 := test1
		test2.Field4 = randStringRunes(10)
		test2.Field5 = rand.Float32()
		test2.Field3 = rand.Int()

		err := CompareHashStruct(&test2, hash)
		if err == nil {
			t.Fatal("the error is nil, but it should not be")
		}

		terr, ok := err.(FieldChangedError)
		if !ok {
			t.Fatal("the error is not FieldChangedError")
		}

		if terr.Field != "" {
			t.Fatal("the field name is not empty")
		}
	}
}

func TestHashStatic(t *testing.T) {
	type testStruct struct {
		Field1  string   `infohash:"Field1"`
		Field2  []string `infohash:"Field2"`
		Field3  int      `infohash:"Field3"`
		Field4  string   `infohash:"Field4"`
		Field5  float32  `infohash:"Field5"`
		Field6  float32  `infohash:"Field6"`
		Field7  float32  `infohash:"Field7"`
		Field8  float32  `infohash:"Field8"`
		Field9  float32  `infohash:"Field9"`
		Field10 float32  `infohash:"Field10"`
		Field11 float32  `infohash:"Field11"`
		Field12 float32  `infohash:"Field12"`
	}

	test1 := testStruct{
		Field1: "test1",
		Field2: []string{"test2", "test3"},
		Field3: 123,
		Field4: "test4",
		Field5: 123.456,
	}

	hash, err := HashStruct(&test1)
	if err != nil {
		t.Fatal(err)
	}

	if hex.EncodeToString(hash) != "5d69bd36c3021e652f75a61a8bfcacedd54dff048b9eb95c" {
		t.Fatal("hash is wrong")
	}
}

func TestHashLength(t *testing.T) {
	makeInfos := func(length int) []fieldInfo {
		infos := make([]fieldInfo, length)
		for i := range infos {
			infos[i] = fieldInfo{
				name:       fmt.Sprintf("Field%d", i),
				fieldValue: nil,
			}
		}
		return infos
	}

	for i := 0; i < 128; i++ {
		hash, err := hashInfo(makeInfos(i))
		if err != nil {
			t.Fatal(err)
		}

		log2NumberOfFieldsPlusOne := int(math.Ceil(math.Log2(float64(i + 1))))

		if len(hash) != (64+log2NumberOfFieldsPlusOne*32)/8 {
			t.Fatalf("the hash length is wrong: %d != %d", len(hash), (64+log2NumberOfFieldsPlusOne*32)/8)
		}

		// hash length in case we would store a hash for each field:
		// 64 + 32 * (i - 1)
		hashForEachField := (64 + 32*math.Max(0, float64(i-1))) / 8

		t.Logf("number of fields: %d, infohash length: %d bytes, per-field hash length: %f bytes, relative change: %f %%", i, len(hash), hashForEachField, 100*(float64(len(hash))-hashForEachField)/hashForEachField)
	}
}

func TestGetFieldInfos(t *testing.T) {
	type testCase struct {
		gen func() (interface{}, []fieldInfo)
		err string
	}

	newFieldInfo := func(name string, fieldValue interface{}) fieldInfo {
		return fieldInfo{
			name:       name,
			fieldValue: fieldValue,
		}
	}

	testCases := []testCase{
		{
			gen: func() (interface{}, []fieldInfo) {
				obj := &struct {
					Field1 string `infohash:"Field1"`
					Field2 string `infohash:"Field2"`
					Field3 string `infohash:"Field3"`
				}{}
				return obj, []fieldInfo{
					newFieldInfo("Field1", &obj.Field1),
					newFieldInfo("Field2", &obj.Field2),
					newFieldInfo("Field3", &obj.Field3),
				}
			},
		},
		{
			gen: func() (interface{}, []fieldInfo) {
				obj := &struct {
					Field1 map[int]*[]string `infohash:"Field1"`
					Field3 [][]string        `infohash:"Field3"`
					Field2 ***string         `infohash:"Field2"`
				}{}
				return obj, []fieldInfo{
					newFieldInfo("Field1", &obj.Field1),
					newFieldInfo("Field2", &obj.Field2),
					newFieldInfo("Field3", &obj.Field3),
				}
			},
		},
		{
			gen: func() (interface{}, []fieldInfo) {
				obj := struct {
					Field2 string `infohash:"Field2"`
				}{}
				return obj, []fieldInfo{}
			},
			err: "the object must be a pointer",
		},
		{
			gen: func() (interface{}, []fieldInfo) {
				obj := &struct {
					Field2 string `infohash:"Field2"`
				}{}
				return &obj, []fieldInfo{}
			},
			err: "the object must be a pointer to a struct",
		},
		{
			gen: func() (interface{}, []fieldInfo) {
				obj := &struct {
					Field1 map[int]*[]string
					Field2 ***string `infohash:"Field2"`
				}{}
				return obj, []fieldInfo{}
			},
			err: "the field Field1 has no tag infohash",
		},
		{
			gen: func() (interface{}, []fieldInfo) {
				obj := &struct {
					Field1 map[int]*[]string `infohash:"Field2"`
					Field2 ***string         `infohash:"Field2"`
				}{}
				return obj, []fieldInfo{}
			},
			err: "the tag \"Field2\" is used more than once",
		},
		{
			gen: func() (interface{}, []fieldInfo) {
				obj := &struct {
					Field1 map[int]*[]string `infohash:""`
					Field2 ***string         `infohash:"Field2"`
				}{}
				return obj, []fieldInfo{}
			},
			err: "the field Field1 has no tag infohash",
		},
	}

	for _, tc := range testCases {
		obj, targetInfos := tc.gen()
		infos, err := getFieldInfos(obj)
		if tc.err != "" && err == nil {
			t.Error("the error is nil, but it should not be")
		} else if tc.err == "" && err != nil {
			t.Errorf("the error is not nil, but it should be: %v", err)
		} else if tc.err != "" && err != nil && tc.err != err.Error() {
			t.Errorf("the error is wrong: %q != %q", tc.err, err.Error())
		}

		if err != nil {
			continue
		}

		for i := range infos {
			if infos[i].name != targetInfos[i].name {
				t.Errorf("the field name is wrong: %q != %q", infos[i].name, targetInfos[i].name)
			}

			if infos[i].fieldValue != targetInfos[i].fieldValue {
				t.Errorf("the field value is wrong: %v != %v", infos[i].fieldValue, targetInfos[i].fieldValue)
			}
		}
	}
}
