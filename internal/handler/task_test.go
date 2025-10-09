package handler

//
//import (
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//)
//
//func TestAbs(t *testing.T) {
//	tests := []struct {
//		name  string
//		value float64
//		want  float64
//	}{
//		{"simple negative value", -5, 5},
//		{"simple positive value", 7, 7},
//		{"zero", 0, 0},
//		{"small value", -0.0001, 0.0001},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			got := Abs(tt.value)
//			assert.Equal(t, tt.want, got)
//		})
//	}
//}
//
//// ------------------- Test User -------------------
//
//func TestUser_FullName(t *testing.T) {
//	tests := []struct {
//		name      string
//		firstName string
//		lastName  string
//		want      string
//	}{
//		{"simple test", "Misha", "Popov", "Misha Popov"},
//		{"long name", "Alexandrina", "Constantinopolis", "Alexandrina Constantinopolis"},
//		{"with numbers", "User123", "Test456", "User123 Test456"},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			u := User{
//				FirstName: tt.firstName,
//				LastName:  tt.lastName,
//			}
//			assert.Equal(t, tt.want, u.FullName())
//		})
//	}
//}
//
//// ------------------- Test Family -------------------
//
//func TestFamily_AddNew(t *testing.T) {
//	tests := []struct {
//		name    string
//		initial map[Relationship]Person
//		r       Relationship
//		p       Person
//		wantErr bool
//	}{
//		{
//			name:    "add father",
//			initial: map[Relationship]Person{},
//			r:       Father,
//			p:       Person{FirstName: "Misha", LastName: "Popov", Age: 56},
//			wantErr: false,
//		},
//		{
//			name: "catch error",
//			initial: map[Relationship]Person{
//				Father: {FirstName: "Misha", LastName: "Popov", Age: 56},
//			},
//			r:       Father,
//			p:       Person{FirstName: "Drug", LastName: "Mishi", Age: 57},
//			wantErr: true,
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			f := &Family{
//				Members: tt.initial,
//			}
//			err := f.AddNew(tt.r, tt.p)
//			if tt.wantErr {
//				require.Error(t, err)
//			} else {
//				require.NoError(t, err)
//				assert.Contains(t, f.Members, tt.r)
//				assert.Equal(t, tt.p, f.Members[tt.r])
//			}
//		})
//	}
//}
//
//func Test_main(t *testing.T) {
//	// Просто проверяем, что main выполняется без паники
//	require.NotPanics(t, func() {
//		main()
//	})
//}
