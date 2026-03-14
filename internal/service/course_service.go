package service

import (
	"time"

	"github.com/Boyeep/flippy-backend/internal/domain"
)

type CourseService struct{}

func NewCourseService() CourseService {
	return CourseService{}
}

func (s CourseService) List() []domain.Course {
	now := time.Now().UTC()

	return []domain.Course{
		{
			ID:          "seed-math-foundations",
			Slug:        "math-foundations",
			Title:       "Math Foundations",
			Category:    "Mathematics",
			Description: "Bangun dasar matematika yang kuat sebelum naik ke materi lanjutan.",
			Summary:     "Foundational mathematics for structured review.",
			Published:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "seed-programming-basics",
			Slug:        "programming-basics",
			Title:       "Programming Basics",
			Category:    "Programming",
			Description: "Kartu belajar untuk memahami syntax, logic, dan pola coding dasar.",
			Summary:     "Programming essentials for quick repetition.",
			Published:   true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}
