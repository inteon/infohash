package infohash

import (
	"fmt"
	"hash/fnv"
	"io"
	"reflect"
	"sort"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

const tagName = "infohash"

type FieldChangedError struct {
	Field string
}

func (e FieldChangedError) Error() string {
	if e.Field == "" {
		return "a field value changed"
	}

	return fmt.Sprintf("the field %q's value changed", e.Field)
}

type fieldInfo struct {
	name       string
	fieldValue interface{}
}

func getFieldInfos(obj interface{}) ([]fieldInfo, error) {
	vObj := reflect.ValueOf(obj)
	if vObj.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("the object must be a pointer")
	}

	v := vObj.Elem()
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("the object must be a pointer to a struct")
	}

	t := reflect.TypeOf(obj).Elem()

	fieldInfos := make([]fieldInfo, 0, t.NumField())
	tags := make(map[string]struct{})
	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)
		fieldValue := v.Field(i)

		tag := structField.Tag.Get(tagName)

		if tag == "" {
			return nil, fmt.Errorf("the field %s has no tag %s", structField.Name, tagName)
		}

		if _, ok := tags[tag]; ok {
			return nil, fmt.Errorf("the tag %q is used more than once", tag)
		}
		tags[tag] = struct{}{}

		fieldInfos = append(fieldInfos, fieldInfo{
			name:       tag,
			fieldValue: fieldValue.Addr().Interface(),
		})
	}

	sort.Slice(fieldInfos, func(a, b int) bool {
		return fieldInfos[a].name < fieldInfos[b].name
	})

	return fieldInfos, nil
}

func getHashes(fieldInfos []fieldInfo) (uint64, []uint32, error) {
	fieldHashes := make([]uint32, 0, len(fieldInfos))

	fullHash := fnv.New64a()
	fieldHash := fnv.New32a()
	multiWriter := io.MultiWriter(fullHash, fieldHash)

	for _, info := range fieldInfos {
		fieldHash.Reset()

		if _, err := multiWriter.Write([]byte(info.name)); err != nil {
			return 0, nil, err
		}

		if _, err := prettyPrintConfigForHash.Fprintf(multiWriter, "%#v", info.fieldValue); err != nil {
			return 0, nil, err
		}

		fieldHashes = append(fieldHashes, fieldHash.Sum32())
	}

	return fullHash.Sum64(), fieldHashes, nil
}

// This test function must be added to the unit tests in your project.
// It will make sure that the defined fields of the struct are not
// changed, which would yield all calculated hashes invalid.
func TestStructDefinition(t *testing.T, obj interface{}, expectedHash []byte) {
	fieldInfos, err := getFieldInfos(obj)
	if err != nil {
		t.Error(err)
	}

	structualHash := fnv.New64a()
	for _, info := range fieldInfos {
		if _, err := structualHash.Write([]byte(info.name)); err != nil {
			t.Error(err)
		}
	}

	if structualHash.Sum64() != uint64FromSlice(expectedHash[:8]) {
		t.Errorf("the struct definition has changed, the hash is invalid")
	}
}

// HashStruct returns an infohash of the given struct.
// The passed object must be a pointer to a struct.
// The struct must have a non-empty tag "infohash" on each of its fields,
// the tag value must be unique for each field. The tag value is used as
// the name of the field in the returned error in case of a mismatch.
//
// The CompareHashStruct function can be used to compare the hash of a struct
// with a previously calculated hash and return an error if the struct has changed.
// The error contains the name of the field that has changed.
func HashStruct(obj interface{}) ([]byte, error) {
	fieldInfos, err := getFieldInfos(obj)
	if err != nil {
		return nil, err
	}
	return hashInfo(fieldInfos)
}

func hashInfo(fieldInfos []fieldInfo) ([]byte, error) {
	fullHash, fieldHashes, err := getHashes(fieldInfos)
	if err != nil {
		return nil, err
	}
	hammingCode := calculateHammingCode(fieldHashes)

	combinedHash := make([]byte, 0, (64+32*len(hammingCode))/8)
	combinedHash = append(combinedHash, uint64ToSlice(fullHash)...)
	combinedHash = append(combinedHash, uint32SliceToByteSlice(hammingCode)...)

	return combinedHash, nil
}

// The CompareHashStruct function compares the hash of the given struct with the given hash.
// The passed object must be a pointer to a struct.
// The struct must have a non-empty tag "infohash" on each of its fields,
// the tag value must be unique for each field. The tag value is used as
// the name of the field in the returned error in case of a mismatch.
//
// If the hash matches, the function returns nil.
// If the hash does not match, the function returns a FieldChangedError.
// If there is only one field that has changed, the error contains the name of the field.
// If there are multiple fields that have changed, the error contains an empty string.
func CompareHashStruct(obj interface{}, existingHash []byte) error {
	fieldInfos, err := getFieldInfos(obj)
	if err != nil {
		return err
	}

	fullHash, fieldHashes, err := getHashes(fieldInfos)
	if err != nil {
		return err
	}

	existingFullHash := uint64FromSlice(existingHash[:8])
	if existingFullHash == fullHash {
		return nil
	}

	existingHammingCode := uint32SliceFromByteSlice(existingHash[8:])
	foundLocation, location := findErrorLocation(fieldHashes, existingHammingCode)

	if !foundLocation {
		return FieldChangedError{}
	}

	mismatchingField := fieldInfos[location]

	return FieldChangedError{
		Field: mismatchingField.name,
	}
}

func uint64FromSlice(s []byte) uint64 {
	return uint64(0) |
		uint64(s[0])<<0 |
		uint64(s[1])<<8 |
		uint64(s[2])<<16 |
		uint64(s[3])<<24 |
		uint64(s[4])<<32 |
		uint64(s[5])<<40 |
		uint64(s[6])<<48 |
		uint64(s[7])<<56
}

func uint64ToSlice(i uint64) []byte {
	return []byte{
		byte(0xff & (i >> 0)),
		byte(0xff & (i >> 8)),
		byte(0xff & (i >> 16)),
		byte(0xff & (i >> 24)),
		byte(0xff & (i >> 32)),
		byte(0xff & (i >> 40)),
		byte(0xff & (i >> 48)),
		byte(0xff & (i >> 56)),
	}
}

func uint32SliceFromByteSlice(s []byte) []uint32 {
	out := make([]uint32, len(s)/4)

	for i := range out {
		out[i] = uint32(s[i*4+0])<<0 |
			uint32(s[i*4+1])<<8 |
			uint32(s[i*4+2])<<16 |
			uint32(s[i*4+3])<<24
	}

	return out
}

func uint32SliceToByteSlice(i []uint32) []byte {
	out := make([]byte, len(i)*4)

	for j := range i {
		out[j*4+0] = byte(0xff & (i[j] >> 0))
		out[j*4+1] = byte(0xff & (i[j] >> 8))
		out[j*4+2] = byte(0xff & (i[j] >> 16))
		out[j*4+3] = byte(0xff & (i[j] >> 24))
	}

	return out
}

// The config MUST NOT be changed because that could change the result of a hash operation
// Based on: https://github.com/kubernetes/kubernetes/blob/f7cb8a5e8a860df643e143cde54d34080372b771/staging/src/k8s.io/apimachinery/pkg/util/dump/dump.go#L47
var prettyPrintConfigForHash = &spew.ConfigState{
	Indent:                  " ",
	SortKeys:                true,
	DisableMethods:          true,
	SpewKeys:                true,
	DisablePointerAddresses: true,
	DisableCapacities:       true,
}
