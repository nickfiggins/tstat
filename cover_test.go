package tstat

import (
	"reflect"
	"testing"
)

func TestCoverage_Package(t *testing.T) {
	type fields struct {
		Percent  float64
		Packages []*PackageCoverage
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *PackageCoverage
	}{
		{
			name: "Package found",
			fields: fields{0, []*PackageCoverage{
				{Name: "pkg1", Percent: 50, Files: []*FileCoverage{}},
				{Name: "pkg2", Percent: 25, Files: []*FileCoverage{}},
				{Name: "pkg3", Percent: 25, Files: []*FileCoverage{}},
			}},
			args: args{"pkg2"},
			want: &PackageCoverage{Name: "pkg2", Percent: 25, Files: []*FileCoverage{}},
		},
		{
			name:   "Package not found",
			fields: fields{0, []*PackageCoverage{}},
			args:   args{"notfound"},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Coverage{
				Percent:  tt.fields.Percent,
				Packages: tt.fields.Packages,
			}
			got, ok := c.Package(tt.args.name)
			if (tt.want != nil) != ok {
				t.Fatalf("Coverage.Package() package %s found=%v, want %v", tt.args.name, ok, tt.want)
			}
			if ok && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Coverage.Package() got = %v, want %v", got, tt.want)
			}
		})
	}
}
