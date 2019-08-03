package wmm

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const (
	MaxLegendreOrder = 12
)

var (
	Epoch     DecimalYear
	COFName   string
	ValidDate time.Time
	Gnm       [][]float64
	Hnm       [][]float64
	DGnm      [][]float64
	DHnm      [][]float64
)

func GetWMMCoefficients(n, m int, t time.Time) (gnm, hnm, dgnm, dhnm float64, err error) {
	if Epoch==0 {
		LoadWMMCOF("")
	}
	if n<0 || n>MaxLegendreOrder || m<0 || m>MaxLegendreOrder {
		return 0, 0, 0, 0, fmt.Errorf("n, m = (%d,%d) must be between 0 and %d",
			n, m, MaxLegendreOrder)
	}
	if m>n {
		return 0, 0, 0, 0, fmt.Errorf("m=%d must be less than n=%d", m, n)
	}
	if t.Sub(ValidDate) < 0 || t.Sub(ValidDate) > 5*SecondsPerYear*time.Second {
		err = fmt.Errorf("requested date %v is outside of validity period beginning %v of WMM.COF file",
				t, ValidDate)
	}
	dt := float64(TimeToDecimalYears(t)- Epoch)
	gnm = Gnm[n][m] + dt*DGnm[n][m]
	hnm = Hnm[n][m] + dt*DHnm[n][m]
	dgnm = DGnm[n][m]
	dhnm = DHnm[n][m]
	return gnm, hnm, dgnm, dhnm, err
}

func LoadWMMCOF(fn string) {
	var (
		data []byte
		err   error
		epoch float64
		n, m  int
	)

	if fn=="" {
		data, err = Asset("WMM.COF")
	} else {
		data, err = ioutil.ReadFile(fn)
	}
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Read and parse header
	if !scanner.Scan() {
		panic("Could not read header line from WMM coefficient file")
	}
	dat := strings.Fields(scanner.Text())
	if epoch, err = strconv.ParseFloat(dat[0], 64); err != nil {
		panic("bad WMM.COF header epoch date")
	}
	Epoch = DecimalYear(epoch)
	COFName = dat[1]
	if ValidDate, err = time.Parse("01/02/2006", dat[2]); err != nil {
		panic("bad WMM.COF header valid date")
	}

	Gnm = make([][]float64, MaxLegendreOrder+1)
	Gnm[0] = []float64{0}
	Hnm = make([][]float64, MaxLegendreOrder+1)
	Hnm[0] = []float64{0}
	DGnm = make([][]float64, MaxLegendreOrder+1)
	DGnm[0] = []float64{0}
	DHnm = make([][]float64, MaxLegendreOrder+1)
	DHnm[0] = []float64{0}

	// Read and parse test_data
	curN := 0
	for scanner.Scan() {
		s := strings.Fields(scanner.Text())
		if len(s)<6 {
			continue
		}
		if n, err = strconv.Atoi(s[0]); err!=nil {
			panic("bad n value in WMM.COF data file")
		}
		if m, err = strconv.Atoi(s[1]); err!=nil {
			panic("bad m value in WMM.COF data file")
		}
		if n>curN {
			Gnm[n] = make([]float64, n+1)
			Hnm[n] = make([]float64, n+1)
			DGnm[n] = make([]float64, n+1)
			DHnm[n] = make([]float64, n+1)
			curN = n
		}
		if Gnm[n][m], err = strconv.ParseFloat(s[2], 64); err != nil {
			panic("bad Gnm value in WMM.COF data file")
		}
		if Hnm[n][m], err = strconv.ParseFloat(s[3], 64); err != nil {
			panic("bad Hnm value in WMM.COF data file")
		}
		if DGnm[n][m], err = strconv.ParseFloat(s[4], 64); err != nil {
			panic("bad DGnm value in WMM.COF data file")
		}
		if DHnm[n][m], err = strconv.ParseFloat(s[5], 64); err != nil {
			panic("bad DHnm value in WMM.COF data file")
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
