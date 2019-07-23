package wmm

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	MaxLegendreOrder = 12
)

var (
	wmmEpoch     time.Time
	wmmCofName   string
	wmmValidDate time.Time
	wmmGnm       [][]float64
	wmmHnm       [][]float64
	wmmdGnm      [][]float64
	wmmdHnm      [][]float64
)

func GetWMMCoefficients(n, m int, t time.Time) (gnm, hnm, dgnm, dhnm float64, err error) {
	if wmmEpoch.IsZero() {
		loadWMMCOF()
	}
	if n<0 || n>MaxLegendreOrder || m<0 || m>MaxLegendreOrder {
		return 0, 0, 0, 0, fmt.Errorf("n, m = (%d,%d) must be between 0 and %d",
			n, m, MaxLegendreOrder)
	}
	if m>n {
		return 0, 0, 0, 0, fmt.Errorf("m=%d must be less than n=%d", m, n)
	}
	if t.Sub(wmmValidDate) < 0 || t.Sub(wmmValidDate) > 5*SecondsPerYear*time.Second {
		return 0, 0, 0, 0,
			fmt.Errorf("requested date %v is outside of validity period beginning %v of WMM.COF file",
				t, wmmValidDate)
	}
	dt := DecimalYearsSinceEpoch(t, wmmEpoch)
	gnm = wmmGnm[n][m] + dt*wmmdGnm[n][m]
	hnm = wmmHnm[n][m] + dt*wmmdHnm[n][m]
	dgnm = wmmdGnm[n][m]
	dhnm = wmmdHnm[n][m]
	return gnm, hnm, dgnm, dhnm, nil
}

func loadWMMCOF() {
	data, err := Asset("WMM.COF")
	if err != nil {
		panic(err)
	}

	var (
		epoch float64
		n, m  int
	)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	// Read and parse header
	if !scanner.Scan() {
		panic("Could not read header line from WMM coefficient file")
	}
	dat := strings.Fields(scanner.Text())
	if epoch, err = strconv.ParseFloat(dat[0], 64); err != nil {
		panic("bad WMM.COF header epoch date")
	}
	wmmEpoch = DecimalYearsToTime(epoch)
	wmmCofName = dat[1]
	if wmmValidDate, err = time.Parse("01/02/2006", dat[2]); err != nil {
		panic("bad WMM.COF header valid date")
	}

	wmmGnm = make([][]float64, MaxLegendreOrder+1)
	wmmGnm[0] = []float64{0}
	wmmHnm = make([][]float64, MaxLegendreOrder+1)
	wmmHnm[0] = []float64{0}
	wmmdGnm = make([][]float64, MaxLegendreOrder+1)
	wmmdGnm[0] = []float64{0}
	wmmdHnm = make([][]float64, MaxLegendreOrder+1)
	wmmdHnm[0] = []float64{0}

	// Read and parse data
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
			wmmGnm[n] = make([]float64, n+1)
			wmmHnm[n] = make([]float64, n+1)
			wmmdGnm[n] = make([]float64, n+1)
			wmmdHnm[n] = make([]float64, n+1)
			curN = n
		}
		if wmmGnm[n][m], err = strconv.ParseFloat(s[2], 64); err != nil {
			panic("bad Gnm value in WMM.COF data file")
		}
		if wmmHnm[n][m], err = strconv.ParseFloat(s[3], 64); err != nil {
			panic("bad Hnm value in WMM.COF data file")
		}
		if wmmdGnm[n][m], err = strconv.ParseFloat(s[4], 64); err != nil {
			panic("bad dGnm value in WMM.COF data file")
		}
		if wmmdHnm[n][m], err = strconv.ParseFloat(s[5], 64); err != nil {
			panic("bad dHnm value in WMM.COF data file")
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}
}
