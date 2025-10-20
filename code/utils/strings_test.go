package utils

import "testing"

func TestEitherCutPrefix(t *testing.T) {
	type args struct {
		s      string
		prefix []string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "Prefix match",
			args: args{
				s:      "/system bar",
				prefix: []string{"/system "},
			},
			want:  "bar",
			want1: true,
		},

		{
			name: "Prefix match",
			args: args{
				s:      "act as bar",
				prefix: []string{"act as "},
			},
			want:  "bar",
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := EitherCutPrefix(tt.args.s, tt.args.prefix...)
			if got != tt.want {
				t.Errorf("EitherCutPrefix() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("EitherCutPrefix() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestEitherTrimEqual(t *testing.T) {
	type args struct {
		s      string
		prefix []string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "Prefix match",
			args: args{
				s:      "clear",
				prefix: []string{"clear"},
			},
			want:  "",
			want1: true,
		},
		{
			name: "Prefix match",
			args: args{
				s:      " /clear ",
				prefix: []string{"clear", "/clear"},
			},
			want:  "",
			want1: true,
		},
		{
			name: "Prefix match",
			args: args{
				s:      " clear ",
				prefix: []string{"clear", "/clear"},
			},
			want:  "",
			want1: true,
		},
		{
			name: "Prefix match",
			args: args{
				s:      " reset ",
				prefix: []string{"clear", "/clear"},
			},
			want:  " reset ",
			want1: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := EitherTrimEqual(tt.args.s, tt.args.prefix...)
			if got != tt.want {
				t.Errorf("EitherTrimEqual() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("EitherTrimEqual() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
