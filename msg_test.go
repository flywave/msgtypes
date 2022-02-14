package msgtypes

import "testing"

func TestEncodeDecode(t *testing.T) {
	recs := []Record{
		{Name: "test1"},
		{Name: "test2"},
	}

	data, err := Encode(recs)

	if err != nil {
		t.FailNow()
	}

	recs2, err := Decode(data)

	if err != nil {
		t.FailNow()
	}

	if len(recs) != len(recs2) {
		t.FailNow()
	}

}
