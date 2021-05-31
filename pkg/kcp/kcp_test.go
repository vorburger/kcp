package kcp

import "testing"

func TestKCP_Run(t *testing.T) {
	tests := []struct {
		name    string
		kcp     *KCP
		wantErr bool
	}{
		{
			name:    "Listen on :6443",
			kcp:     &KCP{Listen: ":6443"},
			wantErr: false,
		},
		/*{
			name:    "No (empty) Listen",
			kcp:     &KCP{},
			wantErr: true,
		},*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.kcp.Run(); (err != nil) != tt.wantErr {
				t.Errorf("KCP.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
