package tstat

import (
	"io"
)

type FileStats struct {
	Cov   FileCov
	FnCov FuncFileCov
}

type Stats struct {
	Total    float64
	coverage map[string]*FileStats
}

func Read(cov io.Reader, fn io.Reader) (*Stats, error) {
	c, err := ReadCover(cov)
	if err != nil {
		return nil, err
	}

	f, err := ReadFunc(fn)
	if err != nil {
		return nil, err
	}

	stats := map[string]*FileStats{}
	for name, fCov := range c.coverage {
		stats[name] = &FileStats{Cov: fCov}
	}

	for name, fCov := range f.file {
		s, ok := stats[name]
		if ok {
			s.FnCov = fCov
			continue
		}

		stats[name] = &FileStats{FnCov: s.FnCov}
	}

	return &Stats{Total: c.Total, coverage: stats}, nil
}
