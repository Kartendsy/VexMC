package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"vexmc/nbt"
)

type Position struct {
	X int32
	Y int16
	Z int32
}

type Slot struct {
	ItemID    int16
	ItemCount byte
	Damage    int16
	NBT       any
}

func MarshalPacket(v any) ([]byte, error) {
	buf := new(bytes.Buffer)
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	typ := val.Type()
	for i := range val.NumField() {
		fieldVal := val.Field(i)
		tag := typ.Field(i).Tag.Get("packet")

		switch tag {
		case "varint":
			if err := WriteVarInt(buf, int32(fieldVal.Int())); err != nil {
				return nil, err
			}
		case "string":
			str := fieldVal.String()
			if err := WriteVarInt(buf, int32(len(str))); err != nil {
				return nil, err
			}
			buf.WriteString(str)
		case "byte":
			if err := binary.Write(buf, binary.BigEndian, uint8(fieldVal.Uint())); err != nil {
				return nil, err
			}
		case "boolean":
			var b byte = 0
			if fieldVal.Bool() {
				b = 1
			}
			buf.WriteByte(b)
		case "short":
			if err := binary.Write(buf, binary.BigEndian, int16(fieldVal.Int())); err != nil {
				return nil, err
			}
		case "ushort":
			if err := binary.Write(buf, binary.BigEndian, uint16(fieldVal.Uint())); err != nil {
				return nil, err
			}
		case "int":
			if err := binary.Write(buf, binary.BigEndian, int32(fieldVal.Int())); err != nil {
				return nil, err
			}
		case "long":
			if err := binary.Write(buf, binary.BigEndian, int64(fieldVal.Int())); err != nil {
				return nil, err
			}
		case "float":
			if err := binary.Write(buf, binary.BigEndian, float32(fieldVal.Float())); err != nil {
				return nil, err
			}
		case "double":
			if err := binary.Write(buf, binary.BigEndian, float64(fieldVal.Float())); err != nil {
				return nil, err
			}
		case "position":
			pos, ok := fieldVal.Interface().(Position)
			if !ok {
				if fieldVal.Kind() == reflect.Ptr && !fieldVal.IsNil() {
					pos = fieldVal.Elem().Interface().(Position)
				} else {
					return nil, fmt.Errorf("field dengan tag position harus struct")
				}
			}

			var packed int64 = (int64(pos.X&0x3FFFFFF) << 38) | (int64(pos.Z&0x3FFFFFF) << 12) | (int64(pos.Y & 0xFFF))
			if err := binary.Write(buf, binary.BigEndian, packed); err != nil {
				return nil, err
			}
		case "slot":
			slot, ok := fieldVal.Interface().(Slot)
			if !ok {
				if fieldVal.Kind() == reflect.Ptr && !fieldVal.IsNil() {
					slot = fieldVal.Elem().Interface().(Slot)
				} else {
					return nil, fmt.Errorf("field dengan tag slot harus bertipe Slot")
				}
			}

			if err := binary.Write(buf, binary.BigEndian, slot); err != nil {
				return nil, err
			}

			if slot.ItemID != -1 {
				if err := binary.Write(buf, binary.BigEndian, uint8(slot.ItemCount)); err != nil {
					return nil, err
				}

				if err := binary.Write(buf, binary.BigEndian, slot.Damage); err != nil {
					return nil, err
				}

				if slot.NBT != nil {
					if err := nbt.Marshal(buf, "", slot.NBT); err != nil {
						return nil, fmt.Errorf("gagal marshal NBT slot 1.8: %v", err)
					}
				} else {
					buf.WriteByte(0)
				}
			}

		case "bytearray":
			bytesData := fieldVal.Bytes()
			if err := WriteVarInt(buf, int32(len(bytesData))); err != nil {
				return nil, err
			}
			buf.Write(bytesData)
		default:
			return nil, fmt.Errorf("tag %s tidak didukung", tag)
		}
	}

	return buf.Bytes(), nil
}

func UnmarshalPacket(r io.Reader, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("unmarshal butuh pointer struct")
	}

	val = val.Elem()
	typ := val.Type()

	for i := range val.NumField() {
		fieldVal := val.Field(i)
		tag := typ.Field(i).Tag.Get("packet")

		switch tag {
		case "varint":
			vInt, err := ReadVarInt(r)
			if err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetInt(int64(vInt))
			}
		case "string":
			length, err := ReadVarInt(r)
			if err != nil {
				return err
			}
			buf := make([]byte, length)
			if _, err := io.ReadFull(r, buf); err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetString(string(buf))
			}

		case "byte":
			var vByte uint8
			if err := binary.Read(r, binary.BigEndian, &vByte); err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetUint(uint64(vByte))
			}
		case "boolean":
			var vByte uint8
			if err := binary.Read(r, binary.BigEndian, &vByte); err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetBool(vByte != 0)
			}
		case "short":
			var vShort int16
			if err := binary.Read(r, binary.BigEndian, &vShort); err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetInt(int64(vShort))
			}
		case "ushort":
			var vShort uint16
			if err := binary.Read(r, binary.BigEndian, &vShort); err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetUint(uint64(vShort))
			}
		case "int":
			var vInt int32
			if err := binary.Read(r, binary.BigEndian, &vInt); err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetInt(int64(vInt))
			}
		case "long":
			var vLong int64
			if err := binary.Read(r, binary.BigEndian, &vLong); err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetInt(vLong)
			}
		case "float":
			var vFloat float32
			if err := binary.Read(r, binary.BigEndian, &vFloat); err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetFloat(float64(vFloat))
			}
		case "double":
			var vDouble float64
			if err := binary.Read(r, binary.BigEndian, &vDouble); err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetFloat(vDouble)
			}
		case "position":
			var packed int64
			if err := binary.Read(r, binary.BigEndian, &packed); err != nil {
				return err
			}

			x := int32(packed >> 38)
			z := int32((packed << 26) >> 38)
			y := int16((packed << 52) >> 52)

			pos := Position{X: x, Y: y, Z: z}

			if fieldVal.CanSet() {
				if fieldVal.Kind() == reflect.Ptr {
					if fieldVal.IsNil() {
						fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
					}
					fieldVal.Elem().Set(reflect.ValueOf(pos))
				} else {
					fieldVal.Set(reflect.ValueOf(pos))
				}
			}
		case "slot":
			var slot Slot

			var itemID int16
			if err := binary.Read(r, binary.BigEndian, &itemID); err != nil {
				return err
			}
			slot.ItemID = itemID

			if itemID != -1 {
				var count uint8
				if err := binary.Read(r, binary.BigEndian, &count); err != nil {
					return err
				}
				slot.ItemCount = count

				var damage int16
				if err := binary.Read(r, binary.BigEndian, &damage); err != nil {
					return err
				}
				slot.Damage = damage

				var nbtType byte
				if err := binary.Read(r, binary.BigEndian, &nbtType); err != nil {
					return err
				}

				if nbtType != 0 {
					if fieldVal.CanSet() {
						currentSlot, ok := fieldVal.Interface().(Slot)
						if ok && currentSlot.NBT != nil {
							slot.NBT = currentSlot.NBT
							if err := nbt.UnmarshalPayload(r, nbtType, slot.NBT); err != nil {
								return fmt.Errorf("gagal unmarshal NBT slot 1.8: %v", err)
							}
						} else {
							if err := nbt.SkipPayload(r, nbtType); err != nil {
								return err
							}
						}
					}
				}
			}
			if fieldVal.CanSet() {
				fieldVal.Set(reflect.ValueOf(slot))
			}

		case "bytearray":
			length, err := ReadVarInt(r)
			if err != nil {
				return err
			}
			buf := make([]byte, length)
			if _, err := io.ReadFull(r, buf); err != nil {
				return err
			}
			if fieldVal.CanSet() {
				fieldVal.SetBytes(buf)
			}
		}
	}
	return nil
}
