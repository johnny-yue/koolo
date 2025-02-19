package pather

import (
	"encoding/gob"
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/area"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func BenchmarkGetPath(b *testing.B) {
	f, err := os.Open("data.bin")
	require.NoError(b, err)
	defer f.Close()
	var d data.Data
	dec := gob.NewDecoder(f)
	err = dec.Decode(&d)
	require.NoError(b, err)

	var p data.Position
	for _, l := range d.AdjacentLevels {
		if l.Area == area.DuranceOfHateLevel3 {
			p = l.Position
		}
	}
	require.NotEqual(b, data.Position{}, p)

	for i := 0; i < b.N; i++ {
		GetPath(d, p)
	}
}
