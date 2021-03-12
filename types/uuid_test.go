package types

import (
	"encoding/json"
	"github.com/google/uuid"
	"testing"
)

func TestUUID_Scan(t *testing.T) {
	uid := uuid.New()
	newId := UUID{}
	err := newId.Scan(uid.String())
	if err != nil {
		t.Error(err)
	}

	if newId.String() != uid.String() {
		t.Errorf("convert value failed")
	}
}

func TestUUID_Value(t *testing.T) {
	uid := uuid.New()
	newId := UUID(uid)

	val, err := newId.Value()
	if err != nil {
		t.Error(err)
	}
	if val.(string) != uid.String() {
		t.Errorf("convert value failed")
	}
}

func TestUUID_JsonMarshal(t *testing.T) {
	type T struct {
		Id UUID
	}

	val := T{Id: NewUUID()}

	marsahlBytes, err := json.Marshal(&val)
	if err != nil {
		t.Error(err)
	}

	var val2 T
	err = json.Unmarshal(marsahlBytes, &val2)
	if err != nil {
		t.Error(err)
	}

	if val2.Id != val.Id {
		t.Errorf("UUID json marshal fail")
	}
}
