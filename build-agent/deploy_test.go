package main

import (
	"deploy/model"
	"reflect"
	"sync"
	"testing"
)

func Test_determineWhatNeedsDeploying(t *testing.T) {
	type args struct {
		alreadyDeployed map[string]string
		wantDeployed    map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Both lists have values and intersection contains same keys but different values",
			args: args{
				alreadyDeployed: map[string]string{
					"staging":    "v0.0.4",
					"dev":        "v0.0.4",
					"production": "v0.0.5",
				},
				wantDeployed: map[string]string{
					"staging":    "v0.0.4",
					"dev":        "v0.0.3",
					"production": "v0.0.3",
				},
			},
			want: map[string]string{
				"dev":        "v0.0.3",
				"production": "v0.0.3",
			},
		},
		{
			name: "Both lists have values and intersection contains same values but different keys",
			args: args{
				alreadyDeployed: map[string]string{
					"staging":    "v0.0.4",
					"dev":        "v0.0.4",
					"production": "v0.0.5",
				},
				wantDeployed: map[string]string{
					"staging": "v0.0.4",
					"dev":     "v0.0.4",
					"qa":      "v0.0.5",
				},
			},
			want: map[string]string{
				"qa": "v0.0.5",
			},
		},
		{
			name: "Both lists have values and intersection contains different values and different keys",
			args: args{
				alreadyDeployed: map[string]string{
					"staging":    "v0.0.4",
					"dev":        "v0.0.4",
					"production": "v0.0.5",
				},
				wantDeployed: map[string]string{
					"staging": "v0.0.4",
					"dev":     "v0.0.5",
					"qa":      "v0.0.5",
				},
			},
			want: map[string]string{
				"dev": "v0.0.5",
				"qa":  "v0.0.5",
			},
		},
		{
			name: "One list has values",
			args: args{
				alreadyDeployed: nil,
				wantDeployed: map[string]string{
					"staging":    "v0.0.4",
					"dev":        "v0.0.3",
					"production": "v0.0.3",
				},
			},
			want: map[string]string{
				"staging":    "v0.0.4",
				"dev":        "v0.0.3",
				"production": "v0.0.3",
			},
		},
		{
			name: "Already deployed list has values",
			args: args{
				alreadyDeployed: map[string]string{
					"staging":    "v0.0.4",
					"dev":        "v0.0.3",
					"production": "v0.0.3",
				},
				wantDeployed: nil,
			},
			want: make(map[string]string),
		},
		{
			name: "Want deployed list has values",
			args: args{
				alreadyDeployed: nil,
				wantDeployed: map[string]string{
					"staging":    "v0.0.4",
					"dev":        "v0.0.3",
					"production": "v0.0.3",
				},
			},
			want: map[string]string{
				"staging":    "v0.0.4",
				"dev":        "v0.0.3",
				"production": "v0.0.3",
			},
		},
		{
			name: "Latest key should always appear in the needs deployment map",
			args: args{
				alreadyDeployed: map[string]string{
					"staging":    "latest",
					"dev":        "v0.0.3",
					"production": "v0.0.3",
				},
				wantDeployed: map[string]string{
					"staging":    "latest",
					"dev":        "v0.0.3",
					"production": "v0.0.3",
				},
			},
			want: map[string]string{
				"staging": "latest",
			},
		},
		{
			name: "No changes",
			args: args{
				alreadyDeployed: nil,
				wantDeployed:    nil,
			},
			want: make(map[string]string),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := determineWhatNeedsDeploying(tt.args.alreadyDeployed, tt.args.wantDeployed); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("determineWhatNeedsDeploying() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unmarshalIntoSyncMap(t *testing.T) {
	type args struct {
		data []byte
	}
	var desiredResult sync.Map
	desiredResult.Store("key1", "value1")
	desiredResult.Store("key2", "value2")
	desiredResult.Store("key3", "value3")
	tests := []struct {
		name    string
		args    args
		want    *sync.Map
		wantErr bool
	}{
		{
			name: "Happy path",
			args: args{
				data: []byte(`{"key1": "value1", "key2": "value2", "key3": "value3"}`),
			},
			want:    &desiredResult,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalIntoSyncMap(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalIntoSyncMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			mapsAreEqual(tt.want, got)
			if !mapsAreEqual(tt.want, got) {
				t.Errorf("unmarshalIntoSyncMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func mapsAreEqual(map1, map2 *sync.Map) bool {
	// Count the number of key-value pairs in both maps
	count1 := 0
	map1.Range(func(_, _ interface{}) bool {
		count1++
		return true
	})

	count2 := 0
	map2.Range(func(_, _ interface{}) bool {
		count2++
		return true
	})

	// Check if both maps have the same number of key-value pairs
	if count1 != count2 {
		return false
	}

	var equal bool

	map1.Range(func(key, value interface{}) bool {
		if val, ok := map2.Load(key); ok {
			if val == value {
				equal = true
			} else {
				equal = false
			}
		} else {
			equal = false
		}
		return equal
	})

	return equal
}

func Test_confirmUniqueNameOfDeploymentDirectories(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Path",
			args: args{
				paths: []string{
					"/home/of/the/enchiladas/deployment_config.json",
					"/home/of/the/burritos/deployment_config.json",
					"/home/of/the/watermelon/deployment_config.json",
				},
			},
		},
		{
			name: "Unhappy Path",
			args: args{
				paths: []string{
					"/home/of/the/enchiladas/deployment_config.json",
					"/home/of/the/burritos/deployment_config.json",
					"/home/of/the/burritos/deployment_config.json",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := confirmUniqueNameOfDeploymentDirectories(tt.args.paths); (err != nil) != tt.wantErr {
				t.Errorf("confirmUniqueNameOfDeploymentDirectories() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getBuildNumber(t *testing.T) {
	type args struct {
		healthResponse model.HealthResponse
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Happy Path",
			args: args{
				healthResponse: model.HealthResponse{
					Status:     "ok",
					AppVersion: "0.0.943",
				},
			},
			want: "943",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBuildNumber(tt.args.healthResponse); got != tt.want {
				t.Errorf("getBuildNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}
