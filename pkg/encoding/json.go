package encoding

import (
	"encoding/json"

	"github.com/go-kratos/kratos/v2/encoding"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var jsonIter = jsoniter.ConfigFastest

// Name is the name registered for the json codec.
const Name = "json"

var (
	// MarshalOptions is a configurable JSON format marshaller.
	MarshalOptions = protojson.MarshalOptions{
		EmitUnpopulated: true,
	}
	// UnmarshalOptions is a configurable JSON format parser.
	UnmarshalOptions = protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
)

func GetJsonCodec() encoding.Codec {
	return privateCodec
}

func init() {
	extra.RegisterFuzzyDecoders()
}

func Init() {
	encoding.RegisterCodec(privateCodec)
}

// SpecialInterface 非常直接的返回，不再Marshal，性能优先的时候使用
type SpecialInterface interface {
	GetSpecialData() []byte
}

// codec is a Codec implementation with json.
type codec struct{}

var privateCodec = codec{}

func (codec) Marshal(v interface{}) ([]byte, error) {
	switch m := v.(type) {
	case json.Marshaler:
		return m.MarshalJSON()
	case proto.Message:
		return MarshalOptions.Marshal(m)
	case SpecialInterface:
		return m.GetSpecialData(), nil
	default:
		return jsonIter.Marshal(m)
	}
}

func (codec) Unmarshal(data []byte, v interface{}) error {
	switch m := v.(type) {
	case json.Unmarshaler:
		return m.UnmarshalJSON(data)
	case proto.Message:
		return UnmarshalOptions.Unmarshal(data, m)
	default:
		//rv := reflect.ValueOf(v)
		//for rv := rv; rv.Kind() == reflect.Ptr; {
		//	if rv.IsNil() {
		//		rv.Set(reflect.New(rv.Type().Elem()))
		//	}
		//	rv = rv.Elem()
		//}
		//if m, ok := reflect.Indirect(rv).Interface().(proto.Message); ok {
		//	return UnmarshalOptions.Unmarshal(data, m)
		//}
		return jsonIter.Unmarshal(data, m)
	}
}

func (codec) Name() string {
	return Name
}
