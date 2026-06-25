package nbt

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

const (
	TagEnd       byte = 0
	TagByte      byte = 1
	TagShort     byte = 2
	TagInt       byte = 3
	TagLong      byte = 4
	TagFloat     byte = 5
	TagDouble    byte = 6
	TagByteArray byte = 7
	TagString    byte = 8
	TagList      byte = 9
	TagCompound  byte = 10
	TagIntArray  byte = 11
)

func Marshal(w io.Writer, rootName string, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	tagType, err := getNBTType(val)
	if err != nil {
		return err
	}

	if _, err := w.Write([]byte{tagType}); err != nil {
		return err
	}

	if err := writeString(w, rootName); err != nil {
		return err
	}

	return writePayload(w, val)
}

func getNBTType(val reflect.Value) (byte, error) {
	switch val.Kind() {
	case reflect.Int8:
		return TagByte, nil
	case reflect.Int16:
		return TagShort, nil
	case reflect.Int32:
		return TagInt, nil
	case reflect.Int64:
		return TagLong, nil
	case reflect.Float32:
		return TagFloat, nil
	case reflect.Float64:
		return TagDouble, nil
	case reflect.String:
		return TagString, nil
	case reflect.Struct:
		return TagCompound, nil
	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.Uint8 || val.Type().Elem().Kind() == reflect.Int8 {
			return TagByteArray, nil
		}
		return TagList, nil
	default:
		return 0, fmt.Errorf("tipe data %s tidak didukung", val)
	}
}

func writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.BigEndian, uint16(len(s))); err != nil {
		return err
	}
	_, err := w.Write([]byte(s))
	return err
}

func writePayload(w io.Writer, val reflect.Value) error {
	switch val.Kind() {
	case reflect.Int8:
		return binary.Write(w, binary.BigEndian, int8(val.Int()))
	case reflect.Int16:
		return binary.Write(w, binary.BigEndian, int16(val.Int()))

	case reflect.Int32:
		return binary.Write(w, binary.BigEndian, int32(val.Int()))

	case reflect.Int64:
		return binary.Write(w, binary.BigEndian, val.Int())

	case reflect.Float32:
		return binary.Write(w, binary.BigEndian, float32(val.Float()))

	case reflect.Float64:
		return binary.Write(w, binary.BigEndian, val.Float())

	case reflect.String:
		return writeString(w, val.String())
	case reflect.Struct:
		typ := val.Type()
		for i := 0; i < val.NumField(); i++ {
			fieldVal := val.Field(i)
			fieldType := typ.Field(i)

			nbtTag := fieldType.Tag.Get("nbt")
			if nbtTag == "" {
				nbtTag = fieldType.Name
			}

			fType, err := getNBTType(fieldVal)
			if err != nil {
				return err
			}

			w.Write([]byte{fType})
			writeString(w, nbtTag)
			if err := writePayload(w, fieldVal); err != nil {
				return err
			}
		}
		_, err := w.Write([]byte{TagEnd})
		return err
	case reflect.Slice:
		if val.Type().Elem().Kind() == reflect.Uint8 || val.Type().Elem().Kind() == reflect.Int8 {
			length := int32(val.Len())
			if err := binary.Write(w, binary.BigEndian, length); err != nil {
				return err
			}
			_, err := w.Write(val.Bytes())
			return err
		}

		length := int32(val.Len())
		var subType byte = TagEnd
		if length > 0 {
			st, err := getNBTType(val.Index(0))
			if err != nil {
				return err
			}
			subType = st
		}
		if _, err := w.Write([]byte{subType}); err != nil {
			return err
		}

		if err := binary.Write(w, binary.BigEndian, length); err != nil {
			return err
		}

		for i := 0; i < val.Len(); i++ {
			if err := writePayload(w, val.Index(i)); err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}

func Unmarshal(r io.Reader, v interface{}) error {
	val := reflect.ValueOf(v)

	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("unmarshal membutuhkan pointer struct")
	}

	var rootType byte
	if err := binary.Read(r, binary.BigEndian, &rootType); err != nil {
		return err
	}

	if _, err := readString(r); err != nil {
		return err
	}

	return readPayload(r, rootType, val.Elem())
}

func readString(r io.Reader) (string, error) {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func readPayload(r io.Reader, tagType byte, target reflect.Value) error {
	switch tagType {
	case TagByte:
		var v int8
		if err := binary.Read(r, binary.BigEndian, &v); err != nil {
			return err
		}
		if target.CanSet() {
			target.SetInt(int64(v))
		}
	case TagShort:
		var v int16
		if err := binary.Read(r, binary.BigEndian, &v); err != nil {
			return err
		}
		if target.CanSet() {
			target.SetInt(int64(v))
		}
	case TagInt:
		var v int32
		if err := binary.Read(r, binary.BigEndian, &v); err != nil {
			return err
		}
		if target.CanSet() {
			target.SetInt(int64(v))
		}
	case TagLong:
		var v int64
		if err := binary.Read(r, binary.BigEndian, &v); err != nil {
			return err
		}
		if target.CanSet() {
			target.SetInt(v)
		}
	case TagFloat:
		var v float32
		if err := binary.Read(r, binary.BigEndian, &v); err != nil {
			return err
		}
		if target.CanSet() {
			target.SetFloat(float64(v))
		}
	case TagDouble:
		var v float64
		if err := binary.Read(r, binary.BigEndian, &v); err != nil {
			return err
		}
		if target.CanSet() {
			target.SetFloat(v)
		}
	case TagString:
		s, err := readString(r)
		if err != nil {
			return err
		}
		if target.CanSet() {
			target.SetString(s)
		}
	case TagByteArray:
		var length int32
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return err
		}
		buf := make([]byte, length)
		if _, err := io.ReadFull(r, buf); err != nil {
			return err
		}
		if target.CanSet() {
			target.SetBytes(buf)
		}
	case TagList:
		var subType byte
		var length int32
		if err := binary.Read(r, binary.BigEndian, &subType); err != nil {
			return err
		}
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return err
		}

		if target.CanSet() {
			sliceType := target.Type()
			newSlice := reflect.MakeSlice(sliceType, int(length), int(length))
			for i := range int(length) {
				if err := readPayload(r, subType, newSlice.Index(i)); err != nil {
					return err
				}
			}

			target.Set(newSlice)
		}
	case TagCompound:
		for {
			var fType byte
			if err := binary.Read(r, binary.BigEndian, &fType); err != nil {
				return err
			}
			if fType == TagEnd {
				break
			}

			fName, err := readString(r)
			if err != nil {
				return err
			}
			foundField := findFieldByTag(target, fName)
			if foundField.IsValid() {
				if err := readPayload(r, fType, foundField); err != nil {
					return err
				}
			} else {
				if err := SkipPayload(r, fType); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func findFieldByTag(strct reflect.Value, tagName string) reflect.Value {
	if strct.Kind() != reflect.Struct {
		return reflect.Value{}
	}

	typ := strct.Type()
	for i := range strct.NumField() {
		tag := typ.Field(i).Tag.Get("nbt")
		if tag == tagName || (tag == "" && typ.Field(i).Name == tagName) {
			return strct.Field(i)
		}
	}

	return reflect.Value{}
}

func SkipPayload(r io.Reader, tagType byte) error {
	switch tagType {
	case TagByte:
		_, err := io.CopyN(io.Discard, r, 1)
		return err
	case TagShort:
		_, err := io.CopyN(io.Discard, r, 2)
		return err
	case TagInt, TagFloat:
		_, err := io.CopyN(io.Discard, r, 4)
		return err
	case TagLong, TagDouble:
		_, err := io.CopyN(io.Discard, r, 8)
		return err
	case TagString:
		var length uint16
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return err
		}
		_, err := io.CopyN(io.Discard, r, int64(length))
		return err
	case TagByteArray:
		var length int32
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return err
		}
		_, err := io.CopyN(io.Discard, r, int64(length))
		return err
	case TagList:
		var subType byte
		var length int32
		if err := binary.Read(r, binary.BigEndian, &subType); err != nil {
			return err
		}
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			return err
		}
		for i := 0; i < int(length); i++ {
			if err := SkipPayload(r, subType); err != nil {
				return err
			}
		}
	case TagCompound:
		for {
			var t byte
			if err := binary.Read(r, binary.BigEndian, &t); err != nil {
				return err
			}
			if t == TagEnd {
				break
			}
			var length uint16
			if err := binary.Read(r, binary.BigEndian, &length); err != nil {
				return err
			}
			if _, err := io.CopyN(io.Discard, r, int64(length)); err != nil {
				return err
			}
			if err := SkipPayload(r, t); err != nil {
				return err
			}
		}
	}
	return nil
}

func UnmarshalPayload(r io.Reader, tagType byte, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("unmarshalPayload membutuhkan pointer struct")
	}
	return readPayload(r, tagType, val.Elem())
}
