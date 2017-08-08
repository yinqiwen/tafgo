package tafgo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"reflect"
)

const (
	TafHeadeChar        = 0
	TafHeadeShort       = 1
	TafHeadeInt32       = 2
	TafHeadeInt64       = 3
	TafHeadeFloat       = 4
	TafHeadeDouble      = 5
	TafHeadeString1     = 6
	TafHeadeString4     = 7
	TafHeadeMap         = 8
	TafHeadeList        = 9
	TafHeadeStructBegin = 10
	TafHeadeStructEnd   = 11
	TafHeadeZeroTag     = 12
	TafHeadeSimpleList  = 13
)

var ErrBufferPeekOverflow = errors.New("Buffer overflow when peekBuf")
var ErrJceDecodeRequireNotExist = errors.New("require field not exist, tag:")
var ErrNotTafStruct = errors.New("Invalid 'TafStruct' value")

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "taf: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Ptr {
		return "taf: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "taf: Unmarshal(nil " + e.Type.String() + ")"
}

func encodeHeaderTag(tag uint8, tagType uint8, buf *bytes.Buffer) {
	if tag < 15 {
		b := byte((tag << 4) + tagType)
		buf.Write([]byte{b})
	} else {
		b1 := byte(tagType + 240)
		b2 := byte(tag)
		buf.Write([]byte{b1, b2})
	}
}

func encodeTagBoolValue(buf *bytes.Buffer, tag uint8, bv bool) error {
	if !bv {
		encodeHeaderTag(tag, uint8(TafHeadeZeroTag), buf)
	} else {
		encodeHeaderTag(tag, uint8(TafHeadeChar), buf)
		buf.Write([]byte{byte(1)})
	}
	return nil
}
func encodeTagInt8Value(buf *bytes.Buffer, tag uint8, bv int8) error {
	if bv == 0 {
		encodeHeaderTag(tag, uint8(TafHeadeZeroTag), buf)
	} else {
		encodeHeaderTag(tag, uint8(TafHeadeChar), buf)
		buf.Write([]byte{byte(bv)})
	}
	return nil
}
func encodeTagShortValue(buf *bytes.Buffer, tag uint8, sv int16) error {
	if sv < (-128) && sv <= 127 {
		return encodeTagInt8Value(buf, tag, int8(sv))
	} else {
		encodeHeaderTag(tag, uint8(TafHeadeShort), buf)
		binary.Write(buf, binary.BigEndian, sv)
	}
	return nil
}
func encodeTagIntValue(buf *bytes.Buffer, tag uint8, iv int32) error {
	if iv >= (-32768) && iv <= 32767 {
		return encodeTagShortValue(buf, tag, int16(iv))
	} else {
		encodeHeaderTag(tag, uint8(TafHeadeInt32), buf)
		binary.Write(buf, binary.BigEndian, iv)
	}
	return nil
}
func encodeTagLongValue(buf *bytes.Buffer, tag uint8, iv int64) error {
	if iv >= (-2147483647-1) && iv <= 2147483647 {
		return encodeTagIntValue(buf, tag, int32(iv))
	} else {
		encodeHeaderTag(tag, uint8(TafHeadeInt64), buf)
		binary.Write(buf, binary.BigEndian, iv)
	}
	return nil
}
func encodeTagFloatValue(buf *bytes.Buffer, tag uint8, fv float32) error {
	encodeHeaderTag(tag, uint8(TafHeadeFloat), buf)
	binary.Write(buf, binary.BigEndian, fv)
	return nil
}
func encodeTagDoubleValue(buf *bytes.Buffer, tag uint8, dv float64) error {
	encodeHeaderTag(tag, uint8(TafHeadeDouble), buf)
	binary.Write(buf, binary.BigEndian, dv)
	return nil
}

func encodeTagStringValue(buf *bytes.Buffer, tag uint8, str string) error {
	if len(str) > 255 {
		encodeHeaderTag(tag, uint8(TafHeadeString4), buf)
		slen := uint32(len(str))
		binary.Write(buf, binary.BigEndian, slen)
	} else {
		encodeHeaderTag(tag, uint8(TafHeadeString1), buf)
		buf.Write([]byte{byte(len(str))})
	}
	buf.Write([]byte(str))
	return nil
}

func encodeValueWithTag(buf *bytes.Buffer, tag uint8, v *reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Bool:
		bv := v.Bool()
		return encodeTagBoolValue(buf, tag, bv)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return encodeTagLongValue(buf, tag, v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return encodeTagLongValue(buf, tag, int64(v.Uint()))
	case reflect.String:
		str := v.String()
		return encodeTagStringValue(buf, tag, str)
	case reflect.Float32:
		return encodeTagFloatValue(buf, tag, float32(v.Float()))
	case reflect.Float64:
		return encodeTagDoubleValue(buf, tag, v.Float())
	case reflect.Array, reflect.Slice:
		if v.IsNil() {
			rv := reflect.MakeSlice(v.Type(), 0, 0)
			v = &rv
		}
		if v.Elem().Type().Kind() == reflect.Uint8 {
			encodeHeaderTag(tag, uint8(TafHeadeSimpleList), buf)
			encodeHeaderTag(0, uint8(TafHeadeChar), buf)
			encodeTagIntValue(buf, 0, int32(v.Len()))
			buf.Write(v.Bytes())
		} else {
			encodeHeaderTag(tag, uint8(TafHeadeList), buf)
			if v.Type().Kind() == reflect.Slice && v.IsNil() {
				encodeTagIntValue(buf, 0, 0)
			} else {
				encodeTagIntValue(buf, 0, int32(v.Len()))
				for i := 0; i < v.Len(); i++ {
					iv := v.Index(i)
					encodeValueWithTag(buf, 0, &iv)
				}
			}
		}
		return nil
	case reflect.Map:
		encodeHeaderTag(tag, uint8(TafHeadeMap), buf)
		if v.IsNil() {
			encodeTagIntValue(buf, 0, 0)
		} else {
			ks := v.MapKeys()
			encodeTagIntValue(buf, 0, int32(len(ks)))
			for i := 0; i < len(ks); i++ {
				encodeValueWithTag(buf, 0, &(ks[i]))
				vv := v.MapIndex(ks[i])
				encodeValueWithTag(buf, 1, &vv)
			}
		}
		return nil
	case reflect.Ptr:
		rv := reflect.Indirect(*v)
		return encodeValueWithTag(buf, tag, &rv)
	case reflect.Interface:
		rv := reflect.ValueOf(v.Interface())
		return encodeValueWithTag(buf, tag, &rv)
	case reflect.Struct:
		encodeHeaderTag(tag, uint8(TafHeadeStructBegin), buf)
		ts, ok := v.Interface().(TafStruct)
		if !ok {
			log.Printf("Invalid type:%v", v.Type())
		} else {
			ts.Encode(buf)
		}
		// num := v.NumField()
		// for i := 0; i < num; i++ {
		// 	fv := v.Field(i)
		// 	tagstr := v.Type().Field(i).Tag.Get("tag")
		// 	if len(tagstr) > 0 {
		// 		tag, _ := strconv.Atoi(tagstr)
		// 		encodeValueWithTag(buf, uint8(tag), &fv)
		// 	}
		// }
		encodeHeaderTag(0, uint8(TafHeadeStructEnd), buf)
	}
	return nil
}

// func EncodeTagValue(v interface{}, tag uint8, buf *bytes.Buffer) error {
// 	rv := reflect.ValueOf(v)
// 	return encodeValueWithTag(buf, tag, &rv)
// }

func peekTypeTag(buf *bytes.Buffer) (uint8, uint8, int, error) {
	if buf.Len() < 1 {
		return 0, 0, 0, ErrBufferPeekOverflow
	}
	typeTag := uint8(buf.Bytes()[0])
	tmpTag := typeTag >> 4
	typeValue := (typeTag & 0x0F)
	if tmpTag == 15 {
		tmpTag = uint8(buf.Bytes()[1])
		return tmpTag, typeValue, 2, nil
	} else {
		return tmpTag, typeValue, 1, nil
	}
}

func skipOneField(buf *bytes.Buffer) error {
	_, headType, len, err := peekTypeTag(buf)
	if nil != err {
		return err
	}
	buf.Next(len)
	return skipField(buf, headType)
}

func skipToStructEnd(buf *bytes.Buffer) error {
	for buf.Len() > 0 {
		_, headType, len, err := peekTypeTag(buf)
		if nil != err {
			return err
		}
		buf.Next(len)
		err = skipField(buf, headType)
		if nil != err {
			return err
		}
		if headType == TafHeadeStructEnd {
			break
		}

	}
	return nil
}

func skipField(buf *bytes.Buffer, typeValue uint8) error {
	switch typeValue {
	case TafHeadeChar:
		buf.Next(1)
	case TafHeadeShort:
		buf.Next(2)
	case TafHeadeInt32:
		buf.Next(4)
	case TafHeadeInt64:
		buf.Next(8)
	case TafHeadeFloat:
		buf.Next(4)
	case TafHeadeDouble:
		buf.Next(8)
	case TafHeadeString1:
		if buf.Len() < 1 {
			return ErrBufferPeekOverflow
		}
		len := uint8(buf.Bytes()[0])
		buf.Next(int(len + 1))
	case TafHeadeString4:
		len := uint32(0)
		err := binary.Read(buf, binary.BigEndian, &len)
		if nil != err {
			return err
		}
		buf.Next(int(len))
	case TafHeadeMap:
		size, err := decodeTagIntValue(buf, 0, true)
		if nil != err {
			return err
		}
		for i := int32(0); i < (size * 2); i++ {
			err = skipOneField(buf)
			if nil != err {
				return err
			}
		}
	case TafHeadeList:
		size, err := decodeTagIntValue(buf, 0, true)
		if nil != err {
			return err
		}
		for i := int32(0); i < size; i++ {
			err = skipOneField(buf)
			if nil != err {
				return err
			}
		}
	case TafHeadeSimpleList:
		_, headType, len, err := peekTypeTag(buf)
		if nil != err {
			return err
		}
		buf.Next(len)
		if headType != TafHeadeChar {
			return fmt.Errorf("skipField with invalid type, type value: %d, %d.", typeValue, headType)
		}
		size, err := decodeTagIntValue(buf, 0, true)
		if nil != err {
			return err
		}
		buf.Next(int(size))
	case TafHeadeStructBegin:
		err := skipToStructEnd(buf)
		if nil != err {
			return err
		}
	case TafHeadeStructEnd:
		break
	case TafHeadeZeroTag:
		break
	default:
		return fmt.Errorf("skipField with invalid type, type value:%d.", typeValue)
	}
	return nil
}

func skipToTag(buf *bytes.Buffer, tag uint8) (bool, uint8, uint8, error) {
	for buf.Len() > 0 {
		nextHeadTag, nextHeadType, len, err := peekTypeTag(buf)
		if nil != err {
			return false, 0, 0, err
		}
		if nextHeadType == TafHeadeStructEnd || tag < nextHeadTag {
			return false, 0, 0, nil
		}
		if tag == nextHeadTag {
			buf.Next(len)
			return true, nextHeadType, nextHeadTag, nil
		}
		buf.Next(int(len))
		skipField(buf, nextHeadType)
	}
	return false, 0, 0, nil
}

func decodeTagBoolValue(buf *bytes.Buffer, tag uint8, required bool) (bool, error) {
	v, err := decodeTagIntegerValue(buf, tag, required, TafHeadeChar)
	if nil != err {
		return false, err
	}
	if v > 0 {
		return true, nil
	}
	return false, nil
}

func decodeTagCharValue(buf *bytes.Buffer, tag uint8, required bool) (byte, error) {
	v, err := decodeTagIntegerValue(buf, tag, required, TafHeadeChar)
	return byte(v), err
}

func decodeTagInt8Value(buf *bytes.Buffer, tag uint8, required bool) (int8, error) {
	v, err := decodeTagIntegerValue(buf, tag, required, TafHeadeChar)
	return int8(v), err
}
func decodeTagUInt8Value(buf *bytes.Buffer, tag uint8, required bool) (uint8, error) {
	v, err := decodeTagIntegerValue(buf, tag, required, TafHeadeShort)
	return uint8(v), err
}

func decodeTagShortValue(buf *bytes.Buffer, tag uint8, required bool) (int16, error) {
	v, err := decodeTagIntegerValue(buf, tag, required, TafHeadeShort)
	return int16(v), err
}
func decodeTagUInt16Value(buf *bytes.Buffer, tag uint8, required bool) (uint16, error) {
	v, err := decodeTagIntegerValue(buf, tag, required, TafHeadeInt32)
	return uint16(v), err
}
func decodeTagIntValue(buf *bytes.Buffer, tag uint8, required bool) (int32, error) {
	v, err := decodeTagIntegerValue(buf, tag, required, TafHeadeInt32)
	return int32(v), err
}
func decodeTagUInt32Value(buf *bytes.Buffer, tag uint8, required bool) (uint32, error) {
	v, err := decodeTagIntegerValue(buf, tag, required, TafHeadeInt64)
	return uint32(v), err
}
func decodeTagLongValue(buf *bytes.Buffer, tag uint8, required bool) (int64, error) {
	return decodeTagIntegerValue(buf, tag, required, TafHeadeInt64)
}

func decodeTagIntegerValue(buf *bytes.Buffer, tag uint8, required bool, typeValue uint8) (int64, error) {
	flag, headType, _, err := skipToTag(buf, tag)
	if nil != err {
		return 0, err
	}
	if flag {
		if headType > typeValue && headType != TafHeadeZeroTag {
			return 0, fmt.Errorf("read 'Integer' type mismatch, tag: %d, get type: %d", tag, headType)
		}
		switch headType {
		case TafHeadeZeroTag:
			return 0, nil
		case TafHeadeChar:
			if buf.Len() < 1 {
				return 0, ErrBufferPeekOverflow
			}
			return int64(buf.Next(1)[0]), nil
		case TafHeadeShort:
			if buf.Len() < 2 {
				return 0, ErrBufferPeekOverflow
			}
			v := int16(0)
			err := binary.Read(buf, binary.BigEndian, &v)
			return int64(v), err
		case TafHeadeInt32:
			if buf.Len() < 4 {
				return 0, ErrBufferPeekOverflow
			}
			v := int32(0)
			err := binary.Read(buf, binary.BigEndian, &v)
			return int64(v), err
		case TafHeadeInt64:
			if buf.Len() < 8 {
				return 0, ErrBufferPeekOverflow
			}
			v := int64(0)
			err := binary.Read(buf, binary.BigEndian, &v)
			return v, err
		default:
			return 0, fmt.Errorf("read 'Integer' type mismatch, tag: %d, get type: %d", tag, headType)
		}
	} else {
		if required {
			return 0, fmt.Errorf("require field not exist, tag:%d", tag)
		}
	}
	return 0, nil
}
func decodeTagFloatValue(buf *bytes.Buffer, tag uint8, required bool) (float32, error) {
	v, err := decodeTagFloatDoubleValue(buf, tag, required, TafHeadeFloat)
	return float32(v), err
}
func decodeTagDoubleValue(buf *bytes.Buffer, tag uint8, required bool) (float64, error) {
	return decodeTagFloatDoubleValue(buf, tag, required, TafHeadeDouble)
}

func decodeTagFloatDoubleValue(buf *bytes.Buffer, tag uint8, required bool, typeValue uint8) (float64, error) {
	flag, headType, _, err := skipToTag(buf, tag)
	if nil != err {
		return 0, err
	}
	if flag {
		if headType > typeValue {
			return 0, fmt.Errorf("read 'Integer' type mismatch, tag: %d, get type: %d.", tag, headType)
		}
		switch headType {
		case TafHeadeZeroTag:
			return 0, nil
		case TafHeadeFloat:
			if buf.Len() < 4 {
				return 0, ErrBufferPeekOverflow
			}
			v := float32(0)
			err := binary.Read(buf, binary.BigEndian, &v)
			return float64(v), err
		case TafHeadeDouble:
			if buf.Len() < 8 {
				return 0, ErrBufferPeekOverflow
			}
			v := float64(0)
			err := binary.Read(buf, binary.BigEndian, &v)
			return v, err
		default:
			return 0, fmt.Errorf("read 'Float/Double' type mismatch, tag: %d, get type: %d.", tag, headType)
		}
	} else {
		if required {
			return 0, fmt.Errorf("require field not exist, tag:%d", tag)
		}
	}
	return float64(0), nil
}
func decodeTagStringValue(buf *bytes.Buffer, tag uint8, required bool) (string, error) {
	flag, headType, _, err := skipToTag(buf, tag)
	if nil != err {
		return "", err
	}
	if flag {
		strLen := 0
		switch headType {
		case TafHeadeString1:
			if buf.Len() < 1 {
				return "", ErrBufferPeekOverflow
			}
			strLen = int(buf.Next(1)[0])
		case TafHeadeString4:
			if buf.Len() < 4 {
				return "", ErrBufferPeekOverflow
			}
			len := int32(0)
			binary.Read(buf, binary.BigEndian, &len)
			strLen = int(len)
		default:
			return "", fmt.Errorf("read 'Integer' type mismatch, tag: %d, get type: %d.", tag, headType)
		}
		if buf.Len() < strLen {
			return "", ErrBufferPeekOverflow
		}
		return string(buf.Next(strLen)), nil
	} else {
		if required {
			return "", fmt.Errorf("require field not exist, tag:%d", tag)
		}
	}
	return "", nil
}

func decodeTagValue(buf *bytes.Buffer, tag uint8, required bool, v *reflect.Value) error {
	switch v.Type().Kind() {
	case reflect.Bool:
		b, err := decodeTagBoolValue(buf, tag, required)
		if nil == err {
			v.SetBool(b)
		} else {
			return err
		}
	case reflect.Int8:
		b, err := decodeTagInt8Value(buf, tag, required)
		if nil == err {
			v.SetInt(int64(b))
		} else {
			return err
		}
	case reflect.Uint8:
		b, err := decodeTagUInt8Value(buf, tag, required)
		if nil == err {
			v.SetInt(int64(b))
		} else {
			return err
		}
	case reflect.Int16:
		b, err := decodeTagShortValue(buf, tag, required)
		if nil == err {
			v.SetInt(int64(b))
		} else {
			return err
		}
	case reflect.Uint16:
		b, err := decodeTagUInt16Value(buf, tag, required)
		if nil == err {
			v.SetInt(int64(b))
		} else {
			return err
		}
	case reflect.Int32:
		b, err := decodeTagIntValue(buf, tag, required)
		if nil == err {
			v.SetInt(int64(b))
		} else {
			return err
		}
	case reflect.Uint32:
		b, err := decodeTagUInt32Value(buf, tag, required)
		if nil == err {
			v.SetInt(int64(b))
		} else {
			return err
		}
	case reflect.Int64:
		b, err := decodeTagLongValue(buf, tag, required)
		if nil == err {
			v.SetInt(int64(b))
		} else {
			return err
		}
	case reflect.Float32:
		b, err := decodeTagFloatValue(buf, tag, required)
		if nil == err {
			v.SetFloat(float64(b))
		} else {
			return err
		}
	case reflect.Float64:
		b, err := decodeTagDoubleValue(buf, tag, required)
		if nil == err {
			v.SetFloat(b)
		} else {
			return err
		}
	case reflect.String:
		b, err := decodeTagStringValue(buf, tag, required)
		if nil == err {
			v.SetString(b)
		} else {
			return err
		}
	case reflect.Array, reflect.Slice:
		if v.IsNil() {
			v.Set(reflect.MakeSlice(v.Type(), 0, 0))
		}
		switch v.Type().Elem().Kind() {
		case reflect.Uint8:
			var b []byte
			err := DecodeTagBytesValue(buf, &b, tag, required)
			if nil != err {
				return err
			}
			v.SetBytes(b)
			return nil
		case reflect.String:
			var sv []string
			err := DecodeTagStringsValue(buf, &sv, tag, required)
			if nil != err {
				return err
			}
			v.Set(reflect.ValueOf(sv))
			return nil
		default:
			flag, headType, _, err := skipToTag(buf, tag)
			if nil != err {
				return err
			}
			if flag {
				switch headType {
				case TafHeadeList:
					vectorSize, err := decodeTagIntValue(buf, 0, true)
					if nil != err {
						return err
					}
					sv := reflect.MakeSlice(v.Type(), int(vectorSize), int(vectorSize))
					for i := 0; i < int(vectorSize); i++ {
						iv := sv.Index(i)
						err = decodeTagValue(buf, 0, true, &(iv))
						if nil != err {
							return err
						}
					}
					v.Set(sv)
				default:
					return fmt.Errorf("read 'vector' type mismatch, tag: %d, get type: %d.", tag, headType)
				}
			} else {
				if required {
					return fmt.Errorf("require field not exist, tag:%d", tag)
				}
			}
		}
	case reflect.Map:
		flag, headType, _, err := skipToTag(buf, tag)
		if nil != err {
			return err
		}
		if flag {
			switch headType {
			case TafHeadeMap:
				mapSize, err := decodeTagIntValue(buf, 0, true)
				if nil != err {
					return err
				}
				vm := reflect.MakeMap(v.Type())
				for i := 0; i < int(mapSize); i++ {
					kv := reflect.New(v.Type().Key()).Elem()
					vv := reflect.New(v.Type().Elem()).Elem()
					err = decodeTagValue(buf, 0, true, &(kv))
					if nil != err {
						return err
					}
					err = decodeTagValue(buf, 1, true, &(vv))
					if nil != err {
						return err
					}
					vm.SetMapIndex(kv, vv)
				}
				v.Set(vm)
			default:
				return fmt.Errorf("read 'map' type mismatch, tag: %d, get type: %d.", tag, headType)
			}
		} else {
			if required {
				return fmt.Errorf("require field not exist, tag:%d", tag)
			}
		}
	case reflect.Ptr:
		if v.IsNil() {
			return &InvalidUnmarshalError{reflect.TypeOf(v)}
		}
		xv := v.Elem()
		return decodeTagValue(buf, tag, required, &xv)
	case reflect.Struct:
		ts, ok := v.Addr().Interface().(TafStruct)
		if ok {
			return DecodeTagStructValue(buf, ts, tag, required)
		}
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	default:
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	return nil
}

type TafStruct interface {
	Encode(buf *bytes.Buffer) error
	Decode(buf *bytes.Buffer) error
}

func EncodeTagStructValue(buf *bytes.Buffer, v TafStruct, tag uint8) error {
	encodeHeaderTag(tag, uint8(TafHeadeStructBegin), buf)
	v.Encode(buf)
	encodeHeaderTag(0, uint8(TafHeadeStructEnd), buf)
	return nil
}
func EncodeTagInt64Value(buf *bytes.Buffer, v int64, tag uint8) error {
	encodeTagLongValue(buf, tag, int64(v))
	return nil
}
func EncodeTagInt32Value(buf *bytes.Buffer, v int32, tag uint8) error {
	encodeTagLongValue(buf, tag, int64(v))
	return nil
}
func EncodeTagInt16Value(buf *bytes.Buffer, v int16, tag uint8) error {
	encodeTagLongValue(buf, tag, int64(v))
	return nil
}
func EncodeTagInt8Value(buf *bytes.Buffer, v int8, tag uint8) error {
	encodeTagLongValue(buf, tag, int64(v))
	return nil
}
func EncodeTagBoolValue(buf *bytes.Buffer, v bool, tag uint8) error {
	if v {
		encodeTagInt8Value(buf, tag, 1)
	} else {
		encodeTagInt8Value(buf, tag, 0)
	}

	return nil
}
func EncodeTagFloat32Value(buf *bytes.Buffer, v float32, tag uint8) error {
	encodeTagFloatValue(buf, tag, v)
	return nil
}
func EncodeTagFloat64Value(buf *bytes.Buffer, v float64, tag uint8) error {
	encodeTagDoubleValue(buf, tag, v)
	return nil
}
func EncodeTagByteValue(buf *bytes.Buffer, v byte, tag uint8) error {
	encodeTagInt8Value(buf, tag, int8(v))
	return nil
}

func EncodeTagBytesValue(buf *bytes.Buffer, v []byte, tag uint8) error {
	encodeHeaderTag(tag, uint8(TafHeadeSimpleList), buf)
	encodeHeaderTag(0, uint8(TafHeadeChar), buf)
	EncodeTagInt32Value(buf, int32(len(v)), 0)
	buf.Write(v)
	return nil
}
func EncodeTagStringValue(buf *bytes.Buffer, v string, tag uint8) error {
	if len(v) > 255 {
		encodeHeaderTag(tag, uint8(TafHeadeString4), buf)
		vlen := uint32(len(v))
		binary.Write(buf, binary.BigEndian, vlen)
	} else {
		encodeHeaderTag(tag, uint8(TafHeadeString1), buf)
		buf.Write([]byte{byte(len(v))})
	}
	buf.Write([]byte(v))
	return nil
}
func EncodeTagStringsValue(buf *bytes.Buffer, v []string, tag uint8) error {
	encodeHeaderTag(tag, uint8(TafHeadeList), buf)
	EncodeTagInt32Value(buf, int32(len(v)), 0)
	for _, s := range v {
		EncodeTagStringValue(buf, s, 0)
	}
	return nil
}

func EncodeTagVectorValue(buf *bytes.Buffer, v interface{}, tag uint8) error {
	val := reflect.ValueOf(v)
	//tafStructType := reflect.TypeOf((*TafStruct)(nil)).Elem()
	if val.Kind() == reflect.Array || val.Kind() == reflect.Slice {
		encodeHeaderTag(tag, uint8(TafHeadeList), buf)
		EncodeTagInt32Value(buf, int32(val.Len()), 0)
		for i := 0; i < val.Len(); i++ {
			e := val.Index(i)
			ts, ok := e.Addr().Interface().(TafStruct)
			if ok {
				EncodeTagStructValue(buf, ts, 0)
			} else {
				encodeValueWithTag(buf, tag, &e)
			}
		}
	} else {
		return ErrNotTafStruct
	}
	return nil
}

func EncodeTagMapValue(buf *bytes.Buffer, v interface{}, tag uint8) error {
	val := reflect.ValueOf(v)
	encodeValueWithTag(buf, tag, &val)
	return nil
}

func DecodeTagByteValue(buf *bytes.Buffer, v *byte, tag uint8, required bool) error {
	tv, err := decodeTagInt8Value(buf, tag, required)
	if nil != err {
		return err
	}
	*v = byte(tv)
	return nil
}

func DecodeTagBoolValue(buf *bytes.Buffer, v *bool, tag uint8, required bool) error {
	tv, err := decodeTagInt8Value(buf, tag, required)
	if nil != err {
		return err
	}
	if tv == 0 {
		*v = false
	} else {
		*v = true
	}
	return nil
}

func DecodeTagInt8Value(buf *bytes.Buffer, v *int8, tag uint8, required bool) error {
	var err error
	*v, err = decodeTagInt8Value(buf, tag, required)
	return err
}
func DecodeTagInt16Value(buf *bytes.Buffer, v *int16, tag uint8, required bool) error {
	var err error
	*v, err = decodeTagShortValue(buf, tag, required)
	return err
}
func DecodeTagInt32Value(buf *bytes.Buffer, v *int32, tag uint8, required bool) error {
	var err error
	*v, err = decodeTagIntValue(buf, tag, required)
	return err
}
func DecodeTagInt64Value(buf *bytes.Buffer, v *int64, tag uint8, required bool) error {
	var err error
	*v, err = decodeTagLongValue(buf, tag, required)
	return err
}
func DecodeTagFloat64Value(buf *bytes.Buffer, v *float64, tag uint8, required bool) error {
	var err error
	*v, err = decodeTagDoubleValue(buf, tag, required)
	return err
}
func DecodeTagFloat32Value(buf *bytes.Buffer, v *float32, tag uint8, required bool) error {
	var err error
	*v, err = decodeTagFloatValue(buf, tag, required)
	return err
}

func DecodeTagStringValue(buf *bytes.Buffer, v *string, tag uint8, required bool) error {
	var err error
	*v, err = decodeTagStringValue(buf, tag, required)
	return err
}

func DecodeTagBytesValue(buf *bytes.Buffer, v *[]byte, tag uint8, required bool) error {
	flag, headType, _, err := skipToTag(buf, tag)
	if nil != err {
		return err
	}
	if !flag {
		if required {
			return fmt.Errorf("require field not exist, tag:%d", tag)
		}
		return nil
	}
	if headType != TafHeadeSimpleList {
		return fmt.Errorf("read 'vector<byte>' type mismatch, tag: %d, get type: %d", tag, headType)
	}
	_, cheadType, clen, err := peekTypeTag(buf)
	if nil != err {
		return err
	}
	buf.Next(clen)
	if cheadType != TafHeadeChar {
		return fmt.Errorf("type mismatch, tag: %d, type: %d, %d", tag, headType, cheadType)
	}
	vlen, err := decodeTagIntValue(buf, 0, true)
	if nil != err {
		return err
	}
	if buf.Len() < int(vlen) {
		return ErrBufferPeekOverflow
	}
	*v = buf.Next(int(vlen))
	return nil
}
func DecodeTagStringsValue(buf *bytes.Buffer, v *[]string, tag uint8, required bool) error {
	flag, headType, _, err := skipToTag(buf, tag)
	if nil != err {
		return err
	}
	if !flag {
		if required {
			return fmt.Errorf("require field not exist, tag:%d", tag)
		}
		return nil
	}
	if headType != TafHeadeList {
		return fmt.Errorf("read 'vector<string>' type mismatch, tag: %d, get type: %d", tag, headType)
	}
	vlen, err := decodeTagIntValue(buf, 0, true)
	if nil != err {
		return err
	}
	sv := make([]string, int(vlen))
	*v = sv
	for i := 0; i < int(vlen); i++ {
		err = DecodeTagStringValue(buf, &(sv[i]), 0, true)
		if nil != err {
			return err
		}
	}
	return nil
}

func DecodeTagMapValue(buf *bytes.Buffer, v interface{}, tag uint8, required bool) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	return decodeTagValue(buf, tag, required, &rv)
}

func DecodeTagVectorValue(buf *bytes.Buffer, v interface{}, tag uint8, required bool) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	return decodeTagValue(buf, tag, required, &rv)
}

func DecodeTagStructValue(buf *bytes.Buffer, v TafStruct, tag uint8, required bool) error {
	flag, headType, _, err := skipToTag(buf, tag)
	if nil != err {
		return err
	}
	if !flag {
		if required {
			return fmt.Errorf("require field not exist, tag:%d", tag)
		}
		return nil
	}
	if headType != TafHeadeStructBegin {
		return fmt.Errorf("read 'struct' type mismatch, tag: %d, get type: %d", tag, headType)
	}
	err = v.Decode(buf)
	if nil != err {
		return err
	}
	skipToStructEnd(buf)
	return nil
}
