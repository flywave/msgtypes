package msgtypes

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"sort"

	"github.com/flywave/go-pbf"
	"github.com/fxamacker/cbor"
)

const (
	xmlns          = "urn:ietf:params:xml:ns:senml"
	defaultVersion = 10
)

type Format int

const (
	JSON Format = 1 + iota
	XML
	CBOR
	PROTO
)

var (
	ErrVersionChange     = errors.New("version change")
	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrEmptyName         = errors.New("empty name")
	ErrBadChar           = errors.New("invalid char")
	ErrTooManyValues     = errors.New("more than one value in the record")
	ErrNoValues          = errors.New("no value or sum field found")
)

type Record struct {
	XMLName     *bool      `json:"-" xml:"senml" cbor:"-"`
	Link        string     `json:"l,omitempty"  xml:"l,attr,omitempty" cbor:"-"`
	BaseName    string     `json:"bn,omitempty" xml:"bn,attr,omitempty" cbor:"-2,keyasint,omitempty"`
	BaseTime    float64    `json:"bt,omitempty" xml:"bt,attr,omitempty" cbor:"-3,keyasint,omitempty"`
	BaseUnit    string     `json:"bu,omitempty" xml:"bu,attr,omitempty" cbor:"-4,keyasint,omitempty"`
	BaseVersion uint       `json:"bver,omitempty" xml:"bver,attr,omitempty" cbor:"-1,keyasint,omitempty"`
	BaseValue   float64    `json:"bv,omitempty" xml:"bv,attr,omitempty" cbor:"-5,keyasint,omitempty"`
	BaseSum     float64    `json:"bs,omitempty" xml:"bs,attr,omitempty" cbor:"-6,keyasint,omitempty"`
	Name        string     `json:"n,omitempty" xml:"n,attr,omitempty" cbor:"0,keyasint,omitempty"`
	Unit        string     `json:"u,omitempty" xml:"u,attr,omitempty" cbor:"1,keyasint,omitempty"`
	Time        float64    `json:"t,omitempty" xml:"t,attr,omitempty" cbor:"6,keyasint,omitempty"`
	UpdateTime  float64    `json:"ut,omitempty" xml:"ut,attr,omitempty" cbor:"7,keyasint,omitempty"`
	Value       *float64   `json:"v,omitempty" xml:"v,attr,omitempty" cbor:"2,keyasint,omitempty"`
	StringValue *string    `json:"vs,omitempty" xml:"vs,attr,omitempty" cbor:"3,keyasint,omitempty"`
	DataValue   *string    `json:"vd,omitempty" xml:"vd,attr,omitempty" cbor:"8,keyasint,omitempty"`
	BoolValue   *bool      `json:"vb,omitempty" xml:"vb,attr,omitempty" cbor:"4,keyasint,omitempty"`
	CoordValue  *[]float64 `json:"vc,omitempty" xml:"vc,attr,omitempty" cbor:"9,keyasint,omitempty"`
	LongValue   *int64     `json:"vl,omitempty" xml:"vl,attr,omitempty" cbor:"10,keyasint,omitempty"`
	Sum         *float64   `json:"s,omitempty" xml:"s,attr,omitempty" cbor:"5,keyasint,omitempty"`
}

func (r *Record) ToJson() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func RecordFromJson(data io.Reader) *Record {
	var o *Record
	json.NewDecoder(data).Decode(&o)
	return o
}

type Records []Record

func (p Records) Len() int {
	return len(p)
}

func (p Records) Less(i, j int) bool {
	return p[i].Time < p[j].Time
}

func (p Records) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func Normalize(p Pack) (Pack, error) {
	if err := Validate(p); err != nil {
		return Pack{}, err
	}
	records := make([]Record, len(p.Records))
	var bname string
	var btime float64
	var bsum float64
	var bunit string

	for i, r := range p.Records {
		if r.BaseTime != 0 {
			btime = r.BaseTime
		}
		if r.BaseSum != 0 {
			bsum = r.BaseSum
		}
		if r.BaseUnit != "" {
			bunit = r.BaseUnit
		}
		if len(r.BaseName) > 0 {
			bname = r.BaseName
		}
		r.Name = bname + r.Name
		r.Time = btime + r.Time
		if r.Sum != nil {
			*r.Sum = bsum + *r.Sum
		}
		if r.Unit == "" {
			r.Unit = bunit
		}
		if r.Value != nil && r.BaseValue != 0 {
			*r.Value = r.BaseValue + *r.Value
		}
		if r.BaseVersion == defaultVersion {
			r.BaseVersion = 0
		}

		r.BaseTime = 0
		r.BaseValue = 0
		r.BaseUnit = ""
		r.BaseName = ""
		r.BaseSum = 0
		records[i] = r
	}
	p.Records = records
	sort.Sort(&p)
	return p, nil
}

func Validate(p Pack) error {
	var bver uint
	var bname string
	var bsum float64
	for _, r := range p.Records {
		if bver == 0 && r.BaseVersion != 0 {
			bver = r.BaseVersion
		}
		if bver != 0 && r.BaseVersion == 0 {
			r.BaseVersion = bver
		}
		if r.BaseVersion != bver {
			return ErrVersionChange
		}
		if r.BaseName != "" {
			bname = r.BaseName
		}
		if r.BaseSum != 0 {
			bsum = r.BaseSum
		}
		name := bname + r.Name
		if len(name) == 0 {
			return ErrEmptyName
		}
		var valCnt int
		if r.Value != nil {
			valCnt++
		}
		if r.BoolValue != nil {
			valCnt++
		}
		if r.DataValue != nil {
			valCnt++
		}
		if r.StringValue != nil {
			valCnt++
		}
		if r.CoordValue != nil {
			valCnt++
		}
		if r.LongValue != nil {
			valCnt++
		}
		if valCnt > 1 {
			return ErrTooManyValues
		}
		if r.Sum != nil || bsum != 0 {
			valCnt++
		}
		if valCnt < 1 {
			return ErrNoValues
		}
		if err := validateName(name); err != nil {
			return err
		}
	}
	return nil
}

func validateName(name string) error {
	l := name[0]
	if (l == '-') || (l == ':') || (l == '.') || (l == '/') || (l == '_') {
		return ErrBadChar
	}
	for _, l := range name {
		if (l < 'a' || l > 'z') && (l < 'A' || l > 'Z') && (l < '0' || l > '9') && (l != '-') && (l != ':') && (l != '.') && (l != '/') && (l != '_') {
			return ErrBadChar
		}
	}
	return nil
}

const (
	LinkTag        pbf.TagType = 1
	BaseNameTag    pbf.TagType = 2
	BaseTimeTag    pbf.TagType = 3
	BaseUnitTag    pbf.TagType = 4
	BaseVersionTag pbf.TagType = 5
	BaseValueTag   pbf.TagType = 6
	BaseSumTag     pbf.TagType = 7
	NameTag        pbf.TagType = 8
	UnitTag        pbf.TagType = 9
	TimeTag        pbf.TagType = 10
	UpdateTimeTag  pbf.TagType = 11
	ValueTag       pbf.TagType = 12
	StringValueTag pbf.TagType = 13
	DataValueTag   pbf.TagType = 14
	BoolValueTag   pbf.TagType = 15
	SumTag         pbf.TagType = 16
	CoordValueTag  pbf.TagType = 17
	LongValueTag   pbf.TagType = 18

	RecordsTag pbf.TagType = 1
)

func decodeRecordfunc(key pbf.TagType, val pbf.WireType, result interface{}, reader *pbf.Reader) {
	record := result.(*Record)
	if key == LinkTag && val == pbf.Bytes {
		record.Link = reader.ReadString()
	}
	if key == BaseNameTag && val == pbf.Bytes {
		record.BaseName = reader.ReadString()
	}
	if key == BaseTimeTag && val == pbf.Fixed64 {
		record.BaseTime = reader.ReadDouble()
	}
	if key == BaseUnitTag && val == pbf.Bytes {
		record.BaseUnit = reader.ReadString()
	}
	if key == BaseVersionTag && val == pbf.Varint {
		record.BaseVersion = uint(reader.ReadUInt64())
	}
	if key == BaseValueTag && val == pbf.Fixed64 {
		record.BaseValue = reader.ReadDouble()
	}
	if key == BaseSumTag && val == pbf.Fixed64 {
		record.BaseSum = reader.ReadDouble()
	}
	if key == NameTag && val == pbf.Bytes {
		record.Name = reader.ReadString()
	}
	if key == UnitTag && val == pbf.Bytes {
		record.Unit = reader.ReadString()
	}
	if key == TimeTag && val == pbf.Fixed64 {
		record.Time = reader.ReadDouble()
	}
	if key == UpdateTimeTag && val == pbf.Fixed64 {
		record.UpdateTime = reader.ReadDouble()
	}
	if key == ValueTag && val == pbf.Fixed64 {
		v := reader.ReadDouble()
		record.Value = &v
	}
	if key == StringValueTag && val == pbf.Bytes {
		v := reader.ReadString()
		record.StringValue = &v
	}
	if key == DataValueTag && val == pbf.Bytes {
		v := reader.ReadString()
		record.DataValue = &v
	}
	if key == BoolValueTag && val == pbf.Varint {
		v := reader.ReadBool()
		record.BoolValue = &v
	}
	if key == SumTag && val == pbf.Fixed64 {
		v := reader.ReadDouble()
		record.Sum = &v
	}
	if key == CoordValueTag && val == pbf.Bytes {
		v := reader.ReadPackedDouble()
		record.CoordValue = &v
	}
	if key == LongValueTag && val == pbf.Varint {
		v := reader.ReadInt64()
		record.LongValue = &v
	}
}

func decodeProto(bytevals []byte) (records Records, err error) {
	r := &pbf.Reader{Pbf: bytevals, Length: len(bytevals)}

	records = Records{}

	for r.Pos < r.Length {
		key, val := r.ReadTag()
		if key == RecordsTag && val == pbf.Bytes {
			record := &Record{}
			r.ReadMessage(decodeRecordfunc, record)
			records = append(records, *record)
		}
	}
	return records, nil
}

func encodeRecord(writer *pbf.Writer, record *Record) error {
	if record.Link != "" {
		writer.WriteString(LinkTag, record.Link)
	}
	if record.BaseName != "" {
		writer.WriteString(BaseNameTag, record.BaseName)
	}
	writer.WriteDouble(BaseTimeTag, record.BaseTime)
	if record.BaseUnit != "" {
		writer.WriteString(BaseUnitTag, record.BaseUnit)
	}
	writer.WriteUInt64(BaseVersionTag, uint64(record.BaseVersion))
	writer.WriteDouble(BaseValueTag, record.BaseValue)
	writer.WriteDouble(BaseSumTag, record.BaseSum)
	if record.Name != "" {
		writer.WriteString(NameTag, record.Name)
	}
	if record.Unit != "" {
		writer.WriteString(UnitTag, record.Unit)
	}
	writer.WriteDouble(TimeTag, record.Time)
	writer.WriteDouble(UpdateTimeTag, record.UpdateTime)

	if record.Value != nil {
		writer.WriteDouble(ValueTag, *record.Value)
	}
	if record.StringValue != nil {
		writer.WriteString(StringValueTag, *record.StringValue)
	}
	if record.DataValue != nil {
		writer.WriteString(DataValueTag, *record.DataValue)
	}
	if record.BoolValue != nil {
		writer.WriteBool(BoolValueTag, *record.BoolValue)
	}
	if record.Sum != nil {
		writer.WriteDouble(SumTag, *record.Sum)
	}
	if record.CoordValue != nil {
		writer.WritePackedDouble(CoordValueTag, *record.CoordValue)
	}
	if record.LongValue != nil {
		writer.WriteInt64(LongValueTag, *record.LongValue)
	}
	return nil
}

func encodeProto(records Records) ([]byte, error) {
	w := pbf.NewWriter()

	for _, record := range records {
		w.WriteMessage(RecordsTag, func(w *pbf.Writer) {
			encodeRecord(w, &record)
		})
	}

	return w.Finish(), nil
}

type Pack struct {
	XMLName *bool    `json:"-" xml:"sensml"`
	XMLNS   string   `json:"-" xml:"xmlns,attr"`
	Records []Record `xml:"senml"`
}

func (p *Pack) Len() int {
	return len(p.Records)
}

func (p *Pack) Less(i, j int) bool {
	return p.Records[i].Time < p.Records[j].Time
}

func (p *Pack) Swap(i, j int) {
	p.Records[i], p.Records[j] = p.Records[j], p.Records[i]
}

func Decode(msg []byte, format Format) (Pack, error) {
	var p Pack
	switch format {
	case JSON:
		if err := json.Unmarshal(msg, &p.Records); err != nil {
			return Pack{}, err
		}
	case XML:
		if err := xml.Unmarshal(msg, &p); err != nil {
			return Pack{}, err
		}
		p.XMLNS = xmlns
	case CBOR:
		if err := cbor.Unmarshal(msg, &p.Records); err != nil {
			return Pack{}, err
		}
	case PROTO:
		var err error
		if p.Records, err = decodeProto(msg); err != nil {
			return Pack{}, err
		}
	default:
		return Pack{}, ErrUnsupportedFormat
	}

	return p, Validate(p)
}

func Encode(p Pack, format Format) ([]byte, error) {
	switch format {
	case JSON:
		return json.Marshal(p.Records)
	case XML:
		p.XMLNS = xmlns
		return xml.Marshal(p)
	case CBOR:
		return cbor.Marshal(p.Records, cbor.CanonicalEncOptions())
	case PROTO:
		return encodeProto(p.Records)
	default:
		return nil, ErrUnsupportedFormat
	}
}
