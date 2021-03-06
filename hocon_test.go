package hocon

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

//todo test letters in int values
//todo default overrides by value
//todo path for whole containers

func makeTestFile(filemask string) (*os.File, error) {
	return ioutil.TempFile("", filemask)
}

func deleteTestFile(file *os.File) error {
	err := os.Remove(file.Name())
	if err != nil {
		log.Printf("cannot delete temp-file (%s): %s", file.Name(), err.Error())
	}
	return err
}

func assertErrDefaultIsOutOfRange(t *testing.T, err error) {
	assertErrRegex(t, err, "^wrong default value.*value out of range$")
}

func assertErrValueIsOutOfRange(t *testing.T, err error) {
	assertErrRegex(t, err, "^wrong value.*value out of range$")
}

func assertErrValueInvalidSyntax(t *testing.T, err error) {
	assertErrRegex(t, err, "^wrong value.*invalid syntax$")
}

func assertErrDefaultInvalidSyntax(t *testing.T, err error) {
	assertErrRegex(t, err, "^wrong default value.*invalid syntax$")
}

func assertErrRegex(t *testing.T, err error, regex string) {
	if assert.Error(t, err) {
		assert.Regexp(t, regex, err, "got wrong error")
	}
}

func TestCorrectDefaultRanges(t *testing.T) {
	props1 := struct {
		FMin1 int8  `hocon:"default=-128"`
		FMax1 int8  `hocon:"default=127"`
		FMin3 int16 `hocon:"default=-32768"`
		FMax3 int16 `hocon:"default=32767"`
		FMin5 int32 `hocon:"default=-2147483648"`
		FMax5 int32 `hocon:"default=2147483647"`
		FMin7 int64 `hocon:"default=-9223372036854775808"`
		FMax7 int64 `hocon:"default=9223372036854775807"`
	}{}
	err := LoadConfigText("{key1: 1}", &props1)
	assert.Nil(t, err)
	assert.Equal(t, int8(-128), props1.FMin1)
	assert.Equal(t, int8(127), props1.FMax1)
	assert.Equal(t, int16(-32768), props1.FMin3)
	assert.Equal(t, int16(32767), props1.FMax3)
	assert.Equal(t, int32(-2147483648), props1.FMin5)
	assert.Equal(t, int32(2147483647), props1.FMax5)
	assert.Equal(t, int64(-9223372036854775808), props1.FMin7)
	assert.Equal(t, int64(9223372036854775807), props1.FMax7)
}

func TestCorrectRanges(t *testing.T) {
	props1 := struct {
		FMin1 int8
		FMax1 int8
		FMin3 int16
		FMax3 int16
		FMin5 int32
		FMax5 int32
		FMin7 int64
		FMax7 int64
	}{}
	err := LoadConfigText("{FMin1:-128,FMax1:127,FMin2:0,FMax2:255,"+
		"FMin3:-32768,FMax3:32767,FMin4:0,FMax4:65535,"+
		"FMin5:-2147483648,FMax5:2147483647,FMin6:0,FMax6:4294967295,"+
		"FMin7:-9223372036854775808,FMax7:9223372036854775807,FMin8:0,FMax8:18446744073709551615}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, int8(-128), props1.FMin1)
		assert.Equal(t, int8(127), props1.FMax1)
		assert.Equal(t, int16(-32768), props1.FMin3)
		assert.Equal(t, int16(32767), props1.FMax3)
		assert.Equal(t, int32(-2147483648), props1.FMin5)
		assert.Equal(t, int32(2147483647), props1.FMax5)
		assert.Equal(t, int64(-9223372036854775808), props1.FMin7)
		assert.Equal(t, int64(9223372036854775807), props1.FMax7)
	}
}

func TestInt8DefaultRanges(t *testing.T) {
	props1 := struct {
		Field1 int8 `hocon:"default=128"`
	}{}
	err := LoadConfigText("{}", &props1)
	assertErrDefaultIsOutOfRange(t, err)

	props2 := struct {
		Field1 int8 `hocon:"default=-129"`
	}{}
	err2 := LoadConfigText("{key1: 1}", &props2)
	assertErrDefaultIsOutOfRange(t, err2)
}

func TestInt16DefaultRanges(t *testing.T) {
	props1 := struct {
		Field1 int16 `hocon:"default=32768"`
	}{}
	assertErrDefaultIsOutOfRange(t, LoadConfigText("{key1: 1}", &props1))

	props2 := struct {
		Field1 int16 `hocon:"default=-32769"`
	}{}
	assertErrDefaultIsOutOfRange(t, LoadConfigText("{key1: 1}", &props2))
}

func TestInt32DefaultRanges(t *testing.T) {
	props1 := struct {
		Field1 int32 `hocon:"default=2147483648"`
	}{}
	assertErrDefaultIsOutOfRange(t, LoadConfigText("{key1: 1}", &props1))

	props2 := struct {
		Field1 int32 `hocon:"default=-2147483649"`
	}{}
	assertErrDefaultIsOutOfRange(t, LoadConfigText("{key1: 1}", &props2))
}

func TestInt64DefaultRanges(t *testing.T) {
	props1 := struct {
		Field1 int64 `hocon:"default=9223372036854775808"`
	}{}
	assertErrDefaultIsOutOfRange(t, LoadConfigText("{key1: 1}", &props1))

	props2 := struct {
		Field1 int64 `hocon:"default=-9223372036854775809"`
	}{}
	assertErrDefaultIsOutOfRange(t, LoadConfigText("{key1: 1}", &props2))
}

func TestInt8Ranges(t *testing.T) {
	props1 := struct {
		Field1 int8
	}{}
	assertErrValueIsOutOfRange(t, LoadConfigText("{Field1: 128}", &props1))
	assertErrValueIsOutOfRange(t, LoadConfigText("{Field1: -129}", &props1))
}

func TestInt16Ranges(t *testing.T) {
	props1 := struct {
		Field1 int16
	}{}
	assertErrValueIsOutOfRange(t, LoadConfigText("{Field1: 32768}", &props1))
	assertErrValueIsOutOfRange(t, LoadConfigText("{Field1: -32769}", &props1))
}

func TestInt32Ranges(t *testing.T) {
	props1 := struct {
		Field1 int32
	}{}
	assertErrValueIsOutOfRange(t, LoadConfigText("{Field1: 2147483648}", &props1))
	assertErrValueIsOutOfRange(t, LoadConfigText("{Field1: -2147483649}", &props1))
}

func TestInt64Ranges(t *testing.T) {
	props1 := struct {
		Field1 int64
	}{}
	assertErrValueIsOutOfRange(t, LoadConfigText("{Field1: 9223372036854775808}", &props1))
	assertErrValueIsOutOfRange(t, LoadConfigText("{Field1: -9223372036854775809}", &props1))
}

func TestCorrectFloat(t *testing.T) {
	props1 := struct {
		Field1 float32
		Field2 float64
	}{}
	err := LoadConfigText("{Field1:1.2,Field2:1.7,}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, float32(1.2), props1.Field1)
		assert.Equal(t, 1.7, props1.Field2)
	}
}

func TestCorrectFloatDefault(t *testing.T) {
	props1 := struct {
		Field1 float32 `hocon:"default=1e3"`
		Field2 float64 `hocon:"default=1e-3"`
	}{}
	err := LoadConfigText("{}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, float32(1000), props1.Field1)
		assert.Equal(t, 0.001, props1.Field2)
	}
}

func TestCorrectStringDefault(t *testing.T) {
	props1 := struct {
		Field1 string `hocon:"default=1e3"`
		Field2 string `hocon:"default=abc"`
		Field3 string `hocon:"default="`
	}{}
	err := LoadConfigText("{Field2:}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, "1e3", props1.Field1)
		assert.Equal(t, "", props1.Field2)
		assert.Equal(t, "", props1.Field3)
	}
}

func TestIncorrectFloat32(t *testing.T) {
	props1 := struct {
		Field1 float32
	}{}
	err := LoadConfigText("{Field1:1.2.3,}", &props1)
	assert.Error(t, err)
}

func TestIncorrectFloat64(t *testing.T) {
	props1 := struct {
		Field1 float64
	}{}
	err := LoadConfigText("{Field1:7hh7,}", &props1)
	assert.Error(t, err)
}

func TestIncorrectFloat32Default(t *testing.T) {
	props1 := struct {
		Field1 float32 `hocon:"default=1ee3"`
	}{}
	err := LoadConfigText("{}", &props1)
	assert.Error(t, err)
}

func TestIncorrectFloat64Default(t *testing.T) {
	props1 := struct {
		Field1 float64 `hocon:"default=1ee3"`
	}{}
	err := LoadConfigText("{}", &props1)
	assert.Error(t, err)
}

func TestCorrectBool(t *testing.T) {
	props1 := struct {
		Field1 bool
		Field2 bool
		Field3 bool
		Field4 bool
	}{}
	err := LoadConfigText("{Field1:true,Field2:false,Field3:True,Field4:False,}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, true, props1.Field1)
		assert.Equal(t, false, props1.Field2)
		assert.Equal(t, true, props1.Field3)
		assert.Equal(t, false, props1.Field4)
	}
}

func TestCorrectBoolYesNo(t *testing.T) {
	props1 := struct {
		Field1 bool
		Field2 bool
		Field3 bool
		Field4 bool
	}{}
	err := LoadConfigText("{Field1:yes,Field2:no,Field3:Yes,Field4:No,}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, true, props1.Field1)
		assert.Equal(t, false, props1.Field2)
		assert.Equal(t, true, props1.Field3)
		assert.Equal(t, false, props1.Field4)
	}
}

func TestCorrectBoolOnOff(t *testing.T) {
	props1 := struct {
		Field1 bool
		Field2 bool
		Field3 bool
		Field4 bool
	}{}
	err := LoadConfigText("{Field1:on,Field2:off,Field3:On,Field4:Off,}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, true, props1.Field1)
		assert.Equal(t, false, props1.Field2)
		assert.Equal(t, true, props1.Field3)
		assert.Equal(t, false, props1.Field4)
	}
}

func TestK(t *testing.T) {
	props1 := struct {
		Field1 int64
	}{}
	err := LoadConfigText("{Field1:32000,}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, int64(32000), props1.Field1)
	}

	props2 := struct {
		Field1 int64
	}{}
	err2 := LoadConfigText("{Orig:100,Field1:${Orig},}", &props2)
	if assert.Nil(t, err2) {
		assert.Equal(t, int64(100), props2.Field1)
	}
}

func TestCorrectBoolDefault(t *testing.T) {
	props1 := struct {
		Field1 bool `hocon:"default=true"`
		Field2 bool `hocon:"default=false"`
		Field3 bool `hocon:"default=True"`
		Field4 bool `hocon:"default=False"`
	}{}
	err := LoadConfigText("{}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, true, props1.Field1)
		assert.Equal(t, false, props1.Field2)
		assert.Equal(t, true, props1.Field3)
		assert.Equal(t, false, props1.Field4)
	}
}

func TestCorrectBoolYesNoDefault(t *testing.T) {
	props1 := struct {
		Field1 bool `hocon:"default=yes"`
		Field2 bool `hocon:"default=no"`
		Field3 bool `hocon:"default=Yes"`
		Field4 bool `hocon:"default=No"`
	}{}
	err := LoadConfigText("{}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, true, props1.Field1)
		assert.Equal(t, false, props1.Field2)
		assert.Equal(t, true, props1.Field3)
		assert.Equal(t, false, props1.Field4)
	}
}

func TestCorrectBoolOnOffDefault(t *testing.T) {
	props1 := struct {
		Field1 bool `hocon:"default=on"`
		Field2 bool `hocon:"default=off"`
		Field3 bool `hocon:"default=On"`
		Field4 bool `hocon:"default=Off"`
	}{}
	err := LoadConfigText("{}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, true, props1.Field1)
		assert.Equal(t, false, props1.Field2)
		assert.Equal(t, true, props1.Field3)
		assert.Equal(t, false, props1.Field4)
	}
}

func TestIncorrectBool(t *testing.T) {
	props1 := struct {
		Field1 bool
	}{}
	err := LoadConfigText("{Field1:225,}", &props1)
	assert.Error(t, err)
}

func TestIncorrectBoolDefault(t *testing.T) {
	props1 := struct {
		Field1 bool `hocon:"default=4567"`
	}{}
	err := LoadConfigText("{}", &props1)
	assert.Error(t, err)
}

func TestCorrectStringSlice(t *testing.T) {
	props1 := struct {
		Field1 []string
	}{}
	err := LoadConfigText("{Field1:[a,b,c,d]}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, []string{"a", "b", "c", "d"}, props1.Field1)
	}
}

func TestCorrectInt8Slice(t *testing.T) {
	props1 := struct {
		Field1 []int8
	}{}
	err := LoadConfigText("{Field1:[1,2,3,4]}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, []int8{1, 2, 3, 4}, props1.Field1)
	}
}

func TestCorrectInt32Slice(t *testing.T) {
	props1 := struct {
		Field1 []int32
	}{}
	err := LoadConfigText("{Field1:[1,2,3,4]}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, []int32{1, 2, 3, 4}, props1.Field1)
	}
}

func TestIncorrectInt8Slice(t *testing.T) {
	props1 := struct {
		Field1 []int8
	}{}
	err := LoadConfigText("{Field1:[1,2,3,1000]}", &props1)
	assert.Error(t, err)
}

func TestIncorrectSliceDefault(t *testing.T) {
	props1 := struct {
		Field1 []string `hocon:"default=11"`
	}{}
	err := LoadConfigText("{}", &props1)
	assert.Error(t, err)
}

func TestPathByStruct(t *testing.T) {
	props1 := struct {
		Inner1 struct {
			Inner1 struct {
				Field1 string
				Field2 string
			}
			Field1 string
			Field2 string
		}
		Inner2 struct {
			Field1 string
			Field2 string
		}
		Field1 string
		Field2 string
	}{}
	err := LoadConfigText("{Inner1:{Inner1:{Field1:c111,Field2:c112,},Field1:c11,Field2:c12,},"+
		"Inner2:{Field1:c21,Field2:c22,},Field1:c1,Field2:c2,}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, "c111", props1.Inner1.Inner1.Field1)
		assert.Equal(t, "c112", props1.Inner1.Inner1.Field2)
		assert.Equal(t, "c11", props1.Inner1.Field1)
		assert.Equal(t, "c12", props1.Inner1.Field2)
		assert.Equal(t, "c21", props1.Inner2.Field1)
		assert.Equal(t, "c22", props1.Inner2.Field2)
		assert.Equal(t, "c1", props1.Field1)
		assert.Equal(t, "c2", props1.Field2)
	}
}

func TestPathByNode(t *testing.T) {
	props1 := struct {
		Inner1 struct {
			Inner1 struct {
				Field1 string `hocon:"node=f1"`
				Field2 string `hocon:"node=f2"`
			} `hocon:"node=i1"`
			Field1 string `hocon:"node=f1"`
			Field2 string `hocon:"node=f2"`
		} `hocon:"node=i1"`
		Inner2 struct {
			Field1 string `hocon:"node=f1"`
			Field2 string `hocon:"node=f2"`
		} `hocon:"node=i2"`
		Field1 string `hocon:"node=f1"`
		Field2 string `hocon:"node=f2"`
	}{}
	err := LoadConfigText("{i1:{i1:{f1:c111,f2:c112,},f1:c11,f2:c12,},"+
		"i2:{f1:c21,f2:c22,},f1:c1,f2:c2,}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, "c111", props1.Inner1.Inner1.Field1)
		assert.Equal(t, "c112", props1.Inner1.Inner1.Field2)
		assert.Equal(t, "c11", props1.Inner1.Field1)
		assert.Equal(t, "c12", props1.Inner1.Field2)
		assert.Equal(t, "c21", props1.Inner2.Field1)
		assert.Equal(t, "c22", props1.Inner2.Field2)
		assert.Equal(t, "c1", props1.Field1)
		assert.Equal(t, "c2", props1.Field2)
	}
}

func TestPathByPath(t *testing.T) {
	props1 := struct {
		Inner1 struct {
			Inner1 struct {
				Field1 string `hocon:"path=i1.i1.f1"`
				Field2 string `hocon:"path=i1.i1.f2"`
			}
			Field1 string `hocon:"path=i1.f1"`
			Field2 string `hocon:"path=i1.f2"`
		}
		Inner2 struct {
			Field1 string `hocon:"path=i2.f1"`
			Field2 string `hocon:"path=i2.f2"`
		}
		Field1 string `hocon:"path=f1"`
		Field2 string `hocon:"path=f2"`
	}{}
	err := LoadConfigText("{i1:{i1:{f1:c111,f2:c112,},f1:c11,f2:c12,},"+
		"i2:{f1:c21,f2:c22,},f1:c1,f2:c2,}", &props1)
	if assert.Nil(t, err) {
		assert.Equal(t, "c111", props1.Inner1.Inner1.Field1)
		assert.Equal(t, "c112", props1.Inner1.Inner1.Field2)
		assert.Equal(t, "c11", props1.Inner1.Field1)
		assert.Equal(t, "c12", props1.Inner1.Field2)
		assert.Equal(t, "c21", props1.Inner2.Field1)
		assert.Equal(t, "c22", props1.Inner2.Field2)
		assert.Equal(t, "c1", props1.Field1)
		assert.Equal(t, "c2", props1.Field2)
	}
}

func TestNodeAndPath(t *testing.T) {
	props1 := struct {
		Inner1 struct {
			Field1 string `hocon:"node=n1,path=i1.f1"`
			Field2 string `hocon:"path=i1.f2"`
		} `hocon:"node=i1"`
		Field1 string `hocon:"node=n1,path=f1"`
		Field2 string `hocon:"path=f2"`
	}{}
	err1 := LoadConfigText("{i1:{f1:c11,f2:c12,},f1:c1,f2:c2,}", &props1)
	if assert.Nil(t, err1) {
		assert.Equal(t, "c11", props1.Inner1.Field1)
		assert.Equal(t, "c12", props1.Inner1.Field2)
		assert.Equal(t, "c1", props1.Field1)
		assert.Equal(t, "c2", props1.Field2)
	}
}

func TestIncorrectTag(t *testing.T) {
	props1 := struct {
		Field1 string `hocon:"node=n1 path=a.b.c"`
	}{}
	err1 := LoadConfigText("{}", &props1)
	assert.Error(t, err1)

	props2 := struct {
		Inner struct{} `hocon:"node=n1 path=a.b.c"`
	}{}
	err2 := LoadConfigText("{}", &props2)
	assert.Error(t, err2)
}

func TestAbsentFile(t *testing.T) {
	file, err1 := makeTestFile("denied.conf")
	if assert.Nil(t, err1, "cannot create temp file") {
		filename := file.Name()
		err2 := deleteTestFile(file)
		if assert.Nil(t, err2, "cannot remove temp file") {
			props1 := struct{}{}

			err3 := LoadConfigFile(filename, &props1)
			if assert.Error(t, err3) {
				assert.Regexp(t, "no such file or directory$", err3)
			}
		}
	}
}

func TestDeniedFile(t *testing.T) {
	file, err1 := makeTestFile("denied.conf")
	if assert.Nil(t, err1, "cannot create temp file") {
		defer func() {
			_ = deleteTestFile(file)
		}()
		err2 := file.Chmod(os.FileMode(0))
		if assert.Nil(t, err2, "cannot set file permissions") {
			props1 := struct{}{}
			err1 := LoadConfigFile(file.Name(), &props1)
			if assert.Error(t, err1) {
				assert.Regexp(t, "permission denied$", err1)
			}
		}
	}
}

func TestDirectoryAsFile(t *testing.T) {
	props1 := struct{}{}
	err1 := LoadConfigFile("tests", &props1)
	if assert.Error(t, err1) {
		assert.Regexp(t, "is a directory$", err1)
	}
}

func TestLoadConfigFile(t *testing.T) {
	props1 := struct {
		Inner struct {
			Key9 int32
		} `hocon:"node=container1"`
	}{}
	err1 := LoadConfigFile("tests/conf1.conf", &props1)
	if assert.Empty(t, err1) {
		assert.Equal(t, int32(-999), props1.Inner.Key9)
	}
}

func TestNoValue(t *testing.T) {
	props1 := struct{ Key int32 }{}
	err1 := LoadConfigText("{}", &props1)
	if assert.Error(t, err1) {
		assert.Regexp(t, "^no value either default value provided", err1)
	}
}

func TestUnsupported(t *testing.T) {
	props1 := struct {
		Key int `hocon:"default=0"`
	}{}
	err1 := LoadConfigText("{}", &props1)
	assert.Error(t, err1)

	props2 := struct {
		Key uint `hocon:"default=0"`
	}{}
	err2 := LoadConfigText("{}", &props2)
	assert.Error(t, err2)

	props3 := struct {
		Key uint8 `hocon:"default=0"`
	}{}
	err3 := LoadConfigText("{}", &props3)
	assert.Error(t, err3)

	props4 := struct {
		Key uint16 `hocon:"default=0"`
	}{}
	err4 := LoadConfigText("{}", &props4)
	assert.Error(t, err4)

	props5 := struct {
		Key uint32 `hocon:"default=0"`
	}{}
	err5 := LoadConfigText("{}", &props5)
	assert.Error(t, err5)

	props6 := struct {
		Key uint64 `hocon:"default=0"`
	}{}
	err6 := LoadConfigText("{}", &props6)
	assert.Error(t, err6)
}

func TestUnimplemented(t *testing.T) {
	props1 := struct {
		Key func() `hocon:"default=0"`
	}{}
	err1 := LoadConfigText("{}", &props1)
	assert.Error(t, err1)
}

func TestPanicTextConfig(t *testing.T) {
	assert.Panics(t, func() {
		props1 := struct{}{}
		err1 := LoadConfigText("{F:1+1}", &props1)
		assert.Error(t, err1)
	})
}

func TestPanicFileConfig(t *testing.T) {
	file, err1 := makeTestFile("panic.conf")
	if assert.Nil(t, err1, "cannot create temp file") {
		defer func() {
			_ = deleteTestFile(file)
		}()
		_, err2 := file.WriteString("{F:1+1}")
		if assert.Nil(t, err2) {
			assert.Panics(t, func() {
				props1 := struct{}{}
				err1 := LoadConfigFile(file.Name(), &props1)
				assert.Error(t, err1)
			})
		}
	}
}
