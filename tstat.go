package tstat

import (
	"io"
)

type FileStats struct {
	Cov   FileCov
	FnCov FuncCov
}

type Stats struct {
	Total   float64
	fileCov map[string]*FileStats
}

type Options struct {
	trimModule string
}

type ParseOpts func(*Options)

func TrimModule(name string) ParseOpts {
	return func(o *Options) {
		o.trimModule = name
	}
}

type parser struct {
	opts Options
}

func Parse(covProfile io.Reader, fnCov io.Reader, opts ...ParseOpts) (*Stats, error) {
	c, err := ParseCovProfile(covProfile, opts...)
	if err != nil {
		return nil, err
	}

	f, err := ParseFuncCoverage(fnCov, opts...)
	if err != nil {
		return nil, err
	}

	stats := map[string]*FileStats{}
	for name, fCov := range c.fileCov {
		stats[name] = &FileStats{Cov: fCov}
	}

	for name, fCov := range f.file {
		s, ok := stats[name]
		if ok {
			s.FnCov = fCov
			continue
		}

		stats[name] = &FileStats{FnCov: fCov}
	}

	return &Stats{Total: c.Total, fileCov: stats}, nil
}
