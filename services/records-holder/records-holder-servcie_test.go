package records_holder

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func teardown(t *testing.T) {
	err := os.RemoveAll(baseDir)
	if err != nil {
		t.Fatal(err)
	}
}
func TestRecordsHolder_Record(t *testing.T) {
	defer teardown(t)
	rh := NewRecordsHolder(nil)
	err := rh.Record("1")
	assert.NoError(t, err)
	err = rh.Record("2")
	assert.NoError(t, err)
	err = rh.Record("1")
	assert.Error(t, err)
	err = rh.Stop("1")
	assert.NoError(t, err)
	err = rh.Stop("1")
	assert.Error(t, err)
	err = rh.Stop("2")
	assert.NoError(t, err)

}
