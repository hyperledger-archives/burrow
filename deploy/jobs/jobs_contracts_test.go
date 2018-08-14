package jobs

import "testing"

func Test_matchInstanceName(t *testing.T) {
	type args struct {
		objectName     string
		deployInstance string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"",
			args{
				objectName:     "contracts/storage.sol:SimpleConstructorArray",
				deployInstance: "SimpleConstructorArray",
			},
			true,
		},
		{
			"",
			args{
				objectName:     "storage.sol:SimpleConstructorArray",
				deployInstance: "simpleConstructorArray",
			},
			true,
		},
		{
			"",
			args{
				objectName:     "SimpleConstructorArray",
				deployInstance: "simpleconstructorarray",
			},
			true,
		},
		{
			"",
			args{
				objectName:     "",
				deployInstance: "Simpleconstructorarray",
			},
			false,
		},
		{
			"",
			args{
				objectName:     "SimpleConstructorArray:",
				deployInstance: "SimpleConstructorArray",
			},
			false,
		},
		{
			"",
			args{
				objectName:     ":",
				deployInstance: "SimpleConstructorArray",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchInstanceName(tt.args.objectName, tt.args.deployInstance); got != tt.want {
				t.Errorf("matchInstanceName() = %v, want %v", got, tt.want)
			}
		})
	}
}
