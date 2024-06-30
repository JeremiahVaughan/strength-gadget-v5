package main

import (
	"fmt"
	"strings"
)

// const (
// 	Weightlifting = "6bdb3624-bed1-41a9-bf8c-7b1066411446"
// 	Calisthenics  = "8ffe7196-4e3d-4439-ae19-3159ad5387bd"
// 	Cardio        = "982d0b18-a67c-401a-95f2-ddb702ba80b5"
// 	WarmUp        = "ce6133be-2bd8-48e9-adbb-05f03ad7b4f9"
// 	CoolDown      = "db085937-cd84-406a-b9db-34f9e091816b"
// )

type SuperSet struct {
	Exercises              []Exercise `json:"exercise"`
	CurrentExercisePointer int        `json:"currentExercisePointer"`
	SetCompletionCount     int        `json:"completionCount"`
	SuperSetProgress
}

type SuperSetProgress struct {
	WorkoutComplete bool `json:"workoutComplete"`
}

type ExerciseUserData struct {
	Measurement     int `json:"measurement"`
	SelectionOffset int `json:"selectionOffset"`
}

type ExerciseType int

// const (
// 	ExerciseTypeWeightlifting ExerciseType = "6bdb3624-bed1-41a9-bf8c-7b1066411446"
// 	ExerciseTypeCalisthenics  ExerciseType = "8ffe7196-4e3d-4439-ae19-3159ad5387bd"
// 	ExerciseTypeCardio        ExerciseType = "982d0b18-a67c-401a-95f2-ddb702ba80b5"
// 	ExerciseTypeWarmUp        ExerciseType = "ce6133be-2bd8-48e9-adbb-05f03ad7b4f9"
// 	ExerciseTypeCoolDown      ExerciseType = "db085937-cd84-406a-b9db-34f9e091816b"
// )

const (
	ExerciseTypeWeightlifting ExerciseType = iota
	ExerciseTypeCalisthenics
	ExerciseTypeCardio
	ExerciseTypeWarmUp
	ExerciseTypeCoolDown
)

var (
	MuscleGroupHipAdductors = MuscleGroup{
		Id:      0,
		Name:    "Hip Adductors",
		Routine: LOWER,
	}
	MuscleGroupGlutes = MuscleGroup{
		Id:      1,
		Name:    "Glutes",
		Routine: LOWER,
	}
	MuscleGroupQuadriceps = MuscleGroup{
		Id:      2,
		Name:    "Quadriceps",
		Routine: LOWER,
	}
	MuscleGroupAbductors = MuscleGroup{
		Id:      3,
		Name:    "Abductors",
		Routine: LOWER,
	}
	MuscleGroupHamstrings = MuscleGroup{
		Id:      4,
		Name:    "Hamstrings",
		Routine: LOWER,
	}
	MuscleGroupCalves = MuscleGroup{
		Id:      5,
		Name:    "Calves",
		Routine: LOWER,
	}

	MuscleGroupObliques = MuscleGroup{
		Id:      6,
		Name:    "Obliques",
		Routine: CORE,
	}
	MuscleGroupTransverseAbdominis = MuscleGroup{
		Id:      7,
		Name:    "Transverse Abdominis",
		Routine: CORE,
	}
	MuscleGroupRectusAbdominis = MuscleGroup{
		Id:      8,
		Name:    "Rectus Abdominis",
		Routine: CORE,
	}
	MuscleGroupMultifidus = MuscleGroup{
		Id:      9,
		Name:    "Multifidus",
		Routine: CORE,
	}
	MuscleGroupHipFlexors = MuscleGroup{
		Id:      10,
		Name:    "Hip Flexors",
		Routine: CORE,
	}
	MuscleGroupQuadratusLumborum = MuscleGroup{
		Id:      11,
		Name:    "Quadratus Lumborum",
		Routine: CORE,
	}

	MuscleGroupForearmsAndGripStrength = MuscleGroup{
		Id:      12,
		Name:    "Forearms and Grip Strength",
		Routine: UPPER,
	}
	MuscleGroupTriceps = MuscleGroup{
		Id:      13,
		Name:    "Triceps",
		Routine: UPPER,
	}
	MuscleGroupBiceps = MuscleGroup{
		Id:      14,
		Name:    "Biceps",
		Routine: UPPER,
	}
	MuscleGroupBack = MuscleGroup{
		Id:      15,
		Name:    "Back",
		Routine: UPPER,
	}
	MuscleGroupChest = MuscleGroup{
		Id:      16,
		Name:    "Chest",
		Routine: UPPER,
	}
	MuscleGroupErectorSpinae = MuscleGroup{
		Id:      17,
		Name:    "Erector Spinae",
		Routine: UPPER,
	}
	MuscleGroupShoulders = MuscleGroup{
		Id:      18,
		Name:    "Shoulders",
		Routine: UPPER,
	}
	MuscleGroupCardio = MuscleGroup{
		Id:      19,
		Name:    "Cardio",
		Routine: ALL,
	}
)

var AllExercises []Exercise = []Exercise{
	{
		Id:                   0,
		Name:                 "Arnold Press",
		DemonstrationGiphyId: "VKAMruVRegKGtl82Bh",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupShoulders,
		},
	},
	{
		Id:                   1,
		Name:                 "Back Squats",
		DemonstrationGiphyId: "W5gFEeJmRhvElyatmF",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadriceps,
			MuscleGroupGlutes,
		},
	},
	{
		Id:                   2,
		Name:                 "Barbell Curl",
		DemonstrationGiphyId: "l3q2UKzEToTKEvlK0",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupBiceps,
		},
	},
	{
		Id:                   3,
		Name:                 "Bench Dip",
		DemonstrationGiphyId: "13HOBYXe87LjvW",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTriceps,
		},
	},
	{
		Id:                   4,
		Name:                 "Bench Press",
		DemonstrationGiphyId: "DgAyvWzSjOhSNzaRwU",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupChest,
		},
	},
	{
		Id:                   5,
		Name:                 "Bent Over Reverse Fly",
		DemonstrationGiphyId: "zBvSThvnE0Cj2ikfTC",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupBack,
		},
	},
	{
		Id:                   6,
		Name:                 "Bicycle Crunches",
		DemonstrationGiphyId: "IkFw3Mnwi6g7tO5AN1",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupRectusAbdominis,
			MuscleGroupObliques,
			MuscleGroupTransverseAbdominis,
		},
	},
	{
		Id:                   7,
		Name:                 "Bird Dog",
		DemonstrationGiphyId: "RpwQmzE45R3NTBaszK",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupErectorSpinae,
			MuscleGroupTransverseAbdominis,
		},
	},
	{
		Id:                   8,
		Name:                 "Body Weight Row",
		DemonstrationGiphyId: "znlXkg1Q61ObI75Dfx",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupBack,
		},
	},
	{
		Id:                   9,
		Name:                 "Body Weight Squats",
		DemonstrationGiphyId: "wxNwwnoYUyxxWqvySO",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadriceps,
			MuscleGroupGlutes,
		},
	},
	{
		Id:                   10,
		Name:                 "Burpees",
		DemonstrationGiphyId: "sdPLLtuVeRdJ9DyCej",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupChest,
		},
	},
	{
		Id:                   11,
		Name:                 "Cat-Cow Stretch",
		DemonstrationGiphyId: "y9t6xaAUh1ECA",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTransverseAbdominis,
		},
	},
	{
		Id:                   12,
		Name:                 "Child's Pose",
		DemonstrationGiphyId: "tvZw5zZHKxqrsN5efG",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupBack,
		},
	},
	{
		Id:                   13,
		Name:                 "Child's Pose with a Twist",
		DemonstrationGiphyId: "aspKT1lhYZFMHG2hGI",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadratusLumborum,
		},
	},
	{
		Id:                   14,
		Name:                 "Chin ups",
		DemonstrationGiphyId: "CyESeFgx6xgNG",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupBack,
		},
	},
	{
		Id:                   15,
		Name:                 "Clasped hands behind back arm extension stretch",
		DemonstrationGiphyId: "0TPz7BQ3C6stkAvMsg",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupBiceps,
		},
	},
	{
		Id:                   16,
		Name:                 "Cossack Squat",
		DemonstrationGiphyId: "cBpGYSVowzPooqn5EY",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadriceps,
			MuscleGroupGlutes,
			MuscleGroupHamstrings,
			MuscleGroupHipAdductors,
		},
	},
	{
		Id:                   17,
		Name:                 "Cross-Body Shoulder Stretch",
		DemonstrationGiphyId: "yBUSNdYGhZZep3omGU",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupShoulders,
		},
	},
	{
		Id:                   18,
		Name:                 "Dead Bug",
		DemonstrationGiphyId: "XjvQ1Tfu8y1xYK23vi",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTransverseAbdominis,
			MuscleGroupRectusAbdominis,
		},
	},
	{
		Id:                   19,
		Name:                 "Dead Hang",
		DemonstrationGiphyId: "5efhbxS9FnTSYLIP2R",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupForearmsAndGripStrength,
		},
	},
	{
		Id:                   20,
		Name:                 "Dips",
		DemonstrationGiphyId: "kojJRlvnjCsIB26YUD",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTriceps,
			MuscleGroupChest,
		},
	},
	{
		Id:                   21,
		Name:                 "Donkey Kicks",
		DemonstrationGiphyId: "h9GtNqlr3GqPxqigxO",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupGlutes,
		},
	},
	{
		Id:                   22,
		Name:                 "Doorway Stretch",
		DemonstrationGiphyId: "zUBf9S1NQ6lxuhxm6k",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupChest,
		},
	},
	{
		Id:                   23,
		Name:                 "Downward Dog Pose",
		DemonstrationGiphyId: "Stha3DqTeY9wIlXnCW",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupCalves,
		},
	},
	{
		Id:                   24,
		Name:                 "Dumbell Bicep Curl",
		DemonstrationGiphyId: "HMzadBUQG3y53pYmOK",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupBiceps,
		},
	},
	{
		Id:                   25,
		Name:                 "Dumbell Shoulder Press",
		DemonstrationGiphyId: "7lugb7ObGYiXe",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupShoulders,
		},
	},
	{
		Id:                   26,
		Name:                 "Dumbell Skull Crushers",
		DemonstrationGiphyId: "hNrpV2ksuav6sVFMrg",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTriceps,
		},
	},
	{
		Id:                   27,
		Name:                 "Extended-Range, Side-Lying Hip Abduction",
		DemonstrationGiphyId: "RtBE6pTJU3zm9qng6e",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupAbductors,
		},
	},
	{
		Id:                   28,
		Name:                 "Front Raise with Plate",
		DemonstrationGiphyId: "FfJxatjLUw6S5MZIzq",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupShoulders,
		},
	},
	{
		Id:                   29,
		Name:                 "Front Squats",
		DemonstrationGiphyId: "AmYRXqZyVy0GVigC2l",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadriceps,
		},
	},
	{
		Id:                   30,
		Name:                 "Hand Grip Strengthener",
		DemonstrationGiphyId: "vw7QhHQwQGjeh4GO0Z",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupForearmsAndGripStrength,
		},
	},
	{
		Id:                   31,
		Name:                 "Hanging Leg Raise",
		DemonstrationGiphyId: "HAuk68YCmu1bI4N8Km",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupHipFlexors,
		},
	},
	{
		Id:                   32,
		Name:                 "High Plank Knee-to-Elbow",
		DemonstrationGiphyId: "JFdWzEYkbK7tyf1A1y",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupRectusAbdominis,
			MuscleGroupTransverseAbdominis,
			MuscleGroupObliques,
		},
	},
	{
		Id:                   33,
		Name:                 "Hip Thrust With Barbell",
		DemonstrationGiphyId: "hVhEI2iYXhrmi1B2EU",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupGlutes,
		},
	},
	{
		Id:                   34,
		Name:                 "Incline Dumbell Bench Press",
		DemonstrationGiphyId: "uT4BlvvPm0TQYvuXCO",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupChest,
			MuscleGroupShoulders,
		},
	},
	{
		Id:                   35,
		Name:                 "Jogging",
		DemonstrationGiphyId: "JquWhb4LRd8ULLvuxJ",
		MeasurementType:      MeasurementTypeMile,
		ExerciseType:         ExerciseTypeCardio,
		MuscleGroups: []MuscleGroup{
			MuscleGroupCardio,
		},
	},
	{
		Id:                   36,
		Name:                 "Jump Squats",
		DemonstrationGiphyId: "I6YSGpRMYGNwzxAo0d",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadriceps,
			MuscleGroupGlutes,
		},
	},
	{
		Id:                   37,
		Name:                 "Kettlebell Farmers Carry",
		DemonstrationGiphyId: "1hpB1Qo3it4B6qThql",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupForearmsAndGripStrength,
		},
	},
	{
		Id:                   38,
		Name:                 "Knee-to-Chest Stretch",
		DemonstrationGiphyId: "tdvrnNICQzTX3CSrkC",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupErectorSpinae,
		},
	},
	{
		Id:                   39,
		Name:                 "Kneeling hip flexor stretch",
		DemonstrationGiphyId: "LZfFjb2AxgCp9GPVEN",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupHipFlexors,
		},
	},
	{
		Id:                   40,
		Name:                 "Leg Raises",
		DemonstrationGiphyId: "55atXlETBRZm0h9NrO",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupRectusAbdominis,
			MuscleGroupTransverseAbdominis,
		},
	},
	{
		Id:                   41,
		Name:                 "Lumbar Rotation Stretch",
		DemonstrationGiphyId: "S7GLaRnDukO8QbtwZc",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupMultifidus,
		},
	},
	{
		Id:                   42,
		Name:                 "Lunges",
		DemonstrationGiphyId: "N9hhNLh26xarKxizrY",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupGlutes,
			MuscleGroupQuadriceps,
		},
	},
	{
		Id:                   43,
		Name:                 "Mountain Climbers",
		DemonstrationGiphyId: "4NojW5eV2t2yY4R0JA",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupRectusAbdominis,
			MuscleGroupTransverseAbdominis,
		},
	},
	{
		Id:                   44,
		Name:                 "Overhead Press",
		DemonstrationGiphyId: "SieD4F7finpC5Bdal5",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupShoulders,
		},
	},
	{
		Id:                   45,
		Name:                 "Pigeon Pose",
		DemonstrationGiphyId: "RtDeFx5s3t5rCyv5yl",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupGlutes,
		},
	},
	{
		Id:                   46,
		Name:                 "Plank",
		DemonstrationGiphyId: "kTdej6DP88WuWkggCp",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTransverseAbdominis,
			MuscleGroupRectusAbdominis,
		},
	},
	{
		Id:                   47,
		Name:                 "Plank Jacks",
		DemonstrationGiphyId: "T8JPUqEXpl5ilx3uom",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTransverseAbdominis,
			MuscleGroupRectusAbdominis,
		},
	},
	{
		Id:                   48,
		Name:                 "Plank to Forearm Plank",
		DemonstrationGiphyId: "mvmrk7kZ8nORvgloyt",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTransverseAbdominis,
			MuscleGroupRectusAbdominis,
		},
	},
	{
		Id:                   49,
		Name:                 "Plank with Hip Twist",
		DemonstrationGiphyId: "F9JNNNn1YVGe3cACYZ",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTransverseAbdominis,
			MuscleGroupRectusAbdominis,
			MuscleGroupQuadratusLumborum,
			MuscleGroupObliques,
		},
	},
	{
		Id:                   50,
		Name:                 "Prayer Stretch",
		DemonstrationGiphyId: "4ARyPx0M5JxNxy7V37",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupForearmsAndGripStrength,
		},
	},
	{
		Id:                   51,
		Name:                 "Prone Quad Stretch",
		DemonstrationGiphyId: "mFLMAKvYU0UgIYONi8",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadriceps,
		},
	},
	{
		Id:                   52,
		Name:                 "Push Up",
		DemonstrationGiphyId: "eWTbWzSVTBjONkm7Nm",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupChest,
		},
	},
	{
		Id:                   53,
		Name:                 "Reverse Lunges",
		DemonstrationGiphyId: "JO4nBEsc6Dlyw3FAvX",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupGlutes,
			MuscleGroupQuadriceps,
		},
	},
	{
		Id:                   54,
		Name:                 "Reverse Plank",
		DemonstrationGiphyId: "25a0Zl7launMhiD6LU",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupErectorSpinae,
		},
	},
	{
		Id:                   55,
		Name:                 "Rower",
		DemonstrationGiphyId: "939CzM7Wu408RQpQcQ",
		MeasurementType:      MeasurementTypeMile,
		ExerciseType:         ExerciseTypeCardio,
		MuscleGroups: []MuscleGroup{
			MuscleGroupCardio,
		},
	},
	{
		Id:                   56,
		Name:                 "Russian Twists",
		DemonstrationGiphyId: "8mRpGumgGIswSG2dD5",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTransverseAbdominis,
			MuscleGroupObliques,
		},
	},
	{
		Id:                   57,
		Name:                 "Scissors",
		DemonstrationGiphyId: "adUjSmrdqTP8Wi0pUG",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupRectusAbdominis,
			MuscleGroupTransverseAbdominis,
		},
	},
	{
		Id:                   58,
		Name:                 "Seated Butterfly Stretch",
		DemonstrationGiphyId: "nYfhcrdFUYNFzcZ7Dh",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupHipAdductors,
		},
	},
	{
		Id:                   59,
		Name:                 "Seated Torso Twist",
		DemonstrationGiphyId: "HSJozjUevFYqJPyQGU",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupObliques,
		},
	},
	{
		Id:                   60,
		Name:                 "Shoulder Shrug with Dumbbells",
		DemonstrationGiphyId: "I8etqXSAvVUm2L4SQm",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupShoulders,
		},
	},
	{
		Id:                   61,
		Name:                 "Side Bends With Dumbell",
		DemonstrationGiphyId: "HwFdMdIGm0hk7k13jn",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadratusLumborum,
			MuscleGroupObliques,
		},
	},
	{
		Id:                   62,
		Name:                 "Side Lateral Raise",
		DemonstrationGiphyId: "xT8qBikQMiEkfvI2yc",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupShoulders,
		},
	},
	{
		Id:                   63,
		Name:                 "Side Leg Raises",
		DemonstrationGiphyId: "jgMpJCqtCij8XGsvIe",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupAbductors,
		},
	},
	{
		Id:                   64,
		Name:                 "Side Lunge Stretch",
		DemonstrationGiphyId: "wjITHGzAXciebSu1Wn",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupAbductors,
		},
	},
	{
		Id:                   65,
		Name:                 "Side Plank",
		DemonstrationGiphyId: "YeYfAFgpamPVi29TNI",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadratusLumborum,
			MuscleGroupObliques,
		},
	},
	{
		Id:                   66,
		Name:                 "Single-leg deadlift with Kettlebell",
		DemonstrationGiphyId: "JUx7urz9o0N65oA4W7",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupHamstrings,
		},
	},
	{
		Id:                   67,
		Name:                 "Split Squat",
		DemonstrationGiphyId: "hTapq4yx9cZ9Ya0iHH",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadriceps,
			MuscleGroupGlutes,
		},
	},
	{
		Id:                   68,
		Name:                 "Squat Jack",
		DemonstrationGiphyId: "bfqGjbLXbbAn8uiryX",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadriceps,
			MuscleGroupGlutes,
		},
	},
	{
		Id:                   69,
		Name:                 "Standing Ab Wheel Rollouts",
		DemonstrationGiphyId: "Z55ueeg3MfUwp8qEa2",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupErectorSpinae,
			MuscleGroupTransverseAbdominis,
			MuscleGroupRectusAbdominis,
			MuscleGroupBack,
		},
	},
	{
		Id:                   70,
		Name:                 "Standing Calve Raises",
		DemonstrationGiphyId: "2wXXVCek2NfkneGqz9",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupCalves,
		},
	},
	{
		Id:                   71,
		Name:                 "Standing Hamstring Stretch",
		DemonstrationGiphyId: "l0COHO3tcUVOzDy6c",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupHamstrings,
		},
	},
	{
		Id:                   72,
		Name:                 "Strait Bar Dips",
		DemonstrationGiphyId: "NpZa6bI0VHMY2ZryfK",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTriceps,
			MuscleGroupChest,
		},
	},
	{
		Id:                   73,
		Name:                 "Strait Leg Deadlift with Barbell",
		DemonstrationGiphyId: "oYK8O344YusZHZKW7S",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupHamstrings,
		},
	},
	{
		Id:                   74,
		Name:                 "Strait Leg Deadlift with Dumbbells",
		DemonstrationGiphyId: "xT0xenc4lKQlhf1Ohi",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupHamstrings,
		},
	},
	{
		Id:                   75,
		Name:                 "Stretch Triceps Behind Back with Towel",
		DemonstrationGiphyId: "EqxmwZIjTpAdzmmDkr",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupTriceps,
		},
	},
	{
		Id:                   76,
		Name:                 "Sumo Squat with Barbell",
		DemonstrationGiphyId: "a7Y2DhvZX4D3EdluiZ",
		MeasurementType:      MeasurementTypePounds,
		ExerciseType:         ExerciseTypeWeightlifting,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadriceps,
			MuscleGroupHipAdductors,
			MuscleGroupGlutes,
		},
	},
	{
		Id:                   77,
		Name:                 "Superman",
		DemonstrationGiphyId: "xxQIlMGciMFUsDA6DX",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupMultifidus,
			MuscleGroupErectorSpinae,
		},
	},
	{
		Id:                   78,
		Name:                 "Swiss Ball Hamstring Curl",
		DemonstrationGiphyId: "61SeVwDFrI0ZfnMLaN",
		MeasurementType:      MeasurementTypeRepetition,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupHamstrings,
		},
	},
	{
		Id:                   79,
		Name:                 "Upward Dog",
		DemonstrationGiphyId: "mqjnXsqlSmrFJDjpNM",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCoolDown,
		MuscleGroups: []MuscleGroup{
			MuscleGroupRectusAbdominis,
		},
	},
	{
		Id:                   80,
		Name:                 "Wall Sit",
		DemonstrationGiphyId: "kAsOw4LRzKvKZPh6fc",
		MeasurementType:      MeasurementTypeSecond,
		ExerciseType:         ExerciseTypeCalisthenics,
		MuscleGroups: []MuscleGroup{
			MuscleGroupQuadriceps,
			MuscleGroupGlutes,
		},
	},
}

type MeasurementType int

type Exercise struct {
	Id                       int
	Name                     string
	DemonstrationGiphyId     string
	LastCompletedMeasurement int
	MeasurementType          MeasurementType
	ExerciseType             ExerciseType
	MuscleGroups             []MuscleGroup
}

type ExerciseDisplay struct {
	SelectMode        bool
	Cool              Button
	Hot               Button
	Yes               Button
	No                Button
	Complete          Button
	NextProgressIndex int
	WorkoutCompleted  bool

	Exercise Exercise
}

func hasMuscleGroupWorkedSessionLimitBeenReached(totalMuscleGroupsCount int, count int) bool {
	// Adding one before division if totalMuscleGroupsCount be odd to handle ceiling
	halfMuscleGroups := totalMuscleGroupsCount / 2
	if totalMuscleGroupsCount%2 != 0 {
		halfMuscleGroups++
	}

	return halfMuscleGroups <= count
}

// func markPreviousExerciseAsCompleted(currentSuperset *SuperSet, numberOfAvailableMuscleGroups int, numberOfExerciseInSuperset int) *SuperSet {
// 	numberOfActiveExercises := len(currentSuperset.Exercises)
// 	currentExerciseNumber := currentSuperset.CurrentExercisePointer + 1
// 	if currentExerciseNumber == numberOfExerciseInSuperset || (numberOfAvailableMuscleGroups == 0 && numberOfActiveExercises == currentExerciseNumber) {
// 		currentSuperset.CurrentExercisePointer = 0
// 		currentSuperset.SetCompletionCount++
// 	} else {
// 		currentSuperset.CurrentExercisePointer++
// 	}
// 	return currentSuperset
// }

func getExerciseArgsAndInsertValues(exerciseIds []string) (string, []any) {
	var exercisesArgsSlice []string
	var insertValues []any
	for i, exerciseId := range exerciseIds {
		exercisesArgsSlice = append(exercisesArgsSlice, fmt.Sprintf("$%d", i+1))
		insertValues = append(insertValues, exerciseId)
	}
	return strings.Join(exercisesArgsSlice, ", "), insertValues
}

// func selectRandomMuscleGroup(availableMuscleGroups []MuscleGroup) *MuscleGroup {
// 	muscleGroupCount := len(availableMuscleGroups)
// 	if muscleGroupCount == 0 {
// 		return nil
// 	}
// 	result := availableMuscleGroups[rand.Intn(muscleGroupCount)]
// 	return &result
// }

// generateExerciseMap return value third key is muscle group id, the value is the exercises that target the muscle group
func generateExerciseMap() map[RoutineType]map[ExerciseType]map[int][]Exercise {
	result := make(map[RoutineType]map[ExerciseType]map[int][]Exercise)
	for _, exercise := range AllExercises {
		for _, mg := range exercise.MuscleGroups {
			// Initialize nested maps and slices if they do not exist yet
			if result[mg.Routine] == nil {
				result[mg.Routine] = make(map[ExerciseType]map[int][]Exercise)
			}
			if result[mg.Routine][exercise.ExerciseType] == nil {
				result[mg.Routine][exercise.ExerciseType] = make(map[int][]Exercise)
			}
			// Categorize the exercise
			result[mg.Routine][exercise.ExerciseType][mg.Id] =
				append(result[mg.Routine][exercise.ExerciseType][mg.Id], exercise)
		}
	}
	return result
}
