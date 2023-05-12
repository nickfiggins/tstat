package tstat

// func TestParseTestOutput(t *testing.T) {
// 	f, _ := os.Open("testdata/prog/test.json")
// 	tests := []struct {
// 		name    string
// 		jsonOut io.Reader
// 		want    TestRun
// 		wantErr bool
// 	}{
// 		{
// 			name:    "ha",
// 			jsonOut: f,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := parseTestOutput(tt.jsonOut)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ParseTestOutput() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			assert.Equal(t, got, tt.want)
// 		})
// 	}
// }
