package msgtypes

import "testing"

func TestEncodeDecode(t *testing.T) {
	v := 20.6
	cv := []float64{36.5, 118.4}
	recs := Pack{Records: []Record{
		{Name: "test1", Value: &v},
		{Name: "test2", CoordValue: &cv},
	},
	}

	data, err := Encode(recs, PROTO)

	if err != nil {
		t.FailNow()
	}

	recs2, err := Decode(data, PROTO)

	if err != nil {
		t.FailNow()
	}

	if len(recs.Records) != len(recs2.Records) {
		t.FailNow()
	}

}
