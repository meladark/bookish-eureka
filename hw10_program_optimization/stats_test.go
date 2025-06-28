//go:build !bench
// +build !bench

package hw10programoptimization

import (
	"archive/zip"
	"bytes"
	"os"
	"runtime/pprof"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDomainStat(t *testing.T) {
	data := `{"Id":1,"Name":"Howard Mendoza","Username":"0Oliver","Email":"aliquid_qui_ea@Browsedrive.gov","Phone":"6-866-899-36-79","Password":"InAQJvsq","Address":"Blackbird Place 25"}
{"Id":2,"Name":"Jesse Vasquez","Username":"qRichardson","Email":"mLynch@broWsecat.com","Phone":"9-373-949-64-00","Password":"SiZLeNSGn","Address":"Fulton Hill 80"}
{"Id":3,"Name":"Clarence Olson","Username":"RachelAdams","Email":"RoseSmith@Browsecat.com","Phone":"988-48-97","Password":"71kuz3gA5w","Address":"Monterey Park 39"}
{"Id":4,"Name":"Gregory Reid","Username":"tButler","Email":"5Moore@Teklist.net","Phone":"520-04-16","Password":"r639qLNu","Address":"Sunfield Park 20"}
{"Id":5,"Name":"Janice Rose","Username":"KeithHart","Email":"nulla@Linktype.com","Phone":"146-91-01","Password":"acSBF5","Address":"Russell Trail 61"}
{"Id":6,"Name":"Франческо Фирталини","Username":"McKvin","Email":"42@","Phone":"8-800-555-35","Password":"verystrong","Address":"No house no street"}`

	t.Run("find 'com'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "com")
		require.NoError(t, err)
		require.Equal(t, DomainStat{
			"browsecat.com": 2,
			"linktype.com":  1,
		}, result)
	})

	t.Run("find 'gov'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "gov")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"browsedrive.gov": 1}, result)
	})

	t.Run("find 'unknown'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "unknown")
		require.NoError(t, err)
		require.Equal(t, DomainStat{}, result)
	})
	t.Run("find 'net'", func(t *testing.T) {
		result, err := GetDomainStat(bytes.NewBufferString(data), "net")
		require.NoError(t, err)
		require.Equal(t, DomainStat{"teklist.net": 1}, result)
	})
}
func BenchmarkGetDomainStat(b *testing.B) {
	cpuProfile, _ := os.Create("cpu.prof")
	defer cpuProfile.Close()
	pprof.StartCPUProfile(cpuProfile)
	defer pprof.StopCPUProfile()
	memProfile, _ := os.Create("mem.prof")
	defer memProfile.Close()
	zr, err := zip.OpenReader("testdata/users.dat.zip")
	if err != nil {
		b.Fatalf("failed to open zip: %v", err)
	}
	defer zr.Close()
	zf, _ := zr.File[0].Open()
	defer zf.Close()
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(zf)
	if err != nil {
		b.Fatalf("failed to read file: %v", err)
	}
	data := buf.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		_, err := GetDomainStat(r, "com")
		if err != nil {
			b.Fatalf("GetDomainStat failed: %v", err)
		}
	}
	b.StopTimer()
	pprof.WriteHeapProfile(memProfile)
}
