package grades

func init() {
	students = []Student{
		{
			ID:        1,
			FirstName: "harusame",
			LastName:  "Z",
			Grades: []Grade{
				{
					Title: "Quiz 1",
					Type:  GradeQuiz,
					Score: 85,
				},
				{
					Title: "Quiz 2",
					Type:  GradeTest,
					Score: 90,
				},
				{
					Title: "Quiz 3",
					Type:  GradeQuiz,
					Score: 95,
				},
				{
					Title: "Quiz 4",
					Type:  GradeExam,
					Score: 100,
				},
			},
		},
		{
			ID:        2,
			FirstName: "coco",
			LastName:  "S",
			Grades: []Grade{
				{
					Title: "Quiz 1",
					Type:  GradeQuiz,
					Score: 85,
				},
				{
					Title: "Quiz 2",
					Type:  GradeTest,
					Score: 90,
				},
				{
					Title: "Quiz 3",
					Type:  GradeQuiz,
					Score: 95,
				},
				{
					Title: "Quiz 4",
					Type:  GradeExam,
					Score: 100,
				},
			},
		},
	}
}
