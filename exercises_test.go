package main

import (
	"reflect"
	"testing"
)

func Test_selectRandomMuscleGroup(t *testing.T) {
	type args struct {
		availableMuscleGroups []MuscleGroup
	}
	tests := []struct {
		name   string
		args   args
		want   *MuscleGroup
		orWant *MuscleGroup
	}{
		{
			name: "1",
			args: args{
				availableMuscleGroups: []MuscleGroup{
					{
						Id:   "1",
						Name: "abs",
					},
					{
						Id:   "2",
						Name: "chest",
					},
				},
			},
			want: &MuscleGroup{
				Id:   "1",
				Name: "abs",
			},
			orWant: &MuscleGroup{
				Id:   "2",
				Name: "chest",
			},
		},
		{
			name: "2",
			args: args{
				availableMuscleGroups: []MuscleGroup{
					{
						Id:   "1",
						Name: "abs",
					},
				},
			},
			want: &MuscleGroup{
				Id:   "1",
				Name: "abs",
			},
			orWant: &MuscleGroup{
				Id:   "1",
				Name: "abs",
			},
		},
		{
			name: "3",
			args: args{
				availableMuscleGroups: []MuscleGroup{},
			},
			want:   nil,
			orWant: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectRandomMuscleGroup(tt.args.availableMuscleGroups); !reflect.DeepEqual(got, tt.want) && !reflect.DeepEqual(got, tt.orWant) {
				t.Errorf("selectRandomMuscleGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getExerciseArgsAndInsertValues(t *testing.T) {
	type args struct {
		exerciseIds []string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []any
	}{
		{
			name: "1",
			args: args{
				exerciseIds: []string{"1", "2"},
			},
			want:  "$1, $2",
			want1: []any{"1", "2"},
		},
		{
			name: "2",
			args: args{
				exerciseIds: []string{"1"},
			},
			want:  "$1",
			want1: []any{"1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getExerciseArgsAndInsertValues(tt.args.exerciseIds)
			if got != tt.want {
				t.Errorf("getExerciseArgsAndInsertValues() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getExerciseArgsAndInsertValues() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_markPreviousExerciseAsCompleted(t *testing.T) {
	type args struct {
		currentSuperset               *SuperSet
		numberOfAvailableMuscleGroups int
		numberOfExerciseInSuperset    int
	}
	tests := []struct {
		name string
		args args
		want *SuperSet
	}{
		{
			name: "1",
			args: args{
				currentSuperset: &SuperSet{
					Exercises: []Exercise{{
						Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b",
					}},
					CurrentExercisePointer: 0,
					SetCompletionCount:     0,
				},
				numberOfAvailableMuscleGroups: 6,
				numberOfExerciseInSuperset:    3,
			},
			want: &SuperSet{
				Exercises: []Exercise{{
					Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b",
				}},
				CurrentExercisePointer: 1,
				SetCompletionCount:     0,
			},
		},
		{
			name: "2",
			args: args{
				currentSuperset: &SuperSet{
					Exercises: []Exercise{{
						Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b",
					}, {Id: "db1e81f0-f47e-4713-8c27-99822cb651c4"}},
					CurrentExercisePointer: 1,
					SetCompletionCount:     0,
				},
				numberOfAvailableMuscleGroups: 6,
				numberOfExerciseInSuperset:    3,
			},
			want: &SuperSet{
				Exercises: []Exercise{{
					Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b",
				}, {Id: "db1e81f0-f47e-4713-8c27-99822cb651c4"}},
				CurrentExercisePointer: 2,
				SetCompletionCount:     0,
			},
		},
		{
			name: "3",
			args: args{
				currentSuperset: &SuperSet{
					Exercises: []Exercise{
						{Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b"},
						{Id: "db1e81f0-f47e-4713-8c27-99822cb651c4"},
						{Id: "15cb4343-1dd6-42c9-8211-10780e8a11a9"},
					},
					CurrentExercisePointer: 2,
					SetCompletionCount:     0,
				},
				numberOfAvailableMuscleGroups: 6,
				numberOfExerciseInSuperset:    3,
			},
			want: &SuperSet{
				Exercises: []Exercise{
					{Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b"},
					{Id: "db1e81f0-f47e-4713-8c27-99822cb651c4"},
					{Id: "15cb4343-1dd6-42c9-8211-10780e8a11a9"},
				},
				CurrentExercisePointer: 0,
				SetCompletionCount:     1,
			}},
		{
			name: "4",
			args: args{
				currentSuperset: &SuperSet{
					Exercises: []Exercise{{
						Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b",
					}, {Id: "db1e81f0-f47e-4713-8c27-99822cb651c4"}, {Id: "15cb4343-1dd6-42c9-8211-10780e8a11a9"}},
					CurrentExercisePointer: 0,
					SetCompletionCount:     1,
				},
				numberOfAvailableMuscleGroups: 6,
				numberOfExerciseInSuperset:    3,
			},
			want: &SuperSet{
				Exercises: []Exercise{{
					Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b",
				}, {Id: "db1e81f0-f47e-4713-8c27-99822cb651c4"},
					{Id: "15cb4343-1dd6-42c9-8211-10780e8a11a9"},
				},
				CurrentExercisePointer: 1,
				SetCompletionCount:     1,
			}},
		{
			name: "5",
			args: args{
				currentSuperset: &SuperSet{
					Exercises: []Exercise{{
						Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b"},
					},
					CurrentExercisePointer: 0,
					SetCompletionCount:     0,
				},
				numberOfAvailableMuscleGroups: 2,
				numberOfExerciseInSuperset:    3,
			},
			want: &SuperSet{
				Exercises: []Exercise{{
					Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b",
				}},
				CurrentExercisePointer: 1,
				SetCompletionCount:     0,
			},
		},
		{
			name: "6",
			args: args{
				currentSuperset: &SuperSet{
					Exercises: []Exercise{{
						Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b",
					}, {Id: "db1e81f0-f47e-4713-8c27-99822cb651c4"}},
					CurrentExercisePointer: 1,
					SetCompletionCount:     0,
				},
				numberOfAvailableMuscleGroups: 0,
				numberOfExerciseInSuperset:    3,
			},
			want: &SuperSet{
				Exercises: []Exercise{{
					Id: "878cdd10-e11f-4925-bd2e-d0909d616b1b",
				}, {Id: "db1e81f0-f47e-4713-8c27-99822cb651c4"}},
				CurrentExercisePointer: 0,
				SetCompletionCount:     1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := markPreviousExerciseAsCompleted(tt.args.currentSuperset, tt.args.numberOfAvailableMuscleGroups, tt.args.numberOfExerciseInSuperset); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("markPreviousExerciseAsCompleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasMuscleGroupWorkedSessionLimitBeenReached(t *testing.T) {
	type args struct {
		totalMuscleGroupsCount int
		count                  int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1",
			args: args{
				totalMuscleGroupsCount: 19,
				count:                  9,
			},
			want: false,
		},
		{
			name: "1",
			args: args{
				totalMuscleGroupsCount: 9,
				count:                  4,
			},
			want: false,
		},
		{
			name: "2",
			args: args{
				totalMuscleGroupsCount: 20,
				count:                  10,
			},
			want: true,
		},
		{
			name: "3",
			args: args{
				totalMuscleGroupsCount: 20,
				count:                  11,
			},
			want: true,
		},
		{
			name: "4",
			args: args{
				totalMuscleGroupsCount: 20,
				count:                  9,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasMuscleGroupWorkedSessionLimitBeenReached(tt.args.totalMuscleGroupsCount, tt.args.count); got != tt.want {
				t.Errorf("hasMuscleGroupWorkedSessionLimitBeenReached() = %v, want %v", got, tt.want)
			}
		})
	}
}
