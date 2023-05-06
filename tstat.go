package tstat

import (
	"io"
)

type FileStats struct {
	Cov   FileCov
	FnCov FuncCov
}

type Coverage struct {
	Function  FunctionStats
	Statement StatementStats
}

type Stats struct {
	Coverage
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

func Parse(covProfile io.Reader, fnCov io.Reader, opts ...ParseOpts) (*Stats, error) {
	c, err := ParseCovProfile(covProfile, opts...)
	if err != nil {
		return nil, err
	}

	f, err := ParseFuncCoverage(fnCov, opts...)
	if err != nil {
		return nil, err
	}

	return &Stats{Coverage: Coverage{Function: f, Statement: c}}, nil
}
