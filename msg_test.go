package msgtypes

import "testing"

func TestEncodeDecode(t *testing.T) {
	recs := Pack{Records: []Record{
		{Name: "test1"},
		{Name: "test2"},
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
