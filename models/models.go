// Package models defines the core data structures and interfaces used across the project.
package models

import "fmt"

// Processor is an interface for types that can process data.
type Processor interface {
	Process(input string) (string, error)
	Name() string
}

// Validator is an interface for types that can validate data.
type Validator interface {
	Validate(input string) bool
}

// User represents an application user with an ID and name.
type User struct {
	ID    int
	Name  string
	Email string
	roles []string
}

// NewUser creates and returns a new User.
func NewUser(id int, name, email string) *User {
	return &User{ID: id, Name: name, Email: email, roles: []string{"viewer"}}
}

// AddRole appends a role to the user's role list.
func (u *User) AddRole(role string) {
	u.roles = append(u.roles, role)
}

// HasRole reports whether the user holds the given role.
func (u *User) HasRole(role string) bool {
	for _, r := range u.roles {
		if r == role {
			return true
		}
	}
	return false
}

// String returns a human-readable representation of the user.
func (u *User) String() string {
	return fmt.Sprintf("User{ID:%d, Name:%q, Email:%q}", u.ID, u.Name, u.Email)
}

// Task represents a unit of work.
type Task struct {
	ID       int
	Title    string
	Priority int
	done     bool
}

// NewTask creates a new Task with the given title and priority.
func NewTask(id int, title string, priority int) *Task {
	return &Task{ID: id, Title: title, Priority: priority}
}

// Complete marks the task as done.
func (t *Task) Complete() {
	t.done = true
}

// IsDone reports whether the task has been completed.
func (t *Task) IsDone() bool {
	return t.done
}

// String returns a human-readable representation of the task.
func (t *Task) String() string {
	status := "pending"
	if t.done {
		status = "done"
	}
	return fmt.Sprintf("Task{ID:%d, Title:%q, Priority:%d, Status:%s}", t.ID, t.Title, t.Priority, status)
}

// Report bundles multiple tasks for reporting.
type Report struct {
	Title string
	Tasks []*Task
}

// NewReport creates an empty Report with the given title.
func NewReport(title string) *Report {
	return &Report{Title: title}
}

// AddTask appends a task to the report.
func (r *Report) AddTask(t *Task) {
	r.Tasks = append(r.Tasks, t)
}

// Summary returns a short summary of task completion.
func (r *Report) Summary() string {
	total := len(r.Tasks)
	done := 0
	for _, t := range r.Tasks {
		if t.IsDone() {
			done++
		}
	}
	return fmt.Sprintf("Report %q: %d/%d tasks completed", r.Title, done, total)
}
