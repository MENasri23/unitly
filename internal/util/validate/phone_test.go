package validate

import "testing"

func TestIsValidPhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		// Valid
		{
			name:  "hamrah aval",
			phone: "+989121234567",
			want:  true,
		},
		{
			name:  "irancell",
			phone: "+989351234567",
			want:  true,
		},
		{
			name:  "rightel",
			phone: "+989201234567",
			want:  true,
		},
		{
			name:  "mvno",
			phone: "+989901234567",
			want:  true,
		},
		{
			name:  "minimum length",
			phone: "+12345678",
			want:  true,
		},
		{
			name:  "US",
			phone: "+12125551234",
			want:  true,
		},
		{
			name:  "Germany",
			phone: "+4915123456789",
			want:  true,
		},
		{
			name:  "UK",
			phone: "+447911123456",
			want:  true,
		},
		{
			name:  "Iran",
			phone: "+989121234567",
			want:  true,
		},
		{
			name:  "maximum length",
			phone: "+123456789012345",
			want:  true,
		},

		// Invalid
		{
			name:  "empty",
			phone: "",
			want:  false,
		},
		{
			name:  "missing plus",
			phone: "989121234567",
			want:  false,
		},
		{
			name:  "local format",
			phone: "09121234567",
			want:  false,
		},
		{
			name:  "too short",
			phone: "+98912123456",
			want:  false,
		},
		{
			name:  "too long",
			phone: "+9891212345678",
			want:  false,
		},
		{
			name:  "contains spaces",
			phone: "+98912 1234567",
			want:  false,
		},
		{
			name:  "contains dash",
			phone: "+98912-1234567",
			want:  false,
		},
		{
			name:  "contains letters",
			phone: "+98912ABC4567",
			want:  false,
		},
		{
			name:  "country code with leading zero",
			phone: "+9809121234567",
			want:  false,
		},
		{
			name:  "double plus",
			phone: "++989121234567",
			want:  false,
		},
		{
			name:  "leading space",
			phone: " +989121234567",
			want:  false,
		},
		{
			name:  "trailing space",
			phone: "+989121234567 ",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPhone(tt.phone)
			if got != tt.want {
				t.Errorf("IsValidPhone(%q) = %v, want %v", tt.phone, got, tt.want)
			}
		})
	}
}
