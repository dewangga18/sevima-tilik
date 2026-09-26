package domain

// AdminClass is the classroom shape used by admin management screens. It is
// separate from TeacherClass because admin needs the teacher roster size, while
// the teacher view needs the assessed-student count for its own classes.
type AdminClass struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	GradeLevel   int    `json:"grade_level"`
	StudentCount int    `json:"student_count"`
	TeacherCount int    `json:"teacher_count"`
}

// ClassRoster lists the enrolled students and assigned teachers of one
// classroom. Admin uses it to place and remove people; teachers never receive
// this shape because it exposes assignments outside their own classes.
type ClassRoster struct {
	ClassroomID string   `json:"classroom_id"`
	StudentIDs  []string `json:"student_ids"`
	TeacherIDs  []string `json:"teacher_ids"`
}
