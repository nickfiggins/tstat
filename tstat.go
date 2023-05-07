package tstat

import (
	"io"
)

type Coverage struct {
	Function  FunctionStats
	Statement CoverStats
}

type CoverageParser struct {
	profile     io.Reader
	funcProfile io.Reader
	opts        []ParseOpts
}

func NewCoverageParser(profile, funcProfile io.Reader, opts ...ParseOpts) *CoverageParser {
	return &CoverageParser{profile: profile, funcProfile: funcProfile, opts: opts}
}

func (cp *CoverageParser) Parse(opts ...ParseOpts) (*Coverage, error) {
	opts = append(opts, cp.opts...)
	c, err := ParseCoverProfileFromReader(cp.profile, opts...)
	if err != nil {
		return nil, err
	}

	f, err := ParseFuncProfileFromReader(cp.funcProfile, opts...)
	if err != nil {
		return nil, err
	}

	return &Coverage{Function: f, Statement: c}, nil
}

type TestParser struct {
	out  io.Reader
	opts []ParseOpts
}

func NewTestParser(jsonOut io.Reader, opts ...ParseOpts) *TestParser {
	return &TestParser{out: jsonOut, opts: opts}
}

func (tp *TestParser) Parse(_ ...ParseOpts) (*TestStats, error) {
	return ParseTestOutput(tp.out)
}
