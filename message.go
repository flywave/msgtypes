package msgtypes

import (
	"encoding/json"
	"errors"
	"io"
	"sort"

	"github.com/flywave/go-pbf"
)

const (
	defaultVersion = 10
)

var (
	ErrVersionChange     = errors.New("version change")
	ErrUnsupportedFormat = errors.New("unsupported format")
	ErrEmptyName         = errors.New("empty name")
	ErrBadChar           = errors.New("invalid char")
	ErrTooManyValues     = errors.New("more than one value in the record")
	ErrNoValues          = errors.New("no value or sum field found")
)

// https://datatracker.ietf.org/doc/html/rfc8428
type Record struct {
	Link        string   `json:"l,omitempty"`
	BaseName    string   `json:"bn,omitempty"`
	BaseTime    float64  `json:"bt,omitempty"`
	BaseUnit    string   `json:"bu,omitempty"`
	BaseVersion uint64   `json:"bver,omitempty"`
	BaseValue   float64  `json:"bv,omitempty"`
	BaseSum     float64  `json:"bs,omitempty"`
	Name        string   `json:"n,omitempty"`
	Unit        string   `json:"u,omitempty"`
	Time        float64  `json:"t,omitempty"`
	UpdateTime  float64  `json:"ut,omitempty"`
	Value       *float64 `json:"v,omitempty"`
	StringValue *string  `json:"vs,omitempty"`
	DataValue   *string  `json:"vd,omitempty"`
	BoolValue   *bool    `json:"vb,omitempty"`
	Sum         *float64 `json:"s,omitempty"`
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

func Normalize(p Records) (Records, error) {
	if err := Validate(p); err != nil {
		return Records{}, err
	}
	records := make([]Record, len(p))
	var bname string
	var btime float64
	var bsum float64
	var bunit string

	for i, r := range p {
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
	p = records
	sort.Sort(&p)
	return p, nil
}

func Validate(p Records) error {
	var bver uint64
	var bname string
	var bsum float64
	for _, r := range p {
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
		record.BaseVersion = reader.ReadUInt64()
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
}

func Decode(bytevals []byte) (records Records, err error) {
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
	writer.WriteUInt64(BaseVersionTag, record.BaseVersion)
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
	return nil
}

func Encode(records Records) ([]byte, error) {
	w := pbf.NewWriter()

	for _, record := range records {
		w.WriteMessage(RecordsTag, func(w *pbf.Writer) {
			encodeRecord(w, &record)
		})
	}

	return w.Finish(), nil
}
